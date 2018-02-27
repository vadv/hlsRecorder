package writer

import (
	"context"
	"log"
	"net/url"
	"time"

	keys "hlsRecorder/keys"
	parser "hlsRecorder/parser"
)

func Stream(stream *parser.Stream, ctx context.Context) {

	resource := (ctx.Value(`content.channel`)).(string)
	storageDir := (ctx.Value(`path.storage.dir`)).(string)
	indexDir := (ctx.Value(`path.index.dir`)).(string)
	vmx := (ctx.Value(`keys.vmx`)).(*keys.VMX)

	mainURI := stream.MainURI
	iframeURI := stream.IFrameURI
	mainURL, errMainURL := url.Parse(stream.MainURI)
	iframeURL, errIFrameURL := url.Parse(stream.IFrameURI)
	if errMainURL != nil {
		panic(errMainURL)
	}
	if errIFrameURL != nil {
		panic(errIFrameURL)
	}

	log.Printf("[INFO] {stream: %s} старт процессинга с параметрами: storage=`%s`, index=`%s`\n", mainURI, storageDir, indexDir)

	// запускаем бесконечный тред
	wasStopped := false
	go func() {

		prevChunkMediaSEQ, prevIframeMediaSEQ := int64(-1), int64(-1)
		equalMediaSEQCount := 0

		//var dataOffset indexOffset int64

		// каждый цикл мы должны должны скачать главный плейлист
		// пробежаться по нему сверху до низу:
		// если мы имеем целую минутку в плейлисте
		// проверяем, есть ли у index-файла есть окончание
		// если окончания нет, то проверям не открыты ли дескриптор на закачку
		// и если надо открываем и скачиваем туда данные
		for {

			// обработка главного плейлиста
			r1, err := fetchURL(mainURI, nil)
			if err != nil {
				log.Printf("[ERROR] {stream: %s} в процессе скачивания chunks-плейлиста: %s\n", mainURI, err.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			chunkPL, err := parser.ParsePlayList(r1)
			r1.Close()
			if err != nil {
				log.Printf("[ERROR] {stream: %s} при парсинге chunk-плейлиста: %s\n", mainURI, err.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			chunkPL.SetURL(mainURL)

			/*
				MediaSeq в памяти транскодера - поэтому мы не можем на него орентироваться
					if chunkPL.MediaSeq < prevChunkMediaSEQ {
						log.Printf("[ERROR] {stream: %s} текущий media sequence меньше чем предыдущий: текущий=%d предыдущий=%d\n", mainURI, chunkPL.MediaSeq, prevChunkMediaSEQ)
						time.Sleep(5 * time.Second)
						continue
					}
			*/
			if chunkPL.MediaSeq == prevChunkMediaSEQ {
				equalMediaSEQCount++
				if equalMediaSEQCount > 10 && equalMediaSEQCount%5 == 0 {
					log.Printf("[ERROR] {stream: %s} media sequence в chunks не изменился за последние %d попыток\n", mainURI, equalMediaSEQCount)
				}
				time.Sleep(time.Second)
				continue
			}
			equalMediaSEQCount, prevChunkMediaSEQ = 0, chunkPL.MediaSeq

			// обработка iframe плейлиста
			r2, err := fetchURL(iframeURI, nil)
			if err != nil {
				log.Printf("[ERROR] {stream: %s} в процессе скачивания iframe-плейлиста: %s\n", mainURI, err.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			iframePL, err := parser.ParsePlayList(r2)
			r2.Close()
			if err != nil {
				log.Printf("[ERROR] {stream: %s} при парсинге iframe-плейлиста: %s\n", mainURI, err.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			if !iframePL.IFrame {
				log.Printf("[ERROR] {stream: %s} проблема с iframe-плейлистом %s: это не iframe-плейлист\n", mainURI, iframeURI)
			}
			iframePL.SetURL(iframeURL)

			if iframePL.MediaSeq == prevIframeMediaSEQ {
				equalMediaSEQCount++
				if equalMediaSEQCount > 10 && equalMediaSEQCount%5 == 0 {
					log.Printf("[ERROR] {stream: %s} media sequence в iframes не изменился за последние %d попыток\n", mainURI, equalMediaSEQCount)
				}
				time.Sleep(time.Second)
				continue
			}
			equalMediaSEQCount, prevIframeMediaSEQ = 0, iframePL.MediaSeq

			/*
				if iframePL.MediaSeq < chunkPL.MediaSeq {
					log.Printf("[ERROR] {stream: %s} media sequence iframes %d < chunks %d\n", mainURI, iframePL.MediaSeq, chunkPL.MediaSeq)
					continue
				}
			*/

			minutes, err := makeMinutes(chunkPL.Segments, iframePL.Segments)
			if err != nil {
				log.Printf("[ERROR] {stream: %s} при создании плана минуток: %s\n", mainURI, err.Error())
				time.Sleep(5 * time.Second)
				continue
			}

			lastMinute := minutes.last()
			for _, m := range minutes.list {
				if !m.indexHasEOF(indexDir) {
					// необходимо записать минутку
					if m.full && m.beginAt != lastMinute.beginAt {
						// записываем минутку полностью
						log.Printf("[INFO] {stream: %s} старт записи полной минуты %d\n", mainURI, m.beginAt)
						if err := m.writeFull(indexDir, storageDir, resource, vmx); err != nil {
							log.Printf("[ERROR] {stream: %s} запись полной минуты %d: %s\n", mainURI, m.beginAt, err.Error())
							continue
						}
						log.Printf("[INFO] {stream: %s} успешная запись полной минуты %d\n", mainURI, m.beginAt)
					} else {
						err, chunks, iframes, last := m.writePartical(indexDir, storageDir, resource, vmx)
						if err != nil {
							log.Printf("[ERROR] {stream: %s} обработка минуты минуты %d: %s\n", mainURI, m.beginAt, err.Error())
							continue
						}
						if chunks+iframes > 0 {
							log.Printf("[DEBUG] {stream: %s} обработка минуты %d было записано: [chunks:%d iframes:%d lag:%.2f]", mainURI, m.beginAt, chunks, iframes, float64(time.Now().Unix())-last)
						}
					}

				}
			}

			if wasStopped {
				return
			}
			time.Sleep(time.Second)
		}
	}()

	select {
	case <-ctx.Done():
		log.Printf("[INFO] stop process playlist `%s` \n", stream.MainURI)
		wasStopped = true
		return
	}

}
