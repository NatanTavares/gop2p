package main

import (
	"context"
	"encoding/binary"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = "/example/1.0.0"
const discoveryNamespace = "example"

var logger = log.New(os.Stdout, "", log.LstdFlags)

func main() {
	peerAddr := flag.String("peer-address", "", "peer address")
	flag.Parse()

	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.Security(tls.ID, tls.New),
	)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	logger.Println("Addresses:", h.Addrs())
	logger.Println("ID:", h.ID())

	h.SetStreamHandler(protocolID, func(s network.Stream) {
		logger.Println("Got a new stream!")
		go writeCounter(s)
		go readCounter(s)
	})

	discoveryService := mdns.NewMdnsService(h, discoveryNamespace, &discoveryNotifee{h: h})
	defer discoveryService.Close()
	if err := discoveryService.Start(); err != nil {
		panic(err)
	}

	if *peerAddr != "" {
		peerMA, err := multiaddr.NewMultiaddr(*peerAddr)
		if err != nil {
			panic(err)
		}
		peerAddrInfo, err := peer.AddrInfoFromP2pAddr(peerMA)
		if err != nil {
			panic(err)
		}

		if err := h.Connect(context.Background(), *peerAddrInfo); err != nil {
			panic(err)
		}
		logger.Println("Connected to", peerAddrInfo.String())

		s, err := h.NewStream(context.Background(), peerAddrInfo.ID, protocolID)
		if err != nil {
			panic(err)
		}

		go writeCounter(s)
		go readCounter(s)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGKILL, syscall.SIGINT)
	<-sigCh
}

func writeCounter(s network.Stream) {
	var counter uint64
	for {
		<-time.After(time.Second)
		counter++
		if err := binary.Write(s, binary.BigEndian, counter); err != nil {
			panic(err)
		}
	}
}

func readCounter(s network.Stream) {
	for {
		var counter uint64
		if err := binary.Read(s, binary.BigEndian, &counter); err != nil {
			panic(err)
		}
		logger.Printf("Received %d from %s\n", counter, s.ID())
	}
}

type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(peerInfo peer.AddrInfo) {
	logger.Println("Found peer", peerInfo.String())
}
