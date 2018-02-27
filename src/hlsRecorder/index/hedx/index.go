package hedx

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// FileExt зафиксированное окончание файла индекса
	FileExt = `.hedx`
	// TypeInvalid невалидная запись
	TypeInvalid = 0
	// TypeChunk информация о данных
	TypeChunk = 1
	// TypeIframe iframe в данных
	TypeIframe = 2
	// TypeKey информация о ключе для данных, которые предоставлены ниже
	TypeKey = 3
	// TypeEOF конец файла или chunk
	TypeEOF = 4
	// TypeDiscontinuity discontinuity, который вставил smartrecord
	TypeDiscontinuity = 5
	// TypeNoSignal - если в промежуток определенного времени smartrecord не получал сигнал
	TypeNoSignal       = 6
	inIframeFlag uint8 = 0x01
)

// Index представляет структурку для парсинга с диска при помощи binary.Read
type Index struct {
	Type          uint8
	OffsetBytes   uint64
	SizeBytes     uint64
	TimeStampUsec uint64
	Flags         uint8
	Reserved      uint32
}

// TimeStampInSec возвращает время в секундах с момента начала файла
func (i *Index) TimeStampInSec() float64 {
	return (float64(i.TimeStampUsec) / 1000000.0)
}

// Read читает и декодирует содержимое
func (i *Index) Read(r io.Reader) error {
	buf := make([]byte, 1+8+8+8+1+4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	i.Type = uint8(buf[0])
	i.OffsetBytes = uint64(binary.LittleEndian.Uint64(buf[1:9]))
	i.SizeBytes = uint64(binary.LittleEndian.Uint64(buf[9:17]))
	i.TimeStampUsec = uint64(binary.LittleEndian.Uint64(buf[17:25]))
	i.Flags = uint8(buf[25])
	i.Reserved = uint32(binary.LittleEndian.Uint32(buf[26:30]))
	return nil
}

// Write записывает текущее содержимое
func (i *Index) Write(w io.Writer) error {
	buf := make([]byte, 1+8+8+8+1+4)
	buf[0] = i.Type
	binary.LittleEndian.PutUint64(buf[1:9], i.OffsetBytes)
	binary.LittleEndian.PutUint64(buf[9:17], i.SizeBytes)
	binary.LittleEndian.PutUint64(buf[17:25], i.TimeStampUsec)
	buf[25] = i.Flags
	binary.LittleEndian.PutUint32(buf[26:30], i.Reserved)
	count, err := w.Write(buf)
	if err != nil {
		return err
	}
	if count != len(buf) {
		return fmt.Errorf("only %d of %d was written", count, len(buf))
	}
	return nil
}

func (i *Index) IsInFrame() bool {
	return 0 != (i.Flags & inIframeFlag)
}
