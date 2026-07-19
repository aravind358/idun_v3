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

func TestReasoningCascadeService_HelloTrace(t *testing.T) {
	// Proves Phase 8: Exact "Hello" trace showing ProposedGoal -> Fingerprint -> ResolvedGoal
	mockCal := &mockCalibService{multiplier: 1.0}
	srv := NewService(DefaultConfig(), nil, nil, WithCalibrationService(mockCal))

	frameJSON := []byte(`{
		"FrameVersion": "2.0",
		"EnvelopeID": "env-hello-trace",
		"Status": "UNAMBIGUOUS",
		"PrimaryHypothesis": {
			"Intent": "greet_user",
			"CalibratedConfidence": 0.95
		}
	}`)

	env := communication.Envelope{
		ID:            "env-hello-trace",
		Source:        "perception",
		PayloadRef:    string(frameJSON),
		RawConfidence: 0.95,
	}

	spec := StrategySpec{
		StrategyID: StrategySymbolicFast,
		EnabledStages: []StageIdentifier{
			StageS0ContextAssembly,
			StageS1SymbolicFast,
			StageS4EvidenceFusion,
			StageS6BeamSelection,
			StageS7Calibration,
			StageS10ResultAssembly,
		},
	}

	result, err := srv.ReasonEnvelope(context.Background(), env, spec)
	if err != nil {
		t.Fatalf("ReasonEnvelope failed for Hello trace: %v", err)
	}

	if result.PrimaryHypothesis.Conclusion != `Derived symbolic conclusion for intent "greet_user"` {
		t.Errorf("unexpected primary conclusion: %q", result.PrimaryHypothesis.Conclusion)
	}

	proposed := result.PrimaryHypothesis.ProposedGoal
	if proposed == nil {
		t.Fatalf("expected primary hypothesis to have non-nil ProposedGoal")
	}
	if proposed.Kind != GoalKindCommunicative || proposed.Intent != "greet_user" || proposed.Target != "user" {
		t.Errorf("unexpected ProposedGoal fields: %+v", proposed)
	}

	fp := proposed.Fingerprint()
	if fp == "" {
		t.Fatalf("expected valid non-empty fingerprint for ProposedGoal")
	}

	if result.ResolvedGoal == nil {
		t.Fatalf("expected result.ResolvedGoal to be populated from valid primary ProposedGoal")
	}
	if result.ResolvedGoal.Fingerprint() != fp {
		t.Errorf("expected ResolvedGoal fingerprint %q to match ProposedGoal fingerprint %q", result.ResolvedGoal.Fingerprint(), fp)
	}

	// Verify Planning boundary invariant: ResolvedGoal and PrimaryHypothesis are both preserved
	// and Planning does not consume ResolvedGoal inside Reasoning service
	if result.PrimaryHypothesis.ProposedGoal != proposed {
		t.Errorf("expected PrimaryHypothesis.ProposedGoal pointer invariance")
	}
}

func TestReasoningCascadeService_S8Interaction(t *testing.T) {
	// Proves Phase 11 Scenarios A, B, and C
	specialist := NewBayesianFusionSpecialist()

	goalGreet := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
	}

	// Scenario C: Internal specialist and S8 both propose SAME goal
	hypsScenarioC := []ReasoningHypothesis{
		{
			ID:                  "hyp-internal-greet",
			Type:                HypothesisType("Symbolic"),
			Conclusion:          `Derived symbolic conclusion for intent "greet_user"`,
			ReasoningConfidence: 0.75,
			ProposedGoal:        goalGreet,
			SupportingPremises:  []string{"rule_match=dialogue_intent:greet_user"},
			ContributingStages:  []StageIdentifier{StageS1SymbolicFast},
		},
		{
			ID:                  "hyp-s8-greet",
			Type:                HypothesisType("Deliberative"),
			Conclusion:          `Deliberative LLM conclusion for greet`,
			ReasoningConfidence: 0.82,
			ProposedGoal:        goalGreet.Clone(),
			SupportingPremises:  []string{"llm_reasoning=greeting"},
			ContributingStages:  []StageIdentifier{StageS8DeliberativeLLM},
		},
	}

	fusedC, err := specialist.FuseEvidence(context.Background(), hypsScenarioC)
	if err != nil {
		t.Fatalf("Scenario C fusion failed: %v", err)
	}
	if len(fusedC) != 1 {
		t.Fatalf("expected Scenario C to fuse duplicate S1 and S8 goals into 1 corroborated hypothesis, got %d", len(fusedC))
	}
	if fusedC[0].ReasoningConfidence <= 0.82 {
		t.Errorf("expected corroborated confidence > 0.82, got %f", fusedC[0].ReasoningConfidence)
	}

	// Scenario B: S8 produces no valid ProposedGoal while internal produced low/nil goal
	hypsScenarioB := []ReasoningHypothesis{
		{
			ID:                  "hyp-internal-fallback",
			Type:                HypothesisInference,
			Conclusion:          `Incomplete fallback without goal`,
			ReasoningConfidence: 0.40,
			ProposedGoal:        nil,
		},
		{
			ID:                  "hyp-s8-nogoal",
			Type:                HypothesisType("Deliberative"),
			Conclusion:          `LLM analysis complete but no actionable goal`,
			ReasoningConfidence: 0.60,
			ProposedGoal:        nil,
		},
	}
	fusedB, _ := specialist.FuseEvidence(context.Background(), hypsScenarioB)
	beamSpec := NewBeamSelectionSpecialist()
	primaryB, _, _ := beamSpec.SelectBeam(fusedB, MaxBeamWidth, 0.25)
	if primaryB.ProposedGoal != nil {
		t.Errorf("expected primary to have nil ProposedGoal when neither specialist produced valid goal")
	}
}
