package secretary

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/codeharik/secretary/api/apiconnect"
	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/secretary/utils/file"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	_ "go.uber.org/automaxprocs"
)

func New() (*Secretary, error) {
	utils.Log("Hello Secretary!")

	secretary := &Secretary{
		trees: map[string]*BTree{},
	}

	dirPath := "./SECRETARY"

	err := file.EnsureDir(dirPath)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {

			tree, err := secretary.NewBTreeReadHeader(file.Name())
			if err == nil && tree.CollectionName == file.Name() {
				secretary.AddTree(tree)
				utils.Log("[DIR] *", file.Name())
			} else {
				// utils.Log("[DIR] ", file.Name(), " ", err)
			}
		}
	}

	// Create a TCP listener on a random available port, OS assigns a free port
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle(apiconnect.NewSecretaryHandler(&secretary))

	handler := secretary.setupRouter(mux)

	server := &http.Server{
		Addr: listener.Addr().String(), // Eg:"127.0.0.1:54321"
		Handler: h2c.NewHandler(
			handler,
			&http2.Server{},
		),
	}

	secretary.listener = listener
	secretary.server = server

	return secretary, nil
}

func (s *Secretary) Tree(name string) (*BTree, error) {
	tree, ok := s.trees[name]
	if !ok {
		return nil, ErrorTreeNotFound
	}
	return tree, nil
}

func (s *Secretary) AddTree(tree *BTree) {
	s.trees[tree.CollectionName] = tree
}

func (s *Secretary) Close() {
	for _, ss := range s.trees {
		if err := ss.close(); err != nil {
			utils.Log("Error closing", ss.CollectionName, err)
		}
	}
	s.Shutdown()
}
