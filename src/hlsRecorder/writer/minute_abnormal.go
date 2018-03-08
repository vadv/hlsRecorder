package writer

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func calcMedian(numbers []float64) float64 {
	middle := len(numbers) / 2
	result := numbers[middle]
	if len(numbers)%2 == 0 {
		result = (result + numbers[middle-1]) / 2
	}
	return result
}

// находим подозрительного размера
func (m *minute) findAndSaveAbnormal(storageDir string) {

	sizes := make([]float64, 0)
	for _, s := range m.chunks {
		sizes = append(sizes, s.Duration)
	}
	median := calcMedian(sizes)

	for _, s := range m.chunks {
		if s.Duration > median*1.5 || s.Duration < (median/1.5) {
			// находим подозрительно большого размера сегмент
			filename := filepath.Join(storageDir, fmt.Sprintf("%d-%s", s.URI, time.Now().Unix()))
			log.Printf("[ERROR] найден странный чанк %s (медиана: %.2f текущий: %.2f) пытаемся сохранить его в %s\n", s.ToString(), median, s.Duration, filename)
			log.Printf("[ERROR] плейлист:\n%s\n", m.chunkPlayList.Body)
			http, err := fetchURL(filename, nil)
			if err == nil {
				defer http.Close()
				fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
				if err == nil {
					defer fd.Close()
					_, err := io.Copy(fd, http)
					if err == nil {
						log.Printf("[ERROR] странный чанк %s сохранен в %s\n", s.ToString(), filename)
					}
				}
			}
		}
	}
}
