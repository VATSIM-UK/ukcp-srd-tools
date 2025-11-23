package main

import (
	"fmt"
	"os"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/cli"
)

// Where to store files
const dir = "/tmp"

func main() {
	err := cli.Run(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
