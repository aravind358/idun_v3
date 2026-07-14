package planning

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockSpecialist struct {
	name     string
	domains  []string
	fn       func(ctx context.Context, req *PlanningRequest) (*PlanningStepLog, []Subgoal, []DependencyEdge, error)
	panics   bool
	duration time.Duration
}

func (m *mockSpecialist) Name() string             { return m.name }
func (m *mockSpecialist) SupportedDomains() []string { return m.domains }
func (m *mockSpecialist) Contribute(
	ctx context.Context,
	req *PlanningRequest,
	graph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) (*PlanningStepLog, []Subgoal, []DependencyEdge, error) {
	if m.panics {
		panic("intentional specialist panic for isolation test")
	}
	if m.duration > 0 {
		select {
		case <-time.After(m.duration):
		case <-ctx.Done():
			return nil, nil, nil, ctx.Err()
		}
	}
	if m.fn != nil {
		return m.fn(ctx, req)
	}
	return &PlanningStepLog{SpecialistName: m.name, Status: "DONE"}, []Subgoal{
		{SubgoalID: "sg-" + m.name, Title: "Task from " + m.name},
	}, nil, nil
}

func TestSpecialistRegistry_RegistrationAndOrdering(t *testing.T) {
	reg := NewSpecialistRegistry()

	s1 := &mockSpecialist{name: "GoalDecomposition", domains: []string{"General"}}
	s2 := &mockSpecialist{name: "TaskSequencing", domains: []string{"General"}}
	s3 := &mockSpecialist{name: "RoboticsSpec", domains: []string{"Robotics"}}

	if err := reg.Register(s1); err != nil {
		t.Fatalf("failed to register s1: %v", err)
	}
	if err := reg.Register(s2); err != nil {
		t.Fatalf("failed to register s2: %v", err)
	}
	if err := reg.Register(s3); err != nil {
		t.Fatalf("failed to register s3: %v", err)
	}

	// Duplicate registration check
	if err := reg.Register(s1); err == nil {
		t.Error("expected error on duplicate specialist registration, got nil")
	}
	if err := reg.Register(nil); err == nil {
		t.Error("expected error when registering nil specialist")
	}

	profile := DefaultPlanningPolicyProfile()
	profile.SpecialistWeights["GoalDecomposition"] = 0.5
	profile.SpecialistWeights["TaskSequencing"] = 0.9

	specs := reg.GetSpecialistsForDomain("General", profile)
	if len(specs) != 2 {
		t.Fatalf("expected 2 specialists for General domain, got %d", len(specs))
	}
	// TaskSequencing (weight 0.9) should be ordered before GoalDecomposition (weight 0.5)
	if specs[0].Name() != "TaskSequencing" || specs[1].Name() != "GoalDecomposition" {
		t.Errorf("unexpected specialist order: %s, %s", specs[0].Name(), specs[1].Name())
	}
}

func TestSpecialistRegistry_ConcurrentExecutionAndPanicIsolation(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(&mockSpecialist{name: "NormalSpec", domains: []string{"General"}})
	_ = reg.Register(&mockSpecialist{name: "PanicSpec", domains: []string{"General"}, panics: true})

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-panic-test").
		WithGoal("Test panic recovery").
		Build()

	cache := NewReflexivePlanningCache("ep-panic", "1.0")
	defer cache.Close()

	steps, subgoals, _, err := reg.ExecuteSpecialists(context.Background(), req, &DependencyGraphSnapshot{}, DefaultPlanningPolicyProfile(), cache)

	// Verify that panic did NOT crash the process and we got outputs from NormalSpec plus panic error/log
	if err == nil {
		t.Fatal("expected error returned from panicking specialist, got nil")
	}
	if len(subgoals) != 1 || subgoals[0].SubgoalID != "sg-NormalSpec" {
		t.Errorf("expected subgoals from NormalSpec despite panic in other specialist, got %+v", subgoals)
	}
	if len(steps) != 2 {
		t.Fatalf("expected 2 step logs (1 normal + 1 panic recovered), got %d", len(steps))
	}

	panicStepFound := false
	for _, st := range steps {
		if st.SpecialistName == "PanicSpec" && st.Status == "ERROR_PANIC" {
			panicStepFound = true
		}
	}
	if !panicStepFound {
		t.Error("expected ERROR_PANIC step log for PanicSpec")
	}

	summary := cache.Summary()
	if summary.CacheHits+summary.CacheMisses != 2 {
		t.Errorf("expected 2 cache evaluations recorded across both specialists, got %+v", summary)
	}
}

func TestSpecialistRegistry_TimeoutAndCancellation(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(&mockSpecialist{name: "SlowSpec", domains: []string{"General"}, duration: 500 * time.Millisecond})

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-timeout").
		WithGoal("Test timeout").
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	steps, _, _, err := reg.ExecuteSpecialists(ctx, req, &DependencyGraphSnapshot{}, DefaultPlanningPolicyProfile(), nil)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("expected deadline exceeded error, got: %v / ctx.Err=%v", err, ctx.Err())
	}
	if len(steps) == 0 {
		t.Error("expected cancelled step log produced on timeout")
	}
}
