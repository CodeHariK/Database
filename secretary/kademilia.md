# Kademilia

* https://kelseyc18.github.io/kademlia_vis//basics/1/

* [IPFS](https://research.protocol.ai/publications/ipfs-content-addressed-versioned-p2p-file-system/benet2014.pdf)
* [Kademilia Paper](https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf)
* [Distributed Hash Tables with Kademlia](https://codethechange.stanford.edu/guides/guide_kademlia.html#supporting-dynamic-leaves-and-joins)

* [P2P Networks](https://www.youtube.com/playlist?list=PLL8woMHwr36F-1h7BE92ynHHOE3zebGpA)
* [Kademlia: A Peer-to-Peer Information System Based on the XOR Metric](https://www.youtube.com/watch?v=NxhZ_c8YX8E&list=PLL8woMHwr36F-1h7BE92ynHHOE3zebGpA&index=9)
* [Kademlia, Explained](https://www.youtube.com/watch?v=1QdKhNpsj8M)
* [Kademlia - a Distributed Hash Table implementation | Paper Dissection and Deep-dive](https://www.youtube.com/watch?v=_kCHOpINA5g&list=PLsdq-3Z1EPT1rNeq2GXpnivaWINnOaCd0&index=7)
* [Playlist](https://www.youtube.com/playlist?list=PLiYqQVdgdw_sSDkdIZzDRQR9xZlsukIxD)

* [IPFS Kademilia](https://docs.ipfs.tech/concepts/dht/#kademlia)
* [Bitswap : Google Docs](https://docs.google.com/presentation/d/1mbFFGIIKNvboHyLn-k26egOSWkt9nXjlNbxpmCEQfqQ/edit#slide=id.p)

* [Implementing Kademlia in Go](https://medium.com/princeton-systems-course/implementing-kademlia-in-go-65ec9e3c1735)

* https://github.com/libp2p/specs/blob/master/kad-dht/README.md
* https://github.com/libp2p/go-libp2p/blob/master/examples/libp2p-host/host.go
* https://github.com/libp2p/go-libp2p-kad-dht
* https://github.com/nictuku/dht
* https://github.com/pdelong/Kademlia
* https://github.com/prettymuchbryce/kademlia

* [Consistent Hashing with Bounded Loads](https://research.google/blog/consistent-hashing-with-bounded-loads/)
* https://github.com/buraksezer/consistent

###

You’re looking for a Kademlia DHT implementation in Golang. I’ll explain the key components and then provide an implementation outline.

🔹 Key Components of Kademlia in Golang
	1.	Node ID & XOR Distance
	•	Each node has a unique 160-bit ID.
	•	The distance between two nodes is calculated as XOR(NodeA, NodeB).
	2.	k-Buckets for Routing Table
	•	Each node maintains k closest nodes per distance range.
	•	Nodes are stored in Least Recently Used (LRU) order.
	3.	Lookup Algorithm (Recursive Search)
	•	A node finds the k closest peers to a target ID.
	•	Queries are performed in parallel to reduce latency.
	4.	Storing and Retrieving Values
	•	Values are stored at nodes closest to the key.
	•	Nodes must periodically refresh data.
	5.	Network Communication
	•	UDP or TCP for efficient message passing.
	•	Standard RPC messages: PING, STORE, FIND_NODE, FIND_VALUE.

📌 Golang Implementation Outline

We’ll implement:
	1.	Node struct – Holds the ID and network info.
	2.	Routing table (k-buckets) – Manages closest nodes.
	3.	Kademlia DHT – Implements storage, lookup, and messaging.

1️⃣ Node Definition

Each node in the network has a 160-bit ID and an IP address.

package main

import (
	"crypto/sha1"
	"encoding/hex"
	"math/big"
	"net"
)

const IDLength = 20 // 160-bit IDs

// Node represents a Kademlia node.
type Node struct {
	ID   [IDLength]byte // 160-bit ID
	IP   string
	Port int
}

// NewNode creates a new node with a hashed ID.
func NewNode(ip string, port int) Node {
	data := ip + ":" + string(port)
	hash := sha1.Sum([]byte(data)) // Hash IP:Port to get Node ID
	return Node{ID: hash, IP: ip, Port: port}
}

// XOR Distance Calculation
func XORDistance(a, b [IDLength]byte) *big.Int {
	aInt := new(big.Int).SetBytes(a[:])
	bInt := new(big.Int).SetBytes(b[:])
	return new(big.Int).Xor(aInt, bInt)
}

2️⃣ Routing Table (k-buckets)

Each node keeps a list of closest nodes, organized by XOR distance.

package main

import (
	"container/list"
	"sync"
)

const BucketSize = 20 // Kademlia typically uses k=20

// Bucket stores up to K closest nodes.
type Bucket struct {
	nodes *list.List // LRU: Least Recently Used
}

// NewBucket creates an empty bucket.
func NewBucket() *Bucket {
	return &Bucket{nodes: list.New()}
}

// RoutingTable holds multiple k-buckets.
type RoutingTable struct {
	selfID   [IDLength]byte
	buckets  []*Bucket
	bucketMu sync.Mutex
}

// NewRoutingTable initializes a routing table.
func NewRoutingTable(selfID [IDLength]byte) *RoutingTable {
	rt := &RoutingTable{selfID: selfID}
	for i := 0; i < IDLength*8; i++ { // 160-bit ID space
		rt.buckets = append(rt.buckets, NewBucket())
	}
	return rt
}

// Update Routing Table when discovering a new node
func (rt *RoutingTable) Update(node Node) {
	rt.bucketMu.Lock()
	defer rt.bucketMu.Unlock()

	distance := XORDistance(rt.selfID, node.ID)
	index := distance.BitLen() // Determine the bucket index

	bucket := rt.buckets[index]
	for e := bucket.nodes.Front(); e != nil; e = e.Next() {
		if e.Value.(Node).ID == node.ID {
			bucket.nodes.MoveToFront(e) // Update LRU
			return
		}
	}

	if bucket.nodes.Len() < BucketSize {
		bucket.nodes.PushFront(node)
	} else {
		// TODO: Implement eviction strategy for full bucket
	}
}

3️⃣ Kademlia DHT Core (Lookup & Store)

Now we implement:
	•	FindClosestNodes()
	•	Store(key, value)
	•	FindValue(key)

package main

import (
	"fmt"
	"sync"
)

type Kademlia struct {
	selfNode Node
	routing  *RoutingTable
	data     map[[IDLength]byte]string
	dataMu   sync.RWMutex
}

// NewKademlia initializes a DHT node.
func NewKademlia(ip string, port int) *Kademlia {
	node := NewNode(ip, port)
	return &Kademlia{
		selfNode: node,
		routing:  NewRoutingTable(node.ID),
		data:     make(map[[IDLength]byte]string),
	}
}

// Store a value in the network.
func (k *Kademlia) Store(key string, value string) {
	hash := sha1.Sum([]byte(key)) // Hash key for storage
	k.dataMu.Lock()
	k.data[hash] = value
	k.dataMu.Unlock()
}

// FindValue searches for a key in the network.
func (k *Kademlia) FindValue(key string) (string, bool) {
	hash := sha1.Sum([]byte(key))
	k.dataMu.RLock()
	defer k.dataMu.RUnlock()
	val, found := k.data[hash]
	return val, found
}

// FindClosestNodes returns the closest nodes to a target ID.
func (k *Kademlia) FindClosestNodes(target [IDLength]byte) []Node {
	var closest []Node
	// TODO: Implement search logic using XOR distance and k-buckets.
	return closest
}

4️⃣ Network Communication (UDP Messaging)

Kademlia uses:
	•	PING (Check if node is alive)
	•	STORE (Store a key-value pair)
	•	FIND_NODE (Lookup closest nodes)
	•	FIND_VALUE (Lookup key-value)

package main

import (
	"encoding/json"
	"fmt"
	"net"
)

type Message struct {
	Type string
	Key  string
	Value string
}

// Handle incoming UDP requests.
func handleRequest(conn *net.UDPConn, k *Kademlia) {
	buf := make([]byte, 1024)
	n, addr, _ := conn.ReadFromUDP(buf)

	var msg Message
	json.Unmarshal(buf[:n], &msg)

	switch msg.Type {
	case "STORE":
		k.Store(msg.Key, msg.Value)
		fmt.Println("Stored:", msg.Key)
	case "FIND_VALUE":
		value, found := k.FindValue(msg.Key)
		response := Message{Type: "RESPONSE", Key: msg.Key, Value: ""}
		if found {
			response.Value = value
		}
		data, _ := json.Marshal(response)
		conn.WriteToUDP(data, addr)
	}
}

// Start UDP server.
func (k *Kademlia) StartServer(port int) {
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	for {
		handleRequest(conn, k)
	}
}

🚀 Running the Kademlia DHT
	1.	Start a node:

k := NewKademlia("127.0.0.1", 8080)
go k.StartServer(8080)


	2.	Store a value:

k.Store("myfile.txt", "some_hash")


	3.	Lookup a value:

value, found := k.FindValue("myfile.txt")
fmt.Println("Found:", found, "Value:", value)

🎯 Next Steps
	•	Implement full routing table lookup.
	•	Add peer-to-peer node discovery.
	•	Optimize message handling and node eviction.

Would you like a complete peer-to-peer test setup or improvements for production use? 🚀
