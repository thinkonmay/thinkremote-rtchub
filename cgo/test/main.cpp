/**
 * @file main.cpp
 * @author {Do Huy Hoang} ({huyhoangdo0205@gmail.com})
 * @brief 
 * @version 1.0
 * @date 2022-11-28
 * 
 * @copyright Copyright (c) 2022
 * 
 */


extern "C" {
#include <ads.h>
}

#include <ads_context.h>

#include <thread>
#include <string.h>





int main(void) {
    void* ctx = new_ads_context();
    

    bool shutdown = false;
    std::thread {[&](){
        while (!shutdown) {
            ads_push_rtt(ctx,1000000);
            SLEEP_MILLISEC(1000);
        }
    }}.detach();
    std::thread {[&](){
        while (!shutdown) {
            ads_push_video_packets_lost(ctx,1);
            SLEEP_MILLISEC(1000);
        }
    }}.detach();
    std::this_thread::sleep_for(1000s);
    shutdown = true;
    std::this_thread::sleep_for(2s);
}