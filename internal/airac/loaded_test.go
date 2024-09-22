package airac

import (
	"bufio"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLoadedAirac(t *testing.T) {
	testDir := t.TempDir()

	// Create a temporary file with a test cycle ident
	filePath := filePath(testDir, "ukcp-srd-import-loaded-cycle")
	file, err := os.Create(filePath)
	require.NoError(t, err, "Failed to create temp file")
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString("test_cycle")
	require.NoError(t, err, "Failed to write to temp file")
	writer.Flush()
	file.Close()

	loadedAirac, err := NewLoadedAirac(testDir)
	require.NoError(t, err, "NewLoadedAirac returned an error")
	defer loadedAirac.Close()

	require.Equal(t, "test_cycle", loadedAirac.Ident(), "Expected ident 'test_cycle'")
}

func TestLoadedAirac_Set(t *testing.T) {
	testDir := t.TempDir()

	loadedAirac, err := NewLoadedAirac(testDir)
	require.NoError(t, err, "NewLoadedAirac returned an error")
	defer loadedAirac.Close()

	newCycle := NewAirac(nil).CurrentCycle()
	err = loadedAirac.Set(newCycle)
	require.NoError(t, err, "Set returned an error")

	// Reopen the file to check the new content
	loadedAirac.Close()
	loadedAirac, err = NewLoadedAirac(testDir)
	require.NoError(t, err, "NewLoadedAirac returned an error")
	defer loadedAirac.Close()

	require.Equal(t, newCycle.Ident, loadedAirac.Ident())
}

func TestLoadedAirac_Close(t *testing.T) {
	testDir := t.TempDir()

	loadedAirac, err := NewLoadedAirac(testDir)
	require.NoError(t, err, "NewLoadedAirac returned an error")

	err = loadedAirac.Close()
	require.NoError(t, err, "Close returned an error")

	// Try closing again to see if it handles already closed file
	err = loadedAirac.Close()
	require.Error(t, err, "Expected an error when closing already closed file, got nil")
}
func TestLoadedAirac_Is(t *testing.T) {
	testDir := t.TempDir()

	// Create a temporary file with a test cycle ident
	filePath := filePath(testDir, "ukcp-srd-import-loaded-cycle")
	file, err := os.Create(filePath)
	require.NoError(t, err, "Failed to create temp file")
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString("test_cycle")
	require.NoError(t, err, "Failed to write to temp file")
	writer.Flush()
	file.Close()

	loadedAirac, err := NewLoadedAirac(testDir)
	require.NoError(t, err, "NewLoadedAirac returned an error")
	defer loadedAirac.Close()

	require.True(t, loadedAirac.Is("test_cycle"), "Expected Is to return true for ident 'test_cycle'")
	require.False(t, loadedAirac.Is("wrong_cycle"), "Expected Is to return false for ident 'wrong_cycle'")
}
