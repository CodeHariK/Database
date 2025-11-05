test ?= ".*"  # Default: Run all tests if no name is given

clean:
	@([ -d "SECRETARY" ] && rm -rf SECRETARY/* || true)
	@go clean -testcache

one:
	go test ./... -v -run $(test) -- TEST

vul:
	go test ./... -v -run "^$(test).*$$" -- TEST
mul:
	go test ./... -run "^$(test).*$$" -- TEST
	echo ""
wul:
	go test ./... -run "^$(test).*$$" -- TEST WASM
	echo ""

wea: clean
	make mul test=TestBTree
	make mul test=TestNode

tea: clean
	make mul test=TestSecretary
	make mul test=TestServer
	make mul test=TestBTree
	make mul test=TestNode
	make mul test=TestPager
	make testutils

test: clean
	go test ./... -cover -coverprofile=cover.txt -- TEST

vtest: clean
	go test ./... -v -cover -coverprofile=cover.txt -- TEST

cover: clean
	go test ./... -cover -coverprofile=cover.txt -- TEST
	go tool cover -func=cover.txt
	go tool cover -html=cover.txt

run:
	go run example/main.go

ui:
	cd secretaryui && bun run dev

testutils:
	go test ./utils/... -- TEST

gen:
	buf dep update
	buf format -w
	buf generate

.PHONY: wasm
wasm:
	GOOS=js GOARCH=wasm go build -o ./secretaryui/secretary.wasm ./wasm

cpwasm:
	cp $(shell go env GOROOT)/lib/wasm/wasm_exec.js ./secretaryui
