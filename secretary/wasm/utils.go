//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"reflect"
	"syscall/js"
)

func SetWASMLibaray(lib map[string]any) {
	fmt.Println("Go WebAssembly Loaded!")

	jslib := js.Global().Get("Object").New()
	for key, val := range lib {

		if reflect.TypeOf(val).Kind() == reflect.String {
			fmt.Println(val)
		}

		switch v := val.(type) {
		case js.Func, string:
			jslib.Set(key, val)
		default:
			fmt.Println("Invalid type in lib:", v)
		}
	}
	js.Global().Set("lib", js.ValueOf(jslib))

	// Keep the program running
	select {}
}

func jsResponse(data []byte, err error) any {
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": err.Error(),
		})
	}
	return js.ValueOf(map[string]interface{}{
		"data": string(data),
	})
}
