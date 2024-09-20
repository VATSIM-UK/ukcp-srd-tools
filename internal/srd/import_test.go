package srd

import (
	"context"
	"database/sql"
	"errors"
	"iter"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/db"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/file"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/route"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

type mysqlContainer struct {
	container     *mysql.MySQLContainer
	terminateFunc func()
}

const (
	TestDatabase = "uk_plugin"
	TestUsername = "uk_plugin_test"
	TestPassword = "test_password_123"
)

type testingT interface {
	Fatalf(format string, args ...interface{})
}

func getMysqlContainer(ctx context.Context, t testingT) (*mysqlContainer, error) {
	container, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithDatabase(TestDatabase),
		mysql.WithUsername(TestUsername),
		mysql.WithPassword(TestPassword),
		mysql.WithDefaultCredentials(),
		mysql.WithScripts("../../test/db/db_setup.sql"),
	)
	if err != nil {
		return nil, err
	}
	terminateFunc := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}

	return &mysqlContainer{container, terminateFunc}, nil
}

func TestImport_Successful(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)
	container, err := getMysqlContainer(ctx, t)
	require.NoError(err)
	defer container.terminateFunc()

	// Create a fake SRD file
	mockSrdFile := &mockSrdFile{
		notes: srdNoteList{
			{
				note: note.NewNote(1, "Note 1 Text"),
				err:  nil,
			},
			{
				note: note.NewNote(2, "Note 2 Text"),
				err:  nil,
			},
			{
				note: note.NewNote(3, "Note 3 Text"),
				err:  nil,
			},
		},
		routes: srdRouteList{
			{
				route: route.NewRoute("EGLL", ptr("SID1"), ptr(uint64(35000)), ptr(uint64(37000)), "SEGMENT", ptr("STAR1"), "EGKK", []uint64{1, 2}),
				err:   nil,
			},
			{
				route: route.NewRoute("EGKK",
					nil, ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", nil, "EGLL", []uint64{3}),
				err: nil,
			},
			{
				route: route.NewRoute("EGGD",
					nil, ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", nil, "EGLL", []uint64{}),
				err: nil,
			},
		},
	}

	containerHost, err := container.container.Host(ctx)
	containerInspect, err := container.container.Inspect(ctx)
	containerPort := containerInspect.NetworkSettings.Ports["3306/tcp"][0].HostPort
	// Convert port to int
	containerPortInt, err := strconv.Atoi(containerPort)
	require.NoError(err)

	// Create db
	db, err := db.NewDatabase(db.DatabaseConnectionParams{
		Host:     containerHost,
		Port:     containerPortInt,
		Username: TestUsername,
		Password: TestPassword,
		Database: TestDatabase,
	})
	require.NoError(err)

	defer db.Close()

	// Create the importer and go
	importer := NewImport(mockSrdFile, db)

	err = importer.Import(ctx)
	require.NoError(err)

	// Check we have the notes in the database
	dbHandle := db.Handle()

	// Check notes
	noteRows := allNotes(ctx, require, dbHandle)
	require.Len(noteRows, 3)

	// Check first note
	require.Equal("1", noteRows[0].id)
	require.Equal("Note 1 Text", noteRows[0].text)

	// Check second note
	require.Equal("2", noteRows[1].id)
	require.Equal("Note 2 Text", noteRows[1].text)

	// Check thord note
	require.Equal("3", noteRows[2].id)
	require.Equal("Note 3 Text", noteRows[2].text)

	// Check routes
	routeRows := allRoutes(ctx, require, dbHandle)
	require.Len(routeRows, 3)

	// Check first route
	require.Equal("EGLL", routeRows[0].origin)
	require.Equal("EGKK", routeRows[0].destination)
	require.Equal("SID1", *routeRows[0].sid)
	require.Equal("STAR1", *routeRows[0].star)
	require.Equal("35000", routeRows[0].minimum_level)
	require.Equal("37000", routeRows[0].maximum_level)
	require.Equal("SEGMENT", routeRows[0].route_segment)

	// Check second route
	require.Equal("EGKK", routeRows[1].origin)
	require.Equal("EGLL", routeRows[1].destination)
	require.Nil(routeRows[1].sid)
	require.Nil(routeRows[1].star)
	require.Equal("37000", routeRows[1].minimum_level)
	require.Equal("39000", routeRows[1].maximum_level)
	require.Equal("SEGMENT", routeRows[1].route_segment)

	// Check third route
	require.Equal("EGGD", routeRows[2].origin)
	require.Equal("EGLL", routeRows[2].destination)
	require.Nil(routeRows[2].sid)
	require.Nil(routeRows[2].star)
	require.Equal("37000", routeRows[2].minimum_level)
	require.Equal("39000", routeRows[2].maximum_level)
	require.Equal("SEGMENT", routeRows[2].route_segment)

	// Check the mappings
	allMappings := allRouteNoteLinks(ctx, require, dbHandle)
	require.Len(allMappings, 3)

	// Check the mappings
	require.Equal(routeRows[0].id, allMappings[0].route)
	require.Equal("1", allMappings[0].note)
	require.Equal(routeRows[0].id, allMappings[1].route)
	require.Equal("2", allMappings[1].note)
	require.Equal(routeRows[1].id, allMappings[2].route)
	require.Equal("3", allMappings[2].note)
}

func TestImport_ErrornousRoutes(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)
	container, err := getMysqlContainer(ctx, t)
	require.NoError(err)
	defer container.terminateFunc()

	// Create a fake SRD file
	mockSrdFile := &mockSrdFile{
		notes: srdNoteList{
			{
				note: note.NewNote(1, "Note 1 Text"),
				err:  nil,
			},
			{
				note: note.NewNote(2, "Note 2 Text"),
				err:  nil,
			},
			{
				note: note.NewNote(3, "Note 3 Text"),
				err:  nil,
			},
		},
		routes: srdRouteList{
			{
				route: route.NewRoute("EGLL", ptr("SID1"), ptr(uint64(35000)), ptr(uint64(37000)), "SEGMENT", ptr("STAR1"), "EGKK", []uint64{1, 2}),
				err:   nil,
			},
			{
				route: nil,
				err:   errors.New("foo"),
			},
			{
				route: route.NewRoute("EGGD",
					nil, ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", nil, "EGLL", []uint64{}),
				err: nil,
			},
		},
	}

	containerHost, err := container.container.Host(ctx)
	containerInspect, err := container.container.Inspect(ctx)
	containerPort := containerInspect.NetworkSettings.Ports["3306/tcp"][0].HostPort
	// Convert port to int
	containerPortInt, err := strconv.Atoi(containerPort)
	require.NoError(err)

	// Create db
	db, err := db.NewDatabase(db.DatabaseConnectionParams{
		Host:     containerHost,
		Port:     containerPortInt,
		Username: TestUsername,
		Password: TestPassword,
		Database: TestDatabase,
	})
	require.NoError(err)

	defer db.Close()

	// Create the importer and go
	importer := NewImport(mockSrdFile, db)

	err = importer.Import(ctx)
	require.NoError(err)

	// Check we have the notes in the database
	dbHandle := db.Handle()

	// Check notes
	noteRows := allNotes(ctx, require, dbHandle)
	require.Len(noteRows, 3)

	// Check first note
	require.Equal("1", noteRows[0].id)
	require.Equal("Note 1 Text", noteRows[0].text)

	// Check second note
	require.Equal("2", noteRows[1].id)
	require.Equal("Note 2 Text", noteRows[1].text)

	// Check thord note
	require.Equal("3", noteRows[2].id)
	require.Equal("Note 3 Text", noteRows[2].text)

	// Check routes
	routeRows := allRoutes(ctx, require, dbHandle)
	require.Len(routeRows, 2)

	// Check first route
	require.Equal("EGLL", routeRows[0].origin)
	require.Equal("EGKK", routeRows[0].destination)
	require.Equal("SID1", *routeRows[0].sid)
	require.Equal("STAR1", *routeRows[0].star)
	require.Equal("35000", routeRows[0].minimum_level)
	require.Equal("37000", routeRows[0].maximum_level)
	require.Equal("SEGMENT", routeRows[0].route_segment)

	// Check second route
	require.Equal("EGGD", routeRows[1].origin)
	require.Equal("EGLL", routeRows[1].destination)
	require.Nil(routeRows[1].sid)
	require.Nil(routeRows[1].star)
	require.Equal("37000", routeRows[1].minimum_level)
	require.Equal("39000", routeRows[1].maximum_level)
	require.Equal("SEGMENT", routeRows[1].route_segment)

	// Check the mappings
	allMappings := allRouteNoteLinks(ctx, require, dbHandle)
	require.Len(allMappings, 2)

	// Check the mappings
	require.Equal(routeRows[0].id, allMappings[0].route)
	require.Equal("1", allMappings[0].note)
	require.Equal(routeRows[0].id, allMappings[1].route)
	require.Equal("2", allMappings[1].note)
}

func TestImport_ErroneousNotes(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)
	container, err := getMysqlContainer(ctx, t)
	require.NoError(err)
	defer container.terminateFunc()

	// Create a fake SRD file
	mockSrdFile := &mockSrdFile{
		notes: srdNoteList{
			{
				note: note.NewNote(1, "Note 1 Text"),
				err:  nil,
			},
			{
				note: note.NewNote(2, "Note 2 Text"),
				err:  nil,
			},
			{
				note: note.NewNote(3, "Note 3 Text"),
				err:  nil,
			},
			{
				note: nil,
				err:  errors.New("foo"),
			},
		},
		routes: srdRouteList{
			{
				route: route.NewRoute("EGLL", ptr("SID1"), ptr(uint64(35000)), ptr(uint64(37000)), "SEGMENT", ptr("STAR1"), "EGKK", []uint64{1, 2}),
				err:   nil,
			},
			{
				route: route.NewRoute("EGKK",
					nil, ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", nil, "EGLL", []uint64{3}),
				err: nil,
			},
			{
				route: route.NewRoute("EGGD",
					nil, ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", nil, "EGLL", []uint64{}),
				err: nil,
			},
		},
	}

	containerHost, err := container.container.Host(ctx)
	containerInspect, err := container.container.Inspect(ctx)
	containerPort := containerInspect.NetworkSettings.Ports["3306/tcp"][0].HostPort
	// Convert port to int
	containerPortInt, err := strconv.Atoi(containerPort)
	require.NoError(err)

	// Create db
	db, err := db.NewDatabase(db.DatabaseConnectionParams{
		Host:     containerHost,
		Port:     containerPortInt,
		Username: TestUsername,
		Password: TestPassword,
		Database: TestDatabase,
	})
	require.NoError(err)

	defer db.Close()

	// Create the importer and go
	importer := NewImport(mockSrdFile, db)

	err = importer.Import(ctx)
	require.NoError(err)

	// Check we have the notes in the database
	dbHandle := db.Handle()

	// Check notes
	noteRows := allNotes(ctx, require, dbHandle)
	require.Len(noteRows, 3)

	// Check first note
	require.Equal("1", noteRows[0].id)
	require.Equal("Note 1 Text", noteRows[0].text)

	// Check second note
	require.Equal("2", noteRows[1].id)
	require.Equal("Note 2 Text", noteRows[1].text)

	// Check thord note
	require.Equal("3", noteRows[2].id)
	require.Equal("Note 3 Text", noteRows[2].text)

	// Check routes
	routeRows := allRoutes(ctx, require, dbHandle)
	require.Len(routeRows, 3)

	// Check first route
	require.Equal("EGLL", routeRows[0].origin)
	require.Equal("EGKK", routeRows[0].destination)
	require.Equal("SID1", *routeRows[0].sid)
	require.Equal("STAR1", *routeRows[0].star)
	require.Equal("35000", routeRows[0].minimum_level)
	require.Equal("37000", routeRows[0].maximum_level)
	require.Equal("SEGMENT", routeRows[0].route_segment)

	// Check second route
	require.Equal("EGKK", routeRows[1].origin)
	require.Equal("EGLL", routeRows[1].destination)
	require.Nil(routeRows[1].sid)
	require.Nil(routeRows[1].star)
	require.Equal("37000", routeRows[1].minimum_level)
	require.Equal("39000", routeRows[1].maximum_level)
	require.Equal("SEGMENT", routeRows[1].route_segment)

	// Check third route
	require.Equal("EGGD", routeRows[2].origin)
	require.Equal("EGLL", routeRows[2].destination)
	require.Nil(routeRows[2].sid)
	require.Nil(routeRows[2].star)
	require.Equal("37000", routeRows[2].minimum_level)
	require.Equal("39000", routeRows[2].maximum_level)
	require.Equal("SEGMENT", routeRows[2].route_segment)

	// Check the mappings
	allMappings := allRouteNoteLinks(ctx, require, dbHandle)
	require.Len(allMappings, 3)

	// Check the mappings
	require.Equal(routeRows[0].id, allMappings[0].route)
	require.Equal("1", allMappings[0].note)
	require.Equal(routeRows[0].id, allMappings[1].route)
	require.Equal("2", allMappings[1].note)
	require.Equal(routeRows[1].id, allMappings[2].route)
	require.Equal("3", allMappings[2].note)
}

func TestImport_BadRouteNoteLinks(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)
	container, err := getMysqlContainer(ctx, t)
	require.NoError(err)
	defer container.terminateFunc()

	// Create a fake SRD file
	mockSrdFile := &mockSrdFile{
		notes: srdNoteList{
			{
				note: note.NewNote(1, "Note 1 Text"),
				err:  nil,
			},
			{
				note: note.NewNote(2, "Note 2 Text"),
				err:  nil,
			},
			{
				note: note.NewNote(3, "Note 3 Text"),
				err:  nil,
			},
			{
				note: nil,
				err:  errors.New("foo"),
			},
		},
		routes: srdRouteList{
			{
				route: route.NewRoute("EGLL", ptr("SID1"), ptr(uint64(35000)), ptr(uint64(37000)), "SEGMENT", ptr("STAR1"), "EGKK", []uint64{1, 2, 55}),
				err:   nil,
			},
			{
				route: route.NewRoute("EGKK",
					nil, ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", nil, "EGLL", []uint64{3}),
				err: nil,
			},
			{
				route: route.NewRoute("EGGD",
					nil, ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", nil, "EGLL", []uint64{24}),
				err: nil,
			},
		},
	}

	containerHost, err := container.container.Host(ctx)
	containerInspect, err := container.container.Inspect(ctx)
	containerPort := containerInspect.NetworkSettings.Ports["3306/tcp"][0].HostPort
	// Convert port to int
	containerPortInt, err := strconv.Atoi(containerPort)
	require.NoError(err)

	// Create db
	db, err := db.NewDatabase(db.DatabaseConnectionParams{
		Host:     containerHost,
		Port:     containerPortInt,
		Username: TestUsername,
		Password: TestPassword,
		Database: TestDatabase,
	})
	require.NoError(err)

	defer db.Close()

	// Create the importer and go
	importer := NewImport(mockSrdFile, db)

	err = importer.Import(ctx)
	require.NoError(err)

	// Check we have the notes in the database
	dbHandle := db.Handle()

	// Check notes
	noteRows := allNotes(ctx, require, dbHandle)
	require.Len(noteRows, 3)

	// Check first note
	require.Equal("1", noteRows[0].id)
	require.Equal("Note 1 Text", noteRows[0].text)

	// Check second note
	require.Equal("2", noteRows[1].id)
	require.Equal("Note 2 Text", noteRows[1].text)

	// Check thord note
	require.Equal("3", noteRows[2].id)
	require.Equal("Note 3 Text", noteRows[2].text)

	// Check routes
	routeRows := allRoutes(ctx, require, dbHandle)
	require.Len(routeRows, 3)

	// Check first route
	require.Equal("EGLL", routeRows[0].origin)
	require.Equal("EGKK", routeRows[0].destination)
	require.Equal("SID1", *routeRows[0].sid)
	require.Equal("STAR1", *routeRows[0].star)
	require.Equal("35000", routeRows[0].minimum_level)
	require.Equal("37000", routeRows[0].maximum_level)
	require.Equal("SEGMENT", routeRows[0].route_segment)

	// Check second route
	require.Equal("EGKK", routeRows[1].origin)
	require.Equal("EGLL", routeRows[1].destination)
	require.Nil(routeRows[1].sid)
	require.Nil(routeRows[1].star)
	require.Equal("37000", routeRows[1].minimum_level)
	require.Equal("39000", routeRows[1].maximum_level)
	require.Equal("SEGMENT", routeRows[1].route_segment)

	// Check third route
	require.Equal("EGGD", routeRows[2].origin)
	require.Equal("EGLL", routeRows[2].destination)
	require.Nil(routeRows[2].sid)
	require.Nil(routeRows[2].star)
	require.Equal("37000", routeRows[2].minimum_level)
	require.Equal("39000", routeRows[2].maximum_level)
	require.Equal("SEGMENT", routeRows[2].route_segment)

	// Check the mappings
	allMappings := allRouteNoteLinks(ctx, require, dbHandle)
	require.Len(allMappings, 3)

	// Check the mappings
	require.Equal(routeRows[0].id, allMappings[0].route)
	require.Equal("1", allMappings[0].note)
	require.Equal(routeRows[0].id, allMappings[1].route)
	require.Equal("2", allMappings[1].note)
	require.Equal(routeRows[1].id, allMappings[2].route)
	require.Equal("3", allMappings[2].note)
}

func BenchmarkImport(b *testing.B) {
	ctx := context.Background()
	require := require.New(b)
	container, err := getMysqlContainer(ctx, b)
	require.NoError(err)
	defer container.terminateFunc()

	containerHost, err := container.container.Host(ctx)
	containerInspect, err := container.container.Inspect(ctx)
	containerPort := containerInspect.NetworkSettings.Ports["3306/tcp"][0].HostPort
	// Convert port to int
	containerPortInt, err := strconv.Atoi(containerPort)
	require.NoError(err)

	// Create db
	db, err := db.NewDatabase(db.DatabaseConnectionParams{
		Host:     containerHost,
		Port:     containerPortInt,
		Username: TestUsername,
		Password: TestPassword,
		Database: TestDatabase,
	})
	require.NoError(err)

	defer db.Close()

	// Reset the benchmark timer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		path, _ := filepath.Abs("../../test/srd/test.xls")

		// Check if the file exists
		_, err = os.Stat(path)
		require.NoError(err)

		// Get the heap allocated right now
		memStats := new(runtime.MemStats)
		runtime.ReadMemStats(memStats)
		allocatedBefore := memStats.HeapAlloc

		excelFile, err := excel.NewExcelFile(path)
		defer excelFile.Close()
		require.NoError(err)

		file, err := file.NewSrdFile(excelFile)
		require.NoError(err)

		// Create the importer and go
		importer := NewImport(file, db)

		err = importer.Import(ctx)
		require.NoError(err)

		// Get the heap allocated right now
		runtime.ReadMemStats(memStats)
		allocatedAfter := memStats.HeapAlloc

		b.Logf("Heap allocated for import: %d", allocatedAfter-allocatedBefore)
	}
}

func BenchmarkImportXlsx(b *testing.B) {
	ctx := context.Background()
	require := require.New(b)
	container, err := getMysqlContainer(ctx, b)
	require.NoError(err)
	defer container.terminateFunc()

	containerHost, err := container.container.Host(ctx)
	containerInspect, err := container.container.Inspect(ctx)
	containerPort := containerInspect.NetworkSettings.Ports["3306/tcp"][0].HostPort
	// Convert port to int
	containerPortInt, err := strconv.Atoi(containerPort)
	require.NoError(err)

	// Create db
	db, err := db.NewDatabase(db.DatabaseConnectionParams{
		Host:     containerHost,
		Port:     containerPortInt,
		Username: TestUsername,
		Password: TestPassword,
		Database: TestDatabase,
	})
	require.NoError(err)

	defer db.Close()

	// Reset the benchmark timer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		path, _ := filepath.Abs("../../test/srd/test.xlsx")

		// Check if the file exists
		_, err = os.Stat(path)
		require.NoError(err)

		// Get the heap allocated right now
		memStats := new(runtime.MemStats)
		runtime.ReadMemStats(memStats)
		allocatedBefore := memStats.HeapAlloc

		excelFile, err := excel.NewExcelExtendedFile(path)
		defer excelFile.Close()
		require.NoError(err)

		file, err := file.NewSrdFile(excelFile)
		require.NoError(err)

		// Create the importer and go
		importer := NewImport(file, db)

		err = importer.Import(ctx)
		require.NoError(err)

		// Get the heap allocated right now
		runtime.ReadMemStats(memStats)
		allocatedAfter := memStats.HeapAlloc

		b.Logf("Heap allocated for import: %d", allocatedAfter-allocatedBefore)
	}
}

func allNotes(ctx context.Context, require *require.Assertions, db *sql.DB) []NoteRow {
	notes, err := db.QueryContext(ctx, "SELECT id, note_text FROM srd_notes")
	require.NoError(err)

	defer notes.Close()

	foundNotes := make([]NoteRow, 0)

	for notes.Next() {
		var noteRow NoteRow
		notes.Scan(&noteRow.id, &noteRow.text)
		foundNotes = append(foundNotes, noteRow)
	}

	return foundNotes
}

func allRoutes(ctx context.Context, require *require.Assertions, db *sql.DB) []RouteRow {
	routes, err := db.QueryContext(ctx, "SELECT id, origin, destination, minimum_level, maximum_level, route_segment, sid, star FROM srd_routes")
	require.NoError(err)

	defer routes.Close()

	foundRoutes := make([]RouteRow, 0)

	for routes.Next() {
		var routeRow RouteRow
		routes.Scan(&routeRow.id, &routeRow.origin, &routeRow.destination, &routeRow.minimum_level, &routeRow.maximum_level, &routeRow.route_segment, &routeRow.sid, &routeRow.star)
		foundRoutes = append(foundRoutes, routeRow)
	}

	return foundRoutes
}

func allRouteNoteLinks(ctx context.Context, require *require.Assertions, db *sql.DB) []NoteRouteRow {
	rows, err := db.QueryContext(ctx, "SELECT srd_note_id, srd_route_id FROM srd_note_srd_route")
	require.NoError(err)

	defer rows.Close()

	foundRows := make([]NoteRouteRow, 0)

	for rows.Next() {
		var routeRow NoteRouteRow
		rows.Scan(&routeRow.note, &routeRow.route)
		foundRows = append(foundRows, routeRow)
	}

	return foundRows
}

type NoteRow struct {
	id   string
	text string
}

type RouteRow struct {
	id            string
	origin        string
	destination   string
	minimum_level string
	maximum_level string
	route_segment string
	sid           *string
	star          *string
}

type NoteRouteRow struct {
	route string
	note  string
}

type srdNoteEntry struct {
	note *note.Note
	err  error
}

type srdNoteList []srdNoteEntry

type srdRouteEntry struct {
	route *route.Route
	err   error
}

type srdRouteList []srdRouteEntry

type mockSrdFile struct {
	routes srdRouteList
	notes  srdNoteList
}

func (m *mockSrdFile) Routes() iter.Seq2[*route.Route, error] {
	return func(yield func(*route.Route, error) bool) {
		for _, route := range m.routes {
			if !yield(route.route, route.err) {
				return
			}
		}
	}
}

func (m *mockSrdFile) Notes() iter.Seq2[*note.Note, error] {
	return func(yield func(*note.Note, error) bool) {
		for _, note := range m.notes {
			if !yield(note.note, note.err) {
				return
			}
		}
	}
}

func ptr[V string | uint64](v V) *V {
	return &v
}
