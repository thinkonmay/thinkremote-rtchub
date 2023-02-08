package proxy

import (
	"fmt"
	"time"

	webrtclib "github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/signalling"
	grpc "github.com/thinkonmay/thinkremote-rtchub/signalling/gRPC"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
	"github.com/thinkonmay/thinkremote-rtchub/util/tool"
	"github.com/thinkonmay/thinkremote-rtchub/webrtc"
)

type Proxy struct {
	listeners []listener.Listener

	chan_conf *config.DataChannelConfig

	signallingClient signalling.Signalling
	webrtcClient     *webrtc.WebRTCClient

	Shutdown chan bool
}

func InitWebRTCProxy(sock *config.WebsocketConfig,
	grpc_conf *config.GrpcConfig,
	webrtc_conf *config.WebRTCConfig,
	chan_conf *config.DataChannelConfig,
	devices *tool.MediaDevice,
	lis []listener.Listener,
	onTrack webrtc.OnTrackFunc,
	deviceSelect signalling.OnDeviceSelectFunc,
) (proxy *Proxy, err error) {

	fmt.Printf("started proxy\n")
	proxy = &Proxy{
		Shutdown:  make(chan bool),
		chan_conf: chan_conf,
		listeners: lis,
	}

	if grpc_conf != nil {
		if proxy.signallingClient, err = grpc.InitGRPCClient(grpc_conf, devices, webrtc_conf, proxy.Shutdown); err != nil {
			return
		}
	} else if sock != nil {
		err = fmt.Errorf("Unimplemented")
		return
	} else {
		err = fmt.Errorf("Unimplemented")
		return
	}

	proxy.handleTimeout()
	proxy.signallingClient.OnDeviceSelect(deviceSelect)
	if proxy.webrtcClient, err = webrtc.InitWebRtcClient(onTrack, *webrtc_conf); err != nil {
		return
	}

	go func() {
		for {
			state := proxy.webrtcClient.GatherStateChange()
			switch state {
			case webrtclib.ICEGathererStateGathering:
			case webrtclib.ICEGathererStateComplete:
			case webrtclib.ICEGathererStateClosed:
			}
		}
	}()
	go func() {
		for {
			state := proxy.webrtcClient.ConnectionStateChange()

			switch state {
			case webrtclib.ICEConnectionStateConnected:
				proxy.webrtcClient.Listen(proxy.listeners)
			case webrtclib.ICEConnectionStateClosed:
			case webrtclib.ICEConnectionStateFailed:
				proxy.Stop()
			case webrtclib.ICEConnectionStateDisconnected:
				proxy.Stop()
			}
		}
	}()

	go func() {
		for {
			proxy.signallingClient.SendICE(proxy.webrtcClient.OnLocalICE())
		}
	}()
	go func() {
		for {
			proxy.signallingClient.SendSDP(proxy.webrtcClient.OnLocalSDP())
		}
	}()
	proxy.signallingClient.OnICE(func(i *webrtclib.ICECandidateInit) {
		proxy.webrtcClient.OnIncomingICE(i)
	})
	proxy.signallingClient.OnSDP(func(i *webrtclib.SessionDescription) {
		proxy.webrtcClient.OnIncominSDP(i)
	})
	return
}

func (proxy *Proxy) handleTimeout() {
	start := make(chan bool, 2)
	go func() {
		proxy.signallingClient.WaitForConnected()
		time.Sleep(30 * time.Second)
		start <- false
	}()
	go func() {
		proxy.signallingClient.WaitForStart()
		start <- true
	}()
	go func() {
		if <-start {
			proxy.Start()
		} else {
			fmt.Printf("application start timeout, closing\n")
			proxy.Stop()
		}
	}()
}
func (prox *Proxy) Start() {
	prox.webrtcClient.RegisterDataChannel(prox.chan_conf)
}

func (prox *Proxy) Stop() {
	prox.webrtcClient.Close()
	prox.signallingClient.Stop()
	for _, lis := range prox.listeners {
		lis.Close()
	}
	prox.Shutdown <- true
}
