package understanding_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/understanding"
)

func TestService_StressConcurrencyAndTelemetry(t *testing.T) {
	ws := &mockDeliberativeWorkspace{}
	svc := understanding.NewService(
		understanding.WithConfigOptions(),
		ws,
	)

	const numGoroutines = 100
	const iterationsPerGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterationsPerGoroutine; i++ {
				utterance := "status"
				if (id+i)%2 == 0 {
					utterance = "Can we reschedule the meeting?"
				}
				env := communication.Envelope{
					ID:         fmt.Sprintf("stress-%d-%d", id, i),
					PayloadRef: utterance,
				}
				_, err := svc.InterpretEnvelope(context.Background(), env)
				if err != nil {
					t.Errorf("stress InterpretEnvelope failed: %v", err)
					return
				}
			}
		}(g)
	}

	wg.Wait()

	telemetry := svc.GetTelemetry()
	expectedTotal := int64(numGoroutines * iterationsPerGoroutine)
	if telemetry.TotalInterpretations != expectedTotal {
		t.Fatalf("expected %d total interpretations in telemetry, got %d", expectedTotal, telemetry.TotalInterpretations)
	}

	ws.mu.Lock()
	pubCount := len(ws.published)
	ws.mu.Unlock()
	if int64(pubCount) != expectedTotal {
		t.Fatalf("expected %d publications to workspace, got %d", expectedTotal, pubCount)
	}
}

func TestService_StressCancellationStormAndShutdown(t *testing.T) {
	infSvc := &mockInferenceService{
		delay: 50 * time.Millisecond,
	}
	worker := understanding.NewDeliberativeWorker(infSvc, nil, 100*time.Millisecond)
	svc := understanding.NewService(
		understanding.WithConfigOptions(),
		nil,
		understanding.WithDeliberativeWorker(worker),
	)

	const numWorkers = 50
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(id%15)*time.Millisecond)
			defer cancel()

			env := communication.Envelope{
				ID:         fmt.Sprintf("storm-%d", id),
				PayloadRef: "complex utterance triggering slow deliberative worker",
			}
			_, _ = svc.InterpretEnvelope(ctx, env)
		}(i)
	}

	time.Sleep(10 * time.Millisecond)
	_ = svc.Close()
	wg.Wait()
}
