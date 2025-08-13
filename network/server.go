package network

import (
	"bytes"
	"fmt"
	"github.com/matrix-go/block/core"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/matrix-go/block/crypto"
)

type ServerOpt struct {
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}
type Server struct {
	ServerOpt
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
	if opt.RPCDecodeFunc == nil {
		opt.RPCDecodeFunc = DefaultRPCDecodeFunc
	}

	server := &Server{
		ServerOpt:   opt,
		Transports:  opt.Transports,
		isValidator: opt.PrivateKey != nil,
		memPool:     NewTxPool(),
		blockTime:   opt.BlockTime,
		rpcChan:     make(chan RPC, 1024),
		quit:        make(chan struct{}, 1),
	}

	if opt.RPCProcessor == nil {
		server.RPCProcessor = server
	}
	return server
}

func (s *Server) Start() {
	s.initTransport()
	tk := time.NewTicker(s.blockTime)
quit:
	for {
		select {
		case msg := <-s.rpcChan:
			deMsg, err := s.RPCDecodeFunc(msg)
			if err != nil {
				logrus.Error(err)
			}
			if err := s.RPCProcessor.ProcessMessage(deMsg); err != nil {
				logrus.Error(err)
			}
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

func (s *Server) ProcessMessage(msg *DecodeMessage) error {

	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	default:
		return fmt.Errorf("unknown msg type: %T", t)
	}
}

func (s *Server) broadcast(payload []byte) error {
	for _, transport := range s.Transports {
		// TODO: broadcast
		if err := transport.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {

	txHash := tx.Hash(core.NewTransactionHasher())
	if s.memPool.HasTx(txHash) {
		logrus.WithFields(logrus.Fields{
			"hash": txHash,
		}).Info("mempool already has tx")
		return nil
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	if err := tx.Verify(); err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"hash":        txHash,
		"mempool len": s.memPool.Len(),
	}).Info("adding tx to mempool")

	// TODO: broadcast the tx to peers

	go func() {
		if err := s.broadcastTx(tx); err != nil {
			logrus.WithFields(logrus.Fields{
				"hash": txHash,
			}).Errorf("broadcast tx failed: %v", err)
		}
	}()

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

func (s *Server) broadcastTx(tx *core.Transaction) error {
	var buf bytes.Buffer
	if err := tx.Encode(core.NewTxEncoder(&buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
}

func (s *Server) createNewBlock() error {
	fmt.Println("prepare to create a new block")
	return nil
}
