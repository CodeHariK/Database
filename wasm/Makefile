cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .

GOOS=js GOARCH=wasm go build -o wasm/add.wasm wasm/add.go

GOOS=js GOARCH=wasm go build -o secretary/secretary.wasm secretary/secretary.go

go run server/main.go
