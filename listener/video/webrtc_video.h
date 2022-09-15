#ifndef __WEBRTC_VIDEO_H__
#define __WEBRTC_VIDEO_H__
#include <stdlib.h>

extern void goHandlePipelineBuffer(void *buffer, int bufferLen, int samples);

void*create_video_pipeline(char *pipeline,void** err);
void start_video_pipeline(void*pipeline);
void stop_video_pipeline(void*pipeline);
void start_video_mainloop(void);

#endif
