//go:build js

package secretary

import (
	"sync"
)

type Secretary struct {
	trees map[string]*BTree

	quit chan any
	wg   sync.WaitGroup
	once sync.Once
}
