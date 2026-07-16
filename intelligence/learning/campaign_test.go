package learning

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/core/memory"
)

func TestLearningCampaignAndSummaryBuildersAndValidation(t *testing.T) {
	now := time.Now()
	camp, err := NewLearningCampaignBuilder(
		"camp-planning-opt-01",
		"Improve planning heuristic consistency across high-complexity domains",
		now.Add(-24*time.Hour),
		now.Add(24*time.Hour),
		"fp-camp-sha256",
		"fp-pol-sha256",
	).
		WithStatus(CampaignStatusActive).
		WithCounters(300, 95, 80, 8, 2).
		Build()
	if err != nil {
		t.Fatalf("campaign build failed: %v", err)
	}

	if camp.CampaignID != "camp-planning-opt-01" || camp.CampaignStatus != CampaignStatusActive || camp.LearningCycles != 300 {
		t.Errorf("unexpected campaign fields: %+v", camp)
	}

	// Test invalid campaign
	_, errInvalid := NewLearningCampaignBuilder("", "Objective", now, now, "fp", "fp").Build()
	if errInvalid == nil {
		t.Errorf("expected error building campaign with empty ID")
	}

	summary, errSum := NewLearningCampaignSummaryBuilder("camp-planning-opt-01").
		WithTotals(300, 95, 80, 2, 5).
		WithAverages(0.88, 1450, 0.99).
		WithDuration(48 * time.Hour).
		Build()
	if errSum != nil {
		t.Fatalf("campaign summary build failed: %v", errSum)
	}

	if summary.TotalCycles != 300 || summary.AverageValidationConfidence != 0.88 {
		t.Errorf("unexpected summary fields: %+v", summary)
	}

	// Test invalid summary bounds
	_, errSumInvalid := NewLearningCampaignSummaryBuilder("camp-01").
		WithAverages(1.5, 100, 0.5). // out of bounds confidence
		Build()
	if errSumInvalid == nil {
		t.Errorf("expected error building summary with confidence > 1.0")
	}
}

func TestLearningCampaignConcurrencySafetyAndPropagation(t *testing.T) {
	ctx := context.Background()
	store := newMockMemory()
	now := time.Now()
	for i := 0; i < 100; i++ {
		_ = store.CreateRecord(memory.Record{
			ID:        fmt.Sprintf("t-%d", i),
			Type:      "idun.reasoning.trace.v1",
			Payload:   []byte(`{"step":"mock"}`),
			CreatedAt: now,
		})
	}

	agg := NewDefaultAggregator(store)
	val := NewDefaultValidationPipeline(nil)
	svc, err := NewService(WithAggregator(agg), WithValidationPipeline(val))
	if err != nil {
		t.Fatalf("service init failed: %v", err)
	}
	_ = svc.Start()
	defer svc.Close()

	var wg sync.WaitGroup
	const workers = 10
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			req, _ := NewLearningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-camp-%d", workerID)).
				WithCampaignID("camp-concurrent-verification").
				WithDomainSchemaID("idun.reasoning.strategy.v1").
				WithTimeWindow(now.Add(-2*time.Hour), now).
				WithPolicyFingerprint(svc.config.PolicyProfile.PolicyFingerprint).
				Build()

			res, errCycle := svc.RunCycle(ctx, req)
			if errCycle != nil {
				t.Errorf("RunCycle failed for worker %d: %v", workerID, errCycle)
				return
			}
			if res.CampaignID != "camp-concurrent-verification" {
				t.Errorf("expected res.CampaignID camp-concurrent-verification, got %s", res.CampaignID)
			}
			if len(res.Traces) > 0 && res.Traces[0].CampaignID != "camp-concurrent-verification" {
				t.Errorf("expected trace.CampaignID camp-concurrent-verification, got %s", res.Traces[0].CampaignID)
			}
		}(w)
	}
	wg.Wait()
}
