// Package period computes reporting date ranges (daily, weekly, monthly,
// quarterly, and named shortcuts like "yesterday" or "month-to-date") in a
// tenant's own timezone, then expresses the result as a UTC [Start, End)
// range for querying storage.
//
// Every stored timestamp is UTC (see shop_docs/docs/database-model.md). A
// tenant's "today" only makes sense in their timezone, so period boundaries
// are always computed there and converted to UTC last, not the other way
// around.
package period

import (
	"fmt"
	"time"
)

// Range is a half-open UTC interval [Start, End) — Start is inclusive, End
// is exclusive, so adjacent periods never double-count a boundary instant.
type Range struct {
	Start time.Time
	End   time.Time
	Label string // human-readable, e.g. "Sep 11" or "Q3 2026"
}

// Kind identifies a named period shortcut.
type Kind string

const (
	Today           Kind = "today"
	Yesterday       Kind = "yesterday"
	WeekToDate      Kind = "week_to_date"
	PreviousWeek    Kind = "previous_week"
	MonthToDate     Kind = "month_to_date"
	PreviousMonth   Kind = "previous_month"
	QuarterToDate   Kind = "quarter_to_date"
	PreviousQuarter Kind = "previous_quarter"
	YearToDate      Kind = "year_to_date"
	Daily           Kind = "daily"
	Weekly          Kind = "weekly"
	Monthly         Kind = "monthly"
	Quarterly       Kind = "quarterly"
)

// Resolve computes a Range for the given Kind, anchored at `now` (evaluated
// in loc). `now` is normally time.Now(), passed explicitly so callers (and
// tests) don't depend on the wall clock.
//
// Daily/Weekly/Monthly/Quarterly resolve to the period containing `now` —
// e.g. Weekly returns the Monday-to-Monday week `now` falls in. Callers
// wanting an arbitrary explicit range should use NewRange directly instead
// of going through Resolve.
func Resolve(kind Kind, now time.Time, loc *time.Location) (Range, error) {
	local := now.In(loc)

	switch kind {
	case Today, Daily:
		start := startOfDay(local)
		return newDayRange(start, dayLabel(start)), nil

	case Yesterday:
		start := startOfDay(local).AddDate(0, 0, -1)
		return newDayRange(start, dayLabel(start)), nil

	case WeekToDate:
		start := startOfWeek(local)
		return Range{Start: start.UTC(), End: local.UTC(), Label: fmt.Sprintf("Week to date (%s)", dayLabel(start))}, nil

	case PreviousWeek, Weekly:
		start := startOfWeek(local).AddDate(0, 0, -7)
		end := start.AddDate(0, 0, 7)
		return Range{Start: start.UTC(), End: end.UTC(), Label: fmt.Sprintf("Week of %s", dayLabel(start))}, nil

	case MonthToDate:
		start := startOfMonth(local)
		return Range{Start: start.UTC(), End: local.UTC(), Label: fmt.Sprintf("Month to date (%s)", monthLabel(start))}, nil

	case PreviousMonth, Monthly:
		start := startOfMonth(local).AddDate(0, -1, 0)
		end := startOfMonth(local)
		return Range{Start: start.UTC(), End: end.UTC(), Label: monthLabel(start)}, nil

	case QuarterToDate:
		start := startOfQuarter(local)
		return Range{Start: start.UTC(), End: local.UTC(), Label: fmt.Sprintf("Quarter to date (%s)", quarterLabel(start))}, nil

	case PreviousQuarter, Quarterly:
		start := startOfQuarter(local).AddDate(0, -3, 0)
		end := startOfQuarter(local)
		return Range{Start: start.UTC(), End: end.UTC(), Label: quarterLabel(start)}, nil

	case YearToDate:
		start := time.Date(local.Year(), 1, 1, 0, 0, 0, 0, loc)
		return Range{Start: start.UTC(), End: local.UTC(), Label: fmt.Sprintf("Year to date (%d)", local.Year())}, nil

	default:
		return Range{}, fmt.Errorf("period: unknown kind %q", kind)
	}
}

// NewRange builds an arbitrary explicit range from tenant-local start/end
// dates (inclusive start day, inclusive end day), converting to UTC.
func NewRange(startDay, endDay time.Time, loc *time.Location) Range {
	start := startOfDay(startDay.In(loc))
	end := startOfDay(endDay.In(loc)).AddDate(0, 0, 1) // make the end day inclusive
	return Range{
		Start: start.UTC(),
		End:   end.UTC(),
		Label: fmt.Sprintf("%s to %s", dayLabel(start), dayLabel(end.AddDate(0, 0, -1))),
	}
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func newDayRange(start time.Time, label string) Range {
	return Range{Start: start.UTC(), End: start.AddDate(0, 0, 1).UTC(), Label: label}
}

// startOfWeek returns the most recent Monday at or before t, at midnight.
func startOfWeek(t time.Time) time.Time {
	day := startOfDay(t)
	// time.Monday == 1; Sunday == 0. Compute days since Monday.
	offset := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -offset)
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func startOfQuarter(t time.Time) time.Time {
	quarterStartMonth := ((int(t.Month())-1)/3)*3 + 1
	return time.Date(t.Year(), time.Month(quarterStartMonth), 1, 0, 0, 0, 0, t.Location())
}

func dayLabel(t time.Time) string   { return t.Format("Jan 2") }
func monthLabel(t time.Time) string { return t.Format("January 2006") }
func quarterLabel(t time.Time) string {
	q := (int(t.Month())-1)/3 + 1
	return fmt.Sprintf("Q%d %d", q, t.Year())
}
