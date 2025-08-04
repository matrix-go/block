package network

import "fmt"

type ServerOpt struct {
	Transports []Transport
}
type Server struct {
	Transports []Transport
	rpcChan    chan RPC
	quit       chan struct{}
}

func (s *Server) Quit() {
	s.quit <- struct{}{}
}

func NewServer(opt ServerOpt) *Server {
	return &Server{
		Transports: opt.Transports,
		rpcChan:    make(chan RPC, 1024),
		quit:       make(chan struct{}, 1),
	}
}

func (s *Server) Start() {
	s.initTransport()

quit:
	for {
		select {
		case msg := <-s.rpcChan:
			fmt.Printf("got msg: %+v\n", msg)
		case <-s.quit:
			break quit
		}
	}

	fmt.Printf("shutdown server...\n")
}

func (s *Server) initTransport() {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			for msg := range tr.Consume() {
				s.rpcChan <- msg
			}
		}(tr)
	}
}
