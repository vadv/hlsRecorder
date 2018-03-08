package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	flagVMX      = flag.String("vmx", "http://192.168.184.49:12684/CAB/keyfile", "vmx url")
	flagStorage  = flag.String("storage", "/video", "video storage path")
	flagIndex    = flag.String("index", "/videoidx-rec", "video index path")
	flagChannels = flag.String("channels", "/etc/hlsRecorder/channels.yml", "channel config")
)

func main() {

	if !flag.Parsed() {
		flag.Parse()
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	channels, err := channelConfigParser(*flagChannels)
	if err != nil {
		log.Printf("[FATAL] при парсинге конфига %s: %s\n", *flagChannels, err.Error())
		os.Exit(1)
	}

	config := &Config{
		Channels:    channels,
		VMXURL:      *flagVMX,
		StoragePath: *flagStorage,
		IndexPath:   *flagIndex,
	}

	go config.Start()

	// перехватываем Ctr+C
	halt := make(chan os.Signal, 1)
	signal.Notify(halt, os.Interrupt)
	signal.Notify(halt, syscall.SIGTERM)

	<-halt
	config.Stop()
	time.Sleep(5 * time.Second)
	os.Exit(1)
}
