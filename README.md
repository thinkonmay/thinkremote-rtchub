# Webrtc-proxy
### A webrtc application that forward RTP packet and data channel between two client



# Repository structure

```
|-- cmd                                  | Program
    |-- signalling                       | Signalling server program
    |-- agent                            | Webrtc-proxy-agent program
|-- data-channel                         | Datachannel package
    |-- datachannel.go                   | Datachannel interface definition
|-- listener                             | RTP listener package
    |-- tcp                              | TCP listener 
    |-- udp                              | UDP listener
|-- signalling                           | Signalling package
    |-- gRPC                             | gRPC signalling client implementation
    |-- websocket                        | websocket signalling client implementation
    |-- signalling.go                    | Signalling interface definition
|-- util                                 | Utilities package
    |-- config                           | Configuration (decode from environment variable)
|-- proxy.go                             | Webrtc-proxy-agent 