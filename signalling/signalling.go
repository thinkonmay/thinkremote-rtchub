package signalling



type Signalling interface {
	SendSDP()
	SendICE()
	OnICE()
	OnSDP()
}