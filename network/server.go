package network

import (
	"fmt"
	"github.com/matrix-go/block/core"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/matrix-go/block/crypto"
)

type ServerOpt struct {
	Transports []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}
type Server struct {
	Transports  []Transport
	isValidator bool
	memPool     *TxPool
	blockTime   time.Duration
	rpcChan     chan RPC
	quit        chan struct{}
}

func (s *Server) Quit() {
	s.quit <- struct{}{}
}

func NewServer(opt ServerOpt) *Server {
	if opt.BlockTime == 0 {
		opt.BlockTime = time.Second // default block time
	}
	return &Server{
		Transports:  opt.Transports,
		isValidator: opt.PrivateKey != nil,
		memPool:     NewTxPool(),
		blockTime:   opt.BlockTime,
		rpcChan:     make(chan RPC, 1024),
		quit:        make(chan struct{}, 1),
	}
}

func (s *Server) Start() {
	s.initTransport()
	tk := time.NewTicker(s.blockTime)
quit:
	for {
		select {
		case msg := <-s.rpcChan:
			fmt.Printf("got msg: %+v\n", msg)
		case <-tk.C:
			if s.isValidator {
				_ = s.createNewBlock()
			}
		case <-s.quit:
			break quit
		}
	}

	fmt.Printf("shutdown server...\n")
}

func (s *Server) handleTransaction(tx *core.Transaction) error {
	if err := tx.Verify(); err != nil {
		return err
	}
	txHash := tx.Hash(core.NewTransactionHasher())
	if s.memPool.HasTx(txHash) {
		logrus.WithFields(logrus.Fields{
			"hash": txHash,
		}).Info("mempool already has tx")
		return nil
	}
	logrus.WithFields(logrus.Fields{
		"hash": txHash,
	}).Info("adding tx to mempool")

	return s.memPool.AddTx(tx)
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

func (s *Server) createNewBlock() error {
	fmt.Println("prepare to create a new block")
	return nil
}
