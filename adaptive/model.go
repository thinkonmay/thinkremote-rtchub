package adaptive


type VideoMetrics struct {
    FrameWidth float64   `json:"frameWidth"`
    FrameHeight float64  `json:"frameHeight"`

    CodecId string  `json:"codecId"`
    DecoderImplementation string  `json:"decoderImplementation"`

    TotalSquaredInterFrameDelay float64  `json:"totalSquaredInterFrameDelay"`
    TotalInterFrameDelay float64  `json:"totalInterFrameDelay"`

    TotalProcessingDelay float64  `json:"totalProcessingDelay"`
    TotalDecodeTime float64  `json:"totalDecodeTime"`
    
    KeyFramesDecoded float64  `json:"keyFramesDecoded"`
    FramesDecoded float64  `json:"framesDecoded"`
    FramesReceived float64  `json:"framesReceived"`
    
    HeaderBytesReceived float64  `json:"headerBytesReceived"`
    BytesReceived float64  `json:"bytesReceived"`
    PacketsReceived float64  `json:"packetsReceived"`
    
    FramesDropped float64  `json:"framesDropped"`
    PacketsLost float64  `json:"packetsLost"`

    JitterBufferEmittedCount float64  `json:"jitterBufferEmittedCount"`
    JitterBufferDelay float64  `json:"jitterBufferDelay"`
    Jitter float64  `json:"jitter"`

    Timestamp float64  `json:"timestamp"`
}
type AudioMetric struct {
    AudioLevel float64    `json:"audioLevel"`
    TotalAudioEnergy float64    `json:"totalAudioEnergy"`
    TotalSamplesReceived float64    `json:"totalSamplesReceived"`
    HeaderBytesReceived float64    `json:"headerBytesReceived"`
    BytesReceived float64    `json:"bytesReceived"`
    PacketsReceived float64    `json:"packetsReceived"`
    PacketsLost float64    `json:"packetsLost"`
    Timestamp float64    `json:"timestamp"`
}

type NetworkMetric struct {
    PacketsReceived float64   `json:"packetsReceived"`
    PacketsSent float64   `json:"packetsSent"`
    BytesSent float64   `json:"bytesSent"`
    BytesReceived float64   `json:"bytesReceived"`
    AvailableIncomingBitrate float64   `json:"availableIncomingBitrate"`
    AvailableOutgoingBitrate float64   `json:"availableOutgoingBitrate"`
    CurrentRoundTripTime float64   `json:"currentRoundTripTime"`
    TotalRoundTripTime float64   `json:"totalRoundTripTime"`
    LocalPort float64   `json:"localPort"`
    RemotePort float64   `json:"remotePort"`
    Priority float64   `json:"priority"`
    Timestamp float64   `json:"timestamp"`
    LocalIP  string   `json:"localIP"`
    RemoteIP  string   `json:"remoteIP"`
}