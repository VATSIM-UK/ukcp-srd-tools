package parse

import (
	"iter"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/route"
)

type srdFile interface {
	Routes() iter.Seq2[*route.Route, error]
	Notes() iter.Seq2[*note.Note, error]
}

type ParseSummary struct {
	RouteCount      int
	RouteErrorCount int
	NoteCount       int
	NoteErrorCount  int
}

// ParseSrd parses the SRD file and returns a summary of the parsing
func ParseSrd(file srdFile) ParseSummary {
	routeCount := 0
	routeErrorCount := 0
	noteCount := 0
	noteErrorCount := 0

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

	return ParseSummary{
		RouteCount:      routeCount,
		RouteErrorCount: routeErrorCount,
		NoteCount:       noteCount,
		NoteErrorCount:  noteErrorCount,
	}
}
