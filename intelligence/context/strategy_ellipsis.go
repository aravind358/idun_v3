package context

import (
	"context"

	underv3 "idun/intelligence/understanding/v3"
)

// EllipsisStrategy reconstrucs intent and context for conversational ellipsis (e.g. "what about X", "and Y").
type EllipsisStrategy struct{}

func (s *EllipsisStrategy) Execute(ctx context.Context, orig *underv3.SemanticInterpretation, builder *underv3.Builder, state DialogueStateReader, resolvedEntities map[string]string) (bool, ResolutionStatus) {
	if orig.PrimaryIntent() != "context_ellipsis" {
		return false, StatusContextUnnecessary
	}

	prevBatch := state.GetPreviousBatch()
	if prevBatch == nil || len(prevBatch.Interpretations()) == 0 {
		return true, StatusFailed
	}

	// For U7.5, we inherit from the first interpretation of the previous batch.
	prevInterp := prevBatch.Interpretations()[0]
	if prevInterp.PrimaryIntent() == "" {
		return true, StatusFailed
	}

	// Reconstruct the intent from the previous frame
	builder.PrimaryIntent(prevInterp.PrimaryIntent())

	// Add missing slots from the previous frame that aren't overridden in the current frame
	existingSlots := make(map[string]bool)
	for _, slot := range orig.PrimaryHypothesis().Slots() {
		existingSlots[slot.Name()] = true
	}

	newSlots := orig.PrimaryHypothesis().Slots()
	for _, prevSlot := range prevInterp.PrimaryHypothesis().Slots() {
		if !existingSlots[prevSlot.Name()] {
			// Inherit context slot
			newSlots = append(newSlots, prevSlot)
			resolvedEntities[prevSlot.Name()] = prevSlot.GroundingID()
		}
	}
	builder.Slots(newSlots)
	builder.OpenSlots(nil)

	return true, StatusResolved
}
