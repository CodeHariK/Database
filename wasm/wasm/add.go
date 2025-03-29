//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/codeharik/wasm"
)

func add(this js.Value, args []js.Value) any {
	sum := args[0].Int() + args[1].Int()
	return js.ValueOf(sum)
}

func sub(this js.Value, args []js.Value) any {
	sum := args[0].Int() - args[1].Int()
	return js.ValueOf(sum)
}

func main() {
	var lib map[string]any = map[string]any{
		"add":      js.FuncOf(add),
		"add_info": "add(a int, b int) -> int",

		"sub":      js.FuncOf(sub),
		"sub_info": "sub(a int, b int) -> int",
	}

	wasm.SetWASMLibaray(lib)
}
