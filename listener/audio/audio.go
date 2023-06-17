// Package gst provides an easy API to create an appsink pipeline
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

#include <gst/app/gstappsrc.h>

void PushBufferAudio (void  *buffer, 
                      int bufferLen, 
                      int samples) {

}

int  PopBufferAudio	(void **buffer, 
                     int *samples) {
	return 0;
}

void handleAudioStopOrError() {

}






GMainLoop *gstreamer_audio_main_loop = NULL;
void
start_audio_mainloop(void) {
    gstreamer_audio_main_loop = g_main_loop_new(NULL, FALSE);
    g_main_loop_run(gstreamer_audio_main_loop);
}




static gboolean
handle_gstreamer_bus_call(GstBus *bus, GstMessage *msg, gpointer data) {
    switch (GST_MESSAGE_TYPE(msg)) {

    case GST_MESSAGE_EOS:
        g_print("End of stream\n");
        handleAudioStopOrError();
        break;

    case GST_MESSAGE_ERROR: {
        gchar *debug;
        GError *error;
        gst_message_parse_error(msg, &error, &debug);
        g_free(debug);
        g_printerr("Audio pipeline error: %s\n", error->message);
        g_error_free(error);
        handleAudioStopOrError();
    }
    default:
        break;
    }

    return TRUE;
}

GstFlowReturn
handle_audio_buffer(GstElement *object, gpointer user_data) {
    GstSample *sample = NULL;
    GstBuffer *buffer = NULL;
    gsize copy_size = 0;

    g_signal_emit_by_name (object, "pull-sample", &sample);
    if (sample) {
        buffer = gst_sample_get_buffer(sample);
        if (buffer) {
            copy_size = gst_buffer_get_size(buffer);
            if(copy_size) {
                gpointer copy = malloc(copy_size);
                gst_buffer_extract(buffer, 0, (gpointer)copy, copy_size);
                PushBufferAudio(copy, copy_size, GST_BUFFER_DURATION(buffer));
                free(copy);
            }
        }
        gst_sample_unref (sample);
    }

    return GST_FLOW_OK;
}

void*
create_audio_pipeline(char *pipeline,
                      char** err)
{
    gst_init(NULL, NULL);

    *err = NULL;
    GError *error = NULL;
    GstElement * ret = gst_parse_launch(pipeline, &error);
    if (error) {
        *err = error->message;
        return NULL;
    }
    return ret;
}

void
start_audio_pipeline(void* pipeline)
{
    GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
    gst_bus_add_watch(bus, handle_gstreamer_bus_call, pipeline);
    gst_object_unref(bus);


    GstElement *appsink = gst_bin_get_by_name(GST_BIN(pipeline), "appsink");
    g_object_set(appsink, "emit-signals", TRUE, NULL);
    g_signal_connect(appsink, "new-sample", G_CALLBACK(handle_audio_buffer), NULL);
    gst_object_unref(appsink);

    gst_element_set_state((GstElement*)pipeline, GST_STATE_PLAYING);
}

void
audio_pipeline_set_bitrate(void* pipeline, int bitrate) {
    GstElement *encoder = gst_bin_get_by_name(GST_BIN(pipeline), "encoder");
    g_object_set(encoder, "bitrate", bitrate, NULL);
}

void
stop_audio_pipeline(void* pipeline) {
    gst_element_set_state((GstElement*)pipeline, GST_STATE_NULL);
}

int
string_get_length(void* string)
{
    return strlen((char*)string);
}
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
*/
import "C"

func init() {
	go C.start_audio_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	closed      bool
	pipeline    unsafe.Pointer
	pipelineStr string

	clockRate float64

	codec string

	Multiplexer *multiplexer.Multiplexer
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelinestr string) (*Pipeline, error) {
	pipeline = &Pipeline{
		pipeline:    unsafe.Pointer(nil),
		pipelineStr: pipelinestr,
		clockRate:   48000,
		codec:       webrtc.MimeTypeOpus,

		Multiplexer: multiplexer.NewMultiplexer("audio", func() rtppay.Packetizer {
			return opus.NewOpusPayloader()
		}),
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


    go func ()  {
        var buffer unsafe.Pointer
        var duration C.int
        for {
            bufferLen  := C.PopBufferAudio(&buffer, &duration) 
            samples := uint32(time.Duration(duration).Seconds() * pipeline.clockRate)
            pipeline.Multiplexer.Send(buffer, uint32(bufferLen), uint32(samples))
        }
    }()

	return pipeline, nil
}


func handleAudioStopOrError() {
	pipeline.Close()

	err := C.CString("")
	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	pipeline.pipeline = C.create_audio_pipeline(pipelineStrUnsafe, &err)
	err_str := C.GoString(err)
	if len(err_str) != 0 {
		fmt.Printf("fail to create pipeline %s", err_str)
	}

	pipeline.Open()
}

func (p *Pipeline) GetCodec() string {
	return p.codec
}

func (p *Pipeline) Open() {
	fmt.Println("starting audio pipeline")
	C.start_audio_pipeline(pipeline.pipeline)
}

func (p *Pipeline) Close() {
	fmt.Println("stoping audio pipeline")
	C.stop_audio_pipeline(p.pipeline)
}

func (p *Pipeline) SetProperty(name string, val int) error {
	switch name {
	case "audio-reset":
		handleAudioStopOrError()
		return nil
	default:
	}
	return fmt.Errorf("unknown prop")
}


func (p *Pipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *Pipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
