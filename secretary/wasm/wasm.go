//go:build js && wasm
// +build js,wasm

package main

import (
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

	secretary.DummyInitTrees(SECRETARY)
}

func main() {
	var lib map[string]any = map[string]any{
		"getAllTree":      js.FuncOf(getAllTree),
		"getAllTree_info": "getAllTree() -> data, error",

		"getTree":      js.FuncOf(getTree),
		"getTree_info": "getTree() -> data, error",

		"newTree":      js.FuncOf(newTree),
		"newTree_info": "newTree(collectionName string, order int, numLevel int, baseSize int, increment int, compactionBatchSize int) -> data, error",

		"clearTree":      js.FuncOf(clearTree),
		"clearTree_info": "clearTree() -> data, error",

		"getRecord":      js.FuncOf(getRecord),
		"getRecord_info": "getRecord(tree string, key string) -> data, error",

		"setRecord":      js.FuncOf(setRecord),
		"setRecord_info": "setRecord(tree string, key string, value string) -> data, error",

		"sortedSetRecord":      js.FuncOf(sortedSetRecord),
		"sortedSetRecord_info": "sortedSetRecord(tree string, value int) -> data, error",

		"deleteRecord":      js.FuncOf(deleteRecord),
		"deleteRecord_info": "deleteRecord(tree string, key string) -> data, error",
	}

	SetWASMLibaray(lib)

	select {}
}

func getAllTree(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleGetAllTree()
	return jsResponse(data, err)
}

func getTree(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleGetTree(args[0].String())
	return jsResponse(data, err)
}

func newTree(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleNewTree(
		args[0].String(),
		args[1].Int(),
		args[2].Int(),
		args[3].Int(),
		args[4].Int(),
		args[5].Int())
	return jsResponse(data, err)
}

func clearTree(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleClearTree(args[0].String())
	return jsResponse(data, err)
}

func setRecord(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleSetRecord(args[0].String(), args[1].String(), args[2].String())
	return jsResponse(data, err)
}

func sortedSetRecord(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleSortedSetRecord(args[0].String(), args[1].Int())
	return jsResponse(data, err)
}

func getRecord(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleGetRecord(args[0].String(), args[1].String())
	return jsResponse(data, err)
}

func deleteRecord(this js.Value, args []js.Value) any {
	data, err := SECRETARY.HandleDeleteRecord(args[0].String(), args[1].String())
	return jsResponse(data, err)
}
