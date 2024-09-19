package note

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNote(t *testing.T) {
	id := uint64(1)
	text := "This is a test note"
	note := NewNote(id, text)

	require.Equal(t, id, note.ID(), "expected ID %d, got %d", id, note.ID())
	require.Equal(t, text, note.Text(), "expected text %q, got %q", text, note.Text())
}

func TestNoteToJSON(t *testing.T) {
	id := uint64(1)
	text := "This is a test note"
	note := NewNote(id, text)
	expectedJSON := `{"id": 1, "text": "This is a test note"}`

	require.Equal(t, expectedJSON, note.ToJSON(), "expected JSON %q, got %q", expectedJSON, note.ToJSON())
}
