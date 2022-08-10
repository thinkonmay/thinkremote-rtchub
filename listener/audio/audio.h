#ifndef GST_H
#define GST_H

#include <glib.h>
#include <gst/gst.h>
#include <stdint.h>
#include <stdlib.h>

extern void   goHandlePipelineBufferAudio             (void *buffer, 
                                                  int bufferLen, 
                                                  int samples);

GstElement *  gstreamer_audio_create_pipeline     (char *pipeline);

void          gstreamer_audio_start_pipeline      (GstElement *pipeline);

void          gstreamer_audio_stop_pipeline       (GstElement *pipeline);

void          gstreamer_audio_start_mainloop      (void);

#endif
