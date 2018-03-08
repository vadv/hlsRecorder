package main

import (
	"context"
)

type Config struct {
	VMXURL      string
	StoragePath string
	IndexPath   string
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
}
