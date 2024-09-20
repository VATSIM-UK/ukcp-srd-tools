package airac

import (
	"fmt"
	"time"

	clockLib "github.com/benbjohnson/clock"
)

const AiracInterval = time.Duration(AiracIntervalDays * time.Hour * 24)
const AiracIntervalDays = 28

var BaseAiracDate = time.Date(2021, time.January, 28, 0, 0, 0, 0, time.UTC)

type Airac struct {
	clock clockLib.Clock
}

func NewAirac(clock clockLib.Clock) *Airac {
	if clock == nil {
		clock = clockLib.New()
	}

	return &Airac{
		clock,
	}
}

type AiracCycle struct {
	Ident string
	Start time.Time
	End   time.Time
}

// GetCurrentAirac returns the current AIRAC cycle
func (a *Airac) CurrentCycle() *AiracCycle {
	return a.nextAiracFromDate(a.previousAiracDayFromDate(a.currentDay()).Add(-time.Hour * 24))
}

// NextCycle returns the next AIRAC cycle
func (a *Airac) NextCycle() *AiracCycle {
	return a.nextAiracFromDate(a.currentDay())
}

// NextCycleFrom returns the next AIRAC cycle from a cycle
func (a *Airac) NextCycleFrom(cycle *AiracCycle) *AiracCycle {
	return a.nextAiracFromDate(cycle.Start)
}

func (a *Airac) daysSinceBase(date time.Time) int {
	// Calculate the number of days since the base AIRAC date
	return int(date.Sub(BaseAiracDate).Hours() / 24)
}

func (a *Airac) nextAiracFromDate(date time.Time) *AiracCycle {
	// Calculate the start of the next AIRAC cycle
	daysIntoCycle := a.daysSinceBase(date) % AiracIntervalDays
	nextAiracStart := date.Add(AiracInterval - (time.Duration(daysIntoCycle) * time.Hour * 24))

	return &AiracCycle{
		Ident: a.identFromStartDate(nextAiracStart),
		Start: nextAiracStart,
		End:   nextAiracStart.Add(AiracInterval),
	}
}

func (a *Airac) identFromStartDate(start time.Time) string {
	// Get the first airac of the year
	firstAirac := a.firstAiracDateOfYear(start)

	// If our start date is before the first airac of the year, then we are in the previous year, and it's a 13 cycle year
	if start.Before(firstAirac) {
		return formatAiracIdent(13, start.Year()-1)
	}

	// Calculate the number of days since the first airac of the year, and add 1 to get the cycle number
	daysSinceFirst := start.Sub(firstAirac).Hours() / 24
	cycleNumber := int(daysSinceFirst/AiracIntervalDays) + 1

	return formatAiracIdent(cycleNumber, start.Year())
}

func (a *Airac) firstAiracDateOfYear(start time.Time) time.Time {
	// Get the start of the year date
	return time.Date(start.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
}

func (a *Airac) previousAiracDayFromDate(date time.Time) time.Time {
	daysSinceBase := a.daysSinceBase(date)
	daysIntoCycle := daysSinceBase % AiracIntervalDays

	return date.Add(-time.Duration(daysIntoCycle) * time.Hour * 24)
}

func (a *Airac) currentDay() time.Time {
	return startOfDay(a.clock.Now())
}

// formatAiracIdent formats the AIRAC cycle identifier
// Its the last two digits of the year followed by the cycle number, which is padded with a leading zero
// For example 2024 1st cycle would be 2401
func formatAiracIdent(cycleNumber, year int) string {
	// Get the last two digits of the year
	yearIdent := year % 100

	// Format the AIRAC cycle identifier
	return fmt.Sprintf("%02d%02d", yearIdent, cycleNumber)
}

func startOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}
