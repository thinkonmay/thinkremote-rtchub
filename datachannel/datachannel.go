package datachannel

import (
	"fmt"
	"sync"
)

const (
	internal_close = "internal close"
)

type IDatachannel interface {
	Groups() []string
	Send(group string, pkt string)

	RegisterHandle(group string,id string, handler func(pkt string))
	DeregisterHandle(group string,id string)

	RegisterConsumer(group string, consumer DatachannelConsumer) 
}

type DatachannelConsumer interface {
	Send(pkt string)
	Recv() (pkt string)
}

type Handler struct {
	handler func(string)
	handle_queue chan string
}

type DatachannelGroup struct {
	send chan string
	recv chan string

	mutext *sync.Mutex
	handlers map[string]*Handler

	consumer DatachannelConsumer
}

type Datachannel struct {
	groups map[string]*DatachannelGroup
}

func NewDatachannel(names []string) IDatachannel {
	dc := &Datachannel{
		groups :map[string]*DatachannelGroup{},
	}

	for _,name := range names {
		dc.groups[name] = &DatachannelGroup{
			send:    make(chan string,10),
			recv:    make(chan string,10),
			mutext: &sync.Mutex{},
			handlers: map[string]*Handler{},
		}

		go func(group *DatachannelGroup) {
			msg := <-group.recv

			group.mutext.Lock()
			for _,handler := range group.handlers{
				if len(handler.handle_queue) <10 {
					handler.handle_queue<-msg
				}
			}
			group.mutext.Unlock()
		}(dc.groups[name])

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
func (dc *Datachannel) Send(group string, pkt string) {
	dc.groups[group].send<-pkt
}
func (dc *Datachannel) RegisterHandle(group string, id string, handler func(pkt string)) {
	if dc.groups[group] == nil {
		fmt.Printf("no group name %s available\n",group)
		return
	}

	dc.groups[group].mutext.Lock()
	defer dc.groups[group].mutext.Unlock()

	dc.groups[group].handlers[id] = &Handler{
		handler : handler,
		handle_queue: make(chan string,10),
	}

	go func() { for { 
		msg:=<-dc.groups[group].handlers[id].handle_queue
		if msg == internal_close { 
			fmt.Println("closed data channel handler")
			return 
		}
		dc.groups[group].handlers[id].handler(msg) 
	}}()
}
func (dc *Datachannel) DeregisterHandle(group string,id string) {
	if dc.groups[group] == nil {
		fmt.Printf("no group name %s available\n",group)
		return
	}

	dc.groups[group].mutext.Lock()
	defer dc.groups[group].mutext.Unlock()

	if dc.groups[group].handlers[id] == nil {
		fmt.Printf("no handler name %s available\n",id)
	}
	fmt.Printf("deregister datachannel %s:%s available\n",group,id)
	dc.groups[group].handlers[id].handle_queue<-internal_close
	delete(dc.groups[group].handlers,id)
}


func (dc *Datachannel) RegisterConsumer(group string, consumer DatachannelConsumer) {
	if dc.groups[group] == nil {
		fmt.Printf("no group name %s available\n",group)
		return
	}

	dc.groups[group].mutext.Lock()
	defer dc.groups[group].mutext.Unlock()

	if dc.groups[group].consumer != nil{
		fmt.Printf("consumer for group %s available\n",group)
		return
	}

	dc.groups[group].consumer = consumer
	go func(_consumer DatachannelConsumer) {
		for {
			msg := _consumer.Recv()
			dc.groups[group].recv<-msg 
		}
	}(consumer)
	go func(_consumer DatachannelConsumer) {
		for {
			msg :=<- dc.groups[group].send
			_consumer.Send(msg)
		}
	}(consumer)
}