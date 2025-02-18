test ?= ".*"  # Default: Run all tests if no name is given

clean:
	@([ -d "SECRETARY" ] && rm -rf SECRETARY/* || true)
	@go clean -testcache

one: clean
	go test ./... -run $(test)

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
	