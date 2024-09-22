package excel

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VATSIM-UK/ukcp-srd-import/test/logging"
)

func TestExcel_ReturnsErrorOnFileNotFound(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	_, err := NewExcelFile("missing.xls")
	require.Error(err)
}

func TestExcel_ReturnsErrorOnInvalidFile(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	_, err := NewExcelFile(testDataFile("invalid.xls"))
	require.Error(err)
}

func TestExcel_ReturnsErrorOnExcelExtendedFile(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	_, err := NewExcelFile(testDataFile("simple1.xlsx"))
	require.Error(err)
}

func TestExcel_HasSheets(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	excelFile, err := NewExcelFile(testDataFile("simple1.xls"))
	require.NoError(err)

	defer excelFile.Close()

	require.True(excelFile.HasSheet(SheetRoutes))
	require.True(excelFile.HasSheet(SheetNotes))
	require.False(excelFile.HasSheet(55))
}

func TestExcel_SheetRows(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	excelFile, err := NewExcelFile(testDataFile("simple1.xls"))
	require.NoError(err)

	defer excelFile.Close()

	rows := excelFile.SheetRows(SheetRoutes)
	require.NotNil(rows)

	rowCount := 0
	rowValues := []string{}
	for row := range rows {
		rowCount++
		rowValues = append(rowValues, row[0])
	}

	require.Equal(4, rowCount)

	// Our rows should contain EGKK, EGLL, EGGD in that order
	require.Equal([]string{"A", "EGKK", "EGLL", "EGGD"}, rowValues)
}

func testDataFile(filename string) string {
	return fmt.Sprintf("../../test/data/%s", filename)
}
