package main

import (
	"bytes"
	"fmt"
	"github.com/matrix-go/block/core"
	"github.com/matrix-go/block/crypto"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/matrix-go/block/network"
)

func main() {
	localTr := network.NewLocalTransport("LOCAL")
	remoteTrA := network.NewLocalTransport("REMOTE_1")
	remoteTrB := network.NewLocalTransport("REMOTE_2")
	remoteTrC := network.NewLocalTransport("REMOTE_3")
	_ = localTr.Connect(remoteTrA)
	_ = remoteTrA.Connect(remoteTrB)
	_ = remoteTrB.Connect(remoteTrC)
	_ = remoteTrA.Connect(localTr)

	if err := initRemoteSevers(remoteTrA, remoteTrB, remoteTrC); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := sendTransaction(remoteTrA, localTr.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(time.Second * 2)
		}
	}()

	privateKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	localServer := makeServer("LOCAL", privateKey, localTr)
	localServer.Start()
}

func initRemoteSevers(trs ...network.Transport) error {
	for idx, tr := range trs {
		id := fmt.Sprintf("REMOTE_%d", idx+1)
		s := makeServer(id, nil, tr)
		go s.Start()
	}
	return nil
}

func makeServer(id string, privateKey *crypto.PrivateKey, tr network.Transport) *network.Server {
	opt := network.ServerOpt{
		ID:         id,
		Transports: []network.Transport{tr},
		BlockTime:  time.Second * 5,
		PrivateKey: privateKey,
	}

	server, err := network.NewServer(opt)
	if err != nil {
		panic(err)
	}
	return server
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	tx := core.NewTransaction(contract())
	privateKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %s", err)
	}
	if err = tx.Sign(privateKey); err != nil {
		return fmt.Errorf("failed to sign tx: %s", err)
	}
	var buf bytes.Buffer
	if err := tx.Encode(core.NewTxEncoder(&buf)); err != nil {
		return fmt.Errorf("failed to encode tx: %s", err)
	}
	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())
}

func contract() []byte {
	return []byte{
		0x03, 0x0a, 0x02, 0x0a, 0x0e, // push 3, push 2 and sub
		0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, // push FOO and pack
		0x0f,                         // store [FOO,1]
		0x03, 0x0a, 0x02, 0x0a, 0x0b, // push 3, push 2 and add
		0x46, 0x0c, 0x4f, 0x0c, 0x4d, 0x0c, 0x03, 0x0a, 0x0d, // push FOM and pack
		0x0f,                                                 // store [FOM,1]
		0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, // push FOO and pack
		0x10, // get foo
	}
}
