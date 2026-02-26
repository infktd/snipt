package main

import (
	"fmt"
	"os"
)

// Version is set by goreleaser at build time.
var version = "dev"

func main() {
	fmt.Fprintln(os.Stderr, "snipt", version)
	os.Exit(0)
}
