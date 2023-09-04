package datachannel

import (
	"fmt"
	"sync"
)

type Msg struct{
	Id string
	Msg string
}

const (
	internal_close = "internal close"
	limit = 10
)

type IDatachannel interface {
	Groups() []string
	Send(group string,id string, pkt string)

	RegisterHandle(group string,id string, handler func(msg string))
	DeregisterHandle(group string,id string)

	RegisterConsumer(group string, consumer DatachannelConsumer) 
}

type DatachannelConsumer interface {
	Send(id string,pkt string)
	Recv() (id string,pkt string)
	SetContext(id []string)
}

type Handler struct {
	handler func(msg string)
	handle_queue chan Msg 
}

type DatachannelGroup struct {
	send chan Msg
	recv chan Msg 

	mutext *sync.Mutex
	handlers map[string]*Handler

	consumer DatachannelConsumer
}

type Datachannel struct {
	groups map[string]*DatachannelGroup
}

func NewDatachannel(names ...string) IDatachannel {
	dc := &Datachannel{
		groups :map[string]*DatachannelGroup{},
	}


	group_loop := func(group *DatachannelGroup) {
		for {
			msg := <-group.recv

			group.mutext.Lock()
			for _,handler := range group.handlers{
				handler.handle_queue<-msg
			}
			group.mutext.Unlock()
		}
	}

	for _,name := range names {
		dc.groups[name] = &DatachannelGroup{
			send:    make(chan Msg,limit),
			recv:    make(chan Msg,limit),
			handlers: map[string]*Handler{},
			mutext: &sync.Mutex{},
		}

		go group_loop(dc.groups[name])
	}
	return dc
}
func (dc *Datachannel) Groups()[]string {
	keys := make([]string, 0, len(dc.groups))
	for k := range dc.groups {
		keys = append(keys, k)
	}
	
	return keys
}







func (dc *Datachannel) Send(group string,id string, pkt string) {
	if dc.groups[group] == nil {
		return
	} else if len(dc.groups[group].send) == limit {
		return
	}

	dc.groups[group].send<-Msg{ Id:id, Msg:pkt }
}












func (dc *Datachannel) RegisterHandle(group_name string, 
									  id string, 
									  handler func(msg string)) {
	group := dc.groups[group_name]
	if group == nil {
		fmt.Printf("no group name %s available\n",group)
		return
	}

	new_handler := &Handler{
		handler : handler,
		handle_queue: make(chan Msg,limit),
	}

	group.mutext.Lock()
	defer group.mutext.Unlock()

	group.handlers[id] = new_handler
	dc.setContext(group_name)

	go func() { for { 
		msg:=<-new_handler.handle_queue
		if msg.Id != id {
			continue
		} else if msg.Msg == internal_close { 
			fmt.Println("closed data channel handler")
			return 
		}


		new_handler.handler(msg.Msg) 
	}}()

}











func (dc *Datachannel) DeregisterHandle(group_name string,id string) {
	group := dc.groups[group_name] 
	if group == nil {
		return
	} 

	handler := group.handlers[id]
	if handler == nil {
		fmt.Printf("no handler name %s available\n",id)
		return
	}

	group.mutext.Lock()
	defer group.mutext.Unlock()

	handler.handle_queue<-Msg{Msg: internal_close}
	delete(group.handlers,id)
	dc.setContext(group_name)
}


func (dc *Datachannel) RegisterConsumer(group_name string, consumer DatachannelConsumer) {
	group := dc.groups[group_name]
	if group == nil {
		fmt.Printf("no group name %s available\n",group)
		return
	} else if group.consumer != nil{
		fmt.Printf("consumer for group %s available\n",group)
		return
	}

	group.consumer = consumer
	go func() { for {
		id,msg := consumer.Recv()
		group.recv<-Msg{
			Id: id,
			Msg: msg,
		}
	}}()
	go func() { for {
		pkt :=<- group.send
		consumer.Send(pkt.Id,pkt.Msg)
	}}()
}












func (dc *Datachannel) setContext(group string) {
	keys := make([]string, 0, len(dc.groups[group].handlers))
	for k := range dc.groups[group].handlers {
		keys = append(keys, k)
	}
	
	dc.groups[group].consumer.SetContext(keys)
}
