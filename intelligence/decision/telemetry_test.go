package decision

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestReflexiveDecisionTrace_RecordAndAnomalyBuffer(t *testing.T) {
	trace := NewReflexiveDecisionTrace("ep-test", "v1.0")

	// 1. Record 25 decisions with anomalies to verify bounded ring buffer (max 16)
	for i := 1; i <= 25; i++ {
		rec := &DecisionRecord{
			DecisionID:      fmt.Sprintf("dec-%d", i),
			SelectedOutcome: OutcomeCommit,
			Confidence:      0.85,
		}
		anomaly := &MicroDecisionAnomaly{
			DecisionID:      rec.DecisionID,
			Timestamp:       time.Now(),
			AnomalyType:     "CONSTITUTIONAL_TENSION",
			TopCandidateID:  "cand-x",
			ConfidenceScore: 0.85,
		}
		trace.RecordDecision(rec, uint32(i*10), 0.15, false, anomaly)
	}

	snap := trace.Snapshot()

	if snap.TotalEvaluated != 25 {
		t.Errorf("expected 25 total evaluated, got %d", snap.TotalEvaluated)
	}
	if snap.CommitCount != 25 {
		t.Errorf("expected 25 commit count, got %d", snap.CommitCount)
	}
	if len(snap.Anomalies) != 16 {
		t.Fatalf("expected bounded anomaly buffer length 16, got %d", len(snap.Anomalies))
	}

	// Verify that the oldest anomaly was dropped and dec-10 is the first entry
	if snap.Anomalies[0].DecisionID != "dec-10" {
		t.Errorf("expected oldest anomaly in buffer to be 'dec-10', got '%s'", snap.Anomalies[0].DecisionID)
	}
	if snap.Anomalies[15].DecisionID != "dec-25" {
		t.Errorf("expected newest anomaly in buffer to be 'dec-25', got '%s'", snap.Anomalies[15].DecisionID)
	}
}

func TestReflexiveDecisionTrace_ConcurrentHighFrequencyRecording(t *testing.T) {
	trace := NewReflexiveDecisionTrace("ep-concurrent", "v1.0")

	const workers = 20
	const decisionsPerWorker = 500

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < decisionsPerWorker; i++ {
				rec := &DecisionRecord{
					DecisionID:      fmt.Sprintf("dec-%d-%d", workerID, i),
					SelectedOutcome: OutcomeCommit,
					Confidence:      0.80,
				}
				trace.RecordDecision(rec, 120, 0.10, i%10 == 0, nil)
			}
		}(w)
	}

	wg.Wait()

	snap := trace.Snapshot()
	expectedTotal := uint64(workers * decisionsPerWorker)
	if snap.TotalEvaluated != expectedTotal {
		t.Errorf("expected TotalEvaluated %d, got %d", expectedTotal, snap.TotalEvaluated)
	}
	if snap.CommitCount != expectedTotal {
		t.Errorf("expected CommitCount %d, got %d", expectedTotal, snap.CommitCount)
	}
}
