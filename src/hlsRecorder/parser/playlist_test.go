package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseChunkList(t *testing.T) {
	f, err := os.Open(filepath.Join(`test`, `chunklist.m3u8`))
	if err != nil {
		t.Fatalf("open: %s\n", err.Error())
	}
	pl, err := ParsePlayList(f)
	if err != nil {
		t.Fatalf("parse error: %s\n", err.Error())
	}
	if pl.IFrame {
		t.Fatalf("not must be iframe playlist")
	}
	if pl.MediaSeq != 67016 {
		t.Fatalf("parse media seq: get %d except: %d\n", pl.MediaSeq, 67016)
	}
	for _, s := range pl.Segments {
		if !strings.HasPrefix(s.URI, fmt.Sprintf("%.2f-", s.BeginAt)) {
			t.Errorf("bad BeginAt in segment: %#v\n", s)
		}
		if !strings.HasSuffix(s.URI, fmt.Sprintf("%.2f.ts", s.BeginAt+s.Duration)) {
			t.Errorf("bad Duration in segment: %#v\n", s)
		}
	}
}

func TestParseIframeChunkList(t *testing.T) {
	f, err := os.Open(filepath.Join(`test`, `iframe_chunklist.m3u8`))
	if err != nil {
		t.Fatalf("open: %s\n", err.Error())
	}
	pl, err := ParsePlayList(f)
	if err != nil {
		t.Fatalf("parse error: %s\n", err.Error())
	}
	if pl.MediaSeq != 67028 {
		t.Fatalf("parse media seq: get %d except: %d\n", pl.MediaSeq, 67028)
	}
	if !pl.IFrame {
		t.Fatalf("must be iframe playlist")
	}
	for _, s := range pl.Segments {
		if !strings.HasPrefix(s.URI, fmt.Sprintf("%.2f-", s.BeginAt)) {
			t.Errorf("bad BeginAt in segment: %#v\n", s)
		}
		if !strings.HasSuffix(s.URI, fmt.Sprintf("%.2f.ts", s.BeginAt+s.Duration)) {
			t.Errorf("bad Duration in segment: %#v\n", s)
		}
		byteRange := s.ByteRange
		if byteRange == nil {
			t.Fatalf("bad byte range in segment: %#v\n", s)
		}
		if !(byteRange.Offset%188 == 0) || !(byteRange.Length%188 == 0) {
			t.Errorf("bad byte range %#v in segment: %#v\n", byteRange, s)
		}
	}
}

func TestParseIframeChunkListWithOneProgram(t *testing.T) {
	f, err := os.Open(filepath.Join(`test`, `iframe_chunklist_with_one_program_date.m3u8`))
	if err != nil {
		t.Fatalf("open: %s\n", err.Error())
	}
	pl, err := ParsePlayList(f)
	if err != nil {
		t.Fatalf("parse error: %s\n", err.Error())
	}
	if pl.MediaSeq != 67028 {
		t.Fatalf("parse media seq: get %d except: %d\n", pl.MediaSeq, 67028)
	}
	for _, s := range pl.Segments {
		if !strings.HasPrefix(s.URI, fmt.Sprintf("%.2f-", s.BeginAt)) {
			t.Errorf("bad BeginAt in segment: %#v\n", s)
		}
		if !strings.HasSuffix(s.URI, fmt.Sprintf("%.2f.ts", s.BeginAt+s.Duration)) {
			t.Errorf("bad Duration in segment: %#v\n", s)
		}
		byteRange := s.ByteRange
		if byteRange == nil {
			t.Fatalf("bad byte range in segment: %#v\n", s)
		}
		if !(byteRange.Offset%188 == 0) || !(byteRange.Length%188 == 0) {
			t.Errorf("bad byte range %#v in segment: %#v\n", byteRange, s)
		}
	}
}
