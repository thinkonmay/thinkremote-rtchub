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
	result := gsttest.GstTestAudio(&config.ListenerConfig{
		Source: soundcards[0],
		Bitrate: 128000,
	})
	fmt.Printf("%s\n",result);

	result = gsttest.GstTestNvCodec(&config.ListenerConfig{
		Source: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)

	result = gsttest.GstTestMediaFoundation(&config.ListenerConfig{
		Source: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)
	result = gsttest.GstTestSoftwareEncoder(&config.ListenerConfig{
		Source: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)
}