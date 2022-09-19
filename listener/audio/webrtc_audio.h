#ifndef __WEBRTC_AUDIO_H__
#define __WEBRTC_AUDIO_H__
#include <stdint.h>
#include <stdlib.h>


extern void   goHandlePipelineBufferAudio         (void *buffer, 
                                                   int bufferLen, 
                                                   int samples);

extern void          handleAudioStopOrError            ();

void*  create_audio_pipeline            (char *pipeline,
                                         char* device,
                                         void** err);

void          start_audio_pipeline      (void* pipeline);

void          stop_audio_pipeline       (void* pipeline);

void          start_audio_mainloop      (void);

void*         set_media_device();

int           string_get_length(void* string);



#endif