package main

import (
	"bytes"
	"fmt"
	"github.com/matrix-go/block/core"
	"github.com/matrix-go/block/crypto"
	"math/rand"
	"strconv"
	"time"

	"github.com/matrix-go/block/network"
)

func main() {
	localTr := network.NewLocalTransport("LOCAL")
	remoteTr := network.NewLocalTransport("REMOTE")
	err := localTr.Connect(remoteTr)
	if err != nil {
		panic(err)
	}
	err = remoteTr.Connect(localTr)
	if err != nil {
		panic(err)
	}
	opt := network.ServerOpt{
		Transports: []network.Transport{localTr, remoteTr},
		BlockTime:  time.Second,
	}

	server := network.NewServer(opt)
	// TODO: mock remote rpc call
	go func() {
		for {
			if err := sendTransaction(remoteTr, localTr.Addr()); err != nil {
				fmt.Printf("failed to send transaction: %s", err)
			}
			time.Sleep(time.Second * 1)
		}
	}()

	server.Start()
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	data := []byte(strconv.FormatInt(rand.Int63n(1000), 10))
	tx := core.NewTransaction(data)
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
