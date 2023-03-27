package signalling

import (
	"github.com/pion/webrtc/v3"
)

type DeviceSelection struct {
	SoundCard string `json:"soundcard"`
	Monitor   string `json:"monitor"`
	Bitrate   int    `json:"bitrate"`
	Framerate int    `json:"framerate"`
}

type OnIceFunc func(*webrtc.ICECandidateInit)

type OnSDPFunc func(*webrtc.SessionDescription)

type Signalling interface {
	SendSDP(*webrtc.SessionDescription) error
	SendICE(*webrtc.ICECandidateInit) error

	OnICE(OnIceFunc)
	OnSDP(OnSDPFunc)

	WaitForStart()
	WaitForEnd()

	Stop()
}
