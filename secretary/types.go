//go:build !js

package secretary

import (
	"net"
	"net/http"
	"sync"
)

type Secretary struct {
	trees map[string]*BTree

	listener net.Listener
	server   *http.Server

	httpClient http.Client

	quit chan any
	wg   sync.WaitGroup
	once sync.Once
}
