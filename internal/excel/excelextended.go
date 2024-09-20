package excel

import (
	"iter"

	"github.com/xuri/excelize/v2"
)

type excelExtendedFile struct {
	file *excelize.File
}

var sheetMap = map[int]string{
	SheetRoutes: "Routes",
	SheetNotes:  "Notes",
}

func NewExcelExtendedFile(absPath string) (ExcelFile, error) {
	file, err := excelize.OpenFile(absPath)
	if err != nil {
		return nil, err
	}

	return &excelExtendedFile{file}, nil
}

func (f *excelExtendedFile) Close() error {
	return f.file.Close()
}

func (f *excelExtendedFile) HasSheet(sheet int) bool {
	sheetName, sheetKnown := sheetMap[sheet]
	if !sheetKnown {
		return false
	}

	_, err := f.file.GetSheetIndex(sheetName)
	return err == nil
}

func (f *excelExtendedFile) sheetIndexToString(sheet int) string {
	sheetName, sheetKnown := sheetMap[sheet]
	if !sheetKnown {
		return ""
	}

	return sheetName
}

func (f *excelExtendedFile) SheetRows(sheet int) iter.Seq[[]string] {
	return func(yield func([]string) bool) {
		sheet := f.sheetIndexToString(sheet)
		if sheet == "" {
			panic("Sheet not found")
		}

		rows, _ := f.file.Rows(sheet)
		defer rows.Close()
		for rows.Next() {
			row, _ := rows.Columns()
			if !yield(row) {
				return
			}
		}
	}
}
