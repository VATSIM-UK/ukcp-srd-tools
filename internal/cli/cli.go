package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/airac"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/db"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/download"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/file"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/lock"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/parse"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/srd"
)

// CLI is the command line interface structure
var CLI struct {
	Loaded struct {
	} `cmd:"" help:"Show information about the currently loaded airac version"`
	Parse struct {
		Filename string `arg:"" name:"filename" type:"path" help:"The filename of the SRD file to parse"`
	} `cmd:"" help:"Parse an SRD file"`
	Airac struct {
	} `cmd:"" help:"Get information about the current AIRAC cycle"`
	Import struct {
		Cycle string `arg:"" name:"cycle" help:"The identfier of the AIRAC cycle being imported"`

		Filename string `arg:"" name:"filename" type:"path" help:"The filename of the SRD file to import"`

		// EnvPath is an optional argument, presented as --env-path or -e, its default value is .env
		EnvPath string `short:"e" help:"Path to the .env file" default:".env"`
	} `cmd:"" help:"Import an SRD file"`
	Download struct {
		// Force is an argument presented as --force or -f
		Force bool `short:"f" help:"Force download of the SRD file"`

		// EnvPath is an optional argument, presented as --env-path or -e, its default value is .env
		EnvPath string `short:"e" help:"Path to the .env file" default:".env"`

		// Cycle is an optional argument, presented as --cycle or -c, used with the Force argument to force set the AIRAC cycle
		Cycle string `short:"c" help:"The identfier of the AIRAC cycle to download"`

		// A forced URL to download the SRD file from
		Url string `short:"u" help:"The URL to download the SRD file from"`
	} `cmd:"" help:"Download the SRD file"`
	// Add a verbosity flag to the CLI, represented as -v or --verbose. This increases the log level to debug
	Verbose bool `short:"v" help:"Enable debug logging"`

	// Add a debug flag to the CLI, represented as -d or --debug. This increases the log level to trace
	Debug bool `short:"d" help:"Enable trace logging"`

	// The Testing flag is used to enable testing mode, it prevents logger modifications.
	// The flag takes no arguments and is represented as -t or --testing
	Testing bool `short:"t" help:"Enable testing mode"`
}

var (
	// Invalid command errors
	ErrInvalidCommandFormat = errors.New("invalid command format - check the code")
	ErrAlreadyRunning       = errors.New("another process is already running")
	ErrFailedProcessLock    = errors.New("failed to acquire process lock")
	ErrUnknownFileExtension = errors.New("unknown file extension, must be .xls or .xlsx")
	ErrCannotLoadDotenv     = errors.New("failed to load environment file")

	// Misc runtime errors
	ErrUpToDate = errors.New("SRD file is up to date, use --force to download anyway")

	// Database-specific errors
	ErrMissingHost     = errors.New("missing database host")
	ErrMissingUser     = errors.New("missing database user")
	ErrMissingPass     = errors.New("missing database password")
	ErrMissingDatabase = errors.New("missing database name")
	ErrMissingPort     = errors.New("missing database port")
	ErrPortInvalid     = errors.New("invalid database port")
)

// Run runs the CLI, parsing the command line arguments and executing the appropriate command
func Run(dir string) error {
	cmd := kong.Parse(&CLI)
	configureLogging()

	switch cmd.Command() {
	case "parse <filename>":
		return doParse()
	case "import <cycle> <filename>":
		return doImport(CLI.Import.Filename, CLI.Import.Cycle, CLI.Import.EnvPath, dir)
	case "airac":
		return doAirac()
	case "download":
		return doDownload(CLI.Download.Force, CLI.Download.Cycle, CLI.Download.EnvPath, dir)
	case "loaded":
		return doLoaded(dir)
	default:
		return ErrInvalidCommandFormat
	}
}

// configureLogging configures the logging based on the CLI flags
// if the testing flag is set, logging is controlled by the test
func configureLogging() {
	if CLI.Testing {
		return
	}

	// Set the log level to trace if the debug flag is set
	if CLI.Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if CLI.Debug {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Turn on pretty console logging, if the unit tests are running, disable this
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// doParse parses an SRD file to check for errors
func doParse() error {
	// Get the filename from the command line
	path, err := filepath.Abs(CLI.Parse.Filename)
	if err != nil {
		return err
	}

	file, err := loadSrdFile(path)
	if err != nil {
		return err
	}

	// Parse the SRD file
	log.Info().Msgf("Parsing SRD file %v", path)
	summary := parse.ParseSrd(file)

	printStats(summary)
	return nil
}

func doLoaded(dir string) error {
	loadedCycle, err := airac.NewLoadedAirac(dir)
	if err != nil {
		return err
	}

	if loadedCycle.Ident() == "" {
		log.Info().Msgf("No AIRAC cycle loaded")
		return nil
	}

	// Get the current AIRAC cycle
	airacManager := airac.NewAirac(nil)
	currentCycle, _ := airacManager.CycleFromIdent(loadedCycle.Ident())

	log.Info().Msgf(
		"Loaded AIRAC cycle is %v (%v - %v)",
		currentCycle.Ident,
		currentCycle.Start.Format("2006-01-02"),
		currentCycle.End.Format("2006-01-02"),
	)

	return nil
}

func printStats(stats file.SrdStats) {
	log.Info().Msgf("processed %v routes with %v errors", stats.RouteCount, stats.RouteErrorCount)
	log.Info().Msgf("processed %v notes with %v errors", stats.NoteCount, stats.NoteErrorCount)
}

// doAirac gets information about the current AIRAC cycle
func doAirac() error {
	// Get the current AIRAC cycle
	airac := airac.NewAirac(nil)

	currentCycle := airac.CurrentCycle()

	// Print the current cycle identifier with the start and end dates, dates to be formatted as "YYYY-MM-DD"
	log.Info().Msgf(
		"Current AIRAC cycle is %v (%v - %v)",
		currentCycle.Ident,
		currentCycle.Start.Format("2006-01-02"),
		currentCycle.End.Format("2006-01-02"),
	)

	// Print the next cycle identifier with the start and end dates
	nextCycle := airac.NextCycleFrom(currentCycle)
	log.Info().Msgf(
		"Next AIRAC cycle is %v (%v - %v)",
		nextCycle.Ident,
		nextCycle.Start.Format("2006-01-02"),
		nextCycle.End.Format("2006-01-02"),
	)

	return nil
}

// doImport imports an SRD file into the database
// it requires that the process lock is acquired before calling this function
func doImport(filePath string, cycle string, envPath string, fileDir string) error {
	lock, err := processLock()
	if err != nil {
		return err
	}
	defer lock.Unlock()

	return importProcess(filePath, cycle, envPath, fileDir)
}

// importProcess performs the import process and is shared between the import command and the download command
func importProcess(filePath string, cycle string, envPath string, fileDir string) error {
	// Get the filename from the command line
	path, _ := filepath.Abs(filePath)

	file, err := loadSrdFile(path)
	if err != nil {
		return err
	}

	// Check the cycle is valid
	airacCycles := airac.NewAirac(nil)
	airacCycle, err := airacCycles.CycleFromIdent(cycle)
	if err != nil {
		return err
	}

	log.Info().Msgf("importing SRD file %v for cycle %v", path, airacCycle.Ident)

	// Load the .env file
	err = godotenv.Overload(envPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to load environment file")
		return ErrCannotLoadDotenv
	}

	// Get the database connection parameters
	dbParams, err := getDatabaseConnectionParams()
	if err != nil {
		log.Error().Err(err).Msgf("failed to get database connection parameters: %v", err)
		return err
	}

	// Create a database connection
	db, err := db.NewDatabase(dbParams)
	if err != nil {
		return err
	}

	defer db.Close()

	// Create the importer and go
	importer := srd.NewImport(file, db)

	err = importer.Import(context.Background())
	if err != nil {
		return err
	}

	log.Info().Msgf("imported SRD for cycle %v", CLI.Import.Cycle)

	// Set the SRD cycle
	loadedCycle, err := airac.NewLoadedAirac(fileDir)
	if err != nil {
		return err
	}

	err = loadedCycle.Set(airacCycle)
	if err != nil {
		return err
	}

	// Print the stats
	printStats(file.Stats())

	return nil
}

// doDownload downloads the SRD file and imports it into the database
func doDownload(force bool, forceCycle string, envPath string, fileDir string) error {
	processLock, err := processLock()
	if err != nil {
		return err
	}

	defer processLock.Unlock()

	// Get the current AIRAC cycle
	airacManager := airac.NewAirac(nil)
	cycleToDownload := airacManager.CurrentCycle()

	// Set the import cycle, if not set use the current cycle
	if forceCycle != "" {
		cycleToDownload, err = airacManager.CycleFromIdent(forceCycle)
		if err != nil {
			return err
		}
	}

	// Get the currently loaded cycle
	loadedCycle, err := airac.NewLoadedAirac(fileDir)
	if err != nil {
		return err
	}

	// Download the SRD file
	downloadUrl := download.DownloadUrl(cycleToDownload)
	if CLI.Download.Url != "" {
		downloadUrl = CLI.Download.Url
	}

	downloader, err := download.NewSrdDownloader(cycleToDownload, loadedCycle, fileDir, downloadUrl)
	if err != nil {
		return err
	}

	err = downloader.Download(context.Background(), force)
	if err == download.ErrUpToDate {
		return ErrUpToDate
	} else if err != nil {
		return err
	}

	// Download happened, so now we do the import
	return importProcess(downloader.LatestFileLocation(), cycleToDownload.Ident, envPath, fileDir)
}

// loadSrdFile loads an SRD file from the given path
func loadSrdFile(path string) (file.SrdFile, error) {
	excelFile, err := loadExcelFile(path)
	if err != nil {
		log.Error().Err(err).Msgf("failed to load excel file %v", path)
		return nil, err
	}

	srdFile, err := file.NewSrdFile(excelFile)
	if err != nil {
		log.Error().Err(err).Msgf("failed to create SRD file from %v", path)
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

	log.Error().Msgf("unknown file extension %v", ext)
	return nil, ErrUnknownFileExtension
}

// Get the database connection parameters from the .env file
func getDatabaseConnectionParams() (db.DatabaseConnectionParams, error) {
	port := os.Getenv("DB_PORT")
	if port == "" {
		return db.DatabaseConnectionParams{}, ErrMissingPort
	}

	// Convert port to an integer
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return db.DatabaseConnectionParams{}, ErrPortInvalid
	}

	// Check the other required parameters
	host := os.Getenv("DB_HOST")
	if host == "" {
		return db.DatabaseConnectionParams{}, ErrMissingHost
	}

	user := os.Getenv("DB_USERNAME")
	if user == "" {
		return db.DatabaseConnectionParams{}, ErrMissingUser
	}

	pass := os.Getenv("DB_PASSWORD")
	if pass == "" {
		return db.DatabaseConnectionParams{}, ErrMissingPass
	}

	database := os.Getenv("DB_DATABASE")
	if database == "" {
		return db.DatabaseConnectionParams{}, ErrMissingDatabase
	}

	// Return the connection parameters
	return db.DatabaseConnectionParams{
		Host:     host,
		Port:     portInt,
		Username: user,
		Password: pass,
		Database: database,
	}, nil
}

// processLock attempts to acquire a process lock to prevent multiple instances of the application running
func processLock() (*lock.Lock, error) {
	lockfile, err := lock.NewLock()
	if err == lock.ErrAlreadyLocked {
		return nil, ErrAlreadyRunning
	} else if err != nil {
		return nil, ErrFailedProcessLock
	}

	return lockfile, nil
}
