// Package gst provides an easy API to create an appsink pipeline
package video

/*

#include <gst/app/gstappsrc.h>
#include <gst/video/video-event.h>
#include <stdint.h>
#include <stdlib.h>
#include <glib.h>
#include <pthread.h>

void PushBufferVideo	   (void  *buffer, int bufferLen , int samples) {

}

int  PopBufferVideo	   				   (void **buffer, int *samples) {
	return 0;
}




void handleVideoStopOrError            () {

}

void*create_video_pipeline(char *pipeline,char** err);
void video_pipeline_set_bitrate(void* pipeline, int bitrate);
void video_pipeline_set_framerate(void* pipeline, int framerate);
void start_video_pipeline(void*pipeline);
void stop_video_pipeline(void*pipeline);
void force_gen_idr_frame_video_pipeline(void*pipeline);
void start_video_mainloop(void);


GMainLoop *gstreamer_video_main_loop = NULL;
void
start_video_mainloop(void) {
    gstreamer_video_main_loop = g_main_loop_new(NULL, FALSE);
    g_main_loop_run(gstreamer_video_main_loop);
}

static gboolean
gstreamer_send_video_bus_call(GstBus *bus, GstMessage *msg, gpointer data) {
    switch (GST_MESSAGE_TYPE(msg)) {

    case GST_MESSAGE_EOS:
        g_print("End of stream\n");
        handleVideoStopOrError();
        break;

    case GST_MESSAGE_ERROR: {
        gchar *debug;
        GError *error;

        gst_message_parse_error(msg, &error, &debug);
        g_free(debug);

        g_printerr("Video pipeline error: %s\n", error->message);
        g_error_free(error);
        handleVideoStopOrError();
    }
    default:
        break;
    }

    return TRUE;
}

GstFlowReturn
handle_video_sample(GstElement *object, gpointer user_data) {
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
                gst_buffer_extract(buffer, 0, (gpointer)copy, copy_size); // linking gstreamer to go limited available stack frame // very dangerous to modify
                PushBufferVideo(copy, copy_size, GST_BUFFER_DURATION(buffer));
                free(copy);
            }
        }
        gst_sample_unref (sample);
    }
    return GST_FLOW_OK;
}

void*
create_video_pipeline(char *pipeline,
                      char** err) {
    if (!pipeline) {
        *err = (void*)"empty pipeline";
        return NULL;
    }

    gst_init(NULL, NULL);

    *err = NULL;
    GError *error = NULL;
    GstElement* el = gst_parse_launch(pipeline, &error);
    if (error) {
        *err = error->message;
        return NULL;
    }

    return (void*)el;
}

void
start_video_pipeline(void* pipeline) {
    if (!pipeline)
        return;

    GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
    gst_bus_add_watch(bus, gstreamer_send_video_bus_call, NULL);
    gst_object_unref(bus);

    GstElement *appsink = gst_bin_get_by_name(GST_BIN(pipeline), "appsink");
    g_object_set(appsink, "emit-signals", TRUE, NULL);
    g_signal_connect(appsink, "new-sample", G_CALLBACK(handle_video_sample), NULL);
    gst_object_unref(appsink);

    gst_element_set_state((GstElement*)pipeline, GST_STATE_PLAYING);
}


void
video_pipeline_set_framerate(void* pipeline, int framerate) {
    if (!pipeline)
        return;

    GstElement *framerateFilter = gst_bin_get_by_name(GST_BIN(pipeline), "framerateFilter");
    if (!framerateFilter)
        return;

    char* capsstr = g_strdup_printf ("video/x-raw(memory:D3D11Memory),framerate=%d/1",framerate);
    GstCaps* caps = gst_caps_from_string (capsstr);
    g_free (capsstr);

    g_object_set (framerateFilter, "caps", caps, NULL);
    gst_caps_unref (caps);
}


void
video_pipeline_set_bitrate(void* pipeline, int bitrate) {
    if (!pipeline)
        return;

    GstElement *encoder = gst_bin_get_by_name(GST_BIN(pipeline), "encoder");
    g_object_set(encoder, "bitrate", bitrate, NULL);
    g_object_set(encoder, "gop-size", 0, NULL);
}

void
stop_video_pipeline(void* pipeline) {
    if (!pipeline)
        return;

    gst_element_set_state((GstElement*)pipeline, GST_STATE_NULL);
}
void
force_gen_idr_frame_video_pipeline(void* pipeline) {
    if (!pipeline)
        return;

    GstElement *encoder = gst_bin_get_by_name(GST_BIN(pipeline), "encoder");
    if (!encoder)
        return;

    GstPad* srcpad = gst_element_get_static_pad(encoder,"src");

	   GstEvent* event = gst_video_event_new_upstream_force_key_unit(GST_CLOCK_TIME_NONE,TRUE,0);
    gst_pad_send_event(srcpad,event);
    gst_object_unref(srcpad);
}

#include <stdlib.h>
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0 gstreamer-video-1.0
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
	"github.com/thinkonmay/thinkremote-rtchub/listener/video/adaptive"
)

func init() {
	go C.start_video_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	closed     bool
	pipeline   unsafe.Pointer
	properties map[string]int

	pipelineStr string
	clockRate   float64

	Multiplexer *multiplexer.Multiplexer

	codec string

	AdsContext datachannel.DatachannelConsumer
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string) (
	*Pipeline,
	error) {
	pipeline = &Pipeline{
		closed:      false,
		pipeline:    unsafe.Pointer(nil),
		pipelineStr: pipelineStr,
		codec:       webrtc.MimeTypeH264,

		clockRate: 90000,

		properties: make(map[string]int),
		AdsContext: adaptive.NewAdsContext(
			func(bitrate int) { pipeline.SetProperty("bitrate", bitrate) },
			func() { pipeline.SetProperty("reset", 0) },
		),
		Multiplexer: multiplexer.NewMultiplexer("video", func() rtppay.Packetizer {
			return h264.NewH264Payloader()
		}),
	}

	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	err := C.CString("")
	fmt.Printf("starting video pipeline %s\n", pipelineStr)
	pipeline.pipeline = C.create_video_pipeline(pipelineStrUnsafe, &err)
	err_str := C.GoString(err)

	if len(err_str) != 0 {
		return nil, fmt.Errorf("failed to create pipeline %s", err_str)
	}

	return pipeline, nil
}

func (pipeline *Pipeline) BufferVideo() {
	var samplesInt C.int
	var buffer unsafe.Pointer
	size := C.PopBufferVideo(&buffer, &samplesInt)
	samples := uint32(time.Duration(samplesInt).Seconds() * pipeline.clockRate)
	pipeline.Multiplexer.Send(buffer, uint32(size), uint32(samples))
}

func (p *Pipeline) GetCodec() string {
	return p.codec
}

func (p *Pipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	if p.pipeline == nil || p.properties == nil {
		return fmt.Errorf("attemping to set property while pipeline is not running, aborting")
	}

	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		C.video_pipeline_set_bitrate(pipeline.pipeline, C.int(val))
		C.force_gen_idr_frame_video_pipeline(pipeline.pipeline)
	case "framerate":
		pipeline.properties["framerate"] = val
		C.video_pipeline_set_framerate(pipeline.pipeline, C.int(val))
		C.force_gen_idr_frame_video_pipeline(pipeline.pipeline)
	case "reset":
		C.force_gen_idr_frame_video_pipeline(pipeline.pipeline)
	default:
		return fmt.Errorf("unknown prop")
	}
	return nil
}

func handleVideoStopOrError() {
	pipeline.Close()

	fmt.Printf("starting video pipeline %s\n", pipeline.pipelineStr)
	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	err := C.CString("")
	pipeline.pipeline = C.create_video_pipeline(pipelineStrUnsafe, &err)
	err_str := C.GoString(err)
	if len(err_str) != 0 {
		fmt.Printf("failed to create pipeline %s", err_str)
		return
	}

	pipeline.Open()
	for key, val := range pipeline.properties {
		pipeline.SetProperty(key, val)
	}
}

func (p *Pipeline) Open() {
	fmt.Println("starting video pipeline")
	C.start_video_pipeline(pipeline.pipeline)
}
func (p *Pipeline) Close() {
	fmt.Println("stopping video pipeline")
	C.stop_video_pipeline(p.pipeline)
}


func (p *Pipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *Pipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
