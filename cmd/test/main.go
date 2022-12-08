package main

import (
	"fmt"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
)

func main() {
	dev := tool.GetDevice()
	soundcards := dev.Soundcards;
	result,_ := gsttest.GstTestAudio(&config.ListenerConfig{
		Source: soundcards[0],
		Bitrate: 128000,
	})
	fmt.Printf("%s\n",result);

	result,_ = gsttest.GstTestNvCodec(&config.ListenerConfig{
		Source: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)

	result,_ = gsttest.GstTestMediaFoundation(&config.ListenerConfig{
		Source: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)
	result,_ = gsttest.GstTestSoftwareEncoder(&config.ListenerConfig{
		Source: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)
}