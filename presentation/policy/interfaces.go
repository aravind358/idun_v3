package policy

import (
	"context"

	"idun/presentation"
)

// RealizationPolicy is the single component responsible for realization-engine selection.
// The Router delegates engine selection exclusively to this interface.
// The Router must remain unchanged when this implementation is replaced.
//
// Future evolution (Phase 5/6): Replace DeterministicRealizationPolicy with a learned
// realization selector. Possible future inputs include: response type, semantic content,
// context, confidence, user preferences, conversation state, execution environment,
// and other optimization criteria. The Router must remain unchanged when this occurs.
type RealizationPolicy interface {
	Select(ctx context.Context, pctx presentation.PresentationContext) (presentation.RealizationEngine, error)
}
