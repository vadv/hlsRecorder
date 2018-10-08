package main

import (
	"context"

	stat "hlsRecorder/stat"
)

type Config struct {
	VMXURL      string
	StoragePath string
	IndexPath   string
	GlobalStat  *stat.GlobalStat
	Channels    map[string][]*Channel
}

type Channel struct {
	UseVMX      bool
	DeleteOlder int64
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
	Hosts     []string
}
