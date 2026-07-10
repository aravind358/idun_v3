// Tests for the Communication Bus.
//
// All tests live in package kernel (white-box), consistent with kernel_test.go
// and registry_test.go. This allows direct access to unexported internals
// when needed, but the tests only exercise public behaviour.
//
// Stubs defined here:
//   - stubValidator  — configurable Validator (always pass or always fail)
//   - stubAuthorizer — configurable Authorizer (always pass or always fail)
//   - stubHandler    — a Component that also implements Handler
//
// stubComponent and newStub are already defined in kernel_test.go and are
// available here because all test files in the same package compile together.
// Do NOT redefine them.
package kernel

import (
	"errors"
	"fmt"
	"testing"
)

// ============================================================
// Stubs
// ============================================================

// stubValidator is a configurable Validator for tests.
// When err is nil, Validate succeeds. When err is non-nil, Validate fails.
type stubValidator struct {
	err error
}

func (v *stubValidator) Validate(_ Message) error { return v.err }

// stubAuthorizer is a configurable Authorizer for tests.
// When err is nil, Authorize succeeds. When err is non-nil, Authorize fails.
type stubAuthorizer struct {
	err error
}

func (a *stubAuthorizer) Authorize(_, _ string) error { return a.err }

// stubHandler is a Component that also implements Handler.
// It records the last message it received so tests can inspect it.
// response and handleErr are returned verbatim by Handle.
type stubHandler struct {
	handlerName string
	response    any
	handleErr   error
	received    Message // last message delivered to this handler
}

func (h *stubHandler) Name() string { return h.handlerName }

func (h *stubHandler) Handle(msg Message) (any, error) {
	h.received = msg
	return h.response, h.handleErr
}

// ============================================================
// Helpers
// ============================================================

// passingValidator returns a Validator that always succeeds.
func passingValidator() Validator { return &stubValidator{err: nil} }

// passingAuthorizer returns an Authorizer that always succeeds.
func passingAuthorizer() Authorizer { return &stubAuthorizer{err: nil} }

// validBus returns a Bus wired to a fresh Registry with passing stubs.
// The Registry is also returned so individual tests can populate it.
func validBus(t *testing.T) (*Bus, *Registry) {
	t.Helper()
	reg := NewRegistry()
	bus, err := NewBus(reg, passingValidator(), passingAuthorizer())
	if err != nil {
		t.Fatalf("NewBus() returned unexpected error: %v", err)
	}
	return bus, reg
}

// newHandler creates a stubHandler with the given name and response.
func newHandler(name string, response any) *stubHandler {
	return &stubHandler{handlerName: name, response: response}
}

// ============================================================
// NewBus — construction
// ============================================================

// TestNewBus_Success verifies that NewBus succeeds when all dependencies
// are provided.
func TestNewBus_Success(t *testing.T) {
	reg := NewRegistry()
	bus, err := NewBus(reg, passingValidator(), passingAuthorizer())

	if err != nil {
		t.Fatalf("NewBus() returned unexpected error: %v", err)
	}
	if bus == nil {
		t.Fatal("NewBus() returned nil Bus, expected a valid *Bus")
	}
}

// TestNewBus_NilDependencies verifies that NewBus rejects any nil dependency.
func TestNewBus_NilDependencies(t *testing.T) {
	reg := NewRegistry()

	cases := []struct {
		name       string
		registry   Locator // must be Locator (not *Registry) so nil stays a true interface nil
		validator  Validator
		authorizer Authorizer
	}{
		{
			name:       "nil registry",
			registry:   nil,
			validator:  passingValidator(),
			authorizer: passingAuthorizer(),
		},
		{
			name:       "nil validator",
			registry:   reg,
			validator:  nil,
			authorizer: passingAuthorizer(),
		},
		{
			name:       "nil authorizer",
			registry:   reg,
			validator:  passingValidator(),
			authorizer: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bus, err := NewBus(tc.registry, tc.validator, tc.authorizer)
			if err == nil {
				t.Errorf("NewBus() should have returned an error for: %s", tc.name)
			}
			if bus != nil {
				t.Errorf("NewBus() should return nil Bus on error, got non-nil for: %s", tc.name)
			}
		})
	}
}

// ============================================================
// Name
// ============================================================

// TestBus_Name verifies the Bus reports the correct component name and
// confirms at compile time that *Bus satisfies Component.
func TestBus_Name(t *testing.T) {
	bus, _ := validBus(t)

	// Compile-time check: *Bus must satisfy Component.
	var _ Component = bus

	got := bus.Name()
	want := "CommunicationBus"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// ============================================================
// Send — message structure validation via Boundary Engine (step 1)
// ============================================================

// TestBus_Send_EmptyFrom verifies that Send rejects a message with no sender
// via delegation to the Boundary Engine.
func TestBus_Send_EmptyFrom(t *testing.T) {
	reg := NewRegistry()
	bus, err := NewBus(reg, NewBoundaryEngine(), passingAuthorizer())
	if err != nil {
		t.Fatalf("NewBus() failed: %v", err)
	}

	_, err = bus.Send(Message{From: "", To: "SomeService", Body: nil})
	if err == nil {
		t.Error("Send() with empty From should have returned an error")
	}
}

// TestBus_Send_EmptyTo verifies that Send rejects a message with no target
// via delegation to the Boundary Engine.
func TestBus_Send_EmptyTo(t *testing.T) {
	reg := NewRegistry()
	bus, err := NewBus(reg, NewBoundaryEngine(), passingAuthorizer())
	if err != nil {
		t.Fatalf("NewBus() failed: %v", err)
	}

	_, err = bus.Send(Message{From: "SomeSender", To: "", Body: nil})
	if err == nil {
		t.Error("Send() with empty To should have returned an error")
	}
}

// ============================================================
// Send — validation failure (step 2)
// ============================================================

// TestBus_Send_ValidationFails verifies that Send stops and returns an error
// when the Validator rejects the message.
func TestBus_Send_ValidationFails(t *testing.T) {
	reg := NewRegistry()
	failingValidator := &stubValidator{err: errors.New("invalid message body")}
	bus, err := NewBus(reg, failingValidator, passingAuthorizer())
	if err != nil {
		t.Fatalf("NewBus() failed unexpectedly: %v", err)
	}

	_, sendErr := bus.Send(Message{From: "Sender", To: "Target", Body: "bad"})
	if sendErr == nil {
		t.Error("Send() should have returned an error when Validator fails")
	}
	// Confirm the original error is wrapped and accessible.
	if !errors.Is(sendErr, failingValidator.err) {
		t.Errorf("Send() error should wrap the validator error; got: %v", sendErr)
	}
}

// ============================================================
// Send — authorization failure (step 3)
// ============================================================

// TestBus_Send_AuthorizationFails verifies that Send stops and returns an error
// when the Authorizer rejects the caller.
func TestBus_Send_AuthorizationFails(t *testing.T) {
	reg := NewRegistry()
	failingAuthorizer := &stubAuthorizer{err: errors.New("permission denied")}
	bus, err := NewBus(reg, passingValidator(), failingAuthorizer)
	if err != nil {
		t.Fatalf("NewBus() failed unexpectedly: %v", err)
	}

	_, sendErr := bus.Send(Message{From: "Sender", To: "Target", Body: nil})
	if sendErr == nil {
		t.Error("Send() should have returned an error when Authorizer fails")
	}
	// Confirm the original error is wrapped and accessible.
	if !errors.Is(sendErr, failingAuthorizer.err) {
		t.Errorf("Send() error should wrap the authorizer error; got: %v", sendErr)
	}
}

// TestBus_Send_AuthorizationBeforeLookup verifies that authorization is checked
// before the registry is consulted.
//
// This prevents an unauthorized caller from inferring whether a service
// exists by observing the difference between "authorization denied" and
// "service not found".
func TestBus_Send_AuthorizationBeforeLookup(t *testing.T) {
	reg := NewRegistry()
	// "Ghost" is never registered — if lookup ran first, we'd get a
	// "not registered" error, not an authorization error.
	failingAuthorizer := &stubAuthorizer{err: errors.New("permission denied")}
	bus, err := NewBus(reg, passingValidator(), failingAuthorizer)
	if err != nil {
		t.Fatalf("NewBus() failed unexpectedly: %v", err)
	}

	_, sendErr := bus.Send(Message{From: "Sender", To: "Ghost", Body: nil})
	if sendErr == nil {
		t.Fatal("Send() should have returned an error")
	}
	if !errors.Is(sendErr, failingAuthorizer.err) {
		t.Errorf("expected authorization error, got: %v", sendErr)
	}
}

// ============================================================
// Send — registry lookup failure (step 4)
// ============================================================

// TestBus_Send_TargetNotRegistered verifies that Send returns an error when
// the target service has not been registered.
func TestBus_Send_TargetNotRegistered(t *testing.T) {
	bus, _ := validBus(t) // Registry is empty.

	_, err := bus.Send(Message{From: "Sender", To: "UnknownService", Body: nil})
	if err == nil {
		t.Error("Send() to an unregistered service should have returned an error")
	}
}

// ============================================================
// Send — handler check failure (step 5)
// ============================================================

// TestBus_Send_TargetNotHandler verifies that Send returns an error when the
// target service is registered but does not implement Handler.
//
// newStub returns a *stubComponent which has Name() but no Handle().
func TestBus_Send_TargetNotHandler(t *testing.T) {
	bus, reg := validBus(t)

	nonHandler := newStub("PassiveService") // Component only, no Handle method.
	if err := reg.Register(nonHandler); err != nil {
		t.Fatalf("Register() failed unexpectedly: %v", err)
	}

	_, err := bus.Send(Message{From: "Sender", To: "PassiveService", Body: nil})
	if err == nil {
		t.Error("Send() to a non-Handler service should have returned an error")
	}
}

// ============================================================
// Send — handler error (step 6)
// ============================================================

// TestBus_Send_HandlerError verifies that an error returned by Handle is
// propagated back to the caller unchanged.
func TestBus_Send_HandlerError(t *testing.T) {
	bus, reg := validBus(t)

	handlerErr := errors.New("handler internal error")
	h := &stubHandler{handlerName: "FaultyService", handleErr: handlerErr}
	if err := reg.Register(h); err != nil {
		t.Fatalf("Register() failed unexpectedly: %v", err)
	}

	_, sendErr := bus.Send(Message{From: "Sender", To: "FaultyService", Body: nil})
	if sendErr == nil {
		t.Error("Send() should propagate the handler's error")
	}
	if !errors.Is(sendErr, handlerErr) {
		t.Errorf("Send() error should be (or wrap) the handler error; got: %v", sendErr)
	}
}

// ============================================================
// Send — success (happy path)
// ============================================================

// TestBus_Send_Success verifies the full happy path: message is validated,
// authorized, routed to the correct handler, and the response is returned.
func TestBus_Send_Success(t *testing.T) {
	bus, reg := validBus(t)

	want := "hello back"
	h := newHandler("EchoService", want)
	if err := reg.Register(h); err != nil {
		t.Fatalf("Register() failed unexpectedly: %v", err)
	}

	got, err := bus.Send(Message{From: "Sender", To: "EchoService", Body: "hello"})
	if err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("Send() returned %v, want %v", got, want)
	}
}

// TestBus_Send_MessageDeliveredCorrectly verifies that the handler receives
// the exact Message the caller sent — From, To, and Body are all preserved.
func TestBus_Send_MessageDeliveredCorrectly(t *testing.T) {
	bus, reg := validBus(t)

	h := newHandler("TargetService", nil)
	if err := reg.Register(h); err != nil {
		t.Fatalf("Register() failed unexpectedly: %v", err)
	}

	msg := Message{From: "SourceService", To: "TargetService", Body: 42}
	_, err := bus.Send(msg)
	if err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}

	if h.received.From != msg.From {
		t.Errorf("handler received From = %q, want %q", h.received.From, msg.From)
	}
	if h.received.To != msg.To {
		t.Errorf("handler received To = %q, want %q", h.received.To, msg.To)
	}
	if h.received.Body != msg.Body {
		t.Errorf("handler received Body = %v, want %v", h.received.Body, msg.Body)
	}
}

// ============================================================
// Send — nil body is valid
// ============================================================

// TestBus_Send_NilBodyAllowed verifies that a nil Body is not treated as
// an error by the Bus itself. Body validation is the Boundary Engine's job.
func TestBus_Send_NilBodyAllowed(t *testing.T) {
	bus, reg := validBus(t)

	h := newHandler("ReceiverService", "ok")
	if err := reg.Register(h); err != nil {
		t.Fatalf("Register() failed unexpectedly: %v", err)
	}

	_, err := bus.Send(Message{From: "Sender", To: "ReceiverService", Body: nil})
	if err != nil {
		t.Errorf("Send() with nil Body should not fail at the Bus level: %v", err)
	}
}

// ============================================================
// Send — Bus is usable in Kernel Boot
// ============================================================

// TestBus_IntegratesWithKernelBoot verifies that a fully wired Bus can be
// passed to the Kernel's Boot function without any changes to kernel.go.
// This is the most important integration invariant.
func TestBus_IntegratesWithKernelBoot(t *testing.T) {
	reg := NewRegistry()
	bus, err := NewBus(reg, passingValidator(), passingAuthorizer())
	if err != nil {
		t.Fatalf("NewBus() failed: %v", err)
	}

	k, err := Boot(Config{
		Registry:   reg,
		Bus:        bus,
		Boundary:   newStub("StubBoundary"),
		Permission: newStub("StubPermission"),
	})
	if err != nil {
		t.Fatalf("Boot() failed with real Bus: %v", err)
	}
	if !k.IsRunning() {
		t.Error("Kernel should be running after Boot with a real Bus")
	}
}

// ============================================================
// Compile-time interface checks
// ============================================================

// TestBus_ValidatorInterface confirms that stubValidator satisfies Validator.
// TestBus_AuthorizerInterface confirms that stubAuthorizer satisfies Authorizer.
// TestBus_HandlerInterface confirms that stubHandler satisfies Handler.
// These never run at runtime — they fail at compile time if an interface drifts.
func TestBus_InterfaceCompliance(t *testing.T) {
	var _ Validator  = &stubValidator{}
	var _ Authorizer = &stubAuthorizer{}
	var _ Handler    = &stubHandler{}
	var _ Component  = &stubHandler{}

	// Suppress "declared and not used" from the compiler.
	_ = fmt.Sprintf
}
