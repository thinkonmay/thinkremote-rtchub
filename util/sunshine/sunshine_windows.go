package sunshine

/*
#include <Windows.h>

typedef struct _VideoPipeline  VideoPipeline;
typedef enum _EventType {
    POINTER_VISIBLE,
    CHANGE_BITRATE,
    CHANGE_FRAMERATE,
    CHANGE_DISPLAY,
    IDR_FRAME,

    STOP
} EventType;

typedef enum _Codec {
    H264 = 1,
    H265,
    AV1,
    OPUS,
}Codec;

typedef VideoPipeline* (*STARTQUEUE)				  ( int codec);

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

void* StartQueue ( int video_codec){
	return (void*)callstart(video_codec);
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
	"unsafe"
)

func init(){
	if C.initlibrary() == 1 {
		panic(fmt.Errorf("failed to load libsunshine.dll"))
	}
}

func StartQueue (codec int) unsafe.Pointer {
    return C.StartQueue(C.int(codec))
} 

func WaitEvent (pipline unsafe.Pointer, code int) {
    C.WaitEvent(pipline,C.EventType(code))
} 
func RaiseEvent(pipline unsafe.Pointer,code int,val int) {
    C.RaiseEvent(pipline,C.EventType(code),C.int(val))
} 
func RaiseEventS(pipline unsafe.Pointer,code int,val string) {
    C.RaiseEventS(pipline,C.EventType(code),C.CString(val))
} 

var duration C.int
func PopFromQueue(pipline unsafe.Pointer,buff unsafe.Pointer) int {
    return int(C.PopFromQueue(pipline,buff,&duration))
} 

var (

	IDR_FRAME = C.IDR_FRAME
    POINTER_VISIBLE = C.POINTER_VISIBLE
    CHANGE_BITRATE = C.CHANGE_BITRATE
    CHANGE_FRAMERATE = C.CHANGE_FRAMERATE
    CHANGE_DISPLAY = C.CHANGE_DISPLAY
    STOP = C.STOP

    H264 = C.H264
    OPUS = C.OPUS
)