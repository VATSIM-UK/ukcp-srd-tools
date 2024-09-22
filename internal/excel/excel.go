package excel

import (
	"io"
	"iter"

	"github.com/rs/zerolog/log"
	"github.com/youkuang/xls"
)

const (
	SheetRoutes = 2
	SheetNotes  = 4
)

type ExcelFile interface {
	Close() error
	HasSheet(sheet int) bool
	SheetRows(sheet int) iter.Seq[[]string]
}

type excelFile struct {
	file   *xls.WorkBook
	closer io.Closer
}

func NewExcelFile(absPath string) (ExcelFile, error) {
	file, closer, err := xls.OpenWithCloser(absPath, "utf-8")
	if err != nil {
		log.Error().Err(err).Msg("failed to open excel file")
		return nil, err
	}

	return &excelFile{file, closer}, nil
}

func (f *excelFile) Close() error {
	return f.closer.Close()
}

func (f *excelFile) HasSheet(sheet int) bool {
	return f.file.GetSheet(sheet) != nil
}

func (f *excelFile) SheetRows(sheet int) iter.Seq[[]string] {
	return func(yield func([]string) bool) {
		worksheet := f.file.GetSheet(sheet)

		rowNumber := 0
		for rowNumber <= int(worksheet.MaxRow) {
			row := worksheet.Row(rowNumber)
			if !yield(rowToColumns(row)) {
				return
			}
			rowNumber++
		}
	}
}

func rowToColumns(row *xls.Row) []string {
	cells := make([]string, int(row.LastCol()))
	for i := 0; i < len(cells); i++ {
		cells[i] = row.Col(i)
	}
	return cells
}
