package datachannel

import (
	"fmt"
	"sync"

	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

type Handler struct {
	handler      func(string)
	handle_queue chan interface{}
	stop         chan bool
}

type DatachannelGroup struct {
	send chan interface{}
	recv chan interface{}
	stop chan bool

	mutext   *sync.Mutex
	handlers map[string]*Handler

	consumer DatachannelConsumer
}

type Datachannel struct {
	groups map[string]*DatachannelGroup
}

func NewDatachannel(names ...string) IDatachannel {
	dc := &Datachannel{
		groups: map[string]*DatachannelGroup{},
	}

	for _, name := range names {
		group := &DatachannelGroup{
			send:     make(chan interface{}, queue_size),
			recv:     make(chan interface{}, queue_size),
			stop:     make(chan bool, 2),
			handlers: map[string]*Handler{},
			mutext:   &sync.Mutex{},
		}

		thread.SafeSelect(group.stop, group.recv, func(_msg interface{}) {
			msg := _msg.(string)
			group.mutext.Lock()
			defer group.mutext.Unlock()
			for _, handler := range group.handlers {
				handler.handle_queue <- msg
			}
		})

		dc.groups[name] = group
	}

	return dc
}
func (dc *Datachannel) Groups() []string {
	keys := make([]string, 0, len(dc.groups))
	for k := range dc.groups {
		keys = append(keys, k)
	}

	return keys
}

func (dc *Datachannel) Send(group string, pkt string) {
	if dc.groups[group] == nil {
		return
	} else if len(dc.groups[group].send) == queue_size {
		return
	}

	dc.groups[group].send <- pkt
}

func (dc *Datachannel) RegisterHandle(group_name string,
	id string,
	fun func(msg string)) {

	if group, found := dc.groups[group_name]; !found {
		fmt.Printf("no group name %s available\n", group_name)
	} else {
		handler := &Handler{
			handler:      fun,
			handle_queue: make(chan interface{}, queue_size),
			stop:         make(chan bool, 2),
		}

		thread.SafeSelect(handler.stop, handler.handle_queue, func(_msg interface{}) {
			handler.handler(_msg.(string))
		})

		group.mutext.Lock()
		defer group.mutext.Unlock()
		group.handlers[id] = handler
	}
}

func (dc *Datachannel) DeregisterHandle(group_name string, id string) {
	group, found := dc.groups[group_name]
	if !found {
		fmt.Printf("no group name %s available\n", group_name)
		return
	}

	group.mutext.Lock()
	defer group.mutext.Unlock()
	if handler, found := group.handlers[id]; !found {
		fmt.Printf("no handler name %s available\n", id)
	} else {
		thread.TriggerStop(handler.stop)
		delete(group.handlers, id)
	}
}

func (dc *Datachannel) RegisterConsumer(group_name string, consumer DatachannelConsumer) {
	if group, found := dc.groups[group_name]; !found {
		fmt.Printf("no group name %s available\n", group_name)
	} else if group.consumer != nil {
		fmt.Printf("consumer for group %s available\n", group_name)
	} else {
		thread.SafeSelect(group.stop, consumer.Recv(), func(data interface{}) {
			group.recv <- data
		})
		thread.SafeSelect(group.stop, group.send, func(data interface{}) {
			consumer.Send(data.(string))
		})

		group.consumer = consumer
	}
}

func (dc *Datachannel) DeregisterConsumer(group_name string) {
	if group, found := dc.groups[group_name]; !found {
		fmt.Printf("no group name %s available\n", group_name)
	} else {
		thread.TriggerStop(group.stop)
	}
}
