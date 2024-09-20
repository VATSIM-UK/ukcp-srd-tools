package airac

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	clockLib "github.com/benbjohnson/clock"
)

// Test the CurrentCycle method
func TestCurrentCycle(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		expected *AiracCycle
	}{
		{
			name: "First cycle of the year",
			now:  time.Date(2024, time.January, 25, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2401",
				Start: time.Date(2024, time.January, 25, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.February, 22, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Second cycle of the year",
			now:  time.Date(2024, time.February, 23, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2402",
				Start: time.Date(2024, time.February, 22, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year",
			now:  time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2413",
				Start: time.Date(2024, time.December, 26, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year last second",
			now:  time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2413",
				Start: time.Date(2024, time.December, 26, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year (next year)",
			now:  time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2413",
				Start: time.Date(2024, time.December, 26, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year (next year last second)",
			now:  time.Date(2025, time.January, 22, 23, 59, 59, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2413",
				Start: time.Date(2024, time.December, 26, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "2409",
			now:  time.Date(2024, time.September, 6, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2409",
				Start: time.Date(2024, time.September, 5, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.October, 3, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "2410",
			now:  time.Date(2024, time.October, 3, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2410",
				Start: time.Date(2024, time.October, 3, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.October, 31, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "2506",
			now:  time.Date(2025, time.June, 12, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2506",
				Start: time.Date(2025, time.June, 12, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.July, 10, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			"3004",
			time.Date(2030, time.April, 11, 0, 0, 0, 0, time.UTC),
			&AiracCycle{
				Ident: "3004",
				Start: time.Date(2030, time.April, 11, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2030, time.May, 9, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "2103",
			now:  time.Date(2021, time.March, 25, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2103",
				Start: time.Date(2021, time.March, 25, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2021, time.April, 22, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := clockLib.NewMock()
			clock.Set(tt.now)
			airac := NewAirac(clock)
			actual := airac.CurrentCycle()
			require.Equal(t, tt.expected, actual)
		})
	}
}

// Test the NextCycle method
func TestNextCycle(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		expected *AiracCycle
	}{
		{
			name: "First cycle of the year",
			now:  time.Date(2024, time.January, 25, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2402",
				Start: time.Date(2024, time.February, 22, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Second cycle of the year",
			now:  time.Date(2024, time.February, 23, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2403",
				Start: time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.April, 18, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year",
			now:  time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2501",
				Start: time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.February, 20, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year last second",
			now:  time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2501",
				Start: time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.February, 20, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year (next year)",
			now:  time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2501",
				Start: time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.February, 20, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year (next year last second)",
			now:  time.Date(2025, time.January, 22, 23, 59, 59, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2501",
				Start: time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.February, 20, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "2409",
			now:  time.Date(2024, time.September, 6, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2410",
				Start: time.Date(2024, time.October, 3, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.October, 31, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "2410",
			now:  time.Date(2024, time.October, 3, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2411",
				Start: time.Date(2024, time.October, 31, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.November, 28, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "2506",
			now:  time.Date(2025, time.June, 12, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2507",
				Start: time.Date(2025, time.July, 10, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.August, 7, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "3004",
			now:  time.Date(2030, time.April, 11, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "3005",
				Start: time.Date(2030, time.May, 9, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2030, time.June, 6, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "2103",
			now:  time.Date(2021, time.March, 25, 0, 0, 0, 0, time.UTC),
			expected: &AiracCycle{
				Ident: "2104",
				Start: time.Date(2021, time.April, 22, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2021, time.May, 20, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := clockLib.NewMock()
			clock.Set(tt.now)
			airac := NewAirac(clock)
			actual := airac.NextCycle()
			require.Equal(t, tt.expected, actual)
		})
	}
}

// Test the NextCycleFrom method
func TestNextCycleFrom(t *testing.T) {
	tests := []struct {
		name     string
		cycle    *AiracCycle
		expected *AiracCycle
	}{
		{
			name: "First cycle of the year",
			cycle: &AiracCycle{
				Ident: "2401",
				Start: time.Date(2024, time.January, 25, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.February, 22, 0, 0, 0, 0, time.UTC),
			},
			expected: &AiracCycle{
				Ident: "2402",
				Start: time.Date(2024, time.February, 22, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Second cycle of the year",
			cycle: &AiracCycle{
				Ident: "2402",
				Start: time.Date(2024, time.February, 22, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC),
			},
			expected: &AiracCycle{
				Ident: "2403",
				Start: time.Date(2024, time.March, 21, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.April, 18, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Last cycle of the year",
			cycle: &AiracCycle{
				Ident: "2413",
				Start: time.Date(2024, time.December, 26, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
			},
			expected: &AiracCycle{
				Ident: "2501",
				Start: time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.February, 20, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Middle of the year",
			cycle: &AiracCycle{
				Ident: "2406",
				Start: time.Date(2024, time.June, 13, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.July, 11, 0, 0, 0, 0, time.UTC),
			},
			expected: &AiracCycle{
				Ident: "2407",
				Start: time.Date(2024, time.July, 11, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.August, 8, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "End of the year",
			cycle: &AiracCycle{
				Ident: "2412",
				Start: time.Date(2024, time.November, 28, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, time.December, 26, 0, 0, 0, 0, time.UTC),
			},
			expected: &AiracCycle{
				Ident: "2413",
				Start: time.Date(2024, time.December, 26, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, time.January, 23, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := clockLib.NewMock()
			airac := NewAirac(clock)
			actual := airac.NextCycleFrom(tt.cycle)
			require.Equal(t, tt.expected, actual)
		})
	}
}
