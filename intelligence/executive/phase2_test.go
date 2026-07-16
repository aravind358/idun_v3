package executive_test

import (
	"context"
	"testing"
	"time"

	"idun/intelligence/attention"
	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

func TestPhase2CoordinationTerminationSummaryAndVersion(t *testing.T) {
	t.Run("valid termination summary builds and validates inside summary", func(t *testing.T) {
		termSummary := executive.CoordinationTerminationSummary{
			SuccessCount:             10,
			UserCancelledCount:       1,
			InterruptedCount:         2,
			TimeBudgetExceededCount:  1,
			DependencyFailureCount:   0,
			ResourceExhaustedCount:   0,
			ConstitutionBlockedCount: 1,
			ExecutiveAbortCount:      0,
		}
		if err := termSummary.Validate(); err != nil {
			t.Fatalf("expected valid term summary, got %v", err)
		}

		coordSummary := executive.ExecutiveCoordinationSummary{
			EpisodesCoordinated:         15,
			AverageCoordinationDuration: 10 * time.Millisecond,
			TotalCoordinationDuration:   150 * time.Millisecond,
			SuccessfulCoordinations:     10,
			FailedCoordinations:         5,
			TerminationSummary:          termSummary,
		}
		if err := coordSummary.Validate(); err != nil {
			t.Fatalf("expected valid coord summary, got %v", err)
		}
	})

	t.Run("result builder with termination summary and version", func(t *testing.T) {
		termSummary := executive.CoordinationTerminationSummary{
			SuccessCount: 1,
		}
		res, err := executive.NewExecutiveResultBuilder().
			WithEpisodeID("ep-phase2").
			WithWorkflowID("wf-phase2").
			WithStatus(executive.StatusSuccess).
			WithTerminationReason(executive.ReasonSuccess).
			WithTerminationSummary(termSummary).
			WithProvenance("policy-fp-1", "caps-fp-1", "2.0.0-FROZEN").
			WithReplayMetadata(executive.ReplayMetadata{
				PolicyFingerprint:     "policy-fp-1",
				CapabilityFingerprint: "caps-fp-1",
				ExecutiveVersion:      "2.0.0-FROZEN",
			}).
			Build()
		if err != nil {
			t.Fatalf("expected clean build, got %v", err)
		}
		if res.ExecutiveVersion != "2.0.0-FROZEN" {
			t.Errorf("expected version 2.0.0-FROZEN, got %q", res.ExecutiveVersion)
		}
		if res.TerminationSummary.SuccessCount != 1 {
			t.Errorf("expected success count 1, got %d", res.TerminationSummary.SuccessCount)
		}
	})
}

func TestPhase2AttentionDelegation(t *testing.T) {
	attSvc := attention.NewService()
	v1Svc := executive.NewExecutiveService(executive.Config{
		AttentionGate: attSvc,
	})

	goal := executive.ActiveGoalContext{
		ID:             "goal-delegated",
		Summary:        "Test attention delegation",
		PriorityWeight: 80,
	}
	v1Svc.SetActiveGoal(goal)
	if got := attSvc.GetActiveGoal(); got.ID != "goal-delegated" {
		t.Fatalf("expected goal propagated to attention service, got %q", got.ID)
	}

	stim := executive.Stimulus{
		ID:            "stim-1",
		SalienceScore: 90,
	}
	dec, band := v1Svc.Evaluate(stim)
	if dec != executive.SalienceFocusImmediately || band != executive.PriorityBand1RealTime {
		t.Errorf("expected FOCUS_IMMEDIATELY and Band 1, got %s and %d", dec, band)
	}
}

func TestPhase2PendingCandidateStorageInWorkspace(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	cal := calibration.NewService()
	constGate := constitution.NewGate()

	v2Svc, err := executive.NewServiceV2(ws, cal, constGate, 1000)
	if err != nil {
		t.Fatalf("failed to create ServiceV2: %v", err)
	}

	env, _ := communication.NewEnvelopeBuilder().
		WithSource("Test.Source").
		WithTopic(communication.TopicCandidatePlans).
		WithPayloadRef("storage://ref/test").
		WithConfidence(0.95).
		WithUrgency(80).
		WithCostEstimate(50).
		Build()

	// Submit bid via Executive
	if err := v2Svc.SubmitBid(context.Background(), env, executive.HorizonDeliberative); err != nil {
		t.Fatalf("SubmitBid failed: %v", err)
	}

	// Verify pending candidate is actually stored inside Workspace
	pending := ws.GetPendingCandidates(communication.TopicCandidatePlans)
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending candidate in Workspace, got %d", len(pending))
	}
	if pending[0].Envelope.ID != env.ID {
		t.Errorf("expected envelope ID %s, got %s", env.ID, pending[0].Envelope.ID)
	}

	// Arbitrate competition via Executive
	dec, err := v2Svc.ArbitrateCompetition(context.Background(), communication.TopicCandidatePlans, 0.1)
	if err != nil {
		t.Fatalf("ArbitrateCompetition failed: %v", err)
	}
	if !dec.Admitted {
		t.Fatalf("expected bid admitted, got reason: %s", dec.Reason)
	}

	// Verify winning candidate was removed from Workspace pending list after arbitration
	pendingAfter := ws.GetPendingCandidates(communication.TopicCandidatePlans)
	if len(pendingAfter) != 0 {
		t.Errorf("expected 0 pending candidates in Workspace after arbitration, got %d", len(pendingAfter))
	}
}
