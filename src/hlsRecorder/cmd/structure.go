package main

import (
	"context"
	"fmt"
	"path/filepath"

	keys "hlsRecorder/keys"
	parser "hlsRecorder/parser"
	writer "hlsRecorder/writer"
)

type Config struct {
	VMXURL      string
	StoragePath string
	IndexPath   string
	Channels    []*Channel
}

type Channel struct {
	Resource    string
	BW          string
	storagePath string
	indexPath   string
	Stream      Stream
	cancelFunc  context.CancelFunc
}

type Stream struct {
	MainURI   string
	IFrameURI string
}

func (c *Channel) Start(vmx *keys.VMX) {
	ctx, f := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, `content.channel`, c.Resource)
	ctx = context.WithValue(ctx, `path.storage.dir`, filepath.Join(c.storagePath, c.BW))
	ctx = context.WithValue(ctx, `path.index.dir`, filepath.Join(c.indexPath, c.BW))
	ctx = context.WithValue(ctx, `keys.vmx`, vmx)
	c.cancelFunc = f

	writer.Stream(&parser.Stream{
		LogName:   fmt.Sprintf("%s-%s", c.Resource, c.BW),
		MainURI:   c.Stream.MainURI,
		IFrameURI: c.Stream.IFrameURI,
	}, ctx)
}

func (c *Config) Start() {
	vmx := keys.New(c.VMXURL)
	for _, channel := range c.Channels {
		channel.storagePath = filepath.Join(c.StoragePath, channel.Resource)
		channel.indexPath = filepath.Join(c.IndexPath, channel.Resource)
		go channel.Start(vmx)
	}
}

func (c *Config) Stop() {
	for _, channel := range c.Channels {
		if channel.cancelFunc != nil {
			go channel.cancelFunc()
		}
	}
}
