package secretary

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// getBTreeHandler handles the /getbtree route
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
	if err != nil {
		http.Error(w, "Tree not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func main() {
	secretary := New()

	http.HandleFunc("/getbtree", secretary.getBTreeHandler)

	port := 8080
	fmt.Printf("Server running on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
