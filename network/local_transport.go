package network

import (
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr    NetAddr
	lock    sync.RWMutex
	peers   map[NetAddr]*LocalTransport
	rpcChan chan RPC
}

func NewLocalTransport(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		addr:    addr,
		lock:    sync.RWMutex{},
		peers:   make(map[NetAddr]*LocalTransport),
		rpcChan: make(chan RPC, 1),
	}
}

// Addr implements Transport.
func (l *LocalTransport) Addr() NetAddr {
	return l.addr
}

// Connect implements Transport.
func (l *LocalTransport) Connect(tr Transport) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.peers[tr.Addr()] = tr.(*LocalTransport)
	return nil
}

// Consume implements Transport.
func (l *LocalTransport) Consume() <-chan RPC {
	return l.rpcChan
}

// SendMessage implements Transport.
func (l *LocalTransport) SendMessage(to NetAddr, msg []byte) error {
	l.lock.RLock()
	defer l.lock.RUnlock()
	peer, ok := l.peers[to]
	if !ok {
		return fmt.Errorf("peer not exists: %s", to)
	}
	peer.rpcChan <- RPC{
		From:    l.addr,
		Payload: msg,
	}
	return nil
}

var _ Transport = (*LocalTransport)(nil)
