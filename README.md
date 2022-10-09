# Webrtc-proxy
### Peer to peer battery included WebRTC application built base on pion/webrtc





# Dependencies 
  - [Ninja](https://ninja-build.org) - Simple and efficient build tool.  
  - [Msys2](https://ninja-build.org) - Unix like terminal for building Window application using MinGW
    1. [Download Msys2](https://www.msys2.org/) 
    2. Prebuild script (Note: always run in mingw_x86_64 terminal)
      * `pacman -S mingw-w64-x86_64-binutils mingw-w64-x86_64-cmake mingw-w64-x86_64-toolchain mingw-w64-x86_64-make cmake make gcc`
  - [Gstreamer 1.20.3](https://gstreamer.freedesktop.org/) - Multimedia framework for getting monitor/soundcard information and audio encoding
    1. [Download Gstreamer package](https://gstreamer.freedesktop.org/data/pkg/windows/1.19.2/msvc/gstreamer-1.0-devel-msvc-x86_64-1.20.3.msi) 
    2. [Download Gstreamer development package](https://gstreamer.freedesktop.org/data/pkg/windows/1.19.2/msvc/gstreamer-1.0-msvc-x86_64-1.20.3.msi) 
    3. Once the installer is downloaded, install gstreamer to `C:/gstreamer`




# Repository structure

```
|-- cmd                                  | Program
    |-- client                           | example client application
    |-- server                           | example server application
    |-- test                             | Test tool for GStreamer pipeline
|-- cgo                                  | Source code for C-GO binding
|-- hid                                  | HID (mouse/keyboard) adapter 
|-- lib                                  | Static library (built from cgo)
|-- broadcaster                          | RTP broadcaster package
|-- listener                             | RTP listener package
    |-- video                            | Video listener 
    |-- audio                            | Audio listener
    | listener.go                        | Listener interface definition
|-- signalling                           | Signalling package
    |-- gRPC                             | gRPC signalling client 
    |-- websocket                        | websocket signalling client 
    | signalling.go                      | Signalling interface definition
|-- util                                 | Utilities package
    |-- config                           | Configuration 
    |-- child-process                    | Child-process module (use by test module)
    |-- test                             | Gstreamer pipeline testcase
    |-- tool                             | Media device query tool
|-- proxy.go                             | Webrtc-proxy
```