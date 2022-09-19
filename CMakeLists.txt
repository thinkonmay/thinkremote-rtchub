cmake_minimum_required(VERSION 2.20.0)
project(WebrtcProxy)



set(CMAKE_PREFIX_PATH C:/gstreamer/1.0/msvc_x86_64)
set(INCLUDE_LIBRARY C:/gstreamer/1.0/msvc_x86_64/lib)

foreach( OUTPUTCONFIG ${CMAKE_CONFIGURATION_TYPES} )
    string( TOUPPER ${OUTPUTCONFIG} OUTPUTCONFIG )
    set( CMAKE_ARCHIVE_OUTPUT_DIRECTORY_${OUTPUTCONFIG} ${CMAKE_SOURCE_DIR}/bin )
    set( CMAKE_LIBRARY_OUTPUT_DIRECTORY_${OUTPUTCONFIG} ${CMAKE_SOURCE_DIR}/bin )
    set( CMAKE_RUNTIME_OUTPUT_DIRECTORY_${OUTPUTCONFIG} ${CMAKE_SOURCE_DIR}/bin )
endforeach( OUTPUTCONFIG CMAKE_CONFIGURATION_TYPES )


set(CMAKE_RUNTIME_OUTPUT_DIRECTORY  ${CMAKE_SOURCE_DIR}/bin) 



find_package(PkgConfig REQUIRED)

pkg_check_modules(GLIB REQUIRED glib-2.0)
pkg_check_modules(GOBJECT REQUIRED gobject-2.0)
pkg_check_modules(GST REQUIRED gstreamer-1.0)
pkg_check_modules(WEBRTC REQUIRED gstreamer-webrtc-1.0)
pkg_check_modules(VIDEO REQUIRED gstreamer-video-1.0)


file(GLOB SOURCE_LIST 
${CMAKE_SOURCE_DIR}/cgo/*.cpp
${CMAKE_SOURCE_DIR}/cgo/*.c
${CMAKE_SOURCE_DIR}/listener/audio/*.cc
${CMAKE_SOURCE_DIR}/listener/video/*.cc
)

add_library(shared STATIC ${SOURCE_LIST})

target_include_directories(shared PUBLIC
    ${CMAKE_SOURCE_DIR}/broadcaster/gstreamer 
    ${CMAKE_SOURCE_DIR}/listener/audio
    ${CMAKE_SOURCE_DIR}/listener/video
    ${CMAKE_SOURCE_DIR}/cmd/tool
    ${GLIB_INCLUDE_DIRS} 
    ${GOBJECT_INCLUDE_DIRS}
    ${GST_INCLUDE_DIRS}
    ${WEBRTC_INCLUDE_DIRS}
    ${VIDEO_INCLUDE_DIRS}
)


set_target_properties(shared PROPERTIES 
 RUNTIME_OUTPUT_DIRECTORY ${CMAKE_SOURCE_DIR}/bin)

target_compile_options(shared PUBLIC ${GLIB_CFLAGS})
target_compile_options(shared PUBLIC ${GOBJECT_CFLAGS})
target_compile_options(shared PUBLIC ${GST_CFLAGS})
target_compile_options(shared PUBLIC ${WEBRTC_CFLAGS})
target_compile_options(shared PUBLIC ${VIDEO_CFLAGS})

target_link_directories(shared PUBLIC ${INCLUDE_LIBRARY} )    

target_link_libraries(shared PUBLIC 
	libstdc++.a
    ${GLIB_LIBRARIES} 
    ${GOBJECT_LIBRARIES}
    ${GST_LIBRARIES}
    ${WEBRTC_LIBRARIES}
    ${VIDEO_LIBRARIES}
)




