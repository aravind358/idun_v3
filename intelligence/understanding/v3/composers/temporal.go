package composers

import (
	"fmt"
	"idun/intelligence/understanding/v3"
	"idun/intelligence/understanding/v3/ontology"
)

type deterministicTemporalComposer struct {
}

func NewDeterministicTemporalComposer() Runner {
	return &deterministicTemporalComposer{}
}

func (c *deterministicTemporalComposer) Run(b *v3.Builder) {
	anchors := b.GetTemporalAnchors()
	
	// Fast path: if 0 or 1 anchor, nothing to compose.
	if len(anchors) < 2 {
		return
	}

	var composed []string
	var currentDate string

	for _, a := range anchors {
		if a.Normalized() == "" {
			continue
		}
		
		t := a.Type()
		if t == ontology.TempRelativeDate || t == ontology.TempRelativeWeekday || t == ontology.TempAbsoluteDate {
			currentDate = a.Normalized()
		} else if t == ontology.TempClockTime {
			if currentDate != "" {
				c := fmt.Sprintf("%sT%s:00Z", currentDate, a.Normalized())
				composed = append(composed, c)
				currentDate = "" // Reset to avoid reusing the same date for the next time without a new date
			}
		}
	}

	if len(composed) > 0 {
		existingComposed := b.GetComposedTimestamps()
		existingComposed = append(existingComposed, composed...)
		b.ComposedTimestamps(existingComposed)
	}
}
