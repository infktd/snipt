package main

import (
	"os"

	"github.com/infktd/snipt/src/internal/cli"
	"github.com/infktd/snipt/src/internal/model"
)

var version = "dev"

func main() {
	root := cli.NewRootCmd(version)
	if err := root.Execute(); err != nil {
		os.Exit(model.ExitError)
	}
}
