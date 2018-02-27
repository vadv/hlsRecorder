package main

import (
	"context"

	keys "hlsRecorder/keys"
	parser "hlsRecorder/parser"
	writer "hlsRecorder/writer"
)

func main() {

	//	vmx := keys.New(`https://vmxott.svc.iptv.rt.ru/CAB/keyfile`)
	vmx := keys.New(`http://192.168.184.49:12684/CAB/keyfile`)

	stream := &parser.Stream{
		MainURI:   `http://192.168.185.148/indextest/4/chunklist.m3u8`,
		IFrameURI: `http://192.168.185.148/indextest/4/iframe_chunklist.m3u8`,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, `content.channel`, `CH_hlsRecorder`)
	ctx = context.WithValue(ctx, `path.storage.dir`, `/video/CH_hlsRecorder/bw4/`)
	ctx = context.WithValue(ctx, `path.index.dir`, `/videoidx-rec/CH_hlsRecorder/bw4/`)
	ctx = context.WithValue(ctx, `keys.vmx`, vmx)

	writer.Stream(stream, ctx)
}
