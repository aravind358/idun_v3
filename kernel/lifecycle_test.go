package kernel

import (
	"context"
	"errors"
	"testing"
)

type mockLifecycleComponent struct {
	name      string
	phase     Phase
	startFunc func(ctx context.Context) error
	closeFunc func() error
}

func (m *mockLifecycleComponent) Name() string {
	return m.name
}

func (m *mockLifecycleComponent) BootPhase() Phase {
	return m.phase
}

func (m *mockLifecycleComponent) Start(ctx context.Context) error {
	if m.startFunc != nil {
		return m.startFunc(ctx)
	}
	return nil
}

func (m *mockLifecycleComponent) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestLifecycleTopologicalOrder(t *testing.T) {
	var order []string

	c1 := &mockLifecycleComponent{
		name:  "CoreComp",
		phase: PhaseCore,
		startFunc: func(ctx context.Context) error {
			order = append(order, "start-core")
			return nil
		},
		closeFunc: func() error {
			order = append(order, "close-core")
			return nil
		},
	}

	c2 := &mockLifecycleComponent{
		name:  "CogComp",
		phase: PhaseCognitive,
		startFunc: func(ctx context.Context) error {
			order = append(order, "start-cog")
			return nil
		},
		closeFunc: func() error {
			order = append(order, "close-cog")
			return nil
		},
	}

	reg := NewRegistry()
	_ = reg.Register(c1)
	_ = reg.Register(c2)

	cfg := Config{
		Registry:   reg,
		Bus:        newStub("StubBus"),
		Boundary:   newStub("StubBoundary"),
		Permission: newStub("StubPermission"),
	}

	k, err := Boot(cfg)
	if err != nil {
		t.Fatalf("Boot failed: %v", err)
	}
	if k.StartedCount() != 2 {
		t.Fatalf("Expected 2 started components, got %d", k.StartedCount())
	}

	if len(order) < 2 || order[0] != "start-core" || order[1] != "start-cog" {
		t.Fatalf("Unexpected start order: %v", order)
	}

	k.Shutdown()

	if len(order) != 4 || order[2] != "close-cog" || order[3] != "close-core" {
		t.Fatalf("Unexpected shutdown order: %v", order)
	}
}

func TestLifecycleStartErrorHaltsBoot(t *testing.T) {
	var closed bool
	c1 := &mockLifecycleComponent{
		name:  "GoodCore",
		phase: PhaseCore,
		closeFunc: func() error {
			closed = true
			return nil
		},
	}
	c2 := &mockLifecycleComponent{
		name:  "BadInfrastructure",
		phase: PhaseInfrastructure,
		startFunc: func(ctx context.Context) error {
			return errors.New("start exploded")
		},
	}

	reg := NewRegistry()
	_ = reg.Register(c1)
	_ = reg.Register(c2)

	cfg := Config{
		Registry:   reg,
		Bus:        newStub("StubBus"),
		Boundary:   newStub("StubBoundary"),
		Permission: newStub("StubPermission"),
	}

	k, err := Boot(cfg)
	if err == nil {
		t.Fatal("Expected Boot to fail when component Start errors")
	}
	if k != nil {
		t.Fatal("Expected nil Kernel when Boot fails")
	}
	if !closed {
		t.Fatal("Expected successfully started components to be closed when subsequent component fails")
	}
}
