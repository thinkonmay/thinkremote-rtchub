package adaptive

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)


const (
	queue_size = 100000
	evaluation_period = 30
)

type AdsCtx struct {
	vqueue chan *VideoMetric
	aqueue chan *AudioMetric
	nqueue chan *NetworkMetric

	last *struct {
		audio   *AudioMetricRaw
		video   *VideoMetricRaw
		network *NetworkMetricRaw
	}

	print_fps_queue chan int
}

type AdsMultiCtxs struct {
	In  chan string
	Out chan string

	videoResetQueue chan bool
	triggerVideoReset func()
	bitrateChangeFunc func(bitrate int)



	mut *sync.Mutex
	ctxs map[string]*AdsCtx
}

func NewAdsContext(BitrateCallback func(bitrate int),
	IDRcallback func()) datachannel.DatachannelConsumer {
	ret := &AdsMultiCtxs{
		In:  make(chan string),
		Out: make(chan string),

		triggerVideoReset: IDRcallback,
		videoResetQueue: make(chan bool,50),
		bitrateChangeFunc: BitrateCallback,
		mut: &sync.Mutex{},
		ctxs: make(map[string]*AdsCtx),

	}

	go func() {
		for {
			ret.mut.Lock()
			for name,ac := range ret.ctxs {
				if len(ac.print_fps_queue) < evaluation_period {
					continue
				}


				fpses := []int{}
				for i := 0; i < evaluation_period; i++ {
					vid:=<-ac.print_fps_queue
					fpses = append(fpses, int(vid))
				}

				fmt.Printf("[%s] fps for worker context %s: %v ",time.Now().Format(time.RFC3339),name,fpses)
			}
			ret.mut.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}()
	go func() {
		for {
			do_reset := false
			for i := 0; i < len(ret.videoResetQueue); i++ {
				_=<-ret.videoResetQueue
				do_reset = true
			}

			if do_reset {
				ret.triggerVideoReset()
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()
	go func() {
		for {
			ret.mut.Lock()
			for _,ac := range ret.ctxs {
				if len(ac.vqueue) == 0 {
					continue
				}

				vid:=<-ac.vqueue
				if vid.DecodedFps < 5 {
					ret.videoResetQueue<-true
				}
				ac.print_fps_queue<-int(vid.DecodedFps)
			}
			ret.mut.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		for {
			metricRaw := <-ret.In
			var out map[string]interface{}
			json.Unmarshal([]byte(metricRaw), &out)

			switch out["type"] {
			case "video":
				video := VideoMetricRaw{}
				json.Unmarshal([]byte(metricRaw), &video)
				ret.handleVideoMetric(&video)
			case "audio":
				audio := AudioMetricRaw{}
				json.Unmarshal([]byte(metricRaw), &audio)
				ret.handleAudioMetric(&audio)
			case "network":
				network := NetworkMetricRaw{}
				json.Unmarshal([]byte(metricRaw), &network)
				ret.handleNetworkMetric(&network)
			}
		}
	}()

	return ret
}

func (ads *AdsMultiCtxs) handleNewContext(source string) {
	ads.mut.Lock()
	ads.ctxs[source] = &AdsCtx{
		aqueue: make(chan *AudioMetric,queue_size),
		vqueue: make(chan *VideoMetric,queue_size),
		nqueue: make(chan *NetworkMetric,queue_size),
		print_fps_queue: make(chan int,queue_size),
		last: &struct{audio *AudioMetricRaw; video *VideoMetricRaw; network *NetworkMetricRaw}{
			audio: nil,
			video: nil,
			network: nil,
		},
	}
	ads.mut.Unlock()
}

func (ads *AdsMultiCtxs) handleVideoMetric(metric *VideoMetricRaw) {
	if ads.ctxs[metric.Source] == nil {
		ads.handleNewContext(metric.Source)
	}

	last := ads.ctxs[metric.Source].last
	lastVideoMetric := last.video
	last.video = metric
	if lastVideoMetric == nil {
		return
	}

	timedif := (metric.Timestamp - lastVideoMetric.Timestamp) //nanosecond
	video := &VideoMetric{
		DecodedFps : 					(metric.FramesDecoded 	- lastVideoMetric.FramesDecoded) 			/ (timedif / float64(time.Second.Milliseconds())),
		ReceivedFps : 					(metric.FramesReceived 	- lastVideoMetric.FramesReceived) 			/ (timedif / float64(time.Second.Milliseconds())),
		VideoBandwidthConsumption : 	(metric.BytesReceived 	- lastVideoMetric.BytesReceived) 			/ (timedif / float64(time.Second.Milliseconds())),
		DecodeTimePerFrame : 			(metric.TotalDecodeTime - lastVideoMetric.TotalDecodeTime) 			/ (metric.FramesDecoded   - lastVideoMetric.FramesDecoded),
		VideoPacketsLostpercent : 		(metric.PacketsLost 	- lastVideoMetric.PacketsLost) 			    / (metric.PacketsReceived - lastVideoMetric.PacketsReceived),
		VideoJitter : metric.Jitter,
		VideoJitterBufferDelay : metric.JitterBufferDelay,
		Timestamp: metric.Timestamp,
	}


	if len(ads.ctxs[metric.Source].vqueue) < queue_size{
		ads.ctxs[metric.Source].vqueue<-video
	}
}

func (ads *AdsMultiCtxs) handleNetworkMetric(metric *NetworkMetricRaw) {
	if ads.ctxs[metric.Source] == nil {
		ads.handleNewContext(metric.Source)
	}

	last := ads.ctxs[metric.Source].last
	lastNetworkMetric := last.network
	last.network = metric
	if lastNetworkMetric == nil {
		return
	}

	timedif := metric.Timestamp - lastNetworkMetric.Timestamp //nanosecond
	network := &NetworkMetric{
		TotalBandwidthConsumption 	: (metric.BytesReceived - lastNetworkMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds())),
		RTT 						: metric.CurrentRoundTripTime * float64(time.Second.Nanoseconds()),
		AvailableIncomingBandwidth 	: metric.AvailableIncomingBitrate,
		Timestamp					: metric.Timestamp,
	}
	
	if len(ads.ctxs[metric.Source].nqueue) < queue_size{
		ads.ctxs[metric.Source].nqueue<-network
	}
}

func (ads *AdsMultiCtxs) handleAudioMetric(metric *AudioMetricRaw) {
	if ads.ctxs[metric.Source] == nil {
		ads.handleNewContext(metric.Source)
	}

	last := ads.ctxs[metric.Source].last
	lastAudioMetric := last.audio
	last.audio = metric
	if lastAudioMetric == nil {
		return
	}

	timedif := metric.Timestamp - lastAudioMetric.Timestamp //nanosecond
	audio:=&AudioMetric{
		AudioBandwidthConsumption : (metric.BytesReceived - lastAudioMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds())),
		Timestamp: metric.Timestamp,
	}

	if len(ads.ctxs[metric.Source].aqueue) < queue_size{
		ads.ctxs[metric.Source].aqueue<-audio
	}
}

func (ads *AdsMultiCtxs) Send(msg string) {
	ads.In <- msg
}

func (ads *AdsMultiCtxs) Recv() string {
	return <-ads.Out
}
