package workspace_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/workspace"
)

func newTestEnvelope(topic communication.TopicID, source string) communication.Envelope {
	env, _ := communication.NewEnvelopeBuilder().
		WithSource(source).
		WithTopic(topic).
		WithPayloadRef("storage://test/payload-1").
		WithModality("structured").
		WithConfidence(0.85).
		WithUrgency(10).
		WithCostEstimate(50).
		Build()
	return env
}

func TestLifecycle(t *testing.T) {
	ws := workspace.NewEngine()
	if ws.Name() != "Intelligence.Workspace" {
		t.Fatalf("unexpected Name: %s", ws.Name())
	}
	if err := ws.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := ws.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	err := ws.Publish(context.Background(), newTestEnvelope(communication.TopicCandidatePlans, "Reasoning"))
	if !errors.Is(err, workspace.ErrWorkspaceClosed) {
		t.Fatalf("expected ErrWorkspaceClosed after close, got: %v", err)
	}
}

func TestPublishSubscribeLeveled(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	var received int32
	sub, err := ws.Subscribe(communication.TopicCandidatePlans, "PlanningAbility", func(ctx context.Context, env communication.Envelope) error {
		atomic.AddInt32(&received, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	if sub.Topic() != communication.TopicCandidatePlans {
		t.Fatalf("unexpected topic: %s", sub.Topic())
	}

	// Publish to matching topic
	ctx := context.Background()
	if err := ws.Publish(ctx, newTestEnvelope(communication.TopicCandidatePlans, "Reasoning")); err != nil {
		t.Fatalf("Publish matching failed: %v", err)
	}

	// Publish to non-matching topic
	if err := ws.Publish(ctx, newTestEnvelope(communication.TopicPerception, "Vision")); err != nil {
		t.Fatalf("Publish non-matching failed: %v", err)
	}

	if atomic.LoadInt32(&received) != 1 {
		t.Fatalf("expected exactly 1 delivery, got %d", received)
	}
}

func TestPublishGlobalBroadcast(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	var received int32
	_, err := ws.Subscribe(communication.TopicPerception, "VisionAbility", func(ctx context.Context, env communication.Envelope) error {
		atomic.AddInt32(&received, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish to different topic WITH global broadcast override
	ctx := context.Background()
	env := newTestEnvelope(communication.TopicValueFlags, "ValueAbility")
	if err := ws.Publish(ctx, env, workspace.WithGlobalBroadcast(true)); err != nil {
		t.Fatalf("Publish global failed: %v", err)
	}

	if atomic.LoadInt32(&received) != 1 {
		t.Fatalf("expected global broadcast delivery to subscriber, got %d", received)
	}
}

func TestSubscribeAllAndUnsubscribe(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	var count int32
	subs, err := ws.SubscribeAll("ExecutiveMonitor", func(ctx context.Context, env communication.Envelope) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("SubscribeAll failed: %v", err)
	}
	if len(subs) != 9 {
		t.Fatalf("expected 9 subscriptions from SubscribeAll, got %d", len(subs))
	}

	ctx := context.Background()
	_ = ws.Publish(ctx, newTestEnvelope(communication.TopicPerception, "s1"))
	_ = ws.Publish(ctx, newTestEnvelope(communication.TopicActiveGoals, "s2"))

	if atomic.LoadInt32(&count) != 2 {
		t.Fatalf("expected 2 deliveries, got %d", count)
	}

	// Unsubscribe first sub and test cancel
	_ = subs[0].Cancel()
}

func TestBufferLimitsAndLookup(t *testing.T) {
	ws := workspace.NewEngine(workspace.WithBufferLimit(5))
	_ = ws.Start()
	defer ws.Close()

	ctx := context.Background()
	var lastID string
	for i := 0; i < 10; i++ {
		env := newTestEnvelope(communication.TopicEvaluatedOptions, "Decision")
		lastID = env.ID
		_ = ws.Publish(ctx, env)
	}

	list := ws.ListTopicEnvelopes(communication.TopicEvaluatedOptions, 100)
	if len(list) != 5 {
		t.Fatalf("expected buffer limit 5 enforced, got %d", len(list))
	}

	lookup, ok := ws.GetEnvelope(lastID)
	if !ok || lookup.ID != lastID {
		t.Fatalf("GetEnvelope failed for ID %s", lastID)
	}
}

func TestConcurrentWorkspaceLoad(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	var delivered int64
	topics := communication.AllTopics()
	for _, topic := range topics {
		_, _ = ws.Subscribe(topic, "StressSubscriber", func(ctx context.Context, env communication.Envelope) error {
			atomic.AddInt64(&delivered, 1)
			time.Sleep(1 * time.Millisecond)
			return nil
		})
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			topic := topics[idx%len(topics)]
			env := newTestEnvelope(topic, "ConcurrentPublisher")
			_ = ws.Publish(ctx, env)
		}(i)
	}

	wg.Wait()
	if atomic.LoadInt64(&delivered) != 100 {
		t.Fatalf("expected 100 deliveries, got %d", delivered)
	}
}

func TestWorkspaceEpisodeGraphCoordination(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	ctx := context.Background()

	// Register hierarchy
	_ = ws.RegisterEpisodeChild(ctx, "ep-root", "ep-child-1")
	_ = ws.RegisterEpisodeChild(ctx, "ep-root", "ep-child-2")
	children, err := ws.GetEpisodeChildren(ctx, "ep-root")
	if err != nil || len(children) != 2 {
		t.Fatalf("expected 2 children, got %v (err=%v)", children, err)
	}

	// Register dependencies
	_ = ws.RegisterEpisodeDependencies(ctx, "ep-child-2", []string{"ep-child-1"})
	ready, err := ws.IsEpisodeReady(ctx, "ep-child-2")
	if err != nil || ready {
		t.Fatalf("expected ep-child-2 not ready, got %v (err=%v)", ready, err)
	}

	// Notify dependency complete
	_ = ws.NotifyDependencyComplete(ctx, "ep-child-1")
	ready, err = ws.IsEpisodeReady(ctx, "ep-child-2")
	if err != nil || !ready {
		t.Fatalf("expected ep-child-2 ready after dependency notification, got %v (err=%v)", ready, err)
	}
}
