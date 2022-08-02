package proxy

import (
	"fmt"

	"github.com/pigeatgarlic/webrtc-proxy/broadcaster"
	"github.com/pigeatgarlic/webrtc-proxy/broadcaster/file"
	udpbr "github.com/pigeatgarlic/webrtc-proxy/broadcaster/udp"

	// datachannel "github.com/pigeatgarlic/webrtc-proxy/data-channel"
	"github.com/pigeatgarlic/webrtc-proxy/listener"
	gst "github.com/pigeatgarlic/webrtc-proxy/listener/gstreamer"
	"github.com/pigeatgarlic/webrtc-proxy/listener/udp"

	"github.com/pigeatgarlic/webrtc-proxy/signalling"
	grpc "github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pigeatgarlic/webrtc-proxy/webrtc"
	webrtclib "github.com/pion/webrtc/v3"
)

type Proxy struct {
	listeners []listener.Listener

	// TODO
	chan_conf *config.DataChannelConfig
	// datachannels []datachannel.Datachannel

	signallingClient signalling.Signalling
	webrtcClient     *webrtc.WebRTCClient

	Shutdown		 chan bool
}

func InitWebRTCProxy(sock *config.WebsocketConfig,
	grpc_conf *config.GrpcConfig,
	webrtc_conf *config.WebRTCConfig,
	br_conf []*config.BroadcasterConfig,
	chan_conf *config.DataChannelConfig,
	lis []*config.ListenerConfig) (proxy *Proxy, err error) {
	proxy = &Proxy{}
	proxy.chan_conf = chan_conf
	proxy.Shutdown = make(chan bool)

	fmt.Printf("added listener\n")
	for _, lis_conf := range lis {

		var Lis listener.Listener
		if lis_conf.Source == "udp" {
			udpLis, err := udp.NewUDPListener(lis_conf)
			Lis = &udpLis;
			if err != nil {
				fmt.Printf("%s\n",err.Error())
				continue;
			}
		} else if lis_conf.Source == "gstreamer" {
			Lis = gst.CreatePipeline(lis_conf)
		} else {
				fmt.Printf("Unimplemented listener\n");
			continue;
		}

		proxy.listeners = append(proxy.listeners, Lis)
	}

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
				if conf.Protocol == "udp" {
					br, err = udpbr.NewUDPBroadcaster(conf)
					if err != nil {
						fmt.Printf("%s\n", err.Error())
					}
					return
				} else if conf.Protocol == "file" {
					br, err = file.NewUDPBroadcaster(conf)
					if err != nil {
						fmt.Printf("%s\n", err.Error())
					}
					return
				}
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
			break
			case webrtclib.ICEConnectionStateFailed:
			case webrtclib.ICEConnectionStateDisconnected:
			proxy.Stop()
			break
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
	for _,lis := range prox.listeners {
		lis.Close()
	}
	prox.Shutdown <- true;
}
