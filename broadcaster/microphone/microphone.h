#ifndef __WEBRTC_MICROPHONE_H__
#define __WEBRTC_MICROPHONE_H__
#include <stdlib.h>
void*create_sink_pipeline(char *pipeline,void** err);
void start_sink_pipeline(void*pipeline);
void stop_sink_pipeline(void*pipeline);
void start_sink_mainloop(void);

extern void handleSinkStopOrError            ();
void goPushSinkBuffer (void* pipeline,void *buffer, int bufferLen);

int  string_get_length(void* string);
#endif
