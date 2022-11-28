package adaptive

import (
	"encoding/json"
	"testing"
	"time"
)


var sampleNetwork NetworkMetric = NetworkMetric{
    Type: "network",
	LocalIP: 					"",
	RemoteIP: 					"192.168.2.6",
	LocalPort: 					63411,
	RemotePort: 				52622,
	PacketsReceived: 			8003,
	PacketsSent: 				127,
	BytesSent: 					26554,
	BytesReceived: 				8791712,
	AvailableIncomingBitrate: 	1569803,
	AvailableOutgoingBitrate: 	300000,
	CurrentRoundTripTime: 		0.100,
	TotalRoundTripTime: 		0.064,
	Priority: 					9079290933605827000,
	Timestamp: 					66962249529,
}

var sampleAudio AudioMetric = AudioMetric{
    Type: "audio",
	TotalAudioEnergy:		0,
	TotalSamplesReceived: 	0,
	HeaderBytesReceived:	0,
	BytesReceived:			0,
	PacketsReceived:		0,
	PacketsLost:			0,
	Timestamp:				1669622495297,
}
var sampleVideo VideoMetrics = VideoMetrics{
    Type: "video",
	FrameWidth:						3840,
	FrameHeight:					2400,
	CodecId:						"CIT01_102_level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
	DecoderImplementation:			"ExternalDecoder (VDAVideoDecoder)",
	TotalSquaredInterFrameDelay:	0.6084750000000014,
	TotalInterFrameDelay:			15.441000000000018,
	TotalProcessingDelay:			59.807531999999995,
	TotalDecodeTime:				2.556301,
	KeyFramesDecoded:				2,
	FramesDecoded:					460,
	FramesReceived:					466,
	HeaderBytesReceived:			95136,
	BytesReceived:					8612552,
	PacketsReceived:				7928,
	FramesDropped:					1,
	PacketsLost:					0,
	JitterBufferEmittedCount:		450,
	JitterBufferDelay:				41.631,
	Jitter:							0.003,
	Timestamp:						1669622495297,
}



func TestAdaptiveStream(t *testing.T) {
	strChan := make(chan string)
	NewAdsContext(strChan);

	go func() {
		for {
			audioBytes,_ := json.Marshal(sampleAudio);
			strChan <- string(audioBytes);
			time.Sleep(time.Second);

			sampleAudio.Timestamp += float64(time.Second.Nanoseconds())
		}
	}()
	go func() {
		for {
			videoBytes,_ := json.Marshal(sampleVideo);
			strChan <- string(videoBytes);
			time.Sleep(time.Second);

			sampleVideo.FramesDecoded += 60
			sampleVideo.FramesReceived += 56
			sampleVideo.Timestamp += float64(time.Second.Nanoseconds())
		}
	}()
	go func() {
		for {
			networkBytes,_ := json.Marshal(sampleNetwork);
			strChan <- string(networkBytes);
			time.Sleep(time.Second);

			sampleNetwork.Timestamp += float64(time.Second.Nanoseconds())
		}
	}()

	time.Sleep(100 * time.Second);
}