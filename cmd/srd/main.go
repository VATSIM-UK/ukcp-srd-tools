package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/airac"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/db"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/file"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/lock"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/parse"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/srd"
	"github.com/alecthomas/kong"
)

var CLI struct {
	Parse struct {
		Filename string `arg:"" name:"filename" type:"path" help:"The filename of the SRD file to parse"`
	} `cmd:"" help:"Parse an SRD file"`
	Airac struct {
	} "cmd help:\"Get information about the current AIRAC cycle\""
	Import struct {
		Filename string `arg:"" name:"filename" type:"path" help:"The filename of the SRD file to import"`
	} `cmd:"" help:"Import an SRD file"`
}

func main() {
	// Get the filename from the command line
	cmd := kong.Parse(&CLI)
	switch cmd.Command() {
	case "parse <filename>":
		doParse()
	case "import <filename>":
		doImport()
	case "airac":
		doAirac()
	}
}

func doParse() {
	// Get the filename from the command line
	path, _ := filepath.Abs(CLI.Parse.Filename)

	file, err := loadSrdFile(path)
	if err != nil {
		fmt.Printf("Failed to load SRD file %v: %v\n", path, err)
		os.Exit(1)
	}

	// Parse the SRD file
	fmt.Printf("Parsing SRD file %v\n", path)
	summary := parse.ParseSrd(file)

	fmt.Printf("Parsed %v routes with %v errors\n", summary.RouteCount, summary.RouteErrorCount)
	fmt.Printf("Parsed %v notes with %v errors\n", summary.NoteCount, summary.NoteErrorCount)
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

func doImport() {
	lockfile, err := lock.NewLock()
	if err == lock.ErrAlreadyLocked {
		fmt.Println("Another process is already running")
		os.Exit(1)
	} else if err != nil {
		fmt.Printf("Failed to acquire lock: %v\n", err)
		os.Exit(1)
	}
	defer lockfile.Unlock()

	// Get the filename from the command line
	path, _ := filepath.Abs(CLI.Import.Filename)

	file, err := loadSrdFile(path)
	if err != nil {
		fmt.Printf("Failed to load SRD file %v: %v\n", path, err)
		os.Exit(1)
	}

	// Create a database connection
	db, err := db.NewDatabase(db.DatabaseConnectionParams{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "secret",
		Database: "uk_plugin",
	})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	defer db.Close()

	// Create the importer and go
	importer := srd.NewImport(file, db)

	err = importer.Import(context.Background())
	if err != nil {
		fmt.Printf("Failed to import SRD file %v: %v\n", path, err)
		os.Exit(1)
	}

	fmt.Printf("Imported SRD file %v\n", path)
}

func loadSrdFile(path string) (file.SrdFile, error) {
	excelFile, err := excel.NewExcelFile(path)
	if err != nil {
		return nil, err
	}

	srdFile, err := file.NewSrdFile(excelFile)
	if err != nil {
		return nil, err
	}

	return srdFile, nil
}
