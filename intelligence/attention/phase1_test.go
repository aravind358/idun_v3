package attention_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"idun/intelligence/attention"
	"idun/intelligence/communication"
)

// mockPublisher is a thread-safe mock implementation of WorkspacePublisher and PayloadStorer.
type mockPublisher struct {
	mu           sync.Mutex
	published    []communication.Envelope
	stored       map[string][]byte
	storeCounter int64
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{
		published: make([]communication.Envelope, 0),
		stored:    make(map[string][]byte),
	}
}

func (m *mockPublisher) Publish(ctx context.Context, env communication.Envelope) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, env)
	return nil
}

func (m *mockPublisher) Store(ctx context.Context, data []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := atomic.AddInt64(&m.storeCounter, 1)
	ref := fmt.Sprintf("cas://payload-%d", id)
	m.stored[ref] = data
	return ref, nil
}

func (m *mockPublisher) getPublished() []communication.Envelope {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]communication.Envelope, len(m.published))
	copy(out, m.published)
	return out
}

func TestValidationFirewalls(t *testing.T) {
	// 1. Stimulus Validation
	if err := (attention.Stimulus{ID: "", SalienceScore: 50}).Validate(); err == nil {
		t.Fatal("expected error on empty Stimulus ID")
	}
	if err := (attention.Stimulus{ID: "s-bad", SalienceScore: 105}).Validate(); err == nil {
		t.Fatal("expected error on invalid SalienceScore > 100")
	}

	// 2. Policy Profile Validation
	if err := ((*attention.AttentionPolicyProfile)(nil)).Validate(); err != attention.ErrNilProfile {
		t.Fatalf("expected ErrNilProfile, got %v", err)
	}
	badProfile := attention.DefaultAttentionPolicyProfile()
	badProfile.Band1Threshold = 40
	badProfile.Band2Threshold = 60 // Non-monotonic
	if err := badProfile.Validate(); err == nil {
		t.Fatal("expected error on non-monotonic thresholds")
	}

	// 3. Capabilities Validation
	if err := ((*attention.AttentionCapabilities)(nil)).Validate(); err != attention.ErrNilCapabilities {
		t.Fatalf("expected ErrNilCapabilities, got %v", err)
	}
	badCaps := &attention.AttentionCapabilities{SupportsInterruptions: true}
	if err := badCaps.Validate(); err == nil {
		t.Fatal("expected error on empty capability fingerprint")
	}

	// 4. Trace & Metadata Validation
	if err := ((*attention.AttentionTrace)(nil)).Validate(); err != attention.ErrNilTrace {
		t.Fatalf("expected ErrNilTrace, got %v", err)
	}
	trace, err := attention.NewTraceBuilder().
		WithIdentifiers("trace-1", "stim-1", "source-1").
		WithDecision(attention.SalienceFocusImmediately, attention.PriorityBand1RealTime).
		WithOutcome(attention.ResultStatusFocused, attention.ReasonHighSalience).
		WithReplay(attention.AttentionReplayMetadata{
			PolicyFingerprint:     "fp-policy",
			CapabilityFingerprint: "fp-cap",
			AttentionVersion:      attention.AttentionVersion,
		}).Build()
	if err != nil {
		t.Fatalf("unexpected trace build error: %v", err)
	}
	if trace.TraceID != "trace-1" {
		t.Fatalf("expected TraceID trace-1, got %s", trace.TraceID)
	}

	// 5. Summary Validation
	if err := ((*attention.AttentionSummary)(nil)).Validate(); err != attention.ErrNilSummary {
		t.Fatalf("expected ErrNilSummary, got %v", err)
	}
	if err := (&attention.AttentionSummary{TotalStimuli: -1}).Validate(); err == nil {
		t.Fatal("expected error on negative TotalStimuli")
	}
}

func TestFluentBuildersAndSnapshots(t *testing.T) {
	// Builders
	p, err := attention.NewPolicyProfileBuilder().
		WithVersions("2.0.0", "2.0.0").
		WithThresholds(100, 80, 45, 15).
		WithMargins(5, 10).
		WithMaximumTracked(50).
		Build()
	if err != nil {
		t.Fatalf("policy profile build error: %v", err)
	}
	if p.PolicyFingerprint == "" {
		t.Fatal("expected non-empty PolicyFingerprint")
	}

	c, err := attention.NewCapabilitiesBuilder().
		WithInterruptions(true).
		WithBackgroundAttention(true).
		WithFocusSwitching(true).
		WithMultimodalAttention(true).
		WithDistributedAttention(true).
		WithFocusHistory(true).
		Build()
	if err != nil {
		t.Fatalf("capabilities build error: %v", err)
	}
	if !c.SupportsDistributedAttention {
		t.Fatal("expected SupportsDistributedAttention true")
	}

	// Atomic Snapshots
	pHolder := attention.NewPolicySnapshotHolder(p)
	loadedP := pHolder.Load()
	if loadedP.PolicyFingerprint != p.PolicyFingerprint {
		t.Fatalf("expected policy fingerprint %s, got %s", p.PolicyFingerprint, loadedP.PolicyFingerprint)
	}

	cHolder := attention.NewCapabilitiesSnapshotHolder(c)
	loadedC := cHolder.Load()
	if loadedC.CapabilityFingerprint != c.CapabilityFingerprint {
		t.Fatalf("expected capability fingerprint %s, got %s", c.CapabilityFingerprint, loadedC.CapabilityFingerprint)
	}
}

func TestPhase1EvaluateTraceAndLifecycle(t *testing.T) {
	svc := attention.NewService(attention.WithReplaySeed(12345))
	ctx := context.Background()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("expected clean Start, got %v", err)
	}
	defer svc.Close()

	stim, err := attention.NewStimulusBuilder().
		WithID("stim-trace-1").
		WithSource("CognitiveAbility.Perception").
		WithPayloadRef("cas://perception/101").
		WithSalienceScore(92).
		Build()
	if err != nil {
		t.Fatalf("stimulus build error: %v", err)
	}

	trace, err := svc.EvaluateTrace(ctx, stim)
	if err != nil {
		t.Fatalf("EvaluateTrace failed: %v", err)
	}

	if trace.Decision != attention.SalienceFocusImmediately || trace.PriorityBand != attention.PriorityBand1RealTime {
		t.Fatalf("expected FocusImmediately/Band1, got %s/%d", trace.Decision, trace.PriorityBand)
	}
	if trace.ResultStatus != attention.ResultStatusFocused {
		t.Fatalf("expected ResultStatus %s, got %s", attention.ResultStatusFocused, trace.ResultStatus)
	}
	if trace.TerminationReason != attention.ReasonHighSalience {
		t.Fatalf("expected Reason %s, got %s", attention.ReasonHighSalience, trace.TerminationReason)
	}
	if trace.ReplayMetadata.ReplaySeed != 12345 {
		t.Fatalf("expected ReplaySeed 12345, got %d", trace.ReplayMetadata.ReplaySeed)
	}
	if trace.AttentionVersion != attention.AttentionVersion {
		t.Fatalf("expected AttentionVersion %s, got %s", attention.AttentionVersion, trace.AttentionVersion)
	}

	sum := svc.GetSummary()
	if sum.TotalStimuli != 1 || sum.ImmediateFocusCount != 1 {
		t.Fatalf("expected summary counts Total=1, Immediate=1, got %+v", sum)
	}
}

func TestBoundedFocusHistoryAndRingBuffer(t *testing.T) {
	svc := attention.NewService()
	_ = svc.Start(context.Background())
	defer svc.Close()

	// Perform 50 consecutive focus switches
	for i := 1; i <= 50; i++ {
		stim := attention.Stimulus{
			ID:            fmt.Sprintf("stim-focus-%d", i),
			SalienceScore: 90,
		}
		_, _ = svc.EvaluateTrace(context.Background(), stim)
	}

	hist := svc.GetFocusHistory()
	if len(hist) != 16 {
		t.Fatalf("expected exactly 16 bounded ring buffer entries, got %d", len(hist))
	}
	// Verify the last entry matches stim-focus-50
	last := hist[len(hist)-1]
	if last.CurrentFocus != "stim-focus-50" {
		t.Fatalf("expected last focus stim-focus-50, got %s", last.CurrentFocus)
	}

	sum := svc.GetSummary()
	if sum.FocusSwitches != 50 {
		t.Fatalf("expected 50 focus switches in summary, got %d", sum.FocusSwitches)
	}
}

func TestWorkspaceEventPublishingAndReplayDeterminism(t *testing.T) {
	mockPub := newMockPublisher()
	svc := attention.NewService(
		attention.WithWorkspacePublisher(mockPub, mockPub),
		attention.WithReplaySeed(99999),
	)
	_ = svc.Start(context.Background())
	defer svc.Close()

	// 1. Focus shift stimulus
	stim1 := attention.Stimulus{ID: "stim-ws-1", Source: "test", SalienceScore: 88}
	trace1, err := svc.EvaluateTrace(context.Background(), stim1)
	if err != nil {
		t.Fatalf("trace1 error: %v", err)
	}

	// 2. Interrupt stimulus (higher priority / real-time arrival)
	stim2 := attention.Stimulus{ID: "stim-ws-2", Source: "test", SalienceScore: 95}
	trace2, err := svc.EvaluateTrace(context.Background(), stim2)
	if err != nil {
		t.Fatalf("trace2 error: %v", err)
	}

	// Verify exact determinism across both traces
	if trace1.PolicyFingerprint != trace2.PolicyFingerprint || trace1.PolicyFingerprint == "" {
		t.Fatalf("deterministic policy fingerprint mismatch or empty: %s vs %s", trace1.PolicyFingerprint, trace2.PolicyFingerprint)
	}

	// Verify published envelopes
	published := mockPub.getPublished()
	if len(published) < 2 {
		t.Fatalf("expected at least 2 published envelopes (focus change + interrupt), got %d", len(published))
	}

	// Verify envelopes have valid control-plane topics and payload refs
	for i, env := range published {
		if !env.Topic.IsValid() {
			t.Errorf("envelope %d has invalid topic %s", i, env.Topic)
		}
		if env.PayloadRef == "" {
			t.Errorf("envelope %d has empty PayloadRef", i)
		}
	}
}

func TestPhase1ConcurrentRaceSafety(t *testing.T) {
	mockPub := newMockPublisher()
	svc := attention.NewService(
		attention.WithWorkspacePublisher(mockPub, mockPub),
	)
	_ = svc.Start(context.Background())
	defer svc.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%4 == 0 {
				svc.SetActiveGoal(attention.ActiveGoalContext{
					ID:             fmt.Sprintf("goal-%d", idx),
					Summary:        "Concurrency test goal",
					PriorityWeight: idx * 10,
				})
			} else if idx%4 == 1 {
				_ = svc.GetActiveGoal()
				_ = svc.GetPolicyProfile()
				_ = svc.GetCapabilities()
			} else if idx%4 == 2 {
				_ = svc.GetSummary()
				_ = svc.GetFocusHistory()
				_ = svc.GetEventSummary()
			} else {
				score := (idx * 7) % 100
				stim := attention.Stimulus{
					ID:            fmt.Sprintf("stim-conc-%d", idx),
					Source:        "CognitiveAbility.Concurrency",
					SalienceScore: score,
				}
				_, _ = svc.EvaluateTrace(context.Background(), stim)
			}
		}(i)
	}
	wg.Wait()

	sum := svc.GetSummary()
	if sum.TotalStimuli == 0 {
		t.Fatalf("expected evaluations to occur during concurrent test")
	}
}
