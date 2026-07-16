package scheduler_test

import (
	"errors"
	"testing"
	"time"
)

func TestSchedulePeriodic_SuccessAndRetry(t *testing.T) {
	disp := newMockDispatcher()
	s, mem := newTestScheduler(t, disp)

	// Schedule a periodic job every 40ms
	err := s.SchedulePeriodic("job/periodic", "TargetPeriodic", []byte("ping"), 40*time.Millisecond)
	if err != nil {
		t.Fatalf("SchedulePeriodic: %v", err)
	}

	// Wait for the first dispatch
	select {
	case <-disp.notify:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for first periodic dispatch")
	}

	calls := disp.getCalls()
	if len(calls) == 0 || calls[0].Target != "TargetPeriodic" {
		t.Fatalf("unexpected calls after first dispatch: %+v", calls)
	}

	// Verify that record still exists in Memory and was updated for next cycle
	rec, err := mem.GetRecord("job/periodic")
	if err != nil {
		t.Fatalf("expected periodic job to persist in memory after dispatch, got err: %v", err)
	}
	if len(rec.Payload) == 0 {
		t.Fatal("expected non-empty record payload")
	}

	// Now make the dispatcher return an error to test retry backoff
	disp.err = errors.New("temporary failure")

	// Wait for second dispatch attempt (and retry)
	select {
	case <-disp.notify:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for second periodic dispatch attempt")
	}

	time.Sleep(15 * time.Millisecond)

	// Cancel the job cleanly
	if err := s.Cancel("job/periodic"); err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}
}
