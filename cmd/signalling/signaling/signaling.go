package signalling

import (
	"fmt"
	"sync"
	"time"

	grpc "github.com/pigeatgarlic/webrtc-proxy/cmd/signalling/gRPC"
	"github.com/pigeatgarlic/webrtc-proxy/cmd/signalling/protocol"
	"github.com/pigeatgarlic/webrtc-proxy/cmd/signalling/websocket"
	"github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC/packet"
)


type Signalling struct {
	waitLine []*WaitingTenant
	pairs    []*Pair
	
	handlers []protocol.ProtocolHandler
	mut 	 *sync.Mutex
}

func removePair(slice []*Pair, s int) []*Pair {
    return append(slice[:s], slice[s+1:]...)
}
func removeTenant(slice []*WaitingTenant, s int) []*WaitingTenant{
    return append(slice[:s], slice[s+1:]...)
}

func ProcessReq(req *packet.UserRequest)*packet.UserResponse  {
	if req == nil {
		return nil;			
	}

	var res packet.UserResponse; 
	res.Data = req.Data
	if req.Target == "ICE" || req.Target == "SDP" {
		res.Data["Target"] = req.Target;
	}

	return &res;
}



type WaitingTenant struct {
	stop bool
	token string
	waiter protocol.Tenant
}

func (wait *WaitingTenant) handle(){
	go func() {
		for { 
			if wait.stop {
				return;	
			}	

			if wait.waiter.IsExited() {
				wait.stop = true;
				return;
			}

			// wait.waiter.Receive();
		}	
	}()
}

type Pair struct {
	Id     string
	stop   bool
	client protocol.Tenant
	worker protocol.Tenant
}

func (pair *Pair) handlePair(){
	go func ()  {
		for {
			if pair.stop {
				return;	
			}
			dat := pair.worker.Receive()	
			pair.client.Send(ProcessReq(dat))
		}	
	}()

	go func ()  {
		for {
			if pair.stop {
				return;	
			}
			dat := pair.client.Receive()	
			pair.worker.Send(ProcessReq(dat))
		}	
	}()
	go func ()  {
		for {
			time.Sleep(10*time.Millisecond);
			if pair.client.IsExited() || pair.worker.IsExited() {
				pair.stop = true;	
			}
		}	
	}()
}

func InitSignallingServer(conf *protocol.SignalingConfig) *Signalling {
	var err error
	var signaling Signalling
	signaling.handlers = make([]protocol.ProtocolHandler, 2)
	signaling.pairs = make([]*Pair, 0)
	signaling.waitLine = make([]*WaitingTenant, 0)
	signaling.mut = &sync.Mutex{}

	signaling.handlers[0] = grpc.InitSignallingServer(conf);
	signaling.handlers[1] = ws.InitSignallingWs(conf);
	if err != nil  {
		fmt.Printf("%s\n",err.Error())
		return nil;
	}

	fun := func (token string, tent protocol.Tenant) error {
		signaling.mut.Lock()
		var found bool;
		var waiter_index int;
		var waiter *WaitingTenant;
		for index,wait := range signaling.waitLine{
			if wait.token == token {
				waiter_index = index;
				waiter = wait;	
				found = true;
			}
		}


		if found{
			waiter.waiter.Send(&packet.UserResponse{
				Id: 0,	
				Error: "",
				Data: map[string]string{
					"Target": "START",
				},
			})

			waiter.stop = true;
			signaling.waitLine = removeTenant(signaling.waitLine,waiter_index);
			pair := &Pair{
				Id: token,
				client: waiter.waiter,
				worker: tent,
			}
			signaling.pairs = append(signaling.pairs, pair)	
			pair.handlePair()
		} else {
			wait := &WaitingTenant{
				stop: false,
				token: token,
				waiter: tent,
			};
			signaling.waitLine = append(signaling.waitLine, wait)
			wait.handle()
		}
		signaling.mut.Unlock()
		return nil;
	};

	go func() {
		for {
			signaling.mut.Lock()
			var rev []int;
			for index, pair := range signaling.pairs{
				if pair.stop {
					rev = append(rev, index);	
				}
			}
			for i := range rev {
				signaling.pairs = removePair(signaling.pairs,i);
			}
			signaling.mut.Unlock()
			time.Sleep(10*time.Millisecond);
		}	
	}()

	go func() {
		for {
			signaling.mut.Lock()
			var rev []int;
			for index, wait := range signaling.waitLine{
				if wait.stop {
					rev = append(rev, index);
				}
			}
			for _,i := range rev {
				signaling.waitLine = removeTenant(signaling.waitLine,i)
			}
			signaling.mut.Unlock()
			time.Sleep(10*time.Millisecond);
		}	
	}()

	for _,handler := range signaling.handlers {
		handler.OnTenant(fun);
	}
	return &signaling;
}


