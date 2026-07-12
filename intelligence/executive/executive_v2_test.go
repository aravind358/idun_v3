package executive_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

func newTestBid(topic communication.TopicID, source string, conf float64, cost int) communication.Envelope {
	env, _ := communication.NewEnvelopeBuilder().
		WithSource(source).
		WithTopic(topic).
		WithPayloadRef("storage://bids/" + source).
		WithModality("structured-bid").
		WithConfidence(conf).
		WithUrgency(0).
		WithCostEstimate(cost).
		Build()
	return env
}

func TestServiceV2InheritanceAndLifecycle(t *testing.T) {
	ws := workspace.NewEngine()
	cal := calibration.NewService()
	constGate := constitution.NewGate()

	execV2, err := executive.NewServiceV2(ws, cal, constGate, 1000)
	if err != nil {
		t.Fatalf("NewServiceV2 failed: %v", err)
	}

	if execV2.Name() != "Intelligence.Executive" {
		t.Fatalf("unexpected Name: %s", execV2.Name())
	}

	if err := execV2.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify inherited Version 1 AttentionGate capability works seamlessly
	_, _ = execV2.Evaluate(executive.Stimulus{})

	if err := execV2.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestArbitrateCompetitionWinner(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	cal := calibration.NewService()
	_ = cal.Start()
	defer cal.Close()

	constGate := constitution.NewGate()
	_ = constGate.Start()
	defer constGate.Close()

	execV2, _ := executive.NewServiceV2(ws, cal, constGate, 1000)
	_ = execV2.Start()
	defer execV2.Close()

	// Discount overconfident module source
	_ = cal.RecordAudit(calibration.AuditRecord{
		Source:             "OverconfidentPlanner",
		Topic:              communication.TopicCandidatePlans,
		ReportedConfidence: 0.95,
		ActualAccuracy:     0.19, // ~0.20 calibration ratio
	})

	ctx := context.Background()
	var publishedCount int32
	var winningSource string

	_, _ = ws.Subscribe(communication.TopicCandidatePlans, "PlanSubscriber", func(ctx context.Context, env communication.Envelope) error {
		atomic.AddInt32(&publishedCount, 1)
		winningSource = env.Source
		return nil
	})

	// Submit two competing bids
	bid1 := newTestBid(communication.TopicCandidatePlans, "OverconfidentPlanner", 0.95, 50)
	bid2 := newTestBid(communication.TopicCandidatePlans, "CalibratedPlanner", 0.85, 50)

	_ = execV2.SubmitBid(ctx, bid1, executive.HorizonDeliberative)
	_ = execV2.SubmitBid(ctx, bid2, executive.HorizonDeliberative)

	dec, err := execV2.ArbitrateCompetition(ctx, communication.TopicCandidatePlans, 0.50)
	if err != nil {
		t.Fatalf("ArbitrateCompetition failed: %v", err)
	}
	if !dec.Admitted {
		t.Fatalf("expected winning bid admitted, got Reason: %s", dec.Reason)
	}
	if dec.Winner.Source != "CalibratedPlanner" {
		t.Fatalf("expected CalibratedPlanner to win over discounted OverconfidentPlanner, got %s", dec.Winner.Source)
	}
	if atomic.LoadInt32(&publishedCount) != 1 || winningSource != "CalibratedPlanner" {
		t.Fatalf("expected winning bid published to workspace")
	}
}

func TestArbitrateCompetitionImpasse(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	cal := calibration.NewService()
	constGate := constitution.NewGate()
	execV2, _ := executive.NewServiceV2(ws, cal, constGate, 1000)
	_ = execV2.Start()
	defer execV2.Close()

	var impasses int32
	_, _ = ws.Subscribe(communication.TopicImpasses, "ImpasseMonitor", func(ctx context.Context, env communication.Envelope) error {
		atomic.AddInt32(&impasses, 1)
		return nil
	})

	ctx := context.Background()
	dec, err := execV2.ArbitrateCompetition(ctx, communication.TopicEvaluatedOptions, 0.50)
	if err != nil {
		t.Fatalf("ArbitrateCompetition failed: %v", err)
	}
	if dec.Admitted || !dec.ImpasseEmitted {
		t.Fatalf("expected impasse emitted for empty topic queue")
	}
	if atomic.LoadInt32(&impasses) != 1 {
		t.Fatalf("expected impasse envelope published to TopicImpasses, got %d", impasses)
	}
}

func TestArbitrateActionConstitutionalGate(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	cal := calibration.NewService()
	constGate := constitution.NewGate()
	_ = constGate.Start()
	defer constGate.Close()

	// Add a rule that vetoes actions from untrusted driver
	_ = constGate.RegisterRule(constitution.NewFunctionalRule(
		"test.veto",
		"vetoes untrusted driver",
		func(ctx context.Context, env communication.Envelope) (constitution.Verdict, string, error) {
			if env.Source == "UntrustedDriver" {
				return constitution.VerdictVetoed, "untrusted", nil
			}
			return constitution.VerdictApproved, "", nil
		},
	))

	execV2, _ := executive.NewServiceV2(ws, cal, constGate, 1000)
	_ = execV2.Start()
	defer execV2.Close()

	ctx := context.Background()
	_ = execV2.SubmitBid(ctx, newTestBid(communication.TopicActionExecution, "UntrustedDriver", 0.90, 50), executive.HorizonDeliberative)

	_, err := execV2.ArbitrateCompetition(ctx, communication.TopicActionExecution, 0.50)
	if !errors.Is(err, constitution.ErrActionVetoed) {
		t.Fatalf("expected ErrActionVetoed from Constitutional Gate interception, got: %v", err)
	}
}

func TestConcurrentExecutiveV2Load(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	cal := calibration.NewService()
	constGate := constitution.NewGate()
	execV2, _ := executive.NewServiceV2(ws, cal, constGate, 100000)
	_ = execV2.Start()
	defer execV2.Close()

	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			topic := communication.TopicCandidatePlans
			bid := newTestBid(topic, "ConcurrentDriver", 0.85, 10)
			_ = execV2.SubmitBid(ctx, bid, executive.HorizonDeliberative)
			time.Sleep(1 * time.Millisecond)
			_, _ = execV2.ArbitrateCompetition(ctx, topic, 0.50)
		}(i)
	}

	wg.Wait()
}
