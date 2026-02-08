package main

import (
	"os"

	"github.com/nicolas/arch-index/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
