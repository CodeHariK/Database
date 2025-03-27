package secretary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/codeharik/secretary/api/apiconnect"
	"github.com/codeharik/secretary/utils"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var COMMAND_LOGS = ""

func ServerLog(msgs ...any) {
	msg, _ := utils.LogMessage(msgs...)
	// fmt.Println(msg)
	COMMAND_LOGS += fmt.Sprintf("<div style='color:%s;background:#000'>%s</div><br>", utils.LightColor().Hex, strings.ReplaceAll(msg, "\n", "<br>"))
	if len(COMMAND_LOGS) > 10000 {
		COMMAND_LOGS = ""
	}
}

type JsonResponse struct {
	Data any    `json:"data"`
	Logs string `json:"logs"`
}

func writeJson(w http.ResponseWriter, code int, data any) {
	response := JsonResponse{
		Data: data,
		Logs: COMMAND_LOGS,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	w.Write(jsonData)

	COMMAND_LOGS = ""
}

func (s *Secretary) getAllTreeHandler(w http.ResponseWriter, r *http.Request) {
	var trees []*BTree
	for _, o := range s.trees {
		trees = append(trees, o)
	}
	writeJson(w, http.StatusOK, trees)
}

func (s *Secretary) getTreeHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	tree, exists := s.trees[table]
	if !exists {
		writeJson(w, http.StatusNotFound, "Tree not found")
		return
	}

	if errs := tree.TreeVerify(); errs != nil {
		writeJson(w, http.StatusConflict, tree.ToJSON())
		return
	}
	writeJson(w, http.StatusOK, tree.ToJSON())
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
		writeJson(w, http.StatusBadRequest, "Invalid Json")
		return
	}

	tree, err := s.NewBTree(
		req.CollectionName,
		req.Order,
		req.NumLevel,
		req.BaseSize,
		req.Increment,
		req.CompactionBatchSize,
	)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = tree.SaveHeader()
	if err != nil {
		writeJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddTree(tree)

	writeJson(w, http.StatusOK, "New tree created")
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var keySeq uint64 = 0

func (s *Secretary) setRecordHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (len(strings.Trim(req.Value, " ")) == 0) {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	tree, exists := s.trees[table]
	if !exists {
		writeJson(w, http.StatusNotFound, "Tree not found")
		return
	}

	var key []byte = []byte(req.Key)
	if len(req.Key) == 0 || len(req.Key) != KEY_SIZE {
		key = []byte(utils.GenerateSeqRandomString(&keySeq, KEY_SIZE, 5, 4, req.Value))
	}
	err := tree.Set(key, []byte(req.Value))
	if err == ErrorDuplicateKey {
		err := tree.Update(key, []byte(req.Value))
		if err != nil {
			writeJson(w, http.StatusNotFound, err.Error())
			return
		}
	} else if err != nil {
		writeJson(w, http.StatusNotFound, err.Error())
		return
	}

	if errs := tree.TreeVerify(); errs != nil {
		writeJson(w, http.StatusConflict, utils.ArrayToStrings(errs))
		return
	}

	response := map[string]any{
		"message": "Data set successfully",
		"table":   table,
	}

	writeJson(w, http.StatusOK, response)
}

func (s *Secretary) getRecordHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	id := r.PathValue("id")

	tree, exists := s.trees[table]
	if !exists {
		writeJson(w, http.StatusNotFound, "Tree not found")
		return
	}

	node, index, found := tree.getLeafNode([]byte(id))
	var record string
	if found {
		record = string(node.records[index].Value)
	} else {
		writeJson(w, http.StatusNoContent, "Key not found")
		return
	}

	response := map[string]any{
		"table":  table,
		"nodeID": node.NodeID,
		"found":  found,
		"record": record,
	}

	writeJson(w, http.StatusOK, response)
}

func (s *Secretary) deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	id := r.PathValue("id")

	tree, exists := s.trees[table]
	if !exists {
		writeJson(w, http.StatusNotFound, "Tree not found")
		return
	}

	err := tree.Delete([]byte(id))
	if err != nil {
		writeJson(w, http.StatusInternalServerError, err.Error())
		return
	}
	if errs := tree.TreeVerify(); errs != nil {
		writeJson(w, http.StatusConflict, utils.ArrayToStrings(errs))
		return
	}

	response := map[string]any{
		"table":  table,
		"result": "Delete success " + id,
	}

	writeJson(w, http.StatusOK, response)
}

func (s *Secretary) setupRouter(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("GET /getalltree", s.getAllTreeHandler)
	mux.HandleFunc("GET /gettree/{table}", s.getTreeHandler)
	mux.HandleFunc("POST /newtree", s.newTreeHandler)
	mux.HandleFunc("POST /set/{table}", s.setRecordHandler)
	mux.HandleFunc("GET /get/{table}/{id}", s.getRecordHandler)
	mux.HandleFunc("DELETE /delete/{table}/{id}", s.deleteRecordHandler)

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

func (s *Secretary) Shutdown() {
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
