package srd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/file"
)

// Do a benchmark of the Import function
func BenchmarkImport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		path, _ := filepath.Abs("../../test.xls")

		// Check if the file exists
		_, err := os.Stat(path)
		if err != nil {
			panic(err)
		}

		excelFile, err := excel.NewExcelFile(path)
		defer excelFile.Close()
		if err != nil {
			panic(err)
		}

		file, err := file.NewSrdFile(excelFile)
		if err != nil {
			panic(err)
		}

		// Benchmark iterating over the routes
		for _, err := range file.Routes() {
			if err != nil {
				panic(err)
			}
		}

		// Benchmark iterating over the notes
		for _, err := range file.Notes() {
			if err != nil {
				panic(err)
			}
		}
	}
}
