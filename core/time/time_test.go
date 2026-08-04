package time

import (
	"testing"
	"time"
)

func TestTimeService_Now(t *testing.T) {
	svc := NewTimeService(nil)
	now := svc.Now()
	if now.IsZero() {
		t.Errorf("Now() returned zero time")
	}
	
	if now.Location().String() != time.Local.String() {
		t.Errorf("Now() expected location %v, got %v", time.Local, now.Location())
	}
}

func TestTimeService_Location(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	svc := NewTimeService(loc)
	if svc.Location() != loc {
		t.Errorf("Location() expected %v, got %v", loc, svc.Location())
	}
}

func TestTimeService_Today(t *testing.T) {
	svc := NewTimeService(time.UTC)
	today := svc.Today()
	
	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 || today.Nanosecond() != 0 {
		t.Errorf("Today() should have 00:00:00 time, got %v", today)
	}
	
	if today.Location() != time.UTC {
		t.Errorf("Today() expected location UTC, got %v", today.Location())
	}
}

func TestTimeService_Parse(t *testing.T) {
	svc := NewTimeService(time.UTC)
	
	validInputs := []string{
		"2023-10-25T15:30:00Z",             // RFC3339
		"Wed, 25 Oct 2023 15:30:00 GMT",    // RFC1123
		"2023-10-25",                       // DateOnly
	}
	
	for _, input := range validInputs {
		parsed, err := svc.Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", input, err)
		}
		if parsed.IsZero() {
			t.Errorf("Parse(%q) returned zero time", input)
		}
	}
	
	_, err := svc.Parse("invalid-date-format")
	if err == nil {
		t.Errorf("Parse() expected error for invalid input, got nil")
	}
}

func TestTimeService_ParseWithLayout(t *testing.T) {
	svc := NewTimeService(time.UTC)
	
	layout := "2006/01/02 15:04"
	input := "2023/10/25 15:30"
	
	parsed, err := svc.ParseWithLayout(layout, input)
	if err != nil {
		t.Errorf("ParseWithLayout() failed: %v", err)
	}
	
	if parsed.Year() != 2023 || parsed.Month() != 10 || parsed.Day() != 25 {
		t.Errorf("ParseWithLayout() parsed incorrect date: %v", parsed)
	}
}

func TestTimeService_Format(t *testing.T) {
	svc := NewTimeService(time.UTC)
	ts := time.Date(2023, 10, 25, 15, 30, 0, 0, time.UTC)
	
	formatted := svc.Format(ts, time.DateOnly)
	if formatted != "2023-10-25" {
		t.Errorf("Format() expected %q, got %q", "2023-10-25", formatted)
	}
}

func TestTimeService_Arithmetic(t *testing.T) {
	svc := NewTimeService(time.UTC)
	now := svc.Now()
	
	future := svc.Add(now, 2*time.Hour)
	if future.Sub(now) != 2*time.Hour {
		t.Errorf("Add() failed, expected difference of 2h, got %v", future.Sub(now))
	}
	
	// Test Since and Until
	past := now.Add(-1 * time.Hour)
	since := svc.Since(past)
	if since < time.Hour || since > time.Hour+1*time.Second {
		t.Errorf("Since() returned unexpected duration: %v", since)
	}
	
	future2 := now.Add(1 * time.Hour)
	until := svc.Until(future2)
	if until < time.Hour-1*time.Second || until > time.Hour {
		t.Errorf("Until() returned unexpected duration: %v", until)
	}
}

func TestTimeService_IsExpired(t *testing.T) {
	svc := NewTimeService(time.UTC)
	now := svc.Now()
	
	past := now.Add(-1 * time.Hour)
	if !svc.IsExpired(past) {
		t.Errorf("IsExpired() expected true for past time")
	}
	
	future := now.Add(1 * time.Hour)
	if svc.IsExpired(future) {
		t.Errorf("IsExpired() expected false for future time")
	}
}

func TestTimeService_CalendarMath(t *testing.T) {
	svc := NewTimeService(time.UTC)
	now := svc.Now()

	// Test AddDays
	tomorrow := svc.AddDays(now, 1)
	if tomorrow.Day() == now.Day() {
		t.Errorf("AddDays() failed, day did not advance")
	}

	// Test NextWeekday
	nextFri := svc.NextWeekday(now, time.Friday)
	if nextFri.Weekday() != time.Friday {
		t.Errorf("NextWeekday() expected Friday, got %v", nextFri.Weekday())
	}
	if nextFri.Before(now) {
		t.Errorf("NextWeekday() returned a past date")
	}

	// Test StartOfDay
	sod := svc.StartOfDay(now)
	if sod.Hour() != 0 || sod.Minute() != 0 || sod.Second() != 0 {
		t.Errorf("StartOfDay() failed, got %v", sod)
	}

	// Test EndOfDay
	eod := svc.EndOfDay(now)
	if eod.Hour() != 23 || eod.Minute() != 59 || eod.Second() != 59 {
		t.Errorf("EndOfDay() failed, got %v", eod)
	}
}

func TestTimeService_ParseClock(t *testing.T) {
	svc := NewTimeService(time.UTC)

	tests := []struct {
		input string
		hour  int
		min   int
	}{
		{"5 PM", 17, 0},
		{"17:30", 17, 30},
		{"noon", 12, 0},
		{"midnight", 0, 0},
		{"3:04 PM", 15, 4},
	}

	for _, tc := range tests {
		parsed, err := svc.ParseClock(tc.input)
		if err != nil {
			t.Errorf("ParseClock(%q) failed: %v", tc.input, err)
		}
		if parsed.Hour() != tc.hour || parsed.Minute() != tc.min {
			t.Errorf("ParseClock(%q) expected %d:%02d, got %d:%02d", tc.input, tc.hour, tc.min, parsed.Hour(), parsed.Minute())
		}
	}
}
