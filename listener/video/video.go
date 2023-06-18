package video

/*

#include <gst/app/gstappsink.h>
#include <gst/video/video-event.h>
#include <stdint.h>
#include <stdlib.h>
#include <glib.h>

void
start_video_mainloop(void) {
    GMainLoop *gstreamer_video_main_loop = NULL;
    gstreamer_video_main_loop = g_main_loop_new(NULL, FALSE);

    g_main_loop_run(gstreamer_video_main_loop);
}


typedef struct _VideoPipeline {
    GstElement* appsink;
    GstElement* pipeline;
    GstElement* framerate_filter;
    GstElement* encoder;

    int frame_count;
} VideoPipeline;





void video_pipeline_stop(void* pipeline);
void video_pipeline_start(void* pipeline);




static void
handle_pipeline_failure(VideoPipeline* video_pipeline)
{
    video_pipeline_stop(video_pipeline);
    video_pipeline_start(video_pipeline);
}


static gboolean
gstreamer_send_video_bus_call(GstBus *bus,
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
create_video_pipeline(char *pipeline_desc,
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
    gst_bus_add_watch(bus, gstreamer_send_video_bus_call, pipeline);
    gst_object_unref(bus);

    pipeline->appsink            = gst_bin_get_by_name(GST_BIN(pipeline->pipeline), "appsink");
    pipeline->framerate_filter   = gst_bin_get_by_name(GST_BIN(pipeline->pipeline), "framerateFilter");
    pipeline->encoder            = gst_bin_get_by_name(GST_BIN(pipeline->pipeline), "encoder");
    *err = NULL;
    return pipeline;
}

void
video_pipeline_start(void* pipelineIn) {
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_PIPELINE(pipeline->pipeline))
        return;

    gst_element_set_state(pipeline->pipeline, GST_STATE_PLAYING);
}

void
video_pipeline_stop(void* pipelineIn) {
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_PIPELINE(pipeline->pipeline))
        return;
    else if (!GST_IS_ELEMENT(pipeline->appsink))
        return;
    else if (!GST_IS_ELEMENT(pipeline->framerate_filter))
        return;
    else if (!GST_IS_ELEMENT(pipeline->encoder))
        return;

    pipeline->frame_count = 0;
    gst_element_set_state(pipeline->pipeline, GST_STATE_NULL);
}



void
video_pipeline_set_framerate(void* pipelineIn,
                             int framerate) {
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_ELEMENT(pipeline->framerate_filter))
        return;

    char* capsstr = g_strdup_printf ("video/x-raw(memory:D3D11Memory),framerate=%d/1",framerate);
    GstCaps* caps = gst_caps_from_string (capsstr);
    g_free (capsstr);

    g_object_set (pipeline->framerate_filter, "caps", caps, NULL);
    gst_caps_unref (caps);
}


void
video_pipeline_set_bitrate(void* pipelineIn,
                            int bitrate) {
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_ELEMENT(pipeline->encoder))
        return;

    g_object_set(pipeline->encoder, "bitrate", bitrate, NULL);
}


void
video_pipeline_generate_idr(void* pipelineIn) {
    VideoPipeline* pipeline = (VideoPipeline*)pipelineIn;
    if (!pipeline)
        return;
    else if (!GST_IS_ELEMENT(pipeline->encoder))
        return;

    GstPad* srcpad = gst_element_get_static_pad(pipeline->encoder,"src");
    if (!srcpad)
        return;

    GstEvent* event = gst_video_event_new_upstream_force_key_unit(GST_CLOCK_TIME_NONE,TRUE,0);
    gst_pad_send_event(srcpad,event);
    gst_object_unref(srcpad);
}


int
video_pipeline_pop_buffer	(void* pipelineIn,
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

	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	err := C.CString("")
	fmt.Printf("starting video pipeline %s\n", pipelineStr)
	pipeline.pipeline = C.create_video_pipeline(pipelineStrUnsafe, &err)
	err_str := C.GoString(err)

	if len(err_str) != 0 {
		return nil, fmt.Errorf("failed to create pipeline %s", err_str)
	}

	go func() {
		var duration C.int
		var size C.int
		var samples uint32
		buffer := make([]byte, 100*1000*1000) //100MB
		for {
			size = C.video_pipeline_pop_buffer(
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

func (p *VideoPipeline) GetCodec() string {
	return p.codec
}

func (pipeline *VideoPipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		C.video_pipeline_set_bitrate(pipeline.pipeline, C.int(val))
		C.video_pipeline_generate_idr(pipeline.pipeline)
	case "framerate":
		pipeline.properties["framerate"] = val
		C.video_pipeline_set_framerate(pipeline.pipeline, C.int(val))
		C.video_pipeline_generate_idr(pipeline.pipeline)
	case "reset":
		C.video_pipeline_generate_idr(pipeline.pipeline)
	default:
		return fmt.Errorf("unknown prop")
	}
	return nil
}

func (p *VideoPipeline) Open() {
	fmt.Println("starting video pipeline")
	C.video_pipeline_start(p.pipeline)
}
func (p *VideoPipeline) Close() {
	fmt.Println("stopping video pipeline")
	C.video_pipeline_stop(p.pipeline)
}

func (p *VideoPipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *VideoPipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
