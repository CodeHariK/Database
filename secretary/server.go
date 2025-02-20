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

func writeJson(w http.ResponseWriter, code int, data any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (s *Secretary) getAllTreeHandler(w http.ResponseWriter, r *http.Request) {
	var hello []*BTree
	for _, o := range s.trees {
		hello = append(hello, o)
	}
	writeJson(w, http.StatusOK, hello)
}

func (s *Secretary) getTreeHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	tree, exists := s.trees[table]
	if !exists {
		writeJson(w, http.StatusNotFound, "Tree not found")
		return
	}

	m, err := tree.MarshalGraphJSON()
	if err != nil || m == nil {
		writeJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	if errs := tree.TreeVerify(); errs != nil {
		w.WriteHeader(http.StatusConflict)
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
		writeJson(w, http.StatusBadRequest, "Invalid Json")
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
		writeJson(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = tree.SaveHeader()
	if err != nil {
		writeJson(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.AddTree(tree)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("New tree created"))
}

type SetRequest struct {
	Value string `json:"value"`
}

var keySeq uint64 = 0

func (s *Secretary) setRecordHandler(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(strings.Trim(req.Value, " ")) == 0 {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	tree, exists := s.trees[table]
	if !exists {
		writeJson(w, http.StatusNotFound, "Tree not found")
		return
	}

	key := []byte(utils.GenerateSeqRandomString(&keySeq, 16, 4, req.Value))
	err := tree.Set(key, key)
	if err != nil {
		writeJson(w, http.StatusNotFound, "Tree not found")
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

func (s *Secretary) setupRouter() http.Handler {
	mux := http.NewServeMux()

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
	port := 8080
	utils.Log("Server running on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), s.setupRouter()))
}
