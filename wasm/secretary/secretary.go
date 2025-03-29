//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"log"
	"os"
	"syscall/js"

	"github.com/codeharik/secretary"
	"github.com/codeharik/secretary/utils"
	"github.com/codeharik/wasm"
)

var SECRETARY *secretary.Secretary

func init() {
	MODEWASM := true
	s, err := secretary.New(&MODEWASM)
	if err != nil {
		utils.Log(err)
		os.Exit(1)
	}

	SECRETARY = s

	users, userErr := s.NewBTree(
		"users",
		4,
		32,
		1024,
		125,
		1000,
	)

	if userErr != nil || users == nil {
		log.Fatal(userErr)
	}

	fmt.Println(users.CollectionName)
}

func main() {
	var lib map[string]any = map[string]any{
		"set":      js.FuncOf(set),
		"set_info": "set(value string) -> bool",
	}

	wasm.SetWASMLibaray(lib)

	select {}
}

func set(this js.Value, args []js.Value) any {
	value := args[0].String()

	tree, err := SECRETARY.Tree("users")
	if err != nil {
		return js.ValueOf(err.Error())
	}

	key, err := tree.Set([]byte(value))
	if err != nil {
		return js.ValueOf(err.Error())
	}

	v, err := tree.Get(key)
	if err != nil {
		return js.ValueOf(err.Error())
	}

	return js.ValueOf(fmt.Sprintf("-> key:%b, value:%s, %v", key, string(v.Value), err))
}
