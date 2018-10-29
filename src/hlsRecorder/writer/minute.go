package writer

import (
	"fmt"
	hedx "hlsRecorder/index/hedx"
	parser "hlsRecorder/parser"
	"os"
	"path/filepath"
)

// сущность - набор сегментом которые необходимо записать на диск
type minute struct {
	beginAt       int64
	full          bool
	chunks        []*parser.Segment
	iframes       []*parser.Segment
	chunkPlayList *parser.PlayList
}

// важная функция - по сути прибитый гвозядями путь на диске
// который выглядит как <data_dir>/%h/%m.fullExt
func (m *minute) getPath(dataPath string, fullExt string) string {
	return filepath.Join(
		dataPath,
		fmt.Sprintf("%d", m.hour()),
		fmt.Sprintf("%d%s", m.beginAt, fullExt),
	)
}

func getMinute(t float64) int64 {
	return int64(t) - (int64(t) % 60.0)
}

func newMinute(t float64) *minute {
	return &minute{
		beginAt: getMinute(t),
		chunks:  make([]*parser.Segment, 0),
		iframes: nil,
	}
}

func (m *minute) hour() int64 {
	return int64(m.beginAt) - (int64(m.beginAt) % 3600.0)
}

func (m *minute) indexExists(indexDir string) bool {
	_, err := os.Stat(m.getPath(indexDir, `.ets.hedx`))
	return err == nil // os.IsNotExists, directory
}

// проверяем записана ли на диск минута
func (m *minute) indexHasEOF(indexDir string) bool {
	if !m.indexExists(indexDir) {
		return false
	}
	fd, err := os.Open(m.getPath(indexDir, `.ets.hedx`))
	if err != nil {
		return false
	}
	defer fd.Close()
	return hedx.HaveEOF(fd)
}

// ротейшен полиси ключа
func (m *minute) KeyTime() int64 {
	return (m.beginAt - (m.beginAt % (60 * 60 * 24))) // раз в день
}
