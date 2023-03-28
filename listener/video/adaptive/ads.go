package adaptive

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)


const (
	queue_size = 100000
)

type AdsCtx struct {
	vqueue chan *VideoMetric
	aqueue chan *AudioMetric
	nqueue chan *NetworkMetric
}

type AdsMultiCtxs struct {
	In  chan string
	Out chan string

	triggerVideoReset func()
	bitrateChangeFunc func(bitrate int)

	last struct {
		audio   *AudioMetricRaw
		video   *VideoMetricRaw
		network *NetworkMetricRaw
	}


	mut *sync.Mutex
	ctxs map[string]*AdsCtx
}

func NewAdsContext(BitrateCallback func(bitrate int),
	IDRcallback func()) datachannel.DatachannelConsumer {
	ret := &AdsMultiCtxs{
		In:  make(chan string),
		Out: make(chan string),

		triggerVideoReset: IDRcallback,
		bitrateChangeFunc: BitrateCallback,
		last: struct {
			audio   *AudioMetricRaw
			video   *VideoMetricRaw
			network *NetworkMetricRaw
		}{
			audio:   nil,
			video:   nil,
			network: nil,
		},
		mut: &sync.Mutex{},
		ctxs: make(map[string]*AdsCtx),
	}

	go func() {
		count := 1
		for {
			ret.mut.Lock()
			for _,ac := range ret.ctxs {
				if len(ac.vqueue) < count {
					continue
				}


				fpses := []int{}
				for i := 0; i < count; i++ {
					vid:=<-ac.vqueue
					fpses = append(fpses, int(vid.DecodedFps))
				}

				// v_fps := fpses/count
				// if v_fps < 25 {
				// 	ret.triggerVideoReset()
				// }
			}
			ret.mut.Unlock()
			time.Sleep(1 * time.Second)
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
				break
			case "audio":
				audio := AudioMetricRaw{}
				json.Unmarshal([]byte(metricRaw), &audio)
				ret.handleAudioMetric(&audio)
				break
			case "network":
				network := NetworkMetricRaw{}
				json.Unmarshal([]byte(metricRaw), &network)
				ret.handleNetworkMetric(&network)
				break
			}
		}
	}()

	return ret
}

func (ads *AdsMultiCtxs) handleVideoMetric(metric *VideoMetricRaw) {
	lastVideoMetric := ads.last.video
	if lastVideoMetric == nil {
		ads.last.video = metric
		return
	}

	timedif := (metric.Timestamp - lastVideoMetric.Timestamp) //nanosecond

	ads.last.video = metric
	video := &VideoMetric{
		DecodedFps : (metric.FramesDecoded - lastVideoMetric.FramesDecoded) / (timedif / float64(time.Second.Milliseconds())),
		ReceivedFps : (metric.FramesReceived - lastVideoMetric.FramesReceived) / (timedif / float64(time.Second.Milliseconds())),
		VideoBandwidthConsumption : (metric.BytesReceived - lastVideoMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds())),
		DecodeTimePerFrame : (metric.TotalDecodeTime - lastVideoMetric.TotalDecodeTime) / (metric.FramesDecoded - lastVideoMetric.FramesDecoded),
		VideoPacketsLostpercent : (metric.PacketsLost - lastVideoMetric.PacketsLost) / (metric.PacketsReceived - lastVideoMetric.PacketsReceived),
		VideoJitter : metric.Jitter,
		VideoJitterBufferDelay : metric.JitterBufferDelay,
		Timestamp: metric.Timestamp,
	}

	if ads.ctxs[metric.Source] == nil {
		ads.mut.Lock()
		ads.ctxs[metric.Source] = &AdsCtx{
			aqueue: make(chan *AudioMetric,queue_size),
			vqueue: make(chan *VideoMetric,queue_size),
			nqueue: make(chan *NetworkMetric,queue_size),
		}
		ads.mut.Unlock()
	}

	ads.ctxs[metric.Source].vqueue<-video

}

func (ads *AdsMultiCtxs) handleNetworkMetric(metric *NetworkMetricRaw) {
	lastNetworkMetric := ads.last.network
	if lastNetworkMetric == nil {
		ads.last.network = metric
		return
	}

	timedif := metric.Timestamp - lastNetworkMetric.Timestamp //nanosecond

	ads.last.network = lastNetworkMetric
	network := &NetworkMetric{
		TotalBandwidthConsumption : (metric.BytesReceived - lastNetworkMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds())),
		RTT : metric.CurrentRoundTripTime * float64(time.Second.Nanoseconds()),
		AvailableIncomingBandwidth : metric.AvailableIncomingBitrate,
		Timestamp: metric.Timestamp,
	}

	if ads.ctxs[metric.Source] == nil {
		ads.mut.Lock()
		ads.ctxs[metric.Source] = &AdsCtx{
			aqueue: make(chan *AudioMetric,queue_size),
			vqueue: make(chan *VideoMetric,queue_size),
			nqueue: make(chan *NetworkMetric,queue_size),
		}
		ads.mut.Unlock()
	}

	ads.ctxs[metric.Source].nqueue<-network
}

func (ads *AdsMultiCtxs) handleAudioMetric(metric *AudioMetricRaw) {
	lastAudioMetric := ads.last.audio
	if lastAudioMetric == nil {
		ads.last.audio = metric
		return
	}

	timedif := metric.Timestamp - lastAudioMetric.Timestamp //nanosecond


	ads.last.audio = lastAudioMetric
	audio:=&AudioMetric{
		AudioBandwidthConsumption : (metric.BytesReceived - lastAudioMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds())),
		Timestamp: metric.Timestamp,
	}

	if ads.ctxs[metric.Source] == nil {
		ads.mut.Lock()
		ads.ctxs[metric.Source] = &AdsCtx{
			aqueue: make(chan *AudioMetric,queue_size),
			vqueue: make(chan *VideoMetric,queue_size),
			nqueue: make(chan *NetworkMetric,queue_size),
		}
		ads.mut.Unlock()
	}

	ads.ctxs[metric.Source].aqueue<-audio
}

func (ads *AdsMultiCtxs) Send(msg string) {
	ads.In <- msg
}

func (ads *AdsMultiCtxs) Recv() string {
	return <-ads.Out
}
