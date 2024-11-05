package proxy

import (
	"fmt"
	"time"

	webrtclib "github.com/pion/webrtc/v4"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/signalling"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
	"github.com/thinkonmay/thinkremote-rtchub/webrtc"
)

type Proxy struct {
	listeners []listener.Listener

	chan_conf        datachannel.IDatachannel
	signallingClient signalling.Signalling
	webrtcClient     *webrtc.WebRTCClient

	stop chan bool
}

func InitWebRTCProxy(grpc_conf signalling.Signalling,
	webrtc_conf *config.WebRTCConfig,
	chan_conf datachannel.IDatachannel,
	lis []listener.Listener,
	onTrack webrtc.OnTrackFunc,
	onIDR webrtc.OnIDRFunc,
) (err error) {
	fmt.Printf("started proxy\n")
	proxy := &Proxy{
		chan_conf:        chan_conf,
		signallingClient: grpc_conf,
		listeners:        lis,
		stop:             make(chan bool, 2),
	}

	if proxy.webrtcClient, err = webrtc.InitWebRtcClient(onTrack, onIDR, *webrtc_conf); err != nil {
		return
	}

	thread.SafeSelect(proxy.stop, proxy.webrtcClient.GatherStateChange(), func(_state interface{}) {
		state := _state.(webrtclib.ICEGatheringState)
		switch state {
		case webrtclib.ICEGatheringStateGathering:
		case webrtclib.ICEGatheringStateComplete:
		case webrtclib.ICEGatheringStateUnknown:
		}
	})
	thread.SafeSelect(proxy.stop, proxy.webrtcClient.ConnectionStateChange(), func(_state interface{}) {
		state := _state.(webrtclib.ICEConnectionState)
		switch state {
		case webrtclib.ICEConnectionStateConnected:
		case webrtclib.ICEConnectionStateCompleted:
		case webrtclib.ICEConnectionStateClosed:
			proxy.Stop()
		case webrtclib.ICEConnectionStateFailed:
			proxy.Stop()
		case webrtclib.ICEConnectionStateDisconnected:
			proxy.Stop()
		}
	})
	thread.SafeSelect(proxy.stop, proxy.webrtcClient.OnLocalICE(), func(ice interface{}) {
		proxy.signallingClient.SendICE(ice.(*webrtclib.ICECandidateInit))
	})
	thread.SafeSelect(proxy.stop, proxy.webrtcClient.OnLocalSDP(), func(sdp interface{}) {
		proxy.signallingClient.SendSDP(sdp.(*webrtclib.SessionDescription))
	})
	proxy.signallingClient.OnICE(func(i *webrtclib.ICECandidateInit) {
		proxy.webrtcClient.OnIncomingICE(i)
	})
	proxy.signallingClient.OnSDP(func(i *webrtclib.SessionDescription) {
		proxy.webrtcClient.OnIncominSDP(i)
	})

	return proxy.start()
}

func (proxy *Proxy) start() error {
	proxy.webrtcClient.RegisterDataChannels(proxy.chan_conf)
	proxy.webrtcClient.Listen(proxy.listeners)
	defer proxy.webrtcClient.StopSignaling()

	success := make(chan bool, 2)
	proxy.signallingClient.WaitForEnd(func() {
		success <- true
	})
	thread.SafeThread(func() {
		time.Sleep(time.Second * 60)
		success <- false
	})

	if !<-success {
		return fmt.Errorf("application exchange signaling timeout, closing")
	} else {
		fmt.Println("webrtc connection established successfully")
		return nil
	}
}

func (prox *Proxy) Stop() {
	fmt.Println("proxy stopped")
	prox.webrtcClient.Close()
	prox.signallingClient.Stop()
	thread.TriggerStop(prox.stop)
}
