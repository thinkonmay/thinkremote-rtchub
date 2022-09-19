#include <sink.h>

#include <gst/app/gstappsrc.h>

GMainLoop *sink_main_loop = NULL;
void 
start_sink_mainloop(void) {
  sink_main_loop = g_main_loop_new(NULL, FALSE);

  g_main_loop_run(sink_main_loop);
}

static gboolean 
sink_bus_call(GstBus *bus, 
              GstMessage *msg, 
              gpointer data) 
{
  switch (GST_MESSAGE_TYPE(msg)) {

  case GST_MESSAGE_EOS:
    g_print("End of stream\n");
    exit(1);
    break;

  case GST_MESSAGE_ERROR: {
    gchar *debug;
    GError *error;

    gst_message_parse_error(msg, &error, &debug);
    g_free(debug);

    g_printerr("Error: %s\n", error->message);
    g_error_free(error);
    exit(1);
  }
  default:
    break;
  }

  return TRUE;
}

GstElement *
create_sink_pipeline(char *pipeline) 
{
  gst_init(NULL, NULL);
  GError *error = NULL;
  return gst_parse_launch(pipeline, &error);
}

void 
start_sink_pipeline(GstElement *pipeline) 
{
    GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
    gst_bus_add_watch(bus, sink_bus_call, NULL);
    gst_object_unref(bus);

    gst_element_set_state(pipeline, GST_STATE_PLAYING);
}

void 
stop_sink_pipeline(GstElement *pipeline) 
{ 
    gst_element_set_state(pipeline, GST_STATE_NULL); 
}

void 
push_sink_buffer(GstElement *pipeline, 
                 void *buffer, 
                 int len) 
{
    GstElement *src = gst_bin_get_by_name(GST_BIN(pipeline), "src");
    if (src != NULL) 
    {
        gpointer p = g_memdup(buffer, len);
        GstBuffer *buffer = gst_buffer_new_wrapped(p, len);
        gst_app_src_push_buffer(GST_APP_SRC(src), buffer);
        gst_object_unref(src);
    }
}
