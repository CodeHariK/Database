package secretary

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/codeharik/secretary/utils"
	"github.com/rs/cors"
)

func (s *Secretary) getAllBTreeHandler(w http.ResponseWriter, r *http.Request) {
	var hello []*BTree

	for _, o := range s.trees {
		hello = append(hello, o)
	}

	jsonData, err := json.MarshalIndent(hello, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (s *Secretary) getBTreeHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	tree, exists := s.trees[table]
	if !exists {
		http.Error(w, "Tree not found", http.StatusInternalServerError)
		return
	}

	m, err := tree.ConvertBTreeToJSON()
	if err != nil || m == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(m)
}

type NewTreeRequest struct {
	CollectionName string `json:"CollectionName"`
	Order          uint8  `json:"Order"`
	BatchNumLevel  uint8  `json:"BatchNumLevel"`
	BatchBaseSize  uint32 `json:"BatchBaseSize"`
	BatchIncrement uint8  `json:"BatchIncrement"`
	BatchLength    uint8  `json:"BatchLength"`
}

func (s *Secretary) newTreeHandler(w http.ResponseWriter, r *http.Request) {
	var req NewTreeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tree, err := s.NewBTree(
		req.CollectionName,
		req.Order,
		req.BatchNumLevel,
		req.BatchBaseSize,
		req.BatchIncrement,
		req.BatchLength,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tree.SaveHeader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.AddTree(tree)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("New tree created"))
}

type InsertRequest struct {
	Value string `json:"value"`
}

func (s *Secretary) insertHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	var req InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(strings.Trim(req.Value, " ")) == 0 {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tree, exists := s.trees[table]
	if !exists {
		http.Error(w, "Tree not found", http.StatusInternalServerError)
		return
	}

	key := []byte(utils.GenerateSeqRandomString(16, 4, req.Value))
	err := tree.Set(key, key)
	if err != nil {
		http.Error(w, "Tree not found", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Data inserted successfully",
		"table":   table,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Secretary) searchHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	id := r.PathValue("id")

	tree, exists := s.trees[table]
	if !exists {
		http.Error(w, "Tree not found", http.StatusInternalServerError)
		return
	}

	node, index, found := tree.SearchLeafNode([]byte(id))
	var record string
	if found {
		record = string(node.records[index].Value)
	} else {
		http.Error(w, "Key not found", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"table":  table,
		"result": fmt.Sprintf("[NodeID:%d  Index:%d  Found:%v  Value:%s]", node.NodeID, index, found, record),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Secretary) deleteHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	id := r.PathValue("id")

	tree, exists := s.trees[table]
	if !exists {
		http.Error(w, "Tree not found", http.StatusInternalServerError)
		return
	}

	err := tree.Delete([]byte(id))
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"table":  table,
		"result": "Delete success " + table + id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Secretary) setupRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /getalltree", s.getAllBTreeHandler)
	mux.HandleFunc("GET /gettree/{table}", s.getBTreeHandler)
	mux.HandleFunc("POST /newtree", s.newTreeHandler)
	mux.HandleFunc("POST /insert/{table}", s.insertHandler)
	mux.HandleFunc("GET /search/{table}/{id}", s.searchHandler)
	mux.HandleFunc("DELETE /delete/{table}/{id}", s.deleteHandler)

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
	port := 8080
	fmt.Printf("\nServer running on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), s.setupRouter()))
}
