package network

type NetAddr string

func (a NetAddr) String() string {
	return string(a)
}

type Peer interface {
	Addr() NetAddr
	Write(addr NetAddr, msg []byte) error
}

type Transport interface {
	Start() error
	Stop() error
	Consume() <-chan RPC
	ConsumePeer() <-chan Peer
	Connect(peer Peer) error // peer should be pointer value
	SendMessage(to Peer, msg []byte) error
	Addr() NetAddr
}
