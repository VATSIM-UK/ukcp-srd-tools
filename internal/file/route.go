package file

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/route"
)

func mapRoute(row []string) (*route.Route, error) {
	if len(row) < 7 {
		return nil, fmt.Errorf("expected 7 or 8 fields, got %d, row: %v", len(row), row)
	}

	ADEPOrEntry, err := convertStringField(row, 0, "ADEP or Entry")
	if err != nil {
		return nil, err
	}

	SID := convertOptionalStringField(row, 1)

	minLevel, err := convertMinFlightLevelToAltitude(row)
	if err != nil {
		return nil, err
	}

	maxLevel, err := convertMaxFlightLevelToAltitude(row)
	if err != nil {
		return nil, err
	}

	routeSegment := convertStringFieldDefaultEmpty(row, 4)

	STAR := convertOptionalStringField(row, 5)

	ADESOrExit, err := convertStringField(row, 6, "ADES or Exit")
	if err != nil {
		return nil, err
	}

	noteIDs, err := convertNoteIDs(row)
	if err != nil {
		return nil, err
	}

	return route.NewRoute(ADEPOrEntry, SID, minLevel, maxLevel, routeSegment, STAR, ADESOrExit, noteIDs), nil
}

// convertStringField returns the string value of the field, or nil if the field is empty (with an error)
func convertStringField(row []string, index int, fieldName string) (string, error) {
	value := strings.TrimSpace(row[index])
	if value == "" {
		return "", missingValueError(row, fieldName)
	}
	return value, nil
}

// convertStringField returns the string value of the field, or empty if the field is empty
func convertStringFieldDefaultEmpty(row []string, index int) string {
	return strings.TrimSpace(row[index])
}

// convertOptionalStringField returns the string value of the field, or nil if the field is empty
func convertOptionalStringField(row []string, index int) *string {
	value := strings.TrimSpace(row[index])
	if value == "" {
		return nil
	}
	return &value
}

// flight levels in the SRD are either 3 digit numbers, of "MC"
// If MC or no value is provided, the altitude is not known
func convertFlightLevelToAltitude(row []string, index int) (*uint64, error) {
	minLevelString := strings.TrimSpace(row[index])
	if minLevelString == "MC" || minLevelString == "" {
		return nil, nil
	}

	fl, err := strconv.ParseUint(minLevelString, 10, 64)
	if err != nil {
		return nil, err
	}

	altitude := fl * 100
	return &altitude, nil
}

func convertMinFlightLevelToAltitude(row []string) (*uint64, error) {
	return convertFlightLevelToAltitude(row, 2)
}

func convertMaxFlightLevelToAltitude(row []string) (*uint64, error) {
	return convertFlightLevelToAltitude(row, 3)
}

// convertNoteIds converts the raw note ID column into a slice of note IDs
func convertNoteIDs(row []string) ([]uint64, error) {
	// If the length of the row is less than 8, there are no notes
	if len(row) < 8 {
		return nil, nil
	}

	// The first 7 characters of the remarks are "Notes: ", so exclude that
	remarks := strings.TrimSpace(row[7])
	if !strings.HasPrefix(remarks, "Notes: ") {
		return nil, nil
	}

	trimmedNotes := strings.TrimPrefix(remarks, "Notes: ")

	// The notes themselves are a list, separated by "-"
	noteStrings := strings.Split(trimmedNotes, "-")

	noteIDs := make([]uint64, 0)
	for _, noteValue := range noteStrings {
		// If for some reason we have an empty string, skip it
		trimmedNoteValue := strings.TrimSpace(noteValue)
		if trimmedNoteValue == "" {
			continue
		}

		noteID, err := strconv.ParseUint(trimmedNoteValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse note in remarks %s: %v", row, err)
		}

		noteIDs = append(noteIDs, noteID)
	}

	return noteIDs, nil
}

func missingValueError(row []string, fieldName string) error {
	return fmt.Errorf("missing value for %s, row: %v", fieldName, row)
}
