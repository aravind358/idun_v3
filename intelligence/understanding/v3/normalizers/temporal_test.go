package normalizers

import (
	coretime "idun/core/time"
	"idun/intelligence/understanding/v3"
	"idun/intelligence/understanding/v3/ontology"
	"testing"
	"time"
)

func TestTemporalNormalizer(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	timeSvc := coretime.NewTimeService(loc)
	norm := NewDeterministicTemporalNormalizer(timeSvc)

	// Mock current time by overriding today in a real test we'd mock TimeService,
	// but here we just compute expected dates dynamically using the same timeSvc.
	today := timeSvc.StartOfDay(timeSvc.Now())

	tests := []struct {
		surface  string
		tType    ontology.TemporalType
		expected string
	}{
		{"tomorrow", ontology.TempRelativeDate, timeSvc.Format(timeSvc.AddDays(today, 1), time.DateOnly)},
		{"today", ontology.TempRelativeDate, timeSvc.Format(today, time.DateOnly)},
		{"5 PM", ontology.TempClockTime, "17:00"},
		{"noon", ontology.TempClockTime, "12:00"},
		{"next friday", ontology.TempRelativeWeekday, timeSvc.Format(timeSvc.NextWeekday(today, time.Friday), time.DateOnly)},
		{"later", ontology.TempRelativeDate, ""}, // Ambiguous
		{"in 3 hours", ontology.TempRelativeDuration, "PT3H"},
		{"morning", ontology.TempDaypart, "[06:00,12:00)"},
		{"2026-08-04", ontology.TempAbsoluteDate, "2026-08-04"},
		{"August 4", ontology.TempAbsoluteDate, timeSvc.Format(time.Date(timeSvc.Now().Year(), time.August, 4, 0, 0, 0, 0, loc), time.DateOnly)},
	}

	for _, tc := range tests {
		anchor := v3.NewTemporalAnchor(tc.surface, tc.tType, "", 1.0)
		normalized := norm.Normalize([]v3.TemporalAnchor{anchor})

		if len(normalized) != 1 {
			t.Fatalf("Normalize() returned %d anchors, expected 1", len(normalized))
		}

		res := normalized[0]
		if res.Surface() != tc.surface {
			t.Errorf("Normalize() modified surface: got %q, want %q", res.Surface(), tc.surface)
		}
		if res.Type() != tc.tType {
			t.Errorf("Normalize() modified type: got %v, want %v", res.Type(), tc.tType)
		}
		if res.Normalized() != tc.expected {
			t.Errorf("Normalize(%q) normalized value wrong: got %q, want %q", tc.surface, res.Normalized(), tc.expected)
		}
	}
}
