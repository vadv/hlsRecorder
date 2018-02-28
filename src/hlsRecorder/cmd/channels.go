package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

/*

# Пример конфига с каналами

  CH_TV1HDREC:
    bw700000:
      chunks:  http://192.168.185.148/indextest/0/chunklist.m3u8
      iframes: http://192.168.185.148/indextest/0/iframe_chunklist.m3u8
    bw1500000:
      chunks:  http://192.168.185.148/indextest/1/chunklist.m3u8
      iframes: http://192.168.185.148/indextest/1/iframe_chunklist.m3u8
    bw2000000:
      chunks:  http://192.168.185.148/indextest/2/chunklist.m3u8
      iframes: http://192.168.185.148/indextest/2/iframe_chunklist.m3u8
    bw4000000:
      chunks:  http://192.168.185.148/indextest/3/chunklist.m3u8
      iframes: http://192.168.185.148/indextest/3/iframe_chunklist.m3u8
    bw5000000:
      chunks:  http://192.168.185.148/indextest/4/chunklist.m3u8
      iframes: http://192.168.185.148/indextest/4/iframe_chunklist.m3u8
*/

func channelParser(filename string) ([]*Channel, error) {

	type Urls struct {
		Chunks  string `yaml:"chunks"`
		IFrames string `yaml:"iframes"`
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c, result := make(map[string]map[string]Urls, 0), make([]*Channel, 0)
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	for channel, channelConfig := range c {
		for bw, urls := range channelConfig {
			newChannel := &Channel{
				Resource: channel,
				BW:       bw,
				Stream: Stream{
					MainURI:   urls.Chunks,
					IFrameURI: urls.IFrames,
				},
			}
			result = append(result, newChannel)
		}
	}

	return result, nil

}
