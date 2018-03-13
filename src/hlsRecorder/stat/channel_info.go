package stat

import (
	"sync"
	"time"
)

type ChannelInfo struct {
	WriteBytes int64                `json:"write_bytes"`
	PlayList   *ChannelDownloadInfo `json:"playlist"`
	Data       *ChannelDownloadInfo `json:"chunks"`
	Lag        float64              `json:"lag"`
}

type ChannelDownloadInfo struct {
	sync.Mutex
	list           []float64
	ErrorCount     int     `json:"error_count"`
	Percentile75ms float64 `json:"percentile_75_ms"`
	Percentile95ms float64 `json:"percentile_95_ms"`
	Percentile99ms float64 `json:"percentile_99_ms"`
}

func newChannelInfo() *ChannelInfo {

	result := &ChannelInfo{
		Data: &ChannelDownloadInfo{
			list: make([]float64, 0),
		},
		PlayList: &ChannelDownloadInfo{
			list: make([]float64, 0),
		},
	}

	go result.compact()

	return result
}

func (i *ChannelInfo) compact() {
	for {

		time.Sleep(time.Minute)

		i.Data.Lock()
		if length := len(i.Data.list); length > 4*1024 {
			i.Data.list = i.Data.list[:length-1024]
		}
		i.Data.Unlock()

		i.PlayList.Lock()
		if length := len(i.PlayList.list); length > 4*1024 {
			i.PlayList.list = i.PlayList.list[:length-1024]
		}
		i.PlayList.Unlock()

	}
}

func (i *ChannelDownloadInfo) AddError() {
	i.Lock()
	i.ErrorCount++
	i.Unlock()
}

func (i *ChannelDownloadInfo) AddTime(t float64) {
	i.Lock()
	i.list = append(i.list, t)
	i.Unlock()
}

func (c *ChannelInfo) AddWrite(i int64) {
	c.WriteBytes += i
}
