package executive_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

type mockReflectionCoordinator struct {
	wakes int32
}

func (m *mockReflectionCoordinator) WakeReflection(ctx context.Context, epID executive.EpisodeID, reason string) error {
	atomic.AddInt32(&m.wakes, 1)
	return nil
}

type mockLearningCoordinator struct {
	wakes int32
}

func (m *mockLearningCoordinator) WakeLearning(ctx context.Context, epID executive.EpisodeID, reason string) error {
	atomic.AddInt32(&m.wakes, 1)
	return nil
}

type mockStrategyCoordinator struct {
	activations int32
}

func (m *mockStrategyCoordinator) ActivateStrategySnapshot(ctx context.Context, snapshotRef string) error {
	atomic.AddInt32(&m.activations, 1)
	return nil
}

func buildTestEpisode(t *testing.T, id string) *executive.ExecutiveEpisode {
	ec := executive.EpisodeContext{
		WorkspaceReference: "workspace://episodes/test-1",
		AttentionReference: "attention://goals/salience-1",
		GoalReference:      "goal://root/goal-1",
		ModuleReferences: map[string]string{
			"planning": "planning://graphs/plan-1",
			"vision":   "vision://buffers/v-1",
		},
	}
	ep, err := executive.NewExecutiveEpisodeBuilder(
		executive.EpisodeID(id),
		executive.EpisodeTypeCognitiveTurn,
		executive.EpisodeIntentConversation,
		executive.EpisodeOriginUser,
	).WithContextReference(ec).
		WithPriorityAndBudget(executive.PriorityBand2Interactive, executive.BudgetStandard, 100).
		WithExecutorID("node-worker-01").
		Build()
	if err != nil {
		t.Fatalf("buildTestEpisode failed: %v", err)
	}
	return ep
}

func TestEpisodeLifecycleTransitions(t *testing.T) {
	manager := executive.NewEpisodeManager()
	ep := buildTestEpisode(t, "ep-lifecycle-01")

	if err := manager.CreateEpisode(ep); err != nil {
		t.Fatalf("CreateEpisode failed: %v", err)
	}

	// Test valid transitions: CREATED -> WAITING -> RUNNING -> PAUSED -> RUNNING -> COMPLETED
	if err := manager.TransitionStatus("ep-lifecycle-01", executive.EpisodeStatusWaiting, executive.EpisodeOutcomePending, "", ""); err != nil {
		t.Fatalf("transition to waiting failed: %v", err)
	}
	if err := manager.TransitionStatus("ep-lifecycle-01", executive.EpisodeStatusRunning, executive.EpisodeOutcomePending, "", executive.ResumeReasonInputReceived); err != nil {
		t.Fatalf("transition to running failed: %v", err)
	}
	if err := manager.TransitionStatus("ep-lifecycle-01", executive.EpisodeStatusPaused, executive.EpisodeOutcomePending, executive.PauseReasonAwaitingDependency, ""); err != nil {
		t.Fatalf("transition to paused failed: %v", err)
	}
	if err := manager.TransitionStatus("ep-lifecycle-01", executive.EpisodeStatusRunning, executive.EpisodeOutcomePending, "", executive.ResumeReasonDependencyResolved); err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	if err := manager.TransitionStatus("ep-lifecycle-01", executive.EpisodeStatusCompleted, executive.EpisodeOutcomeSuccess, "", ""); err != nil {
		t.Fatalf("transition to completed failed: %v", err)
	}

	// Verify terminal state rejection
	err := manager.TransitionStatus("ep-lifecycle-01", executive.EpisodeStatusRunning, executive.EpisodeOutcomePending, "", "")
	if err != executive.ErrTerminalEpisode {
		t.Fatalf("expected ErrTerminalEpisode when transitioning out of completed, got %v", err)
	}
}

func TestEpisodeRollingHistories(t *testing.T) {
	manager := executive.NewEpisodeManager()
	ep := buildTestEpisode(t, "ep-history-01")
	_ = manager.CreateEpisode(ep)

	// Perform 25 priority and budget updates to test rolling ring buffer bounding (max 16)
	for i := 0; i < 25; i++ {
		p := executive.PriorityBand1RealTime
		if i%2 == 0 {
			p = executive.PriorityBand2Interactive
		}
		_ = manager.UpdatePriority("ep-history-01", p, executive.PriorityReasonSalienceOverride)
		_ = manager.UpdateBudget("ep-history-01", executive.BudgetDeliberative, executive.BudgetReasonEscalationGranted)
	}

	updated, exists := manager.GetEpisode("ep-history-01")
	if !exists {
		t.Fatal("episode not found")
	}
	if len(updated.Runtime.PriorityHistory) > 16 {
		t.Fatalf("expected bounded PriorityHistory <= 16, got %d", len(updated.Runtime.PriorityHistory))
	}
	if len(updated.Runtime.BudgetHistory) > 16 {
		t.Fatalf("expected bounded BudgetHistory <= 16, got %d", len(updated.Runtime.BudgetHistory))
	}
}

func TestEpisodeCheckpointAndRestore(t *testing.T) {
	manager := executive.NewEpisodeManager()
	ep := buildTestEpisode(t, "ep-cp-01")
	_ = manager.CreateEpisode(ep)
	_ = manager.TransitionStatus("ep-cp-01", executive.EpisodeStatusRunning, executive.EpisodeOutcomePending, "", "")

	cp, err := manager.CreateCheckpoint("ep-cp-01", executive.CheckpointReasonPreMigration)
	if err != nil {
		t.Fatalf("CreateCheckpoint failed: %v", err)
	}
	if cp.CheckpointID == "" || cp.RuntimeFingerprint == "" {
		t.Fatalf("checkpoint fields missing: %+v", cp)
	}

	// Restore onto a fresh manager
	manager2 := executive.NewEpisodeManager()
	restored, err := manager2.RestoreFromCheckpoint(cp, ep.Definition)
	if err != nil {
		t.Fatalf("RestoreFromCheckpoint failed: %v", err)
	}
	if restored.Runtime.Status != executive.EpisodeStatusPaused || restored.Runtime.PauseReason != executive.PauseReasonCheckpointing {
		t.Fatalf("unexpected restored state: %+v", restored.Runtime)
	}
}

func TestEventDrivenOrchestrationAndCognitiveWaking(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	manager := executive.NewEpisodeManager()
	refCoord := &mockReflectionCoordinator{}
	learnCoord := &mockLearningCoordinator{}
	stratCoord := &mockStrategyCoordinator{}

	orchestrator := executive.NewEpisodeOrchestrator(manager, ws, refCoord, learnCoord, stratCoord)
	ep := buildTestEpisode(t, "ep-orch-01")
	_ = manager.CreateEpisode(ep)
	_ = manager.TransitionStatus("ep-orch-01", executive.EpisodeStatusWaiting, executive.EpisodeOutcomePending, "", "")

	ctx := context.Background()

	// Dependency check - ready by default when no dependencies registered
	if err := orchestrator.HandleEvent(ctx, executive.OrchestrationEvent{
		Type:      executive.EventDependencyResolved,
		EpisodeID: "ep-orch-01",
		Timestamp: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("HandleEvent EventDependencyResolved failed: %v", err)
	}

	// Verify status transitioned to RUNNING
	curr, _ := manager.GetEpisode("ep-orch-01")
	if curr.Runtime.Status != executive.EpisodeStatusRunning {
		t.Fatalf("expected status RUNNING after dependency resolution, got %s", curr.Runtime.Status)
	}

	// Handle decision completed event -> transitions to COMPLETED
	if err := orchestrator.HandleEvent(ctx, executive.OrchestrationEvent{
		Type:      executive.EventDecisionCompleted,
		EpisodeID: "ep-orch-01",
		Timestamp: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("HandleEvent EventDecisionCompleted failed: %v", err)
	}
	curr, _ = manager.GetEpisode("ep-orch-01")
	if curr.Runtime.Status != executive.EpisodeStatusCompleted || curr.Runtime.Outcome != executive.EpisodeOutcomeSuccess {
		t.Fatalf("expected COMPLETED/SUCCESS, got %s/%s", curr.Runtime.Status, curr.Runtime.Outcome)
	}

	// Test cognitive waking coordination
	_ = orchestrator.CoordinateReflection(ctx, "ep-orch-01", "impass contradiction detected")
	_ = orchestrator.CoordinateLearning(ctx, "ep-orch-01", "episode completed successfully")
	_ = orchestrator.CoordinateStrategyActivation(ctx, "strategy://snapshots/snap-999")

	if atomic.LoadInt32(&refCoord.wakes) != 1 || atomic.LoadInt32(&learnCoord.wakes) != 1 || atomic.LoadInt32(&stratCoord.activations) != 1 {
		t.Fatalf("expected exactly 1 wake per coordinator, got ref=%d learn=%d strat=%d",
			atomic.LoadInt32(&refCoord.wakes), atomic.LoadInt32(&learnCoord.wakes), atomic.LoadInt32(&stratCoord.activations))
	}
}

func TestBackgroundEpisodeScheduling(t *testing.T) {
	manager := executive.NewEpisodeManager()
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	orchestrator := executive.NewEpisodeOrchestrator(manager, ws, nil, nil, nil)
	ep := buildTestEpisode(t, "ep-bg-01")
	ep.Definition.EpisodeType = executive.EpisodeTypeBackground

	ctx := context.Background()
	if err := orchestrator.ScheduleBackgroundEpisode(ctx, ep); err != nil {
		t.Fatalf("ScheduleBackgroundEpisode failed: %v", err)
	}

	stored, exists := manager.GetEpisode("ep-bg-01")
	if !exists {
		t.Fatal("background episode not found")
	}
	if stored.Runtime.CurrentPriority != executive.PriorityBand3Background {
		t.Fatalf("expected priority PriorityBand3Background, got %v", stored.Runtime.CurrentPriority)
	}
}

func TestEpisodeConcurrentRaceSafety(t *testing.T) {
	manager := executive.NewEpisodeManager()
	ep := buildTestEpisode(t, "ep-race-01")
	_ = manager.CreateEpisode(ep)
	_ = manager.TransitionStatus("ep-race-01", executive.EpisodeStatusRunning, executive.EpisodeOutcomePending, "", "")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			p := executive.PriorityBand1RealTime
			if idx%2 == 0 {
				p = executive.PriorityBand2Interactive
			}
			_ = manager.UpdatePriority("ep-race-01", p, executive.PriorityReasonSalienceOverride)
			_ = manager.UpdateBudget("ep-race-01", executive.BudgetDeliberative, executive.BudgetReasonEscalationGranted)
			if idx%10 == 0 {
				_, _ = manager.CreateCheckpoint("ep-race-01", executive.CheckpointReasonPeriodic)
			}
			_, _ = manager.GetEpisode("ep-race-01")
		}(i)
	}
	wg.Wait()
}
