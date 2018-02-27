package vmx

import (
	"fmt"
	"io"
)

func (k *VMX) getUnsafeKeyPosition(r string) ([]byte, int64, error) {
	key, okKey := k.lastKeyByResource[r]
	p, okPosition := k.lastPositionByResource[r]
	if okKey && okPosition {
		return key, p, nil
	}
	return nil, 0, fmt.Errorf("ключ не найден в предыдущих позициях")
}

// скачивает и возвращает ключ
func (k *VMX) GetKeyPosition(r string, t resourceType, p int64) ([]byte, int64, error) {

	u := k.url
	q := u.Query()
	q.Set(`r`, r)
	q.Set(`p`, fmt.Sprintf("%d", p))
	q.Set(`t`, string(t))
	u.RawQuery = q.Encode()

	// блокируем на чтение
	k.Lock()
	// нашлось в кэше
	result, ok := k.keyByURL[u.String()]
	k.Unlock()
	if ok {
		return result, p, nil
	}

	// должны сходить по http и дернуть ключ
	http, err := fetchURLWithRetry(u.String(), 3)
	if err != nil {
		// мы не смогли получить ключ, пробуем достать по последнему ресурсу
		k.Lock()
		result, p, err := k.getUnsafeKeyPosition(r)
		k.Unlock()
		return result, p, err
	}
	defer http.Close()

	buff := make([]byte, 16)
	size, err := io.ReadFull(http, buff)
	if err != nil || size != len(buff) {
		// мы не смогли получить ключ, пробуем достать по последнему ресурсу
		k.Lock()
		result, p, err := k.getUnsafeKeyPosition(r)
		k.Unlock()
		return result, p, err
	}

	k.RLock()
	k.keyByURL[u.String()] = buff
	k.lastKeyByResource[r] = buff
	k.lastPositionByResource[r] = p
	k.RUnlock()

	return buff, p, nil
}
