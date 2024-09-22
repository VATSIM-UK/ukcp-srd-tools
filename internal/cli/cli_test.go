package cli_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/airac"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/cli"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/db"
	"github.com/VATSIM-UK/ukcp-srd-tools/test/logging"
)

func TestRun_Airac(t *testing.T) {
	require := require.New(t)

	// Run the CLI test
	test := runCliTest(t, []string{"cmd", "airac"})
	require.NoError(test.testError)

	currentAirac := airac.NewAirac(nil)
	currentCycle := currentAirac.CurrentCycle()
	nextCycle := currentAirac.NextCycle()

	airacString := func(cycleName string, cycle *airac.AiracCycle) string {
		return fmt.Sprintf(
			"%s AIRAC cycle is %v (%v - %v)",
			cycleName,
			cycle.Ident,
			cycle.Start.Format("2006-01-02"),
			cycle.End.Format("2006-01-02"),
		)
	}

	// Check the logs
	test.logRecorder.AssertHasString(require, airacString("Current", currentCycle))
	test.logRecorder.AssertHasString(require, airacString("Next", nextCycle))
}

func TestRun_Parse(t *testing.T) {
	testFolderAbsPath, _ := filepath.Abs("../../test/data")

	tests := []struct {
		name                string
		filename            string
		expectedErr         error
		expectedLogMessages []string
	}{
		{
			name:        "missing file",
			filename:    "missing.xlsx",
			expectedErr: fmt.Errorf("failed to open excel extended file: open %s/missing.xlsx: no such file or directory", testFolderAbsPath),
			expectedLogMessages: []string{
				"failed to open excel extended file",
			},
		},
		{
			"successful parse",
			"simple1.xlsx",
			nil,
			[]string{
				"processed 3 routes with 0 errors",
				"processed 3 notes with 0 errors",
			},
		},
		{
			"successful parse xls",
			"simple1.xls",
			nil,
			[]string{
				"processed 3 routes with 0 errors",
				"processed 3 notes with 0 errors",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Run the CLI test
			test := runCliTest(t, []string{"cmd", "parse", testDataFile(tt.filename)})

			if tt.expectedErr != nil {
				require.Error(test.testError)
				require.Equal(tt.expectedErr, test.testError)
			} else {
				require.NoError(test.testError)
			}

			// Check the logs
			for _, msg := range tt.expectedLogMessages {
				test.logRecorder.AssertHasString(require, msg)
			}
		})
	}
}

func TestRun_Loaded(t *testing.T) {
	test := getCliTest(t, []string{"cmd", "loaded"})
	require := require.New(t)

	// Create the version
	loaded, err := airac.NewLoadedAirac(test.tempDir)
	require.NoError(err)

	// Save the version
	airacManager := airac.NewAirac(nil)
	savedCycle, _ := airacManager.CycleFromIdent("2403")
	err = loaded.Set(savedCycle)
	require.NoError(err)
	require.NoError(loaded.Close())

	// Now run the CLI test
	test.testError = cli.Run(test.tempDir)
	require.NoError(test.testError)

	// Check the logs
	test.logRecorder.AssertHasString(require, "Loaded AIRAC cycle is 2403 (2024-03-21 - 2024-04-18)")
}

func TestRun_LoadedNoCycleLoaded(t *testing.T) {
	require := require.New(t)

	// Run the CLI test
	test := runCliTest(t, []string{"cmd", "loaded"})
	require.NoError(test.testError)

	// Check the logs
	test.logRecorder.AssertHasString(require, "No AIRAC cycle loaded")
}

type importTest struct {
	name                string
	filename            string
	testEnvFile         string
	envFileContent      map[string]string
	expectedErr         error
	expectedLogMessages []string
}

func TestRun_ImportErrors(t *testing.T) {
	tests := []importTest{
		{
			"missing file",
			"missing.xlsx",
			"",
			nil,
			fmt.Errorf("failed to open excel extended file: open %s: no such file or directory", testDataFile("missing.xlsx")),
			[]string{
				"failed to open excel extended file",
			},
		},
		{
			"non excel file",
			"invalid.txt",
			"",
			nil,
			cli.ErrUnknownFileExtension,
			[]string{
				"unknown file extension .txt",
			},
		},
		{
			"missing env file",
			"simple1.xlsx",
			"",
			nil,
			cli.ErrCannotLoadDotenv,
			[]string{
				"failed to load environment file",
			},
		},
		{
			"env file missing host",
			"simple1.xlsx",
			"missing-host.env",
			map[string]string{
				"DB_PORT":     "5432",
				"DB_USERNAME": "user",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
			},
			cli.ErrMissingHost,
			[]string{
				"missing database host",
			},
		},
		{
			"env file missing port",
			"simple1.xlsx",
			"missing-port.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_USERNAME": "user",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
			},
			cli.ErrMissingPort,
			[]string{
				"missing database port",
			},
		},
		{
			"env file invalid port",
			"simple1.xlsx",
			"missing-port.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_USERNAME": "user",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
				"DB_PORT":     "invalid",
			},
			cli.ErrPortInvalid,
			[]string{
				"invalid database port",
			},
		},
		{
			"env file missing username",
			"simple1.xlsx",
			"missing-username.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
			},
			cli.ErrMissingUser,
			[]string{
				"missing database user",
			},
		},
		{
			"env file missing database",
			"simple1.xlsx",
			"missing-database.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USERNAME": "user",
				"DB_PASSWORD": "passwd",
			},
			cli.ErrMissingDatabase,
			[]string{
				"missing database name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			testDir := t.TempDir()
			envFilePath := fmt.Sprintf("%s/%s", testDir, tt.testEnvFile)

			// Get the cliTest struct
			test := getCliTestWithTempDir([]string{"cmd", "import", "2404", testDataFile(tt.filename), "--env-path", envFilePath}, testDir)

			// If we have an env file, create it
			if tt.testEnvFile != "" {
				err := godotenv.Write(
					tt.envFileContent,
					envFilePath,
				)
				require.NoError(err)
			}

			// Run the CLI test
			test.testError = cli.Run(testDir)

			if tt.expectedErr != nil {
				require.Error(test.testError)
				require.Equal(tt.expectedErr, test.testError)
			} else {
				require.NoError(test.testError)
			}

			// Check the logs
			for _, msg := range tt.expectedLogMessages {
				test.logRecorder.AssertHasString(require, msg)
			}

			// Reset the dotenv
			resetEnv()
		})
	}

}

func TestRun_ImportSuccess(t *testing.T) {
	tests := []struct {
		name                string
		filename            string
		expectedLogMessages []string
		extraArgs           []string
		setupFunc           func(*cliTest, *require.Assertions)
	}{
		{
			name:     "import xlsx",
			filename: "simple1.xlsx",
			expectedLogMessages: []string{
				"processed 3 routes with 0 errors",
				"processed 3 notes with 0 errors",
				"imported SRD for cycle",
			},
		},
		{
			name:     "import xls",
			filename: "simple1.xls",
			expectedLogMessages: []string{
				"processed 3 routes with 0 errors",
				"processed 3 notes with 0 errors",
				"imported SRD for cycle",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			testDir := t.TempDir()
			envFilePath := fmt.Sprintf("%s/%s", testDir, "test.env")

			// If there's a setup function, run it
			if tt.setupFunc != nil {
				test := getCliTestWithTempDir(tt.extraArgs, testDir)
				tt.setupFunc(test, require)
			}

			// Get the cliTest struct
			test := getCliTestWithTempDir([]string{"cmd", "import", "2404", testDataFile(tt.filename), "--env-path", envFilePath}, testDir)

			// Create the database
			ctx := context.Background()
			mysqlContainer, err := getMysqlContainer(ctx, t)
			require.NoError(err)

			containerHost, err := mysqlContainer.container.Host(ctx)
			require.NoError(err)

			containerPort, err := mysqlContainer.container.MappedPort(ctx, "3306")
			require.NoError(err)

			// Now write the env file with our database credentaisl
			err = godotenv.Write(
				map[string]string{
					"DB_HOST":     containerHost,
					"DB_PORT":     containerPort.Port(),
					"DB_USERNAME": TestUsername,
					"DB_DATABASE": TestDatabase,
					"DB_PASSWORD": TestPassword,
				},
				envFilePath,
			)

			// Run the CLI test
			require.NoError(cli.Run(testDir))

			// Check the logs
			for _, msg := range tt.expectedLogMessages {
				test.logRecorder.AssertHasString(require, msg)
			}

			// Check the saved version file
			loaded, err := airac.NewLoadedAirac(testDir)
			require.NoError(err)
			require.Equal("2404", loaded.Ident())

			// Close the loaded file
			require.NoError(loaded.Close())

			// Now check the database - just a simple check we have the correct number of routes and notes
			db, err := db.NewDatabase(
				db.DatabaseConnectionParams{
					Host:     containerHost,
					Port:     containerPort.Int(),
					Username: TestUsername,
					Password: TestPassword,
					Database: TestDatabase,
				},
			)
			require.NoError(err)

			// Check the number of routes
			var routeCount int
			err = db.Handle().QueryRow("SELECT COUNT(*) FROM srd_routes").Scan(&routeCount)
			require.NoError(err)
			require.Equal(3, routeCount)

			// Check the number of notes
			var noteCount int
			err = db.Handle().QueryRow("SELECT COUNT(*) FROM srd_notes").Scan(&noteCount)
			require.NoError(err)
			require.Equal(3, noteCount)

			// Terminate the database
			mysqlContainer.terminateFunc()

			// Reset the dotenv
			resetEnv()
		})
	}
}

type downloadTest struct {
	name                string
	filename            string
	testEnvFile         string
	envFileContent      map[string]string
	expectedErr         error
	expectedLogMessages []string
	responseCode        int
}

func TestDownload_Errors(t *testing.T) {
	tests := []downloadTest{
		{
			"missing env file",
			"simple1.xlsx",
			"",
			nil,
			cli.ErrCannotLoadDotenv,
			[]string{
				"failed to load environment file",
			},
			200,
		},
		{
			"env file missing host",
			"simple1.xlsx",
			"missing-host.env",
			map[string]string{
				"DB_PORT":     "5432",
				"DB_USERNAME": "user",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
			},
			cli.ErrMissingHost,
			[]string{
				"missing database host",
			},
			200,
		},
		{
			"env file missing port",
			"simple1.xlsx",
			"missing-port.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_USERNAME": "user",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
			},
			cli.ErrMissingPort,
			[]string{
				"missing database port",
			},
			200,
		},
		{
			"env file invalid port",
			"simple1.xlsx",
			"missing-port.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_USERNAME": "user",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
				"DB_PORT":     "invalid",
			},
			cli.ErrPortInvalid,
			[]string{
				"invalid database port",
			},
			200,
		},
		{
			"env file missing username",
			"simple1.xlsx",
			"missing-username.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
			},
			cli.ErrMissingUser,
			[]string{
				"missing database user",
			},
			200,
		},
		{
			"env file missing database",
			"simple1.xlsx",
			"missing-database.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USERNAME": "user",
				"DB_PASSWORD": "passwd",
			},
			cli.ErrMissingDatabase,
			[]string{
				"missing database name",
			},
			200,
		},
		{
			"download error",
			"simple1.xlsx",
			"test.env",
			map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USERNAME": "user",
				"DB_DATABASE": "name",
				"DB_PASSWORD": "passwd",
			},
			errors.New("unable to download SRD, status code was 500 Internal Server Error"),
			[]string{
				"unable to download SRD, status code was 500",
			},
			500,
		},
		{
			"invalid file",
			"invalid.txt",
			"",
			nil,
			errors.New("failed to open excel extended file: zip: not a valid zip file"),
			[]string{
				"failed to load excel file",
			},
			200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			testDir := t.TempDir()
			envFilePath := fmt.Sprintf("%s/%s", testDir, tt.testEnvFile)
			fileName := testDataFile(tt.filename)

			// Start up a test server
			ts := getTestServer(tt.responseCode, fileName)
			defer ts.server.Close()

			// Get the cliTest struct
			test := getCliTestWithTempDir([]string{"cmd", "download", "--env-path", envFilePath, "--url", ts.server.URL}, testDir)

			// If we have an env file, create it
			if tt.testEnvFile != "" {
				err := godotenv.Write(
					tt.envFileContent,
					envFilePath,
				)
				require.NoError(err)
			}

			// Run the CLI test
			test.testError = cli.Run(testDir)

			if tt.expectedErr != nil {
				require.Error(test.testError)
				require.Equal(tt.expectedErr, test.testError)
			} else {
				require.NoError(test.testError)
			}

			// Check the logs
			for _, msg := range tt.expectedLogMessages {
				test.logRecorder.AssertHasString(require, msg)
			}

			// Reset the dotenv
			resetEnv()
		})
	}

}

type downloadSuccessTest struct {
	name                string
	fileName            string
	expectedLogMessages []string
	extraArgs           []string
	setupFunc           func(*cliTest, *require.Assertions)
}

func TestDownloadSuccess(t *testing.T) {
	tests := []downloadSuccessTest{
		{
			name:     "download xlsx",
			fileName: "simple1.xlsx",
			expectedLogMessages: []string{
				"processed 3 routes with 0 errors",
				"processed 3 notes with 0 errors",
				"imported SRD for cycle",
			},
			extraArgs: nil,
			setupFunc: nil,
		},
		{
			name:     "download xlsx with cycle",
			fileName: "simple1.xlsx",
			expectedLogMessages: []string{
				"processed 3 routes with 0 errors",
				"processed 3 notes with 0 errors",
				"imported SRD for cycle",
			},
			extraArgs: []string{"--cycle", "2404"},
		},
		{
			name:     "download xlsx with cycle force",
			fileName: "simple1.xlsx",
			expectedLogMessages: []string{
				"processed 3 routes with 0 errors",
				"processed 3 notes with 0 errors",
				"imported SRD for cycle",
			},
			extraArgs: []string{"--cycle", "2404", "--force"},
			setupFunc: func(test *cliTest, require *require.Assertions) {
				// Create the version
				loaded, err := airac.NewLoadedAirac(test.tempDir)
				require.NoError(err)

				// Save the version
				airacManager := airac.NewAirac(nil)
				savedCycle, _ := airacManager.CycleFromIdent("2404")
				err = loaded.Set(savedCycle)
				require.NoError(err)
				require.NoError(loaded.Close())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			testDir := t.TempDir()
			envFilePath := fmt.Sprintf("%s/%s", testDir, "test.env")

			// Download file path
			fileName := testDataFile(tt.fileName)

			// Get the cliTest struct
			// Start up a test server
			ts := getTestServer(200, fileName)
			defer ts.server.Close()

			args := []string{"cmd", "download", "--env-path", envFilePath, "--url", ts.server.URL}
			fullArgs := append(args, tt.extraArgs...)

			// If there's not an arg called "cycle" in the extra args, find out what the current cycle is
			cycle := ""
			if !slices.Contains(fullArgs, "--cycle") {
				airac := airac.NewAirac(nil)
				currentCycle := airac.CurrentCycle()
				cycle = currentCycle.Ident
			} else {
				// Get index of cycle
				index := slices.Index(fullArgs, "--cycle")

				// Get the cycle from the next index
				cycle = fullArgs[index+1]
			}

			// If there's a setup function, run it
			if tt.setupFunc != nil {
				test := getCliTestWithTempDir(fullArgs, testDir)
				tt.setupFunc(test, require)
			}

			// Get the cliTest struct
			test := getCliTestWithTempDir(fullArgs, testDir)

			// Create the database
			ctx := context.Background()
			mysqlContainer, err := getMysqlContainer(ctx, t)
			require.NoError(err)

			containerHost, err := mysqlContainer.container.Host(ctx)
			require.NoError(err)

			containerPort, err := mysqlContainer.container.MappedPort(ctx, "3306")
			require.NoError(err)

			// Now write the env file with our database credentaisl
			err = godotenv.Write(
				map[string]string{
					"DB_HOST":     containerHost,
					"DB_PORT":     containerPort.Port(),
					"DB_USERNAME": TestUsername,
					"DB_DATABASE": TestDatabase,
					"DB_PASSWORD": TestPassword,
				},
				envFilePath,
			)

			// Run the CLI test
			require.NoError(cli.Run(testDir))

			// Check the logs
			for _, msg := range tt.expectedLogMessages {
				test.logRecorder.AssertHasString(require, msg)
			}

			// Check the saved version file
			loaded, err := airac.NewLoadedAirac(testDir)
			require.NoError(err)
			require.Equal(cycle, loaded.Ident())

			// Close the loaded file
			require.NoError(loaded.Close())

			// Now check the database - just a simple check we have the correct number of routes and notes
			db, err := db.NewDatabase(
				db.DatabaseConnectionParams{
					Host:     containerHost,
					Port:     containerPort.Int(),
					Username: TestUsername,
					Password: TestPassword,
					Database: TestDatabase,
				},
			)
			require.NoError(err)

			// Check the number of routes
			var routeCount int
			err = db.Handle().QueryRow("SELECT COUNT(*) FROM srd_routes").Scan(&routeCount)
			require.NoError(err)
			require.Equal(3, routeCount)

			// Check the number of notes
			var noteCount int
			err = db.Handle().QueryRow("SELECT COUNT(*) FROM srd_notes").Scan(&noteCount)
			require.NoError(err)
			require.Equal(3, noteCount)

			// Terminate the database
			mysqlContainer.terminateFunc()

			// Reset the dotenv
			resetEnv()
		})
	}
}

type cliTest struct {
	tempDir     string
	logRecorder *logging.LogRecorder
	tearDown    func()
	testError   error
}

func resetEnv() {
	absPath, _ := filepath.Abs("../../test/env//blankenv")
	godotenv.Overload(absPath)
}

func testDataFile(filename string) string {
	relPath := fmt.Sprintf("../../test/data/%s", filename)

	absPath, _ := filepath.Abs(relPath)
	return absPath
}

// runCliTest runs a test for the CLI
// It returns a struct that can be used to check the logs and clean up
func runCliTest(t *testing.T, args []string) *cliTest {
	test := getCliTest(t, args)

	// Run the CLI
	test.testError = cli.Run(test.tempDir)

	return test
}

func getCliTest(t *testing.T, args []string) *cliTest {
	// Create a temporary dir for the test
	return getCliTestWithTempDir(args, t.TempDir())
}

func getCliTestWithTempDir(args []string, tempDir string) *cliTest {
	// Hijack the logs
	logRecorder, teardownLogs := logging.HijackLogs()

	// Hijack the os.Args to pretend we're running on the CLI
	tearDownArgs := hijackArgs(args)

	return &cliTest{
		tempDir:     tempDir,
		logRecorder: logRecorder,
		tearDown: func() {
			tearDownArgs()
			teardownLogs()
		},
	}
}

// hijackArgs hijacks the os.Args to pretend we're running on the CLI
// it returns a function that can be used to restore the original os.Args
func hijackArgs(args []string) func() {
	originalArgs := os.Args

	// Add the --testing flag to the args
	args = append(args, "--testing")

	os.Args = args
	return func() {
		os.Args = originalArgs
	}
}

type mysqlContainer struct {
	container     *mysql.MySQLContainer
	terminateFunc func()
}

const (
	TestDatabase = "uk_plugin"
	TestUsername = "uk_plugin_test"
	TestPassword = "test_password_123"
)

type testingT interface {
	Fatalf(format string, args ...interface{})
}

func getMysqlContainer(ctx context.Context, t testingT) (*mysqlContainer, error) {
	container, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithDatabase(TestDatabase),
		mysql.WithUsername(TestUsername),
		mysql.WithPassword(TestPassword),
		mysql.WithDefaultCredentials(),
		mysql.WithScripts("../../test/db/db_setup.sql"),
	)
	if err != nil {
		return nil, err
	}
	terminateFunc := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}

	return &mysqlContainer{container, terminateFunc}, nil
}

type testServer struct {
	statusCode      int
	filePathToServe string
	callCount       int
	server          *httptest.Server
}

func (t *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.callCount++

	// If the status code isn't 200, return the status code
	if t.statusCode != http.StatusOK {
		w.WriteHeader(t.statusCode)
		return
	}

	// Serve the file
	file, err := os.Open(t.filePathToServe)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(t.statusCode)
	defer file.Close()

	_, err = file.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Copy the file to the response
	_, err = io.Copy(w, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func getTestServer(statusCode int, pathToServe string) *testServer {
	testServer := &testServer{statusCode: statusCode, filePathToServe: pathToServe}
	testServer.server = httptest.NewServer(testServer)

	return testServer
}
