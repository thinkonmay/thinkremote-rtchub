package adaptive

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)


const (
	queue_size = 1000
	evaluation_period = 10
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
	in  chan datachannel.Msg
	out chan datachannel.Msg

	triggerVideoReset func()
	bitrateChangeFunc func(bitrate int)



	mut *sync.Mutex
	ctxs map[string]*AdsCtx
}

type VideoData struct{
	Type string `json:"type"`
	ReceiveFps []int `json:"receivefps"`
	DecodeFps  []int `json:"decodefps"`
	PacketsLoss []int `json:"packetloss"`
	Bandwidth []int `json:"bandwidth"`
	Buffer   []int `json:"buffer"`
}


func NewAdsContext(BitrateCallback func(bitrate int),
	IDRcallback func()) datachannel.DatachannelConsumer {
	ret := &AdsMultiCtxs{
		in:  make(chan datachannel.Msg,queue_size),
		out: make(chan datachannel.Msg,queue_size),

		triggerVideoReset: IDRcallback,
		bitrateChangeFunc: BitrateCallback,
		mut: &sync.Mutex{},
		ctxs: make(map[string]*AdsCtx),
	}

	video_reset := func () {
		ret.triggerVideoReset() 
		form := struct{ 
			Type string `json:"type"` 
		}{ 
			Type: "FRAME_LOSS", 
		}
		data,_ :=json.Marshal(form)
		ret.SendToAll(string(data))
	}

	afterprocess := func() {
		for {
			ret.mut.Lock()
			for id,ac := range ret.ctxs {
				if len(ac.afterVQueue) < evaluation_period {
					continue
				}

				value := VideoData{
					Type: "VIDEO",
					DecodeFps: []int{},
					ReceiveFps: []int{},
					Bandwidth : []int{},
					PacketsLoss : []int{},
					Buffer : []int{},
				}

				for i := 0; i < evaluation_period; i++ {
					vid:=<-ac.afterVQueue
					value.DecodeFps  = append(value.DecodeFps  , int(vid.DecodedFps))
					value.ReceiveFps = append(value.ReceiveFps , int(vid.ReceivedFps))
					value.Bandwidth  = append(value.Bandwidth  , int(vid.VideoBandwidthConsumption * 8 / 1024))
					value.PacketsLoss= append(value.PacketsLoss, int(vid.VideoPacketsLostpercent * 100))
					value.Buffer     = append(value.Buffer     , int(vid.BufferedFrame))
				}


				data,_ :=json.Marshal(&value);
				ret.out<-datachannel.Msg{
					Id : id,
					Msg : string(data),
				}
			}

			for _,ac := range ret.ctxs {
				if len(ac.afterAQueue) < evaluation_period {
					continue
				}

				// TODO
				for i := 0; i < evaluation_period; i++ {
					_=<-ac.afterAQueue
				}
			}

			for _,ac := range ret.ctxs {
				if len(ac.afterNQueue) < evaluation_period {
					continue
				}

				// TODO
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
				if vid.DecodedFps == 0 { video_reset() }
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
			in := <-ret.in
			metricRaw := in.Msg
			var out map[string]interface{}
			json.Unmarshal([]byte(metricRaw), &out)

			switch out["type"] {
			case "video":
				video := VideoMetricRaw{}
				json.Unmarshal([]byte(metricRaw), &video)
				ret.handleVideoMetric(in.Id,&video)
			case "audio":
				audio := AudioMetricRaw{}
				json.Unmarshal([]byte(metricRaw), &audio)
				ret.handleAudioMetric(in.Id,&audio)
			case "network":
				network := NetworkMetricRaw{}
				json.Unmarshal([]byte(metricRaw), &network)
				ret.handleNetworkMetric(in.Id,&network)
			}
		}
	}

	go preprocess()
	go process()
	go afterprocess()

	return ret
}


func (ads *AdsMultiCtxs) handleVideoMetric(id string,metric *VideoMetricRaw) {
	if ads.ctxs[id] == nil {
		return
	}

	last := ads.ctxs[id].last
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


	if len(ads.ctxs[id].vqueue) < queue_size {
		ads.ctxs[id].vqueue<-video
	}
}

func (ads *AdsMultiCtxs) handleNetworkMetric(id string,metric *NetworkMetricRaw) {
	if ads.ctxs[id] == nil {
		return
	}
	last := ads.ctxs[id].last
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
	
	if len(ads.ctxs[id].nqueue) < queue_size{
		ads.ctxs[id].nqueue<-network
	}
}

func (ads *AdsMultiCtxs) handleAudioMetric(id string,metric *AudioMetricRaw) {
	if ads.ctxs[id] == nil {
		return
	}

	last := ads.ctxs[id].last
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

	if len(ads.ctxs[id].aqueue) < queue_size{
		ads.ctxs[id].aqueue<-audio
	}
}

func (ads *AdsMultiCtxs) Send(id string,msg string) {
	ads.in <- datachannel.Msg{
		Id: id,
		Msg: msg,
	}
}

func (ads *AdsMultiCtxs) Recv() (string,string) {
	out := <-ads.out
	return out.Id,out.Msg
}

func (ads *AdsMultiCtxs) SetContext(ids []string) {
	for name,_ := range ads.ctxs {
		found := false
		for _,id := range ids {
			if id == name {
				found = true
			}
		}

		if !found {
			ads.deleteContext(name)
		}
	}

	for _,id := range ids {
		found := false
		for name,_ := range ads.ctxs {
			if id == name {
				found = true
			}
		}

		if !found {
			ads.handleNewContext(id)
		}
	}
}


func (ads *AdsMultiCtxs) SendToAll(msg string) {
	for k := range ads.ctxs {
		ads.out<-datachannel.Msg{
			Msg: msg,
			Id: k,
		}
	}
}


func (ads *AdsMultiCtxs) deleteContext(name string) {
	ads.mut.Lock()
	delete(ads.ctxs,name)
	ads.mut.Unlock()
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