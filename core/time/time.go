package time

import (
	"fmt"
	"strings"
	"time"
)

// TimeService provides foundational temporal operations for the entire system.
// It serves as a mechanical abstraction layer to ensure accurate time retrieval,
// parsing, and arithmetic without imposing cognitive logic or conversational behavior.
type TimeService interface {
	// Clock
	Now() time.Time
	Location() *time.Location

	// Calendar
	Today() time.Time

	// Parsing
	Parse(input string) (time.Time, error)
	ParseWithLayout(layout, input string) (time.Time, error)

	// Formatting
	Format(t time.Time, layout string) string

	// Arithmetic
	Add(t time.Time, d time.Duration) time.Time
	AddDuration(t time.Time, d time.Duration) time.Time
	AddDays(t time.Time, days int) time.Time
	Since(t time.Time) time.Duration
	Until(t time.Time) time.Duration
	NextWeekday(t time.Time, wd time.Weekday) time.Time
	StartOfDay(t time.Time) time.Time
	EndOfDay(t time.Time) time.Time

	// Timezone
	Timezone() *time.Location

	// Specialized Parsing
	ParseClock(input string) (time.Time, error)

	// Utility
	IsExpired(t time.Time) bool
}

var standardLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC822,
	time.RFC822Z,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.Kitchen,
	time.DateOnly,
	time.DateTime,
	time.TimeOnly,
}

type systemTimeService struct {
	loc *time.Location
}

// NewTimeService initializes a new TimeService using the provided location.
// If loc is nil, time.Local is used.
func NewTimeService(loc *time.Location) TimeService {
	if loc == nil {
		loc = time.Local
	}
	return &systemTimeService{
		loc: loc,
	}
}

func (s *systemTimeService) Now() time.Time {
	return time.Now().In(s.loc)
}

func (s *systemTimeService) Location() *time.Location {
	return s.loc
}

func (s *systemTimeService) Today() time.Time {
	now := s.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.loc)
}

func (s *systemTimeService) Parse(input string) (time.Time, error) {
	// Try parsing against standard layouts
	for _, layout := range standardLayouts {
		if t, err := time.ParseInLocation(layout, input, s.loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("core/time: failed to parse %q against standard layouts", input)
}

func (s *systemTimeService) ParseWithLayout(layout, input string) (time.Time, error) {
	return time.ParseInLocation(layout, input, s.loc)
}

func (s *systemTimeService) Format(t time.Time, layout string) string {
	return t.In(s.loc).Format(layout)
}

func (s *systemTimeService) Add(t time.Time, d time.Duration) time.Time {
	return t.Add(d)
}

func (s *systemTimeService) AddDuration(t time.Time, d time.Duration) time.Time {
	return t.Add(d)
}

func (s *systemTimeService) AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

func (s *systemTimeService) Since(t time.Time) time.Duration {
	return s.Now().Sub(t)
}

func (s *systemTimeService) Until(t time.Time) time.Duration {
	return t.Sub(s.Now())
}

func (s *systemTimeService) IsExpired(t time.Time) bool {
	return s.Now().After(t)
}

func (s *systemTimeService) NextWeekday(t time.Time, wd time.Weekday) time.Time {
	days := (int(wd) - int(t.Weekday()) + 7) % 7
	if days == 0 {
		days = 7 // "next Friday" when today is Friday means next week
	}
	return t.AddDate(0, 0, days)
}

func (s *systemTimeService) StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (s *systemTimeService) EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

func (s *systemTimeService) Timezone() *time.Location {
	return s.loc
}

func (s *systemTimeService) ParseClock(input string) (time.Time, error) {
	// Simple clock parser for 5 PM, 17:30, noon, midnight
	layouts := []string{
		"3 PM",
		"3:04 PM",
		"15:04",
	}
	
	lowerInput := strings.ToLower(input)
	if lowerInput == "noon" {
		return time.Date(0, 1, 1, 12, 0, 0, 0, s.loc), nil
	}
	if lowerInput == "midnight" {
		return time.Date(0, 1, 1, 0, 0, 0, 0, s.loc), nil
	}
	
	upperInput := strings.ToUpper(input)
	for _, layout := range layouts {
		if t, err := time.Parse(layout, upperInput); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("core/time: failed to parse clock time %q", input)
}
