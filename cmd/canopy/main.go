package main

import (
	"os"

	"github.com/nhomble/canopy/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
