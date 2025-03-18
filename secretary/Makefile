test ?= ".*"  # Default: Run all tests if no name is given

clean:
	@([ -d "SECRETARY" ] && rm -rf SECRETARY/* || true)
	@go clean -testcache

one: clean
	go test ./... -v -run $(test)

test: clean
	go test -cover -coverprofile=cover.txt ./...

vtest: clean
	go test -cover -coverprofile=cover.txt -v ./...

cover: clean
	go test -cover -coverprofile=cover.txt ./...
	go tool cover -func=cover.txt
	go tool cover -html=cover.txt

run:
	go run example/main.go

ui:
	cd secretaryui && bun run dev

testutils:
	go test -v utils/*.go

callgraph:
	go run utils/ast/main.go

serf:
	go run serf/main.go
	go run serf/main.go node2

gen:
	buf dep update
	buf format -w
	buf generate
