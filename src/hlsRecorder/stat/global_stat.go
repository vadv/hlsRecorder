package stat

import (
	"encoding/json"
	"runtime"
	"sync"
	"time"
)

type GlobalStat struct {
	sync.Mutex
	started_at time.Time
	Uptime     float64                            `json:"uptime"`
	Alloc      uint64                             `json:"memory_alloc"`
	TotalAlloc uint64                             `json:"memory_total_alloc"`
	Sys        uint64                             `json:"memory_sys"`
	GC         uint32                             `json:"num_gc"`
	Channels   map[string]map[string]*ChannelInfo `json:"channels"`
}

func New() *GlobalStat {
	return &GlobalStat{
		started_at: time.Now(),
		Channels:   make(map[string]map[string]*ChannelInfo, 0),
	}
}

func (g *GlobalStat) AddChannel(r string, bw string) *ChannelInfo {
	g.Lock()
	defer g.Unlock()
	i := newChannelInfo()
	if g.Channels[r] == nil {
		g.Channels[r] = make(map[string]*ChannelInfo, 0)
	}
	g.Channels[r][bw] = i
	return i
}

func (g *GlobalStat) ToJson() []byte {

	g.Lock()
	defer g.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	g.Alloc = m.Alloc
	g.TotalAlloc = m.TotalAlloc
	g.Sys = m.Sys
	g.GC = m.NumGC

	g.Uptime = time.Now().Sub(g.started_at).Seconds()

	for _, bws := range g.Channels {
		for _, i := range bws {
			i.CalcStat()
		}
	}

	data, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}

	return data
}
