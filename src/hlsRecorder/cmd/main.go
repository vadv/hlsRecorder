package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	keys "hlsRecorder/keys"
	parser "hlsRecorder/parser"
	writer "hlsRecorder/writer"
)

type Config struct {
	vmxURL      string
	StoragePath string
	IndexPath   string
	Channels    []*Channel
}

type Channel struct {
	Resource    string
	BW          string
	StoragePath string
	IndexPath   string
	Stream      *parser.Stream
	CancelFunc  context.CancelFunc
}

func (c *Channel) Start(vmx *keys.VMX) {
	ctx, f := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, `content.channel`, c.Resource)
	ctx = context.WithValue(ctx, `path.storage.dir`, filepath.Join(c.StoragePath, c.BW))
	ctx = context.WithValue(ctx, `path.index.dir`, filepath.Join(c.IndexPath, c.BW))
	ctx = context.WithValue(ctx, `keys.vmx`, vmx)
	c.CancelFunc = f

	writer.Stream(c.Stream, ctx)
}

func (c *Config) Start() {
	vmx := keys.New(c.vmxURL)
	for _, channel := range c.Channels {
		channel.StoragePath = filepath.Join(c.StoragePath, channel.Resource)
		channel.IndexPath = filepath.Join(c.IndexPath, channel.Resource)
		go channel.Start(vmx)
	}
}

func (c *Config) Stop() {
	for _, channel := range c.Channels {
		if channel.CancelFunc != nil {
			go channel.CancelFunc()
		}
	}
}

func main() {

	config := &Config{
		vmxURL:      `http://192.168.184.49:12684/CAB/keyfile`,
		StoragePath: `/video`,
		IndexPath:   `/videoidx-rec`,
		Channels: []*Channel{
			&Channel{
				Resource: `CH_TV1HDREC`,
				BW:       `bw0`,
				Stream: &parser.Stream{
					MainURI:   `http://192.168.185.148/indextest/0/chunklist.m3u8`,
					IFrameURI: `http://192.168.185.148/indextest/0/iframe_chunklist.m3u8`,
				},
			},
			&Channel{
				Resource: `CH_TV1HDREC`,
				BW:       `bw1`,
				Stream: &parser.Stream{
					MainURI:   `http://192.168.185.148/indextest/1/chunklist.m3u8`,
					IFrameURI: `http://192.168.185.148/indextest/1/iframe_chunklist.m3u8`,
				},
			},
			&Channel{
				Resource: `CH_TV1HDREC`,
				BW:       `bw2`,
				Stream: &parser.Stream{
					MainURI:   `http://192.168.185.148/indextest/2/chunklist.m3u8`,
					IFrameURI: `http://192.168.185.148/indextest/2/iframe_chunklist.m3u8`,
				},
			},
			&Channel{
				Resource: `CH_TV1HDREC`,
				BW:       `bw3`,
				Stream: &parser.Stream{
					MainURI:   `http://192.168.185.148/indextest/3/chunklist.m3u8`,
					IFrameURI: `http://192.168.185.148/indextest/3/iframe_chunklist.m3u8`,
				},
			},
			&Channel{
				Resource: `CH_TV1HDREC`,
				BW:       `bw4`,
				Stream: &parser.Stream{
					MainURI:   `http://192.168.185.148/indextest/4/chunklist.m3u8`,
					IFrameURI: `http://192.168.185.148/indextest/4/iframe_chunklist.m3u8`,
				},
			},
		},
	}

	go config.Start()

	// перехватываем Ctr+C
	halt := make(chan os.Signal, 1)
	signal.Notify(halt, os.Interrupt)
	signal.Notify(halt, syscall.SIGTERM)

	<-halt
	config.Stop()
	os.Exit(1)
}
