package signalling

import "github.com/pion/webrtc/v3"

type OnIceFunc func (*webrtc.ICECandidateInit) 

type OnSDPFunc func (*webrtc.SessionDescription) 

type Signalling interface {
	SendSDP(*webrtc.SessionDescription) error;
	SendICE(*webrtc.ICECandidateInit) error;
	OnICE(OnIceFunc);
	OnSDP(OnSDPFunc);
}