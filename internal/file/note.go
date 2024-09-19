package file

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/note"
)

// mapNote creates a note from the raw data
// The first row is always in the format "Note <number>"
// The subsequent rows are arbitrary text
func mapNote(rows [][]string) (*note.Note, error) {
	// Get the note id using the note id RegExp
	noteIDRegexMatch := NewRowRegxp.FindStringSubmatch(rows[0][0])

	// If the note id is not found, return an error
	if len(noteIDRegexMatch) != 2 {
		return nil, fmt.Errorf("expected note id, got %v", rows[0][0])
	}

	// Convert the note id to a uint64
	noteID, err := strconv.ParseUint(noteIDRegexMatch[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse note id on note %v, error: %v", rows[0], err)
	}

	// Join the subsequent rows to create the note text
	noteText := ""
	if len(rows) < 2 {
		return nil, fmt.Errorf("expected note text, got %v", rows)
	}

	for _, row := range rows[1:] {
		noteText += row[0] + "\n"
	}

	// Create the note
	return note.NewNote(noteID, strings.TrimSpace(noteText)), nil
}
