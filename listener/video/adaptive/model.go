package adaptive

type VideoMetricRaw struct {
	Type        string  `json:"type"`
	FrameWidth  float64 `json:"frameWidth"`
	FrameHeight float64 `json:"frameHeight"`

	CodecId               string `json:"codecId"`
	DecoderImplementation string `json:"decoderImplementation"`

	TotalSquaredInterFrameDelay float64 `json:"totalSquaredInterFrameDelay"`
	TotalInterFrameDelay        float64 `json:"totalInterFrameDelay"`

	TotalProcessingDelay float64 `json:"totalProcessingDelay"`
	TotalDecodeTime      float64 `json:"totalDecodeTime"`

	KeyFramesDecoded float64 `json:"keyFramesDecoded"`
	FramesDecoded    float64 `json:"framesDecoded"`
	FramesReceived   float64 `json:"framesReceived"`

	HeaderBytesReceived float64 `json:"headerBytesReceived"`
	BytesReceived       float64 `json:"bytesReceived"`
	PacketsReceived     float64 `json:"packetsReceived"`

	FramesDropped float64 `json:"framesDropped"`
	PacketsLost   float64 `json:"packetsLost"`

	JitterBufferEmittedCount float64 `json:"jitterBufferEmittedCount"`
	JitterBufferDelay        float64 `json:"jitterBufferDelay"`
	Jitter                   float64 `json:"jitter"`

	Timestamp float64 `json:"timestamp"`

	Source string `json:"__source__"`
}


type VideoMetric struct {
	Timestamp float64                      `json:"timestamp"`
    DecodedFps float64                     `json:"decodedFps"`   
	ReceivedFps float64                    `json:"receivedFps"`
	VideoBandwidthConsumption float64      `json:"videoBandwidthConsumption"`
	DecodeTimePerFrame float64             `json:"decodeTimePerFrame"`
	VideoPacketsLostpercent float64        `json:"videoPacketsLostpercent"`
	VideoJitter float64                    `json:"videoJitter"`
	VideoJitterBufferDelay float64         `json:"videoJitterBufferDelay"`
}










type AudioMetricRaw struct {
	Type                 string  `json:"type"`
	AudioLevel           float64 `json:"audioLevel"`
	TotalAudioEnergy     float64 `json:"totalAudioEnergy"`
	TotalSamplesReceived float64 `json:"totalSamplesReceived"`
	HeaderBytesReceived  float64 `json:"headerBytesReceived"`
	BytesReceived        float64 `json:"bytesReceived"`
	PacketsReceived      float64 `json:"packetsReceived"`
	PacketsLost          float64 `json:"packetsLost"`
	Timestamp            float64 `json:"timestamp"`

	Source string `json:"__source__"`
}

type AudioMetric struct {
	Timestamp                 float64 `json:"timestamp"`
	AudioBandwidthConsumption float64 `json:"audioBandwidthConsumption"`
}

type NetworkMetricRaw struct {
	Type                     string  `json:"type"`
	PacketsReceived          float64 `json:"packetsReceived"`
	PacketsSent              float64 `json:"packetsSent"`
	BytesSent                float64 `json:"bytesSent"`
	BytesReceived            float64 `json:"bytesReceived"`
	AvailableIncomingBitrate float64 `json:"availableIncomingBitrate"`
	AvailableOutgoingBitrate float64 `json:"availableOutgoingBitrate"`
	CurrentRoundTripTime     float64 `json:"currentRoundTripTime"`
	TotalRoundTripTime       float64 `json:"totalRoundTripTime"`
	LocalPort                float64 `json:"localPort"`
	RemotePort               float64 `json:"remotePort"`
	Priority                 float64 `json:"priority"`
	Timestamp                float64 `json:"timestamp"`
	LocalIP                  string  `json:"localIP"`
	RemoteIP                 string  `json:"remoteIP"`

	Source string `json:"__source__"`
}


type NetworkMetric struct {
	Timestamp                 float64 `json:"timestamp"`
	TotalBandwidthConsumption float64     `json:"totalBandwidthConsumption"`
	RTT float64                           `json:"RTT"`
	AvailableIncomingBandwidth float64    `json:"availableIncomingBandwidth"`
}

