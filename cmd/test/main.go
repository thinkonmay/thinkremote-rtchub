package main

import (
	"fmt"

	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
)

func main() {
	soundcards := tool.GetDevice().Soundcards;
	result := gsttest.GstTestAudio(&soundcards[0])
	fmt.Printf("%s\n",result);
}