#include <webrtc_audio.h>
#include <gst/app/gstappsrc.h>

GMainLoop *gstreamer_audio_main_loop = NULL;
void gstreamer_audio_start_mainloop(void) {
  gstreamer_audio_main_loop = g_main_loop_new(NULL, FALSE);
  g_main_loop_run(gstreamer_audio_main_loop);
}


static void
device_foreach(GstDevice* device, 
               gpointer data)
{
    MediaDevice* source = (MediaDevice*) data;

    gchar* name = gst_device_get_display_name(device);
    gchar* klass = gst_device_get_device_class(device);
    GstCaps* cap = gst_device_get_caps(device);
    GstStructure* cap_structure = gst_caps_get_structure (cap, 0);
    GstStructure* device_structure = gst_device_get_properties(device);
    gchar* cap_name = (gchar*)gst_structure_get_name(cap_structure);
    gchar* api = (gchar*)gst_structure_get_string(device_structure,"device.api");

    
    if(!g_strcmp0(api,"wasapi2"))
    {
        gboolean is_default;
        gchar* id;

        id = (gchar*)gst_structure_get_string(device_structure,"device.strid");
        id = id ? id : (gchar*)gst_structure_get_string(device_structure,"device.id");
        gst_structure_get_boolean(device_structure,"device.default",&is_default);

        if(!g_strcmp0(klass,"Audio/Source") &&
           !g_strcmp0(cap_name,"audio/x-raw"))
        {
            if(g_str_has_prefix(name,"CABLE Input"))
                memcpy(source->sound_output_device_id,id,strlen(id));
            else if(is_default && !g_str_has_prefix(name,"Default Audio Capture"))
                memcpy(source->backup_sound_output_device_id,id,strlen(id));
        }

        if(!g_strcmp0(klass,"Audio/Sink") &&
           !g_strcmp0(cap_name,"audio/x-raw"))
        {
            if(g_str_has_prefix(name,"CABLE"))
                memcpy(source->sound_capture_device_id,id,strlen(id));
            else if(is_default)
                memcpy(source->backup_sound_capture_device_id,id,strlen(id));
        }
    }

    gst_caps_unref(cap);
    g_object_unref(device);
}


void*
set_media_device()
{
    static MediaDevice dev = {0};

    GstDeviceMonitor* monitor = gst_device_monitor_new();
    if(!gst_device_monitor_start(monitor)) {
        return NULL;
    }

    GList* device_list = gst_device_monitor_get_devices(monitor);
    g_list_foreach(device_list,(GFunc)device_foreach,&dev);

    MediaDevice* source = &dev;
    return source->sound_capture_device_id ? source->sound_capture_device_id : source->backup_sound_capture_device_id;
}

static gboolean gstreamer_audio_bus_call(GstBus *bus, GstMessage *msg, gpointer data) {
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

GstFlowReturn gstreamer_audio_new_sample_handler(GstElement *object, gpointer user_data) {
  GstSample *sample = NULL;
  GstBuffer *buffer = NULL;
  gsize copy_size = 0;
  char copy[1000] = {0};

  g_signal_emit_by_name (object, "pull-sample", &sample);
  if (sample) {
    buffer = gst_sample_get_buffer(sample);
    if (buffer) {
      copy_size = gst_buffer_get_size(buffer);

      gst_buffer_extract(buffer, 0, (gpointer)copy, copy_size);
      if(copy || copy_size)
        goHandlePipelineBufferAudio(copy, copy_size, GST_BUFFER_DURATION(buffer));
    }
    gst_sample_unref (sample);
  }

  return GST_FLOW_OK;
}

void* gstreamer_audio_create_pipeline(char *pipeline,
                                            char* device) 
{
  gst_init(NULL, NULL);
  GError *error = NULL;
  GstElement * ret = gst_parse_launch(pipeline, &error);
  GstElement *src = gst_bin_get_by_name(GST_BIN(ret), "source");
  g_object_set(src,"device",device,NULL);
  return ret;
}

void gstreamer_audio_start_pipeline(void* pipeline) {

  
  GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
  gst_bus_add_watch(bus, gstreamer_audio_bus_call, NULL);
  gst_object_unref(bus);


  GstElement *appsink = gst_bin_get_by_name(GST_BIN(pipeline), "appsink");
  g_object_set(appsink, "emit-signals", TRUE, NULL);
  g_signal_connect(appsink, "new-sample", G_CALLBACK(gstreamer_audio_new_sample_handler), NULL);
  gst_object_unref(appsink);

  gst_element_set_state((GstElement*)pipeline, GST_STATE_PLAYING);
}

void gstreamer_audio_stop_pipeline(void* pipeline) {
  gst_element_set_state((GstElement*)pipeline, GST_STATE_NULL);
}


int           
string_get_length(void* string)
{
  return strlen((char*)string);
}