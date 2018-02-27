package writer

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

var httpTimeout = time.Duration(15 * time.Second)
var httpTransport = &http.Transport{
	Dial:                (&net.Dialer{Timeout: httpTimeout}).Dial,
	TLSHandshakeTimeout: httpTimeout,
	TLSClientConfig:     &tls.Config{},
	MaxIdleConnsPerHost: 10,
}

func fetchURL(url string, headers map[string]string) (io.ReadCloser, error) {

	client := http.Client{
		Transport: httpTransport,
		Timeout:   httpTimeout,
	}

	request, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return nil, err
	}

	if len(headers) > 0 {
		for k, v := range headers {
			request.Header.Set(k, v)
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK && len(headers) == 0 {
		response.Body.Close()
		return nil, fmt.Errorf("не корректный http-ответ: %d", response.StatusCode)
	}

	if response.StatusCode != http.StatusPartialContent && len(headers) > 1 {
		response.Body.Close()
		return nil, fmt.Errorf("не корректный http-ответ: %d", response.StatusCode)
	}

	return response.Body, nil
}

func fetchURLWithRetry(url string, headers map[string]string, retryCount int) (io.ReadCloser, error) {
	for i := 0; i < retryCount; i++ {
		r, err := fetchURL(url, headers)
		if err == nil {
			return r, err
		}
		log.Printf("[ERROR] скачивание с %s завершилось с ошибкой: %s\n", url, err.Error())
	}
	return nil, fmt.Errorf("достигнуто максимальное количество ошибок при скачивании")
}
