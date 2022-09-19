/**
 * @file sink.h
 * @author {Do Huy Hoang} ({huyhoangdo0205@gmail.com})
 * @brief 
 * @version 1.0
 * @date 2022-09-19
 * 
 * @copyright Copyright (c) 2022
 * 
 */
#ifndef GST_H
#define GST_H

#include <glib.h>
#include <gst/gst.h>
#include <stdint.h>
#include <stdlib.h>

GstElement *create_sink_pipeline(char *pipeline);
void start_sink_pipeline(GstElement *pipeline);
void stop_sink_pipeline(GstElement *pipeline);
void push_sink_buffer(GstElement *pipeline, void *buffer, int len);
void start_sink_mainloop(void);

#endif
