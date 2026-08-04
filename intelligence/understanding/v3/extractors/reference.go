package extractors

import (
	"strings"

	"idun/intelligence/understanding/v3"
	"idun/intelligence/understanding/v3/ontology"
)

type referenceExtractor struct{}

func NewReferenceExtractor() *referenceExtractor {
	return &referenceExtractor{}
}

func (e *referenceExtractor) Extract(hyp v3.Hypothesis) []v3.Reference {
	var refs []v3.Reference
	for _, slot := range hyp.Slots() {
		// Only evaluate slots that are structurally intended to carry references (or generic slots that we check).
		// We strictly limit reference evaluation to "target" or "reference" slots to avoid false positives in "content" or "task".
		if slot.Name() == "target" || slot.Name() == "reference" {
			lower := strings.ToLower(slot.Value())
			if isReference(lower) {
				refs = append(refs, v3.NewReference(slot.Value(), ontology.RefPronoun, slot.Name(), slot.GroundingID(), false, slot.Confidence()))
			}
		}
	}
	return refs
}
