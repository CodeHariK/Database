package main

import (
	"os"

	"github.com/codeharik/secretary"
	"github.com/codeharik/secretary/utils"
)

func main() {
	s, err := secretary.New(nil)
	if err != nil {
		utils.Log(err)
		os.Exit(1)
	}

	secretary.DummyInitTrees(s)

	s.Serve()
}
