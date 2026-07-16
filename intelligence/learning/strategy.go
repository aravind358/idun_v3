package learning

import (
	"fmt"
	"sync/atomic"
	"time"
)

// DefaultStrategyProvider implements StrategyProvider using lock-free atomic pointers.
// It guarantees zero-lock, passive consumption of immutable LearningStrategySnapshot packages
// during active cognitive episodes.
type DefaultStrategyProvider struct {
	ptr atomic.Pointer[LearningStrategySnapshot]
}

// NewDefaultStrategyProvider initializes a new StrategyProvider with a validated initial snapshot.
func NewDefaultStrategyProvider(initial *LearningStrategySnapshot) (*DefaultStrategyProvider, error) {
	if initial == nil {
		initial = &LearningStrategySnapshot{
			SnapshotID:        "snap-learning-default-v2.0.0",
			SchemaVersion:     SchemaVersion,
			ActiveProfile:     DefaultLearningPolicyProfile(),
			Capabilities:      DefaultLearningCapabilities(),
			AggregationPolicy: DefaultAggregationPolicy(),
			CreatedAt:         time.Now(),
		}
	}
	if err := initial.Validate(); err != nil {
		return nil, fmt.Errorf("initial strategy snapshot validation failed: %w", err)
	}
	p := &DefaultStrategyProvider{}
	p.ptr.Store(initial)
	return p, nil
}

// ActiveSnapshot returns the currently live LearningStrategySnapshot without blocking or locking.
func (p *DefaultStrategyProvider) ActiveSnapshot() *LearningStrategySnapshot {
	return p.ptr.Load()
}

// SwapSnapshot atomically replaces the active strategy snapshot after verifying its structural integrity.
// This is invoked strictly at episode/window boundaries or by Executive governance.
func (p *DefaultStrategyProvider) SwapSnapshot(next *LearningStrategySnapshot) error {
	if next == nil {
		return fmt.Errorf("%w: cannot swap to nil snapshot", ErrValidationFailed)
	}
	if err := next.Validate(); err != nil {
		return fmt.Errorf("next strategy snapshot validation failed: %w", err)
	}
	p.ptr.Store(next)
	return nil
}
