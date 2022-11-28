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





static AdsBufferMap*
ads_algorithm(AdsBufferMap* query_result)
{
    AdsBufferMap* ret = ADS_BUFFER_MAP_CLASS->init();

    {
        AdsBuffer* buf = NULL;
        AdsTimeseries* rtt_arr = NULL;
        int32 data;
        ADS_BUFFER_MAP_CLASS->ref(query_result);
        buf = ADS_BUFFER_MAP_CLASS->get(query_result,"rtt");
        if (buf == NULL) 
            goto donertt;

        if (BUFFER_CLASS->datatype(buf) != AdsDataType::ADS_DATATYPE_BUFFER_TIMESERIES) {
            LOG_ERROR("unknown");
        }
        
        rtt_arr = (AdsTimeseries*)BUFFER_REF(buf,NULL);

        float total = 0, total_period = 0;
        for(int i = 1 ; i < ADS_TIMESERIES_CLASS->length(rtt_arr) + 1;i++) {

            time_point time;
            AdsBuffer* element = ADS_TIMESERIES_CLASS->n_th(rtt_arr,i,&time);
            if (element == NULL)
                break;
            

            nanosecond* data = (nanosecond*)BUFFER_REF(element,NULL);

            int period = (time - TIME_STOP).count();

            total_period += period;
            total += ((*data).count() * period);
        }

        data = (float)total / (float)total_period;

        ADS_TIMESERIES_CLASS->unref(rtt_arr);
        BUFFER_UNREF(buf);
    }
donertt:

    {
        AdsBuffer* bwbuf = NULL;
        AdsTimeseries* bandwidth_arr = NULL;

        int32 data;
        bwbuf = ADS_BUFFER_MAP_CLASS->get(query_result,"bandwidth");
        if (bwbuf == NULL) 
            goto donebw;

        if (BUFFER_CLASS->datatype(bwbuf) != AdsDataType::ADS_DATATYPE_BUFFER_TIMESERIES) {
            LOG_ERROR("unknown");
        }

        bandwidth_arr = (AdsTimeseries*)BUFFER_REF(bwbuf,NULL);

        float total = 0, total_period = 0;
        for(int i = 1 ; i < ADS_TIMESERIES_CLASS->length(bandwidth_arr) + 1;i++) {
            time_point time;
            AdsBuffer* element = ADS_TIMESERIES_CLASS->n_th(bandwidth_arr,i,&time);
            if (element == NULL)
                break;
            
            // START query logic
            nanosecond* data = (nanosecond*)BUFFER_REF(element,NULL);

            int period = (time - TIME_STOP).count();

            total_period += period;
            total += ((*data).count() * period);
            // END query logic
        }

        data = (float)total / (float)total_period;

        ADS_TIMESERIES_CLASS->unref(bandwidth_arr);
        ADS_BUFFER_MAP_CLASS->unref(query_result);
    }
donebw:

    {
        AdsBuffer* bwbuf = NULL;
        AdsTimeseries* bandwidth_arr = NULL;

        int32 data;
        bwbuf = ADS_BUFFER_MAP_CLASS->get(query_result,"decodedFps");
        if (bwbuf == NULL) 
            goto donefps;

        if (BUFFER_CLASS->datatype(bwbuf) != AdsDataType::ADS_DATATYPE_BUFFER_TIMESERIES) {
            LOG_ERROR("unknown");
        }

        bandwidth_arr = (AdsTimeseries*)BUFFER_REF(bwbuf,NULL);

        float total = 0, total_period = 0;
        for(int i = 1 ; i < ADS_TIMESERIES_CLASS->length(bandwidth_arr) + 1;i++) {
            time_point time;
            AdsBuffer* element = ADS_TIMESERIES_CLASS->n_th(bandwidth_arr,i,&time);
            if (element == NULL)
                break;
            
            // START query logic
            int* data = (int*)BUFFER_REF(element,NULL);

            int period = (time - TIME_STOP).count();

            total_period += period;
            total += (*data) * period;
            // END query logic
        }

        data = (float)total / (float)total_period;

        ADS_TIMESERIES_CLASS->unref(bandwidth_arr);
        ADS_BUFFER_MAP_CLASS->unref(query_result);
    }
donefps:

    return ret;

}


void
handle_bitrate_change(AdsBuffer* data)
{

}

void* 
new_ads_context()
{
    AdsEvent* shutdown = NEW_EVENT;
    AdsContext* ret = new_adaptive_context(shutdown,NULL,ads_algorithm);
    add_listener(ret,"bitrate",handle_bitrate_change);
    return(void*)ret;
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
ads_push_video_packets_lost(void* context, int count)
{
    AdsContext* ctx = (AdsContext*)context;
    AdsRecordSource* src = get_record_source(ctx,"packetsLost");
    ads_push_record(src,ADS_DATATYPE_INT32,0,&count);
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