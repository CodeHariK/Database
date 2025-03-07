package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"connectrpc.com/connect"
	"github.com/codeharik/k7"
	"github.com/codeharik/k7/example/api"
	"github.com/codeharik/k7/example/api/apiconnect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// curl \
//     --header "Content-Type: application/json" \
//     --data '{"name": "Jane"}' \
//     http://localhost:8080/greet.GreetService/Greet

var COUNTER = 0

type GreetServer struct{}

func (s *GreetServer) Greet(
	ctx context.Context,
	req *connect.Request[api.GreetRequest],
) (*connect.Response[api.GreetResponse], error) {
	res := connect.NewResponse(&api.GreetResponse{
		Greeting: fmt.Sprintf("Hello, %s!", req.Msg.Name),
	})
	res.Header().Set("Greet-Version", "v1")

	COUNTER++

	return res, nil
}

func main() {
	runtime.GOMAXPROCS(8)

	greeter := &GreetServer{}
	mux := http.NewServeMux()
	path, handler := apiconnect.NewGreetServiceHandler(greeter)
	mux.Handle(path, handler)

	go func() {
		http.ListenAndServe(
			"localhost:8080",
			h2c.NewHandler(mux, &http2.Server{}),
		)
	}()

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println(COUNTER)
			}
		}
	}()

	config := k7.BenchmarkConfig{
		Concurrency: 2,                // Number of concurrent users
		Duration:    10 * time.Second, // Run test for 10 seconds
		AttackFunc: func() bool {
			client := apiconnect.NewGreetServiceClient(
				http.DefaultClient,
				"http://localhost:8080",
				connect.WithGRPC(),
			)
			_, err := client.Greet(
				context.Background(),
				connect.NewRequest(&api.GreetRequest{Name: "Jane"}),
			)
			if err != nil {
				fmt.Println(err)
				return false
			}

			return true
		},
	}

	config.Attack()
}

// func main() {
// 	runtime.GOMAXPROCS(8)

// 	go func() {
// 		server := http.ServeMux{}

// 		// Set up handlers
// 		server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 			COUNTER++

// 			// if rand.IntN(10) < 5 {
// 			// 	http.Error(w, "Random error", http.StatusConflict)
// 			// 	return
// 			// }
// 			w.WriteHeader(http.StatusOK)
// 			w.Write([]byte("Hello world!"))
// 			return
// 		})

// 		ticker := time.NewTicker(1 * time.Second)

// 		go func() {
// 			for {
// 				select {
// 				case <-ticker.C:
// 					fmt.Println(COUNTER)
// 				}
// 			}
// 		}()

// 		// Start the server
// 		port := 8080
// 		fmt.Printf("Server running on http://localhost:%d\n", port)
// 		http.ListenAndServe(fmt.Sprintf(":%d", port), &server)
// 	}()

// 	config := k7.BenchmarkConfig{
// 		Concurrency: 2,                // Number of concurrent users
// 		Duration:    10 * time.Second, // Run test for 10 seconds
// 		AttackFunc: func() bool {
// 			client := &http.Client{}

// 			res, err := client.Get("http://localhost:8080")
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 			defer res.Body.Close()
// 			return err == nil && res.StatusCode == 200
// 		},
// 	}

// 	config.Attack()
// }
