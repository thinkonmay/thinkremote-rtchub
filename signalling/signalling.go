package signalling

import "github.com/pion/webrtc/v3"

type Signalling interface {
	SendSDP(*webrtc.SessionDescription);
	SendICE(*webrtc.ICECandidate);
	OnICE() *webrtc.ICECandidate;
	OnSDP() *webrtc.SessionDescription;
}