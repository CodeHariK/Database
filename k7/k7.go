package k7

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	"github.com/gorilla/websocket"
)

// Bucket stores per-second stats
type Bucket struct {
	Requests     int
	Success      int
	Errors       int
	TotalLatency time.Duration
	CPUUsage     float64
	MemUsage     float64
	NetSent      uint64
	NetRecv      uint64
}

// BenchmarkConfig defines the test parameters
type BenchmarkConfig struct {
	Concurrency int
	Duration    time.Duration
	AttackFunc  func() bool
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Store active WebSocket connections
var (
	clients      = make(map[*websocket.Conn]bool)
	clientsMutex sync.Mutex
)

func (config BenchmarkConfig) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Register client
	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	RunBenchmark(config)

	// Keep connection open
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			clientsMutex.Lock()
			delete(clients, conn) // Remove closed connection
			clientsMutex.Unlock()
			break
		}
	}
}

// Function to broadcast benchmark results to all WebSocket clients
func broadcastResults(data Bucket) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	jsonData, _ := json.Marshal(data)

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			fmt.Println("WebSocket send error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// RunBenchmark executes the test
func RunBenchmark(config BenchmarkConfig) []Bucket {
	var wg sync.WaitGroup

	startTime := time.Now()

	// Buckets to store per-100Millisecond stats
	buckets := make([]Bucket, int(config.Duration.Milliseconds()/100))
	var bucketMutex sync.Mutex

	fmt.Println("Starting benchmark...")

	// Start request workers
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			endTime := time.Now().Add(config.Duration)

			bucketFill := 0

			for time.Now().Before(endTime) {
				start := time.Now()

				success := config.AttackFunc()
				duration := time.Since(start)

				// Store results
				timeElapsed := int(time.Since(startTime).Milliseconds()) / 100

				if timeElapsed < len(buckets) {
					bucketMutex.Lock()
					if success {
						buckets[timeElapsed].Success++
					} else {
						buckets[timeElapsed].Errors++
					}
					buckets[timeElapsed].Requests++

					buckets[timeElapsed].TotalLatency += duration

					if bucketFill == timeElapsed {

						// Get system metrics
						cpuPercent, _ := cpu.Percent(0, false)
						memStats, _ := mem.VirtualMemory()
						netStats, _ := net.IOCounters(false)

						// Store in bucket
						buckets[timeElapsed].CPUUsage = cpuPercent[0]
						buckets[timeElapsed].MemUsage = memStats.UsedPercent
						buckets[timeElapsed].NetSent = netStats[0].BytesSent
						buckets[timeElapsed].NetRecv = netStats[0].BytesRecv

						broadcastResults(buckets[timeElapsed])

						fmt.Println(timeElapsed, buckets[timeElapsed].Requests)

						bucketFill++
					}

					bucketMutex.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	return buckets
}

func (config BenchmarkConfig) Attack() {
	server := http.ServeMux{}

	server.HandleFunc("/", httpServeHTML)
	server.HandleFunc("/results", config.serveResults)

	// server.HandleFunc("/", wsServeHTML)
	// server.HandleFunc("/ws", config.serveWebSocket)

	// Start the server
	port := 8888
	fmt.Printf("Server running on http://localhost:%d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), &server)
}
