package reflection

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

type failingSpecialistEvaluator struct{}

func (f *failingSpecialistEvaluator) ID() string { return "failing-specialist" }
func (f *failingSpecialistEvaluator) TargetAbility() executive.CognitiveAbility { return executive.AbilityReasoning }
func (f *failingSpecialistEvaluator) EvaluateEpisode(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error) {
	return SpecialistReport{}, errors.New("simulated evaluation failure")
}

func TestStress_ConcurrentReflectionsAndCancellationStorms(t *testing.T) {
	ws := workspace.NewEngine()
	srv := NewService(WithWorkspace(ws))
	_ = srv.Start()
	defer srv.Close()

	const workers = 20
	var wg sync.WaitGroup

	// 1. Concurrent Episode Reflection under load
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			traces := []communication.Envelope{
				{ID: fmt.Sprintf("env-%d", id), Source: "stress"},
			}
			_, _ = srv.ReflectEpisode(context.Background(), fmt.Sprintf("ep-%d", id), traces)
		}(i)
	}

	// 2. Concurrent Periodic Reflection under load
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sum := HistoricalSummary{
				SchemaVersion:      SchemaVersion,
				SummaryID:          fmt.Sprintf("sum-%d", id),
				GeneratedTimestamp: 100,
				TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
				AverageScores: map[string]float64{
					"Reasoning": 0.80,
				},
				SummaryConfidence: 0.90,
			}
			_, _ = srv.ReflectPeriodic(context.Background(), sum)
		}(i)
	}

	// 3. Concurrent Reflection-on-Reflection under load
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sum := HistoricalSummary{
				SchemaVersion:      SchemaVersion,
				SummaryID:          fmt.Sprintf("sum-meta-%d", id),
				GeneratedTimestamp: 100,
				TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
				AverageScores: map[string]float64{
					"Reasoning": 0.85,
				},
				SummaryConfidence: 0.90,
			}
			_, _ = srv.ReflectOnReflection(context.Background(), nil, sum)
		}(i)
	}

	// 4. Cancellation storms
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // cancel immediately
			_, _ = srv.ReflectEpisode(ctx, fmt.Sprintf("ep-cancel-%d", id), nil)
		}(i)
	}

	wg.Wait()

	snap := srv.Telemetry()
	if snap.TotalEpisodeReflections == 0 || snap.TotalPeriodicReflections == 0 || snap.TotalMetaReflections == 0 {
		t.Errorf("expected positive telemetry counts under stress: %+v", snap)
	}
	if snap.CancellationCount != workers {
		t.Errorf("got %d cancellation count, want %d", snap.CancellationCount, workers)
	}
}

func TestStress_ShutdownDuringEvaluation(t *testing.T) {
	srv := NewService()
	_ = srv.Start()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(2 * time.Millisecond)
			_, _ = srv.ReflectEpisode(context.Background(), fmt.Sprintf("ep-shut-%d", id), nil)
		}(i)
	}

	time.Sleep(1 * time.Millisecond)
	_ = srv.Close()
	wg.Wait()
}
