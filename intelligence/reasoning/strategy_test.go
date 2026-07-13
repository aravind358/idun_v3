package reasoning

import (
	"context"
	"testing"

	"idun/intelligence/communication"
)

func TestDefaultStrategySelector_SelectStrategy(t *testing.T) {
	selector := NewDefaultStrategySelector()
	env := communication.Envelope{ID: "env-strat"}

	spec, err := selector.SelectStrategy(context.Background(), env)
	if err != nil {
		t.Fatalf("expected strategy selection to succeed, got %v", err)
	}

	if spec.StrategyID != StrategySymbolicFast {
		t.Errorf("expected default strategy %s, got %s", StrategySymbolicFast, spec.StrategyID)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("selected spec failed validation: %v", err)
	}
}

func TestSelectStrategyForID(t *testing.T) {
	ids := []StrategyIdentifier{
		StrategySymbolicFast,
		StrategyAnalogicalBayes,
		StrategyGraphDeliberative,
		StrategyDeliberativeEscalate,
		"UNKNOWN_STRATEGY",
	}

	for _, id := range ids {
		spec := SelectStrategyForID(id)
		if err := spec.Validate(); err != nil {
			t.Errorf("SelectStrategyForID(%q) returned invalid spec: %v", id, err)
		}
	}
}
