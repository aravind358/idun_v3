package strategy

import (
	"context"
	"fmt"

	"idun/world/output"
)

// RealizationStrategy defines how the output should be realized.
type RealizationStrategy int

const (
	StrategyDeterministic RealizationStrategy = iota
	StrategyGenerative
	StrategyHybrid
)

// DeterministicRealizationPolicy selects the appropriate OutputEngine.
type DeterministicRealizationPolicy struct {
	responseTypeRules map[string]output.OutputEngine
	strategyFallback  map[RealizationStrategy]output.OutputEngine
}

// NewDeterministicRealizationPolicy constructs a DeterministicRealizationPolicy.
func NewDeterministicRealizationPolicy(
	responseTypeRules map[string]output.OutputEngine,
	strategyFallback map[RealizationStrategy]output.OutputEngine,
) *DeterministicRealizationPolicy {
	return &DeterministicRealizationPolicy{
		responseTypeRules: responseTypeRules,
		strategyFallback:  strategyFallback,
	}
}

// Select returns the OutputEngine appropriate for the given Context.
func (p *DeterministicRealizationPolicy) Select(_ context.Context, responseType string, strategy RealizationStrategy) (output.OutputEngine, error) {
	if responseType != "" {
		if engine, ok := p.responseTypeRules[responseType]; ok {
			return engine, nil
		}
	}
	if engine, ok := p.strategyFallback[strategy]; ok {
		return engine, nil
	}
	return nil, fmt.Errorf(
		"realization policy: no engine registered for ResponseType=%q Strategy=%v",
		responseType, strategy,
	)
}
