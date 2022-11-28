/**
 * @file ads.h
 * @author {Do Huy Hoang} ({huyhoangdo0205@gmail.com})
 * @brief 
 * @version 1.0
 * @date 2022-11-25
 * 
 * @copyright Copyright (c) 2022
 * 
 */



void* new_ads_context();

void ads_push_rtt                                    (void* context, int nanosec);

void ads_push_total_incoming_bandwidth_consumption   (void* context, int bytes);
void ads_push_available_incoming_bandwidth           (void* context, int bytes);

void ads_push_audio_incoming_bandwidth_consumption   (void* context, int bytes);

void ads_push_video_incoming_bandwidth_consumption   (void* context, int bytes);

void ads_push_frame_decoded_per_second               (void* context, int count);
void ads_push_decode_time_per_frame                  (void* context, int nanosecond);
void ads_push_key_frame_per_frame                    (void* context, int count);

void ads_push_video_packets_lost                     (void* context, int count);

void ads_push_video_jitter                           (void* context, int count);
void ads_push_video_jitter_buffer_delay              (void* context, int count);





