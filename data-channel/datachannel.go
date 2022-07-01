package datachannel



	
type Datachannel interface {
	Open(port int);
	Write(data []byte, err error);
	Close();
}