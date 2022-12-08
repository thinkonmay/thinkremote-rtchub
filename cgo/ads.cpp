/**
 * @file ads.c
 * @author {Do Huy Hoang} ({huyhoangdo0205@gmail.com})
 * @brief 
 * @version 1.0
 * @date 2022-11-25
 * 
 * @copyright Copyright (c) 2022
 * 
 */


#include <ads_context.h>
extern "C" {
#include <ads.h>
}


typedef void (*QueryFunction) (void* data, 
                               time_point timestamp,
                               void* user_data);


bool
ads_query_buffer_map_contain_time_series(AdsBufferMap* map, 
                                         char* key,
                                         void* user_data,
                                         QueryFunction func)
{
    ADS_BUFFER_MAP_CLASS->ref(map);

    AdsBuffer* buf = ADS_BUFFER_MAP_CLASS->get(map,key);
    if (buf == NULL) {
        LOG_ERROR("key not exist");
        return false;
    }

    if (BUFFER_CLASS->datatype(buf) != AdsDataType::ADS_DATATYPE_BUFFER_TIMESERIES) {
        LOG_ERROR("map buffer not contain timeseries");
        return false;
    }
    
    AdsTimeseries* buf_array = (AdsTimeseries*)BUFFER_REF(buf,NULL);
    for(int i = 1 ; i < ADS_TIMESERIES_CLASS->length(buf_array) + 1;i++) {
        time_point time;
        AdsBuffer* element = ADS_TIMESERIES_CLASS->n_th(buf_array,i,&time);
        if (element == NULL)
            break;
        

        void* _data = BUFFER_REF(element,NULL);
        func(_data,time,user_data);
        BUFFER_UNREF(element);
    }

    ADS_TIMESERIES_CLASS->unref(buf_array);
    BUFFER_UNREF(buf);
    return true;
}


struct AdsCallbackData {
    uint64 rtt_x_timestamp_total;
    uint64 rtt_timestamp_total;

    uint64 decodedfps_x_timestamp_total;
    uint64 decodedfps_timestamp_total;

    uint64 receivedfps_x_timestamp_total;
    uint64 receivedfps_timestamp_total;

    uint64 videoBandwidth_x_timestamp_total;
    uint64 videoBandwidth_timestamp_total;

    uint64 availableIncomingBandwidth_x_timestamp_total;
    uint64 availableIncomingBandwidth_timestamp_total;
};

static void 
query_rtt_field  (void* data, 
                  time_point timestamp,
                  void* user_data)
{
    AdsCallbackData* cbdata = (AdsCallbackData*) user_data;

    cbdata->rtt_x_timestamp_total += ((*(nanosecond*)data).count() * GET_TIMESTAMP_MILLISEC(timestamp));
    cbdata->rtt_timestamp_total += GET_TIMESTAMP_MILLISEC(timestamp);
}


static void 
query_decoded_fps_field  (void* data, 
                  time_point timestamp,
                  void* user_data)
{
    AdsCallbackData* cbdata = (AdsCallbackData*) user_data;

    cbdata->decodedfps_x_timestamp_total += *(int*)data * GET_TIMESTAMP_MILLISEC(timestamp);
    cbdata->decodedfps_timestamp_total += GET_TIMESTAMP_MILLISEC(timestamp);
}


static void 
query_received_fps_field  (void* data, 
                  time_point timestamp,
                  void* user_data)
{
    AdsCallbackData* cbdata = (AdsCallbackData*) user_data;

    cbdata->receivedfps_x_timestamp_total += *(int*)data * GET_TIMESTAMP_MILLISEC(timestamp);
    cbdata->receivedfps_timestamp_total   += GET_TIMESTAMP_MILLISEC(timestamp);
}

static void 
query_video_bandwidth_field  (void* data, 
                              time_point timestamp,
                              void* user_data)
{
    AdsCallbackData* cbdata = (AdsCallbackData*) user_data;

    cbdata->videoBandwidth_x_timestamp_total += *(int*)data * GET_TIMESTAMP_MILLISEC(timestamp);
    cbdata->videoBandwidth_timestamp_total += GET_TIMESTAMP_MILLISEC(timestamp);
}
static void 
query_available_incoming_bandwidth_field  (void* data, 
                              time_point timestamp,
                              void* user_data)
{
    AdsCallbackData* cbdata = (AdsCallbackData*) user_data;

    cbdata->availableIncomingBandwidth_x_timestamp_total += *(int*)data * GET_TIMESTAMP_MILLISEC(timestamp);
    cbdata->availableIncomingBandwidth_timestamp_total   += GET_TIMESTAMP_MILLISEC(timestamp);
}


static AdsBufferMap*
ads_algorithm(AdsBufferMap* query_result)
{
    AdsBufferMap* ret = ADS_BUFFER_MAP_CLASS->init();

    AdsCallbackData data = {0};

    if(ads_query_buffer_map_contain_time_series(query_result,"rtt",&data,query_rtt_field)){
        float medium_rtt = (float)data.rtt_x_timestamp_total / (float)data.rtt_timestamp_total;
        LOG_DEBUG("Medium Round Trip Time: %fms",medium_rtt / 1000000);
    }
    if(ads_query_buffer_map_contain_time_series(query_result,"receivedFps",&data,query_received_fps_field)){
        float medium_receivedfps = (float)data.receivedfps_x_timestamp_total / (float)data.receivedfps_timestamp_total;
        LOG_DEBUG("Medium Received Fps: %dfps",(int)medium_receivedfps);
    }
    if(ads_query_buffer_map_contain_time_series(query_result,"decodedFps",&data,query_decoded_fps_field)){
        float medium_decodedfps = (float)data.decodedfps_x_timestamp_total / (float)data.decodedfps_timestamp_total;
        LOG_DEBUG("Medium Decoded Fps: %dfps",(int)medium_decodedfps);
    }
    if(ads_query_buffer_map_contain_time_series(query_result,"videoBWincoming",&data,query_video_bandwidth_field)){
        float medium = (float)data.videoBandwidth_x_timestamp_total / (float)data.videoBandwidth_timestamp_total;
        LOG_DEBUG("Medium video bandwidth: %fmbps",medium * 8 / 1000000);
    }
    if(ads_query_buffer_map_contain_time_series(query_result,"availableBWincoming",&data,query_available_incoming_bandwidth_field)){
        float medium = (float)data.availableIncomingBandwidth_x_timestamp_total / (float)data.availableIncomingBandwidth_timestamp_total;
        LOG_DEBUG("Medium available incoming bandwidth: %fmbps",medium / 1000000);
    }

    return ret;
}


void
handle_bitrate_change_event(AdsBuffer* data,void* user_data)
{
    int bitrate = 0;
    bitrate = *(int*)BUFFER_REF(data,NULL);
    *(int*)user_data = bitrate;
    BUFFER_UNREF(data);
}

void* 
new_ads_context()
{
    AdsEvent* shutdown = NEW_EVENT;
    AdsContext* ret = new_adaptive_context(shutdown,NULL,ads_algorithm);
    add_record_source(ret,"rtt");
    add_record_source(ret,"totalBWincoming");
    add_record_source(ret,"availableBWincoming");
    add_record_source(ret,"audioBWincoming");
    add_record_source(ret,"videoBWincoming");
    add_record_source(ret,"decodedFps");
    add_record_source(ret,"receivedFps");
    add_record_source(ret,"decodeTimeperFrame");
    add_record_source(ret,"keyFrameperFrame");
    add_record_source(ret,"packetsLost");
    add_record_source(ret,"videoJitter");
    add_record_source(ret,"videoJitterBufferDelay");
    return(void*)ret;
}


int
wait_for_bitrate_change(void* ctx)
{
    int val = -1;
    AdsContext* context = (AdsContext*)ctx;
    int id = add_listener_to_ctx(context,"bitrate",handle_bitrate_change_event,(void*)&val);
    while (val == -1) { SLEEP_MILLISEC(10); }
    remove_listener_from_ctx(context,id);
    return val;
}

void 
ads_push_rtt(void* context, int nanosec)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"rtt");
    nanosecond data = NANOSEC(nanosec);
    ads_push_record(src,ADS_DATATYPE_TIMERANGE_NANOSECOND,0,&data);
}

void 
ads_push_total_incoming_bandwidth_consumption(void* context, int bytes)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"totalBWincoming");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&bytes);
}

void 
ads_push_available_incoming_bandwidth(void* context, int bytes)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"availableBWincoming");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&bytes);
}

void 
ads_push_audio_incoming_bandwidth_consumption(void* context, int bytes)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"audioBWincoming");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&bytes);
}

void 
ads_push_video_incoming_bandwidth_consumption(void* context, int bytes)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"videoBWincoming");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&bytes);
}

void 
ads_push_frame_decoded_per_second(void* context, int count)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"decodedFps");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&count);
}

void 
ads_push_frame_received_per_second(void* context, int count)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"receivedFps");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&count);
}

void 
ads_push_decode_time_per_frame(void* context, int nanosec)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"decodeTimeperFrame");
    nanosecond data = NANOSEC(nanosec);
    ads_push_record(src,ADS_DATATYPE_TIMERANGE_NANOSECOND,0,&data);
}

void 
ads_push_key_frame_per_frame(void* context, int count)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"keyFrameperFrame");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&count);
}

void 
ads_push_video_packets_lost(void* context, float count)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"packetsLost");
    ads_push_record(src,ADS_DATATYPE_FLOAT,0,&count);
}

void 
ads_push_video_jitter(void* context, int count)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"videoJitter");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&count);
}

void 
ads_push_video_jitter_buffer_delay(void* context, int count)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"videoJitterBufferDelay");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&count);
}