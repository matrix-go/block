package network

type NetAddr string

type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(to NetAddr, msg []byte) error
	Addr() NetAddr
	Broadcast(msg []byte) error
}
