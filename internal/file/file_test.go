package file

import (
	"errors"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/route"
)

// Mock for excelFile interface
type mockExcelFile struct {
	sheetRows map[int][][]string
}

func (m *mockExcelFile) Close() error {
	return nil
}

func (m *mockExcelFile) HasSheet(sheet int) bool {
	_, ok := m.sheetRows[sheet]
	return ok
}

func (m *mockExcelFile) SheetRows(sheet int) iter.Seq[[]string] {
	return func(yield func([]string) bool) {
		for _, row := range m.sheetRows[sheet] {
			if !yield(row) {
				return
			}
		}
	}
}

func TestNewSrdFile(t *testing.T) {
	tests := []struct {
		name          string
		hasSheet      map[int]bool
		expectedError error
	}{
		{
			name: "Routes sheet not found",
			hasSheet: map[int]bool{
				excel.SheetRoutes: false,
			},
			expectedError: errors.New("Routes sheet, 2 not found"),
		},
		{
			name: "Notes sheet not found",
			hasSheet: map[int]bool{
				excel.SheetRoutes: true,
				excel.SheetNotes:  false,
			},
			expectedError: errors.New("Notes sheet, 4 not found"),
		},
		{
			name: "Both sheets found",
			hasSheet: map[int]bool{
				excel.SheetRoutes: true,
				excel.SheetNotes:  true,
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExcel := &mockExcelFile{
				sheetRows: map[int][][]string{},
			}

			for sheet, hasSheet := range tt.hasSheet {
				if hasSheet {
					mockExcel.sheetRows[sheet] = [][]string{}
				}
			}

			srdFile, err := NewSrdFile(mockExcel)

			if tt.expectedError != nil {
				assert.Nil(t, srdFile)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NotNil(t, srdFile)
				assert.NoError(t, err)
			}
		})
	}
}
func ptr[V string | uint64](v V) *V {
	return &v
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		name           string
		sheetRows      map[int][][]string
		expectedRoutes []*route.Route
		expectedErrors []error
	}{
		{
			name: "No routes",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes:  {},
			},
			expectedRoutes: nil,
			expectedErrors: nil,
		},
		{
			name: "Single route",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {
					{"Some", "Header", "Row"},
					{"EGLL", "SID1", "350", "370", "SEGMENT", "STAR1", "EGKK", "Notes: 123-456"},
				},
				excel.SheetNotes: {},
			},
			expectedRoutes: []*route.Route{
				route.NewRoute("EGLL", ptr("SID1"), ptr(uint64(35000)), ptr(uint64(37000)), "SEGMENT", ptr("STAR1"), "EGKK", []uint64{123, 456}),
			},
			expectedErrors: []error{nil},
		},
		{
			name: "Multiple routes",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {
					{"Some", "Header", "Row"},
					{"EGLL", "SID1", "350", "370", "SEGMENT", "STAR1", "EGKK", "Notes: 123-456"},
					{"EGKK", "SID2", "370", "390", "SEGMENT", "STAR2", "EGLL", "Notes: 789-012"},
				},
				excel.SheetNotes: {},
			},
			expectedRoutes: []*route.Route{
				route.NewRoute("EGLL", ptr("SID1"), ptr(uint64(35000)), ptr(uint64(37000)), "SEGMENT", ptr("STAR1"), "EGKK", []uint64{123, 456}),
				route.NewRoute("EGKK", ptr("SID2"), ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", ptr("STAR2"), "EGLL", []uint64{789, 12}),
			},
			expectedErrors: []error{nil, nil},
		},
		{
			name: "Invalid route",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {
					{"Some", "Header", "Row"},
					{"EGLL", "SID1", "350", "370", "SEGMENT", "STAR1", "EGKK", "Notes: 123-abc"},
					{"EGKK", "SID2", "370", "390", "SEGMENT", "STAR2", "EGLL", "Notes: 789-012"},
				},
				excel.SheetNotes: {},
			},
			expectedRoutes: []*route.Route{
				nil,
				route.NewRoute("EGKK", ptr("SID2"), ptr(uint64(37000)), ptr(uint64(39000)), "SEGMENT", ptr("STAR2"), "EGLL", []uint64{789, 12}),
			},
			expectedErrors: []error{errors.New("failed to parse note in remarks [EGLL SID1 350 370 SEGMENT STAR1 EGKK Notes: 123-abc]: strconv.ParseUint: parsing \"abc\": invalid syntax"), nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			mockExcel := &mockExcelFile{
				sheetRows: tt.sheetRows,
			}

			srdFile, err := NewSrdFile(mockExcel)
			require.NoError(err)
			require.NotNil(srdFile)

			var routes []*route.Route
			var routeErrors []error
			for route, err := range srdFile.Routes() {
				routes = append(routes, route)
				routeErrors = append(routeErrors, err)
			}

			require.Equal(tt.expectedRoutes, routes)
			require.Equal(tt.expectedErrors, routeErrors)
		})
	}
}
func TestNotes(t *testing.T) {
	tests := []struct {
		name           string
		sheetRows      map[int][][]string
		expectedNotes  []*note.Note
		expectedErrors []error
	}{
		{
			name: "No notes",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes:  {},
			},
			expectedNotes:  nil,
			expectedErrors: nil,
		},
		{
			name: "Empty line",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{""},
				},
			},
			expectedNotes:  nil,
			expectedErrors: nil,
		},
		{
			name: "Empty string",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{""},
				},
			},
			expectedNotes:  nil,
			expectedErrors: nil,
		},
		{
			name: "Single note",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{"Note 1"},
					{"Some Note Content"},
				},
			},
			expectedNotes: []*note.Note{
				note.NewNote(
					uint64(1),
					"Some Note Content",
				),
			},
			expectedErrors: []error{nil},
		},
		{
			name: "Multiple notes",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{"Note 1"},
					{"Some Note Content"},
					{"Note 2"},
					{"Another Note Content"},
				},
			},
			expectedNotes: []*note.Note{
				note.NewNote(
					1,
					"Some Note Content",
				),
				note.NewNote(
					2,
					"Another Note Content",
				),
			},
			expectedErrors: []error{nil, nil},
		},
		{
			name: "Note with multiple lines",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{"Note 1"},
					{"Some Note Content"},
					{"Another Note Content"},
				},
			},
			expectedNotes: []*note.Note{
				note.NewNote(
					1,
					"Some Note Content\nAnother Note Content",
				),
			},
			expectedErrors: []error{nil},
		},
		{
			name: "Multiple notes with multiple lines",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{"Note 1"},
					{"Some Note Content"},
					{"Another Note Content"},
					{"Note 2"},
					{"Another Note Content 2"},
					{"Another Note Content 3"},
				},
			},
			expectedNotes: []*note.Note{
				note.NewNote(
					1,
					"Some Note Content\nAnother Note Content",
				),
				note.NewNote(
					2,
					"Another Note Content 2\nAnother Note Content 3",
				),
			},
			expectedErrors: []error{nil, nil},
		},
		{
			name: "Note with scenario row",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{"Note 1"},
					{"Some Note Content"},
					{"Scenario S1"},
					{"Note 2"},
					{"Another Note Content"},
				},
			},
			expectedNotes: []*note.Note{
				note.NewNote(
					1,
					"Some Note Content",
				),
				note.NewNote(
					2,
					"Another Note Content",
				),
			},
			expectedErrors: []error{nil, nil},
		},
		{
			name: "Notes ending with multiple blank lines",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{"Note 1"},
					{"Some Note Content"},
					{""},
					{""},
				},
			},
			expectedNotes: []*note.Note{
				note.NewNote(
					1,
					"Some Note Content",
				),
			},
			expectedErrors: []error{nil},
		},
		{
			name: "Notes ending with multiple scenario rows",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{"Note 1"},
					{"Some Note Content"},
					{"Scenario S1"},
					{"some scenario content"},
					{"Scenario S2"},
					{"another scenario content"},
				},
			},
			expectedNotes: []*note.Note{
				note.NewNote(
					1,
					"Some Note Content",
				),
			},
			expectedErrors: []error{nil},
		},
		{
			name: "Invalid note",
			sheetRows: map[int][][]string{
				excel.SheetRoutes: {},
				excel.SheetNotes: {
					{"Note abc"},
					{"Some Note Content"},
				},
			},
			expectedNotes:  nil,
			expectedErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			mockExcel := &mockExcelFile{
				sheetRows: tt.sheetRows,
			}

			srdFile, err := NewSrdFile(mockExcel)
			require.NoError(err)
			require.NotNil(srdFile)

			var notes []*note.Note
			var noteErrors []error
			for note, err := range srdFile.Notes() {
				notes = append(notes, note)
				noteErrors = append(noteErrors, err)
			}

			require.Equal(tt.expectedNotes, notes)
			require.Equal(tt.expectedErrors, noteErrors)
		})
	}
}
