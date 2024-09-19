package file

import (
	"testing"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/route"
	"github.com/stretchr/testify/require"
)

func TestConvertNoteIDs(t *testing.T) {
	tests := []struct {
		name     string
		row      []string
		expected []uint64
		wantErr  bool
	}{
		{
			name:     "Valid notes",
			row:      []string{"", "", "", "", "", "", "", "Notes: 123-456-789"},
			expected: []uint64{123, 456, 789},
			wantErr:  false,
		},
		{
			name:     "Empty notes",
			row:      []string{"", "", "", "", "", "", "", "Notes: "},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "No notes prefix",
			row:      []string{"", "", "", "", "", "", "", ""},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "Invalid note ID",
			row:      []string{"", "", "", "", "", "", "", "Notes: 123-abc-789"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Trailing dash",
			row:      []string{"", "", "", "", "", "", "", "Notes: 123-456-"},
			expected: []uint64{123, 456},
			wantErr:  false,
		},
		{
			name:     "Leading dash",
			row:      []string{"", "", "", "", "", "", "", "Notes: -123-456"},
			expected: []uint64{123, 456},
			wantErr:  false,
		},
		{
			name:     "Extra spaces",
			row:      []string{"", "", "", "", "", "", "", "Notes:  123 - 456 - 789 "},
			expected: []uint64{123, 456, 789},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertNoteIDs(tt.row)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, got)
		})
	}
}
func TestConvertFlightLevelToAltitude(t *testing.T) {
	tests := []struct {
		name     string
		row      []string
		index    int
		expected *uint64
		wantErr  bool
	}{
		{
			name:     "Valid flight level",
			row:      []string{"", "", "350"},
			index:    2,
			expected: func() *uint64 { v := uint64(35000); return &v }(),
			wantErr:  false,
		},
		{
			name:     "Empty flight level",
			row:      []string{"", "", ""},
			index:    2,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "MC flight level",
			row:      []string{"", "", "MC"},
			index:    2,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "Invalid flight level",
			row:      []string{"", "", "ABC"},
			index:    2,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Leading and trailing spaces",
			row:      []string{"", "", " 350 "},
			index:    2,
			expected: func() *uint64 { v := uint64(35000); return &v }(),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertFlightLevelToAltitude(tt.row, tt.index)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, got)
		})
	}
}
func TestMapRoute(t *testing.T) {
	tests := []struct {
		name     string
		row      []string
		expected *route.Route
		wantErr  bool
	}{
		{
			name: "Valid route",
			row:  []string{"EGLL", "SID1", "350", "370", "SEGMENT", "STAR1", "EGKK", "Notes: 123-456"},
			expected: route.NewRoute(
				"EGLL",
				func() *string { v := "SID1"; return &v }(),
				func() *uint64 { v := uint64(35000); return &v }(),
				func() *uint64 { v := uint64(37000); return &v }(),
				func() *string { v := "SEGMENT"; return &v }(),
				func() *string { v := "STAR1"; return &v }(),
				"EGKK",
				[]uint64{123, 456},
			),
			wantErr: false,
		},
		{
			name:     "Invalid number of fields",
			row:      []string{"EGLL", "SID1", "350", "370", "SEGMENT", "STAR1", "EGKK"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Missing ADEP or Entry",
			row:      []string{"", "SID1", "350", "370", "SEGMENT", "STAR1", "EGKK", "Notes: 123-456"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid min flight level",
			row:      []string{"EGLL", "SID1", "ABC", "370", "SEGMENT", "STAR1", "EGKK", "Notes: 123-456"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid max flight level",
			row:      []string{"EGLL", "SID1", "350", "ABC", "SEGMENT", "STAR1", "EGKK", "Notes: 123-456"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Missing ADES or Exit",
			row:      []string{"EGLL", "SID1", "350", "370", "SEGMENT", "STAR1", "", "Notes: 123-456"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid note ID",
			row:      []string{"EGLL", "SID1", "350", "370", "SEGMENT", "STAR1", "EGKK", "Notes: 123-abc"},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Optional fields empty",
			row:  []string{"EGLL", "", "350", "370", "", "", "EGKK", ""},
			expected: route.NewRoute(
				"EGLL",
				nil,
				func() *uint64 { v := uint64(35000); return &v }(),
				func() *uint64 { v := uint64(37000); return &v }(),
				nil,
				nil,
				"EGKK",
				nil,
			),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapRoute(tt.row)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, got)
		})
	}
}
