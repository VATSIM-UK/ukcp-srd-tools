package parse

import (
	"errors"
	"iter"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/file"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/route"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	require := require.New(t)

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
				note: note.NewNote(4, "Note 4 Text"),
				err:  nil,
			},
			{
				note: nil,
				err:  errors.New("foo"),
			},
			{
				note: nil,
				err:  errors.New("foo 2"),
			},
			{
				note: nil,
				err:  errors.New("foo 3"),
			},
		},
		routes: srdRouteList{
			{
				route: route.NewRoute("EGLL", ptr("SID1"), ptr(uint64(35000)), ptr(uint64(37000)), "SEGMENT", ptr("STAR1"), "EGKK", []uint64{1, 2}),
				err:   nil,
			},
			{
				route: route.NewRoute("EGFF", ptr("SID1"), ptr(uint64(35000)), ptr(uint64(37000)), "SEGMENT", ptr("STAR1"), "EGKK", []uint64{1, 2}),
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
			{
				route: nil,
				err:   errors.New("foo"),
			},
		},
	}

	// Parse the SRD file
	summary := ParseSrd(mockSrdFile)

	// Check the summary
	require.Equal(3, summary.RouteCount)
	require.Equal(2, summary.RouteErrorCount)
	require.Equal(4, summary.NoteCount)
	require.Equal(3, summary.NoteErrorCount)
}

// Do a benchmark of the Parse function
func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		path, _ := filepath.Abs("../../test/srd/test.xls")

		// Check if the file exists
		_, err := os.Stat(path)
		if err != nil {
			panic(err)
		}

		// Get the heap allocated right now
		memStats := new(runtime.MemStats)
		runtime.ReadMemStats(memStats)
		allocatedBefore := memStats.HeapAlloc

		excelFile, err := excel.NewExcelFile(path)
		defer excelFile.Close()
		if err != nil {
			panic(err)
		}

		file, err := file.NewSrdFile(excelFile)
		if err != nil {
			panic(err)
		}

		// Benchmark parsing the SRD file
		ParseSrd(file)

		// Get the heap allocated right now
		runtime.ReadMemStats(memStats)
		allocatedAfter := memStats.HeapAlloc

		// Print the heap allocated
		b.Logf("Allocated: %d", allocatedAfter-allocatedBefore)
	}
}

func BenchmarkParseXlsx(b *testing.B) {
	for i := 0; i < b.N; i++ {
		path, _ := filepath.Abs("../../test/srd/test.xlsx")

		// Check if the file exists
		_, err := os.Stat(path)
		if err != nil {
			panic(err)
		}

		// Get the heap allocated right now
		memStats := new(runtime.MemStats)
		runtime.ReadMemStats(memStats)
		allocatedBefore := memStats.HeapAlloc

		excelFile, err := excel.NewExcelExtendedFile(path)
		defer excelFile.Close()
		if err != nil {
			panic(err)
		}

		file, err := file.NewSrdFile(excelFile)
		if err != nil {
			panic(err)
		}

		// Benchmark parsing the SRD file
		ParseSrd(file)

		// Get the heap allocated right now
		runtime.ReadMemStats(memStats)
		allocatedAfter := memStats.HeapAlloc

		// Print the heap allocated
		b.Logf("Allocated: %d", allocatedAfter-allocatedBefore)
	}
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
