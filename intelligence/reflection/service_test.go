package reflection

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

type mockSpecialistEvaluator struct {
	id      string
	ability executive.CognitiveAbility
	fn      func(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error)
}

func (m *mockSpecialistEvaluator) ID() string {
	return m.id
}

func (m *mockSpecialistEvaluator) TargetAbility() executive.CognitiveAbility {
	return m.ability
}

func (m *mockSpecialistEvaluator) EvaluateEpisode(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
	return m.fn(ctx, traces)
}

func TestService_LifecycleAndMetadata(t *testing.T) {
	srv := NewService()
	if srv.Name() != "Reflection" {
		t.Errorf("got Name %s, want Reflection", srv.Name())
	}
	if srv.Ability() != executive.AbilityReflection {
		t.Errorf("got Ability %s, want %s", srv.Ability(), executive.AbilityReflection)
	}

	if err := srv.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := srv.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if err := srv.Start(); err != ErrServiceClosed {
		t.Errorf("expected ErrServiceClosed after Close, got: %v", err)
	}
}

func TestService_SpecialistRegistrationAndIsolation(t *testing.T) {
	srv := NewService()

	s1 := NewSpecialistEvaluator("spec-reasoning", executive.AbilityReasoning,
		NewHeuristicEvaluationStrategy("spec-reasoning", string(executive.AbilityReasoning)))
	if err := srv.RegisterSpecialist(s1); err != nil {
		t.Fatalf("RegisterSpecialist failed: %v", err)
	}

	// Duplicate registration error
	if err := srv.RegisterSpecialist(s1); !errors.Is(err, ErrSpecialistAlreadyRegistered) {
		t.Errorf("expected ErrSpecialistAlreadyRegistered, got: %v", err)
	}

	specs := srv.GetSpecialists()
	if len(specs) != 1 || specs[0].ID() != "spec-reasoning" {
		t.Errorf("unexpected specialists list: %v", specs)
	}
}

func TestService_ReflectEpisodeAndWorkspacePublication(t *testing.T) {
	ws := workspace.NewEngine()
	srv := NewService(WithWorkspace(ws))
	if err := srv.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv.Close()

	// Register 3 specialists:
	// 1. Evaluated successfully
	sEval := &mockSpecialistEvaluator{
		id:      "spec-eval",
		ability: executive.AbilityReasoning,
		fn: func(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
			refs := make([]TraceReference, len(traces))
			for i, tr := range traces {
				ts := tr.CreatedAt.Unix()
				if ts <= 0 {
					ts = time.Now().Unix()
				}
				refs[i] = TraceReference{
					EnvelopeID:     tr.ID,
					SourceAbility:  tr.Source,
					TraceTimestamp: ts,
					PayloadHashRef: "sha256-hash",
				}
			}
			return SpecialistReport{
				SpecialistID:         "spec-eval",
				TargetAbility:        string(executive.AbilityReasoning),
				Verdict:              VerdictEvaluated,
				WentWell:             []string{"Good reasoning step"},
				ReflectionConfidence: 0.92,
				SourceTraceRefs:      refs,
			}, nil
		},
	}

	// 2. Insufficient data
	sInsuf := &mockSpecialistEvaluator{
		id:      "spec-insuf",
		ability: executive.AbilityPlanning,
		fn: func(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
			return SpecialistReport{
				SpecialistID:         "spec-insuf",
				TargetAbility:        string(executive.AbilityPlanning),
				Verdict:              VerdictInsufficientData,
				ReflectionConfidence: 1.0,
				SourceTraceRefs:      []TraceReference{},
			}, nil
		},
	}

	// 3. Panicking specialist (should be isolated as ABSTAIN without breaking the others!)
	sPanic := &mockSpecialistEvaluator{
		id:      "spec-panic",
		ability: executive.AbilityDecision,
		fn: func(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
			panic("intentional panic for isolation test")
		},
	}

	if err := srv.RegisterSpecialist(sEval); err != nil {
		t.Fatalf("register sEval failed: %v", err)
	}
	if err := srv.RegisterSpecialist(sInsuf); err != nil {
		t.Fatalf("register sInsuf failed: %v", err)
	}
	if err := srv.RegisterSpecialist(sPanic); err != nil {
		t.Fatalf("register sPanic failed: %v", err)
	}

	traces := []communication.Envelope{
		{
			ID:        "env-trace-1",
			Source:    string(executive.AbilityReasoning),
			CreatedAt: time.Now(),
		},
	}

	report, err := srv.ReflectEpisode(context.Background(), "ep-200", traces)
	if err != nil {
		t.Fatalf("ReflectEpisode failed: %v", err)
	}

	if err := report.Validate(); err != nil {
		t.Fatalf("generated ReflectionReport failed validation: %v", err)
	}
	if report.EpisodeID != "ep-200" {
		t.Errorf("got EpisodeID %s, want ep-200", report.EpisodeID)
	}
	if len(report.SpecialistReports) != 3 {
		t.Fatalf("expected 3 specialist reports, got %d", len(report.SpecialistReports))
	}

	// Verify trace lineage on evaluated report
	var foundEval bool
	for _, sr := range report.SpecialistReports {
		if sr.SpecialistID == "spec-eval" {
			foundEval = true
			if len(sr.SourceTraceRefs) != 1 || sr.SourceTraceRefs[0].EnvelopeID != "env-trace-1" {
				t.Errorf("unexpected trace lineage: %v", sr.SourceTraceRefs)
			}
		} else if sr.SpecialistID == "spec-panic" {
			if sr.Verdict != VerdictAbstain {
				t.Errorf("panicking specialist should have verdict ABSTAIN, got %s", sr.Verdict)
			}
		}
	}
	if !foundEval {
		t.Error("expected to find spec-eval report")
	}
}

func TestService_ValidationEnforcementPreventsPublication(t *testing.T) {
	ws := workspace.NewEngine()
	srv := NewService(WithWorkspace(ws))
	_ = srv.Start()
	defer srv.Close()

	// Specialist returning invalid confidence (e.g. 5.0)
	badSpec := &mockSpecialistEvaluator{
		id:      "spec-bad",
		ability: executive.AbilityReasoning,
		fn: func(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
			return SpecialistReport{
				SpecialistID:         "spec-bad",
				TargetAbility:        string(executive.AbilityReasoning),
				Verdict:              VerdictEvaluated,
				ReflectionConfidence: 5.0, // out of bounds
			}, nil
		},
	}

	_ = srv.RegisterSpecialist(badSpec)

	_, err := srv.ReflectEpisode(context.Background(), "ep-invalid", []communication.Envelope{})
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed on invalid report, got: %v", err)
	}
}

func TestService_PeriodicReflectionPhase3(t *testing.T) {
	ws := workspace.NewEngine()
	srv := NewService(WithWorkspace(ws))
	_ = srv.Start()
	defer srv.Close()

	summary := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-phase3",
		GeneratedTimestamp: time.Now().UnixNano(),
		TimeWindow:         TimeWindowSpec{StartTime: 1000, EndTime: 2000},
		EpisodeCount:       50,
		AverageScores: map[string]float64{
			"Reasoning":     0.92,
			"Communication": 0.71,
		},
		TrendMetrics: map[string]float64{
			"AttentionDrift": 0.08,
		},
		FailureRates: map[string]float64{
			"Planning": 0.12,
		},
		ImprovementRates: map[string]float64{
			"Reasoning": 0.06,
		},
		SummaryConfidence: 0.94,
	}

	report, err := srv.ReflectPeriodic(context.Background(), summary)
	if err != nil {
		t.Fatalf("ReflectPeriodic failed: %v", err)
	}

	if err := report.Validate(); err != nil {
		t.Fatalf("generated periodic report failed validation: %v", err)
	}
	if report.Mode != ModePeriodic {
		t.Errorf("got Mode %s, want %s", report.Mode, ModePeriodic)
	}
	if report.SummaryID != "sum-phase3" {
		t.Errorf("got SummaryID %s, want sum-phase3", report.SummaryID)
	}
	if len(report.TrendFindings) == 0 {
		t.Error("expected periodic reflection to populate trend findings")
	}
	if len(report.CrossCognitiveFindings) == 0 {
		t.Error("expected periodic reflection to populate cross cognitive findings")
	}
	if len(report.GrowthPotentialEstimates) == 0 {
		t.Error("expected periodic reflection to populate growth potential estimates")
	}
}

func TestService_ConcurrentRaceSafety(t *testing.T) {
	srv := NewService()
	_ = srv.Start()

	spec := NewSpecialistEvaluator("spec-race", executive.AbilityReasoning,
		NewHeuristicEvaluationStrategy("spec-race", string(executive.AbilityReasoning)))
	_ = srv.RegisterSpecialist(spec)

	var wg sync.WaitGroup
	traces := []communication.Envelope{
		{
			ID:        "env-trace-race",
			Source:    string(executive.AbilityReasoning),
			CreatedAt: time.Now(),
		},
	}

	summary := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-race",
		GeneratedTimestamp: time.Now().UnixNano(),
		AverageScores: map[string]float64{
			"Reasoning": 0.90,
		},
		SummaryConfidence: 0.90,
	}

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = srv.ReflectEpisode(context.Background(), "ep-race", traces)
			_, _ = srv.ReflectPeriodic(context.Background(), summary)
			_ = srv.GetSpecialists()
		}(i)
	}

	wg.Wait()
	_ = srv.Close()
}
