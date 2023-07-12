package microphone

/*
#include <gst/app/gstappsrc.h>
#include <stdint.h>
#include <stdlib.h>
#include <glib.h>

void
start_microphone_mainloop(void) {
    GMainLoop *gstreamer_microphone_main_loop = NULL;
    gstreamer_microphone_main_loop = g_main_loop_new(NULL, FALSE);

    g_main_loop_run(gstreamer_microphone_main_loop);
}


typedef struct _MicPipeline {
    GstElement* appsrc;
    GstElement* pipeline;

} MicPipeline;


void mic_pipeline_stop(void* pipeline);
void mic_pipeline_start(void* pipeline);




static void
handle_pipeline_failure(MicPipeline* mic_pipeline)
{
    mic_pipeline_stop(mic_pipeline);
    mic_pipeline_start(mic_pipeline);
}


static gboolean
gstreamer_send_mic_bus_call(GstBus *bus,
                              GstMessage *msg,
                              gpointer data)
{
    GError *error;
    gchar debug[1000] = {0};
    MicPipeline* pipeline = (MicPipeline*)data;


    switch (GST_MESSAGE_TYPE(msg)) {
    case GST_MESSAGE_EOS:
        g_print("End of stream\n");
        handle_pipeline_failure(pipeline);
        break;

    case GST_MESSAGE_ERROR: {
        gst_message_parse_error(msg, &error, (gchar**)&debug);
        g_printerr("Mic pipeline error: %s\n", error->message);
        g_printerr("Debug message: %s\n", debug);

        handle_pipeline_failure(pipeline);
        g_error_free(error);
    }
    default:
        break;
    }

    return TRUE;
}



void*
create_mic_pipeline(char *pipeline_desc,
                    char** err) {
    GError *error = NULL;
    GstBus *bus = NULL;
    if (!pipeline_desc) {
        *err = (void*)"empty pipeline";
        return NULL;
    }



    gst_init(NULL, NULL);
    MicPipeline* pipeline = (MicPipeline*) malloc(sizeof(MicPipeline));
    pipeline->pipeline           = gst_parse_launch(pipeline_desc, &error);
    if (error) {
        *err = error->message;
        return NULL;
    }

    bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline->pipeline));
    gst_bus_add_watch(bus, gstreamer_send_mic_bus_call, pipeline);
    gst_object_unref(bus);

    pipeline->appsrc            = gst_bin_get_by_name(GST_BIN(pipeline->pipeline), "appsrc");

    *err = NULL;
    return pipeline;
}

void
mic_pipeline_start(void* pipelineIn) {
    MicPipeline* pipeline = (MicPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_PIPELINE(pipeline->pipeline))
        return;

    gst_element_set_state(pipeline->pipeline, GST_STATE_PLAYING);
}

void
mic_pipeline_stop(void* pipelineIn) {
    MicPipeline* pipeline = (MicPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_PIPELINE(pipeline->pipeline))
        return;


    gst_element_set_state(pipeline->pipeline, GST_STATE_NULL);
}

void
mic_pipeline_push_buffer(void* pipelineIn,
						 void* buffer,
						 int len)
{
    MicPipeline* pipeline = (MicPipeline*)pipelineIn;
    if (!GST_IS_ELEMENT(pipeline->appsrc))
        return;


	gpointer p = g_memdup2(buffer, len);
	GstBuffer *gbuffer = gst_buffer_new_wrapped(p, len);
	gst_app_src_push_buffer(GST_APP_SRC(pipeline->appsrc), gbuffer);
}



#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0 gstreamer-audio-1.0
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
)


func init() {
	go C.start_microphone_mainloop()
}


// MicPipeline is a wrapper for a GStreamer MicPipeline

type MicPipelineC unsafe.Pointer
type MicPipeline struct {
	closed     bool
	pipeline   unsafe.Pointer

	pipelineStr string
	clockRate   float64

	codec string
	Multiplexer *multiplexer.Multiplexer
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(payloadType webrtc.PayloadType) ( *MicPipeline,
	                                       error) {
    pipelineStr  := fmt.Sprintf(
        "appsrc format=time is-live=true do-timestamp=true name=appsrc ! application/x-rtp,payload=%d,encoding-name=OPUS,clock-rate=48000 ! rtpopusdepay ! opusdec ! audioconvert ! audioresample ! wasapisink device=\"\\{0.0.0.00000000\\}.\\{e9e2e411-614e-4ba0-8584-aca0f67853cd\\}\"", payloadType)
	pipeline := &MicPipeline{
		closed:      false,
		pipeline:    nil,
		pipelineStr: pipelineStr,
		codec:       webrtc.MimeTypeH264,

		clockRate: 90000,
	}


	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	err := C.CString("")
	fmt.Printf("starting mic pipeline %s\n", pipelineStr)
	pipeline.pipeline = C.create_mic_pipeline(pipelineStrUnsafe, &err)
	err_str := C.GoString(err)

	if len(err_str) != 0 {
		return nil, fmt.Errorf("failed to create pipeline %s", err_str)
	}

	return pipeline, nil
}



func (p *MicPipeline) Open() {
	fmt.Println("starting mic pipeline")
	C.mic_pipeline_start(p.pipeline)
}
func (p *MicPipeline) Close() {
	fmt.Println("stopping mic pipeline")
	C.mic_pipeline_stop(p.pipeline)
}

func (p *MicPipeline) Push(buff []byte) {
    b := C.CBytes(buff)
	defer C.free(b)
    C.mic_pipeline_push_buffer(p.pipeline,b,C.int(len(buff)))
}


