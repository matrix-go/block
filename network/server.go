package network

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/go-kit/log"
	"github.com/matrix-go/block/api"
	"github.com/matrix-go/block/core"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"

	"github.com/matrix-go/block/crypto"
)

type ServerOpt struct {
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transport     Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
	SeedPeers     []Peer // peers wait for connection to sync block status
	ApiAddr       string
}
type Server struct {
	ServerOpt
	Transport Transport
	peerMap   map[NetAddr]Peer
	lock      sync.RWMutex
	//Transports  []Transport // transport wait for connection
	isValidator bool
	memPool     *TxPool
	chain       *core.Blockchain
	blockTime   time.Duration
	rpcChan     chan RPC
	quit        chan struct{}
	apiServer   *api.Server
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
		ServerOpt: opt,
		peerMap:   make(map[NetAddr]Peer), // already connected peers
		lock:      sync.RWMutex{},         // peerMap lock
		Transport: opt.Transport,
		//Transports:  opt.Transports,
		isValidator: opt.PrivateKey != nil,
		memPool:     NewTxPool(10),
		chain:       chain,
		blockTime:   opt.BlockTime,
		rpcChan:     make(chan RPC, 1024),
		quit:        make(chan struct{}, 1),
	}

	if opt.ApiAddr != "" {
		apiServerConfig := api.ServerConfig{
			Logger: opt.Logger,
			Addr:   opt.ApiAddr,
		}
		server.apiServer = api.NewServer(apiServerConfig, chain)
	}

	if opt.RPCProcessor == nil {
		server.RPCProcessor = server
	}
	return server, nil
}

func (s *Server) Start() {
	if err := s.Transport.Start(); err != nil {
		panic(err)
	}
	if s.isValidator {
		go s.validatorLoop()
	}
	time.Sleep(time.Second)

	s.bootstrapNetwork()

	if s.ApiAddr != "" {
		go s.apiServer.Start()
	}

quit:
	for {
		select {
		case peer := <-s.Transport.ConsumePeer():
			s.lock.RLock()
			_, exists := s.peerMap[peer.Addr()]
			s.lock.RUnlock()
			if exists {
				s.Logger.Log("peer", peer.Addr(), "msg", "peer exists")
				continue
			}
			s.lock.Lock()
			s.peerMap[peer.Addr()] = peer
			s.lock.Unlock()
		case msg := <-s.Transport.Consume():
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
			logrus.Error(err)
		}
	}
}

func (s *Server) bootstrapNetwork() {
	for _, peer := range s.SeedPeers {
		fmt.Println("trying to connect to ", peer.Addr())
		go func(peer Peer) {
			if err := s.Transport.Connect(peer); err != nil {
				fmt.Printf("could not connect to %+v\n", peer.Addr())
				return
			}
			time.Sleep(1 * time.Second)
			if err := s.processSendGetStatusMessage(peer); err != nil {
				s.Logger.Log("err", err)
			}
		}(peer)
	}
}

// TODO: find a situation to stop
func (s *Server) requestBlockLoop(peer Peer) error {
	ticker := time.NewTicker(time.Second * 3)
	for {
		getBlockMsg := NewGetBlocksMessage(s.chain.Height()+1, 0)
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(&getBlockMsg); err != nil {
			s.Logger.Log("err", err, "msg", "requestBlockLoop encode err", "peer", peer.Addr().String())
		} else {
			msg := NewMessage(MessageTypeGetBlock, buf.Bytes())
			if err = s.Transport.SendMessage(peer, msg.Bytes()); err != nil {
				s.Logger.Log("err", err, "msg", "requestBlockLoop send message err", "peer", peer.Addr().String())
			}
		}
		<-ticker.C
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
		return s.processSendGetBlocksMessage(msg.From, t)
	case *BlockMessage:
		return s.processSyncBlocks(t)
	default:
		return fmt.Errorf("unknown msg type: %T", t)
	}
}

func (s *Server) broadcast(payload []byte) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for addr, peer := range s.peerMap {
		if err := s.Transport.SendMessage(peer, payload); err != nil {
			s.Logger.Log("err", err, "msg", "broadcast to peer", "addr", addr)
			continue
		}
	}
	return nil
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

func (s *Server) processTransaction(tx *core.Transaction) error {

	txHash := tx.GetHash(core.NewTransactionHasher())
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
	s.lock.RLock()
	peer, ok := s.peerMap[to]
	s.lock.RUnlock()
	if !ok {
		return fmt.Errorf("peer not found to %s", to)
	}
	go s.requestBlockLoop(peer)
	return nil
}

func (s *Server) processSendGetStatusMessage(peer Peer) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(GetStatusMessage{}); err != nil {
		s.Logger.Log("err", err, "msg", "send GetStatusMessage failed with encode message")
		return err
	}
	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	if err := s.Transport.SendMessage(peer, msg.Bytes()); err != nil {
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
	s.lock.RLock()
	peer, ok := s.peerMap[to]
	s.lock.RUnlock()
	if !ok {
		return fmt.Errorf("peer not found to %s", to)
	}
	return s.Transport.SendMessage(peer, msg.Bytes())
}

func (s *Server) processSendGetBlocksMessage(to NetAddr, data *GetBlocksMessage) error {
	heightStart := data.From
	heightEnd := data.To
	// return all
	if data.To == 0 {
		heightEnd = s.chain.Height()
	}
	blocks := make([]*core.Block, 0, heightEnd-heightStart+1)
	for i := heightStart; i <= heightEnd; i++ {
		block, err := s.chain.GetBlock(i)
		if err != nil {
			return err
		}
		blocks = append(blocks, block)
		fmt.Printf("height: %d, block: %v\n", i, block)
	}
	fmt.Printf("===> process send blocks message to %s, msg %+v\n", to, data)
	blkMsg := NewBlockMessage(blocks)
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(blkMsg); err != nil {
		s.Logger.Log("err", err, "msg", "processSendGetBlocksMessage failed with encode message")
		return err
	}
	msg := NewMessage(MessageTypeBlocks, buf.Bytes())
	s.lock.RLock()
	peer, ok := s.peerMap[to]
	s.lock.RUnlock()
	if !ok {
		return fmt.Errorf("peer not found to %s", to)
	}
	return s.Transport.SendMessage(peer, msg.Bytes())
}

func (s *Server) processSyncBlocks(t *BlockMessage) error {
	fmt.Printf("process sync blocks: %+v\n", t.Data)
	for _, block := range t.Data {
		if err := s.chain.AddBlock(block); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Stop() {
	//s.apiServer.Stop()
	//s.Transport.Stop()
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
