package main

import (
	"net"
	"net/http"
)

func (c *Config) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		if req.URL.Path == "/stat.json" {
			data := c.GlobalStat.ToJson()
			w.Write(data)
			return
		}

	}
	w.WriteHeader(http.StatusNotAcceptable)
}

func (c *Config) RunWeb() {
	listener, err := net.Listen("tcp", *flagListen)
	if err != nil {
		panic(err)
	}
	if err := http.Serve(listener, c); err != nil {
		panic(err)
	}
}
