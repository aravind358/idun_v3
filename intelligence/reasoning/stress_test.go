package reasoning

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/constitution"
)

func TestService_StressStormsAndConcurrentLifecycle(t *testing.T) {
	mockCal := &mockCalibService{multiplier: 0.90}
	mockInf := &mockInferenceService{}
	mockGate := &mockActionGate{verdict: constitution.VerdictApproved, sig: "STRESS-SIG"}

	srv := NewService(DefaultConfig(), nil, nil,
		WithCalibrationService(mockCal),
		WithInferenceService(mockInf),
		WithConstitutionGate(mockGate),
	)

	if err := srv.Start(); err != nil {
		t.Fatalf("unexpected Start error: %v", err)
	}

	var wg sync.WaitGroup
	const numEpisodes = 150

	for i := 0; i < numEpisodes; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			var ctx context.Context
			var cancel context.CancelFunc

			switch idx % 4 {
			case 0:
				// Normal execution
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
			case 1:
				// Cancellation storm (cancel almost immediately)
				ctx, cancel = context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Millisecond)
					cancel()
				}()
			case 2:
				// Timeout storm (microsecond timeout)
				ctx, cancel = context.WithTimeout(context.Background(), 10*time.Microsecond)
			case 3:
				// Normal execution with deliberative escalation trigger
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
			}
			defer cancel()

			env := communication.Envelope{
				ID:            fmt.Sprintf("stress-env-%d", idx),
				Source:        "stress.test",
				Topic:         communication.TopicResolvedIntent,
				RawConfidence: 0.85,
			}

			_, _ = srv.ReasonEnvelope(ctx, env, StrategySpec{})
		}(i)
	}

	// Trigger concurrent shutdown while episodes are actively processing
	go func() {
		time.Sleep(15 * time.Millisecond)
		_ = srv.Close()
	}()

	wg.Wait()
}
