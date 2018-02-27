package parser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var reKeyValue = regexp.MustCompile(`([a-zA-Z0-9_-]+)=("[^"]+"|[^",]+)`)

func parseKeyValue(line string) map[string]string {
	out := make(map[string]string, 0)
	for _, kv := range reKeyValue.FindAllStringSubmatch(line, -1) {
		k, v := kv[1], kv[2]
		out[k] = strings.Trim(v, ` "`)
	}
	return out
}

func ParseIndex(r io.ReadCloser) (*Index, error) {
	defer r.Close()
	result, scanner := &Index{Streams: make([]*Stream, 0)}, bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}
		if strings.HasPrefix(line, `#`) {
			// #
			kv := parseKeyValue(line)
			switch {

			case strings.HasPrefix(line, `#EXT-X-I-FRAME-STREAM-INF`):
				uri, ok := kv["URI"]
				if !ok {
					return result, fmt.Errorf("can't found uri in stream info line: `%s`", line)
				}
				if len(result.Streams) == 0 {
					result.Streams = append(result.Streams, &Stream{IFrameURI: uri})
					continue
				}
				if lastStream := result.Streams[len(result.Streams)-1]; lastStream.IFrameURI == `` {
					lastStream.IFrameURI = uri
					continue
				}
				result.Streams = append(result.Streams, &Stream{IFrameURI: uri})

			case strings.HasPrefix(line, `#EXT-X-STREAM-INF`):
				bandwidth, ok := kv["BANDWIDTH"]
				if !ok {
					return result, fmt.Errorf("can't found bandwidth in stream info line: `%s`", line)
				}
				if len(result.Streams) == 0 {
					result.Streams = append(result.Streams, &Stream{MainBandwidth: bandwidth})
					continue
				}
				if lastStream := result.Streams[len(result.Streams)-1]; lastStream.MainBandwidth == `` {
					lastStream.MainBandwidth = bandwidth
					continue
				}
				result.Streams = append(result.Streams, &Stream{MainBandwidth: bandwidth})
			}
		} else {
			// main url
			if len(result.Streams) == 0 {
				result.Streams = append(result.Streams, &Stream{MainURI: line})
				continue
			}
			if lastStream := result.Streams[len(result.Streams)-1]; lastStream.MainURI == `` {
				lastStream.MainURI = line
				continue
			}
			result.Streams = append(result.Streams, &Stream{MainURI: line})
		}
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}
	return result, nil
}
