package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

/*

# Пример конфига с каналами

CH_TV1HDREC:
  vmx: true
  rotate_hours: 72
  bandwidths:
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

func channelConfigParser(filename string) (map[string][]*Channel, error) {

	type Urls struct {
		Chunks  string `yaml:"chunks"`
		IFrames string `yaml:"iframes"`
	}

	type ChannelConfig struct {
		VMX         bool             `yaml:"vmx"`
		RotateHours int64            `yaml:"rotate_hours"`
		Bandwidths  map[string]*Urls `yaml:"bandwidths"`
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	maps, result := make(map[string]ChannelConfig, 0), make(map[string][]*Channel, 0)
	err = yaml.Unmarshal(data, &maps)
	if err != nil {
		return nil, err
	}

	for resource, settings := range maps {
		result[resource] = make([]*Channel, 0)
		for dirName, bw := range settings.Bandwidths {
			channel := &Channel{
				UseVMX:      settings.VMX,
				DeleteOlder: settings.RotateHours * 60 * 60,
				Resource:    resource,
				BW:          dirName,
				Stream: Stream{
					MainURI:   bw.Chunks,
					IFrameURI: bw.IFrames,
				},
			}
			result[resource] = append(result[resource], channel)
		}

	}

	return result, nil

}
