package parse

import (
	"iter"

	"github.com/rs/zerolog/log"

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
	for _, err := range file.Routes() {
		if err != nil {
			log.Error().Msgf("Error parsing route: %v", err)
		}
	}

	for _, err := range file.Notes() {
		if err != nil {
			log.Error().Msgf("Error parsing note: %v", err)
		}
	}

	return file.Stats()
}
