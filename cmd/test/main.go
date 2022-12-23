package main

import (
	"fmt"

	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
)

func main() {
	dev := tool.GetDevice()
	result := gsttest.GstTestAudio(&dev.Soundcards[0])
	fmt.Printf("%s\n",result);
	result = gsttest.GstTestVideo(&dev.Monitors[0])
	fmt.Printf("%s\n",result)
}