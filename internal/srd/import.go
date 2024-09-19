package srd

import (
	"iter"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/route"
)

type srdFile interface {
	Routes() iter.Seq2[*route.Route, error]
	Notes() iter.Seq2[*note.Note, error]
}

func Import(file srdFile) error {
	// First of all, iterate the notes
	for _, err := range file.Notes() {
		if err != nil {
			return err
		}

		//fmt.Println(note.ID())
		// TODO: Write the note to the database
	}

	// Iterate the routes
	for _, err := range file.Routes() {
		if err != nil {
			return err
		}

		//fmt.Println(route.ADEPOrEntry())

		// TODO: Write the route to the database
	}

	return nil
}
