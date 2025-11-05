//go:build !js

package secretary

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/codeharik/secretary/api/apiconnect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func writeJson(w http.ResponseWriter, data []byte, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

	COMMAND_LOGS = ""
}

func (s *Secretary) getAllTreeHandler(w http.ResponseWriter, r *http.Request) {
	data, err := s.HandleGetAllTree()
	writeJson(w, data, err)
}

func (s *Secretary) getTreeHandler(w http.ResponseWriter, r *http.Request) {
	collectionName := r.PathValue("collectionName")
	data, err := s.HandleGetTree(collectionName)
	writeJson(w, data, err)
}

type NewTreeRequest struct {
	CollectionName      string `json:"CollectionName"`
	Order               uint8  `json:"Order"`
	NumLevel            uint8  `json:"NumLevel"`
	BaseSize            uint32 `json:"BaseSize"`
	Increment           uint8  `json:"Increment"`
	CompactionBatchSize uint32 `json:"compactionBatchSize"`
}

func (s *Secretary) newTreeHandler(w http.ResponseWriter, r *http.Request) {
	var req NewTreeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJson(w, nil, ErrorInvalidJson)
		return
	}

	data, err := s.HandleNewTree(
		req.CollectionName,
		int(req.Order),
		int(req.NumLevel),
		int(req.BaseSize),
		int(req.Increment),
		int(req.CompactionBatchSize))
	writeJson(w, data, err)
}

func (s *Secretary) setRecordHandler(w http.ResponseWriter, r *http.Request) {
	collectionName := r.PathValue("collectionName")

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (len(strings.Trim(req.Value, " ")) == 0) {
		writeJson(w, nil, err)
		return
	}

	data, err := s.HandleSetRecord(collectionName, req.Key, req.Value)
	writeJson(w, data, err)
}

func (s *Secretary) sortedSetRecordHandler(w http.ResponseWriter, r *http.Request) {
	collectionName := r.PathValue("collectionName")
	v := r.PathValue("value")

	value, err := strconv.Atoi(v)

	if err != nil || value < 1 {
		writeJson(w, nil, err)
		return
	}

	data, err := s.HandleSortedSetRecord(collectionName, value)
	writeJson(w, data, err)
}

func (s *Secretary) getRecordHandler(w http.ResponseWriter, r *http.Request) {
	collectionName := r.PathValue("collectionName")
	id := r.PathValue("id")

	data, err := s.HandleGetRecord(collectionName, id)
	writeJson(w, data, err)
}

func (s *Secretary) deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	collectionName := r.PathValue("collectionName")
	id := r.PathValue("id")

	data, err := s.HandleDeleteRecord(collectionName, id)
	writeJson(w, data, err)
}

func (s *Secretary) clearTreeHandler(w http.ResponseWriter, r *http.Request) {
	collectionName := r.PathValue("collectionName")

	data, err := s.HandleClearTree(collectionName)

	writeJson(w, data, err)
}

func (s *Secretary) setupRouter(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("GET /getalltree", s.getAllTreeHandler)
	mux.HandleFunc("GET /gettree/{collectionName}", s.getTreeHandler)
	mux.HandleFunc("POST /newtree", s.newTreeHandler)
	mux.HandleFunc("POST /set/{collectionName}", s.setRecordHandler)
	mux.HandleFunc("POST /sortedset/{collectionName}/{value}", s.sortedSetRecordHandler)
	mux.HandleFunc("GET /get/{collectionName}/{id}", s.getRecordHandler)
	mux.HandleFunc("DELETE /delete/{collectionName}/{id}", s.deleteRecordHandler)
	mux.HandleFunc("DELETE /clear/{collectionName}", s.clearTreeHandler)

	// Enable CORS with custom settings
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "OPTIONS", "POST", "DELETE"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	return handler
}

func (s *Secretary) Serve() {
	if MODE_WASM {
		return
	}

	// Create a TCP listener on a random available port, OS assigns a free port
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle(apiconnect.NewSecretaryHandler(&s))

	handler := s.setupRouter(mux)

	server := &http.Server{
		Addr: listener.Addr().String(), // Eg:"127.0.0.1:54321"
		Handler: h2c.NewHandler(
			handler,
			&http2.Server{},
		),
	}

	s.listener = listener
	s.server = server

	s.wg.Add(1)
	defer s.wg.Done()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Track if the server stops
	serverExited := make(chan struct{})

	go func() {
		log.Printf("Server listening at %s", s.server.Addr)
		if err := s.server.Serve(s.listener); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
		close(serverExited) // Signal that the server has stopped
	}()

	// Wait for signal
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Shutting down...", sig)
	case <-s.quit:
		log.Printf("Received quit signal. Shutting down...")
	case <-serverExited:
		log.Printf("Server exited unexpectedly.")
	}
}

func (s *Secretary) ServerShutdown() {
	s.once.Do(func() { // Ensures this runs only once

		// Close quit channel only if this call initiated shutdown
		select {
		case <-s.quit:
			// Already closed, do nothing
		default:
			close(s.quit)
		}

		// Gracefully shut down the HTTP server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
			if err := s.server.Close(); err != nil {
				log.Fatalf("Server force close error: %v", err)
			}
		}

		if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Printf("Listener close error: %v", err)
		}

		s.wg.Wait() // the program waits for all goroutines to exit

		log.Printf("Server terminated!")
	})
}
