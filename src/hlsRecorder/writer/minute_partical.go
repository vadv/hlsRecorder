package writer

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	hedx "hlsRecorder/index/hedx"
	keys "hlsRecorder/keys"
	stat "hlsRecorder/stat"
)

func round(f float64) float64 {
	return float64(int64(f*100)) / 100
}

// открываем индексный файл и проверяем последние индексы
// и записываем все сегменты что больше чем эти последние индексы
func (m *minute) writePartical(indexDir, storageDir, resource string, vmx *keys.VMX, channelInfo *stat.ChannelInfo) (error, int64, int64, float64) {

	chunkWrited, iframeWrited, last := int64(0), int64(0), float64(0)

	if len(m.chunks) == 0 {
		return fmt.Errorf("пустая минутка"), chunkWrited, iframeWrited, last
	}

	chunkFile, iframeFile, indexFile := m.getPath(storageDir, `.ets`), m.getPath(storageDir, `.ets.ifr`), m.getPath(indexDir, `.ets.hedx`)

	if err := os.MkdirAll(filepath.Dir(chunkFile), 0755); err != nil {
		return err, chunkWrited, iframeWrited, last
	}
	if err := os.MkdirAll(filepath.Dir(indexFile), 0755); err != nil {
		return err, chunkWrited, iframeWrited, last
	}

	chunkFD, err := os.OpenFile(chunkFile, os.O_RDWR|os.O_CREATE|syscall.O_APPEND, 0644)
	if err != nil {
		return err, chunkWrited, iframeWrited, last
	}
	defer chunkFD.Close()

	indexFD, err := os.OpenFile(indexFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err, chunkWrited, iframeWrited, last
	}
	defer indexFD.Close()
	if _, err := indexFD.Seek(0, 0); err != nil {
		return err, chunkWrited, iframeWrited, last
	}

	// создаем заготовки для хидеров
	headersRange, headers := make(map[string]string, 1), make(map[string]string, 0)

	// параметры ключа
	var keyData []byte
	keyPosition := m.KeyTime()

	// в LastIndex мы передвигаемся до конца дескриптора
	lastChunk, chunkKey, lastIFrame, iframeKey := hedx.LastIndexesWithKeys(indexFD)

	// разбираемся с ключами
	if vmx != nil {
		// если мы только открыли index и там нет ключей
		if chunkKey.Type == hedx.TypeInvalid && (m.iframes == nil || iframeKey.Type == hedx.TypeInvalid) {
			// получаем их
			newKeyData, newKeyPosition, err := vmx.GetKeyPosition(resource, keys.ResourceTypeDTV, m.KeyTime())
			if err != nil {
				return err, chunkWrited, iframeWrited, last
			}
			keyData, keyPosition = newKeyData, newKeyPosition
			chunkKey.ChunkKey(0, keyPosition, m.chunks[0].BeginAt-float64(m.beginAt))
			if err := chunkKey.Write(indexFD); err != nil {
				return err, chunkWrited, iframeWrited, last
			}
			if m.iframes != nil {
				iframeKey.IFrameKey(0, keyPosition, m.iframes[0].BeginAt-float64(m.beginAt))
				if err := iframeKey.Write(indexFD); err != nil {
					return err, chunkWrited, iframeWrited, last
				}
			}
		} else {
			// если в chunkKey и в iframeKey лежат какие-то данные, попробуем получить keyData
			newKeyData, newKeyPosition, err := vmx.GetKeyPosition(resource, keys.ResourceTypeDTV, int64(chunkKey.SizeBytes))
			if err != nil {
				return err, chunkWrited, iframeWrited, last
			}
			keyData, keyPosition = newKeyData, newKeyPosition
			if uint64(newKeyPosition) != chunkKey.SizeBytes || chunkKey.SizeBytes != iframeKey.SizeBytes {
				// нужно записать новые ключи
				chunkKey.ChunkKey(0, newKeyPosition, m.chunks[0].BeginAt-float64(m.beginAt))
				if err := chunkKey.Write(indexFD); err != nil {
					return err, chunkWrited, iframeWrited, last
				}
				if m.iframes != nil {
					iframeKey.IFrameKey(0, newKeyPosition, m.iframes[0].BeginAt-float64(m.beginAt))
					if err := iframeKey.Write(indexFD); err != nil {
						return err, chunkWrited, iframeWrited, last
					}
				}
			}
		}
	}

	stat, err := chunkFD.Stat()
	if err != nil {
		return err, chunkWrited, iframeWrited, last
	}
	chunkOffset := stat.Size()

	chunkLength := len(m.chunks)

	m.findAndSaveAbnormal(storageDir)

	lastChunkTs, isFirstChunk := float64(m.beginAt)+lastChunk.TimeStampInSec(), (lastChunk.TimeStampUsec == 0)

	// проверяем что дописать по chunks
	for i, s := range m.chunks {

		// определяем что чанк необходимо дописать
		if round(s.BeginAt) > round(lastChunkTs+0.1) ||
			// если первый чанк в минуте
			(isFirstChunk && int64(s.BeginAt)+1 > int64(lastChunkTs)) {
			lastChunkTs, isFirstChunk = s.BeginAt, false
			index := &hedx.Index{}
			if s.ByteRange != nil {
				headersRange["Range"] = s.ByteRange.Range()
				headers = headersRange
			}
			http, err := fetchURLWithRetry(s.URL, headers, 3)
			if err != nil {
				return err, chunkWrited, iframeWrited, last
			}

			// если небходимо пишем с шифрованием
			startAt := time.Now()
			var writeSize int64
			if vmx != nil {
				writeSize, err = vmx.Crypto(http, chunkFD, keyPosition, keyData)
			} else {
				writeSize, err = io.Copy(chunkFD, http)
			}
			channelInfo.AddWrite(writeSize)
			channelInfo.Data.AddTime(time.Now().Sub(startAt).Seconds())

			if err != nil {
				log.Printf("[ERROR] при записи/шифровании %s: %s\n", s.ToString(), err.Error())
				return err, chunkWrited, iframeWrited, last
			}

			// обычный CHUNK
			index.Chunk(chunkOffset, writeSize, s.BeginAt-float64(m.beginAt))
			if err := index.Write(indexFD); err != nil {
				return err, chunkWrited, iframeWrited, last
			}
			if err := indexFD.Sync(); err != nil {
				return err, chunkWrited, iframeWrited, last
			}
			// записываем CHUNK OEF
			if m.full && i == chunkLength-1 {
				index.ChunkEOF(chunkOffset+writeSize, 0, s.EndAt-float64(m.beginAt))
				if err := index.Write(indexFD); err != nil {
					return err, chunkWrited, iframeWrited, last
				}
				if err := indexFD.Sync(); err != nil {
					return err, chunkWrited, iframeWrited, last
				}
				chunkWrited++
			}
			headers = nil
			chunkOffset = chunkOffset + writeSize
			http.Close()
			chunkWrited++
			if s.EndAt > last {
				last = s.EndAt
			}
		}
	}

	// записываем iframes
	if m.iframes != nil {
		iframeFD, err := os.OpenFile(iframeFile, os.O_RDWR|os.O_CREATE|syscall.O_APPEND, 0644)
		if err != nil {
			return err, chunkWrited, iframeWrited, last
		}
		defer iframeFD.Close()
		stat, err = iframeFD.Stat()
		if err != nil {
			return err, chunkWrited, iframeWrited, last
		}
		iframeOffset := stat.Size()

		iframeLength := len(m.iframes)
		// проверяем что дописать по iframes
		for i, s := range m.iframes {
			if round(s.BeginAt-float64(m.beginAt)) > round(lastIFrame.TimeStampInSec()+0.1) {
				index := &hedx.Index{}
				if s.ByteRange != nil {
					headersRange["Range"] = s.ByteRange.Range()
					headers = headersRange
				}
				http, err := fetchURLWithRetry(s.URL, headers, 3)
				if err != nil {
					return err, chunkWrited, iframeWrited, last
				}

				startAt := time.Now()
				var writeSize int64
				// если небходимо пишем с шифрованием
				if vmx != nil {
					writeSize, err = vmx.Crypto(http, iframeFD, keyPosition, keyData)
				} else {
					writeSize, err = io.Copy(iframeFD, http)
				}
				channelInfo.AddWrite(writeSize)
				channelInfo.Data.AddTime(time.Now().Sub(startAt).Seconds())

				if err != nil {
					log.Printf("[ERROR] при шифровании %s: %s\n", s.ToString(), err.Error())
					return err, chunkWrited, iframeWrited, last
				}
				// обычный IFRAME
				index.IFrame(iframeOffset, writeSize, s.BeginAt-float64(m.beginAt))
				if err := index.Write(indexFD); err != nil {
					return err, chunkWrited, iframeWrited, last
				}
				if err := indexFD.Sync(); err != nil {
					return err, chunkWrited, iframeWrited, last
				}
				// записываем IFRAME OEF
				if m.full && i == iframeLength-1 {
					index.IFrameEOF(iframeOffset+writeSize, 0, s.BeginAt-float64(m.beginAt))
					if err := index.Write(indexFD); err != nil {
						return err, chunkWrited, iframeWrited, last
					}
					if err := indexFD.Sync(); err != nil {
						return err, chunkWrited, iframeWrited, last
					}
					iframeWrited++
				}
				headers = nil
				iframeOffset = iframeOffset + writeSize
				http.Close()
				iframeWrited++
				if s.EndAt > last {
					last = s.EndAt
				}
			}
		}
		if err := iframeFD.Sync(); err != nil {
			return err, chunkWrited, iframeWrited, last
		}
	}

	// синкаем все
	if err := chunkFD.Sync(); err != nil {
		return err, chunkWrited, iframeWrited, last
	}
	if err := indexFD.Sync(); err != nil {
		return err, chunkWrited, iframeWrited, last
	}

	channelInfo.Lag = float64(time.Now().UnixNano())/float64(time.Second) - last

	return nil, chunkWrited, iframeWrited, last

}
