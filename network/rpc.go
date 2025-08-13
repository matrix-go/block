package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/matrix-go/block/core"
	"github.com/sirupsen/logrus"
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
	logrus.WithFields(logrus.Fields{
		"from": rpc.From,
		"type": msg.Header,
	}).Infof("process message from: %v", rpc.From)
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
	MessageTypeTx    MessageType = 0x01
	MessageTypeBlock MessageType = 0x02
)
