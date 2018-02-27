package vmx

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

var httpTimeout = time.Duration(2 * time.Second)
var httpTransport = &http.Transport{
	Dial:                (&net.Dialer{Timeout: httpTimeout}).Dial,
	TLSHandshakeTimeout: httpTimeout,
	TLSClientConfig:     &tls.Config{},
	MaxIdleConnsPerHost: 5,
}

func fetchURL(url string) (io.ReadCloser, error) {

	client := http.Client{
		Transport: httpTransport,
		Timeout:   httpTimeout,
	}

	request, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("не корректный http-ответ: %d", response.StatusCode)
	}

	return response.Body, nil
}

func fetchURLWithRetry(url string, retryCount int) (io.ReadCloser, error) {
	for i := 0; i < retryCount; i++ {
		r, err := fetchURL(url)
		if err == nil {
			return r, err
		}
		log.Printf("[ERROR] скачивание с %s завершилось с ошибкой: %s\n", url, err.Error())
	}
	return nil, fmt.Errorf("достигнуто максимальное количество ошибок при скачивании")
}
