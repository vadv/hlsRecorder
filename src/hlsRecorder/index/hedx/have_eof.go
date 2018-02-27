package hedx

import (
	"bufio"
	"io"
)

func HaveEOF(r io.Reader) bool {

	i := &Index{}
	reader := bufio.NewReaderSize(r, 32*1024)
	haveChunkEOF, haveIFrameEOF := false, false

	for {
		if err := i.Read(reader); err != nil {
			if err == io.EOF {
				break
			}
			return false
		}
		if i.Type == TypeEOF {
			if i.IsInFrame() {
				haveIFrameEOF = true
			} else {
				haveChunkEOF = true
			}
		}
	}

	return haveChunkEOF && haveIFrameEOF

}
