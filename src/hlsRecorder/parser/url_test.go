package parser

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestSetUrl(t *testing.T) {
	f, err := os.Open(filepath.Join(`test`, `url.m3u8`))
	if err != nil {
		t.Fatalf("open: %s\n", err.Error())
	}
	pl, err := ParsePlayList(f)
	if err != nil {
		t.Fatalf("parse error: %s\n", err.Error())
	}

	if len(pl.Segments) < 3 {
		t.Fatalf("bad segments: %#v\n", pl.Segments)
	}

	u, _ := url.Parse(`http://server1/index/0/playlist.m3u8`)
	pl.SetURL(u)

	for i, s := range pl.Segments {
		switch i {

		case 0:
			if s.URL != `http://server2/i0/1519484820.94-1519484823.94.ts` {
				t.Errorf("bad URL %#v\n", s)
			}

		case 1:
			if s.URL != `http://server1/index/0/i1/1519484823.94-1519484826.94.ts` {
				t.Errorf("bad URL %#v\n", s)
			}

		case 2:
			if s.URL != `http://server1/index/0/1519484826.94-1519484829.94.ts` {
				t.Errorf("bad URL %#v\n", s)
			}

		}
	}

}
