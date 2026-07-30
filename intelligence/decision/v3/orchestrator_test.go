package v3

import (
	"context"
	"idun/core/foundation"
	planning "idun/intelligence/planning/v3"
	"testing"
	"time"
)

type mockSafety struct{}

func (m *mockSafety) CheckSafety(ctx context.Context, node planning.PlanNode) (bool, DecisionFinding, error) {
	if node.Capability() == "urn:capability:nuclear.launch" {
		return false, NewDecisionFinding("Safety", node.NodeID(), false, "ERR_NUCLEAR", "Nuclear launch rejected"), nil
	}
	return true, DecisionFinding{}, nil
}

type mockAuth struct{}

func (m *mockAuth) CheckPermissions(ctx context.Context, node planning.PlanNode) (bool, DecisionFinding, error) {
	if node.Capability() == "urn:capability:admin.only" {
		return false, NewDecisionFinding("Auth", node.NodeID(), false, "ERR_NO_AUTH", "Missing admin rights"), nil
	}
	return true, DecisionFinding{}, nil
}

func TestOrchestrator_Decide_Approved(t *testing.T) {
	planID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	n1 := planning.NewPlanNode("n1", "urn:capability:device.turn_on", nil)

	plan, _ := planning.NewBuilder().
		ArtifactID(foundation.ArtifactID(planID)).
		ParentArtifactID(foundation.ParentArtifactID("reason-1")).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Nodes([]planning.PlanNode{n1}).
		Build()

	orc := NewOrchestrator(&mockSafety{}, &mockAuth{}, nil, nil)
	rec, err := orc.Decide(context.Background(), nil, nil, plan)

	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if rec.Resolution() != StatusApproved {
		t.Errorf("expected APPROVED, got %s", rec.Resolution())
	}
	if rec.ParentArtifactID() != foundation.ParentArtifactID(planID) {
		t.Errorf("mismatch parent artifact ID")
	}
	if !rec.SafetyPassed() || !rec.PermissionsPassed() {
		t.Errorf("expected evaluation flags to pass")
	}
}

func TestOrchestrator_Decide_Rejected(t *testing.T) {
	planID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	n1 := planning.NewPlanNode("n1", "urn:capability:nuclear.launch", nil)

	plan, _ := planning.NewBuilder().
		ArtifactID(foundation.ArtifactID(planID)).
		ParentArtifactID(foundation.ParentArtifactID("reason-1")).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Nodes([]planning.PlanNode{n1}).
		Build()

	orc := NewOrchestrator(&mockSafety{}, &mockAuth{}, nil, nil)
	rec, err := orc.Decide(context.Background(), nil, nil, plan)

	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if rec.Resolution() != StatusRejected {
		t.Errorf("expected REJECTED, got %s", rec.Resolution())
	}
	if rec.SafetyPassed() {
		t.Errorf("expected safety to fail")
	}
	if len(rec.Findings()) != 1 {
		t.Fatalf("expected 1 finding")
	}
	if rec.Findings()[0].Code() != "ERR_NUCLEAR" {
		t.Errorf("mismatch finding code")
	}
}
