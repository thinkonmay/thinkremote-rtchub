package listener



type Listener interface {
	Open(port int);
	Read() (size int, data []byte, err error);
	Close();
}