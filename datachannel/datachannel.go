package datachannel

const (
	queue_size = 32
)

type IDatachannel interface {
	Groups() []string
	Send(group string, pkt string)

	RegisterHandle(group string, id string, handler func(msg string))
	DeregisterHandle(group string, id string)

	RegisterConsumer(group string, consumer DatachannelConsumer)
	DeregisterConsumer(group string)
}

type DatachannelConsumer interface {
	Send(string)
	Recv() chan interface{}
}
