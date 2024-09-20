package parse

import (
	"iter"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/file"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/route"
)

type srdFile interface {
	Routes() iter.Seq2[*route.Route, error]
	Notes() iter.Seq2[*note.Note, error]
	Stats() file.SrdStats
}

// ParseSrd parses the SRD file and returns a summary of the parsing
func ParseSrd(file srdFile) file.SrdStats {
	for range file.Routes() {
	}

	for range file.Notes() {
	}

	return file.Stats()
}
