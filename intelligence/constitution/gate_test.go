package constitution_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/workspace"
)

func newCandidateAction(urgency int, source string) communication.Envelope {
	env, _ := communication.NewEnvelopeBuilder().
		WithSource(source).
		WithTopic(communication.TopicActionExecution).
		WithPayloadRef("storage://actions/act-1").
		WithConfidence(0.90).
		WithUrgency(urgency).
		Build()
	return env
}

func TestLifecycle(t *testing.T) {
	gate := constitution.NewGate()
	if gate.Name() != "Intelligence.Constitution" {
		t.Fatalf("unexpected Name: %s", gate.Name())
	}
	if err := gate.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := gate.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	_, err := gate.EvaluateAction(context.Background(), newCandidateAction(10, "Decision"))
	if !errors.Is(err, constitution.ErrGateClosed) {
		t.Fatalf("expected ErrGateClosed after close, got: %v", err)
	}
}

func TestApprovedActionFlow(t *testing.T) {
	gate := constitution.NewGate()
	_ = gate.Start()
	defer gate.Close()

	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	var published int32
	_, _ = ws.Subscribe(communication.TopicActionExecution, "ActionExecutor", func(ctx context.Context, env communication.Envelope) error {
		atomic.AddInt32(&published, 1)
		return nil
	})

	env := newCandidateAction(10, "Decision")
	res, err := gate.InterceptAndPublish(context.Background(), env, ws)
	if err != nil {
		t.Fatalf("expected approved action to succeed, got: %v", err)
	}
	if res.Verdict != constitution.VerdictApproved {
		t.Fatalf("expected VerdictApproved, got: %v", res.Verdict)
	}
	if res.Signature == "" {
		t.Fatalf("expected cryptographic signature token on approved action")
	}
	if atomic.LoadInt32(&published) != 1 {
		t.Fatalf("expected action envelope published to workspace, got %d", published)
	}
}

func TestVetoActionFlow(t *testing.T) {
	gate := constitution.NewGate()
	_ = gate.Start()
	defer gate.Close()

	// Register veto rule blocking untrusted sources
	err := gate.RegisterRule(constitution.NewFunctionalRule(
		"test.veto.untrusted",
		"Vetoes untrusted sources",
		func(ctx context.Context, env communication.Envelope) (constitution.Verdict, string, error) {
			if env.Source == "UntrustedDriver" {
				return constitution.VerdictVetoed, "source is untrusted", nil
			}
			return constitution.VerdictApproved, "", nil
		},
	))
	if err != nil {
		t.Fatalf("RegisterRule failed: %v", err)
	}

	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	var alerts int32
	_, _ = ws.Subscribe(communication.TopicValueFlags, "ValueMonitor", func(ctx context.Context, env communication.Envelope) error {
		atomic.AddInt32(&alerts, 1)
		return nil
	})

	env := newCandidateAction(20, "UntrustedDriver")
	res, err := gate.InterceptAndPublish(context.Background(), env, ws)
	if !errors.Is(err, constitution.ErrActionVetoed) {
		t.Fatalf("expected ErrActionVetoed, got: %v", err)
	}
	if res.Verdict != constitution.VerdictVetoed {
		t.Fatalf("expected VerdictVetoed, got: %v", res.Verdict)
	}
	if atomic.LoadInt32(&alerts) != 1 {
		t.Fatalf("expected veto alert published to TopicValueFlags, got %d", alerts)
	}
}

func TestEscalationFlow(t *testing.T) {
	gate := constitution.NewGate()
	_ = gate.Start()
	defer gate.Close()

	_ = gate.RegisterRule(constitution.NewMaxUrgencyEscalationRule(90))

	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	var userAlerts int32
	_, _ = ws.Subscribe(communication.TopicUserIntent, "UserEscalator", func(ctx context.Context, env communication.Envelope) error {
		atomic.AddInt32(&userAlerts, 1)
		return nil
	})

	env := newCandidateAction(95, "EmergencyAction")
	res, err := gate.InterceptAndPublish(context.Background(), env, ws)
	if !errors.Is(err, constitution.ErrActionEscalation) {
		t.Fatalf("expected ErrActionEscalation, got: %v", err)
	}
	if res.Verdict != constitution.VerdictEscalateToUser {
		t.Fatalf("expected VerdictEscalateToUser, got: %v", res.Verdict)
	}
	if atomic.LoadInt32(&userAlerts) != 1 {
		t.Fatalf("expected user inquiry published to TopicUserIntent, got %d", userAlerts)
	}
}

func TestConcurrentGateLoad(t *testing.T) {
	gate := constitution.NewGate()
	_ = gate.Start()
	defer gate.Close()

	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			env := newCandidateAction(10, "ConcurrentDriver")
			_, _ = gate.EvaluateAction(ctx, env)
			_ = gate.ListRules()
		}(i)
	}

	wg.Wait()
}
