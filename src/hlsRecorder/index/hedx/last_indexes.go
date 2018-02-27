package hedx

import (
	"bufio"
	"io"
)

func LastIndexes(r io.Reader) (chunk *Index, iframe *Index) {
	reader := bufio.NewReaderSize(r, 1024)
	chunk, iframe = &Index{}, &Index{}
	for {
		i := &Index{}
		if err := i.Read(reader); err != nil {
			if err == io.EOF {
				break
			}
		}
		if i.IsInFrame() {
			iframe = i
		} else {
			chunk = i
		}
	}
	return
}

func LastIndexesWithKeys(r io.Reader) (chunk *Index, chunkKey *Index, iframe *Index, iframeKey *Index) {
	reader := bufio.NewReaderSize(r, 1024)
	chunk, iframe, iframeKey, chunkKey = &Index{}, &Index{}, &Index{}, &Index{}
	for {
		i := &Index{}
		if err := i.Read(reader); err != nil {
			if err == io.EOF {
				break
			}
		}
		if i.IsInFrame() {
			if i.Type == TypeKey {
				iframeKey = i
			} else {
				iframe = i
			}
		} else {
			if i.Type == TypeKey {
				chunkKey = i
			} else {
				chunk = i
			}
		}
	}
	return
}
