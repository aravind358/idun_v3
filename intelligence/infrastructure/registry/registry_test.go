package registry_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"idun/intelligence/infrastructure/registry"
)

func validDescriptor(id, version string) registry.BackendDescriptor {
	return registry.BackendDescriptor{
		ID:             id,
		DriverScheme:   "grpc",
		Endpoint:       "localhost:8080",
		Version:        version,
		MaxConcurrency: 4,
		SupportedBudgets: []string{"REFLEXIVE", "STANDARD"},
		DriverConfig:   map[string]string{"timeout_ms": "100"},
	}
}

func TestLifecycle(t *testing.T) {
	svc := registry.NewService()
	if svc.Name() != "Intelligence.Infrastructure.Registry" {
		t.Fatalf("unexpected name: %s", svc.Name())
	}

	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Operations after close should return ErrRegistryClosed
	ctx := context.Background()
	err := svc.Register(ctx, "model-1", validDescriptor("b-1", "v1.0"))
	if !errors.Is(err, registry.ErrRegistryClosed) {
		t.Fatalf("expected ErrRegistryClosed, got: %v", err)
	}
}

func TestRegisterAndResolve(t *testing.T) {
	svc := registry.NewService()
	ctx := context.Background()

	// Resolve unregistered should fail
	_, err := svc.Resolve(ctx, "missing-model")
	if !errors.Is(err, registry.ErrModelNotFound) {
		t.Fatalf("expected ErrModelNotFound, got: %v", err)
	}

	// Register valid backend
	bd := validDescriptor("backend-v1", "1.0.0")
	if err := svc.Register(ctx, "language-reasoner", bd); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Resolve should succeed
	resolved, err := svc.Resolve(ctx, "language-reasoner")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if resolved.ID != "backend-v1" || resolved.Version != "1.0.0" {
		t.Fatalf("unexpected resolved descriptor: %+v", resolved)
	}

	// Modifying resolved descriptor should not mutate internal state (cloning check)
	resolved.DriverConfig["timeout_ms"] = "9999"
	resolvedAgain, _ := svc.Resolve(ctx, "language-reasoner")
	if resolvedAgain.DriverConfig["timeout_ms"] != "100" {
		t.Fatalf("internal map mutated! cloning failed")
	}
}

func TestRegisterInvalid(t *testing.T) {
	svc := registry.NewService()
	ctx := context.Background()

	err := svc.Register(ctx, "", validDescriptor("b-1", "1.0"))
	if !errors.Is(err, registry.ErrInvalidDescriptor) {
		t.Fatalf("expected ErrInvalidDescriptor on empty modelID, got: %v", err)
	}

	invalidBD := validDescriptor("b-1", "1.0")
	invalidBD.MaxConcurrency = 0
	err = svc.Register(ctx, "test", invalidBD)
	if !errors.Is(err, registry.ErrInvalidDescriptor) {
		t.Fatalf("expected ErrInvalidDescriptor on 0 MaxConcurrency, got: %v", err)
	}
}

func TestDeregister(t *testing.T) {
	svc := registry.NewService()
	ctx := context.Background()

	_ = svc.Register(ctx, "model-a", validDescriptor("b-1", "v1"))
	if err := svc.Deregister(ctx, "model-a"); err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	_, err := svc.Resolve(ctx, "model-a")
	if !errors.Is(err, registry.ErrModelNotFound) {
		t.Fatalf("expected ErrModelNotFound after deregister, got: %v", err)
	}
}

func TestHealthTracking(t *testing.T) {
	svc := registry.NewService()
	ctx := context.Background()

	_ = svc.Register(ctx, "model-health", validDescriptor("b-1", "v1"))

	// Set unhealthy
	if err := svc.SetHealth(ctx, "model-health", registry.HealthUnhealthy, "GPU OOM"); err != nil {
		t.Fatalf("SetHealth failed: %v", err)
	}

	// Resolve should fail with ErrBackendUnavailable
	_, err := svc.Resolve(ctx, "model-health")
	if !errors.Is(err, registry.ErrBackendUnavailable) {
		t.Fatalf("expected ErrBackendUnavailable, got: %v", err)
	}

	// Set back to healthy
	_ = svc.SetHealth(ctx, "model-health", registry.HealthHealthy, "recovered")
	_, err = svc.Resolve(ctx, "model-health")
	if err != nil {
		t.Fatalf("Resolve should succeed after recovery, got: %v", err)
	}
}

func TestRollback(t *testing.T) {
	svc := registry.NewService()
	ctx := context.Background()

	_ = svc.Register(ctx, "model-rollback", validDescriptor("b-1", "1.0.0"))
	_ = svc.Register(ctx, "model-rollback", validDescriptor("b-2", "2.0.0"))

	res, _ := svc.Resolve(ctx, "model-rollback")
	if res.Version != "2.0.0" {
		t.Fatalf("expected 2.0.0 active, got %s", res.Version)
	}

	// Rollback to 1.0.0
	if err := svc.Rollback(ctx, "model-rollback", "1.0.0"); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	res, _ = svc.Resolve(ctx, "model-rollback")
	if res.Version != "1.0.0" {
		t.Fatalf("expected 1.0.0 after rollback, got %s", res.Version)
	}

	// Rollback to missing version should fail
	err := svc.Rollback(ctx, "model-rollback", "9.9.9")
	if !errors.Is(err, registry.ErrVersionNotFound) {
		t.Fatalf("expected ErrVersionNotFound, got: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	svc := registry.NewService()
	ctx := context.Background()

	_ = svc.Register(ctx, "model-concurrency", validDescriptor("b-0", "v0"))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%3 == 0 {
				_ = svc.Register(ctx, "model-concurrency", validDescriptor("b-next", "v1"))
			} else if idx%3 == 1 {
				_, _ = svc.Resolve(ctx, "model-concurrency")
			} else {
				_ = svc.SetHealth(ctx, "model-concurrency", registry.HealthHealthy, "ok")
			}
		}(i)
	}
	wg.Wait()
}

func TestTelemetry(t *testing.T) {
	svc := registry.NewService()
	ctx := context.Background()

	_ = svc.Register(ctx, "m1", validDescriptor("b1", "v1"))
	_, _ = svc.Resolve(ctx, "m1")
	_, _ = svc.Resolve(ctx, "missing")

	snap := svc.GetTelemetry()
	if snap.TotalRegisteredModels != 1 {
		t.Fatalf("expected 1 registered model, got %d", snap.TotalRegisteredModels)
	}
	if snap.TotalResolutions != 1 || snap.FailedResolutions != 1 {
		t.Fatalf("unexpected telemetry snapshot: %+v", snap)
	}
}

func TestContextCancellation(t *testing.T) {
	svc := registry.NewService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.Register(ctx, "m1", validDescriptor("b1", "v1"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}
