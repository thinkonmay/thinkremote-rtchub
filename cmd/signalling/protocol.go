package signalling

import "github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC/packet"

type Tenant interface {
	Send(*packet.UserResponse)
	Receive(*packet.UserRequest)
	IsExited()
}


type ProtocolHandler interface {
	OnTenant(token string, tent Tenant) (error)
}