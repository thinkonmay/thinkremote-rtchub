package video

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	proxy "github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/h264"
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

type VideoPipelineC unsafe.Pointer
type VideoPipeline struct {
	closed      bool
	pipeline    unsafe.Pointer
	mut         *sync.Mutex

	clockRate float64

	codec       string
	Multiplexer *multiplexer.Multiplexer
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(memory *proxy.SharedMemory, channel int) (listener.Listener,
	error) {

	pipeline := &VideoPipeline{
		closed:   false,
		pipeline: nil,
		mut:      &sync.Mutex{},
		codec:    webrtc.MimeTypeH264,

		clockRate:   90000,
		Multiplexer: multiplexer.NewMultiplexer("video", h264.NewH264Payloader()),
	}

	go func(queue *proxy.Queue) {
		thread.HighPriorityThread()
		buffer := make([]byte, 256*1024) //256kB
		timestamp := time.Now().UnixNano()
		local_index := queue.CurrentIndex()

		for {
			for local_index == queue.CurrentIndex() {
				time.Sleep(time.Millisecond)
			}

			local_index++
			queue.Copy(buffer,local_index)
			diff := time.Now().UnixNano() - timestamp
			pipeline.Multiplexer.Send(buffer, uint32(time.Duration(diff).Seconds()*pipeline.clockRate))
			timestamp = timestamp + diff
		}
	}(memory.GetQueue(channel))
	return pipeline, nil
}

func (p *VideoPipeline) GetCodec() string {
	return p.codec
}

func (p *VideoPipeline) Close() {
	fmt.Println("stopping video pipeline")
}

func (p *VideoPipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *VideoPipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
