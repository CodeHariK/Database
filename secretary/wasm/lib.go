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
		"allTree":      js.FuncOf(allTree),
		"allTree_info": "allTree() -> json",

		"set":      js.FuncOf(set),
		"set_info": "set(value string) -> string",

		"add":      js.FuncOf(add),
		"add_info": "add(a int, b int) -> int",

		"sub":      js.FuncOf(sub),
		"sub_info": "sub(a int, b int) -> int",
	}

	SetWASMLibaray(lib)

	select {}
}

func allTree(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleGetAllTree()
	if err != nil {
		return js.ValueOf(err.Error())
	}

	return js.ValueOf(string(data))
}

func set(this js.Value, args []js.Value) any {
	value := args[0].String()

	key, data, err := SECRETARY.HandleSetRecord("users", "", value)
	if err != nil {
		return js.ValueOf(err.Error())
	}

	fmt.Println(string(data))

	v, err := SECRETARY.HandleGetRecord("users", string(key))
	if err != nil {
		return js.ValueOf(err.Error())
	}

	return js.ValueOf(fmt.Sprintf("-> key:%v, value:%s, %v", key, string(v), err))
}

func add(this js.Value, args []js.Value) any {
	sum := args[0].Int() + args[1].Int()
	return js.ValueOf(sum)
}

func sub(this js.Value, args []js.Value) any {
	sum := args[0].Int() - args[1].Int()
	return js.ValueOf(sum)
}
