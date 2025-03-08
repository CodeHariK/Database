package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/codeharik/k7"
	"github.com/codeharik/k7/example/api"
	"github.com/codeharik/k7/example/api/apiconnect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type GreetServer struct{}

func (s *GreetServer) Greet(
	ctx context.Context,
	req *connect.Request[api.GreetRequest],
) (*connect.Response[api.GreetResponse], error) {
	res := connect.NewResponse(&api.GreetResponse{
		Greeting: fmt.Sprintf("Hello, %s!", req.Msg.Name),
	})
	res.Header().Set("Greet-Version", "v1")

	return res, nil
}

func connectServer(resty bool) {
	greeter := &GreetServer{}
	mux := http.NewServeMux()
	path, handler := apiconnect.NewGreetServiceHandler(greeter)
	mux.Handle(path, handler)

	go func() {
		http.ListenAndServe(
			"localhost:8080",
			h2c.NewHandler(mux,
				&http2.Server{},
			),
		)
	}()

	if resty {
		restyAttack()
	} else {
		clients := []apiconnect.GreetServiceClient{}
		for i := 0; i < 3; i++ {
			clients = append(clients, apiconnect.NewGreetServiceClient(
				&http.Client{Timeout: time.Millisecond * 100},
				"http://localhost:8080",
				// connect.WithGRPCWeb(),
				// connect.WithGRPC(),
				// connect.WithHTTPGet(),
			))
		}
		cc := 0
		config := k7.BenchmarkConfig{
			Threads:  3,
			Duration: 10 * time.Second,
			AttackFunc: func() bool {
				_, err := clients[cc].Greet(
					context.Background(),
					connect.NewRequest(&api.GreetRequest{Name: "Jane"}),
				)
				if err != nil {
					fmt.Println(err)
					return false
				}

				cc = (cc + 1) % len(clients)

				return true
			},
		}
		config.Attack()
	}
}
