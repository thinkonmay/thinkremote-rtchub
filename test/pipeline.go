package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pion/rtp"
	// "github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	// video "github.com/thinkonmay/thinkremote-rtchub/listener/video"
	video "github.com/thinkonmay/thinkremote-rtchub/listener/video-sunshine"
)

const (
	source = "tmd3d11screencapturesrc blocksize=8192 do-timestamp=true ! video/x-raw(memory:D3D11Memory),clock-rate=90000,framerate=55/1 ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! tmd3d11convert ! video/x-raw(memory:D3D11Memory),width=192,height=108 ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! "
	nvidia_default = source + "tmnvd3d11h264enc name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
	amd_default    = source + "tmamfh264enc 	name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
	intel_default  = source + "tmqsvh264enc 	name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"

	audio_default = "wasapisrc name=source device=\"\\{0.0.1.00000000\\}.\\{4d54e66c-3242-4385-bff8-9c82dca3682a\\}\" ! audioresample ! audio/x-raw,rate=48000 ! audioconvert ! opusenc name=encoder ! appsink name=appsink"
)

func main() {

	pipeline := ""
	switch os.Args[1] {
	case "amd":
		pipeline = amd_default
		break
	case "nvidia":
		pipeline = nvidia_default
		break
	case "intel":
		pipeline = intel_default
		break
	}

	videopipeline, err := video.CreatePipeline(pipeline)

	if err != nil {
		fmt.Printf("error initiate video pipeline %s\n", err.Error())
		return
	}
	videopipeline.Open()
	count := 1

	udpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:65500")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	go func ()  {
		for {
			videopipeline.SetProperty("reset",0)
			time.Sleep(time.Second)
		}
	}()

	pkts := make(chan *rtp.Packet)
	videopipeline.RegisterRTPHandler("abc",func(pkt *rtp.Packet) { pkts<-pkt })
	go func() { for { pkt := <- pkts
			bytes,err := pkt.Marshal()
			if err != nil {
				panic(err)
			}

			_,err = conn.Write(bytes)
			if err != nil {
				panic(err)
			}

			count++
		}
		
	}()

	<-make(chan bool, 0)
}