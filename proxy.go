package proxy

import (
	"fmt"

	"github.com/pigeatgarlic/webrtc-proxy/broadcaster"
	udpbr "github.com/pigeatgarlic/webrtc-proxy/broadcaster/udp"
	datachannel "github.com/pigeatgarlic/webrtc-proxy/data-channel"
	"github.com/pigeatgarlic/webrtc-proxy/listener"
	"github.com/pigeatgarlic/webrtc-proxy/listener/udp"

	"github.com/pigeatgarlic/webrtc-proxy/signalling"
	grpc "github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pigeatgarlic/webrtc-proxy/webrtc"
	webrtclib "github.com/pion/webrtc/v3"
)

type Proxy struct {
	listeners []listener.Listener
	datachannels []datachannel.Datachannel
	signallingClient signalling.Signalling
	webrtcClient *webrtc.WebRTCClient
	started bool
}



func InitWebRTCProxy(sock *config.WebsocketConfig,
					 grpc_conf *config.GrpcConfig,
					 webrtc_conf *config.WebRTCConfig,
					 br_conf []*config.BroadcasterConfig,
					 lis  []*config.ListenerConfig) (proxy *Proxy, err error) {
	proxy = &Proxy{}	
	proxy.started = false;
	for _,lis_conf := range lis {
		if lis_conf.Protocol == "udp" {
			var udpLis udp.UDPListener;
			udpLis,err = udp.NewUDPListener(lis_conf);
			if err != nil {
				return;	
			}
			fmt.Printf("added listener\n");
			proxy.listeners = append(proxy.listeners, &udpLis);
		}else if lis_conf.Protocol == "tpc" {
			err = fmt.Errorf("Unimplemented");
			return;
		}
	}

	if grpc_conf != nil {
		var rpc grpc.GRPCclient;
		rpc, err = grpc.InitGRPCClient(grpc_conf);
		if err != nil {
			return;	
		}
		proxy.signallingClient = &rpc;
	} else if sock != nil {
		err = fmt.Errorf("Unimplemented");
		return;
	} else {
		err = fmt.Errorf("Unimplemented");
		return;
	}

	proxy.webrtcClient,err = webrtc.InitWebRtcClient(func(tr *webrtclib.TrackRemote) (br broadcaster.Broadcaster, err error) {
		for _,conf := range br_conf {
			if tr.Codec().MimeType == conf.Codec {
				if conf.Protocol == "udp" {
					br,err = udpbr.NewUDPBroadcaster(conf);
					return;
				}	
			}
		}

		err = fmt.Errorf("unimplemented broadcaster");
		return;
	},*webrtc_conf);
	if err != nil {
		panic(err);
	}

	go func ()  {
		proxy.signallingClient.WaitForStart()
		if proxy.started == false {
			proxy.Start()
		}
	}()
	go func ()  {
		proxy.webrtcClient.WaitConnected()
		if proxy.started == false {
			proxy.Start()
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
		proxy.webrtcClient.OnIncomingICE(i);
	})
	proxy.signallingClient.OnSDP(func(i *webrtclib.SessionDescription) {
		proxy.webrtcClient.OnIncominSDP(i);
	})
	return;
}

func (prox *Proxy) Start() {
	prox.started = true;
	prox.webrtcClient.ListenRTP(prox.listeners);	
}