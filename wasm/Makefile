build:
	GOOS=js GOARCH=wasm go build -o wasm/secretary.wasm wasm/secretary.go

run:
	go run server/main.go

cp:
	cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
