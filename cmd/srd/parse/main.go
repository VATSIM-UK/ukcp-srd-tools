package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/file"
	"github.com/alecthomas/kong"
)

var CLI struct {
	Filename string `arg:"" type:"path" help:"The filename of the SRD file to parse"`
}

func main() {
	// Get the filename from the command line
	cmd := kong.Parse(&CLI)
	switch cmd.Command() {
	case "<filename>":
		doParse()
	}
}

func doParse() {
	// Get the filename from the command line
	path, _ := filepath.Abs(CLI.Filename)

	// Check if the file exists
	_, err := os.Stat(path)
	if err != nil {
		fmt.Printf("File %v does not exist\n", path)
		os.Exit(1)
	}

	excelFile, err := excel.NewExcelFile(path)
	defer excelFile.Close()
	if err != nil {
		fmt.Printf("Failed to open excel file %v: %v\n", path, err)
		os.Exit(1)
	}

	file, err := file.NewSrdFile(excelFile)
	if err != nil {
		fmt.Printf("Failed to load SRD file %v: %v\n", path, err)
		os.Exit(1)
	}

	routeCount := 0
	routeErrorCount := 0
	noteCount := 0
	noteErrorCount := 0

	fmt.Printf("Parsing SRD file %v\n", path)
	for _, err := range file.Routes() {
		if err != nil {
			routeErrorCount++
		} else {
			routeCount++
		}
	}

	for _, err := range file.Notes() {
		if err != nil {
			noteErrorCount++
		} else {
			noteCount++
		}
	}

	fmt.Printf("Parsed %v routes with %v errors\n", routeCount, routeErrorCount)
	fmt.Printf("Parsed %v notes with %v errors\n", noteCount, noteErrorCount)
}
