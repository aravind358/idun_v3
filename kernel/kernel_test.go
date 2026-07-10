// Tests for the Kernel Boot Mechanism.
//
// In Go, test files always end in _test.go.
// The testing package is part of the standard library.
// Run all tests with: go test ./kernel/...
//
// These tests verify the Boot Mechanism in isolation.
// They use a minimal stub component so no real implementations are required.
// This proves the Boot Mechanism works correctly before any other component exists.
package kernel

import "testing"

// ============================================================
// Test Helpers
// ============================================================

// stubComponent is a minimal implementation of the Component interface.
// It exists only in this test file. It has no behaviour beyond returning its name.
//
// In Go, test helpers live in the _test.go file.
// They are compiled only during testing and never included in the production binary.
type stubComponent struct {
	name string
}

// Name satisfies the Component interface for stubComponent.
func (s *stubComponent) Name() string {
	return s.name
}

// newStub creates a named Component stub.
// A constructor function like this keeps test cases short and readable.
func newStub(name string) Component {
	return &stubComponent{name: name}
}

// fullConfig returns a Config with all four components populated.
// Used in tests that need a valid Kernel without caring about component details.
func fullConfig() Config {
	return Config{
		Registry:   newStub("StubRegistry"),
		Bus:        newStub("StubBus"),
		Boundary:   newStub("StubBoundary"),
		Permission: newStub("StubPermission"),
	}
}

// ============================================================
// Boot Tests
// ============================================================

// TestBoot_Success verifies that Boot succeeds when all components are provided.
func TestBoot_Success(t *testing.T) {
	k, err := Boot(fullConfig())

	// In Go, t.Fatalf stops the test immediately. Use it when continuing is pointless.
	if err != nil {
		t.Fatalf("Boot() returned unexpected error: %v", err)
	}
	if k == nil {
		t.Fatal("Boot() returned nil Kernel, expected a valid Kernel")
	}
	if !k.IsRunning() {
		t.Error("Kernel should be running immediately after Boot")
	}
}

// TestBoot_MissingComponents verifies that Boot fails if any required component is nil.
//
// This uses a table-driven test — the idiomatic Go pattern for testing
// multiple similar cases. Each case is a struct in a slice.
// The test loops over all cases, which keeps the logic in one place.
func TestBoot_MissingComponents(t *testing.T) {
	cases := []struct {
		name string // human-readable description of the missing component
		cfg  Config // the incomplete Config under test
	}{
		{
			name: "missing Registry",
			cfg: Config{
				// Registry intentionally omitted (nil)
				Bus:        newStub("Bus"),
				Boundary:   newStub("Boundary"),
				Permission: newStub("Permission"),
			},
		},
		{
			name: "missing Bus",
			cfg: Config{
				Registry: newStub("Registry"),
				// Bus intentionally omitted (nil)
				Boundary:   newStub("Boundary"),
				Permission: newStub("Permission"),
			},
		},
		{
			name: "missing Boundary",
			cfg: Config{
				Registry: newStub("Registry"),
				Bus:      newStub("Bus"),
				// Boundary intentionally omitted (nil)
				Permission: newStub("Permission"),
			},
		},
		{
			name: "missing Permission",
			cfg: Config{
				Registry: newStub("Registry"),
				Bus:      newStub("Bus"),
				Boundary: newStub("Boundary"),
				// Permission intentionally omitted (nil)
			},
		},
	}

	for _, tc := range cases {
		// t.Run creates a named sub-test.
		// Each sub-test appears separately in the test output.
		// Run a specific one with: go test -run TestBoot_MissingComponents/missing_Registry
		t.Run(tc.name, func(t *testing.T) {
			k, err := Boot(tc.cfg)

			if err == nil {
				t.Errorf("Boot() should have returned an error for: %s", tc.name)
			}
			if k != nil {
				t.Errorf("Boot() should return nil Kernel on error, got non-nil for: %s", tc.name)
			}
		})
	}
}

// ============================================================
// Shutdown Tests
// ============================================================

// TestShutdown verifies that Shutdown stops the Kernel.
func TestShutdown(t *testing.T) {
	k, err := Boot(fullConfig())
	if err != nil {
		t.Fatalf("Boot() failed unexpectedly: %v", err)
	}

	k.Shutdown()

	if k.IsRunning() {
		t.Error("Kernel should not be running after Shutdown")
	}
}

// TestShutdown_Idempotent verifies that calling Shutdown twice does not panic or error.
//
// Idempotency is a critical property of any shutdown mechanism.
// Real programs often call Shutdown from multiple places:
// signal handlers, deferred calls, error paths.
// The Kernel must handle all of them gracefully.
func TestShutdown_Idempotent(t *testing.T) {
	k, err := Boot(fullConfig())
	if err != nil {
		t.Fatalf("Boot() failed unexpectedly: %v", err)
	}

	k.Shutdown()
	k.Shutdown() // Second call must not panic or change anything

	if k.IsRunning() {
		t.Error("Kernel should remain stopped after double Shutdown")
	}
}
