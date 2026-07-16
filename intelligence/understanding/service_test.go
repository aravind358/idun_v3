package understanding_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
	"idun/intelligence/understanding"
	"idun/intelligence/workspace"
)

type mockSubscription struct {
	id      workspace.SubscriptionID
	topic   communication.TopicID
	handler workspace.EnvelopeHandler
}

func (s *mockSubscription) ID() workspace.SubscriptionID       { return s.id }
func (s *mockSubscription) Topic() communication.TopicID       { return s.topic }
func (s *mockSubscription) Subscriber() string                 { return "sub" }
func (s *mockSubscription) Handler() workspace.EnvelopeHandler { return s.handler }
func (s *mockSubscription) Cancel() error                      { return nil }

type mockWorkspace struct {
	mu            sync.Mutex
	published     []communication.Envelope
	subscriptions map[communication.TopicID]workspace.EnvelopeHandler
}

func (m *mockWorkspace) Publish(ctx context.Context, env communication.Envelope, opts ...workspace.PublishOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, env)
	return nil
}

func (m *mockWorkspace) Subscribe(topic communication.TopicID, subscriberID string, handler workspace.EnvelopeHandler) (workspace.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.subscriptions == nil {
		m.subscriptions = make(map[communication.TopicID]workspace.EnvelopeHandler)
	}
	m.subscriptions[topic] = handler
	return &mockSubscription{id: workspace.SubscriptionID("sub-1"), topic: topic, handler: handler}, nil
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

type mockPayloadStorer struct {
	mu     sync.Mutex
	stored map[string][]byte
}

func newMockPayloadStorer() *mockPayloadStorer {
	return &mockPayloadStorer{stored: make(map[string][]byte)}
}

func (m *mockPayloadStorer) Store(_ context.Context, data []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("cas-key-%d", len(m.stored)+1)
	m.stored[key] = data
	return key, nil
}

func (m *mockPayloadStorer) Retrieve(_ context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.stored[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return data, nil
}

func TestService_PerceptionBridge_Integration(t *testing.T) {
	ws := &mockWorkspace{}
	storer := newMockPayloadStorer()
	cfg := understanding.WithConfigOptions()

	svc := understanding.NewService(
		cfg,
		ws,
		understanding.WithPayloadStorer(storer),
	)

	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer svc.Close()

	ws.mu.Lock()
	handler, ok := ws.subscriptions[communication.TopicPerception]
	ws.mu.Unlock()
	if !ok || handler == nil {
		t.Fatalf("expected Understanding service to subscribe to TopicPerception upon Start")
	}

	ctx := context.Background()
	rawText := "weather in Seattle"
	casKey, err := storer.Store(ctx, []byte(rawText))
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	perceptionEnv := communication.Envelope{
		ID:         "world-interaction-123",
		Source:     "World.Service",
		Topic:      communication.TopicPerception,
		PayloadRef: casKey,
	}

	if err := handler(ctx, perceptionEnv); err != nil {
		t.Fatalf("handlePerceptionEnvelope failed: %v", err)
	}

	ws.mu.Lock()
	pubCount := len(ws.published)
	pubList := append([]communication.Envelope(nil), ws.published...)
	ws.mu.Unlock()

	if pubCount != 1 {
		t.Fatalf("expected 1 published envelope to TopicUserIntent, got %d", pubCount)
	}
	pubEnv := pubList[0]
	if pubEnv.Topic != communication.TopicUserIntent {
		t.Fatalf("expected topic %q, got %q", communication.TopicUserIntent, pubEnv.Topic)
	}
	if pubEnv.ParentRef != "world-interaction-123" {
		t.Fatalf("expected ParentRef 'world-interaction-123', got %q", pubEnv.ParentRef)
	}

	// Verify structured output frame stored in CAS via PayloadStorer
	frameData, err := storer.Retrieve(ctx, pubEnv.PayloadRef)
	if err != nil {
		t.Fatalf("expected published PayloadRef %q to be in CAS: %v", pubEnv.PayloadRef, err)
	}

	var frame understanding.SemanticFrame
	if err := json.Unmarshal(frameData, &frame); err != nil {
		t.Fatalf("unmarshal stored SemanticFrame failed: %v", err)
	}
	if frame.PrimaryHypothesis.Intent != "query_weather" {
		t.Fatalf("expected intent query_weather, got %s", frame.PrimaryHypothesis.Intent)
	}
}
