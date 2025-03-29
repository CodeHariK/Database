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
	if !MODE_TEST {
		msg, _ := utils.LogMessage(msgs...)
		COMMAND_LOGS += fmt.Sprintf("<div style='color:%s;background:#000'>%s</div><br>", utils.LightColor().Hex, strings.ReplaceAll(msg, "\n", "<br>"))
	}
}

type JsonResponse struct {
	Data any    `json:"data"`
	Logs string `json:"logs"`
}

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

func makeJson(data any) ([]byte, error) {
	response := JsonResponse{
		Data: data,
		Logs: COMMAND_LOGS,
	}

	return json.Marshal(response)
}

func (s *Secretary) HandleGetAllTree() ([]byte, error) {
	var trees []*BTree
	for _, o := range s.trees {
		trees = append(trees, o)
	}

	return makeJson(trees)
}

func (s *Secretary) getAllTreeHandler(w http.ResponseWriter, r *http.Request) {
	data, err := s.HandleGetAllTree()
	writeJson(w, data, err)
}

func (s *Secretary) HandleGetTree(table string) ([]byte, error) {
	tree, exists := s.trees[table]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	jsonData, err := makeJson(tree.ToJSON())
	if err != nil {
		return nil, err
	}

	errs := tree.TreeVerify()
	return jsonData, errors.Join(errs...)
}

func (s *Secretary) getTreeHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	data, err := s.HandleGetTree(table)
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

func (s *Secretary) HandleNewTree(req NewTreeRequest) ([]byte, error) {
	tree, err := s.NewBTree(
		req.CollectionName,
		req.Order,
		req.NumLevel,
		req.BaseSize,
		req.Increment,
		req.CompactionBatchSize,
	)
	if err != nil {
		return nil, err
	}
	err = tree.SaveHeader()
	if err != nil {
		return nil, err
	}
	return makeJson("New tree created")
}

func (s *Secretary) newTreeHandler(w http.ResponseWriter, r *http.Request) {
	var req NewTreeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJson(w, nil, ErrorInvalidJson)
		return
	}

	data, err := s.HandleNewTree(req)
	writeJson(w, data, err)
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *Secretary) HandleSetRecord(table string, reqKey string, reqValue string) (key []byte, data []byte, err error) {
	tree, exists := s.trees[table]
	if !exists {
		return nil, nil, ErrorTreeNotFound
	}

	key = []byte(reqKey)
	if len(reqKey) == 0 || len(reqKey) != KEY_SIZE {
		key = []byte(utils.GenerateSeqString(&tree.KeySeq, KEY_SIZE, KEY_INCREMENT))
		key, err = tree.SetKV(key, []byte(reqValue))
	} else {
		key, err = tree.SetKV(key, []byte(reqValue))
	}

	if err == ErrorDuplicateKey {
		err := tree.Update(key, []byte(reqValue))
		if err != nil {
			return nil, nil, err
		}
	} else if err != nil {
		return nil, nil, err
	}

	if errs := tree.TreeVerify(); len(errs) != 0 {
		return nil, nil, errors.Join(errs...)
	}

	response := map[string]any{
		"message": "Data set successfully",
		"table":   table,
		"key":     key,
	}

	data, err = makeJson(response)
	return key, data, err
}

func (s *Secretary) setRecordHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (len(strings.Trim(req.Value, " ")) == 0) {
		writeJson(w, nil, err)
		return
	}

	_, data, err := s.HandleSetRecord(table, req.Key, req.Value)
	writeJson(w, data, err)
}

type SortedSetRequest struct {
	Value int `json:"value"`
}

func (s *Secretary) HandleSortedSetRecord(table string, req SortedSetRequest) ([]byte, error) {
	tree, exists := s.trees[table]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	tree.Erase()

	sortedRecords := SampleSortedKeyRecords(req.Value)

	tree.SortedRecordSet(sortedRecords)

	if errs := tree.TreeVerify(); len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	response := map[string]any{
		"message": "Data set successfully",
		"table":   table,
	}

	return makeJson(response)
}

func (s *Secretary) sortedSetRecordHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	var req SortedSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Value < 1 {
		writeJson(w, nil, err)
		return
	}

	data, err := s.HandleSortedSetRecord(table, req)
	writeJson(w, data, err)
}

func (s *Secretary) HandleGetRecord(table string, key string) ([]byte, error) {
	tree, exists := s.trees[table]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	node, index, found := tree.getLeafNode([]byte(key))
	if found {
		response := map[string]any{
			"table":  table,
			"nodeID": node.NodeID,
			"found":  found,
			"record": node.records[index].Value,
		}
		return makeJson(response)
	}

	return nil, ErrorKeyNotFound
}

func (s *Secretary) getRecordHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	id := r.PathValue("id")

	data, err := s.HandleGetRecord(table, id)
	writeJson(w, data, err)
}

func (s *Secretary) HandleDeleteRecord(table string, id string) ([]byte, error) {
	tree, exists := s.trees[table]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	err := tree.Delete([]byte(id))
	if err != nil {
		return nil, err
	}
	if errs := tree.TreeVerify(); len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	response := map[string]any{
		"table":  table,
		"result": "Delete success " + id,
	}

	return makeJson(response)
}

func (s *Secretary) deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	id := r.PathValue("id")

	data, err := s.HandleDeleteRecord(table, id)
	writeJson(w, data, err)
}

func (s *Secretary) HanldeClearTree(table string) ([]byte, error) {
	tree, exists := s.trees[table]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	tree.Erase()

	response := map[string]any{
		"table":  table,
		"result": "Clear table success",
	}

	return makeJson(response)
}

func (s *Secretary) clearTreeHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	data, err := s.HanldeClearTree(table)

	writeJson(w, data, err)
}

func (s *Secretary) setupRouter(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("GET /getalltree", s.getAllTreeHandler)
	mux.HandleFunc("GET /gettree/{table}", s.getTreeHandler)
	mux.HandleFunc("POST /newtree", s.newTreeHandler)
	mux.HandleFunc("POST /set/{table}", s.setRecordHandler)
	mux.HandleFunc("POST /sortedset/{table}", s.sortedSetRecordHandler)
	mux.HandleFunc("GET /get/{table}/{id}", s.getRecordHandler)
	mux.HandleFunc("DELETE /delete/{table}/{id}", s.deleteRecordHandler)
	mux.HandleFunc("DELETE /clear/{table}", s.clearTreeHandler)

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
