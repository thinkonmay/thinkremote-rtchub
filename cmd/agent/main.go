package main

import (
	"fmt"
	"os/exec"
	"time"

	proxy "github.com/pigeatgarlic/webrtc-proxy"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/webrtc/v3"
)

func main() {
	grpc := config.GrpcConfig{
		Port:          8000,
		ServerAddress: "localhost",
	}
	rtc := config.WebRTCConfig{
		Ices: []webrtc.ICEServer{webrtc.ICEServer{
			URLs: []string{"stun:stun.l.google.com:19302"},
		}},
	}
	br := []*config.BroadcasterConfig{
		// &config.BroadcasterConfig{
		// 	Port: 5001,
		// 	Protocol: "udp",
		// 	BufferSize: 10000,

		// 	Type: "video",
		// 	Name: "rtp2",
		// 	Codec: webrtc.MimeTypeH264,
		// },
	}
	lis := []*config.ListenerConfig{
		&config.ListenerConfig{
			Port: 6000,
			Protocol: "udp",
			BufferSize: 10000,

			Type: "video",
			Name: "rtp2",
			Codec: webrtc.MimeTypeH264,
		},
	}	

	var cmds []*exec.Cmd
	for _,i := range lis {
		cmd := exec.Command("gst-launch-1.0.exe ",fmt.Sprintf( "videotestsrc ! openh264enc ! rtph264pay ! application/x-rtp,payload=97 ! udpsink port=%d",i.Port));
		if err := cmd.Run(); err != nil {
			fmt.Printf(err.Error())
		}
		cmds = append(cmds, cmd)
	}

	defer func ()  {
		for _,cmd := range cmds {
			cmd.Process.Kill();
		}
		
	}();

	
	chans := config.DataChannelConfig {
		Offer: false,
		Confs : map[string]*struct{Send chan string; Recv chan string}{
			"test" : &struct{Send chan string; Recv chan string}{
				Send: make(chan string),
				Recv: make(chan string),
			},
		},
	}
		

	go func() {
		for {
			time.Sleep(1 * time.Second);
			chans.Confs["test"].Send <-"test";
		}	
	}()
	go func() {
		for {
			str := <-chans.Confs["test"].Recv
			fmt.Printf("%s\n",str);
		}	
	}()

	_,err := proxy.InitWebRTCProxy(nil,&grpc,&rtc,br,&chans,lis);
	if err != nil {
		panic(err);
	}
	shut := make(chan bool)
	<- shut
}
