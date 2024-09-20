package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/airac"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/db"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/download"
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
		Cycle    string `arg:"" name:"cycle" help:"The identfier of the AIRAC cycle being imported"`
		Filename string `arg:"" name:"filename" type:"path" help:"The filename of the SRD file to import"`
	} `cmd:"" help:"Import an SRD file"`
	Download struct {
		// Force is an argument presented as --force or -f
		Force bool `short:"f" help:"Force download of the SRD file"`
	} `cmd:"" help:"Download the SRD file"`
}

func main() {
	// Get the filename from the command line
	cmd := kong.Parse(&CLI)
	switch cmd.Command() {
	case "parse <filename>":
		doParse()
	case "import <cycle> <filename>":
		doImport(CLI.Import.Filename)
	case "airac":
		doAirac()
	case "download":
		doDownload(CLI.Download.Force)
	default:
		fmt.Print("Incorrect command format - check code")
		os.Exit(1)
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

	printStats(summary)
}

func printStats(stats file.SrdStats) {
	fmt.Printf("Processed %v routes with %v errors\n", stats.RouteCount, stats.RouteErrorCount)
	fmt.Printf("Processed %v notes with %v errors\n", stats.NoteCount, stats.NoteErrorCount)
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

func doImport(filePath string) {
	lockfile, err := lock.NewLock()
	if err == lock.ErrAlreadyLocked {
		fmt.Println("Another process is already running")
		os.Exit(1)
	} else if err != nil {
		fmt.Printf("Failed to acquire lock: %v\n", err)
		os.Exit(1)
	}
	defer lockfile.Unlock()
	importProcess(filePath)
}

func importProcess(filePath string) {
	// Get the filename from the command line
	path, _ := filepath.Abs(filePath)

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

	fmt.Printf("Imported SRD for cycle %v\n", CLI.Import.Cycle)

	// Print the stats
	printStats(file.Stats())
}

func doDownload(force bool) {
	lockfile, err := lock.NewLock()
	if err == lock.ErrAlreadyLocked {
		fmt.Println("Another process is already running")
		os.Exit(1)
	} else if err != nil {
		fmt.Printf("Failed to acquire lock: %v\n", err)
		os.Exit(1)
	}
	defer lockfile.Unlock()

	// Get the current AIRAC cycle
	airac := airac.NewAirac(nil)

	currentCycle := airac.CurrentCycle()

	// Download the SRD file
	downloadUrl := download.DownloadUrl(currentCycle)
	downloader, err := download.NewSrdDownloader(currentCycle, "/tmp", downloadUrl)
	if err != nil {
		fmt.Printf("Failed to create downloader: %v\n", err)
		os.Exit(1)
	}

	err = downloader.Download(context.Background(), force)
	if err == download.ErrUpToDate {
		fmt.Println("SRD file is up to date, use --force to download anyway")
		os.Exit(0)
	} else if err != nil {
		fmt.Printf("Failed to download SRD file: %v\n", err)
		os.Exit(1)
	}

	// Download happened, so now we do the import
	importProcess(downloader.LatestFileLocation())
}

func loadSrdFile(path string) (file.SrdFile, error) {
	excelFile, err := loadExcelFile(path)
	if err != nil {
		return nil, err
	}

	srdFile, err := file.NewSrdFile(excelFile)
	if err != nil {
		return nil, err
	}

	return srdFile, nil
}

// Load the right excel reader (xls or xlsx) based on the file extension
func loadExcelFile(path string) (excel.ExcelFile, error) {
	ext := filepath.Ext(path)
	if ext == ".xls" {
		return excel.NewExcelFile(path)
	} else if ext == ".xlsx" {
		return excel.NewExcelExtendedFile(path)
	}

	return nil, fmt.Errorf("Unknown file extension %v", ext)
}
