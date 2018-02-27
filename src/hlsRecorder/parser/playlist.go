package parser

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func timeToFloat64(t time.Time) float64 {
	return float64(t.UnixNano()) / float64(time.Second)
}

func timeParse(value string) (float64, error) {
	layouts := []string{
		"2006-01-02T15:04:05.999999999Z0700",
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05.999999999Z07",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return timeToFloat64(t), nil
		}
	}
	return 0, fmt.Errorf("can't found valid layout")
}

// "/.../xx-yy.ts"
func parseStartEndToFloat(value string) (start float64, end float64, err error) {
	value = filepath.Base(value)
	value = strings.TrimSuffix(value, filepath.Ext(value))
	data := strings.Split(value, `-`)
	if len(data) != 2 {
		err = fmt.Errorf("bad format")
		return
	}
	start, err = strconv.ParseFloat(data[0], 64)
	if err != nil {
		return
	}
	end, err = strconv.ParseFloat(data[1], 64)
	return
}

func ParsePlayList(r io.ReadCloser) (*PlayList, error) {

	defer r.Close()

	result, scanner := PlayList{Segments: make([]*Segment, 0)}, bufio.NewScanner(r)

	var segmentByteRangeInfo *ByteRange
	var segmentBeginAt, segmentDuration float64

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}

		switch {

		case strings.HasPrefix(line, `#EXT-X-MEDIA-SEQUENCE:`):
			if _, err := fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", &result.MediaSeq); err != nil {
				return &result, fmt.Errorf("parse media sequence error `%s` from line `%s`", err.Error(), line)
			}

		case strings.HasPrefix(line, `#EXT-X-PROGRAM-DATE-TIME:`):
			programDate, err := timeParse(line[25:])
			if err != nil {
				return &result, fmt.Errorf("parse program date error `%s` from line `%s`", err.Error(), line)
			}
			segmentBeginAt = programDate

		case strings.HasPrefix(line, `#EXTINF:`):
			sepIndex := strings.Index(line, ",")
			if sepIndex == -1 {
				sepIndex = len(line)
			}
			duration, err := strconv.ParseFloat(line[8:sepIndex], 64)
			if err != nil {
				return &result, fmt.Errorf("parse extinf error `%s` from line `%s`", err.Error(), line)
			}
			segmentDuration = duration

		case strings.HasPrefix(line, `#EXT-X-BYTERANGE:`):
			params, byteRange := strings.SplitN(line[17:], "@", 2), &ByteRange{}
			if length, err := strconv.ParseInt(params[0], 10, 64); err != nil {
				return &result, fmt.Errorf("parse byterange length error `%s` from line `%s`", err.Error(), line)
			} else {
				byteRange.Length = length
			}
			if len(params) > 1 {
				if offset, err := strconv.ParseInt(params[1], 10, 64); err != nil {
					return &result, fmt.Errorf("parse byterange offset error `%s` from line `%s`", err.Error(), line)
				} else {
					byteRange.Offset = offset
				}
			}
			segmentByteRangeInfo = byteRange

		case strings.HasPrefix(line, "#EXT-X-DISCONTINUITY"):
			segmentBeginAt, segmentDuration, segmentByteRangeInfo = 0, 0, nil

		case strings.HasPrefix(line, "#EXT-X-I-FRAMES-ONLY"):
			result.IFrame = true

		case !strings.HasPrefix(line, "#"):
			segment := &Segment{
				URI:       line,
				Duration:  segmentDuration,
				BeginAt:   segmentBeginAt,
				ByteRange: segmentByteRangeInfo,
			}
			if segment.BeginAt == 0 || segment.Duration == 0 {
				// проверим есть ли у предыдущего окончание
				if length := len(result.Segments); length > 1 {
					last := result.Segments[length-1]
					if last.BeginAt > 0 && last.Duration > 0 {
						segment.BeginAt = last.BeginAt + last.Duration
					}
				}
			}
			if segment.BeginAt == 0 || segment.Duration == 0 {
				// последний шанс исправить ситуацию
				start, end, err := parseStartEndToFloat(segment.URI)
				if err == nil {
					segment.BeginAt = start
					segment.Duration = end - start
				} else {
					return &result, fmt.Errorf("can't found extinf or program date. Parse uri %s: %s", line, err.Error())
				}
			}
			if segment.BeginAt == 0 || segment.Duration == 0 {
				return &result, fmt.Errorf("parse extinf or program date for: %s failed", line)
			}
			segment.EndAt = segment.BeginAt + segment.Duration
			result.Segments = append(result.Segments, segment)
			segmentBeginAt, segmentDuration, segmentByteRangeInfo = 0, 0, nil

		}
	}
	if err := scanner.Err(); err != nil {
		return &result, err
	}
	return &result, nil
}
