// Tests for the Boundary Engine.
//
// All tests live in package kernel (white-box), consistent with the other
// component test files.
package kernel

import (
	"testing"
)

// ============================================================
// NewBoundaryEngine
// ============================================================

// TestNewBoundaryEngine verifies that NewBoundaryEngine returns a non-nil engine.
func TestNewBoundaryEngine(t *testing.T) {
	b := NewBoundaryEngine()
	if b == nil {
		t.Fatal("NewBoundaryEngine() returned nil, expected a valid *BoundaryEngine")
	}
}

// ============================================================
// Name
// ============================================================

// TestBoundaryEngine_Name verifies the engine reports the correct component
// name and satisfies the Component interface at compile time.
func TestBoundaryEngine_Name(t *testing.T) {
	b := NewBoundaryEngine()

	// Compile-time proof: *BoundaryEngine must satisfy Component.
	var _ Component = b

	got := b.Name()
	want := "BoundaryEngine"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// TestBoundaryEngine_ValidatorCompliance is a compile-time proof that
// *BoundaryEngine satisfies the Validator interface defined in bus.go.
func TestBoundaryEngine_ValidatorCompliance(t *testing.T) {
	var _ Validator = NewBoundaryEngine()
}

// ============================================================
// Validate
// ============================================================

// TestValidate_Success verifies that a structurally valid message succeeds.
func TestValidate_Success(t *testing.T) {
	b := NewBoundaryEngine()

	msg := Message{From: "SenderService", To: "ReceiverService", Body: "hello"}
	if err := b.Validate(msg); err != nil {
		t.Errorf("Validate() returned unexpected error for valid message: %v", err)
	}
}

// TestValidate_InvalidInputs verifies that Validate rejects messages with
// empty From or empty To via a table-driven test.
func TestValidate_InvalidInputs(t *testing.T) {
	cases := []struct {
		name string
		msg  Message
	}{
		{name: "empty From", msg: Message{From: "", To: "Receiver", Body: nil}},
		{name: "empty To", msg: Message{From: "Sender", To: "", Body: nil}},
		{name: "both empty", msg: Message{From: "", To: "", Body: nil}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBoundaryEngine()
			if err := b.Validate(tc.msg); err == nil {
				t.Errorf("Validate() should have returned an error for: %s", tc.name)
			}
		})
	}
}

// TestValidate_NilBodyAllowed verifies that a nil Body is structurally valid in V1.
func TestValidate_NilBodyAllowed(t *testing.T) {
	b := NewBoundaryEngine()

	msg := Message{From: "Sender", To: "Receiver", Body: nil}
	if err := b.Validate(msg); err != nil {
		t.Errorf("Validate() returned unexpected error for nil Body: %v", err)
	}
}

// ============================================================
// Integration with Bus
// ============================================================

// TestBoundaryEngine_IntegratesWithBus verifies that the Bus correctly
// delegates structural validation to a real BoundaryEngine.
func TestBoundaryEngine_IntegratesWithBus(t *testing.T) {
	reg := NewRegistry()
	boundary := NewBoundaryEngine()

	bus, err := NewBus(reg, boundary, passingAuthorizer())
	if err != nil {
		t.Fatalf("NewBus() failed: %v", err)
	}

	h := newHandler("Target", "ok")
	if err := reg.Register(h); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Invalid message structure should be rejected by BoundaryEngine via Bus.Send.
	_, err = bus.Send(Message{From: "", To: "Target", Body: nil})
	if err == nil {
		t.Error("Send() with empty From should be rejected by BoundaryEngine")
	}

	// Valid message structure should succeed.
	resp, err := bus.Send(Message{From: "Sender", To: "Target", Body: nil})
	if err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("Send() returned %v, want %q", resp, "ok")
	}
}

// ============================================================
// Integration with Kernel Boot
// ============================================================

// TestBoundaryEngine_IntegratesWithKernelBoot verifies that a real
// BoundaryEngine is accepted by kernel.Boot without any issues.
func TestBoundaryEngine_IntegratesWithKernelBoot(t *testing.T) {
	reg := NewRegistry()
	boundary := NewBoundaryEngine()

	bus, err := NewBus(reg, boundary, passingAuthorizer())
	if err != nil {
		t.Fatalf("NewBus() failed: %v", err)
	}

	k, err := Boot(Config{
		Registry:   reg,
		Bus:        bus,
		Boundary:   boundary,
		Permission: newStub("StubPermission"),
	})
	if err != nil {
		t.Fatalf("Boot() failed with real BoundaryEngine: %v", err)
	}
	if !k.IsRunning() {
		t.Error("Kernel should be running after Boot with real BoundaryEngine")
	}
}
