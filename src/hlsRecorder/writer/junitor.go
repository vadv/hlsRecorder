package writer

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// передаем корневую директорию из которой надо удалять
// директории и их содержимое, которые похожи на unixts и старее чем now-deleteOlder
func runJunitor(dir string, deleteOlder int64) {
	time.Sleep(time.Minute)
	for {
		list, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Printf("[ERROR] в процессе поиска старых директорий в %s: %s\n", dir, err.Error())
			time.Sleep(time.Minute)
			continue
		}
		for _, d := range list {
			if !d.IsDir() {
				continue
			}
			dateOfDir, err := strconv.ParseUint(d.Name(), 10, 64)
			if err == nil && (time.Now().Unix()-deleteOlder > int64(dateOfDir)) {
				path := filepath.Join(dir, d.Name())
				log.Printf("[INFO] удаляем старую директорию %s\n", path)
				if err := os.RemoveAll(path); err != nil {
					log.Printf("[ERROR] при удалении старой директории %s: %s\n", path, err.Error())
				}
			}
		}
		time.Sleep(time.Hour)
	}
}
