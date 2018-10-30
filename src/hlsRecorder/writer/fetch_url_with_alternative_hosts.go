package writer

import (
	"fmt"
	parser "hlsRecorder/parser"
	"io"
	"log"
)

/* закачка url со списком альтернативных источников

suffixUrl: /indextest/0/chunklist.m3u8
stream: {
    Hosts: [http://192.168.185.148, http://192.168.185.149]
    currentHost: http://192.168.185.148
}

возвращаем интерфейс для чтения, ошибку и менялся ли upstream, и конечный url

*/
func fetchStreamUrlWithAlternativeHosts(suffixUrl string, stream *parser.Stream, headers map[string]string, retryCount int) (io.ReadCloser, error, bool, string) {
	stream.Lock()
	if stream.CurrentHost == `` {
		stream.CurrentHost = stream.Hosts[0]
	}
	stream.Unlock()

	url := stream.CurrentHost + suffixUrl
	result, err := fetchURLWithRetry(url, headers, retryCount)
	if err != nil {
		stream.Lock()
		defer stream.Unlock()
		// пробуем переключить источник
		for i := 0; i < len(stream.Hosts); i++ {
			stream.CurrentHost = stream.Hosts[i]
			log.Printf("[WARN] `%s` пробуем альтернативный источник: %s\n", stream.Name(), stream.CurrentHost)
			url = stream.CurrentHost + suffixUrl
			result, err := fetchURLWithRetry(url, headers, retryCount)
			if err == nil {
				log.Printf("[INFO] `%s` выбрали источник: %s\n", stream.Name(), stream.CurrentHost)
				return result, nil, true, url
			}
		}
	} else {
		return result, nil, false, url
	}
	return nil, fmt.Errorf("не найдено ни одного подходящего источника"), false, ``
}
