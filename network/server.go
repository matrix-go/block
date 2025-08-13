package network

import (
	"bytes"
	"fmt"
	"github.com/go-kit/log"
	"github.com/matrix-go/block/core"
	"github.com/sirupsen/logrus"
	"os"
	"time"

	"github.com/matrix-go/block/crypto"
)

type ServerOpt struct {
	ID            string
	Logger        log.Logger
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
	chain       *core.Blockchain
	blockTime   time.Duration
	rpcChan     chan RPC
	quit        chan struct{}
}

func (s *Server) Quit() {
	s.quit <- struct{}{}
}

func NewServer(opt ServerOpt) (*Server, error) {
	if opt.BlockTime == 0 {
		opt.BlockTime = time.Second // default block time
	}
	if opt.RPCDecodeFunc == nil {
		opt.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if opt.Logger == nil {
		opt.Logger = log.NewLogfmtLogger(os.Stderr)
		opt.Logger = log.With(opt.Logger, "ID", opt.ID)
	}
	chain, err := core.NewBlockchain(genesisBlock(), opt.Logger)
	if err != nil {
		return nil, err
	}
	server := &Server{
		ServerOpt:   opt,
		Transports:  opt.Transports,
		isValidator: opt.PrivateKey != nil,
		memPool:     NewTxPool(),
		chain:       chain,
		blockTime:   opt.BlockTime,
		rpcChan:     make(chan RPC, 1024),
		quit:        make(chan struct{}, 1),
	}

	if opt.RPCProcessor == nil {
		server.RPCProcessor = server
	}
	return server, nil
}

func (s *Server) Start() {
	s.initTransport()
	if s.isValidator {
		go s.validatorLoop()
	}
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
		case <-s.quit:
			break quit
		}
	}

	s.Logger.Log("msg", "server stopped")
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.blockTime)
	s.Logger.Log("msg", "validator loop started", "block", s.blockTime)
	for {
		<-ticker.C
		if err := s.createNewBlock(); err != nil {
			logrus.Error(err)
		}
	}
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
		s.Logger.Log("msg", "mempool already has tx", "hash", txHash)
		return nil
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	if err := tx.Verify(); err != nil {
		return err
	}
	s.Logger.Log("msg", "adding tx to mempool", "hash", txHash, "mempoolLen", s.memPool.Len())
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

	header, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}
	// get txs from mempool
	txs := s.memPool.Transactions()
	block, err := core.NewBlockWithPrevHeader(header, txs)
	if err != nil {
		return err
	}
	if err = block.Sign(s.PrivateKey); err != nil {
		return err
	}
	if err = s.chain.AddBlock(block); err != nil {
		return err
	}
	// TODO: clear cached picked transactions
	s.memPool.Flush()
	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		Height:    0,
		Timestamp: uint64(time.Now().UnixNano()),
	}
	var txs []*core.Transaction
	return core.NewBlock(header, txs)
}
