package v3

import (
	"context"
	"idun/core/foundation"
	reasoning "idun/intelligence/reasoning/v3"
	understanding "idun/intelligence/understanding/v3"
	"testing"
	"time"
)

type mockRegistry struct{}

func (m *mockRegistry) Discover(ctx context.Context, goal string) ([]CapabilityDescriptor, error) {
	if goal == "turn_on" {
		return []CapabilityDescriptor{
			NewCapabilityDescriptor("urn:capability:device.turn_on", "Turns on a device", []string{"target"}),
		}, nil
	}
	return nil, nil
}

func TestOrchestrator_Plan(t *testing.T) {
	interpID, _ := foundation.NewUUID()
	reasonCtxID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	// SemanticInterpretation stub
	interp, _ := understanding.NewBuilder().
		ArtifactID(foundation.ArtifactID(interpID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Status(understanding.StatusUnambiguous).
		PrimaryIntent("turn_on").
		PrimaryHypothesis(understanding.NewHypothesis("turn_on", 0.9, 0, understanding.LayerNeuralClassifier, nil)).
		CompoundIntentCount(1).
		CommunicativeAct(understanding.ActStatement).
		Build()

	// ReasoningContext stub
	origSlot := understanding.NewSlot("target", "Mughal Empire", "", 0.9)
	enriched := reasoning.NewEnrichedSlot(origSlot, "kb-mughal")

	reasonCtx, _ := reasoning.NewBuilder().
		ArtifactID(foundation.ArtifactID(reasonCtxID)).
		ParentArtifactID(foundation.ParentArtifactID(interpID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		ResolvedIntent("turn_on").
		EnrichedSlots([]reasoning.EnrichedSlot{enriched}).
		Build()

	registry := &mockRegistry{}
	orc := NewOrchestrator(registry)

	plan, err := orc.Plan(context.Background(), interp, reasonCtx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Lineage
	if plan.ParentArtifactID() != foundation.ParentArtifactID(reasonCtxID) {
		t.Errorf("expected ParentArtifactID %s, got %s", reasonCtxID, plan.ParentArtifactID())
	}
	if plan.EnvelopeID() != foundation.EnvelopeID(envID) {
		t.Errorf("mismatch EnvelopeID")
	}

	// Graph verification
	nodes := plan.Nodes()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Capability() != "urn:capability:device.turn_on" {
		t.Errorf("expected urn:capability:device.turn_on, got %s", nodes[0].Capability())
	}

	if val, ok := nodes[0].BoundParams()["target"]; !ok || val != "kb-mughal" {
		t.Errorf("expected target param bounded to kb-mughal, got %v", val)
	}

	if len(plan.Edges()) != 0 {
		t.Errorf("expected 0 edges, got %d", len(plan.Edges()))
	}
}
