#ifndef __WEBRTC_VIDEO_H__
#define __WEBRTC_VIDEO_H__
#include <stdlib.h>

extern void goHandlePipelineBuffer(void *buffer, int bufferLen, int samples);

void*gstreamer_send_create_pipeline(char *pipeline);
void gstreamer_send_start_pipeline(void*pipeline);
void gstreamer_send_stop_pipeline(void*pipeline);
void gstreamer_send_start_mainloop(void);

#endif
