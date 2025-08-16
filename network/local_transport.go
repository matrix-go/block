package network

import (
	"bytes"
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr    NetAddr
	lock    sync.RWMutex
	local   NetAddr
	peers   map[NetAddr]*LocalTransport
	rpcChan chan RPC
}

func NewLocalTransport(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		addr:    addr,
		lock:    sync.RWMutex{},
		local:   addr,
		peers:   make(map[NetAddr]*LocalTransport),
		rpcChan: make(chan RPC, 1),
	}
}

// Addr implements Transport.
func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}

// Connect implements Transport.
func (t *LocalTransport) Connect(tr Transport) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.peers[tr.Addr()] = tr.(*LocalTransport)
	return nil
}

// Consume implements Transport.
func (t *LocalTransport) Consume() <-chan RPC {
	return t.rpcChan
}

// SendMessage implements Transport.
func (t *LocalTransport) SendMessage(to NetAddr, msg []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.Addr() == to {
		return nil
	}
	peer, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("peer not exists: %s", to)
	}
	peer.rpcChan <- RPC{
		From:    t.addr,
		Payload: bytes.NewReader(msg),
	}
	return nil
}

func (t *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range t.peers {
		// TODO: err handle
		if err := t.SendMessage(peer.Addr(), payload); err != nil {
			return err
		}
	}
	return nil
}

var _ Transport = (*LocalTransport)(nil)
