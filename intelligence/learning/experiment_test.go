package learning

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestExperimentManagerStartStop(t *testing.T) {
	ctx := context.Background()
	mgr := NewDefaultExperimentManager(nil)

	prof := &ExperimentProfile{
		ExperimentID:     "exp-shadow-1",
		DomainSchemaID:   "idun.reasoning.strategy.v1",
		TargetSnapshotID: "snap-1",
		ShadowRatio:      0.10,
		CanaryRatio:      0.0,
		MaxDuration:      1 * time.Hour,
		ReplaySeed:       12345,
	}

	if err := mgr.StartExperiment(ctx, prof); err != nil {
		t.Fatalf("StartExperiment failed: %v", err)
	}

	// Verify retrieval
	got, err := mgr.GetActiveExperiment(ctx, "exp-shadow-1")
	if err != nil || got.ExperimentID != "exp-shadow-1" {
		t.Errorf("GetActiveExperiment failed: %v, got %+v", err, got)
	}

	// Verify duplicate start is rejected
	if err := mgr.StartExperiment(ctx, prof); err == nil {
		t.Errorf("expected error when starting duplicate experiment ID")
	}

	// Stop experiment
	if err := mgr.StopExperiment(ctx, "exp-shadow-1"); err != nil {
		t.Fatalf("StopExperiment failed: %v", err)
	}

	// Verify stop on non-existent experiment returns ErrNotFound
	if err := mgr.StopExperiment(ctx, "exp-shadow-1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound when stopping inactive experiment, got: %v", err)
	}
}

func TestExperimentManagerBoundaryChecks(t *testing.T) {
	ctx := context.Background()
	mgr := NewDefaultExperimentManager(nil)

	// Rejection when ShadowRatio + CanaryRatio > 1.0
	invalidRatioProf := &ExperimentProfile{
		ExperimentID:     "exp-invalid",
		DomainSchemaID:   "idun.reasoning.strategy.v1",
		TargetSnapshotID: "snap-1",
		ShadowRatio:      0.80,
		CanaryRatio:      0.30, // 0.80 + 0.30 = 1.10 > 1.0
		MaxDuration:      1 * time.Hour,
	}

	if err := mgr.StartExperiment(ctx, invalidRatioProf); err == nil {
		t.Errorf("expected error when combined shadow+canary ratio exceeds 1.0")
	}

	// Rejection when max shadow limits exceeded (default 3)
	for i := 1; i <= 3; i++ {
		p := &ExperimentProfile{
			ExperimentID:     fmt.Sprintf("shadow-%d", i),
			DomainSchemaID:   "idun.reasoning.strategy.v1",
			TargetSnapshotID: "snap-1",
			ShadowRatio:      0.05,
			MaxDuration:      1 * time.Hour,
		}
		if err := mgr.StartExperiment(ctx, p); err != nil {
			t.Fatalf("failed to start shadow experiment %d: %v", i, err)
		}
	}

	// 4th shadow experiment should fail due to cardinality limit
	pOver := &ExperimentProfile{
		ExperimentID:     "shadow-4-over",
		DomainSchemaID:   "idun.reasoning.strategy.v1",
		TargetSnapshotID: "snap-1",
		ShadowRatio:      0.05,
		MaxDuration:      1 * time.Hour,
	}
	if err := mgr.StartExperiment(ctx, pOver); !errors.Is(err, ErrCardinalityExceeded) {
		t.Errorf("expected ErrCardinalityExceeded for 4th shadow experiment, got: %v", err)
	}

	// Canary limit (default 1)
	c1 := &ExperimentProfile{
		ExperimentID:     "canary-1",
		DomainSchemaID:   "idun.reasoning.strategy.v1",
		TargetSnapshotID: "snap-1",
		CanaryRatio:      0.05,
		MaxDuration:      1 * time.Hour,
	}
	if err := mgr.StartExperiment(ctx, c1); err != nil {
		t.Fatalf("failed to start canary experiment: %v", err)
	}
	c2Over := &ExperimentProfile{
		ExperimentID:     "canary-2-over",
		DomainSchemaID:   "idun.reasoning.strategy.v1",
		TargetSnapshotID: "snap-1",
		CanaryRatio:      0.05,
		MaxDuration:      1 * time.Hour,
	}
	if err := mgr.StartExperiment(ctx, c2Over); !errors.Is(err, ErrCardinalityExceeded) {
		t.Errorf("expected ErrCardinalityExceeded for 2nd canary experiment, got: %v", err)
	}
}

func TestExperimentManagerConcurrentSafety(t *testing.T) {
	ctx := context.Background()
	mgr := NewDefaultExperimentManager(nil)

	var wg sync.WaitGroup
	// Start concurrent reads and starts/stops across unique IDs
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			expID := fmt.Sprintf("exp-conc-%d", id)
			prof := &ExperimentProfile{
				ExperimentID:     expID,
				DomainSchemaID:   "idun.reasoning.strategy.v1",
				TargetSnapshotID: "snap-1",
				ShadowRatio:      0.01,
				MaxDuration:      1 * time.Hour,
			}
			// Attempt start (some may hit cardinality limit, which is fine and thread-safe)
			_ = mgr.StartExperiment(ctx, prof)
			_ = mgr.ListActiveExperiments(ctx)
			_, _ = mgr.GetActiveExperiment(ctx, expID)
			_ = mgr.StopExperiment(ctx, expID)
		}(i)
	}
	wg.Wait()
}

func TestExperimentPrioritization(t *testing.T) {
	ctx := context.Background()
	mgr := NewDefaultExperimentManager(nil)

	p1 := &ExperimentProfile{
		ExperimentID:     "exp-low-prio",
		DomainSchemaID:   "idun.reasoning.strategy.v1",
		TargetSnapshotID: "snap-1",
		ShadowRatio:      0.10,
		MaxDuration:      1 * time.Hour,
		Priority:         1,
	}
	p2 := &ExperimentProfile{
		ExperimentID:     "exp-high-prio",
		DomainSchemaID:   "idun.reasoning.strategy.v1",
		TargetSnapshotID: "snap-2",
		ShadowRatio:      0.10,
		MaxDuration:      1 * time.Hour,
		Priority:         10,
	}
	p3 := &ExperimentProfile{
		ExperimentID:     "exp-med-prio",
		DomainSchemaID:   "idun.reasoning.strategy.v1",
		TargetSnapshotID: "snap-3",
		ShadowRatio:      0.20,
		MaxDuration:      1 * time.Hour,
		Priority:         5,
	}

	_ = mgr.StartExperiment(ctx, p1)
	_ = mgr.StartExperiment(ctx, p2)
	_ = mgr.StartExperiment(ctx, p3)

	list := mgr.ListActiveExperimentsPrioritized(ctx)
	if len(list) != 3 {
		t.Fatalf("expected 3 prioritized experiments, got %d", len(list))
	}
	if list[0].ExperimentID != "exp-high-prio" || list[1].ExperimentID != "exp-med-prio" || list[2].ExperimentID != "exp-low-prio" {
		t.Errorf("unexpected priority order: %v, %v, %v", list[0].ExperimentID, list[1].ExperimentID, list[2].ExperimentID)
	}
}
