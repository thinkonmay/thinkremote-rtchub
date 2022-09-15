/**
 * @file util.h
 * @author {Do Huy Hoang} ({huyhoangdo0205@gmail.com})
 * @brief 
 * @version 1.0
 * @date 2022-09-15
 * 
 * @copyright Copyright (c) 2022
 * 
 */
int   string_get_length(void* string);

void* query_media_device();
void* get_monitor_device_name(void* data, int pos);
void* get_monitor_name(void* data, int pos);
int   get_monitor_handle(void* data, int pos);
void* get_monitor_adapter(void* data, int pos);
int   monitor_is_active(void* data, int pos);
int   soundcard_is_active(void* data, int pos);
void* get_soundcard_name(void* data, int pos);
void* get_soundcard_device_id(void* data, int pos);