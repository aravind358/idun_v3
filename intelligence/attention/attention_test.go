package attention_test

import (
	"sync"
	"testing"

	"idun/intelligence/attention"
)

func TestAttentionServiceEvaluationAndGoal(t *testing.T) {
	svc := attention.NewService()
	if svc.Name() != "Intelligence.Attention" {
		t.Fatalf("expected Name Intelligence.Attention, got %q", svc.Name())
	}

	goal := attention.ActiveGoalContext{
		ID:             "goal-att-1",
		Summary:        "Test attention tracking",
		PriorityWeight: 100,
	}
	svc.SetActiveGoal(goal)
	if got := svc.GetActiveGoal(); got.ID != "goal-att-1" {
		t.Fatalf("expected goal ID goal-att-1, got %q", got.ID)
	}

	tests := []struct {
		name         string
		stimulus     attention.Stimulus
		wantDecision attention.SalienceDecision
		wantBand     attention.PriorityBand
	}{
		{
			name:         "safety tripwire overrides score",
			stimulus:     attention.Stimulus{ID: "s1", SafetyFlag: true, SalienceScore: 10},
			wantDecision: attention.SalienceFocusImmediately,
			wantBand:     attention.PriorityBand0CriticalSafety,
		},
		{
			name:         "high salience real-time focus",
			stimulus:     attention.Stimulus{ID: "s2", SalienceScore: 90},
			wantDecision: attention.SalienceFocusImmediately,
			wantBand:     attention.PriorityBand1RealTime,
		},
		{
			name:         "interactive dialogue focus",
			stimulus:     attention.Stimulus{ID: "s3", SalienceScore: 60},
			wantDecision: attention.SalienceFocusImmediately,
			wantBand:     attention.PriorityBand2Interactive,
		},
		{
			name:         "background schedule",
			stimulus:     attention.Stimulus{ID: "s4", SalienceScore: 30},
			wantDecision: attention.SalienceSchedule,
			wantBand:     attention.PriorityBand3Background,
		},
		{
			name:         "low salience filter",
			stimulus:     attention.Stimulus{ID: "s5", SalienceScore: 10},
			wantDecision: attention.SalienceFilter,
			wantBand:     attention.PriorityBand4Idle,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dec, band := svc.Evaluate(tc.stimulus)
			if dec != tc.wantDecision {
				t.Errorf("expected decision %s, got %s", tc.wantDecision, dec)
			}
			if band != tc.wantBand {
				t.Errorf("expected priority band %d, got %d", tc.wantBand, band)
			}
		})
	}
}

func TestAttentionServiceConcurrency(t *testing.T) {
	svc := attention.NewService()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				svc.SetActiveGoal(attention.ActiveGoalContext{ID: "goal-idx"})
			} else {
				_ = svc.GetActiveGoal()
				_, _ = svc.Evaluate(attention.Stimulus{ID: "s-idx", SalienceScore: idx * 2})
			}
		}(i)
	}
	wg.Wait()
}
