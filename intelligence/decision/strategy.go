package decision

import (
	"sync/atomic"
)

// DefaultStrategyProvider implements StrategyProvider with atomic pointer updates,
// ensuring thread-safe lock-free reads of immutable DecisionStrategySnapshots.
type DefaultStrategyProvider struct {
	snapshot atomic.Pointer[DecisionStrategySnapshot]
}

// NewDefaultStrategyProvider initializes a StrategyProvider with a sensible default snapshot.
func NewDefaultStrategyProvider(initial *DecisionStrategySnapshot) *DefaultStrategyProvider {
	if initial == nil {
		initial = NewDefaultStrategySnapshot("v2.0.0-default")
	}
	p := &DefaultStrategyProvider{}
	p.snapshot.Store(initial)
	return p
}

// NewDefaultStrategySnapshot creates a canonical default strategy snapshot with a BALANCED policy profile.
func NewDefaultStrategySnapshot(version string) *DecisionStrategySnapshot {
	weights := map[string]float64{
		"utility":     1.0,
		"safety":      1.5,
		"reliability": 1.2,
		"efficiency":  0.8,
	}
	profile := DecisionPolicyProfile{
		ProfileID:                 "BALANCED",
		PolicyVersion:             "1.0.0",
		PolicySource:              "HANDCRAFTED",
		PolicyFingerprint:         "7C8F29E4BALANCED1",
		Description:               "Default balanced multi-objective utility and constitutional safety profile",
		FeatureWeights:            weights,
		RiskTolerance:             0.25,
		EscalationConfidenceFloor: 0.60,
		EscalationAmbiguityMargin: 0.05,
		ObjectivePriorities:       []string{"safety", "reliability", "utility", "efficiency"},
		MaxReflexiveLatencyUs:     2000,
	}

	return &DecisionStrategySnapshot{
		StrategyVersion:           version,
		ActiveProfileID:           profile.ProfileID,
		ActiveProfile:             profile,
		FeatureWeights:            profile.FeatureWeights,
		EscalationConfidenceFloor: profile.EscalationConfidenceFloor,
		EscalationAmbiguityMargin: profile.EscalationAmbiguityMargin,
		MaxReflexiveLatencyUs:     profile.MaxReflexiveLatencyUs,
	}
}

// ActiveSnapshot returns the active immutable read-only strategy snapshot.
func (p *DefaultStrategyProvider) ActiveSnapshot() (*DecisionStrategySnapshot, error) {
	snap := p.snapshot.Load()
	if snap == nil {
		return nil, ErrInvalidStrategySnapshot
	}
	return snap, nil
}

// UpdateSnapshot atomically swaps the active strategy snapshot.
// This method is called exclusively by Learning during inter-episode adaptation.
func (p *DefaultStrategyProvider) UpdateSnapshot(snap *DecisionStrategySnapshot) error {
	if err := snap.Validate(); err != nil {
		return err
	}
	p.snapshot.Store(snap)
	return nil
}
