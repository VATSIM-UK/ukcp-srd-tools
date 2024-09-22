package file

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/note"
)

func TestMapNote_Success(t *testing.T) {
	rows := [][]string{
		{"Note 1"},
		{"This is the first line of the note."},
		{"This is the second line of the note."},
	}

	expectedNote := note.NewNote(1, "This is the first line of the note.\nThis is the second line of the note.")

	result, err := mapNote(rows)

	require.NoError(t, err)
	require.Equal(t, expectedNote, result)
}

func TestMapNote_InvalidNoteID(t *testing.T) {
	rows := [][]string{
		{"Invalid Note ID"},
		{"This is the first line of the note."},
	}

	result, err := mapNote(rows)

	require.Error(t, err)
	require.Nil(t, result)
}

func TestMapNote_InvalidNoteIDConversion(t *testing.T) {
	rows := [][]string{
		{"Note abc"},
		{"This is the first line of the note."},
	}

	result, err := mapNote(rows)

	require.Error(t, err)
	require.Nil(t, result)
}

func TestMapNote_EmptyRows(t *testing.T) {
	rows := [][]string{}

	result, err := mapNote(rows)

	require.Error(t, err)
	require.Nil(t, result)
}

func TestMapNote_MissingNoteText(t *testing.T) {
	rows := [][]string{
		{"Note 1"},
	}

	result, err := mapNote(rows)

	require.Error(t, err)
	require.Nil(t, result)
}
