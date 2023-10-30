package sunshine

/*
#include <Windows.h>

typedef struct _VideoPipeline  VideoPipeline;
typedef enum _EventType {
    POINTER_VISIBLE,
    CHANGE_BITRATE,
    IDR_FRAME,

    STOP,

	DISPLAY
}EventType;

typedef VideoPipeline* (*STARTQUEUE)				  ( int video_width,
                                                        int video_height,
                                                        int video_bitrate,
                                                        int video_framerate,
                                                        int video_codec,
														char* display_name);

typedef int  		   (*POPFROMQUEUE)			(VideoPipeline* pipeline,
                                                void* data,
                                                int* duration);

typedef void 			(*RAISEEVENT)		 (VideoPipeline* pipeline,
                                              EventType event,
                                              int value);

typedef void  			(*WAITEVENT)			(VideoPipeline* pipeline,
                                                  EventType event,
                                                  int* value);



static HMODULE 			hModule;
static STARTQUEUE 		callstart;
static POPFROMQUEUE 	callpop;
static WAITEVENT		callwait;
static RAISEEVENT       callraise;

int
initlibrary() {
	hModule 	= LoadLibrary(".\\libsunshine.dll");
	callstart 	= (STARTQUEUE)		GetProcAddress( hModule,"StartQueue");
	callpop 	= (POPFROMQUEUE)	GetProcAddress( hModule,"PopFromQueue");
	callraise 	= (RAISEEVENT)		GetProcAddress( hModule,"RaiseEvent");
	callwait	= (WAITEVENT)		GetProcAddress( hModule,"WaitEvent");

	if(callpop ==0 || callstart == 0 || callraise == 0 || callwait == 0)
		return 1;

	return 0;
}

void* StartQueue ( int video_width,
                            int video_height,
                            int video_bitrate,
                            int video_framerate,
                            int video_codec,
							char* display_name){
	return (void*)callstart(video_width,video_height,video_bitrate,video_framerate,video_codec,display_name);
}

int PopFromQueue			(void* pipeline,
                             void* data,
                             int* duration){
	return callpop(pipeline,data,duration);
}

void RaiseEvent	(void* pipeline,
                 EventType event,
                 int value){
	return callraise((VideoPipeline*)pipeline,event,value);
}

void WaitEvent	(void* pipeline,
                 EventType event,
                 int* value){
	return callwait((VideoPipeline*)pipeline,event,value);
}

*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener/adaptive"
	"github.com/thinkonmay/thinkremote-rtchub/listener/display"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/h264"
	"github.com/thinkonmay/thinkremote-rtchub/util/win32"
)

func init(){
	if C.initlibrary() == 1 {
		panic(fmt.Errorf("failed to load libsunshine.dll"))
	}
}

type VideoPipelineC unsafe.Pointer
type VideoPipeline struct {
	closed     bool
	pipeline   unsafe.Pointer
	mut        *sync.Mutex
	properties map[string]int
	sproperties map[string]string

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
		codec:       webrtc.MimeTypeH264,

		clockRate: 90000,

		properties: map[string]int{
			"width": 1920,
			"height": 1080,
			"codec": 1,
			"bitrate": 6000,
			"framerate": 60,
		},
		sproperties: map[string]string{
			"display": display.GetDisplays()[0],
		},
		Multiplexer: multiplexer.NewMultiplexer("video", func() rtppay.Packetizer {
			return h264.NewH264Payloader()
		}),
	}
    pipeline.AdsContext = adaptive.NewAdsContext(
        func(bitrate int) { pipeline.SetProperty("bitrate", bitrate) },
        func() { pipeline.SetProperty("reset", 0) },
    )



	pipeline.reset()
	go func() {
		var duration C.int
		buffer := make([]byte, 256*1024) //256kB
		timestamp := time.Now().UnixNano()

		win32.HighPriorityThread()
		for {
			pipeline.mut.Lock()
			size := C.PopFromQueue(pipeline.pipeline,
                unsafe.Pointer(&buffer[0]), 
                &duration)
			pipeline.mut.Unlock()
			if size == 0 {
				continue
			}

			diff := time.Now().UnixNano() - timestamp
			pipeline.Multiplexer.Send(buffer[:size], uint32(time.Duration(diff).Seconds() * pipeline.clockRate))
			timestamp = timestamp + diff
		}
	}()
	return pipeline, nil
}

func (p *VideoPipeline) GetCodec() string {
	return p.codec
}

func (pipeline *VideoPipeline) reset() {
	pipeline.mut.Lock()
	defer pipeline.mut.Unlock()
	if pipeline.pipeline != nil {
		C.RaiseEvent(pipeline.pipeline,C.STOP,0)
	}

	display.SetResolution(
		pipeline.sproperties["display"],
		pipeline.properties["width"], 
		pipeline.properties["height"], 
	)
	pipeline.pipeline =  C.StartQueue( 
		C.int(pipeline.properties["width"]), 
		C.int(pipeline.properties["height"]), 
		C.int(pipeline.properties["bitrate"]), 
		C.int(pipeline.properties["framerate"]), 
		C.int(pipeline.properties["codec"]), 
		C.CString(pipeline.sproperties["display"]));
}

func (pipeline *VideoPipeline) SetPropertyS(name string, val string) error {
	fmt.Printf("%s change to %d\n", name, val)
	switch name {
	case "display":
		pipeline.sproperties["display"] = val
		pipeline.reset()
	}
	return nil
}
func (pipeline *VideoPipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		pipeline.reset()
	case "framerate":
		pipeline.properties["framerate"] = val
		pipeline.reset()
	case "display":
		pipeline.properties["display"] = val
		pipeline.reset()
	case "codec":
		pipeline.properties["codec"] = val
		pipeline.reset()
	case "width":
		pipeline.properties["width"] = val
	case "height":
		pipeline.properties["height"] = val
	case "pointer":
		pipeline.properties["pointer"] = val
		C.RaiseEvent(pipeline.pipeline,C.POINTER_VISIBLE,0)
	case "reset":
		C.RaiseEvent(pipeline.pipeline,C.IDR_FRAME,0)
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
