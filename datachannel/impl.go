package datachannel

import (
	"fmt"
	"sync"

	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

type Handler struct {
	handler      func(string)
	handle_queue chan string
	stop         chan bool
}

type DatachannelGroup struct {
	send chan string
	recv chan string
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
			send:     make(chan string, queue_size),
			recv:     make(chan string, queue_size),
			stop:     make(chan bool, 2),
			handlers: map[string]*Handler{},
			mutext:   &sync.Mutex{},
		}

		thread.SafeLoop(group.stop, 0, func() {
			select {
			case msg := <-group.recv:
				group.mutext.Lock()
				defer group.mutext.Unlock()
				for _, handler := range group.handlers {
					handler.handle_queue <- msg
				}
			case <-group.stop:
				group.stop <- true
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
			handle_queue: make(chan string, queue_size),
			stop:         make(chan bool, 2),
		}

		thread.SafeLoop(handler.stop, 0, func() {
			select {
			case msg := <-handler.handle_queue:
				handler.handler(msg)
			case <-handler.stop:
				handler.stop <- true
			}
		})

		group.mutext.Lock()
		defer group.mutext.Unlock()
		group.handlers[id] = handler
	}
}

func (dc *Datachannel) DeregisterHandle(group_name string, id string) {
	if group, found := dc.groups[group_name]; !found {
		fmt.Printf("no group name %s available\n", group_name)
	} else if handler, found := group.handlers[id]; !found {
		fmt.Printf("no handler name %s available\n", id)
	} else {
		handler.stop <- true

		group.mutext.Lock()
		defer group.mutext.Unlock()
		delete(group.handlers, id)
	}
}

func (dc *Datachannel) RegisterConsumer(group_name string, consumer DatachannelConsumer) {
	if group, found := dc.groups[group_name]; !found {
		fmt.Printf("no group name %s available\n", group_name)
	} else if group.consumer != nil {
		fmt.Printf("consumer for group %s available\n", group_name)
	} else {
		thread.SafeLoop(group.stop, 0, func() {
			select {
			case data := <-consumer.Recv():
				group.recv <- data
			case <-group.stop:
				group.stop <- true
			}
		})
		thread.SafeLoop(group.stop, 0, func() {
			select {
			case data := <-group.send:
				consumer.Send(data)
			case <-group.stop:
				group.stop <- true
			}
		})

		group.consumer = consumer
	}
}

func (dc *Datachannel) DeregisterConsumer(group_name string) {
	if group, found := dc.groups[group_name]; !found {
		fmt.Printf("no group name %s available\n", group_name)
	} else {
		group.stop <- true
	}
}
