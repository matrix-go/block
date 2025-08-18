package network

import (
	"bytes"
	"sync"
)

type LocalPeer struct {
	addr    NetAddr
	rpcChan chan RPC
}

func NewLocalPeer(addr NetAddr, rpcChan chan RPC) *LocalPeer {
	return &LocalPeer{addr: addr, rpcChan: rpcChan}
}

func (l *LocalPeer) Addr() NetAddr {
	return l.addr
}

func (l *LocalPeer) Write(addr NetAddr, msg []byte) error {
	l.rpcChan <- RPC{
		From:    addr,
		Payload: bytes.NewReader(msg),
	}
	return nil
}

var _ Peer = (*LocalPeer)(nil)

type LocalTransport struct {
	addr     NetAddr
	lock     sync.RWMutex
	local    NetAddr
	peers    map[NetAddr]*LocalTransport
	RpcChan  chan RPC
	peerChan chan Peer
}

func NewLocalTransport(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		addr:     addr,
		lock:     sync.RWMutex{},
		local:    addr,
		peers:    make(map[NetAddr]*LocalTransport),
		RpcChan:  make(chan RPC, 1),
		peerChan: make(chan Peer, 1),
	}
}

func (t *LocalTransport) Start() error {
	return nil
}

func (t *LocalTransport) Stop() error {
	return nil
}

func (t *LocalTransport) Connect(peer Peer) error {
	t.peerChan <- peer
	return nil
}

func (t *LocalTransport) SendMessage(to Peer, msg []byte) error {
	return to.Write(t.Addr(), msg)
}

// Addr implements Transport.
func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}

// Consume implements Transport.
func (t *LocalTransport) Consume() <-chan RPC {
	return t.RpcChan
}

func (t *LocalTransport) ConsumePeer() <-chan Peer {
	return t.peerChan
}

var _ Transport = (*LocalTransport)(nil)
