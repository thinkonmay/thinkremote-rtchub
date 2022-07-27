package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pigeatgarlic/webrtc-proxy/cmd/signalling/protocol"
	"github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC/packet"
)

var wsserver = WebSocketServer{};


type WebSocketServer struct {
	fun protocol.OnTenantFunc
}
func (server *WebSocketServer) OnTenant(fun protocol.OnTenantFunc) {
	server.fun = fun
}

type WebsocketTenant struct {
	exited bool	
	conn *websocket.Conn
	msgType int
}

func (tenant *WebsocketTenant) Send(pkt *packet.UserResponse) {
	if pkt == nil {
		return;	
	}
	data, err := json.Marshal(pkt);
	if err != nil {
		return;
	}
	tenant.conn.WriteMessage(websocket.TextMessage,data);
}



func (tenant *WebsocketTenant) Receive() *packet.UserRequest {
	msgt, data, err := tenant.conn.ReadMessage()
	if err != nil {
		return nil;
	}
	switch msgt {
	case websocket.CloseMessage:	
		tenant.exited = true;
		return nil;
	case websocket.TextMessage:
	}

	var req packet.UserRequest;
	err = json.Unmarshal(data,&req)
	if err != nil {
		return nil;
	}
	return &req;
}

func (tenant *WebsocketTenant) Exit() {
	tenant.exited = true;
}

func (tenant *WebsocketTenant) IsExited() bool {
	return tenant.exited
}


var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	var tenant WebsocketTenant;
	c, err := upgrader.Upgrade(w, r, nil)
	defer c.Close()
	if err != nil {
		token := r.Header["authorization"]
		tenant.exited = false;
		tenant.conn = c;
		wsserver.fun(token[0],&tenant);
		return
	}

	for {
		if tenant.exited == true {
			return;
		}
		time.Sleep(time.Millisecond);
	}
}

func InitSignallingWs(conf *protocol.SignalingConfig) *WebSocketServer{
	http.HandleFunc("/echo", echo)
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d",conf.WebsocketPort), nil)
	return &wsserver;
}
