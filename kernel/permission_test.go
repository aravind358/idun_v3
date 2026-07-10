// Tests for the Permission Engine.
//
// All tests live in package kernel (white-box), consistent with the other
// component test files. No unexported fields are accessed directly — only
// public behaviour is tested.
//
// Stubs from other test files available in this package:
//   - stubComponent, newStub            (kernel_test.go)
//   - emptyNameComponent                (registry_test.go)
//   - stubValidator, stubHandler        (bus_test.go)
//   - passingValidator, passingAuthorizer, newHandler, validBus (bus_test.go)
//
// Do NOT redeclare any of the above.
package kernel

import (
	"testing"
)

// ============================================================
// NewPermissionEngine
// ============================================================

// TestNewPermissionEngine verifies that NewPermissionEngine returns a
// non-nil engine that starts in a deny-all state.
func TestNewPermissionEngine(t *testing.T) {
	p := NewPermissionEngine()

	if p == nil {
		t.Fatal("NewPermissionEngine() returned nil, expected a valid *PermissionEngine")
	}

	// A fresh engine must deny every call — verified more thoroughly in
	// TestPermissionEngine_DefaultDeny, but confirm the basic case here.
	if err := p.Authorize("anyone", "anywhere"); err == nil {
		t.Error("a fresh PermissionEngine should deny all; Authorize returned nil")
	}
}

// ============================================================
// Name
// ============================================================

// TestPermissionEngine_Name verifies the engine reports the correct component
// name and satisfies the Component interface at compile time.
func TestPermissionEngine_Name(t *testing.T) {
	p := NewPermissionEngine()

	// Compile-time proof: *PermissionEngine must satisfy Component.
	var _ Component = p

	got := p.Name()
	want := "PermissionEngine"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// TestPermissionEngine_AuthorizerCompliance is a compile-time proof that
// *PermissionEngine satisfies the Authorizer interface defined in bus.go.
// This never runs as a meaningful assertion — it fails at compile time if
// the interface drifts.
func TestPermissionEngine_AuthorizerCompliance(t *testing.T) {
	var _ Authorizer = NewPermissionEngine()
}

// ============================================================
// Allow
// ============================================================

// TestAllow_Success verifies that a valid (caller, target) pair is stored
// and that a subsequent Authorize call for that pair succeeds.
func TestAllow_Success(t *testing.T) {
	p := NewPermissionEngine()

	if err := p.Allow("Alpha", "Beta"); err != nil {
		t.Fatalf("Allow() returned unexpected error: %v", err)
	}

	if err := p.Authorize("Alpha", "Beta"); err != nil {
		t.Errorf("Authorize() should succeed after Allow; got: %v", err)
	}
}

// TestAllow_EmptyInputs verifies that Allow rejects empty caller and empty
// target via a table-driven test.
func TestAllow_EmptyInputs(t *testing.T) {
	cases := []struct {
		name   string
		caller string
		target string
	}{
		{name: "empty caller", caller: "", target: "Beta"},
		{name: "empty target", caller: "Alpha", target: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPermissionEngine()
			err := p.Allow(tc.caller, tc.target)
			if err == nil {
				t.Errorf("Allow(%q, %q) should have returned an error", tc.caller, tc.target)
			}
		})
	}
}

// TestAllow_Duplicate verifies that registering the same (caller, target)
// pair twice returns an error.
//
// This is consistent with the Registry's Register behaviour: duplicate
// registrations are surfaced as errors rather than silently accepted,
// because they are most likely programming mistakes in the wiring code.
func TestAllow_Duplicate(t *testing.T) {
	p := NewPermissionEngine()

	if err := p.Allow("Alpha", "Beta"); err != nil {
		t.Fatalf("first Allow() returned unexpected error: %v", err)
	}

	err := p.Allow("Alpha", "Beta")
	if err == nil {
		t.Error("second Allow() with the same pair should have returned an error")
	}
}

// TestAllow_DifferentTargetsSameCaller verifies that a caller can be allowed
// to send to multiple distinct targets independently.
func TestAllow_DifferentTargetsSameCaller(t *testing.T) {
	p := NewPermissionEngine()

	if err := p.Allow("Alpha", "Beta"); err != nil {
		t.Fatalf("Allow(Alpha, Beta) failed: %v", err)
	}
	if err := p.Allow("Alpha", "Gamma"); err != nil {
		t.Fatalf("Allow(Alpha, Gamma) failed: %v", err)
	}

	if err := p.Authorize("Alpha", "Beta"); err != nil {
		t.Errorf("Authorize(Alpha, Beta) should succeed: %v", err)
	}
	if err := p.Authorize("Alpha", "Gamma"); err != nil {
		t.Errorf("Authorize(Alpha, Gamma) should succeed: %v", err)
	}
}

// ============================================================
// Authorize
// ============================================================

// TestAuthorize_Allowed verifies that Authorize returns nil when the
// (caller, target) pair has been explicitly permitted.
func TestAuthorize_Allowed(t *testing.T) {
	p := NewPermissionEngine()
	_ = p.Allow("Charlie", "Delta")

	if err := p.Authorize("Charlie", "Delta"); err != nil {
		t.Errorf("Authorize() returned error for a permitted pair: %v", err)
	}
}

// TestAuthorize_Denied verifies that Authorize returns an error when no
// permission exists for the (caller, target) pair.
func TestAuthorize_Denied(t *testing.T) {
	p := NewPermissionEngine()
	// No Allow() call — engine is in deny-all state.

	if err := p.Authorize("Echo", "Foxtrot"); err == nil {
		t.Error("Authorize() should return an error when no permission exists")
	}
}

// TestAuthorize_EmptyInputs verifies that Authorize rejects empty caller
// and empty target via a table-driven test.
func TestAuthorize_EmptyInputs(t *testing.T) {
	cases := []struct {
		name   string
		caller string
		target string
	}{
		{name: "empty caller", caller: "", target: "Delta"},
		{name: "empty target", caller: "Charlie", target: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPermissionEngine()
			svc, err := p.Authorize(tc.caller, tc.target), error(nil)
			_ = svc
			err = p.Authorize(tc.caller, tc.target)
			if err == nil {
				t.Errorf("Authorize(%q, %q) should have returned an error", tc.caller, tc.target)
			}
		})
	}
}

// ============================================================
// Default deny
// ============================================================

// TestPermissionEngine_DefaultDeny verifies that a fresh engine denies
// every caller/target combination — not just the one it was never asked about.
func TestPermissionEngine_DefaultDeny(t *testing.T) {
	p := NewPermissionEngine()

	pairs := [][2]string{
		{"ServiceA", "ServiceB"},
		{"ServiceB", "ServiceA"},
		{"anything", "anything"},
	}

	for _, pair := range pairs {
		caller, target := pair[0], pair[1]
		if err := p.Authorize(caller, target); err == nil {
			t.Errorf("fresh engine should deny %q → %q, but Authorize returned nil", caller, target)
		}
	}
}

// ============================================================
// Permission direction is not symmetric
// ============================================================

// TestPermissionEngine_IndependentPairs verifies that allowing A → B does
// not implicitly allow B → A.
//
// This is a critical invariant: permissions are one-directional. Failing
// to enforce this would mean granting any permission accidentally grants
// bidirectional communication.
func TestPermissionEngine_IndependentPairs(t *testing.T) {
	p := NewPermissionEngine()
	_ = p.Allow("Golf", "Hotel")

	// Forward direction must be allowed.
	if err := p.Authorize("Golf", "Hotel"); err != nil {
		t.Errorf("Authorize(Golf, Hotel) should succeed: %v", err)
	}

	// Reverse direction must remain denied.
	if err := p.Authorize("Hotel", "Golf"); err == nil {
		t.Error("Authorize(Hotel, Golf) should be denied — permissions are not symmetric")
	}
}

// TestPermissionEngine_PermissionDoesNotLeakAcrossCallers verifies that
// allowing Caller A to reach Target X does not allow Caller B to reach Target X.
func TestPermissionEngine_PermissionDoesNotLeakAcrossCallers(t *testing.T) {
	p := NewPermissionEngine()
	_ = p.Allow("India", "Juliet")

	if err := p.Authorize("India", "Juliet"); err != nil {
		t.Errorf("Authorize(India, Juliet) should succeed: %v", err)
	}

	if err := p.Authorize("Kilo", "Juliet"); err == nil {
		t.Error("Authorize(Kilo, Juliet) should be denied — permission was granted to India, not Kilo")
	}
}

// ============================================================
// Integration with the Bus
// ============================================================

// TestPermissionEngine_IntegratesWithBus verifies that the Bus correctly
// blocks a Send when the PermissionEngine denies the caller, and correctly
// delivers a Send when permission has been granted.
//
// This is the most important integration test: it proves the end-to-end
// path from Allow() through Bus.Send() to Handler.Handle() works correctly.
func TestPermissionEngine_IntegratesWithBus(t *testing.T) {
	reg := NewRegistry()
	engine := NewPermissionEngine()

	bus, err := NewBus(reg, passingValidator(), engine)
	if err != nil {
		t.Fatalf("NewBus() failed: %v", err)
	}

	h := newHandler("Lima", "received")
	if err := reg.Register(h); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	msg := Message{From: "Mike", To: "Lima", Body: nil}

	// Without permission, Send must fail.
	if _, err := bus.Send(msg); err == nil {
		t.Error("Send() should be denied before Allow() is called")
	}

	// Grant permission.
	if err := engine.Allow("Mike", "Lima"); err != nil {
		t.Fatalf("Allow() failed: %v", err)
	}

	// With permission, Send must succeed.
	resp, err := bus.Send(msg)
	if err != nil {
		t.Fatalf("Send() should succeed after Allow(); got: %v", err)
	}
	if resp != "received" {
		t.Errorf("Send() returned %v, want %q", resp, "received")
	}
}

// ============================================================
// Integration with Kernel Boot
// ============================================================

// TestPermissionEngine_IntegratesWithKernelBoot verifies that a real
// PermissionEngine is accepted by kernel.Boot without any changes to
// kernel.go. This confirms the Component interface is satisfied.
func TestPermissionEngine_IntegratesWithKernelBoot(t *testing.T) {
	reg := NewRegistry()
	engine := NewPermissionEngine()

	bus, err := NewBus(reg, passingValidator(), engine)
	if err != nil {
		t.Fatalf("NewBus() failed: %v", err)
	}

	k, err := Boot(Config{
		Registry:   reg,
		Bus:        bus,
		Boundary:   newStub("StubBoundary"),
		Permission: engine,
	})
	if err != nil {
		t.Fatalf("Boot() failed with real PermissionEngine: %v", err)
	}
	if !k.IsRunning() {
		t.Error("Kernel should be running after Boot with a real PermissionEngine")
	}
}
