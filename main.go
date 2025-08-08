package main

import (
	"fmt"
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
	go func() {
		for {
			if err := remoteTr.SendMessage(localTr.Addr(), []byte("hello")); err != nil {
				fmt.Printf("got err: %v\n", err)
			}
			time.Sleep(time.Second * 1)
		}
	}()

	server.Start()
}
