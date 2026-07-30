package v3

import (
	"context"
	"idun/core/foundation"
	understanding "idun/intelligence/understanding/v3"
	"testing"
	"time"
)

type mockMemory struct{}

func (m *mockMemory) RetrieveEntity(ctx context.Context, surfaceName string) (string, float64, error) {
	if surfaceName == "Mughal Empire" {
		return "kb-mughal", 0.95, nil
	}
	return "", 0.0, nil
}
func (m *mockMemory) ResolveReference(ctx context.Context, pronoun string) (string, string, float64, error) {
	if pronoun == "it" {
		return "the lights", "device-lights-1", 0.85, nil
	}
	return "", "", 0.0, nil
}
func (m *mockMemory) RetrieveContext(ctx context.Context, intent string, topics []string) ([]ContextEvidence, error) {
	if intent == "turn_on" {
		return []ContextEvidence{NewContextEvidence("semantic", "lights are currently off", 1.0)}, nil
	}
	return nil, nil
}
func (m *mockMemory) EvaluateCondition(ctx context.Context, condition string) (bool, error) {
	if condition == "if it is dark" {
		return true, nil
	}
	return false, nil
}
func (m *mockMemory) EvaluateFact(ctx context.Context, premise string) (bool, error) {
	if premise == "fact_check" {
		return true, nil
	}
	return false, nil
}

func TestOrchestrator_Reason(t *testing.T) {
	artID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	_ = artID
	_ = envID

	hyp := understanding.NewHypothesis("turn_on", 0.9, 0.0, understanding.LayerNeuralClassifier, []understanding.Slot{
		understanding.NewSlot("target", "Mughal Empire", "", 0.9), // Mixing test data to hit all mocks
	})

	interp, _ := understanding.NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Status(understanding.StatusUnambiguous).
		PrimaryIntent("turn_on").
		PrimaryHypothesis(hyp).
		CompoundIntentCount(1).
		CommunicativeAct(understanding.ActStatement).
		IsConditional(true).
		ConditionClause("if it is dark").
		Entities([]understanding.Entity{understanding.NewEntity("Mughal Empire", understanding.EntityLocation, "", "", 0.9)}).
		References([]understanding.Reference{understanding.NewReference("it", understanding.RefPronoun, "", "", false, 0.0)}).
		Build()

	mem := &mockMemory{}
	orc := NewOrchestrator(mem)

	ctxResult, err := orc.Reason(context.Background(), interp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 1. Lineage Verification
	if ctxResult.ParentArtifactID() != foundation.ParentArtifactID(interp.ArtifactID()) {
		t.Errorf("expected ParentArtifactID %s, got %s", interp.ArtifactID(), ctxResult.ParentArtifactID())
	}
	if ctxResult.EnvelopeID() != interp.EnvelopeID() {
		t.Errorf("mismatch envelope id")
	}
	if ctxResult.ArtifactID() == "" {
		t.Errorf("missing artifact id")
	}

	// 2. Intent & Slots
	if ctxResult.ResolvedIntent() != "turn_on" {
		t.Errorf("expected turn_on, got %s", ctxResult.ResolvedIntent())
	}
	if len(ctxResult.EnrichedSlots()) != 1 {
		t.Errorf("expected 1 enriched slot, got %d", len(ctxResult.EnrichedSlots()))
	}
	if ctxResult.EnrichedSlots()[0].EnrichedValue() != "kb-mughal" {
		t.Errorf("expected enriched value kb-mughal, got %s", ctxResult.EnrichedSlots()[0].EnrichedValue())
	}

	// 3. Grounding & References
	if len(ctxResult.GroundedEntities()) != 1 || ctxResult.GroundedEntities()[0].MemoryID() != "kb-mughal" {
		t.Errorf("failed to ground entity")
	}
	if len(ctxResult.ResolvedReferences()) != 1 || ctxResult.ResolvedReferences()[0].MemoryID() != "device-lights-1" {
		t.Errorf("failed to resolve reference")
	}

	// 4. Context & Conditions
	if len(ctxResult.RetrievedContexts()) != 1 {
		t.Errorf("failed to retrieve context")
	}
	if !ctxResult.ConditionEvaluated() || !ctxResult.ConditionMet() {
		t.Errorf("condition failed to evaluate correctly")
	}
	if !ctxResult.TruthEvaluated() || ctxResult.IsFactuallyTrue() {
		// premise is turn_on, so evaluate fact returns false for it
		if ctxResult.IsFactuallyTrue() {
			t.Errorf("truth should be false")
		}
	}
}
