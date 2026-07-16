package learning

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"idun/core/memory"
)

func TestDomainLearnersMetadata(t *testing.T) {
	reasoning := NewReasoningLearner()
	if reasoning.LearnerID() != "learner-reasoning-heuristics-v1" || reasoning.LearnerVersion() != "1.0.0" || reasoning.LearnerFingerprint() == "" {
		t.Errorf("unexpected reasoning learner metadata: %s v%s fp=%s", reasoning.LearnerID(), reasoning.LearnerVersion(), reasoning.LearnerFingerprint())
	}
	if len(reasoning.Consumes()) != 2 || reasoning.Produces()[0] != "idun.reasoning.strategy.v1" {
		t.Errorf("unexpected reasoning consumes/produces")
	}

	planning := NewPlanningLearner()
	if planning.LearnerID() != "learner-planning-specialist-v1" || planning.LearnerVersion() != "1.0.0" || planning.LearnerFingerprint() == "" {
		t.Errorf("unexpected planning learner metadata: %s v%s fp=%s", planning.LearnerID(), planning.LearnerVersion(), planning.LearnerFingerprint())
	}
	if len(planning.Consumes()) != 2 || planning.Produces()[0] != "idun.planning.strategy.v1" {
		t.Errorf("unexpected planning consumes/produces")
	}

	decision := NewDecisionLearner()
	if decision.LearnerID() != "learner-decision-weights-v1" || decision.LearnerVersion() != "1.0.0" || decision.LearnerFingerprint() == "" {
		t.Errorf("unexpected decision learner metadata: %s v%s fp=%s", decision.LearnerID(), decision.LearnerVersion(), decision.LearnerFingerprint())
	}
	if len(decision.Consumes()) != 2 || decision.Produces()[0] != "idun.decision.strategy.v1" {
		t.Errorf("unexpected decision consumes/produces")
	}
}

func TestDomainLearnersGenerateSuccess(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	recReasoning := memory.Record{ID: "r1", Type: "idun.reasoning.trace.v1", Payload: []byte(`{"steps":5}`), CreatedAt: now}
	recPlanning := memory.Record{ID: "p1", Type: "idun.planning.trace.v1", Payload: []byte(`{"htn":true}`), CreatedAt: now}
	recDecision := memory.Record{ID: "d1", Type: "idun.decision.trace.v1", Payload: []byte(`{"score":0.9}`), CreatedAt: now}
	recReflection := memory.Record{ID: "ref1", Type: "idun.reflection.report.v1", Payload: []byte(`{"verdict":"PASS"}`), CreatedAt: now}

	summary := &AggregationSummary{
		SummaryID:              "sum-learn-test",
		TimeWindowStart:        now.Add(-2 * time.Hour),
		TimeWindowEnd:          now,
		TotalArtifactsIngested: 4,
		SourceArtifactHash:     "hash-learn-exact-777",
		DomainSchemaIDs:        []string{"idun.reasoning.trace.v1", "idun.planning.trace.v1", "idun.decision.trace.v1"},
		AggregationPolicyID:    "agg-policy-1",
		Records:                []memory.Record{recReasoning, recPlanning, recDecision, recReflection},
	}

	// 1. ReasoningLearner
	reasoning := NewReasoningLearner()
	snapsR, err := reasoning.Generate(ctx, summary)
	if err != nil || len(snapsR) != 1 {
		t.Fatalf("Reasoning Generate failed: err=%v, len=%d", err, len(snapsR))
	}
	snapR := snapsR[0]
	if snapR.SchemaID != "idun.reasoning.strategy.v1" || snapR.SemVer != "1.0.0" || snapR.Lifecycle != LifecycleDraft {
		t.Errorf("unexpected reasoning candidate snapshot fields: %+v", snapR)
	}
	if snapR.Lineage.SourceArtifactHash != "hash-learn-exact-777" {
		t.Errorf("expected exact lineage source hash, got: %s", snapR.Lineage.SourceArtifactHash)
	}
	var payloadMapR map[string]interface{}
	if err := json.Unmarshal(snapR.Payload, &payloadMapR); err != nil {
		t.Errorf("invalid json in reasoning candidate payload: %v", err)
	}

	// 2. PlanningLearner
	planning := NewPlanningLearner()
	snapsP, err := planning.Generate(ctx, summary)
	if err != nil || len(snapsP) != 1 {
		t.Fatalf("Planning Generate failed: err=%v, len=%d", err, len(snapsP))
	}
	snapP := snapsP[0]
	if snapP.SchemaID != "idun.planning.strategy.v1" || snapP.SemVer != "1.0.0" {
		t.Errorf("unexpected planning candidate snapshot fields")
	}

	// 3. DecisionLearner
	decision := NewDecisionLearner()
	snapsD, err := decision.Generate(ctx, summary)
	if err != nil || len(snapsD) != 1 {
		t.Fatalf("Decision Generate failed: err=%v, len=%d", err, len(snapsD))
	}
	snapD := snapsD[0]
	if snapD.SchemaID != "idun.decision.strategy.v1" || snapD.SemVer != "1.0.0" {
		t.Errorf("unexpected decision candidate snapshot fields")
	}
}

func TestDomainLearnersGenerateEmptyOrNil(t *testing.T) {
	ctx := context.Background()
	learners := []Learner{NewReasoningLearner(), NewPlanningLearner(), NewDecisionLearner()}

	for _, l := range learners {
		// Nil check
		if _, err := l.Generate(ctx, nil); err == nil {
			t.Errorf("expected error on nil summary for %s", l.LearnerID())
		}

		// Empty summary check (TotalArtifactsIngested = 0)
		emptySummary := &AggregationSummary{
			SummaryID:              "sum-empty",
			TimeWindowStart:        time.Now().Add(-1 * time.Hour),
			TimeWindowEnd:          time.Now(),
			TotalArtifactsIngested: 0,
			SourceArtifactHash:     "hash-empty",
			DomainSchemaIDs:        []string{"idun.reasoning.trace.v1"},
		}
		res, err := l.Generate(ctx, emptySummary)
		if err != nil {
			t.Errorf("expected nil error on empty summary for %s, got: %v", l.LearnerID(), err)
		}
		if res != nil {
			t.Errorf("expected nil candidates on empty summary for %s, got: %v", l.LearnerID(), res)
		}
	}
}
