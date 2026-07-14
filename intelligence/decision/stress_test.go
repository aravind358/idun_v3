package decision

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestDecisionService_StressConcurrentPipelines(t *testing.T) {
	service := NewService()
	var wg sync.WaitGroup
	workers := 50
	iterations := 100

	ctx := context.Background()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				cset := CandidateSet{
					EpisodeID: fmt.Sprintf("ep-stress-%d", workerID),
					Candidates: []Candidate{
						{
							ID: "c1",
							Attributes: map[string]float64{
								"utility": 0.95,
								"safety":  1.0,
							},
						},
						{
							ID: "c2",
							Attributes: map[string]float64{
								"utility": 0.85,
								"safety":  0.90,
							},
						},
					},
				}

				rec, err := service.EvaluateReflexive(ctx, cset)
				if err != nil {
					t.Errorf("EvaluateReflexive failed: %v", err)
					return
				}
				if err := rec.Validate(); err != nil {
					t.Errorf("Validate failed on returned record: %v", err)
					return
				}
			}
		}(w)
	}

	wg.Wait()
}
