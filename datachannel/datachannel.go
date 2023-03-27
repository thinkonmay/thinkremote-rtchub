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


type DatachannelGroup struct {
	send chan string
	recv chan string

	mutext *sync.Mutex
	handlers map[string]*struct {
		handler func(string)
		send chan string
	}


	consumer DatachannelConsumer
}

type Datachannel struct {
	groups map[string]*DatachannelGroup
}

func NewDatachannel(names []string) IDatachannel {
	dc := &Datachannel{}

	for _,name := range names {
		dc.groups[name] = &DatachannelGroup{
			send:    make(chan string,10),
			recv:    make(chan string,10),
			mutext: &sync.Mutex{},
			handlers: map[string]*struct{handler func(string); send chan string}{},
			consumer: nil,
		}

		go func(group *DatachannelGroup) {
			msg := <-group.recv

			group.mutext.Lock()
			for _,handler := range group.handlers{
				if len(handler.send) <10 {
					handler.send<-msg
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

	dc.groups[group].handlers[id] = &struct{handler func(string); send chan string}{
		handler : handler,
		send: make(chan string,10),
	}

	go func(sender chan string, handler func(string)) { for { 
		msg:=<-sender
		if msg == internal_close { return }
		handler(msg) 
	}}(dc.groups[group].handlers[id].send,dc.groups[group].handlers[id].handler)
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
	dc.groups[group].send<-internal_close
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