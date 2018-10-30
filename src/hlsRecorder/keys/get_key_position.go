package vmx

import (
	"fmt"
	"io"
	"log"
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
	k.RLock()
	// нашлось в кэше
	result, ok := k.keyByURL[u.String()]
	k.RUnlock()
	if ok {
		log.Printf("[DEBUG] {vmx:%s:%d} ключ из кэша\n", r, p)
		return result, p, nil
	}

	// должны сходить по http и дернуть ключ
	http, err := fetchURLWithRetry(u.String(), 3)
	if err != nil {
		log.Printf("[ERROR] {vmx:%s:%d): %s\n", r, p, err.Error())
		// мы не смогли получить ключ, пробуем достать по последнему ресурсу
		k.RLock()
		result, newP, err := k.getUnsafeKeyPosition(r)
		k.RUnlock()
		log.Printf("[INFO] {vmx:%s:%d}: воспользовались резервным ключом: (%s:%d)\n", r, p, r, newP)
		return result, newP, err
	}
	defer http.Close()

	buff := make([]byte, 16)
	size, err := io.ReadFull(http, buff)
	if err == nil && size != len(buff) {
		err = fmt.Errorf("неправильный размер при скачивании: %d ожидалось: %s", size, len(buff))
	}
	if err != nil {
		log.Printf("[ERROR] {vmx:%s:%d}: %s\n", r, p, err.Error())
		// мы не смогли получить ключ, пробуем достать по последнему ресурсу
		k.RLock()
		result, newP, err := k.getUnsafeKeyPosition(r)
		k.RUnlock()
		log.Printf("[INFO] {vmx:%s:%d}: воспользовались резервным ключом: (%s:%d)\n", r, p, r, newP)
		return result, newP, err
	}

	k.Lock()
	k.keyByURL[u.String()] = buff
	k.lastKeyByResource[r] = buff
	k.lastPositionByResource[r] = p
	log.Printf("[DEBUG] {vmx:%s:%d} установили в кэш\n", r, p)
	k.Unlock()

	return buff, p, nil
}
