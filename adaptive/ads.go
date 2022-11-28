package adaptive

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

// #cgo LDFLAGS: ${SRCDIR}/../build/libshared.a
// #include "ads.h"
import "C"

type AdsContext unsafe.Pointer

type AdaptiveContext struct {
	In  chan string
	ctx AdsContext
}

func NewAdsContext(InChan chan string) *AdaptiveContext {
	ret := &AdaptiveContext{}
	C.new_ads_context()

	ret.In = InChan
	go func() {
		for {
			metricRaw := <-ret.In
			var out map[string]interface{}
			json.Unmarshal([]byte(metricRaw), &out)

			switch out["type"] {
			case "video":
				video := VideoMetrics{}
				json.Unmarshal([]byte(metricRaw), &video)
				fmt.Printf("%v", video)
			case "audio":
				audio := AudioMetric{}
				json.Unmarshal([]byte(metricRaw), &audio)
				fmt.Printf("%v", audio)
			case "network":
				network := NetworkMetric{}
				json.Unmarshal([]byte(metricRaw), &network)
				fmt.Printf("%v", network)
			}
		}
	}()

	return ret
}

// func (ads *AdsContext)
