package reasoning

import (
	"context"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/workspace"
)

type mockActionGate struct {
	verdict constitution.Verdict
	reason  string
	sig     string
}

func (m *mockActionGate) EvaluateAction(ctx context.Context, env communication.Envelope) (constitution.EvaluationResult, error) {
	return constitution.EvaluationResult{
		EnvelopeID:  env.ID,
		Verdict:     m.verdict,
		Reason:      m.reason,
		Signature:   m.sig,
		EvaluatedAt: time.Now(),
	}, nil
}

func (m *mockActionGate) InterceptAndPublish(ctx context.Context, env communication.Envelope, ws workspace.Workspace) (constitution.EvaluationResult, error) {
	return m.EvaluateAction(ctx, env)
}

func (m *mockActionGate) RegisterRule(rule constitution.Rule) error { return nil }
func (m *mockActionGate) ListRules() []string                       { return nil }
func (m *mockActionGate) Name() string                              { return "mock.gate" }
func (m *mockActionGate) Start() error                              { return nil }
func (m *mockActionGate) Close() error                              { return nil }

func TestConstitutionSpecialist_Approved(t *testing.T) {
	mock := &mockActionGate{verdict: constitution.VerdictApproved, sig: "SIG-OK"}
	specialist := NewConstitutionSpecialist(mock)

	if specialist.ID() != StageS9Constitution {
		t.Errorf("expected stage ID S9, got %s", specialist.ID())
	}

	res := &ReasoningResult{
		EnvelopeID: "env-1",
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                   "h1",
			CalibratedConfidence: 0.90,
		},
	}

	err := specialist.EvaluateResult(context.Background(), res)
	if err != nil {
		t.Fatalf("expected approved result to succeed, got %v", err)
	}
	if len(res.ConstitutionAnnotations) != 1 {
		t.Fatalf("expected 1 constitution annotation, got %d", len(res.ConstitutionAnnotations))
	}
}

func TestConstitutionSpecialist_Vetoed(t *testing.T) {
	mock := &mockActionGate{verdict: constitution.VerdictVetoed, reason: "violates rule 1"}
	specialist := NewConstitutionSpecialist(mock)

	res := &ReasoningResult{EnvelopeID: "env-veto"}
	err := specialist.EvaluateResult(context.Background(), res)
	if err == nil {
		t.Errorf("expected vetoed result to return error")
	}
}
