package extractors

import (
	"idun/intelligence/understanding/v3"
)

// EntityExtractor parses a hypothesis into semantic entities.
type EntityExtractor interface {
	Extract(hyp v3.Hypothesis) []v3.Entity
}

// ReferenceExtractor parses a hypothesis into reference anchors.
type ReferenceExtractor interface {
	Extract(hyp v3.Hypothesis) []v3.Reference
}

// TemporalExtractor parses a hypothesis into temporal anchors.
type TemporalExtractor interface {
	Extract(hyp v3.Hypothesis) []v3.TemporalAnchor
}
