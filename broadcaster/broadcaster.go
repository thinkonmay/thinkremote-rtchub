package broadcaster

type Broadcaster interface {
	Push(buff []byte) 
	Close()
	Open()
}
