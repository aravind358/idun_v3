package extractors

import (
	"strings"

	"idun/intelligence/understanding/v3"
	"idun/intelligence/understanding/v3/ontology"
)

type temporalExtractor struct{}

func NewTemporalExtractor() *temporalExtractor {
	return &temporalExtractor{}
}

func (e *temporalExtractor) Extract(hyp v3.Hypothesis) []v3.TemporalAnchor {
	var anchors []v3.TemporalAnchor
	for _, slot := range hyp.Slots() {
		// Identify temporal boundaries and decompose into components.
		// The exact decomposition strategy is an internal detail.
		components := decomposeTemporalPhrase(slot.Value())
		for _, comp := range components {
			switch slot.Name() {
			case "date":
				tType := classifyDate(comp)
				anchors = append(anchors, v3.NewTemporalAnchor(comp, tType, "", slot.Confidence()))
			case "time":
				tType := classifyTime(comp)
				anchors = append(anchors, v3.NewTemporalAnchor(comp, tType, "", slot.Confidence()))
			case "duration":
				anchors = append(anchors, v3.NewTemporalAnchor(comp, ontology.TempRelativeDuration, "", slot.Confidence()))
			case "daypart":
				anchors = append(anchors, v3.NewTemporalAnchor(comp, ontology.TempDaypart, "", slot.Confidence()))
			case "datetime":
				anchors = append(anchors, v3.NewTemporalAnchor(comp, ontology.TempAbsoluteDate, "", slot.Confidence()))
			case "day":
				anchors = append(anchors, v3.NewTemporalAnchor(comp, ontology.TempRelativeWeekday, "", slot.Confidence()))
			}
		}
	}
	return anchors
}

func decomposeTemporalPhrase(phrase string) []string {
	phrase = strings.TrimSpace(phrase)
	lower := strings.ToLower(phrase)
	
	// Check for duration context (keep as single component)
	if strings.HasPrefix(lower, "in ") || strings.Contains(lower, "hour") || strings.Contains(lower, "minute") {
		return []string{phrase}
	}
	
	// Trim prefix prepositions
	if strings.HasPrefix(lower, "on ") {
		phrase = phrase[3:]
	}

	// Simple heuristic delimiter splitting
	var parts []string
	
	// Split by " and " first
	andParts := strings.Split(strings.ToLower(phrase), " and ")
	for _, p := range andParts {
		// Then split by " at "
		if idx := strings.Index(p, " at "); idx != -1 {
			parts = append(parts, strings.TrimSpace(p[:idx]))
			parts = append(parts, strings.TrimSpace(p[idx+4:]))
		} else {
			parts = append(parts, strings.TrimSpace(p))
		}
	}

	// Filter empty
	var final []string
	for _, p := range parts {
		if p != "" {
			final = append(final, p)
		}
	}
	return final
}

func classifyTime(surface string) ontology.TemporalType {
	lower := strings.ToLower(strings.TrimSpace(surface))
	if strings.HasPrefix(lower, "in ") || strings.Contains(lower, "hour") || strings.Contains(lower, "minute") {
		return ontology.TempRelativeDuration
	}
	tType := classifyDate(surface)
	if tType == ontology.TempAbsoluteDate {
		// If classifyDate fell through to absolute date, it's likely a clock time like "5 PM"
		return ontology.TempClockTime
	}
	return tType
}

func classifyDate(surface string) ontology.TemporalType {
	switch surface {
	case "today", "tomorrow", "yesterday":
		return ontology.TempRelativeDate
	case "next week", "this morning", "tonight", "next month":
		return ontology.TempTimeInterval
	case "later", "soon", "now":
		return ontology.TempUnknown
	default:
		// Check for weekday patterns
		if containsWeekday(surface) {
			return ontology.TempRelativeWeekday
		}
		return ontology.TempAbsoluteDate
	}
}

func containsWeekday(s string) bool {
	lower := strings.ToLower(s)
	weekdays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, wd := range weekdays {
		if strings.Contains(lower, wd) {
			return true
		}
	}
	return false
}
