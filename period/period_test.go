package period

import (
	"testing"
	"time"
)

// New York is a good stress-test timezone: it's behind UTC, observes DST,
// and is the actual tenant timezone this platform launches with.
func newYork(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("tzdata not available: %v", err)
	}
	return loc
}

func TestResolve_Today_CrossesUTCDayBoundary(t *testing.T) {
	loc := newYork(t)
	// 11pm Eastern on Sep 11 is already Sep 12 in UTC. "Today" must still
	// mean Sep 11 in the tenant's timezone, not the UTC day.
	now := time.Date(2026, 9, 11, 23, 0, 0, 0, loc)

	r, err := Resolve(Today, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	wantStart := time.Date(2026, 9, 11, 0, 0, 0, 0, loc).UTC()
	wantEnd := time.Date(2026, 9, 12, 0, 0, 0, 0, loc).UTC()
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", r.Start, wantStart)
	}
	if !r.End.Equal(wantEnd) {
		t.Errorf("End = %v, want %v", r.End, wantEnd)
	}
	if r.Label != "Sep 11" {
		t.Errorf("Label = %q, want %q", r.Label, "Sep 11")
	}
}

func TestResolve_Yesterday(t *testing.T) {
	loc := newYork(t)
	now := time.Date(2026, 9, 11, 10, 0, 0, 0, loc)

	r, err := Resolve(Yesterday, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	wantStart := time.Date(2026, 9, 10, 0, 0, 0, 0, loc).UTC()
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", r.Start, wantStart)
	}
}

func TestResolve_WeekToDate_StartsMonday(t *testing.T) {
	loc := newYork(t)
	// Thursday Sep 10, 2026
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, loc)

	r, err := Resolve(WeekToDate, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	wantStart := time.Date(2026, 9, 7, 0, 0, 0, 0, loc).UTC() // Monday Sep 7
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v (Monday)", r.Start, wantStart)
	}
	if !r.End.Equal(now.UTC()) {
		t.Errorf("End = %v, want now (%v)", r.End, now.UTC())
	}
}

func TestResolve_PreviousWeek_IsFullSevenDays(t *testing.T) {
	loc := newYork(t)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, loc) // Thursday

	r, err := Resolve(PreviousWeek, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got := r.End.Sub(r.Start); got != 7*24*time.Hour {
		t.Errorf("previous week span = %v, want 7 days", got)
	}
	wantStart := time.Date(2026, 8, 31, 0, 0, 0, 0, loc).UTC() // Monday before this week
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", r.Start, wantStart)
	}
}

func TestResolve_MonthToDate(t *testing.T) {
	loc := newYork(t)
	now := time.Date(2026, 9, 15, 8, 30, 0, 0, loc)

	r, err := Resolve(MonthToDate, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	wantStart := time.Date(2026, 9, 1, 0, 0, 0, 0, loc).UTC()
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", r.Start, wantStart)
	}
}

func TestResolve_PreviousMonth_HandlesYearBoundary(t *testing.T) {
	loc := newYork(t)
	now := time.Date(2026, 1, 15, 0, 0, 0, 0, loc)

	r, err := Resolve(PreviousMonth, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	wantStart := time.Date(2025, 12, 1, 0, 0, 0, 0, loc).UTC()
	wantEnd := time.Date(2026, 1, 1, 0, 0, 0, 0, loc).UTC()
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", r.Start, wantStart)
	}
	if !r.End.Equal(wantEnd) {
		t.Errorf("End = %v, want %v", r.End, wantEnd)
	}
}

func TestResolve_QuarterToDate(t *testing.T) {
	loc := newYork(t)
	// Aug 15 is in Q3 (Jul-Sep)
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, loc)

	r, err := Resolve(QuarterToDate, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	wantStart := time.Date(2026, 7, 1, 0, 0, 0, 0, loc).UTC()
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", r.Start, wantStart)
	}
}

func TestResolve_PreviousQuarter_HandlesYearBoundary(t *testing.T) {
	loc := newYork(t)
	// Feb 2026 is in Q1 2026; previous quarter is Q4 2025.
	now := time.Date(2026, 2, 10, 0, 0, 0, 0, loc)

	r, err := Resolve(PreviousQuarter, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	wantStart := time.Date(2025, 10, 1, 0, 0, 0, 0, loc).UTC()
	wantEnd := time.Date(2026, 1, 1, 0, 0, 0, 0, loc).UTC()
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", r.Start, wantStart)
	}
	if !r.End.Equal(wantEnd) {
		t.Errorf("End = %v, want %v", r.End, wantEnd)
	}
}

func TestResolve_YearToDate(t *testing.T) {
	loc := newYork(t)
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, loc)

	r, err := Resolve(YearToDate, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	wantStart := time.Date(2026, 1, 1, 0, 0, 0, 0, loc).UTC()
	if !r.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", r.Start, wantStart)
	}
}

func TestResolve_UnknownKind(t *testing.T) {
	if _, err := Resolve(Kind("bogus"), time.Now(), time.UTC); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

func TestNewRange_InclusiveEndDay(t *testing.T) {
	loc := newYork(t)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, loc)
	end := time.Date(2026, 9, 5, 0, 0, 0, 0, loc)

	r := NewRange(start, end, loc)
	wantEnd := time.Date(2026, 9, 6, 0, 0, 0, 0, loc).UTC() // exclusive end = day after Sep 5
	if !r.End.Equal(wantEnd) {
		t.Errorf("End = %v, want %v (Sep 5 inclusive)", r.End, wantEnd)
	}
}

func TestResolve_DailyAcrossDSTSpringForward(t *testing.T) {
	loc := newYork(t)
	// 2026-03-08 is the US DST transition (2am -> 3am). A naive 24h
	// assumption would be wrong; startOfDay/AddDate must handle this via
	// the Location's wall-clock arithmetic, not fixed-duration math.
	now := time.Date(2026, 3, 8, 12, 0, 0, 0, loc)

	r, err := Resolve(Today, now, loc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	span := r.End.Sub(r.Start)
	if span != 23*time.Hour {
		t.Errorf("DST spring-forward day span = %v, want 23h", span)
	}
}
