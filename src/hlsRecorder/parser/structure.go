package parser

import "fmt"

type Index struct {
	Streams []*Stream
}

type Stream struct {
	Resolution    string
	MainBandwidth string
	MainURI       string
	IFrameURI     string
}

type PlayList struct {
	URI      string
	MediaSeq int64
	IFrame   bool
	Segments []*Segment
}

type Segment struct {
	URI       string
	URL       string
	Duration  float64
	BeginAt   float64
	EndAt     float64
	ByteRange *ByteRange
}

type ByteRange struct {
	Offset int64
	Length int64
}

func (s *Segment) ToString() string {
	url := s.URL
	if url == `` {
		url = s.URI
	}
	if br := s.ByteRange; br != nil {
		return fmt.Sprintf("%s [%d:%d]", url, br.Offset, br.Length)
	} else {
		return fmt.Sprintf("%s", url)
	}
}
