package vmx

import (
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"
)

type resourceType string

const (
	ResourceTypeDTV resourceType = `DTV`
	ResourceTypeVOD resourceType = `VOD`
)

type VMX struct {
	sync.RWMutex
	url                    *url.URL
	keyByURL               map[string][]byte // полный урл - ключ
	lastKeyByResource      map[string][]byte // fallback если не можем получить для конкретного ресурса
	lastPositionByResource map[string]int64
}

func (v *VMX) getUrl() url.URL {
	v.Lock()
	defer v.Unlock()
	result := *v.url
	return result
}

// https://vmxott.svc.iptv.rt.ru/CAB/keyfile
func New(rawurl string) *VMX {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	result := &VMX{
		url:                    u,
		keyByURL:               make(map[string][]byte),
		lastKeyByResource:      make(map[string][]byte),
		lastPositionByResource: make(map[string]int64),
	}
	go result.junitor()
	return result
}

func (k *VMX) junitor() {
	for {
		k.Lock()
		if len(k.keyByURL) > 1024 {
			k.keyByURL = make(map[string][]byte)
			// ресурсов не так много по ожиданиям (кол-во каналов, так что их не чистим)
		}
		k.Unlock()
		time.Sleep(time.Minute)
	}
}

func (k *VMX) Crypto(ioR io.Reader, ioW io.Writer, p int64, key []byte) (int64, error) {
	readCount, writeCount, err := crypto(ioW, ioR, p, 16*188, key)
	if err != nil {
		return int64(writeCount), err
	}
	// aes выровнено по 16 байт, 188/16 =11.75
	// значит прочитать мы должны меньше чем записать
	if readCount > writeCount {
		return int64(writeCount), fmt.Errorf("записано %d из прочитаного %d", writeCount, readCount)
	}
	if !(readCount%188 == 0) {
		return int64(writeCount), fmt.Errorf("прочитаное значение %d не выровнено по 188 байт", readCount)
	}
	if !(writeCount%16 == 0) {
		return int64(writeCount), fmt.Errorf("записанное значение %d не выровнено по 16 байт", readCount)
	}
	return int64(writeCount), nil
}
