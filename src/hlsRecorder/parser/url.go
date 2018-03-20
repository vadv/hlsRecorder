package parser

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

func isURL(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	return (u.Scheme == `http` || u.Scheme == `https`)
}

// u1 = second.html, u2 = http://ya.ru/first.html => http://ya.ru/second.html
// u1 = /second/second.html, u2 = http://ya.ru/first/first.html => http://ya.ru/second/second.html
func joinURL(u1 *url.URL, u2 string) string {
	u := *u1
	if strings.HasPrefix(u2, `/`) {
		return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u2)
	} else {
		u.Path = path.Join(path.Dir(u.Path), u2)
	}
	return u.String()
}

func (i *Index) SetURL(u *url.URL) error {
	if i.Streams == nil {
		return fmt.Errorf("empty streams info")
	}
	for _, s := range i.Streams {
		if s.IFrameURI != `` && !isURL(s.IFrameURI) {
			s.IFrameURI = joinURL(u, s.IFrameURI)
		}
		if s.MainURI != `` && !isURL(s.MainURI) {
			s.MainURI = joinURL(u, s.MainURI)
		}
	}
	return nil
}

func (p *PlayList) SetURL(u *url.URL) error {
	if p.Segments == nil {
		return fmt.Errorf("empty segments info")
	}
	for _, s := range p.Segments {
		if !isURL(s.URI) {
			s.URL = joinURL(u, s.URI)
		} else {
			s.URL = s.URI
		}
	}
	return nil
}
