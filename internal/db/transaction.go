package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/route"
)

type NoteRouteLink struct {
	NoteID  uint64
	RouteID uint64
}

type Transaction struct {
	tx *sql.Tx
}

// DeleteAllRoutes deletes all routes from the database
func (t *Transaction) DeleteAllRoutes(ctx context.Context) error {
	_, err := t.tx.ExecContext(ctx, "DELETE FROM srd_routes")
	return err
}

// DeleteAllNotes deletes all notes from the database
func (t *Transaction) DeleteAllNotes(ctx context.Context) error {
	_, err := t.tx.ExecContext(ctx, "DELETE FROM srd_notes")
	return err
}

// InsertNoteBatch inserts a batch of notes into the database
func (t *Transaction) InsertNoteBatch(ctx context.Context, notes []*note.Note) error {
	queryString := "INSERT INTO srd_notes (id, note_text) VALUES "

	// Build the query string
	queryArgs := make([]interface{}, 0)
	for i, note := range notes {
		if i > 0 {
			queryString += ", "
		}
		queryString += fmt.Sprintf("(?, ?)")
		queryArgs = append(queryArgs, note.ID(), note.Text())
	}

	// Prepare the query
	stmt, err := t.tx.PrepareContext(ctx, queryString)
	if err != nil {
		return err
	}

	// Execute the query
	_, err = stmt.ExecContext(ctx, queryArgs...)

	return err
}

// InsertRouteBatch inserts a batch of routes into the database
// It returns the insert ID of the first route in the batch
func (t *Transaction) InsertRouteBatch(ctx context.Context, routes []*route.Route) (int64, error) {
	queryString := "INSERT INTO srd_routes (origin, destination, minimum_level, maximum_level, route_segment, sid, star) VALUES "
	// Build the query string
	queryArgs := make([]interface{}, 0)
	for i, route := range routes {
		if i > 0 {
			queryString += ", "
		}
		queryString += fmt.Sprintf("(?, ?, ?, ?, ?, ?, ?)")
		queryArgs = append(queryArgs, route.ADEPOrEntry(), route.ADESOrExit(), route.MinLevel(), route.MaxLevel(), route.RouteSegment(), route.SID(), route.STAR())
	}

	// Prepare the query
	stmt, err := t.tx.PrepareContext(ctx, queryString)
	if err != nil {
		return 0, err
	}

	// Execute the query
	res, err := stmt.ExecContext(ctx, queryArgs...)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// InsertNoteRouteLinkBatch inserts a batch of note-route links into the database
func (t *Transaction) InsertNoteRouteLinkBatch(ctx context.Context, noteRouteLinks []*NoteRouteLink) error {
	queryString := "INSERT INTO srd_note_srd_route (srd_note_id, srd_route_id) VALUES "
	// Build the query string
	queryArgs := make([]interface{}, 0)
	for i, link := range noteRouteLinks {
		if i > 0 {
			queryString += ", "
		}
		queryString += "(?, ?)"
		queryArgs = append(queryArgs, link.NoteID, link.RouteID)
	}

	// Prepare the query
	stmt, err := t.tx.PrepareContext(ctx, queryString)
	if err != nil {
		return err
	}

	// Execute the query
	_, err = stmt.ExecContext(ctx, queryArgs...)

	return err
}
