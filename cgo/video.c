#include <webrtc_video.h>

#include <gst/app/gstappsrc.h>
#include <stdint.h>
#include <stdlib.h>
#include <glib.h>

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
handle_audio_sample(GstElement *object, gpointer user_data) {
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
            goHandlePipelineBuffer(copy, copy_size, GST_BUFFER_DURATION(buffer));
            free(copy);
        }
        }
        gst_sample_unref (sample);
    }
    return GST_FLOW_OK;
}

void* 
create_video_pipeline(char *pipeline, 
                      void** err) {
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
    GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
    gst_bus_add_watch(bus, gstreamer_send_video_bus_call, NULL);
    gst_object_unref(bus);

    GstElement *appsink = gst_bin_get_by_name(GST_BIN(pipeline), "appsink");
    g_object_set(appsink, "emit-signals", TRUE, NULL);
    g_signal_connect(appsink, "new-sample", G_CALLBACK(handle_audio_sample), NULL);
    gst_object_unref(appsink);

    gst_element_set_state((GstElement*)pipeline, GST_STATE_PLAYING);
}




void 
video_pipeline_set_bitrate(void* pipeline, int bitrate) {
    GstElement *encoder = gst_bin_get_by_name(GST_BIN(pipeline), "encoder");
    g_object_set(encoder, "bitrate", bitrate, NULL);
}

void 
stop_video_pipeline(void* pipeline) {
    gst_element_set_state((GstElement*)pipeline, GST_STATE_NULL);
}