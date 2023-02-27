cmake_minimum_required(VERSION 2.20.0)
project(WebrtcProxy)


set(CMAKE_PREFIX_PATH C:/gstreamer/1.0/msvc_x86_64)
set(INCLUDE_LIBRARY C:/gstreamer/1.0/msvc_x86_64/lib)

find_package(PkgConfig REQUIRED)

pkg_check_modules(GLIB REQUIRED glib-2.0)
pkg_check_modules(GST REQUIRED gstreamer-1.0)
pkg_check_modules(VIDEO REQUIRED gstreamer-video-1.0)




file(GLOB SOURCE_LIST 
    ${CMAKE_SOURCE_DIR}/cgo/*.cpp
)


add_library(shared STATIC ${SOURCE_LIST})
target_include_directories(shared PUBLIC
    ${CMAKE_SOURCE_DIR}/broadcaster/gstreamer 

    ${CMAKE_SOURCE_DIR}/listener/audio
    ${CMAKE_SOURCE_DIR}/listener/video


    ${CMAKE_SOURCE_DIR}/adaptive

    ${CMAKE_SOURCE_DIR}/util/tool

    ${GLIB_INCLUDE_DIRS} 
    ${GST_INCLUDE_DIRS}
    ${VIDEO_INCLUDE_DIRS}
)

target_compile_options(shared PUBLIC ${GLIB_CFLAGS}
                                     ${GST_CFLAGS}
                                     ${VIDEO_CFLAGS})

target_link_directories(shared PUBLIC ${INCLUDE_LIBRARY} )    

target_link_libraries(shared PUBLIC 
    ${GLIB_LIBRARIES} 
    ${GST_LIBRARIES}
    ${VIDEO_LIBRARIES}
)

