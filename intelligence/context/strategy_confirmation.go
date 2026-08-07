package context

import (
	"context"

	underv3 "idun/intelligence/understanding/v3"
)

// ConfirmationStrategy maps confirmation or negation (e.g. yes, no, do it, cancel) to pending active goals.
type ConfirmationStrategy struct{}

func (s *ConfirmationStrategy) Execute(ctx context.Context, orig *underv3.SemanticInterpretation, builder *underv3.Builder, state DialogueStateReader, resolvedEntities map[string]string) (bool, ResolutionStatus) {
	act := orig.CommunicativeAct()
	if act != underv3.ActConfirmation && act != underv3.ActRefusal {
		return false, StatusContextUnnecessary
	}

	activeGoals := state.GetActiveGoals()
	if len(activeGoals) == 0 {
		return true, StatusFailed
	}

	// For confirmation/negation, the context intent targets the most recent active goal
	// We might mutate the intent, e.g., to "confirm_X", or keep it as "confirmation" but add a slot for the Target Goal.
	// For now, we resolve the target to the most recent goal.
	resolvedEntities["target_goal"] = activeGoals[len(activeGoals)-1]

	return true, StatusResolved
}
