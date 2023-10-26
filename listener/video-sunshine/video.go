package sunshine

/*
#include <Windows.h>

typedef void (*INIT) ();
typedef int  (*STARTQUEUE) ();
typedef int  (*POPFROMQUEUE) (void* data,int* duration);
typedef void (*DEINIT) ();
typedef void (*WAIT) ();


static HMODULE 			hModule;
static INIT 			callinit;
static STARTQUEUE 		callstart;
static POPFROMQUEUE 	callpop;

int
initlibrary(void) {
	hModule = LoadLibrary("libsunshinelib.dll");
	callinit = (INIT)GetProcAddress( hModule,"Init");
	callstart = (STARTQUEUE)GetProcAddress( hModule,"StartQueue");
	callpop = (POPFROMQUEUE)GetProcAddress( hModule,"PopFromQueue");
	if(callpop ==0 || callstart == 0 || callinit == 0)
		return 1;

	callinit();
	return 0;
}


void
start(void) {
	callstart();
}

int
pop(void* data,int* duration) {
	return callpop((void*)data,duration);
}

*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/h264"
	"github.com/thinkonmay/thinkremote-rtchub/listener/adaptive"
)



// VideoPipeline is a wrapper for a GStreamer VideoPipeline

type VideoPipelineC unsafe.Pointer
type VideoPipeline struct {
	closed     bool
	pipeline   unsafe.Pointer
	properties map[string]int

	pipelineStr string
	clockRate   float64

	codec string
	Multiplexer *multiplexer.Multiplexer
	AdsContext datachannel.DatachannelConsumer
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string) ( *VideoPipeline,
	                                       error) {

	pipeline := &VideoPipeline{
		closed:      false,
		pipeline:    nil,
		pipelineStr: pipelineStr,
		codec:       webrtc.MimeTypeH264,

		clockRate: 90000,

		properties: make(map[string]int),
		Multiplexer: multiplexer.NewMultiplexer("video", func() rtppay.Packetizer {
			return h264.NewH264Payloader()
		}),
	}
    pipeline.AdsContext = adaptive.NewAdsContext(
        func(bitrate int) { pipeline.SetProperty("bitrate", bitrate) },
        func() { pipeline.SetProperty("reset", 0) },
    )


	if C.initlibrary() == 1 {
		panic(fmt.Errorf("unable to load library"))
	}


	go C.start()
	go func() {
		time.Sleep(10 * time.Second)
		var duration C.int
		var samples uint32
		buffer := make([]byte, 100*1000*1000) //100MB
		for {
			size := C.pop(
                unsafe.Pointer(&buffer[0]), 
                &duration)
			if size == 0 {
				continue
			}

			samples = uint32(time.Duration(duration).Seconds() * pipeline.clockRate)
			pipeline.Multiplexer.Send(buffer[:size], samples)
		}
	}()
	return pipeline, nil
}

func (p *VideoPipeline) GetCodec() string {
	return p.codec
}

func (pipeline *VideoPipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		// C.video_pipeline_set_bitrate(pipeline.pipeline, C.int(val))
	case "framerate":
		pipeline.properties["framerate"] = val
		// C.video_pipeline_set_framerate(pipeline.pipeline, C.int(val))
	case "pointer":
		pipeline.properties["pointer"] = val
		// C.video_pipeline_enable_pointer(pipeline.pipeline, C.int(val))
	case "reset":
		// C.video_pipeline_generate_idr(pipeline.pipeline)
	default:
		return fmt.Errorf("unknown prop")
	}
	return nil
}

func (p *VideoPipeline) Open() {
	fmt.Println("starting video pipeline")
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
