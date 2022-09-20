package main

import (
	"fmt"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
)

func main() {
	soundcards := tool.GetDevice().Soundcards;
	result := gsttest.GstTestAudio(&config.ListenerConfig{
		AudioSource: soundcards,
	})
	fmt.Printf("%s\n",result);
}