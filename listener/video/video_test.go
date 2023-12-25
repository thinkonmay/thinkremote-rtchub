package video

import (
	"fmt"
	"testing"
	"time"

	"github.com/pion/rtp"
)

func TestVideo(t *testing.T) {
	videopipeline,err := CreatePipeline()
	if err != nil {
		fmt.Printf("error initiate video pipeline %s",err.Error())
		return
	}

	videopipeline.RegisterRTPHandler("test",func(pkt *rtp.Packet) {
		fmt.Printf("packet from %s %s \n","test",pkt.String())
	})
	videopipeline.RegisterRTPHandler("test2",func(pkt *rtp.Packet) {
		fmt.Printf("packet from %s %s \n","test",pkt.String())
	})
	videopipeline.RegisterRTPHandler("test3",func(pkt *rtp.Packet) {
		fmt.Printf("packet from %s %s \n","test",pkt.String())
	})

	videopipeline.Open()
	videopipeline.DeregisterRTPHandler("test")
	time.Sleep(10 * time.Second)
	videopipeline.Close()
	time.Sleep(1 * time.Second)
}