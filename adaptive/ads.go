package adaptive

import (
	"encoding/json"
	"fmt"
	"time"
)


type AdaptiveContext struct {
	In         chan string

	// ctx unsafe.Pointer
	triggerVideoReset func()
	bitrateChangeFunc func(bitrate int)

	last struct {
		audio   *AudioMetric
		video   *VideoMetrics
		network *NetworkMetric
	}
}

func NewAdsContext(InChan chan string,
				   BitrateChangeFunc func(bitrate int),
				   TriggerVideoReset func(),
				   ) *AdaptiveContext {
	ret := &AdaptiveContext{
		In:         InChan,
		// ctx:        C.new_ads_context(),
		triggerVideoReset: TriggerVideoReset,
		bitrateChangeFunc: BitrateChangeFunc,
		last: struct {
			audio   *AudioMetric
			video   *VideoMetrics
			network *NetworkMetric
		}{
			audio:   nil,
			video:   nil,
			network: nil,
		},
	}

	go func() {
		for {
			metricRaw := <-ret.In
			var out map[string]interface{}
			json.Unmarshal([]byte(metricRaw), &out)

			switch out["type"] {
			case "video":
				video := VideoMetrics{}
				json.Unmarshal([]byte(metricRaw), &video)
				ret.handleVideoMetric(&video)
			case "audio":
				audio := AudioMetric{}
				json.Unmarshal([]byte(metricRaw), &audio)
				ret.handleAudioMetric(&audio)
			case "network":
				network := NetworkMetric{}
				json.Unmarshal([]byte(metricRaw), &network)
				ret.handleNetworkMetric(&network)
			}
		}
	}()


	return ret
}

func (ads *AdaptiveContext) handleVideoMetric(metric *VideoMetrics) {
	lastVideoMetric := ads.last.video
	if lastVideoMetric == nil {
		ads.last.video = metric
		return
	}

	timedif := (metric.Timestamp - lastVideoMetric.Timestamp) //nanosecond
	decodedFps := (metric.FramesDecoded - lastVideoMetric.FramesDecoded) / (timedif / float64(time.Second.Milliseconds()))
	if decodedFps < 25 { // TODO 
		ads.triggerVideoReset()	
	}


	receivedFps := (metric.FramesReceived - lastVideoMetric.FramesReceived) / (timedif / float64(time.Second.Milliseconds()))
	videoBandwidthConsumption := (metric.BytesReceived - lastVideoMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds()))
	decodeTimePerFrame := (metric.TotalDecodeTime - lastVideoMetric.TotalDecodeTime) / (metric.FramesDecoded - lastVideoMetric.FramesDecoded)
	videoPacketsLostpercent := (metric.PacketsLost - lastVideoMetric.PacketsLost) / (metric.PacketsReceived - lastVideoMetric.PacketsReceived)
	videoJitter := metric.Jitter
	videoJitterBufferDelay := metric.JitterBufferDelay

	fmt.Printf("frame_decoded_per_second %f\n", decodedFps)
	fmt.Printf("frame_received_per_second %f\n", receivedFps)
	fmt.Printf("video_incoming_bandwidth_consumption %f\n", videoBandwidthConsumption)
	fmt.Printf("decode_time_per_frame %f\n", decodeTimePerFrame)
	fmt.Printf("video_packets_lost %f\n", videoPacketsLostpercent)
	fmt.Printf("video_jitter %f\n", videoJitter)
	fmt.Printf("video_jitter_buffer_delay %f\n", videoJitterBufferDelay)

	ads.last.video = metric
}

func (ads *AdaptiveContext) handleNetworkMetric(metric *NetworkMetric) {
	lastNetworkMetric := ads.last.network
	if lastNetworkMetric == nil {
		ads.last.network = metric
		return
	}

	timedif := metric.Timestamp - lastNetworkMetric.Timestamp //nanosecond

	totalBandwidthConsumption := (metric.BytesReceived - lastNetworkMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds()))
	RTT := metric.CurrentRoundTripTime * float64(time.Second.Nanoseconds())
	availableIncomingBandwidth := metric.AvailableIncomingBitrate

	fmt.Printf("rtt %f\n", RTT);
	fmt.Printf("total_incoming_bandwidth_consumption %f\n", totalBandwidthConsumption);
	fmt.Printf("available_incoming_bandwidth %f\n", availableIncomingBandwidth);

	ads.last.network = lastNetworkMetric
}

func (ads *AdaptiveContext) handleAudioMetric(metric *AudioMetric) {
	lastAudioMetric := ads.last.audio
	if lastAudioMetric == nil {
		ads.last.audio = metric
		return
	}



	timedif := metric.Timestamp - lastAudioMetric.Timestamp //nanosecond

	audioBandwidthConsumption := (metric.BytesReceived - lastAudioMetric.BytesReceived) / (timedif / float64(time.Second.Milliseconds()))
	fmt.Printf("audio_incoming_bandwidth_consumption %f\n", audioBandwidthConsumption)

	ads.last.audio = lastAudioMetric
}
