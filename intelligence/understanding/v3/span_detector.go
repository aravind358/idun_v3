package v3

import (
	"fmt"
	"regexp"
	"strings"
)

// TemporalSpanDetector identifies protected temporal spans in an utterance
// that must not be segmented by the Splitter.
type TemporalSpanDetector struct {
	// A naive pattern list for demonstration. In a production system, this
	// would likely reuse ontology primitives or a specialized NLP tokenizer.
	patterns []*regexp.Regexp
}

// NewTemporalSpanDetector creates a new TemporalSpanDetector.
func NewTemporalSpanDetector() *TemporalSpanDetector {
	return &TemporalSpanDetector{
		patterns: []*regexp.Regexp{
			// "between X and Y"
			regexp.MustCompile(`(?i)\bbetween\s+.+?\s+and\s+.+?\b`),
			// "X and Y" for days of the week
			regexp.MustCompile(`(?i)\b(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday|tomorrow|today|yesterday)\s+and\s+(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday|tomorrow|today|yesterday)\b`),
			// "X and Y" for times (e.g. 3pm and 5pm)
			regexp.MustCompile(`(?i)\b\d{1,2}(?::\d{2})?\s*(?:am|pm)?\s+and\s+\d{1,2}(?::\d{2})?\s*(?:am|pm)?\b`),
		},
	}
}

// ProtectedSpan represents a masked region in the utterance.
type ProtectedSpan struct {
	Placeholder string
	Original    string
}

// MaskSpans finds protected temporal boundaries, replaces them with placeholders,
// and returns the masked utterance along with a map to restore them later.
func (d *TemporalSpanDetector) MaskSpans(utterance string) (string, []ProtectedSpan) {
	masked := utterance
	var spans []ProtectedSpan
	counter := 0

	for _, pattern := range d.patterns {
		matches := pattern.FindAllString(masked, -1)
		for _, match := range matches {
			placeholder := fmt.Sprintf("__TEMP%d__", counter)
			spans = append(spans, ProtectedSpan{
				Placeholder: placeholder,
				Original:    match,
			})
			// Only replace one instance at a time to prevent overlapping bugs if any.
			masked = strings.Replace(masked, match, placeholder, 1)
			counter++
		}
	}

	return masked, spans
}

// RestoreSpans replaces the placeholders in a chunk with their original text.
func (d *TemporalSpanDetector) RestoreSpans(chunk string, spans []ProtectedSpan) string {
	restored := chunk
	for _, span := range spans {
		restored = strings.ReplaceAll(restored, span.Placeholder, span.Original)
	}
	return restored
}
