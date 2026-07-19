package planning

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"idun/intelligence/communication"
	"idun/intelligence/decision"
)

// sharedCASStorer implements both planning.PayloadStorer and decision.PayloadStorer.
type sharedCASStorer struct {
	mu   sync.RWMutex
	data map[string][]byte
	seq  int
}

func newSharedCASStorer() *sharedCASStorer {
	return &sharedCASStorer{data: make(map[string][]byte)}
}

func (c *sharedCASStorer) Store(_ context.Context, data []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	key := fmt.Sprintf("cas://key-stage5a-%d", c.seq)
	c.data[key] = data
	return key, nil
}

func (c *sharedCASStorer) Retrieve(_ context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	d, ok := c.data[key]
	if !ok {
		return nil, fmt.Errorf("cas: key not found: %s", key)
	}
	return d, nil
}

// sharedWorkspaceBridge implements planning and decision publishers/subscribers and routes envelopes.
type sharedWorkspaceBridge struct {
	mu            sync.RWMutex
	envelopes     []communication.Envelope
	handlers      map[communication.TopicID]func(ctx context.Context, env communication.Envelope) error
	subscriptions []*dummySubscription
}

type dummySubscription struct {
	bridge *sharedWorkspaceBridge
	topic  communication.TopicID
}

func (d *dummySubscription) Cancel() error {
	d.bridge.mu.Lock()
	defer d.bridge.mu.Unlock()
	delete(d.bridge.handlers, d.topic)
	return nil
}

func newSharedWorkspaceBridge() *sharedWorkspaceBridge {
	return &sharedWorkspaceBridge{
		handlers: make(map[communication.TopicID]func(ctx context.Context, env communication.Envelope) error),
	}
}

func (b *sharedWorkspaceBridge) Publish(ctx context.Context, env communication.Envelope) error {
	b.mu.Lock()
	b.envelopes = append(b.envelopes, env)
	handler, ok := b.handlers[env.Topic]
	b.mu.Unlock()

	if ok && handler != nil {
		return handler(ctx, env)
	}
	return nil
}

func (b *sharedWorkspaceBridge) Subscribe(topic communication.TopicID, _ string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = handler
	sub := &dummySubscription{bridge: b, topic: topic}
	b.subscriptions = append(b.subscriptions, sub)
	return sub, nil
}

func (b *sharedWorkspaceBridge) subscribeDecision(topic communication.TopicID, _ string, handler func(ctx context.Context, env communication.Envelope) error) (decision.WorkspaceSubscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = handler
	sub := &dummySubscription{bridge: b, topic: topic}
	b.subscriptions = append(b.subscriptions, sub)
	return sub, nil
}

// decisionSubscriberBridge adapts sharedWorkspaceBridge for decision.WorkspaceSubscriber.
type decisionSubscriberBridge struct {
	bridge *sharedWorkspaceBridge
}

func (d *decisionSubscriberBridge) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (decision.WorkspaceSubscription, error) {
	return d.bridge.subscribeDecision(topic, subscriberID, handler)
}

// Test A: Multi-Plan Generation
// Verify that when HTN, GOAP, and TreeSearch each contribute subgoals, PlanningResult.Plans contains exactly 3 independent Plan structs.
func TestStage5A_MultiPlanGeneration(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("TACTICAL"))

	svc := NewService(WithSpecialistRegistry(reg))
	if err := svc.Start(); err != nil {
		t.Fatalf("failed to start planning service: %v", err)
	}
	defer svc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-multi-1").
		WithGoal("Accomplish mission X").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical failed: %v", err)
	}

	if len(res.Plans) != 3 {
		t.Fatalf("expected exactly 3 plans in res.Plans, got %d", len(res.Plans))
	}

	seenIDs := make(map[string]bool)
	seenFPs := make(map[string]bool)
	seenTypes := make(map[string]bool)

	for i, p := range res.Plans {
		if p.PlanID == "" {
			t.Errorf("plan[%d] has empty PlanID", i)
		}
		if seenIDs[p.PlanID] {
			t.Errorf("duplicate PlanID detected: %s", p.PlanID)
		}
		seenIDs[p.PlanID] = true

		if p.PlanFingerprint == "" {
			t.Errorf("plan[%d] has empty PlanFingerprint", i)
		}
		if seenFPs[p.PlanFingerprint] {
			t.Errorf("duplicate PlanFingerprint detected: %s", p.PlanFingerprint)
		}
		seenFPs[p.PlanFingerprint] = true

		if p.PlannerID == "" || p.PlannerType == "" {
			t.Errorf("plan[%d] missing PlannerID (%s) or PlannerType (%s)", i, p.PlannerID, p.PlannerType)
		}
		seenTypes[p.PlannerType] = true
	}

	if !seenTypes["HTN"] || !seenTypes["GOAP"] || !seenTypes["TreeSearch"] {
		t.Errorf("expected PlannerTypes HTN, GOAP, and TreeSearch, got seenTypes=%v", seenTypes)
	}
}

// Test B: Plan Isolation
// Ensure subgoals generated by HTN are not contained in GOAP or TreeSearch plans, and vice versa.
func TestStage5A_PlanIsolation(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("TACTICAL"))

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-iso-1").
		WithGoal("Test isolation").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical failed: %v", err)
	}

	for _, p := range res.Plans {
		for _, sg := range p.Subgoals {
			switch p.PlannerType {
			case "HTN":
				if !strings.HasPrefix(sg.SubgoalID, "htn-sg-") {
					t.Errorf("HTN plan contains foreign subgoal ID: %s", sg.SubgoalID)
				}
			case "GOAP":
				if !strings.HasPrefix(sg.SubgoalID, "goap-act-") {
					t.Errorf("GOAP plan contains foreign subgoal ID: %s", sg.SubgoalID)
				}
			case "TreeSearch":
				if !strings.HasPrefix(sg.SubgoalID, "tree-sg-") {
					t.Errorf("TreeSearch plan contains foreign subgoal ID: %s", sg.SubgoalID)
				}
			}
		}
	}
}

// Test C: No Applicable Specialist
// When no specialist generates subgoals, verify ResultNoPlans and exactly 1 fallback infeasible plan.
func TestStage5A_NoApplicableSpecialist(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(&mockSpecialist{name: "RoboSpec", domains: []string{"Robotics"}})

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-noapp-1").
		WithGoal("Test no applicable specialist").
		WithDomain("Finance").
		WithTargetDepth(DepthTactical).
		Build()

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical failed: %v", err)
	}

	if res.ResultStatus != ResultNoPlans {
		t.Errorf("expected ResultStatus=%v, got %v", ResultNoPlans, res.ResultStatus)
	}
	if res.Status != PlanStatusInfeasible {
		t.Errorf("expected Status=%v, got %v", PlanStatusInfeasible, res.Status)
	}
	if len(res.Plans) != 1 {
		t.Fatalf("expected exactly 1 fallback plan when no specialist contributes, got %d", len(res.Plans))
	}
	if len(res.Plans[0].Subgoals) != 0 {
		t.Errorf("fallback plan should have 0 subgoals, got %d", len(res.Plans[0].Subgoals))
	}
}

// Test D: Single Specialist
// When only 1 specialist matches domain, verify exactly 1 complete Plan is generated.
func TestStage5A_SingleSpecialist(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-single-1").
		WithGoal("Execute HTN only").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical failed: %v", err)
	}

	if len(res.Plans) != 1 {
		t.Fatalf("expected exactly 1 plan when only 1 specialist registered, got %d", len(res.Plans))
	}
	if res.Plans[0].PlannerType != "HTN" {
		t.Errorf("expected PlannerType=HTN, got %s", res.Plans[0].PlannerType)
	}
	if res.Plans[0].Status != PlanStatusComplete {
		t.Errorf("expected PlanStatusComplete, got %v", res.Plans[0].Status)
	}
}

// Test E: Fingerprint Deduplication across identical subgoals from different specialists
// If two specialists produce identical subgoals/edges, verify their PlanFingerprint values are distinct due to PlannerID/PlannerType.
func TestStage5A_FingerprintDeduplication(t *testing.T) {
	sharedSubgoals := []Subgoal{
		{SubgoalID: "sg-shared-1", Title: "Shared Subgoal 1", Description: "Shared Subgoal"},
		{SubgoalID: "sg-shared-2", Title: "Shared Subgoal 2", Description: "Shared Subgoal"},
	}
	sharedEdges := []DependencyEdge{
		{EdgeID: "e-shared-1", SourceNodeID: "sg-shared-1", TargetNodeID: "sg-shared-2", DependencyType: "HARD_PREREQUISITE", IsBlocking: true},
	}

	specAlpha := &mockSpecialist{
		name:    "SpecAlpha",
		domains: []string{"General"},
		fn: func(_ context.Context, _ *PlanningRequest) (*PlanningStepLog, []Subgoal, []DependencyEdge, error) {
			return &PlanningStepLog{SpecialistName: "SpecAlpha", Status: "DONE"}, sharedSubgoals, sharedEdges, nil
		},
	}
	specBeta := &mockSpecialist{
		name:    "SpecBeta",
		domains: []string{"General"},
		fn: func(_ context.Context, _ *PlanningRequest) (*PlanningStepLog, []Subgoal, []DependencyEdge, error) {
			return &PlanningStepLog{SpecialistName: "SpecBeta", Status: "DONE"}, sharedSubgoals, sharedEdges, nil
		},
	}

	reg := NewSpecialistRegistry()
	_ = reg.Register(specAlpha)
	_ = reg.Register(specBeta)

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-dedup-1").
		WithGoal("Verify fingerprint differentiation").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical failed: %v", err)
	}

	if len(res.Plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(res.Plans))
	}

	fp0 := res.Plans[0].PlanFingerprint
	fp1 := res.Plans[1].PlanFingerprint

	if fp0 == fp1 {
		t.Errorf("expected distinct fingerprints for identical subgoals produced by different planners (%s vs %s), both got %s", res.Plans[0].PlannerID, res.Plans[1].PlannerID, fp0)
	}
}

// Test F: Panic Isolation Preserves Other Plans
// If one specialist panics, verify the other specialists' candidate plans are still generated cleanly.
func TestStage5A_PanicIsolationPreservesOtherPlans(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(&mockSpecialist{name: "PanickerSpec", domains: []string{"General"}, panics: true})

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-panic-1").
		WithGoal("Test panic isolation multi-plan").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical failed: %v", err)
	}

	if len(res.Plans) != 1 {
		t.Fatalf("expected exactly 1 plan (from HTN, surviving the panic), got %d", len(res.Plans))
	}
	if res.Plans[0].PlannerType != "HTN" {
		t.Errorf("expected surviving plan to be HTN, got %s", res.Plans[0].PlannerType)
	}

	panicLogged := false
	for _, st := range res.Traces[0].PlanningSteps {
		if st.SpecialistName == "PanickerSpec" && st.Status == "ERROR_PANIC" {
			panicLogged = true
		}
	}
	if !panicLogged {
		t.Error("expected trace to record ERROR_PANIC in PlanningSteps for PanickerSpec")
	}
}

// Test G: Decision MCDA Comparison
// Verify that publishing multiple Plan objects via TopicCandidatePlans allows Decision to evaluate each plan and select via MCDA scoring.
func TestStage5A_DecisionMCDAComparison(t *testing.T) {
	storer := newSharedCASStorer()
	bridge := newSharedWorkspaceBridge()

	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("TACTICAL"))

	planSvc := NewService(
		WithSpecialistRegistry(reg),
		WithWorkspaceBridge(storer, bridge, bridge),
	)
	_ = planSvc.Start()
	defer planSvc.Close()

	decSub := &decisionSubscriberBridge{bridge: bridge}
	decSvc := decision.NewService(decision.WithWorkspaceBridge(storer, bridge, decSub))
	if err := decSvc.Start(); err != nil {
		t.Fatalf("failed to start decision service: %v", err)
	}
	defer decSvc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-mcda-1").
		WithGoal("Test Decision comparison across HTN, GOAP, and TreeSearch").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()

	res, err := planSvc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical failed: %v", err)
	}

	if len(res.Plans) != 3 {
		t.Fatalf("expected 3 plans produced by planning, got %d", len(res.Plans))
	}

	// Verify TopicEvaluatedOptions was emitted by Decision when evaluating candidate plans
	bridge.mu.RLock()
	var evalEnv *communication.Envelope
	for _, env := range bridge.envelopes {
		if env.Topic == communication.TopicEvaluatedOptions {
			evalEnv = &env
			break
		}
	}
	bridge.mu.RUnlock()

	if evalEnv == nil {
		t.Fatalf("expected TopicEvaluatedOptions envelope to be published by Decision after receiving TopicCandidatePlans")
	}

	// Build CandidateSet from the 3 independent candidate plans to verify rigorous MCDA evaluation
	cs := decision.CandidateSet{
		EpisodeID:  "ep-mcda-test",
		Candidates: make([]decision.Candidate, 0, len(res.Plans)),
	}
	for _, p := range res.Plans {
		cs.Candidates = append(cs.Candidates, decision.Candidate{
			ID:            p.PlanID,
			Description:   p.Goal,
			SourceAbility: "Planning",
			Attributes: map[string]float64{
				"confidence": p.ConfidenceProfile.OverallConfidence,
				"cost":       p.EstimatedCost,
			},
		})
	}

	rec, err := decSvc.EvaluateDeliberative(context.Background(), cs)
	if err != nil {
		t.Fatalf("EvaluateDeliberative failed across multi-plan candidates: %v", err)
	}

	if len(rec.TradeoffMatrix) != 3 {
		t.Errorf("expected TradeoffMatrix across 3 candidates (one per Plan), got TradeoffMatrix=%d", len(rec.TradeoffMatrix))
	}

	if rec.SelectedCandidateID == "" {
		t.Error("expected Decision to select a winning candidate ID")
	}

	// Verify the selected candidate matches one of the generated plans
	foundWinner := false
	for _, p := range res.Plans {
		if p.PlanID == rec.SelectedCandidateID {
			foundWinner = true
			break
		}
	}
	if !foundWinner {
		t.Errorf("selected candidate ID %s does not match any of the 3 candidate Plan IDs", rec.SelectedCandidateID)
	}
}

// Test H: Concurrency Safety
// Verify concurrent invocations of executePlanningEpisode across multiple goroutines produce isolated multi-plan outputs cleanly (-race).
func TestStage5A_ConcurrencySafety(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("TACTICAL"))

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	var wg sync.WaitGroup
	numWorkers := 10

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req, _ := NewPlanningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-conc-%d", id)).
				WithGoal(fmt.Sprintf("Concurrent goal %d", id)).
				WithDomain("General").
				WithTargetDepth(DepthTactical).
				Build()

			res, err := svc.PlanTactical(context.Background(), req)
			if err != nil {
				t.Errorf("worker %d PlanTactical failed: %v", id, err)
				return
			}
			if len(res.Plans) != 3 {
				t.Errorf("worker %d expected 3 plans, got %d", id, len(res.Plans))
			}
		}(i)
	}

	wg.Wait()
}
