package context

import (
	"context"
	"strings"

	underv3 "idun/intelligence/understanding/v3"
)

// DefaultContextResolver provides the canonical implementation of the ContextResolver interface.
type DefaultContextResolver struct {
	strategies []ResolutionStrategy
}

// NewDefaultContextResolver constructs a new DefaultContextResolver.
func NewDefaultContextResolver(strategies ...ResolutionStrategy) *DefaultContextResolver {
	if len(strategies) == 0 {
		strategies = []ResolutionStrategy{
			&PronounStrategy{},
			&EllipsisStrategy{},
			&ConfirmationStrategy{},
			&TemporalStrategy{},
		}
	}
	return &DefaultContextResolver{
		strategies: strategies,
	}
}

// Resolve implements ContextResolver.Resolve.
func (r *DefaultContextResolver) Resolve(ctx context.Context, batch *underv3.UnderstandingBatch, state DialogueStateReader) (*underv3.UnderstandingBatch, error) {
	if batch == nil {
		return nil, nil
	}

	interps := batch.Interpretations()
	var resolvedInterps []*underv3.SemanticInterpretation
	
	// rolling entities persist across interpretations within the SAME batch
	// This enables sequential resolution (Intent 1 entity grounds Intent 2 pronoun)
	rollingEntities := make(map[string]string)

	for _, interp := range interps {
		if interp.PrimaryIntent() == "" || interp.PrimaryIntent() == "unresolved_intent" {
			resolvedInterps = append(resolvedInterps, interp)
			continue
		}

		builder := underv3.CloneBuilder(interp)
		overallStatus := StatusContextUnnecessary
		anyHandled := false

		// Pipeline iteration
		for _, strategy := range r.strategies {
			handled, status := strategy.Execute(ctx, interp, builder, state, rollingEntities)
			if handled {
				anyHandled = true
				if status == StatusFailed {
					overallStatus = StatusFailed
					break // Stop on failure
				}
				if status == StatusAmbiguous {
					overallStatus = StatusAmbiguous
				} else if status == StatusResolved && overallStatus != StatusAmbiguous {
					overallStatus = StatusResolved
				}
			}
		}

		if !anyHandled {
			intent := interp.PrimaryIntent()
			requiresContext := strings.HasPrefix(intent, "context_")
			if requiresContext {
				overallStatus = StatusFailed
			}
		}

		// Update the interpretation status based on resolution
		if overallStatus == StatusFailed {
			builder.Status(underv3.StatusFailed)
		} else if overallStatus == StatusAmbiguous {
			builder.Status(underv3.StatusAmbiguous)
		}
		
		builtInterp, err := builder.Build()
		if err != nil {
			// If it fails to build, we fallback to the original to prevent panic
			builtInterp = interp
		}
		resolvedInterps = append(resolvedInterps, builtInterp)
	}

	// Construct a new batch with the resolved interpretations
	resolvedBatch := underv3.NewUnderstandingBatch(
		batch.EnvelopeID(),
		batch.ParentArtifactID(),
		batch.OriginalUtterance(),
		resolvedInterps,
	)

	return resolvedBatch, nil
}
