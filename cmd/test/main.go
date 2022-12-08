package main

import (
	"fmt"

	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
)

func main() {
	dev := tool.GetDevice()
	soundcards := dev.Soundcards;
	result,_ := gsttest.GstTestAudio(&soundcards[0])
	fmt.Printf("%s\n",result);
	result,_ = gsttest.GstTestNvCodec( &dev.Monitors[0] )
	fmt.Printf("%s\n",result)
	result,_ = gsttest.GstTestMediaFoundation( &dev.Monitors[0] )
	fmt.Printf("%s\n",result)
	result,_ = gsttest.GstTestSoftwareEncoder(&dev.Monitors[0] )
	fmt.Printf("%s\n",result)
}