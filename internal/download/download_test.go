package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/airac"
)

type testServer struct {
	statusCode int
	body       string
	callCount  int
	server     *httptest.Server
}

func (t *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.callCount++
	w.WriteHeader(t.statusCode)
	w.Write([]byte(t.body))
}

func getTestServer(statusCode int, body string) *testServer {
	testServer := &testServer{statusCode: statusCode, body: body}
	testServer.server = httptest.NewServer(testServer)

	return testServer
}

func TestDownloader_FirstTimeDownload(t *testing.T) {
	require := require.New(t)
	tempDir := t.TempDir()

	ts := getTestServer(http.StatusOK, "test body")
	defer ts.server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	airac := airac.NewAirac(nil)
	cycle := airac.CurrentCycle()
	d, err := NewSrdDownloader(cycle, &mockLoadedAirac{ident: ""}, tempDir, ts.server.URL)
	require.NoError(err)

	// Download the file
	err = d.Download(ctx, false)
	require.NoError(err)

	// Check the downloaded file
	downloadedFile, err := os.Open(tempDir + "/ukcp-srd-import-loaded-download.xlsx")
	require.NoError(err)

	// Check the file content
	buf := make([]byte, 1024)
	n, err := downloadedFile.Read(buf)
	require.NoError(err)
	require.Equal("test body", string(buf[:n]))
}

func TestDownloader_SubsequentDownloads(t *testing.T) {
	require := require.New(t)
	tempDir := t.TempDir()

	ts := getTestServer(http.StatusOK, "test body")
	defer ts.server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	airac := airac.NewAirac(nil)
	nextCycle := airac.NextCycle()

	// Create the previous file
	previousFile, err := os.Create(tempDir + "/ukcp-srd-import-loaded-download.xlsx")
	require.NoError(err)
	previousFile.Write([]byte("previous file"))
	previousFile.Close()

	d, err := NewSrdDownloader(nextCycle, &mockLoadedAirac{ident: "ABCD"}, tempDir, ts.server.URL)
	require.NoError(err)

	// Download the file
	err = d.Download(ctx, false)
	require.NoError(err)

	// Check the downloaded file
	downloadedFile, err := os.Open(tempDir + "/ukcp-srd-import-loaded-download.xlsx")
	require.NoError(err)

	// Check the file content
	buf := make([]byte, 1024)
	n, err := downloadedFile.Read(buf)
	require.NoError(err)
	require.Equal("test body", string(buf[:n]))
}

func TestDownloader_AlreadyUpToDate(t *testing.T) {
	require := require.New(t)
	tempDir := t.TempDir()

	ts := getTestServer(http.StatusOK, "test body")
	defer ts.server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	airac := airac.NewAirac(nil)
	cycle := airac.CurrentCycle()

	// Create the version file
	versionFile, err := os.Create(tempDir + "/ukcp-srd-import-loaded-cycle")
	require.NoError(err)
	versionFile.Write([]byte(cycle.Ident))
	versionFile.Close()

	d, err := NewSrdDownloader(cycle, &mockLoadedAirac{ident: cycle.Ident}, tempDir, ts.server.URL)
	require.NoError(err)

	// Download the file
	err = d.Download(ctx, false)
	require.ErrorIs(err, ErrUpToDate)

	require.Equal(0, ts.callCount)
}

func TestDownloader_ErrorDownloading(t *testing.T) {
	require := require.New(t)
	tempDir := t.TempDir()

	ts := getTestServer(http.StatusInternalServerError, "test body")
	defer ts.server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	airac := airac.NewAirac(nil)
	cycle := airac.CurrentCycle()
	d, err := NewSrdDownloader(cycle, &mockLoadedAirac{ident: ""}, tempDir, ts.server.URL)
	require.NoError(err)

	// Download the file
	err = d.Download(ctx, false)
	require.Error(err)
}

type mockLoadedAirac struct {
	ident string
}

func (m *mockLoadedAirac) Ident() string {
	return m.ident
}

func (m *mockLoadedAirac) Is(ident string) bool {
	return m.ident == ident
}
