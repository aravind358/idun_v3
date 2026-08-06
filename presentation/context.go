package presentation

import "idun/capabilities"

// PresentationContext carries the minimal information required by the Presentation layer
// to select a realization strategy. It is deliberately decoupled from CapabilityResult
// so the Presentation layer does not depend on capability internals.
//
// PresentationContext is designed for forward extension: future phases can add fields
// (e.g. UserPreferences, ConversationState, ConfidenceScore, ExecutionEnvironment)
// without changing the RealizationPolicy interface.
type PresentationContext struct {
	// ResponseType is the semantic category of the capability output (e.g. "calculator", "time").
	ResponseType string

	// Strategy is the coarse-grained realization hint produced by the capability.
	Strategy RealizationStrategy

	// Operation is the semantic operation executed by the capability.
	Operation string

	// ParentRef is the workspace envelope correlation ID.
	ParentRef string

	// PresentationHints supports future stylistic or multimodal metadata.
	PresentationHints map[string]interface{}
}

// RealizationStrategy mirrors the capability-layer strategy enum at the presentation boundary.
type RealizationStrategy int

const (
	StrategyDeterministic RealizationStrategy = iota
	StrategyGenerative
	StrategyHybrid
)

// PresentationContextBuilder constructs a PresentationContext from capability-layer data.
// This builder is the only place in the Presentation layer that imports capabilities.
type PresentationContextBuilder struct {
	pctx PresentationContext
}

// NewPresentationContextBuilder returns a fresh builder.
func NewPresentationContextBuilder() *PresentationContextBuilder {
	return &PresentationContextBuilder{
		pctx: PresentationContext{
			PresentationHints: make(map[string]interface{}),
		},
	}
}

// FromCapabilityResult populates the builder from a CapabilityResult.
func (b *PresentationContextBuilder) FromCapabilityResult(res capabilities.CapabilityResult) *PresentationContextBuilder {
	b.pctx.ResponseType = res.ResponseType
	b.pctx.Strategy = mapStrategy(res.Realization)
	b.pctx.Operation = res.Operation
	return b
}

// WithParentRef sets the workspace envelope correlation ID.
func (b *PresentationContextBuilder) WithParentRef(ref string) *PresentationContextBuilder {
	b.pctx.ParentRef = ref
	return b
}

// Build finalizes the PresentationContext.
func (b *PresentationContextBuilder) Build() PresentationContext {
	return b.pctx
}

func mapStrategy(s capabilities.RealizationStrategy) RealizationStrategy {
	switch s {
	case capabilities.Deterministic:
		return StrategyDeterministic
	case capabilities.Generative:
		return StrategyGenerative
	default:
		return StrategyGenerative
	}
}
