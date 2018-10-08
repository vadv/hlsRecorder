package main

import (
	"context"
	"fmt"
	"path/filepath"

	keys "hlsRecorder/keys"
	parser "hlsRecorder/parser"
	stat "hlsRecorder/stat"
	writer "hlsRecorder/writer"
)

func (c *Channel) Start(vmx *keys.VMX, stat *stat.GlobalStat) {

	ctx, f := context.WithCancel(context.Background())
	c.cancelFunc = f

	ctx = context.WithValue(ctx, `content.channel`, c.Resource)
	ctx = context.WithValue(ctx, `path.storage.dir`, filepath.Join(c.storagePath, c.BW))
	ctx = context.WithValue(ctx, `path.index.dir`, filepath.Join(c.indexPath, c.BW))
	ctx = context.WithValue(ctx, `path.delete_older`, c.DeleteOlder)
	if c.UseVMX {
		ctx = context.WithValue(ctx, `keys.vmx`, vmx)
	}
	channelStat := stat.AddChannel(c.Resource, c.BW)
	ctx = context.WithValue(ctx, `stat.channel_info`, channelStat)

	writer.Stream(&parser.Stream{
		LogName:   fmt.Sprintf("%s/%s", c.Resource, c.BW),
		Hosts:     c.Stream.Hosts,
		MainURI:   c.Stream.MainURI,
		IFrameURI: c.Stream.IFrameURI,
	}, ctx)
}

func (c *Config) StartRecord() {
	vmx := keys.New(c.VMXURL)
	for _, channels := range c.Channels {
		for _, channel := range channels {
			channel.storagePath = filepath.Join(c.StoragePath, channel.Resource)
			channel.indexPath = filepath.Join(c.IndexPath, channel.Resource)
			go channel.Start(vmx, c.GlobalStat)
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
