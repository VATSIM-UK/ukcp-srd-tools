package airac

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	clockLib "github.com/benbjohnson/clock"
)

const AiracInterval = time.Duration(AiracIntervalDays * time.Hour * 24)
const AiracIntervalDays = 28

var BaseAiracDate = time.Date(2021, time.January, 28, 0, 0, 0, 0, time.UTC)
var AiracCycleRegexp = regexp.MustCompile(`^(\d{2})(\d{2})$`)

var (
	ErrInvalidAiracIdent = fmt.Errorf("invalid AIRAC cycle identifier")
)

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

// CycleFromIdent returns the AIRAC cycle from an identifier
func (a *Airac) CycleFromIdent(ident string) (*AiracCycle, error) {
	matches := AiracCycleRegexp.FindStringSubmatch(ident)
	if matches == nil {
		return nil, ErrInvalidAiracIdent
	}

	// The cycle must be between 1 and 13, convert the string to an integer
	cycle, _ := strconv.Atoi(matches[2])
	if cycle < 1 || cycle > 13 {
		return nil, ErrInvalidAiracIdent
	}

	// Convert the year to an integer
	year, _ := strconv.Atoi(matches[1])

	// The full year is 2000 + the first two digits of the year
	cycleYear := 2000 + year

	return a.nthCycleOfYear(cycleYear, cycle), nil
}

func (a *Airac) nthCycleOfYear(year, cycle int) *AiracCycle {
	// Get the first airac day of the year
	firstAirac := a.firstAiracDateOfYear(time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC))

	// The nth cycle is adding the airac interval to the first airac day of the year
	start := firstAirac.Add(time.Duration((cycle-1)*AiracIntervalDays) * time.Hour * 24)

	return &AiracCycle{
		Ident: formatAiracIdent(cycle, year),
		Start: start,
		End:   start.Add(AiracInterval),
	}
}

func (a *Airac) daysSinceBase(date time.Time) int {
	// Calculate the number of days since the base AIRAC date
	return int(date.Sub(BaseAiracDate).Hours() / 24)
}

func (a *Airac) nextAiracFromDate(date time.Time) *AiracCycle {
	nextAiracStart := a.nextAiracDateFromDate(date)

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

func (a *Airac) nextAiracDateFromDate(date time.Time) time.Time {
	daysIntoCycle := a.daysSinceBase(date) % AiracIntervalDays
	return date.Add(AiracInterval - (time.Duration(daysIntoCycle) * time.Hour * 24))
}

func (a *Airac) firstAiracDateOfYear(start time.Time) time.Time {
	// Get the last day of the previous year
	lastDayOfPreviousYear := time.Date(start.Year()-1, time.December, 31, 0, 0, 0, 0, time.UTC)

	// Now get the next airac day
	return a.nextAiracDateFromDate(lastDayOfPreviousYear)
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
