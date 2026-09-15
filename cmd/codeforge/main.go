package main

import (
	"os"

	"github.com/mxl1c/CodeForge/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
