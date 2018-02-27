package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"testing"
)

func TestParseIndex(t *testing.T) {
	f, err := os.Open(filepath.Join(`test`, `index.m3u8`))
	if err != nil {
		t.Fatalf("open: %s\n", err.Error())
	}
	i, err := ParseIndex(f)
	if err != nil {
		t.Fatalf("parse error: %s\n", err.Error())
	}
	if len(i.Streams) != 5 {
		for _, s := range i.Streams {
			t.Errorf("stream: %#v\n", s)
		}
	}
	for i, s := range i.Streams {
		if s.IFrameURI != fmt.Sprintf("%d/iframe_chunklist.m3u8", i) {
			t.Errorf("bad stream: %#v\n", s)
		}
		if s.MainURI != fmt.Sprintf("%d/chunklist.m3u8", i) {
			t.Errorf("bad stream: %#v\n", s)
		}
	}
}
