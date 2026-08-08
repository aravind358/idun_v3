package output

import (
	"context"
	"fmt"
)

// DefaultOutputEngine transforms a CompositeResponse into an OutputDocument.
// It is strictly a transformation layer — it does not access CAS, perform inference,
// OrchestratingEngine implements OutputEngine by strictly delegating
// realization to a configured Strategy, remaining completely decoupled
// from capability logic and formatting.
type OrchestratingEngine struct {
	strategy Strategy
}

// NewOrchestratingEngine creates a new OrchestratingEngine.
func NewOrchestratingEngine(strategy Strategy) *OrchestratingEngine {
	return &OrchestratingEngine{
		strategy: strategy,
	}
}

// Realize routes the CompositeResponse through the configured Strategy.
func (e *OrchestratingEngine) Realize(ctx context.Context, response CompositeResponse) (OutputDocument, error) {
	rt := response.PrimaryResponseType()
	
	// Delegate descriptor lookup to strategy
	desc := e.strategy.Select(rt)

	if desc.Realizer == nil {
		return OutputDocument{}, fmt.Errorf("no realizer configured in descriptor for response type: %q", rt)
	}

	// Delegate realization to the selected realizer
	return desc.Realizer.Realize(ctx, response, desc)
}
