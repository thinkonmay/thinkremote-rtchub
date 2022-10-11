package main

import (
	"fmt"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
)

func main() {
	dev,err := tool.GetDevice()
	if err != nil {
		fmt.Printf("%s\n",err.Error());
	}

	soundcards := dev.Soundcards;
	result := gsttest.GstTestAudio(&config.ListenerConfig{
		AudioSource: soundcards[0],
		Bitrate: 128000,
	})
	fmt.Printf("%s\n",result);

	result = gsttest.GstTestNvCodec(&config.ListenerConfig{
		VideoSource: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)

	result = gsttest.GstTestMediaFoundation(&config.ListenerConfig{
		VideoSource: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)
	result = gsttest.GstTestSoftwareEncoder(&config.ListenerConfig{
		VideoSource: dev.Monitors[0],
		Bitrate: 3000,
	})

	fmt.Printf("%s\n",result)
}