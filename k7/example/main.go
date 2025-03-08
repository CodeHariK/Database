package main

import (
	"fmt"
	"time"

	"github.com/codeharik/k7"
	"github.com/go-resty/resty/v2"
)

var (
	restyClients  = []resty.Client{}
	restyClientId = 0
)

func init() {
	for i := 0; i < 3; i++ {
		restyClients = append(restyClients, *resty.New())
	}
}

func restyAttack() {
	config := k7.BenchmarkConfig{
		Threads:  3,
		Duration: 10 * time.Second,
		AttackFunc: func() bool {
			res, err := restyClients[restyClientId].R().Get("http://localhost:8080")
			if err != nil {
				fmt.Println(err)
				return false
			}

			restyClientId = (restyClientId + 1) % len(restyClients)

			return res.StatusCode() == 200
		},
	}
	config.Attack()
}

type Mode int

const (
	Connect Mode = iota
	Fiber
	NetHTTP
)

func (m Mode) String() string {
	return [...]string{"Connect", "Fiber", "NetHTTP"}[m]
}

func main() {
	mode := Fiber

	switch mode {
	case Connect:
		resty := false
		fmt.Println("Starting Connect Server, Resty:", resty)
		connectServer(resty)
	case Fiber:
		fmt.Println("Starting Fiber Server")
		fiberServer()
	case NetHTTP:
		fmt.Println("Starting NetHTTP Server")
		nethttpServer()
	default:
		fmt.Println("Invalid mode")
	}
}
