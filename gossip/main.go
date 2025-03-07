package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/codeharik/gossip/api"
	"github.com/codeharik/gossip/api/apiconnect"
)

type GossipNetwork struct {
	node *GossipNode

	MIN_PEER int
	MAX_PEER int
	MAX_HOP  int

	BootstrapAddr   string
	BootstrapClient apiconnect.GossipServiceClient
}

func NewGossipNetwork(node *GossipNode) *GossipNetwork {
	BootstrapAddr := "localhost:8888"
	BootstrapClient := createGossipClient(BootstrapAddr)

	node.Peers[BootstrapClient] = GossipPeer{
		Addr:      BootstrapAddr,
		Bootstrap: true,
		StartTime: time.Now().Unix(),
	}

	return &GossipNetwork{
		node: node,

		MIN_PEER:        5,
		MAX_PEER:        20,
		MAX_HOP:         5,
		BootstrapAddr:   BootstrapAddr,
		BootstrapClient: BootstrapClient,
	}
}

func createGossipClient(addr string) apiconnect.GossipServiceClient {
	return apiconnect.NewGossipServiceClient(http.DefaultClient, "http://"+addr)
}

type GossipPeer struct {
	Addr      string
	Bootstrap bool
	StartTime int64
}

// GossipNode represents a node in the gossip network
type GossipNode struct {
	ID        string
	Addr      string
	Peers     map[apiconnect.GossipServiceClient]GossipPeer
	PeersLock sync.Mutex
}

func NewGossipNode(port int) *GossipNode {
	return &GossipNode{
		ID:    "$GossipNode",
		Addr:  fmt.Sprintf("localhost:%d", port),
		Peers: make(map[apiconnect.GossipServiceClient]GossipPeer),
	}
}

func (network *GossipNetwork) SendMessage(
	ctx context.Context,
	req *connect.Request[api.SendMessageRequest],
) (*connect.Response[api.SendMessageResponse], error) {
	// Log message
	fmt.Printf("[%s] Received message from %s (Hops Left: %d): %s\n",
		network.node.ID, req.Msg.SenderId, req.Msg.HopCount, req.Msg.Message)

	// Stop propagating if hop count is 0
	if req.Msg.HopCount == 0 {
		fmt.Println("Dropping message due to hop count limit.")
		return connect.NewResponse(&api.SendMessageResponse{Received: true}), nil
	}

	// Forward the message with decremented hop count

	msg := req.Msg

	network.node.PeersLock.Lock()
	defer network.node.PeersLock.Unlock()

	if len(network.node.Peers) == 0 || msg.HopCount <= 0 {
		return connect.NewResponse(&api.SendMessageResponse{Received: true}), nil
	}

	// Send new message with decremented hop count
	newMsg := &api.SendMessageRequest{
		SenderId:  network.node.ID,
		HopCount:  msg.HopCount - 1,
		Message:   msg.Message,
		Timestamp: msg.Timestamp,
	}

	{
		type peerEntry struct {
			connection apiconnect.GossipServiceClient
			peer       GossipPeer
		}
		peerList := make([]peerEntry, 0, len(network.node.Peers))

		for connection, peer := range network.node.Peers {
			peerList = append(peerList, peerEntry{connection, peer})
		}

		n := network.MAX_HOP
		if n > len(peerList) {
			n = len(peerList)
		}

		// Shuffle the slice
		rand.Shuffle(len(peerList), func(i, j int) {
			peerList[i], peerList[j] = peerList[j], peerList[i]
		})

		p := peerList[:n]

		// connection, peer := range node.Peers
		for _, peerentry := range p {
			_, err := peerentry.connection.SendMessage(context.Background(), connect.NewRequest(newMsg))
			if err != nil {
				fmt.Printf("[%s] Failed to forward gossip to %s: %v\n", network.node.ID, peerentry.peer.Addr, err)
			}
		}
	}
	return connect.NewResponse(&api.SendMessageResponse{Received: true}), nil
}

func (network *GossipNetwork) Connect(
	ctx context.Context,
	req *connect.Request[api.ConnectRequest],
) (*connect.Response[api.ConnectResponse], error) {
	network.node.PeersLock.Lock()
	defer network.node.PeersLock.Unlock()

	peerList := make([]*api.Peer, 0, len(network.node.Peers))
	for _, peer := range network.node.Peers {
		peerList = append(peerList, &api.Peer{
			Addr:      peer.Addr,
			StartTime: peer.StartTime,
		})
	}

	network.node.Peers[createGossipClient(req.Msg.Peer.Addr)] = GossipPeer{
		Addr:      req.Msg.Peer.Addr,
		Bootstrap: false,
		StartTime: req.Msg.Peer.StartTime,
	}

	fmt.Println("-> ", req.Msg.Peer, network.node.Peers)

	return connect.NewResponse(&api.ConnectResponse{KnownPeers: peerList}), nil
}

func (network *GossipNetwork) StartGossipLoop(node *GossipNode) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		node.PeersLock.Lock()
		node.PeersLock.Unlock()

		var newNodes []*api.Peer

		for connection, peer := range node.Peers {
			go func() {
				msg, err := connection.Connect(context.Background(), connect.NewRequest(&api.ConnectRequest{
					Peer: &api.Peer{
						Addr: network.node.Addr,
					},
				}))
				if err != nil {
					if !node.Peers[connection].Bootstrap {
						fmt.Printf("[%s] Peer %s is unresponsive, removing...\n", node.ID, peer.Addr)
						delete(node.Peers, connection) // Remove dead peer
					} else {
						fmt.Println("Bootstrap unresponsive", node.Peers[connection].Addr)
					}
				} else {
					for _, peer := range msg.Msg.KnownPeers {
						newNodes = append(newNodes, peer)
					}
				}
			}()
		}

		if len(node.Peers) < network.MIN_PEER {
			for _, peer := range newNodes {
				node.Peers[createGossipClient(peer.Addr)] = GossipPeer{
					Addr:      peer.Addr,
					StartTime: peer.StartTime,
				}
			}
		}

		if len(node.Peers) > network.MAX_PEER {

			// Convert map to slice (store both key and peer)
			type peerEntry struct {
				key  apiconnect.GossipServiceClient
				peer GossipPeer
			}
			peerList := make([]peerEntry, 0, len(node.Peers))

			for connection, peer := range node.Peers {
				peerList = append(peerList, peerEntry{connection, peer})
			}

			// Sort by StartTime (oldest first)
			sort.Slice(peerList, func(i, j int) bool {
				return peerList[i].peer.StartTime < peerList[j].peer.StartTime
			})

			// Keep only the oldest MAX_PEER nodes, remove newer ones
			if len(peerList) > network.MAX_PEER {
				for i := network.MAX_PEER; i < len(peerList); i++ {
					if !peerList[i].peer.Bootstrap {
						delete(node.Peers, peerList[i].key)
					}
				}
			}
		}
	}
}

func main() {
	// Define a command-line flag for the port
	port := flag.Int("port", 8080, "Port to run the gossip node on")
	flag.Parse() // Parse command-line arguments

	node := NewGossipNode(*port)
	network := NewGossipNetwork(node)

	// Start gossip loop in the background
	go network.StartGossipLoop(node)

	// HTTP server setup
	mux := http.NewServeMux()
	mux.Handle(apiconnect.NewGossipServiceHandler(network))

	server := &http.Server{
		Addr:    node.Addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Node running on ", *port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for signal
	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...", sig)

	// Graceful shutdown
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Node stopped cleanly.")
}
