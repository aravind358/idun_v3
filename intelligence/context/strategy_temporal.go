package context

import (
	"context"
	"time"

	underv3 "idun/intelligence/understanding/v3"
	"idun/intelligence/understanding/v3/ontology"
)

// TemporalStrategy resolves relative time references (e.g. tomorrow, next week) using the dialogue state's temporal anchor.
type TemporalStrategy struct{}

func (s *TemporalStrategy) Execute(ctx context.Context, orig *underv3.SemanticInterpretation, builder *underv3.Builder, state DialogueStateReader, resolvedEntities map[string]string) (bool, ResolutionStatus) {
	handled := false
	status := StatusContextUnnecessary

	anchor := state.GetTemporalAnchor()
	if anchor.IsZero() {
		// No temporal anchor available, fallback to now if we must, but generally we expect state to provide one.
		anchor = time.Now()
	}

	anchors := orig.TemporalAnchors()
	var newAnchors []underv3.TemporalAnchor

	for _, a := range anchors {
		if a.Type() == ontology.TempRelative || a.Type() == ontology.TempRelativeDate || a.Type() == ontology.TempRelativeWeekday {
			handled = true
			
			// For U7.5, we just mark that we "anchored" the time.
			// Actual relative time math goes to Reasoning.
			normalizedStr := anchor.Format(time.RFC3339)
			resolvedEntities["temporal_anchor"] = normalizedStr
			
			newAnchor := underv3.NewTemporalAnchor(
				a.Surface(), a.Type(), normalizedStr, a.Confidence(),
			)
			newAnchors = append(newAnchors, newAnchor)
			status = StatusResolved
		} else {
			newAnchors = append(newAnchors, a)
		}
	}

	if handled {
		builder.TemporalAnchors(newAnchors)
	}

	return handled, status
}
