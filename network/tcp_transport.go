package network

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
)

type TcpPeer struct {
	addr NetAddr
	conn net.Conn
}

func NewTcpPeer(addr NetAddr) *TcpPeer {
	return &TcpPeer{
		addr: addr,
		conn: nil,
	}
}

func (t *TcpPeer) Addr() NetAddr {
	if t.conn != nil {
		return NetAddr(t.conn.RemoteAddr().String())
	}
	return t.addr
}

func (t *TcpPeer) Write(addr NetAddr, msg []byte) error {
	_, err := t.conn.Write(msg)
	return err
}

var _ Peer = (*TcpPeer)(nil)

type TcpTransport struct {
	addr     string
	listener net.Listener
	peerChan chan Peer
	rpcChan  chan RPC
}

func NewTcpTransport(addr string) *TcpTransport {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic(err)
	}
	return &TcpTransport{
		addr:     address.String(),
		rpcChan:  make(chan RPC, 1),
		peerChan: make(chan Peer, 1),
	}
}

func (t *TcpTransport) Start() error {
	listener, err := net.Listen("tcp", t.addr)
	if err != nil {
		return err
	}
	go t.acceptLoop()

	t.listener = listener
	fmt.Printf("tcp transport listening at %v\n", t.addr)
	return nil
}

func (t *TcpTransport) acceptLoop() {

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			// retry count
			if errors.Is(err, net.ErrClosed) {
				return
			}
			fmt.Printf("tcp acceptLoop err: %v\n", err)
			continue
		}
		peer := &TcpPeer{
			conn: conn,
		}
		t.peerChan <- peer
		go t.readLoop(peer)
	}
}

func (t *TcpTransport) readLoop(peer *TcpPeer) {
	for {
		// TODO: tcp protocol
		var buf = make([]byte, 4096)
		n, err := peer.conn.Read(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}
			fmt.Printf("tcp readLoop err: %v\n", err)
			continue
		}
		// TODO: handle message
		t.rpcChan <- RPC{
			From:    NetAddr(peer.conn.RemoteAddr().String()),
			Payload: bytes.NewReader(buf[:n]),
		}
	}
}

func (t *TcpTransport) Stop() error {
	return t.listener.Close()
}

func (t *TcpTransport) Consume() <-chan RPC {
	return t.rpcChan
}

func (t *TcpTransport) ConsumePeer() <-chan Peer {
	return t.peerChan
}

func (t *TcpTransport) Connect(peer Peer) error {
	conn, err := net.Dial("tcp", peer.Addr().String())
	if err != nil {
		return err
	}
	p := peer.(*TcpPeer)

	p.conn = conn
	p.addr = NetAddr(conn.RemoteAddr().String())
	t.peerChan <- p
	fmt.Printf("%s tcp connect to %v\n", t.addr, p.addr)
	go t.readLoop(p)
	return nil
}

func (t *TcpTransport) SendMessage(to Peer, msg []byte) error {
	return to.Write(to.Addr(), msg)
}

func (t *TcpTransport) Addr() NetAddr {
	return NetAddr(t.addr)
}

var _ Transport = (*TcpTransport)(nil)
