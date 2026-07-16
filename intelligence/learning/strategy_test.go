package learning

import (
	"sync"
	"testing"
	"time"
)

func TestDefaultStrategyProvider(t *testing.T) {
	sp, err := NewDefaultStrategyProvider(nil)
	if err != nil {
		t.Fatalf("failed to init strategy provider: %v", err)
	}

	snap := sp.ActiveSnapshot()
	if snap == nil || snap.SnapshotID != "snap-learning-default-v2.0.0" {
		t.Fatalf("unexpected default snapshot ID: %v", snap)
	}

	nextSnap := &LearningStrategySnapshot{
		SnapshotID:    "snap-updated-v2.0.0",
		SchemaVersion: SchemaVersion,
		ActiveProfile: DefaultLearningPolicyProfile(),
		Capabilities:  DefaultLearningCapabilities(),
		CreatedAt:     time.Now(),
	}

	if err := sp.SwapSnapshot(nextSnap); err != nil {
		t.Fatalf("failed to swap snapshot: %v", err)
	}

	if current := sp.ActiveSnapshot(); current.SnapshotID != "snap-updated-v2.0.0" {
		t.Errorf("expected snap-updated-v2.0.0, got %q", current.SnapshotID)
	}
}

func TestStrategyProviderConcurrentSafety(t *testing.T) {
	sp, err := NewDefaultStrategyProvider(nil)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	var wg sync.WaitGroup
	workers := 10
	iterations := 100

	// Concurrent readers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				snap := sp.ActiveSnapshot()
				if snap == nil || snap.SchemaVersion != SchemaVersion {
					t.Errorf("invalid snapshot read concurrently")
				}
			}
		}()
	}

	// Concurrent swappers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				snap := &LearningStrategySnapshot{
					SnapshotID:    "snap-worker-update",
					SchemaVersion: SchemaVersion,
					ActiveProfile: DefaultLearningPolicyProfile(),
					Capabilities:  DefaultLearningCapabilities(),
					CreatedAt:     time.Now(),
				}
				_ = sp.SwapSnapshot(snap)
			}
		}(i)
	}

	wg.Wait()
}
