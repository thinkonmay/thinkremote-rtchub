package sunshine

/*
#include <Windows.h>

typedef struct _VideoPipeline  VideoPipeline;
typedef enum _EventType {
    POINTER_VISIBLE,
    CHANGE_BITRATE,
    CHANGE_DISPLAY,
    IDR_FRAME,

    STOP
}EventType;

typedef enum _Codec {
    H264 = 1,
    H265,
    AV1,
}Codec;

typedef VideoPipeline* (*STARTQUEUE)				  ( int video_codec,
														char* display_name);

typedef int  		   (*POPFROMQUEUE)			(VideoPipeline* pipeline,
                                                void* data,
                                                int* duration);

typedef void 			(*RAISEEVENT)		 (VideoPipeline* pipeline,
                                              EventType event,
                                              int value);

typedef void 			(*RAISEEVENTS)		 (VideoPipeline* pipeline,
                                              EventType event,
                                              char* value);

typedef void  			(*WAITEVENT)			(VideoPipeline* pipeline,
                                                  EventType event);



static HMODULE 			hModule;
static STARTQUEUE 		callstart;
static POPFROMQUEUE 	callpop;
static WAITEVENT		callwait;
static RAISEEVENT       callraise;
static RAISEEVENTS      callraises;

int
initlibrary() {
	hModule 	= LoadLibrary(".\\libsunshine.dll");
	callstart 	= (STARTQUEUE)		GetProcAddress( hModule,"StartQueue");
	callpop 	= (POPFROMQUEUE)	GetProcAddress( hModule,"PopFromQueue");
	callraise 	= (RAISEEVENT)		GetProcAddress( hModule,"RaiseEvent");
	callraises	= (RAISEEVENTS)		GetProcAddress( hModule,"RaiseEventS");
	callwait	= (WAITEVENT)		GetProcAddress( hModule,"WaitEvent");

	if(callpop ==0 || callstart == 0 || callraise == 0 || callwait == 0)
		return 1;

	return 0;
}

void* StartQueue ( int video_codec,
				   char* display_name){
	return (void*)callstart(video_codec,display_name);
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

void RaiseEventS(void* pipeline,
                 EventType event,
                 char* value){
	return callraises((VideoPipeline*)pipeline,event,value);
}

void WaitEvent	(void* pipeline,
                 EventType event){
	return callwait((VideoPipeline*)pipeline,event);
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
	"github.com/thinkonmay/thinkremote-rtchub/listener"
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
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline() ( listener.Listener,
	                                       error) {

	pipeline := &VideoPipeline{
		closed:      false,
		pipeline:    nil,
		mut: 		 &sync.Mutex{},
		codec:       webrtc.MimeTypeH264,

		clockRate: 90000,

		properties: map[string]int{
			"codec": 1,
			"bitrate": 6000,
		},
		sproperties: map[string]string{
			"display": display.GetDisplays()[0],
		},
		Multiplexer: multiplexer.NewMultiplexer("video", func() rtppay.Packetizer {
			return h264.NewH264Payloader()
		}),
	}



	pipeline.reset()
	go func() { win32.HighPriorityThread()
		var duration C.int
		buffer := make([]byte, 256*1024) //256kB
		timestamp := time.Now().UnixNano()

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
		C.WaitEvent(pipeline.pipeline,C.STOP)
	}

	pipeline.pipeline =  C.StartQueue( 
		C.int(pipeline.properties["codec"]), 
		C.CString(pipeline.sproperties["display"]));
}

func (pipeline *VideoPipeline) SetPropertyS(name string, val string) error {
	fmt.Printf("%s change to %s\n", name, val)
	switch name {
	case "display":
		pipeline.sproperties["display"] = val
		C.RaiseEventS(pipeline.pipeline,C.CHANGE_DISPLAY,C.CString(pipeline.sproperties["display"]))
	}
	return nil
}
func (pipeline *VideoPipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		C.RaiseEvent(pipeline.pipeline,C.CHANGE_BITRATE,C.int(pipeline.properties["bitrate"]))
	case "codec":
		pipeline.properties["codec"] = val
		pipeline.reset()
	case "pointer":
		pipeline.properties["pointer"] = val
		C.RaiseEvent(pipeline.pipeline,C.POINTER_VISIBLE,C.int(pipeline.properties["pointer"]))
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
