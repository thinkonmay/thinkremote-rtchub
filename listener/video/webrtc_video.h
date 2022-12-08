#ifndef __WEBRTC_VIDEO_H__
#define __WEBRTC_VIDEO_H__
#include <stdlib.h>

extern void goHandlePipelineBufferVideo(void *buffer, int bufferLen, int samples);
extern void handleVideoStopOrError            ();

void*create_video_pipeline(char *pipeline,void** err);
void video_pipeline_set_bitrate(void* pipeline, int bitrate);
void start_video_pipeline(void*pipeline);
void stop_video_pipeline(void*pipeline);
void start_video_mainloop(void);

#endif
