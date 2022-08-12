#ifndef AUDIO_H
#define AUDIO_H

#include <glib.h>
#include <gst/gst.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct _MediaDevice
{
    char sound_capture_device_id[1000];
    char sound_output_device_id[1000];

    char backup_sound_capture_device_id[1000];
    char backup_sound_output_device_id[1000];

    int monitor_handle;
    int backup_monitor_handle;

    char monitor_name[100];
    char backup_monitor_name[100];
}MediaDevice;

extern void   goHandlePipelineBufferAudio             (void *buffer, 
                                                  int bufferLen, 
                                                  int samples);

GstElement *  gstreamer_audio_create_pipeline     (char *pipeline,
                                                   char* device);

void          gstreamer_audio_start_pipeline      (GstElement *pipeline);

void          gstreamer_audio_stop_pipeline       (GstElement *pipeline);

void          gstreamer_audio_start_mainloop      (void);

void*         set_media_device();

int           
string_get_length(void* string);
#endif

