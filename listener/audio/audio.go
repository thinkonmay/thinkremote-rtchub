package audio

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/opus"
)

/*
#include <gst/app/gstappsink.h>
#include <stdint.h>
#include <stdlib.h>

void
start_audio_mainloop(void) {
    GMainLoop *gstreamer_audio_main_loop = NULL;
    gstreamer_audio_main_loop = g_main_loop_new(NULL, FALSE);

    g_main_loop_run(gstreamer_audio_main_loop);
}


typedef struct _VideoPipeline {
    GstElement* appsink;
    GstElement* pipeline;
    GstElement* encoder;

    int frame_count;
} VideoPipeline;





void audio_pipeline_stop(void* pipeline);
void audio_pipeline_start(void* pipeline);




static void
handle_pipeline_failure(VideoPipeline* audio_pipeline)
{
    audio_pipeline_stop(audio_pipeline);
    audio_pipeline_start(audio_pipeline);
}


static gboolean
gstreamer_send_audio_bus_call(GstBus *bus,
                              GstMessage *msg,
                              gpointer data)
{
    GError *error;
    gchar debug[1000] = {0};
    VideoPipeline* pipeline = (VideoPipeline*)data;


    switch (GST_MESSAGE_TYPE(msg)) {
    case GST_MESSAGE_EOS:
        g_print("End of stream\n");
        handle_pipeline_failure(pipeline);
        break;

    case GST_MESSAGE_ERROR: {
        gst_message_parse_error(msg, &error, (gchar**)&debug);
        g_printerr("Video pipeline error: %s\n", error->message);
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
create_audio_pipeline(char *pipeline_desc,
                      char** err) {
    GError *error = NULL;
    GstBus *bus = NULL;
    if (!pipeline_desc) {
        *err = (void*)"empty pipeline";
        return NULL;
    }



    gst_init(NULL, NULL);
    VideoPipeline* pipeline = (VideoPipeline*) malloc(sizeof(VideoPipeline));
    pipeline->frame_count        = 0;
    pipeline->pipeline           = gst_parse_launch(pipeline_desc, &error);
    if (error) {
        *err = error->message;
        return NULL;
    }

    bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline->pipeline));
    gst_bus_add_watch(bus, gstreamer_send_audio_bus_call, pipeline);
    gst_object_unref(bus);

    pipeline->appsink            = gst_bin_get_by_name(GST_BIN(pipeline->pipeline), "appsink");
    pipeline->encoder            = gst_bin_get_by_name(GST_BIN(pipeline->pipeline), "encoder");
    *err = NULL;
    return pipeline;
}

void
audio_pipeline_start(void* pipelineIn) {
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_PIPELINE(pipeline->pipeline))
        return;

    gst_element_set_state(pipeline->pipeline, GST_STATE_PLAYING);
}

void
audio_pipeline_stop(void* pipelineIn) {
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_PIPELINE(pipeline->pipeline))
        return;
    else if (!GST_IS_ELEMENT(pipeline->appsink))
        return;

    pipeline->frame_count = 0;
    gst_element_set_state(pipeline->pipeline, GST_STATE_NULL);
}




void
audio_pipeline_set_bitrate(void* pipelineIn,
                            int bitrate) {
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_ELEMENT(pipeline->encoder))
        return;

    g_object_set(pipeline->encoder, "bitrate", bitrate, NULL);
}


int
audio_pipeline_pop_buffer	(void* pipelineIn,
                void *bufferOut,
                int *samplesOut)
{
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!GST_IS_ELEMENT(pipeline->appsink))
        return 0;

    GstBuffer* buffer;
    GstSample* sample = NULL;
    gint copy_size;

    g_signal_emit_by_name (pipeline->appsink, "pull-sample", &sample,NULL);
    if (!sample)
        return 0;
    else if (!GST_IS_SAMPLE(sample))
        return 0;

    buffer = gst_sample_get_buffer(sample);
    if (!buffer)
        return 0;

    copy_size = gst_buffer_get_size(buffer);
    if(!copy_size)
        return 0;

    gst_buffer_extract(buffer, 0, bufferOut, copy_size); // linking gstreamer to go limited available stack frame // very dangerous to modify
    gst_sample_unref (sample);

    *samplesOut = GST_BUFFER_DURATION(buffer);

    pipeline->frame_count++;
    GST_INFO("got buffer %d with size %d",pipeline->frame_count,copy_size);
	return copy_size;
}
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
*/
import "C"

func init() {
	go C.start_audio_mainloop()
}

type AudioPipelineC unsafe.Pointer
type AudioPipeline struct {
	closed      bool
	pipelineStr string
	clockRate float64
	codec string

	pipeline    unsafe.Pointer
	Multiplexer *multiplexer.Multiplexer
}


// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelinestr string) (*AudioPipeline, error) {
	pipeline := &AudioPipeline{
		pipeline:    unsafe.Pointer(nil),
		pipelineStr: pipelinestr,
		clockRate:   48000,
		codec:       webrtc.MimeTypeOpus,

		Multiplexer: multiplexer.NewMultiplexer("audio", func() rtppay.Packetizer {
			return opus.NewOpusPayloader()
		}),
	}

    if pipelinestr == "" {
	    return pipeline, nil
    }

	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	err := C.CString("")
	fmt.Printf("starting audio pipeline %s\n", pipelinestr)
	pipeline.pipeline = C.create_audio_pipeline(pipelineStrUnsafe, &err)
	err_str := C.GoString(err)
	if len(err_str) != 0 {
		return nil, fmt.Errorf("fail to create pipeline %s", err_str)
	}


	go func() {
		var duration C.int
		var size C.int
		var samples uint32
		buffer := make([]byte, 256*1024) //256kB

		for {
			size = C.audio_pipeline_pop_buffer(
                pipeline.pipeline,
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


func (p *AudioPipeline) GetCodec() string {
	return p.codec
}

func (p *AudioPipeline) Open() {
	fmt.Println("starting audio pipeline")
	C.audio_pipeline_start(p.pipeline)
}

func (p *AudioPipeline) Close() {
	fmt.Println("stoping audio pipeline")
	C.audio_pipeline_stop(p.pipeline)
}

func (p *AudioPipeline) SetPropertyS(name string, val string) error {
	return fmt.Errorf("unknown prop")
}
func (p *AudioPipeline) SetProperty(name string, val int) error {
	switch name {
	case "audio-reset":
	    C.audio_pipeline_stop(p.pipeline)
	    C.audio_pipeline_start(p.pipeline)
		return nil
	default:
	}
	return fmt.Errorf("unknown prop")
}


func (p *AudioPipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *AudioPipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
