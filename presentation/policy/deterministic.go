package policy

import (
	"context"
	"fmt"

	"idun/presentation"
)

// DeterministicRealizationPolicy is the Phase 4/5 implementation of RealizationPolicy.
// It uses a two-level rule table:
//   1. ResponseType rules (precise): map a specific response type to an engine.
//   2. Strategy fallback (coarse): map a RealizationStrategy to an engine when ResponseType is absent.
//
// A missing mapping is treated as a configuration error (not a silent fallback),
// so architectural misconfigurations surface immediately.
//
// Future evolution (Phase 5/6): Replace this implementation with a learned realization selector —
// for example, NeuralRealizationPolicy or AdaptiveRealizationPolicy — using inputs such as:
// response type, semantic content, context, confidence, user preferences, conversation state,
// execution environment, and other optimization criteria.
// The RealizationPolicy interface and the Router remain unchanged when this replacement occurs.
type DeterministicRealizationPolicy struct {
	// responseTypeRules maps ResponseType strings to their designated engine.
	// Takes priority over strategyFallback when ResponseType is set.
	// Example: "calculator" → deterministicEngine, "weather" → generativeEngine.
	responseTypeRules map[string]presentation.RealizationEngine

	// strategyFallback maps RealizationStrategy to an engine.
	// Used when the CapabilityResult does not provide a ResponseType (e.g. Files, System native caps).
	strategyFallback map[presentation.RealizationStrategy]presentation.RealizationEngine
}

// NewDeterministicRealizationPolicy constructs a DeterministicRealizationPolicy.
//
//   responseTypeRules: maps ResponseType strings (e.g. "calculator", "time") → engine.
//   strategyFallback:  maps RealizationStrategy (StrategyDeterministic, StrategyGenerative) → engine.
func NewDeterministicRealizationPolicy(
	responseTypeRules map[string]presentation.RealizationEngine,
	strategyFallback map[presentation.RealizationStrategy]presentation.RealizationEngine,
) *DeterministicRealizationPolicy {
	return &DeterministicRealizationPolicy{
		responseTypeRules: responseTypeRules,
		strategyFallback:  strategyFallback,
	}
}

// Select returns the RealizationEngine appropriate for the given PresentationContext.
//
// Selection priority:
//  1. ResponseType rule table — precise, explicit per-type routing.
//  2. Strategy fallback — coarse-grained, for capabilities that omit ResponseType.
//  3. Error — missing configuration is surfaced explicitly; no silent fallback.
func (p *DeterministicRealizationPolicy) Select(_ context.Context, pctx presentation.PresentationContext) (presentation.RealizationEngine, error) {
	if pctx.ResponseType != "" {
		if engine, ok := p.responseTypeRules[pctx.ResponseType]; ok {
			return engine, nil
		}
	}
	if engine, ok := p.strategyFallback[pctx.Strategy]; ok {
		return engine, nil
	}
	return nil, fmt.Errorf(
		"realization policy: no engine registered for ResponseType=%q Strategy=%v",
		pctx.ResponseType, pctx.Strategy,
	)
}
