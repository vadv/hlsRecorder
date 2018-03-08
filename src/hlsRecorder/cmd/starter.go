package main

import (
	"context"
	"fmt"
	"path/filepath"

	keys "hlsRecorder/keys"
	parser "hlsRecorder/parser"
	writer "hlsRecorder/writer"
)

func (c *Channel) Start(vmx *keys.VMX) {
	ctx, f := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, `content.channel`, c.Resource)
	ctx = context.WithValue(ctx, `path.storage.dir`, filepath.Join(c.storagePath, c.BW))
	ctx = context.WithValue(ctx, `path.index.dir`, filepath.Join(c.indexPath, c.BW))
	ctx = context.WithValue(ctx, `path.delete_older`, c.DeleteOlder)
	if c.UseVMX {
		ctx = context.WithValue(ctx, `keys.vmx`, vmx)
	}
	c.cancelFunc = f

	writer.Stream(&parser.Stream{
		LogName:   fmt.Sprintf("%s/%s", c.Resource, c.BW),
		MainURI:   c.Stream.MainURI,
		IFrameURI: c.Stream.IFrameURI,
	}, ctx)
}

func (c *Config) Start() {
	vmx := keys.New(c.VMXURL)
	for _, channels := range c.Channels {
		for _, channel := range channels {
			channel.storagePath = filepath.Join(c.StoragePath, channel.Resource)
			channel.indexPath = filepath.Join(c.IndexPath, channel.Resource)
			go channel.Start(vmx)
		}
	}
}

func (c *Config) Stop() {
	for _, channels := range c.Channels {
		for _, channel := range channels {
			if channel.cancelFunc != nil {
				go channel.cancelFunc()
			}
		}
	}
}
