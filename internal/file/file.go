package file

import (
	"fmt"
	"iter"
	"regexp"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/note"
	"github.com/VATSIM-UK/ukcp-srd-tools/internal/route"
)

var NewRowRegxp = regexp.MustCompile(`^Note (\d+)$`)
var ScenarioRowRegxp = regexp.MustCompile(`^Scenario S\d+`)

type SrdFile interface {
	Routes() iter.Seq2[*route.Route, error]
	Notes() iter.Seq2[*note.Note, error]
	Stats() SrdStats
}

type SrdStats struct {
	RouteCount      int
	RouteErrorCount int
	NoteCount       int
	NoteErrorCount  int
}

func (s *SrdStats) RouteError() {
	s.RouteErrorCount++
}

func (s *SrdStats) Route() {
	s.RouteCount++
}

func (s *SrdStats) NoteError() {
	s.NoteErrorCount++
}

func (s *SrdStats) Note() {
	s.NoteCount++
}

func (s *SrdStats) ResetNoteStats() {
	s.NoteCount = 0
	s.NoteErrorCount = 0
}

func (s *SrdStats) ResetRouteStats() {
	s.RouteCount = 0
	s.RouteErrorCount = 0
}

type srdFile struct {
	file  excelFile
	stats SrdStats
}

type excelFile interface {
	Close() error
	HasSheet(sheet int) bool
	SheetRows(sheet int) iter.Seq[[]string]
}

func NewSrdFile(excelFile excelFile) (SrdFile, error) {
	if !excelFile.HasSheet(excel.SheetRoutes) {
		return nil, fmt.Errorf("Routes sheet, %d not found", excel.SheetRoutes)
	}

	if !excelFile.HasSheet(excel.SheetNotes) {
		return nil, fmt.Errorf("Notes sheet, %d not found", excel.SheetNotes)
	}

	return &srdFile{excelFile, SrdStats{}}, nil
}

func (f *srdFile) Routes() iter.Seq2[*route.Route, error] {
	headerRowProcessed := false
	return func(yield func(*route.Route, error) bool) {
		// Reset the route stats
		f.stats.ResetRouteStats()
		yieldWrapper := func(route *route.Route, err error) bool {
			f.incrementRouteStats(err)
			return yield(route, err)
		}

		for row := range f.file.SheetRows(excel.SheetRoutes) {
			if !headerRowProcessed {
				headerRowProcessed = true
				continue
			}

			if !yieldWrapper(mapRoute(row)) {
				return
			}
		}
	}
}

func (f *srdFile) Notes() iter.Seq2[*note.Note, error] {
	return func(yield func(*note.Note, error) bool) {
		rowsToProcess := make([][]string, 0)
		inNote := false

		// Reset the note stats
		f.stats.ResetNoteStats()
		yieldWrapper := func(note *note.Note, err error) bool {
			f.incrementNoteStats(err)
			return yield(note, err)
		}

		for row := range f.file.SheetRows(excel.SheetNotes) {
			if len(row) == 0 {
				continue
			}

			// If the row is empty, skip it
			if row[0] == "" {
				continue
			}

			isHeaderRow := NewRowRegxp.MatchString(row[0])
			if isHeaderRow {
				inNote = true
			}

			// We've found a header row and we already have some lines, so process the existing lines
			if isHeaderRow && len(rowsToProcess) > 0 {
				inNote = true
				if !yieldWrapper(mapNote(rowsToProcess)) {
					return
				}

				// Now we add the new header row to the list of rows to process
				rowsToProcess = [][]string{row}
				continue
			}

			// We've found our first header row, so add it to the list of rows to process
			if isHeaderRow {
				rowsToProcess = append(rowsToProcess, row)
				continue
			}

			// If we find a scenario row, we are now leaving a note, process any existing rows
			if ScenarioRowRegxp.MatchString(row[0]) {
				if len(rowsToProcess) > 0 {
					if !yieldWrapper(mapNote(rowsToProcess)) {
						return
					}

					rowsToProcess = make([][]string, 0)
				}
				inNote = false
				continue
			}

			// Only if we're in a note, add the row to the list of rows to process
			if inNote {
				rowsToProcess = append(rowsToProcess, row)
			}
		}

		// We've reached the end of the sheet, so process the last set of rows if there are any
		if len(rowsToProcess) > 0 {
			yieldWrapper(mapNote(rowsToProcess))
		}
	}
}

func (f *srdFile) incrementNoteStats(err error) {
	if err != nil {
		f.stats.NoteError()
	} else {
		f.stats.Note()
	}
}

func (f *srdFile) incrementRouteStats(err error) {
	if err != nil {
		f.stats.RouteError()
	} else {
		f.stats.Route()
	}
}

// Stats returns the statistics for the SRD file
func (f *srdFile) Stats() SrdStats {
	return f.stats
}
