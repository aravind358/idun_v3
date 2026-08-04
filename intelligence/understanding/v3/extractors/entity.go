package extractors

import (
	"strings"

	"idun/intelligence/understanding/v3"
	"idun/intelligence/understanding/v3/ontology"
)

type slotBasedEntityExtractor struct{}

func NewSlotBasedEntityExtractor() *slotBasedEntityExtractor {
	return &slotBasedEntityExtractor{}
}

func (e *slotBasedEntityExtractor) Extract(hyp v3.Hypothesis) []v3.Entity {
	var entities []v3.Entity
	for _, slot := range hyp.Slots() {
		eType := ontology.EntityUnknown

		switch slot.Name() {
		case "person", "name":
			eType = ontology.EntityPerson
		case "location", "city":
			eType = ontology.EntityLocation
		case "filename", "extension", "path", "source", "destination":
			eType = ontology.EntityFile
		case "directory":
			eType = ontology.EntityDirectory
		case "title", "content", "task":
			eType = ontology.EntityDocument
		case "operand1", "operand2", "expression":
			eType = ontology.EntityNumber
		case "operator", "operation":
			eType = ontology.EntityUnknown // Command/Verb represented generically for now
		case "target":
			lower := strings.ToLower(slot.Value())
			if isSystemResource(lower) {
				eType = ontology.EntitySystemResource
			} else if isReference(lower) {
				// Handled strictly by ReferenceExtractor
				continue
			} else {
				eType = ontology.EntityUnknown
			}
		case "date", "time", "duration", "daypart", "day", "datetime":
			// Handled strictly by TemporalExtractor
			continue
		default:
			eType = ontology.EntityUnknown
		}

		entities = append(entities, v3.NewEntity(slot.Value(), eType, slot.Value(), slot.GroundingID(), slot.Confidence()))
	}
	return entities
}

func isSystemResource(val string) bool {
	switch val {
	case "battery", "cpu", "memory", "ram", "disk", "storage", "system", "computer", "screen":
		return true
	}
	return false
}

func isReference(val string) bool {
	switch val {
	case "me", "us", "it", "this", "that", "them", "him", "her":
		return true
	}
	return false
}
