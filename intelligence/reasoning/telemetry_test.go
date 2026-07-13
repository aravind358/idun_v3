package reasoning

import (
	"context"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
)

func TestService_TelemetryThreadSafety(t *testing.T) {
	srv := NewService(DefaultConfig(), nil, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("unexpected Start error: %v", err)
	}
	defer srv.Close()

	var wg sync.WaitGroup
	const numGoroutines = 30

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			env := communication.Envelope{
				ID:            "env-telemetry",
				Source:        "test",
				RawConfidence: 0.9,
			}
			_, _ = srv.ReasonEnvelope(ctx, env, StrategySpec{})
		}()
	}

	wg.Wait()

	snap := srv.GetTelemetry()
	if snap.TotalReasoningEpisodes != numGoroutines {
		t.Errorf("expected TotalReasoningEpisodes=%d, got %d", numGoroutines, snap.TotalReasoningEpisodes)
	}
	if snap.BeamSelections != numGoroutines {
		t.Errorf("expected BeamSelections=%d, got %d", numGoroutines, snap.BeamSelections)
	}
	if snap.AvgBeamWidth < 1.0 {
		t.Errorf("expected AvgBeamWidth >= 1.0, got %f", snap.AvgBeamWidth)
	}
	if snap.AvgReasoningDurationMs < 0.0 {
		t.Errorf("expected AvgReasoningDurationMs >= 0.0, got %f", snap.AvgReasoningDurationMs)
	}
}
