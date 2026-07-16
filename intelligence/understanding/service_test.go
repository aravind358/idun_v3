package understanding_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
	"idun/intelligence/understanding"
	"idun/intelligence/workspace"
)

// mockWorkspace captures envelopes published to the workspace for test assertions.
type mockWorkspace struct {
	mu        sync.Mutex
	published []communication.Envelope
}

func (m *mockWorkspace) Publish(ctx context.Context, env communication.Envelope, opts ...workspace.PublishOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, env)
	return nil
}

func (m *mockWorkspace) Subscribe(topic communication.TopicID, subscriberID string, handler workspace.EnvelopeHandler) (workspace.Subscription, error) {
	return nil, nil
}

func (m *mockWorkspace) SubscribeAll(subscriberID string, handler workspace.EnvelopeHandler) ([]workspace.Subscription, error) {
	return nil, nil
}

func (m *mockWorkspace) Unsubscribe(id workspace.SubscriptionID) error {
	return nil
}

func (m *mockWorkspace) GetEnvelope(id string) (communication.Envelope, bool) {
	return communication.Envelope{}, false
}

func (m *mockWorkspace) ListTopicEnvelopes(topic communication.TopicID, limit int) []communication.Envelope {
	return nil
}

func (m *mockWorkspace) StorePendingCandidate(ctx context.Context, topic communication.TopicID, candidate workspace.PendingCandidate) error {
	return nil
}
func (m *mockWorkspace) GetPendingCandidates(topic communication.TopicID) []workspace.PendingCandidate {
	return nil
}
func (m *mockWorkspace) RemovePendingCandidate(topic communication.TopicID, envelopeID string) bool {
	return false
}

func (m *mockWorkspace) RegisterEpisodeDependencies(ctx context.Context, epID string, dependsOn []string) error {
	return nil
}
func (m *mockWorkspace) IsEpisodeReady(ctx context.Context, epID string) (bool, error) {
	return true, nil
}
func (m *mockWorkspace) ResolveDependencies(ctx context.Context, epID string) error {
	return nil
}
func (m *mockWorkspace) NotifyDependencyComplete(ctx context.Context, depID string) error {
	return nil
}
func (m *mockWorkspace) RegisterEpisodeChild(ctx context.Context, parentID string, childID string) error {
	return nil
}
func (m *mockWorkspace) GetEpisodeChildren(ctx context.Context, parentID string) ([]string, error) {
	return nil, nil
}

func (m *mockWorkspace) Name() string { return "MockWorkspace" }
func (m *mockWorkspace) Start() error { return nil }
func (m *mockWorkspace) Close() error { return nil }

func TestDefaultNormalizer(t *testing.T) {
	normalizer := understanding.NewDefaultNormalizer()
	norm := normalizer.Normalize("   Set   ALARM   for  07:00!  Call her today. ")

	expectedCleaned := "set alarm for 07:00! call her today."
	if norm.Cleaned != expectedCleaned {
		t.Fatalf("expected cleaned %q, got %q", expectedCleaned, norm.Cleaned)
	}
	if len(norm.PronounsFound) == 0 || norm.PronounsFound[0] != "her" {
		t.Fatalf("expected pronoun 'her' detected, got %v", norm.PronounsFound)
	}
}

func TestDefaultReferentBinder(t *testing.T) {
	normalizer := understanding.NewDefaultNormalizer()
	binder := understanding.NewDefaultReferentBinder()

	candidates := []understanding.ReferentCandidate{
		{ID: "person-alex", Name: "Alex", Role: "Friend"},
	}

	norm := normalizer.Normalize("Tell her I will be late")
	slots := binder.BindReferents(norm, candidates)

	if len(slots) == 0 {
		t.Fatalf("expected pronominal slot bound to person-alex")
	}
	if slots[0].GroundingID != "person-alex" || slots[0].Value != "her" {
		t.Fatalf("unexpected bound slot: %+v", slots[0])
	}
}

func TestDefaultGrammarSpecialist(t *testing.T) {
	grammar := understanding.NewDefaultGrammarSpecialist()
	normalizer := understanding.NewDefaultNormalizer()

	// 1. Exact match
	normStatus := normalizer.Normalize("  STATUS ")
	hyp, matched := grammar.Evaluate(normStatus, nil)
	if !matched || hyp.Intent != "query_status" {
		t.Fatalf("expected query_status match, got matched=%v, intent=%s", matched, hyp.Intent)
	}

	// 2. Prefix match with slot
	normAlarm := normalizer.Normalize("Set alarm for 07:30 AM")
	hypAlarm, matchedAlarm := grammar.Evaluate(normAlarm, nil)
	if !matchedAlarm || hypAlarm.Intent != "set_alarm" {
		t.Fatalf("expected set_alarm match, got %v", matchedAlarm)
	}
	if len(hypAlarm.Slots) != 1 || hypAlarm.Slots[0].Name != "time" || hypAlarm.Slots[0].Value != "07:30 am" {
		t.Fatalf("unexpected slots: %+v", hypAlarm.Slots)
	}
}

func TestServiceDeterministicInterpretationAndWorkspacePublication(t *testing.T) {
	ws := &mockWorkspace{}
	cfg := understanding.WithConfigOptions()
	svc := understanding.NewService(cfg, nil)

	svcWithWS := understanding.NewService(cfg, ws)
	if err := svcWithWS.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer svcWithWS.Close()

	ctx := context.Background()
	env := communication.Envelope{
		ID:         "perception-101",
		PayloadRef: "weather in Seattle",
	}

	frame, err := svcWithWS.InterpretEnvelope(ctx, env)
	if err != nil {
		t.Fatalf("InterpretEnvelope failed: %v", err)
	}

	if frame.Status != understanding.StatusUnambiguous {
		t.Fatalf("expected UNAMBIGUOUS status, got %s", frame.Status)
	}
	if frame.PrimaryHypothesis.Intent != "query_weather" {
		t.Fatalf("expected intent query_weather, got %s", frame.PrimaryHypothesis.Intent)
	}
	if frame.PrimaryHypothesis.SourceLayer != understanding.LayerReflexiveGrammar {
		t.Fatalf("expected source layer LayerReflexiveGrammar, got %s", frame.PrimaryHypothesis.SourceLayer)
	}

	// Verify workspace publication
	ws.mu.Lock()
	pubCount := len(ws.published)
	ws.mu.Unlock()
	if pubCount != 1 {
		t.Fatalf("expected 1 published envelope to workspace, got %d", pubCount)
	}

	// Verify ParseIntent & ExecuteTask interface compatibility
	parsedStr, err := svc.ParseIntent(ctx, "cancel")
	if err != nil {
		t.Fatalf("ParseIntent failed: %v", err)
	}
	if parsedStr == "" {
		t.Fatalf("expected non-empty ParseIntent string")
	}

	status, resRef, err := svc.ExecuteTask(ctx, "wf-1", "n-1", executive.BudgetStandard, "status")
	if err != nil || status != executive.StatusConfident {
		t.Fatalf("ExecuteTask failed: status=%s, err=%v", status, err)
	}
	var decodedFrame understanding.SemanticFrame
	if err := json.Unmarshal([]byte(resRef), &decodedFrame); err != nil {
		t.Fatalf("unmarshal payload failed: %v", err)
	}
	if decodedFrame.PrimaryHypothesis.Intent != "query_status" {
		t.Fatalf("decoded frame intent mismatch")
	}
}

func TestServiceConcurrentRaceSafety(t *testing.T) {
	svc := understanding.NewService(understanding.WithConfigOptions(), nil)
	_ = svc.Start()
	defer svc.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			env := communication.Envelope{
				ID:         "concurrent-env",
				PayloadRef: "set alarm for 08:00",
			}
			_, _ = svc.InterpretEnvelope(ctx, env)
		}(i)
	}
	wg.Wait()
}
