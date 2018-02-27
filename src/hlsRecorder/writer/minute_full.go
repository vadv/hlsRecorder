package writer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	hedx "hlsRecorder/index/hedx"
	keys "hlsRecorder/keys"
)

// переписываем файлы на диске полностью
func (m *minute) writeFull(indexDir, storageDir, resource string, vmx *keys.VMX) error {

	if len(m.iframes) == 0 || len(m.chunks) == 0 {
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

	iframeFD, err := os.OpenFile(iframeFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer iframeFD.Close()
	if _, err := iframeFD.Seek(0, 0); err != nil {
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

	// должны записать ключи, сначала получаем их
	keyData, keyPosition, err := vmx.GetKeyPosition(resource, keys.ResourceTypeDTV, m.beginAt)
	if err != nil {
		return err
	}
	index.ChunkKey(0, keyPosition, m.chunks[0].BeginAt-float64(m.beginAt))
	if err := index.Write(indexFD); err != nil {
		return err
	}
	index.IFrameKey(0, keyPosition, m.iframes[0].BeginAt-float64(m.beginAt))
	if err := index.Write(indexFD); err != nil {
		return err
	}

	// сначала записываем chunks
	chunkLength, chunkOffset := len(m.chunks), int64(0)
	for i, s := range m.chunks {
		if s.ByteRange != nil {
			headersRange["Range"] = fmt.Sprintf(
				"bytes=%d-%d", s.ByteRange.Offset, s.ByteRange.Offset+s.ByteRange.Length-1)
			headers = headersRange
		}
		http, err := fetchURLWithRetry(s.URL, headers, 3)
		if err != nil {
			return err
		}
		//writeSize, err := io.Copy(chunkFD, http)
		writeSize, err := vmx.Crypto(http, chunkFD, keyPosition, keyData)
		if err != nil {
			log.Printf("[ERROR] при шифровании %s: %s\n", s.ToString(), err.Error())
			return err
		}
		if i == chunkLength-1 {
			// записываем CHUNK OEF
			index.ChunkEOF(chunkOffset+writeSize, 0, s.BeginAt-float64(m.beginAt))
		} else {
			// обычный CHUNK
			index.Chunk(chunkOffset, writeSize, s.BeginAt-float64(m.beginAt))
		}
		if err := index.Write(indexFD); err != nil {
			return err
		}
		if err := indexFD.Sync(); err != nil {
			return err
		}
		headers = nil
		chunkOffset = chunkOffset + writeSize
		http.Close()
	}

	// потом записываем iframes
	iframeLength, iframeOffset := len(m.iframes), int64(0)
	for i, s := range m.iframes {
		if s.ByteRange != nil {
			headersRange["Range"] = fmt.Sprintf(
				"bytes=%d-%d", s.ByteRange.Offset, s.ByteRange.Offset+s.ByteRange.Length-1)
			headers = headersRange
		}
		http, err := fetchURLWithRetry(s.URL, headers, 3)
		if err != nil {
			return err
		}
		//writeSize, err := io.Copy(iframeFD, http)
		writeSize, err := vmx.Crypto(http, iframeFD, keyPosition, keyData)
		if err != nil {
			log.Printf("[ERROR] при шифровании %s: %s\n", s.ToString(), err.Error())
			return err
		}
		if i == iframeLength-1 {
			// записываем IFRAME OEF
			index.IFrameEOF(iframeOffset+writeSize, 0, s.BeginAt-float64(m.beginAt))
		} else {
			// записываем IFRAME
			index.IFrame(iframeOffset, writeSize, s.BeginAt-float64(m.beginAt))
		}
		if err := index.Write(indexFD); err != nil {
			return err
		}
		if err := indexFD.Sync(); err != nil {
			return err
		}
		headers = nil
		iframeOffset = iframeOffset + writeSize
		http.Close()
	}

	// синкаем все
	if err := chunkFD.Sync(); err != nil {
		return err
	}
	if err := iframeFD.Sync(); err != nil {
		return err
	}
	if err := indexFD.Sync(); err != nil {
		return err
	}

	return nil
}
