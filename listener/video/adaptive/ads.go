package adaptive

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)


const (
	queue_size = 1000
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

	afterVQueue chan *VideoMetric
	afterAQueue chan *AudioMetric
	afterNQueue chan *NetworkMetric
}

type AdsMultiCtxs struct {
	in  chan string
	out chan string

	triggerVideoReset func()
	bitrateChangeFunc func(bitrate int)



	mut *sync.Mutex
	ctxs map[string]*AdsCtx
}

func NewAdsContext(BitrateCallback func(bitrate int),
	IDRcallback func()) datachannel.DatachannelConsumer {
	ret := &AdsMultiCtxs{
		in:  make(chan string,queue_size),
		out: make(chan string,queue_size),

		triggerVideoReset: IDRcallback,
		bitrateChangeFunc: BitrateCallback,
		mut: &sync.Mutex{},
		ctxs: make(map[string]*AdsCtx),
	}

	afterprocess := func() {
		for {
			ret.mut.Lock()
			for _,ac := range ret.ctxs {
				receivefpses,decodefpses,
				bandwidth,packetloss,buffer := 
				[]int{},[]int{},[]int{},
				[]int{},[]int{}

				if len(ac.afterVQueue) < evaluation_period {
					continue
				}

				for i := 0; i < evaluation_period; i++ {
					vid:=<-ac.afterVQueue
					decodefpses = append(decodefpses, int(vid.DecodedFps))
					receivefpses = append(receivefpses, int(vid.ReceivedFps))
					bandwidth   = append(bandwidth, int(vid.VideoBandwidthConsumption))
					packetloss = append(packetloss, int(vid.VideoPacketsLostpercent))
					buffer = append(buffer, int(vid.BufferedFrame))
				}

				value := struct{
					Type string `json:"type"`
					ReceiveFps []int `json:"receivefps"`
					DecodeFps  []int `json:"decodefps"`
					PacketsLoss []int `json:"packetloss"`
					Bandwidth []int `json:"bandwidth"`
					Buffer   []int `json:"buffer"`
				}{
					Type: "VIDEO",
					DecodeFps: decodefpses,
					ReceiveFps: receivefpses,
					Bandwidth : bandwidth,
					PacketsLoss : packetloss,
					Buffer : buffer,
				}

				data,_ :=json.Marshal(&value);
				ret.out<-string(data)
				receivefpses,decodefpses = []int{},[]int{}
			}

			for _,ac := range ret.ctxs {
				if len(ac.afterAQueue) < evaluation_period {
					continue
				}

				for i := 0; i < evaluation_period; i++ {
					_=<-ac.afterAQueue
				}
			}

			for _,ac := range ret.ctxs {
				if len(ac.afterNQueue) < evaluation_period {
					continue
				}

				for i := 0; i < evaluation_period; i++ {
					_=<-ac.afterNQueue
				}
			}

			ret.mut.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}
	process := func() {
		for {
			ret.mut.Lock()
			for _,ac := range ret.ctxs {
				if len(ac.vqueue) == 0 {
					continue
				}

				vid:=<-ac.vqueue
				if vid.DecodedFps == 0 { 
					ret.triggerVideoReset() 
					data,_ :=json.Marshal(struct{ Type string `json:"type"` }{ Type: "FRAME_LOSS", });
					ret.out<-string(data)
				}
				ac.afterVQueue<-vid
			}

			for _,ac := range ret.ctxs {
				if len(ac.aqueue) == 0 {
					continue
				}

				data:=<-ac.aqueue
				ac.afterAQueue<-data
			}

			for _,ac := range ret.ctxs {
				if len(ac.nqueue) == 0 {
					continue
				}

				data:=<-ac.nqueue
				ac.afterNQueue<-data
			}


			ret.mut.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}

	preprocess := func() {
		for {
			metricRaw := <-ret.in
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
	}

	go preprocess()
	go process()
	go afterprocess()

	return ret
}

func (ads *AdsMultiCtxs) handleNewContext(source string) {
	ads.mut.Lock()
	ads.ctxs[source] = &AdsCtx{
		aqueue: make(chan *AudioMetric,queue_size),
		vqueue: make(chan *VideoMetric,queue_size),
		nqueue: make(chan *NetworkMetric,queue_size),
		afterNQueue: make(chan *NetworkMetric,queue_size),
		afterAQueue: make(chan *AudioMetric,queue_size),
		afterVQueue: make(chan *VideoMetric,queue_size),
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
		Type: "VIDEO",
		DecodedFps : 					(metric.FramesDecoded 	- lastVideoMetric.FramesDecoded) 			/ (timedif / float64(time.Second.Milliseconds())),
		ReceivedFps : 					(metric.FramesReceived 	- lastVideoMetric.FramesReceived) 			/ (timedif / float64(time.Second.Milliseconds())),
		VideoBandwidthConsumption : 	(metric.BytesReceived 	- lastVideoMetric.BytesReceived) 			/ (timedif / float64(time.Second.Milliseconds())),
		DecodeTimePerFrame : 			(metric.TotalDecodeTime - lastVideoMetric.TotalDecodeTime) 			/ (metric.FramesDecoded   - lastVideoMetric.FramesDecoded),
		VideoPacketsLostpercent : 		(metric.PacketsLost 	- lastVideoMetric.PacketsLost) 			    / (metric.PacketsReceived - lastVideoMetric.PacketsReceived),
		BufferedFrame: 					(metric.FramesReceived  - metric.FramesDecoded - metric.FramesDropped),
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
		Type: "NETWORK",
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
		Type: "AUDIO",
		AudioBandwidthConsumption : (metric.BytesReceived - lastAudioMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds())),
		Timestamp: metric.Timestamp,
	}

	if len(ads.ctxs[metric.Source].aqueue) < queue_size{
		ads.ctxs[metric.Source].aqueue<-audio
	}
}

func (ads *AdsMultiCtxs) Send(msg string) {
	ads.in <- msg
}

func (ads *AdsMultiCtxs) Recv() string {
	return <-ads.out
}
