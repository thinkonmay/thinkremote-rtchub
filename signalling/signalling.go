package signalling

import (
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/pion/webrtc/v3"
)

type OnIceFunc func (*webrtc.ICECandidateInit) 

type OnSDPFunc func (*webrtc.SessionDescription) 

type OnDeviceSelectFunc func (tool.Monitor, tool.Soundcard, int) 

type Signalling interface {
	SendSDP(*webrtc.SessionDescription) error;
	SendICE(*webrtc.ICECandidateInit) error;

	OnICE(OnIceFunc);
	OnSDP(OnSDPFunc);
	OnDeviceSelect(OnDeviceSelectFunc);

	WaitForStart();
	Stop();
}


