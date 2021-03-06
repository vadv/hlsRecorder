package writer

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	hedx "hlsRecorder/index/hedx"
	keys "hlsRecorder/keys"
	stat "hlsRecorder/stat"
)

// переписываем файлы на диске полностью
func (m *minute) writeFull(indexDir, storageDir, resource string, vmx *keys.VMX, channelInfo *stat.ChannelInfo) error {

	if len(m.chunks) == 0 {
		return fmt.Errorf("пустая минутка")
	}

	chunkFile, iframeFile, indexFile := m.getPath(storageDir, `.ets`), m.getPath(storageDir, `.ets.ifr`), m.getPath(indexDir, `.ets.hedx`)

	if err := os.MkdirAll(filepath.Dir(chunkFile), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(indexFile), 0755); err != nil {
		return err
	}

	chunkFD, err := os.OpenFile(chunkFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer chunkFD.Close()
	if _, err := chunkFD.Seek(0, 0); err != nil {
		return err
	}

	indexFD, err := os.OpenFile(indexFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer indexFD.Close()
	if _, err := indexFD.Seek(0, 0); err != nil {
		return err
	}

	index := &hedx.Index{}
	headersRange, headers := make(map[string]string, 1), make(map[string]string, 0)

	// если шифруем, то должны записать ключи, сначала получаем их:
	var keyData []byte
	var keyPosition int64
	if vmx != nil {
		keyData, keyPosition, err = vmx.GetKeyPosition(resource, keys.ResourceTypeDTV, m.KeyTime())
		if err != nil {
			return err
		}
		index.ChunkKey(0, keyPosition, m.chunks[0].BeginAt-float64(m.beginAt))
		if err := index.Write(indexFD); err != nil {
			return err
		}
		if m.iframes != nil {
			index.IFrameKey(0, keyPosition, m.iframes[0].BeginAt-float64(m.beginAt))
			if err := index.Write(indexFD); err != nil {
				return err
			}
		}
	}

	// сначала записываем chunks
	chunkLength, chunkOffset := len(m.chunks), int64(0)
	for i, s := range m.chunks {
		if s.ByteRange != nil {
			headersRange["Range"] = s.ByteRange.Range()
			headers = headersRange
		}
		http, err := fetchURLWithRetry(s.URL, headers, 3)
		if err != nil {
			return err
		}

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
			log.Printf("[ERROR] при шифровании %s: %s\n", s.ToString(), err.Error())
			return err
		}

		// обычный CHUNK
		index.Chunk(chunkOffset, writeSize, s.BeginAt-float64(m.beginAt))
		if err := index.Write(indexFD); err != nil {
			return err
		}
		if err := indexFD.Sync(); err != nil {
			return err
		}
		if i == chunkLength-1 {
			// записываем CHUNK OEF
			index.ChunkEOF(chunkOffset+writeSize, 0, s.EndAt-float64(m.beginAt))
			if err := index.Write(indexFD); err != nil {
				return err
			}
			if err := indexFD.Sync(); err != nil {
				return err
			}
		}
		headers = nil
		chunkOffset = chunkOffset + writeSize
		http.Close()
	}

	// потом записываем iframes
	if m.iframes != nil {
		iframeFD, err := os.OpenFile(iframeFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer iframeFD.Close()
		if _, err := iframeFD.Seek(0, 0); err != nil {
			return err
		}
		iframeLength, iframeOffset := len(m.iframes), int64(0)
		for i, s := range m.iframes {
			if s.ByteRange != nil {
				headersRange["Range"] = s.ByteRange.Range()
				headers = headersRange
			}
			http, err := fetchURLWithRetry(s.URL, headers, 3)
			if err != nil {
				return err
			}

			// если небходимо пишем с шифрованием
			startAt := time.Now()
			var writeSize int64
			if vmx != nil {
				writeSize, err = vmx.Crypto(http, iframeFD, keyPosition, keyData)
			} else {
				writeSize, err = io.Copy(iframeFD, http)
			}
			channelInfo.AddWrite(writeSize)
			channelInfo.Data.AddTime(time.Now().Sub(startAt).Seconds())

			if err != nil {
				log.Printf("[ERROR] при шифровании %s: %s\n", s.ToString(), err.Error())
				return err
			}
			// записываем IFRAME
			index.IFrame(iframeOffset, writeSize, s.BeginAt-float64(m.beginAt))
			if err := index.Write(indexFD); err != nil {
				return err
			}
			if err := indexFD.Sync(); err != nil {
				return err
			}
			if i == iframeLength-1 {
				// записываем IFRAME OEF
				index.IFrameEOF(iframeOffset+writeSize, 0, s.EndAt-float64(m.beginAt))
				if err := index.Write(indexFD); err != nil {
					return err
				}
				if err := indexFD.Sync(); err != nil {
					return err
				}
			}
			headers = nil
			iframeOffset = iframeOffset + writeSize
			http.Close()
		}

		if err := iframeFD.Sync(); err != nil {
			return err
		}
	}

	// синкаем все
	if err := chunkFD.Sync(); err != nil {
		return err
	}
	if err := indexFD.Sync(); err != nil {
		return err
	}

	return nil
}
