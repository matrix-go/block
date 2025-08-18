package main

import (
	"bytes"
	"fmt"
	"github.com/matrix-go/block/core"
	"github.com/matrix-go/block/crypto"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/matrix-go/block/network"
)

var transports = []network.Transport{
	network.NewLocalTransport("LOCAL"),
	network.NewLocalTransport("REMOTE_1"),
	network.NewLocalTransport("REMOTE_2"),
	network.NewLocalTransport("REMOTE_3"),
	network.NewLocalTransport("LATE_REMOTE"),
}

var peers []network.Peer

func main() {

	//servers := initLocalTransportServers()

	servers := initTcpTransportSevers()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	// Block until a signal is received.
	s := <-c
	for _, server := range servers {
		server.Stop()
	}
	fmt.Println("shut down:", s)

}

func initLocalTransportServers() []*network.Server {

	peers = []network.Peer{}
	for _, tr := range transports {
		peers = append(peers, network.NewLocalPeer(tr.Addr(), tr.(*network.LocalTransport).RpcChan))
	}
	//if err := initRemoteSevers(transports[1:]...); err != nil {
	//	panic(err)
	//}

	localTr := transports[0]
	localPeer := peers[0]

	// local validator node
	privateKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	localSrv := makeServer("LOCAL", privateKey, localTr, []network.Peer{}, ":9000")
	go localSrv.Start()

	// remote node send transaction
	remoteTr := transports[1]
	remoteSrv := makeServer("REMOTE_1", nil, remoteTr, []network.Peer{localPeer}, "")
	go remoteSrv.Start()
	go func() {
		time.Sleep(1 * time.Second)
		for {
			if err := sendTransaction(remoteTr, localPeer); err != nil {
				logrus.Error(err)
			}
			time.Sleep(time.Second * 2)
		}
	}()

	// mock late server
	lateTr := transports[len(transports)-1]
	lateSrv := makeServer("LATE_REMOTE", nil, lateTr, []network.Peer{localPeer}, "")
	time.Sleep(time.Second * 6)
	go lateSrv.Start()

	// localPeer connect to late server
	latePeer := peers[len(peers)-1]
	localTr.Connect(latePeer)

	// localPeer connect to remote server
	remotePeer := peers[1]
	localTr.Connect(remotePeer)

	return []*network.Server{localSrv, remoteSrv, lateSrv}
}

func initTcpTransportSevers() []*network.Server {
	localTr := network.NewTcpTransport(":8080")
	localPeer := network.NewTcpPeer(localTr.Addr())

	// local validator node
	privateKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	localApiAddr := ":9000"
	localServer := makeServer("LOCAL", privateKey, localTr, nil, localApiAddr)
	go localServer.Start()
	time.Sleep(1 * time.Second)

	// remote node to send transactions
	remoteTr := network.NewTcpTransport(":8082")
	remoteServer := makeServer("REMOTE", nil, remoteTr, []network.Peer{localPeer}, "")
	go remoteServer.Start()
	go func() {
		time.Sleep(2 * time.Second)
		for {
			if err := sendTransaction(remoteTr, localPeer); err != nil {
				logrus.Error(err)
			}
			time.Sleep(time.Second)
		}
	}()

	// late node to sync block
	//lateTr := network.NewTcpTransport(":8081")
	//lateServer := makeServer("LATE", nil, lateTr,
	//	//[]network.Peer{remotePeer},
	//	[]network.Peer{localPeer},
	//	"",
	//)
	//time.Sleep(time.Second * 10)
	//go lateServer.Start()
	//time.Sleep(1 * time.Second)

	return []*network.Server{localServer, remoteServer}
}
func initRemoteSevers(trs ...network.Transport) error {
	for idx, tr := range trs {
		id := fmt.Sprintf("REMOTE_%d", idx+1)
		s := makeServer(id, nil, tr, peers, "")
		go s.Start()
	}
	return nil
}

func makeServer(id string, privateKey *crypto.PrivateKey, tr network.Transport, peers []network.Peer, apiAddr string) *network.Server {
	opt := network.ServerOpt{
		ID:         id,
		Transport:  tr,
		SeedPeers:  peers,
		BlockTime:  time.Second * 5,
		PrivateKey: privateKey,
		ApiAddr:    apiAddr,
	}

	server, err := network.NewServer(opt)
	if err != nil {
		panic(err)
	}
	return server
}

func sendTransaction(tr network.Transport, to network.Peer) error {
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
