package communication_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
)

func TestTopicValidation(t *testing.T) {
	validTopics := communication.AllTopics()
	if len(validTopics) != 10 {
		t.Fatalf("expected 10 canonical topics, got %d", len(validTopics))
	}
	for _, topic := range validTopics {
		if !topic.IsValid() {
			t.Fatalf("expected topic %s to be valid", topic)
		}
	}

	invalidTopic := communication.TopicID("unknown-topic")
	if invalidTopic.IsValid() {
		t.Fatalf("expected unknown-topic to be invalid")
	}
}

func TestEnvelopeValidation(t *testing.T) {
	validEnv := communication.Envelope{
		ID:                "env-123",
		Source:            "CognitiveAbility.Reasoning",
		Topic:             communication.TopicCandidatePlans,
		PayloadRef:        "storage://plans/plan-01",
		RawConfidence:     0.85,
		Urgency:           10,
		CostEstimateUnits: 100,
		CreatedAt:         time.Now().UTC(),
	}

	if err := validEnv.Validate(); err != nil {
		t.Fatalf("expected valid envelope to pass validation, got: %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(e *communication.Envelope)
		wantErr error
	}{
		{
			name:    "missing id",
			mutate:  func(e *communication.Envelope) { e.ID = "" },
			wantErr: communication.ErrMissingID,
		},
		{
			name:    "missing source",
			mutate:  func(e *communication.Envelope) { e.Source = "" },
			wantErr: communication.ErrMissingSource,
		},
		{
			name:    "invalid topic",
			mutate:  func(e *communication.Envelope) { e.Topic = "invalid" },
			wantErr: communication.ErrInvalidTopic,
		},
		{
			name:    "missing payload ref",
			mutate:  func(e *communication.Envelope) { e.PayloadRef = "" },
			wantErr: communication.ErrMissingPayloadRef,
		},
		{
			name:    "confidence too high",
			mutate:  func(e *communication.Envelope) { e.RawConfidence = 1.05 },
			wantErr: communication.ErrInvalidConfidence,
		},
		{
			name:    "confidence negative",
			mutate:  func(e *communication.Envelope) { e.RawConfidence = -0.1 },
			wantErr: communication.ErrInvalidConfidence,
		},
		{
			name:    "urgency out of bounds",
			mutate:  func(e *communication.Envelope) { e.Urgency = 101 },
			wantErr: communication.ErrInvalidUrgency,
		},
		{
			name:    "negative cost",
			mutate:  func(e *communication.Envelope) { e.CostEstimateUnits = -10 },
			wantErr: communication.ErrNegativeCost,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cp := validEnv
			tc.mutate(&cp)
			err := cp.Validate()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestEffectivePriorityCalculation(t *testing.T) {
	env := communication.Envelope{
		RawConfidence:     0.80,
		Urgency:           10,
		CostEstimateUnits: 200,
	}

	// P_eff = (0.80 * 1.0) + (10 * 0.1) - ((200 / 1000) * 0.5) = 0.8 + 1.0 - 0.1 = 1.70
	score := env.EffectivePriority(1.0, 0.1, 0.5, 1000)
	expected := 1.70
	diff := score - expected
	if diff < -0.0001 || diff > 0.0001 {
		t.Fatalf("expected EffectivePriority ~1.70, got %f", score)
	}

	// Discounted by calibration weight 0.5 (overconfident module penalty)
	// P_eff = (0.80 * 0.5) + (10 * 0.1) - 0.1 = 0.4 + 1.0 - 0.1 = 1.30
	discounted := env.EffectivePriority(0.5, 0.1, 0.5, 1000)
	if discounted >= score {
		t.Fatalf("expected discounted score (%f) < undiscounted (%f)", discounted, score)
	}
}

func TestEnvelopeBuilderAndConcurrency(t *testing.T) {
	var wg sync.WaitGroup
	errs := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			env, err := communication.NewEnvelopeBuilder().
				WithSource("CognitiveAbility.Planning").
				WithTopic(communication.TopicCandidatePlans).
				WithPayloadRef("storage://plans/item").
				WithModality("structured-plan").
				WithConfidence(0.90).
				WithUrgency(5).
				WithCostEstimate(50).
				Build()
			if err != nil {
				errs <- err
				return
			}
			if env.ID == "" {
				errs <- errors.New("empty id generated")
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent build failure: %v", err)
		}
	}
}
