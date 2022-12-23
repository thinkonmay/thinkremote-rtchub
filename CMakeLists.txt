cmake_minimum_required(VERSION 2.20.0)
project(WebrtcProxy)


set(CMAKE_PREFIX_PATH C:/gstreamer/1.0/msvc_x86_64)
set(INCLUDE_LIBRARY C:/gstreamer/1.0/msvc_x86_64/lib)

find_package(PkgConfig REQUIRED)

pkg_check_modules(GLIB REQUIRED glib-2.0)
pkg_check_modules(GST REQUIRED gstreamer-1.0)
pkg_check_modules(VIDEO REQUIRED gstreamer-video-1.0)

set(ADS_FOLDER adaptive/ads)
set(ADS_HEADER 
    ${ADS_FOLDER}/header
    ${ADS_FOLDER}/src
    ${ADS_FOLDER}/utils/queue
    ${ADS_FOLDER}/utils/macros
    ${ADS_FOLDER}/utils/logs
    ${ADS_FOLDER}/utils/event
    ${ADS_FOLDER}/utils/buffer
    ${ADS_FOLDER}/utils/array
    ${ADS_FOLDER}/utils/map
    ${ADS_FOLDER}/utils/timeseries
    ${ADS_FOLDER}/utils
)
set(ADS_SOURCE
	${ADS_FOLDER}/utils/*.cpp
	${ADS_FOLDER}/utils/queue/*.cpp
	${ADS_FOLDER}/utils/macros/*.cpp
	${ADS_FOLDER}/utils/logs/*.cpp
	${ADS_FOLDER}/utils/event/*.cpp
	${ADS_FOLDER}/utils/buffer/*.cpp
	${ADS_FOLDER}/utils/array/*.cpp
	${ADS_FOLDER}/utils/timeseries/*.cpp
	${ADS_FOLDER}/utils/map/*.cpp
	${ADS_FOLDER}/src/*.cpp 
)


file(GLOB SOURCE_LIST 
    ${CMAKE_SOURCE_DIR}/cgo/*.cpp
    ${ADS_SOURCE}
)


add_library(shared STATIC ${SOURCE_LIST})
target_include_directories(shared PUBLIC
    ${ADS_HEADER}

    ${CMAKE_SOURCE_DIR}/broadcaster/gstreamer 

    ${CMAKE_SOURCE_DIR}/listener/audio
    ${CMAKE_SOURCE_DIR}/listener/video

    ${CMAKE_SOURCE_DIR}/adaptive/ads/utils
    ${CMAKE_SOURCE_DIR}/adaptive/ads/header
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
	# libstdc++.a
    ${GLIB_LIBRARIES} 
    ${GST_LIBRARIES}
    ${VIDEO_LIBRARIES}
)

add_executable(testCGo ${CMAKE_SOURCE_DIR}/cgo/test/main.cpp)
target_include_directories( testCGo PUBLIC ${CMAKE_SOURCE_DIR}/adaptive)
target_link_libraries(testCGo PUBLIC shared )

add_executable(testPipeline ${CMAKE_SOURCE_DIR}/cgo/test/dynamic_pipeline.cpp)
target_include_directories( testPipeline PUBLIC 
    ${GLIB_INCLUDE_DIRS} 
    ${GST_INCLUDE_DIRS}
    ${VIDEO_INCLUDE_DIRS}
)
target_link_directories(testPipeline PUBLIC ${INCLUDE_LIBRARY} )    
target_link_libraries(testPipeline PUBLIC 
    ${GLIB_LIBRARIES} 
    ${GST_LIBRARIES}
    ${VIDEO_LIBRARIES}
)