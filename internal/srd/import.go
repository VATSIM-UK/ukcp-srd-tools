package srd

import (
	"context"
	"iter"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/db"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/route"
)

const InsertBatchSize = 5000
const InterBatchWait = 1 * time.Second

type srdFile interface {
	Routes() iter.Seq2[*route.Route, error]
	Notes() iter.Seq2[*note.Note, error]
}

type Import struct {
	db   *db.Database
	file srdFile

	// Map of note IDs to route IDs
	routeNotes map[uint64][]uint64
}

func NewImport(file srdFile, db *db.Database) *Import {
	return &Import{db, file, make(map[uint64][]uint64)}
}

func (i *Import) Import(ctx context.Context) error {
	i.routeNotes = make(map[uint64][]uint64)

	return i.db.Transaction(func(tx *db.Transaction) error {
		err := i.deleteCurrentData(ctx, tx)
		if err != nil {
			return err
		}

		err = i.insertNotes(ctx, tx)
		if err != nil {
			return err
		}

		err = i.insertRoutes(ctx, tx)
		if err != nil {
			return err
		}

		err = i.insertRouteNoteLinks(ctx, tx)
		if err != nil {
			return err
		}

		return nil
	})
}

// insertNotes inserts the notes into the database, in batches of InsertBatchSize
func (i *Import) insertNotes(ctx context.Context, tx *db.Transaction) error {
	notes := make([]*note.Note, 0)
	for srdNote, err := range i.file.Notes() {
		if err != nil {
			log.Warn().Msgf("invalid note detected: %v", err)
			continue
		}

		notes = append(notes, srdNote)

		// Add the note to our map of note
		// IDs to route IDs
		i.routeNotes[srdNote.ID()] = make([]uint64, 0)

		// Insert the notes in batches
		if len(notes) >= InsertBatchSize {
			err := i.insertNoteBatch(ctx, tx, notes)
			if err != nil {
				return err
			}

			notes = make([]*note.Note, 0)
		}
	}

	// Insert any remaining notes
	if len(notes) > 0 {
		err := i.insertNoteBatch(ctx, tx, notes)
		if err != nil {
			return err
		}
	}

	return nil
}

// insertNoteBatch inserts a batch of notes into the database and then waits for a bit
func (i *Import) insertNoteBatch(ctx context.Context, tx *db.Transaction, batch []*note.Note) error {
	if err := tx.InsertNoteBatch(ctx, batch); err != nil {
		return err
	}

	// Wait for a bit to avoid overwhelming the database
	i.interBatchWait()

	return nil
}

func (i *Import) insertRoutes(ctx context.Context, tx *db.Transaction) error {
	routes := make([]*route.Route, 0)
	for srdRoute, err := range i.file.Routes() {
		if err != nil {
			log.Warn().Msgf("invalid route detected: %v", err)
			continue
		}

		routes = append(routes, srdRoute)

		// Insert the routes in batches
		if len(routes) >= InsertBatchSize {
			err := i.insertRouteBatch(ctx, tx, routes)
			if err != nil {
				return err
			}
			routes = make([]*route.Route, 0)

			// Wait for a bit to avoid overwhelming the database
			i.interBatchWait()
		}
	}

	// Insert any remaining routes
	if len(routes) > 0 {
		return i.insertRouteBatch(ctx, tx, routes)
	}

	return nil
}

func (i *Import) insertRouteBatch(ctx context.Context, tx *db.Transaction, batch []*route.Route) error {
	firstInsertId, err := tx.InsertRouteBatch(ctx, batch)
	if err != nil {
		return err
	}

	// Now go through each route in the batch, and add the note IDs to the routeNotes map - the note ID is the firstInsertId + the index
	for idx, route := range batch {
		routeID := uint64(firstInsertId) + uint64(idx)

		// Add the note IDs to the routeNotes map
		for _, noteID := range route.NoteIDs() {
			// If there's no entry for this note ID, skip it
			if _, ok := i.routeNotes[noteID]; !ok {
				continue
			}

			i.routeNotes[noteID] = append(i.routeNotes[noteID], routeID)
		}
	}

	// Wait for a bit to avoid overwhelming the database
	i.interBatchWait()

	return nil
}

// insertNoteRouteLinks inserts the note-route links into the database in batches of InsertBatchSize
func (i *Import) insertRouteNoteLinks(ctx context.Context, tx *db.Transaction) error {
	links := make([]*db.NoteRouteLink, 0)
	for noteID, routeIDs := range i.routeNotes {
		for _, routeID := range routeIDs {
			links = append(links, &db.NoteRouteLink{NoteID: noteID, RouteID: routeID})

			// Insert the links in batches
			if len(links) >= InsertBatchSize {
				err := i.insertRouteNoteBatch(ctx, tx, links)
				if err != nil {
					return err
				}
				links = make([]*db.NoteRouteLink, 0)
			}
		}
	}

	// Insert any remaining links
	if len(links) > 0 {
		return i.insertRouteNoteBatch(ctx, tx, links)
	}

	return nil
}

// insertRouteNoteBatch inserts a batch of note-route links into the database and then waits for a bit
func (i *Import) insertRouteNoteBatch(ctx context.Context, tx *db.Transaction, batch []*db.NoteRouteLink) error {
	if err := tx.InsertNoteRouteLinkBatch(ctx, batch); err != nil {
		return err
	}

	// Wait for a bit to avoid overwhelming the database
	i.interBatchWait()

	return nil
}

func (i *Import) deleteCurrentData(ctx context.Context, tx *db.Transaction) error {
	err := tx.DeleteAllRoutes(ctx)
	if err != nil {
		return err
	}

	return tx.DeleteAllNotes(ctx)
}

// If we import too quickly, we might overwhelm the database, so we should wait between batches
func (i *Import) interBatchWait() {
	time.Sleep(InterBatchWait)
}
