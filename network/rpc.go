package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/matrix-go/block/core"
	"io"
)

type RPC struct {
	From    NetAddr
	Payload io.Reader // Message
}

//type RPCHandler interface {
//	HandleRPC(RPC) error
//}

type DecodeMessage struct {
	From NetAddr
	Data any
}
type RPCDecodeFunc func(RPC) (*DecodeMessage, error)

func DefaultRPCDecodeFunc(rpc RPC) (*DecodeMessage, error) {
	var msg Message
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode RPC payload %s: %s", rpc.From, err)
	}
	switch msg.Header {
	case MessageTypeTx:
		var tx = &core.Transaction{}
		if err := tx.Decode(core.NewTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, fmt.Errorf("failed to decode tx: %s", err)
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: tx,
		}, nil
		//return h.processor.processTransaction(rpc.From, tx)
	case MessageTypeBlock:
		var block = &core.Block{}
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, fmt.Errorf("failed to decode block: %s", err)
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: block,
		}, nil
	case MessageTypeGetStatus:
		return &DecodeMessage{From: rpc.From, Data: NewGetStatusMessage()}, nil
	case MessageTypeStatus:
		// get status of current block
		stsMsg := NewStatusMessage()
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(stsMsg); err != nil {
			return nil, fmt.Errorf("failed to decode status: %s", err)
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: stsMsg,
		}, nil
	case MessageTypeGetBlock:
		// get all blocks
		gblMsg := NewGetBlocksMessage(0, 0)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(gblMsg); err != nil {
			return nil, fmt.Errorf("failed to decode block: %s", err)
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: gblMsg,
		}, nil
	default:
		return nil, fmt.Errorf("uinknown message type %v", msg.Header)
	}
	return nil, nil
}

type RPCProcessor interface {
	ProcessMessage(msg *DecodeMessage) error
}

type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(header MessageType, data []byte) *Message {
	return &Message{
		Header: header,
		Data:   data,
	}
}

func (m *Message) Bytes() []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(m); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type MessageType byte

const (
	MessageTypeTx        MessageType = 0x01
	MessageTypeBlock     MessageType = 0x02
	MessageTypeGetBlock  MessageType = 0x03
	MessageTypeStatus    MessageType = 0x04
	MessageTypeGetStatus MessageType = 0x05
)
