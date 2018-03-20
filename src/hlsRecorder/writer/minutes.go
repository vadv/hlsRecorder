package writer

import (
	"fmt"
	"sort"

	parser "hlsRecorder/parser"
)

type minutes struct {
	list map[int64]*minute
}

// возвращаем список всех полных минуток из сегмента
func makeMinutes(chunks, iframes *parser.PlayList) (*minutes, error) {

	if len(chunks.Segments) == 0 {
		return nil, fmt.Errorf("список chunk-сегментов пустой")
	}

	list := make(map[int64]*minute, 0)

	for _, segment := range chunks.Segments {
		at := getMinute(segment.BeginAt)
		if _, ok := list[at]; !ok {
			list[at] = newMinute(segment.BeginAt)
		}
		list[at].chunks = append(list[at].chunks, segment)
	}

	if iframes != nil {
		for _, segment := range iframes.Segments {
			at := getMinute(segment.BeginAt)
			if list[at].iframes == nil {
				list[at].iframes = make([]*parser.Segment, 0)
			}
			if _, ok := list[at]; !ok {
				// iframe-плейлист обгоняет/не догоняет chunks
				continue
			}
			list[at].iframes = append(list[at].iframes, segment)
		}
	}

	// первая минута всегда в непонятном статусе,
	// поэтому мы просто ее удаляем
	delete(list, getMinute(chunks.Segments[0].BeginAt))
	if len(list) == 0 {
		return nil, fmt.Errorf("не одной целой минуты")
	}

	for _, m := range list {
		chunkFull, iframeFull := false, false
		for _, segment := range m.chunks {
			if int64(segment.EndAt) >= m.beginAt+60 {
				chunkFull = true
				break
			}
		}
		if iframes == nil {
			iframeFull = true
		} else {
			for _, segment := range m.iframes {
				if int64(segment.EndAt) >= m.beginAt+60 {
					iframeFull = true
					break
				}
			}
		}
		m.chunkPlayList = chunks
		m.full = chunkFull && iframeFull
	}

	return &minutes{list: list}, nil
}

func (m *minutes) last() (last *minute) {
	for _, next := range m.list {
		if last == nil || next.beginAt > last.beginAt {
			last = next
		}
	}
	return
}

func (m *minutes) sortedMinuteList() []int64 {
	result := make([]int64, 0)
	for i, _ := range m.list {
		result = append(result, i)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
