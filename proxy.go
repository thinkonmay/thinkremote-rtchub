package proxy

import (
	"fmt"
	"time"

	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster"
	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster/dummy"
	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster/gstreamer"

	"github.com/OnePlay-Internet/webrtc-proxy/listener"
	"github.com/OnePlay-Internet/webrtc-proxy/signalling"
	grpc "github.com/OnePlay-Internet/webrtc-proxy/signalling/gRPC"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
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
	lis []listener.Listener,
	devices *tool.MediaDevice) (proxy *Proxy, err error) {
	proxy = &Proxy{}
	proxy.chan_conf = chan_conf
	proxy.listeners = lis

	proxy.Shutdown = make(chan bool)
	fmt.Printf("added listener\n")


	if grpc_conf != nil {
		var rpc *grpc.GRPCclient
		rpc, err = grpc.InitGRPCClient(grpc_conf,devices,proxy.Shutdown)
		if err != nil {
			return
		}
		proxy.signallingClient = rpc
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
				return sink.CreatePipeline(conf);
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


	start := make(chan bool,2)
	go func() {
		proxy.signallingClient.WaitForConnected()
		time.Sleep(60 * time.Second)
		start<-false
	}()
	go func() {
		proxy.signallingClient.WaitForStart()
		start<-true
	}()
	go func() {
		if <-start {
			proxy.Start()
		} else {
			fmt.Printf("application start timeout, closing\n");
			proxy.Shutdown<-true;
		}
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
	proxy.signallingClient.OnDeviceSelect(func(monitor tool.Monitor,soundcard tool.Soundcard, bitrate int) error {
		for _,listener := range proxy.listeners {
			conf := listener.GetConfig()
			if conf.MediaType == "video" {
				conf.VideoSource = monitor;
				conf.Bitrate = bitrate;
				err := listener.UpdateConfig(conf);
				if err != nil {
					return err
				}
			} else if listener.GetConfig().MediaType == "audio" {
				conf.AudioSource = soundcard;
				err := listener.UpdateConfig(conf);
				if err != nil {
					return err
				}
			}
		}
		return nil;
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
