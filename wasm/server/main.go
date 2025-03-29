package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := "8080"
	dir := "."

	// Get port from environment variable if set
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	fs := http.FileServer(http.Dir(dir))

	// Enable directory listing and file streaming
	http.Handle("/", http.StripPrefix("/", fs))

	fmt.Printf("Serving files on http://localhost:%s/\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
