package secretary

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	tree, exists := s.trees[key]
	if !exists {
		http.Error(w, "Tree not found", http.StatusNotFound)
		return
	}

	m, err := tree.ConvertBTreeToJSON()
	if err != nil || m == nil {
		http.Error(w, "Tree not found", http.StatusInternalServerError)
		return
	}

	fmt.Println(string(m))

	w.Header().Set("Content-Type", "application/json")
	w.Write(m)
}

func (s *Secretary) Serve() {
	mux := http.NewServeMux()

	mux.HandleFunc("/getallbtree", s.getAllBTreeHandler)
	mux.HandleFunc("/getbtree", s.getBTreeHandler)

	// Enable CORS with custom settings
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	port := 8080
	fmt.Printf("\nServer running on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}
