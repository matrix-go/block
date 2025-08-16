package network

import (
	"bytes"
	"encoding/gob"
	"errors"
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
	Transport     Transport
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}
type Server struct {
	ServerOpt
	Transport   Transport
	Transports  []Transport // transport wait for connection
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
		opt.Logger = log.With(opt.Logger, "ID", opt.ID, "addr", opt.Transport.Addr())
	}
	chain, err := core.NewBlockchain(genesisBlock(), opt.Logger)
	if err != nil {
		return nil, err
	}
	server := &Server{
		ServerOpt:   opt,
		Transport:   opt.Transport,
		Transports:  opt.Transports,
		isValidator: opt.PrivateKey != nil,
		memPool:     NewTxPool(10),
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
	s.bootstrapNodes()
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
			if err = s.RPCProcessor.ProcessMessage(deMsg); err != nil {
				if !errors.Is(err, core.ErrBlockAlreadyInBlockchain) {
					s.Logger.Log("err", err)
				}
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
			//logrus.Error(err)
		}
	}
}

func (s *Server) ProcessMessage(msg *DecodeMessage) error {

	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		return s.processSendStatusMessage(msg.From, t)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		return s.processSendBlocksMessage(msg.From, t)
	default:
		return fmt.Errorf("unknown msg type: %T", t)
	}
}

func (s *Server) broadcast(payload []byte) error {
	return s.Transport.Broadcast(payload)
}

func (s *Server) processTransaction(tx *core.Transaction) error {

	txHash := tx.Hash(core.NewTransactionHasher())
	if s.memPool.Contains(txHash) {
		//s.Logger.Log("msg", "mempool already has tx", "hash", txHash)
		return nil
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	if err := tx.Verify(); err != nil {
		return err
	}
	//s.Logger.Log("msg", "adding tx to mempool", "hash", txHash, "mempoolPending", s.memPool.PendingCount())
	// TODO: broadcast the tx to peers

	go func() {
		if err := s.broadcastTx(tx); err != nil {
			logrus.WithFields(logrus.Fields{
				"hash": txHash,
			}).Errorf("broadcast tx failed: %v", err)
		}
	}()

	return s.memPool.Add(tx)
}

func (s *Server) initTransport() {
	go func(tr Transport) {
		for msg := range tr.Consume() {
			s.rpcChan <- msg
		}
	}(s.Transport)
}

func (s *Server) bootstrapNodes() {
	for _, transport := range s.Transports {
		if s.Transport.Addr() != transport.Addr() {
			if err := s.Transport.Connect(transport); err != nil {
				s.Logger.Log("msg", "could not connect to transport", "addr", transport.Addr(), "err", err)
				continue
			}
			s.Logger.Log("msg", "connected to transport", "addr", transport.Addr())
		}
	}
	if err := s.sendGetStatusMessage(); err != nil {
		s.Logger.Log("err", err, "msg", "send GetStatusMessage failed")
	}
}

func (s *Server) broadcastBlock(b *core.Block) error {
	var buf bytes.Buffer

	if err := b.Encode(core.NewGobBlockEncoder(&buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeBlock, buf.Bytes())
	return s.broadcast(msg.Bytes())
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
	txs := s.memPool.Pending()
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
	// clear cached pending transactions
	s.memPool.ClearPending()

	// broad cast block
	return s.broadcastBlock(block)
}

func (s *Server) processBlock(data *core.Block) error {
	if err := s.chain.AddBlock(data); err != nil {
		return err
	}
	go s.broadcastBlock(data)
	return nil
}

func (s *Server) processStatusMessage(to NetAddr, data *StatusMessage) error {
	if data.Height <= s.chain.Height() {
		s.Logger.Log("msg", "remote height is less than or equal with local chain height", "height", s.chain.Height())
		return nil
	}
	// remote block height is higher than local block height
	getBlockMsg := NewGetBlocksMessage(s.chain.Height(), data.Height)
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&getBlockMsg); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetBlock, buf.Bytes())
	return s.Transport.SendMessage(to, msg.Bytes())
}

func (s *Server) sendGetStatusMessage() error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(GetStatusMessage{}); err != nil {
		s.Logger.Log("err", err, "msg", "send GetStatusMessage failed with encode message")
		return err
	}
	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	if err := s.Transport.Broadcast(msg.Bytes()); err != nil {
		s.Logger.Log("err", err, "msg", "send GetStatusMessage failed with broadcast message")
	}
	return nil
}

func (s *Server) processSendStatusMessage(to NetAddr, data *GetStatusMessage) error {
	stsMessage := NewStatusMessage()
	stsMessage.ID = s.ID
	stsMessage.Height = s.chain.Height()
	stsMessage.Version = s.chain.Version()
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(stsMessage); err != nil {
		s.Logger.Log("err", err, "msg", "processSendStatusMessage failed with encode message")
		return err
	}
	msg := NewMessage(MessageTypeStatus, buf.Bytes())
	return s.Transport.SendMessage(to, msg.Bytes())
}

func (s *Server) processSendBlocksMessage(to NetAddr, data *GetBlocksMessage) error {
	// TODO: send blocks to xxx
	fmt.Printf("===> process send blocks message to %s, msg %+v\n", to, data)
	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		Height:    0,
		Timestamp: 0,
	}
	var txs []*core.Transaction
	return core.NewBlock(header, txs)
}
