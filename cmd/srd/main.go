package main

import (
	"os"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/cli"
)

// Where to store files
const dir = "/tmp"

func main() {
	err := cli.Run(dir)
	if err != nil {
		os.Exit(1)
	}
}
