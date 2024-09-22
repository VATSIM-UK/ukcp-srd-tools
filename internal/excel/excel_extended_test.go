package excel

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VATSIM-UK/ukcp-srd-tools/test/logging"
)

func TestExcelExtended_ReturnsErrorOnFileNotFound(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	_, err := NewExcelExtendedFile("missing.xlsx")
	require.Error(err)
}

func TestExcelExtended_ReturnsErrorOnInvalidFile(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	_, err := NewExcelExtendedFile(testDataFile("invalid.xlsx"))
	require.Error(err)
}

func TestExcelExtended_ReturnsErrorOnOldExcelFile(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	_, err := NewExcelExtendedFile(testDataFile("simple1.xls"))
	require.Error(err)
}

func TestExcelExtended_HasSheets(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	excelFile, err := NewExcelExtendedFile(testDataFile("simple1.xlsx"))
	require.NoError(err)

	defer excelFile.Close()

	require.True(excelFile.HasSheet(SheetRoutes))
	require.True(excelFile.HasSheet(SheetNotes))
	require.False(excelFile.HasSheet(3))
}

func TestExcelExtended_SheetRows(t *testing.T) {
	require := require.New(t)

	_, cancel := logging.HijackLogs()
	defer cancel()

	excelFile, err := NewExcelExtendedFile(testDataFile("simple1.xlsx"))
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
