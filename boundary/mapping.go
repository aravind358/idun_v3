package boundary

import "idun/intelligence/understanding"

// MapSlotFromUnderstanding explicitly converts an internal understanding.Slot
// into a boundary Slot representation without silently duplicating canonical definitions.
func MapSlotFromUnderstanding(s understanding.Slot) Slot {
	return Slot{
		Name:        s.Name,
		Value:       s.Value,
		GroundingID: s.GroundingID,
		Confidence:  s.Confidence,
	}
}

// MapSlotsFromUnderstanding explicitly converts a slice of internal understanding.Slots
// into boundary Slot representations.
func MapSlotsFromUnderstanding(slots []understanding.Slot) []Slot {
	if len(slots) == 0 {
		return nil
	}
	out := make([]Slot, len(slots))
	for i, s := range slots {
		out[i] = MapSlotFromUnderstanding(s)
	}
	return out
}
