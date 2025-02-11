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
	table := r.URL.Query().Get("table")
	if table == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	tree, exists := s.trees[table]
	if !exists {
		http.Error(w, "Tree not found", http.StatusNotFound)
		return
	}

	m, err := tree.ConvertBTreeToJSON()
	if err != nil || m == nil {
		http.Error(w, "Tree not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(m)
}

type InsertRequest struct {
	Value string `json:"value"`
}

func (s *Secretary) insertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	table := r.URL.Query().Get("table")
	if table == "" {
		http.Error(w, "Missing table parameter", http.StatusBadRequest)
		return
	}

	var req InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(strings.Trim(req.Value, " ")) == 0 {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tree, exists := s.trees[table]
	if !exists {
		http.Error(w, "Tree not found", http.StatusNotFound)
		return
	}

	key := []byte(utils.GenerateSeqRandomString(16, 4, req.Value))
	err := tree.Insert(key, key)
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

func (s *Secretary) Serve() {
	mux := http.NewServeMux()

	mux.HandleFunc("/getallbtree", s.getAllBTreeHandler)
	mux.HandleFunc("/getbtree", s.getBTreeHandler)
	mux.HandleFunc("/insert", s.insertHandler)

	// Enable CORS with custom settings
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "OPTIONS", "POST", "DELETE"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	port := 8080
	fmt.Printf("\nServer running on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}
