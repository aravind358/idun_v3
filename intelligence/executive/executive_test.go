package executive_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"idun/intelligence/executive"
)

// mockAbilityDriver is a thread-safe test double implementing executive.AbilityDriver.
type mockAbilityDriver struct {
	ability     executive.CognitiveAbility
	mu          sync.Mutex
	status      executive.EpistemicStatus
	outputRef   string
	err         error
	callCount   int32
	delay       time.Duration
	customHandler func(ctx context.Context, workflowID, nodeID string, budget executive.BudgetTier, payloadRef string) (executive.EpistemicStatus, string, error)
}

func newMockDriver(ability executive.CognitiveAbility, status executive.EpistemicStatus, outputRef string, err error) *mockAbilityDriver {
	return &mockAbilityDriver{
		ability:   ability,
		status:    status,
		outputRef: outputRef,
		err:       err,
	}
}

func (m *mockAbilityDriver) Ability() executive.CognitiveAbility {
	return m.ability
}

func (m *mockAbilityDriver) ExecuteTask(ctx context.Context, workflowID, nodeID string, budget executive.BudgetTier, payloadRef string) (executive.EpistemicStatus, string, error) {
	atomic.AddInt32(&m.callCount, 1)
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return executive.StatusUnresolvableContradiction, "", ctx.Err()
		case <-time.After(m.delay):
		}
	}
	m.mu.Lock()
	handler := m.customHandler
	status := m.status
	out := m.outputRef
	err := m.err
	m.mu.Unlock()

	if handler != nil {
		return handler(ctx, workflowID, nodeID, budget, payloadRef)
	}
	return status, out, err
}

func (m *mockAbilityDriver) getCallCount() int {
	return int(atomic.LoadInt32(&m.callCount))
}

// TestKernelLifecycle tests Name(), Start(), Close(), and re-start protection.
func TestKernelLifecycle(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})
	if exec.Name() != "Intelligence.Executive" {
		t.Fatalf("unexpected component name: %s", exec.Name())
	}

	if err := exec.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	if err := exec.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	if err := exec.Start(); !errors.Is(err, executive.ErrServiceClosed) {
		t.Fatalf("expected ErrServiceClosed after close, got: %v", err)
	}
}

// TestAttentionGate tests Evaluate and ActiveGoalContext references.
func TestAttentionGate(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})

	goal := executive.ActiveGoalContext{
		ID:             "goal-001",
		Summary:        "Build Arvind's restaurant",
		PriorityWeight: 100,
	}
	exec.SetActiveGoal(goal)

	gotGoal := exec.GetActiveGoal()
	if gotGoal.ID != goal.ID {
		t.Fatalf("expected goal %s, got %s", goal.ID, gotGoal.ID)
	}

	tests := []struct {
		name         string
		stimulus     executive.Stimulus
		wantDecision executive.SalienceDecision
		wantBand     executive.PriorityBand
	}{
		{
			name:         "safety tripwire overrides score",
			stimulus:     executive.Stimulus{ID: "s1", SafetyFlag: true, SalienceScore: 10},
			wantDecision: executive.SalienceFocusImmediately,
			wantBand:     executive.PriorityBand0CriticalSafety,
		},
		{
			name:         "high salience real-time focus",
			stimulus:     executive.Stimulus{ID: "s2", SalienceScore: 90},
			wantDecision: executive.SalienceFocusImmediately,
			wantBand:     executive.PriorityBand1RealTime,
		},
		{
			name:         "interactive dialogue focus",
			stimulus:     executive.Stimulus{ID: "s3", SalienceScore: 60},
			wantDecision: executive.SalienceFocusImmediately,
			wantBand:     executive.PriorityBand2Interactive,
		},
		{
			name:         "background schedule",
			stimulus:     executive.Stimulus{ID: "s4", SalienceScore: 30},
			wantDecision: executive.SalienceSchedule,
			wantBand:     executive.PriorityBand3Background,
		},
		{
			name:         "low salience filter",
			stimulus:     executive.Stimulus{ID: "s5", SalienceScore: 10},
			wantDecision: executive.SalienceFilter,
			wantBand:     executive.PriorityBand4Idle,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dec, band := exec.Evaluate(tc.stimulus)
			if dec != tc.wantDecision || band != tc.wantBand {
				t.Fatalf("Evaluate() = (%v, %v), want (%v, %v)", dec, band, tc.wantDecision, tc.wantBand)
			}
		})
	}
}

// TestPriorityQueue verifies priority band ordering (Bands 0..4).
func TestPriorityQueue(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})

	wgLow := &executive.WorkflowGraph{ID: "low", Priority: executive.PriorityBand3Background}
	wgHigh := &executive.WorkflowGraph{ID: "high", Priority: executive.PriorityBand0CriticalSafety}
	wgMed := &executive.WorkflowGraph{ID: "med", Priority: executive.PriorityBand2Interactive}

	_ = exec.Enqueue(wgLow)
	_ = exec.Enqueue(wgHigh)
	_ = exec.Enqueue(wgMed)

	first, ok := exec.Dequeue()
	if !ok || first.ID != "high" {
		t.Fatalf("expected high priority first, got: %v", first)
	}

	second, _ := exec.Dequeue()
	if second.ID != "med" {
		t.Fatalf("expected med priority second, got: %v", second)
	}

	third, _ := exec.Dequeue()
	if third.ID != "low" {
		t.Fatalf("expected low priority third, got: %v", third)
	}

	_, emptyOk := exec.Dequeue()
	if emptyOk {
		t.Fatalf("expected queue empty")
	}
}

// TestBudgetEscalation verifies AssignBudget and explicit budget upgrade arbitration.
func TestBudgetEscalation(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})

	if b := exec.AssignBudget(executive.SalienceFocusImmediately, executive.PriorityBand0CriticalSafety); b != executive.BudgetReflexive {
		t.Fatalf("expected REFLEXIVE for Band 0, got %s", b)
	}

	if b := exec.AssignBudget(executive.SalienceSchedule, executive.PriorityBand4Idle); b != executive.BudgetDeliberative {
		t.Fatalf("expected DELIBERATIVE for Band 4, got %s", b)
	}

	upgraded, ok := exec.EvaluateEscalation(executive.BudgetReflexive, executive.PriorityBand2Interactive)
	if !ok || upgraded != executive.BudgetStandard {
		t.Fatalf("expected upgrade to STANDARD, got %s (%v)", upgraded, ok)
	}

	_, maxed := exec.EvaluateEscalation(executive.BudgetDeliberative, executive.PriorityBand2Interactive)
	if maxed {
		t.Fatalf("expected no upgrade from DELIBERATIVE")
	}
}

// TestWorkflowExecution verifies normal execution across sequential nodes.
func TestWorkflowExecution(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})

	driverU := newMockDriver(executive.AbilityUnderstanding, executive.StatusConfident, "ref/parsed", nil)
	driverR := newMockDriver(executive.AbilityReasoning, executive.StatusConfident, "ref/result", nil)

	_ = exec.RegisterDriver(driverU)
	_ = exec.RegisterDriver(driverR)

	wg := &executive.WorkflowGraph{
		ID:          "wf-001",
		StartNodeID: "node1",
		Nodes: map[string]*executive.WorkflowNode{
			"node1": {ID: "node1", Ability: executive.AbilityUnderstanding, InputRef: "raw/input", NextNodeID: "node2"},
			"node2": {ID: "node2", Ability: executive.AbilityReasoning, NextNodeID: ""},
		},
		MaxFuel: 10,
	}

	res := exec.Execute(context.Background(), wg)
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if res.FinalStatus != executive.StatusConfident || res.OutputRef != "ref/result" {
		t.Fatalf("unexpected execution result: %+v", res)
	}
	if driverU.getCallCount() != 1 || driverR.getCallCount() != 1 {
		t.Fatalf("expected 1 call per driver, got U=%d, R=%d", driverU.getCallCount(), driverR.getCallCount())
	}
}

// TestBudgetEscalationWorkflow verifies explicit escalation retry when STATUS_ESCALATION_REQUIRED is returned.
func TestBudgetEscalationWorkflow(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})

	driver := newMockDriver(executive.AbilityReasoning, executive.StatusConfident, "ref/solved", nil)
	calls := 0
	driver.customHandler = func(ctx context.Context, workflowID, nodeID string, budget executive.BudgetTier, payloadRef string) (executive.EpistemicStatus, string, error) {
		calls++
		if budget == executive.BudgetReflexive {
			return executive.StatusEscalationRequired, "", nil
		}
		return executive.StatusConfident, "ref/solved_standard", nil
	}
	_ = exec.RegisterDriver(driver)

	wg := &executive.WorkflowGraph{
		ID:          "wf-escalate",
		StartNodeID: "n1",
		Nodes: map[string]*executive.WorkflowNode{
			"n1": {ID: "n1", Ability: executive.AbilityReasoning},
		},
		Budget:  executive.BudgetReflexive,
		MaxFuel: 5,
	}

	res := exec.Execute(context.Background(), wg)
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if res.OutputRef != "ref/solved_standard" || calls != 2 {
		t.Fatalf("expected 2 calls (reflexive then standard), got calls=%d out=%s", calls, res.OutputRef)
	}
}

// TestReflectionTrigger verifies epistemic contradiction routes to Reflection until MaxReflection depth exceeded.
func TestReflectionTrigger(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{DefaultMaxReflection: 2})

	driverR := newMockDriver(executive.AbilityReasoning, executive.StatusUnsureConflicting, "ref/conflict", nil)
	driverReflect := newMockDriver(executive.AbilityReflection, executive.StatusConfident, "ref/critique", nil)

	_ = exec.RegisterDriver(driverR)
	_ = exec.RegisterDriver(driverReflect)

	wg := &executive.WorkflowGraph{
		ID:            "wf-reflect",
		StartNodeID:   "n1",
		MaxReflection: 2,
		MaxFuel:       10,
		Nodes: map[string]*executive.WorkflowNode{
			"n1": {ID: "n1", Ability: executive.AbilityReasoning},
		},
	}

	res := exec.Execute(context.Background(), wg)
	if !errors.Is(res.Error, executive.ErrMaxReflectionExceeded) {
		t.Fatalf("expected ErrMaxReflectionExceeded, got: %v", res.Error)
	}
	if driverReflect.getCallCount() != 2 {
		t.Fatalf("expected 2 reflection calls before exceeding limit, got %d", driverReflect.getCallCount())
	}
}

// TestConstitutionalVeto verifies safety check veto from CognitiveAbility.Value.
func TestConstitutionalVeto(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})

	driverD := newMockDriver(executive.AbilityDecision, executive.StatusUnsureConstitutionalRisk, "ref/risky", nil)
	driverV := newMockDriver(executive.AbilityValue, executive.StatusUnsureConstitutionalRisk, "", nil)

	_ = exec.RegisterDriver(driverD)
	_ = exec.RegisterDriver(driverV)

	wg := &executive.WorkflowGraph{
		ID:          "wf-veto",
		StartNodeID: "n1",
		Nodes: map[string]*executive.WorkflowNode{
			"n1": {ID: "n1", Ability: executive.AbilityDecision},
		},
	}

	res := exec.Execute(context.Background(), wg)
	if !errors.Is(res.Error, executive.ErrConstitutionalVeto) {
		t.Fatalf("expected ErrConstitutionalVeto, got: %v", res.Error)
	}
}

// TestCancellation tests workflow cancellation via CancelTask and context cancellation.
func TestCancellation(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})

	driver := newMockDriver(executive.AbilityReasoning, executive.StatusConfident, "out", nil)
	driver.delay = 100 * time.Millisecond
	_ = exec.RegisterDriver(driver)

	wg := &executive.WorkflowGraph{
		ID:          "wf-cancel",
		StartNodeID: "n1",
		Nodes: map[string]*executive.WorkflowNode{
			"n1": {ID: "n1", Ability: executive.AbilityReasoning},
		},
	}

	done := make(chan executive.ExecutionResult, 1)
	go func() {
		done <- exec.Execute(context.Background(), wg)
	}()

	time.Sleep(20 * time.Millisecond)
	err := exec.CancelTask("wf-cancel")
	if err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}

	res := <-done
	if !errors.Is(res.Error, executive.ErrWorkflowCancelled) {
		t.Fatalf("expected ErrWorkflowCancelled, got: %v", res.Error)
	}
}

// TestHomeostasis verifies idle duration and ShouldConsolidate calculation.
func TestHomeostasis(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{IdleThreshold: 50 * time.Millisecond})

	exec.RecordActivity()
	if exec.ShouldConsolidate() {
		t.Fatalf("should not consolidate immediately after activity")
	}

	time.Sleep(60 * time.Millisecond)
	if !exec.ShouldConsolidate() {
		t.Fatalf("expected ShouldConsolidate true after idle threshold")
	}
}

// TestConcurrency verifies race-safety under concurrent evaluation, enqueue, and execution.
func TestConcurrency(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})
	driver := newMockDriver(executive.AbilityUnderstanding, executive.StatusConfident, "out", nil)
	_ = exec.RegisterDriver(driver)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s := executive.Stimulus{ID: "s", SalienceScore: 60}
			exec.Evaluate(s)

			exec.SetActiveGoal(executive.ActiveGoalContext{ID: "g", Summary: "test"})

			wf := &executive.WorkflowGraph{
				ID:          "wf",
				StartNodeID: "n1",
				Nodes: map[string]*executive.WorkflowNode{
					"n1": {ID: "n1", Ability: executive.AbilityUnderstanding},
				},
			}
			_ = exec.Enqueue(wf)
			exec.Execute(context.Background(), wf)
		}(i)
	}
	wg.Wait()
}

// TestEnumsAndErrorPaths tests String() representations and defensive error paths.
func TestEnumsAndErrorPaths(t *testing.T) {
	statuses := []executive.EpistemicStatus{
		executive.StatusConfident,
		executive.StatusUnsureAmbiguous,
		executive.StatusUnsureConflicting,
		executive.StatusUnsureConstitutionalRisk,
		executive.StatusInsufficientData,
		executive.StatusEscalationRequired,
		executive.StatusUnresolvableContradiction,
		executive.EpistemicStatus(999),
	}
	for _, s := range statuses {
		if s.String() == "" {
			t.Fatalf("empty string representation for status %d", s)
		}
	}

	budgets := []executive.BudgetTier{
		executive.BudgetReflexive,
		executive.BudgetStandard,
		executive.BudgetDeliberative,
		executive.BudgetTier(999),
	}
	for _, b := range budgets {
		if b.String() == "" {
			t.Fatalf("empty string representation for budget %d", b)
		}
	}

	exec := executive.NewExecutiveService(executive.Config{})
	if err := exec.RegisterDriver(nil); err == nil {
		t.Fatalf("expected error registering nil driver")
	}

	if _, err := exec.GetDriver("NonExistentAbility"); !errors.Is(err, executive.ErrAbilityNotRegistered) {
		t.Fatalf("expected ErrAbilityNotRegistered, got: %v", err)
	}

	resNil := exec.Execute(context.Background(), nil)
	if resNil.Error == nil {
		t.Fatalf("expected error executing nil workflow")
	}

	wgMissingNode := &executive.WorkflowGraph{
		ID:          "missing-node",
		StartNodeID: "nonexistent",
		Nodes:       map[string]*executive.WorkflowNode{},
	}
	resMissing := exec.Execute(context.Background(), wgMissingNode)
	if resMissing.Error == nil {
		t.Fatalf("expected error executing workflow with missing start node")
	}
}

// TestExecutiveControlMethods tests ListAbilities, CancelAll, Preempt, and IdleDuration.
func TestExecutiveControlMethods(t *testing.T) {
	exec := executive.NewExecutiveService(executive.Config{})
	driver := newMockDriver(executive.AbilityUnderstanding, executive.StatusConfident, "out", nil)
	_ = exec.RegisterDriver(driver)

	abilities := exec.ListAbilities()
	if len(abilities) != 1 || abilities[0] != executive.AbilityUnderstanding {
		t.Fatalf("unexpected list abilities: %v", abilities)
	}

	dur := exec.IdleDuration()
	if dur < 0 {
		t.Fatalf("invalid idle duration: %v", dur)
	}

	ctx, cancel := context.WithCancel(context.Background())
	exec.RegisterTask("t1", cancel)
	if err := exec.Preempt(ctx, executive.PriorityBand0CriticalSafety); err != nil {
		t.Fatalf("Preempt failed: %v", err)
	}
	if ctx.Err() == nil {
		t.Fatalf("expected task cancelled after preempt")
	}

	ctx2, cancel2 := context.WithCancel(context.Background())
	exec.RegisterTask("t2", cancel2)
	exec.CancelAll()
	if ctx2.Err() == nil {
		t.Fatalf("expected task cancelled after CancelAll")
	}
}

