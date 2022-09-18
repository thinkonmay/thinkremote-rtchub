package proxy

import (
	"fmt"

	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster"
	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster/dummy"
	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster/udp"

	"github.com/OnePlay-Internet/webrtc-proxy/listener"
	"github.com/OnePlay-Internet/webrtc-proxy/signalling"
	grpc "github.com/OnePlay-Internet/webrtc-proxy/signalling/gRPC"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/OnePlay-Internet/webrtc-proxy/webrtc"
	webrtclib "github.com/pion/webrtc/v3"
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
	br_conf []*config.BroadcasterConfig,
	chan_conf *config.DataChannelConfig,
	lis []listener.Listener) (proxy *Proxy, err error) {
	proxy = &Proxy{}
	proxy.chan_conf = chan_conf
	proxy.listeners = lis

	proxy.Shutdown = make(chan bool)
	fmt.Printf("added listener\n")

	if grpc_conf != nil {
		var rpc grpc.GRPCclient
		rpc, err = grpc.InitGRPCClient(grpc_conf)
		if err != nil {
			return
		}
		proxy.signallingClient = &rpc
	} else if sock != nil {
		err = fmt.Errorf("Unimplemented")
		return
	} else {
		err = fmt.Errorf("Unimplemented")
		return
	}

	proxy.webrtcClient, err = webrtc.InitWebRtcClient(func(tr *webrtclib.TrackRemote) (br broadcaster.Broadcaster, err error) {
		for _, conf := range br_conf {
			if tr.Codec().MimeType == conf.Codec {
				return udp.NewUDPBroadcaster(conf)
			} else {
				fmt.Printf("no available codec handler, using dummy sink\n");
				return dummy.NewDummyBroadcaster(conf)
			}

		}

		err = fmt.Errorf("unimplemented broadcaster")
		return
	}, *webrtc_conf)
	if err != nil {
		panic(err)
	}

	go func() {
		proxy.signallingClient.WaitForStart()
		proxy.Start()
	}()
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

func (prox *Proxy) Start() {
	prox.webrtcClient.RegisterDataChannel(prox.chan_conf)
	prox.webrtcClient.Listen(prox.listeners)
}

func (prox *Proxy) Stop() {
	prox.webrtcClient.Close()
	prox.signallingClient.Stop()
	for _, lis := range prox.listeners {
		lis.Close()
	}
	prox.Shutdown <- true
}
