package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/airac"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/file"
	"github.com/alecthomas/kong"
)

var CLI struct {
	Parse struct {
		Filename string `arg:"" name:"filename" type:"path" help:"The filename of the SRD file to parse"`
	} "cmd help:\"Parse an SRD file\""
	Airac struct {
	} "cmd help:\"Get information about the current AIRAC cycle\""
}

func main() {
	// Get the filename from the command line
	cmd := kong.Parse(&CLI)
	switch cmd.Command() {
	case "parse <filename>":
		doParse()
	case "airac":
		doAirac()
	}
}

func doParse() {
	// Get the filename from the command line
	path, _ := filepath.Abs(CLI.Parse.Filename)

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

func doAirac() {
	// Get the current AIRAC cycle
	airac := airac.NewAirac(nil)

	currentCycle := airac.CurrentCycle()

	// Print the current cycle identifier with the start and end dates, dates to be formatted as "YYYY-MM-DD"
	fmt.Printf(
		"Current AIRAC cycle is %v (%v - %v)\n",
		currentCycle.Ident,
		currentCycle.Start.Format("2006-01-02"),
		currentCycle.End.Format("2006-01-02"),
	)

	// Print the next cycle identifier with the start and end dates
	nextCycle := airac.NextCycleFrom(currentCycle)
	fmt.Printf(
		"Next is %v (%v - %v)\n",
		nextCycle.Ident,
		nextCycle.Start.Format("2006-01-02"),
		nextCycle.End.Format("2006-01-02"),
	)
}
