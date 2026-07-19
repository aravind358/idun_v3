package understanding_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"idun/intelligence/communication"
	"idun/intelligence/understanding"
)

func TestService_Phase5_Hardening_NormalInputs(t *testing.T) {
	ws := &mockWorkspace{}
	storer := newMockPayloadStorer()
	cfg := understanding.WithConfigOptions()
	svc := understanding.NewService(cfg, ws, understanding.WithPayloadStorer(storer))

	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer svc.Close()

	testCases := []struct {
		input          string
		expectedIntent string
	}{
		{"Hello", "greet_user"},
		{"Hi", "greet_user"},
		{"Who are you?", "query_identity"},
		{"How are you?", "query_wellbeing"},
		{"Goodbye", "farewell_user"},
		{"status", "query_status"},
	}

	norm := understanding.NewDefaultNormalizer()

	for i, tc := range testCases {
		ctx := context.Background()
		envID := fmt.Sprintf("normal-env-%d", i)
		casKey, err := storer.Store(ctx, []byte(tc.input))
		if err != nil {
			t.Fatalf("Store failed: %v", err)
		}

		perceptionEnv := communication.Envelope{
			ID:         envID,
			Source:     "World.Service",
			Topic:      communication.TopicPerception,
			PayloadRef: casKey,
		}

		ws.mu.Lock()
		handler := ws.subscriptions[communication.TopicPerception]
		ws.mu.Unlock()

		if err := handler(ctx, perceptionEnv); err != nil {
			t.Fatalf("handlePerceptionEnvelope failed for %q: %v", tc.input, err)
		}

		ws.mu.Lock()
		pubEnv := ws.published[len(ws.published)-1]
		ws.mu.Unlock()

		if pubEnv.Topic != communication.TopicUserIntent {
			t.Fatalf("expected TopicUserIntent, got %q", pubEnv.Topic)
		}

		frameData, err := storer.Retrieve(ctx, pubEnv.PayloadRef)
		if err != nil {
			t.Fatalf("Retrieve failed for CAS payload %q: %v", pubEnv.PayloadRef, err)
		}

		var frame understanding.SemanticFrame
		if err := json.Unmarshal(frameData, &frame); err != nil {
			t.Fatalf("Unmarshal SemanticFrame failed: %v", err)
		}

		if err := frame.Validate(); err != nil {
			t.Fatalf("SemanticFrame validation failed for %q: %v", tc.input, err)
		}

		if frame.PrimaryHypothesis.Intent != tc.expectedIntent {
			t.Fatalf("for input %q expected intent %q, got %q", tc.input, tc.expectedIntent, frame.PrimaryHypothesis.Intent)
		}

		normalized := norm.Normalize(tc.input)

		t.Logf("=== Test Case %d: %q ===", i, tc.input)
		t.Logf("Normalized: Cleaned=%q Tokens=%v", normalized.Cleaned, normalized.Tokens)
		t.Logf("SemanticFrame: Intent=%s Confidence=%.2f Status=%s Slots=%d Hypotheses=%d",
			frame.PrimaryHypothesis.Intent, frame.PrimaryHypothesis.CalibratedConfidence, frame.Status, len(frame.PrimaryHypothesis.Slots), len(frame.AmbiguitySet)+1)
		t.Logf("CAS PayloadRef: %s", pubEnv.PayloadRef)
		t.Logf("Published TopicUserIntent: ID=%s ParentRef=%s Source=%s RawConfidence=%.2f",
			pubEnv.ID, pubEnv.ParentRef, pubEnv.Source, pubEnv.RawConfidence)
	}
}

func TestService_Phase5_Hardening_EdgeCases(t *testing.T) {
	storer := newMockPayloadStorer()
	svc := understanding.NewService(understanding.WithConfigOptions(), nil, understanding.WithPayloadStorer(storer))
	_ = svc.Start()
	defer svc.Close()

	ctx := context.Background()

	edgeCases := []struct {
		name           string
		input          string
		expectedIntent string
		expectedStatus understanding.InterpretationStatus
	}{
		{"empty input", "", "unresolved_intent", understanding.StatusFailedImpasse},
		{"whitespace only", "   \t\n  ", "unresolved_intent", understanding.StatusFailedImpasse},
		{"very long input", strings.Repeat("hello ", 500), "greet_user", understanding.StatusUnambiguous},
		{"malformed UTF-8", "\xff\xfe\xfd not utf8", "unresolved_intent", understanding.StatusFailedImpasse},
		{"punctuation only", "!!!???...", "unresolved_intent", understanding.StatusFailedImpasse},
		{"mixed capitalization", "HeLLo IdUn", "greet_user", understanding.StatusUnambiguous},
		{"repeated input", "hello hello hello", "greet_user", understanding.StatusUnambiguous},
		{"unknown intent", "what is the meaning of quantum chromodynamics?", "unresolved_intent", understanding.StatusFailedImpasse},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			env := communication.Envelope{
				ID:         "edge-" + tc.name,
				PayloadRef: tc.input,
			}
			frame, err := svc.InterpretEnvelope(ctx, env)
			if err != nil {
				t.Fatalf("InterpretEnvelope failed for %s: %v", tc.name, err)
			}
			if err := frame.Validate(); err != nil {
				t.Fatalf("frame.Validate() failed for %s: %v", tc.name, err)
			}
			if frame.PrimaryHypothesis.Intent != tc.expectedIntent {
				t.Fatalf("expected intent %q, got %q", tc.expectedIntent, frame.PrimaryHypothesis.Intent)
			}
			if frame.Status != tc.expectedStatus {
				t.Fatalf("expected status %s, got %s", tc.expectedStatus, frame.Status)
			}
		})
	}

	t.Run("cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := svc.InterpretEnvelope(cancelCtx, communication.Envelope{ID: "cancel-env", PayloadRef: "hello"})
		if err == nil {
			t.Fatalf("expected cancellation error")
		}
	})

	t.Run("service shutdown during processing", func(t *testing.T) {
		shutSvc := understanding.NewService(understanding.WithConfigOptions(), nil)
		_ = shutSvc.Start()
		_ = shutSvc.Close()
		_, err := shutSvc.InterpretEnvelope(ctx, communication.Envelope{ID: "shut-env", PayloadRef: "hello"})
		if !errors.Is(err, understanding.ErrServiceClosed) {
			t.Fatalf("expected ErrServiceClosed, got %v", err)
		}
	})
}

func TestService_Phase5_Hardening_ConcurrencyAndParentRef(t *testing.T) {
	ws := &mockWorkspace{}
	storer := newMockPayloadStorer()
	svc := understanding.NewService(understanding.WithConfigOptions(), ws, understanding.WithPayloadStorer(storer))
	_ = svc.Start()
	defer svc.Close()

	ctx := context.Background()
	const numConcurrent = 50
	var wg sync.WaitGroup
	wg.Add(numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(idx int) {
			defer wg.Done()
			parentRef := fmt.Sprintf("world-parent-%d", idx)
			envID := fmt.Sprintf("world-env-%d", idx)
			input := "Hello"
			if idx%2 == 0 {
				input = "Who are you?"
			}

			casKey, _ := storer.Store(ctx, []byte(input))
			env := communication.Envelope{
				ID:         envID,
				ParentRef:  parentRef,
				Source:     "World.Service",
				Topic:      communication.TopicPerception,
				PayloadRef: casKey,
			}

			ws.mu.Lock()
			handler := ws.subscriptions[communication.TopicPerception]
			ws.mu.Unlock()

			if err := handler(ctx, env); err != nil {
				t.Errorf("concurrent handler failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	ws.mu.Lock()
	pubCount := len(ws.published)
	pubCopy := append([]communication.Envelope(nil), ws.published...)
	ws.mu.Unlock()

	if pubCount != numConcurrent {
		t.Fatalf("expected %d publications, got %d", numConcurrent, pubCount)
	}

	for _, pubEnv := range pubCopy {
		if !strings.HasPrefix(pubEnv.ParentRef, "world-parent-") {
			t.Fatalf("unexpected ParentRef %q", pubEnv.ParentRef)
		}
		frameData, err := storer.Retrieve(ctx, pubEnv.PayloadRef)
		if err != nil {
			t.Fatalf("Retrieve failed: %v", err)
		}
		var frame understanding.SemanticFrame
		if err := json.Unmarshal(frameData, &frame); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if err := frame.Validate(); err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
	}
}
