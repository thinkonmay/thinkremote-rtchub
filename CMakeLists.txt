cmake_minimum_required(VERSION 2.20.0)
project(WebrtcProxy)


set(CMAKE_PREFIX_PATH C:/gstreamer/1.22.0/msvc_x86_64)
set(INCLUDE_LIBRARY C:/gstreamer/1.22.0/msvc_x86_64/lib)

find_package(PkgConfig REQUIRED)

pkg_check_modules(GLIB REQUIRED glib-2.0)
pkg_check_modules(GST REQUIRED gstreamer-1.0)
pkg_check_modules(VIDEO REQUIRED gstreamer-video-1.0)




file(GLOB SOURCE_LIST_LISTENER
    ${CMAKE_SOURCE_DIR}/listener/cgo/*.cpp
)
file(GLOB SOURCE_LIST_BROADCASTER
    ${CMAKE_SOURCE_DIR}/broadcaster/cgo/*.cpp
)


add_library(listener    STATIC ${SOURCE_LIST_LISTENER})
add_library(broadcaster STATIC ${SOURCE_LIST_BROADCASTER})

target_include_directories(listener PUBLIC
    ${CMAKE_SOURCE_DIR}/listener/audio
    ${CMAKE_SOURCE_DIR}/listener/video
    ${GLIB_INCLUDE_DIRS} 
    ${GST_INCLUDE_DIRS}
    ${VIDEO_INCLUDE_DIRS}
)
target_include_directories(broadcaster PUBLIC
    ${CMAKE_SOURCE_DIR}/broadcaster/microphone
    ${GLIB_INCLUDE_DIRS} 
    ${GST_INCLUDE_DIRS}
    ${VIDEO_INCLUDE_DIRS}
)

target_compile_options(listener PUBLIC ${GLIB_CFLAGS}
                                     ${GST_CFLAGS}
                                     ${VIDEO_CFLAGS})

target_link_directories(listener PUBLIC ${INCLUDE_LIBRARY} )    

target_link_libraries(listener PUBLIC 
    ${GLIB_LIBRARIES} 
    ${GST_LIBRARIES}
    ${VIDEO_LIBRARIES}
)

target_compile_options(broadcaster PUBLIC ${GLIB_CFLAGS}
                                     ${GST_CFLAGS}
                                     ${VIDEO_CFLAGS})

target_link_directories(broadcaster PUBLIC ${INCLUDE_LIBRARY} )    

target_link_libraries(broadcaster PUBLIC 
    ${GLIB_LIBRARIES} 
    ${GST_LIBRARIES}
    ${VIDEO_LIBRARIES}
)