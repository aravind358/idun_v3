package reasoning

import (
	"context"
	"sync"
	"testing"
	"time"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

func TestService_LifecycleAndExecutiveAbilities(t *testing.T) {
	srv := NewService(DefaultConfig(), nil, nil)

	if srv.Name() != "intelligence.reasoning" {
		t.Errorf("expected canonical name intelligence.reasoning, got %s", srv.Name())
	}
	if srv.Ability() != executive.AbilityReasoning {
		t.Errorf("expected AbilityReasoning, got %s", srv.Ability())
	}

	if err := srv.Start(); err != nil {
		t.Fatalf("expected Start() to succeed, got %v", err)
	}
	// Start idempotency
	if err := srv.Start(); err != nil {
		t.Fatalf("expected idempotent Start() to succeed, got %v", err)
	}

	status, envelopeID, err := srv.ExecuteTask(context.Background(), "storage://payload/ref")
	if err != nil {
		t.Fatalf("expected ExecuteTask to succeed, got %v", err)
	}
	if status != executive.StatusConfident || envelopeID == "" {
		t.Errorf("expected confident status and non-empty envelope ID, got %d / %q", status, envelopeID)
	}

	conclusion, err := srv.SynthesizeInference(context.Background(), "storage://premises/ref")
	if err != nil {
		t.Fatalf("expected SynthesizeInference to succeed, got %v", err)
	}
	if conclusion == "" {
		t.Errorf("expected non-empty conclusion")
	}

	if err := srv.Close(); err != nil {
		t.Fatalf("expected Close() to succeed, got %v", err)
	}

	// Verify operations after Close return ErrServiceClosed
	if _, _, err := srv.ExecuteTask(context.Background(), "ref"); err != ErrServiceClosed {
		t.Errorf("expected ErrServiceClosed after Close(), got %v", err)
	}
	if _, err := srv.SynthesizeInference(context.Background(), "ref"); err != ErrServiceClosed {
		t.Errorf("expected ErrServiceClosed after Close(), got %v", err)
	}
}

func TestService_ReasonEnvelopeWithWorkspaceAndMemory(t *testing.T) {
	ws := workspace.NewEngine()
	if err := ws.Start(); err != nil {
		t.Fatalf("failed to start workspace: %v", err)
	}
	defer ws.Close()

	mem := &mockMemoryProvider{
		records: map[string][]memory.Record{
			"belief": {
				{ID: "bel-user", Type: "belief"},
			},
		},
	}

	srv := NewService(
		DefaultConfig(),
		ws,
		mem,
		WithTopics(communication.TopicUserIntent, communication.TopicActiveGoals),
	)
	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}
	defer srv.Close()

	env := communication.Envelope{
		ID:            "env-test-reason",
		Source:        "understanding",
		Topic:         communication.TopicUserIntent,
		PayloadRef:    "storage://frame/123",
		RawConfidence: 0.92,
	}

	result, err := srv.ReasonEnvelope(context.Background(), env, StrategySpec{})
	if err != nil {
		t.Fatalf("ReasonEnvelope failed: %v", err)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("generated ReasoningResult failed validation: %v", err)
	}
	if result.SourceFrameID != "env-test-reason" {
		t.Errorf("expected source frame ID env-test-reason, got %s", result.SourceFrameID)
	}
	if len(result.ReasoningTrace) < 2 {
		t.Errorf("expected trace logs for S0, S1, S10, got %d logs", len(result.ReasoningTrace))
	}
}

func TestService_ConcurrentRaceSafety(t *testing.T) {
	srv := NewService(DefaultConfig(), nil, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}
	defer srv.Close()

	var wg sync.WaitGroup
	const numGoroutines = 40

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			env := communication.Envelope{
				ID:            "env-race",
				Source:        "exec",
				RawConfidence: 0.9,
			}
			_, _ = srv.ReasonEnvelope(ctx, env, StrategySpec{})
			_, _, _ = srv.ExecuteTask(ctx, "ref")
			_, _ = srv.SynthesizeInference(ctx, "ref")
		}(i)
	}

	wg.Wait()
}

func TestService_FullCascadeWithBeamCalibrationConstitutionAndDeliberative(t *testing.T) {
	mockCal := &mockCalibService{multiplier: 0.95}
	mockInf := &mockInferenceService{}
	mockGate := &mockActionGate{verdict: constitution.VerdictApproved, sig: "SIG-VALID"}

	srv := NewService(DefaultConfig(), nil, nil,
		WithCalibrationService(mockCal),
		WithInferenceService(mockInf),
		WithConstitutionGate(mockGate),
	)

	env := communication.Envelope{
		ID:            "env-full-cascade",
		Source:        "perception",
		RawConfidence: 0.90,
	}

	result, err := srv.ReasonEnvelope(context.Background(), env, StrategySpec{})
	if err != nil {
		t.Fatalf("expected ReasonEnvelope cascade to succeed, got %v", err)
	}

	if result.PrimaryHypothesis.CalibratedConfidence == 0.0 {
		t.Errorf("expected Stage S7 Calibration to populate CalibratedConfidence > 0")
	}

	if len(result.ConstitutionAnnotations) == 0 {
		t.Errorf("expected Stage S9 Constitution to annotate approval")
	}
}
