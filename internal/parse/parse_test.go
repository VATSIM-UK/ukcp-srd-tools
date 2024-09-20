package parse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/excel"
	"github.com/VATSIM-UK/ukcp-srd-import/internal/file"
)

// Do a benchmark of the Parse function
func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		path, _ := filepath.Abs("../../test/srd/test.xls")

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

		// Benchmark parsing the SRD file
		ParseSrd(file)
	}
}
