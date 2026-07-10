// This file implements the Communication Bus — Kernel component three of five.
//
// Responsibility: route synchronous messages between services.
//
// The Bus is a thin router. It does not:
//   - Interpret message content   (that is the Boundary Engine's job)
//   - Decide who may talk to whom (that is the Permission Engine's job)
//   - Store or replay messages    (no queue, no retry, no pub/sub in V1)
//   - Communicate over a network  (all routing is in-process)
//
// The Bus depends on three other components:
//   - Registry   — locates the target service by name
//   - Validator  — asks the Boundary Engine whether the message is well-formed
//   - Authorizer — asks the Permission Engine whether the sender may reach the target
//
// Validator and Authorizer are defined in this file because they are the
// contracts the Bus requires. When the Boundary and Permission engines are
// built they will implement these interfaces. The Bus never imports those
// packages; the dependency arrow always points inward, toward the kernel.
//
// Ordering inside Send (validate → authorize → lookup → deliver):
//   1. Envelope validation first — catch trivially bad input before any I/O.
//   2. Boundary validation second — check message content before touching auth.
//   3. Authorization third — do not reveal whether a target exists to an
//      unauthorized caller (registry lookup comes after, not before).
//   4. Registry lookup fourth — only now locate the concrete target.
//   5. Delivery last — call the handler.
package kernel

import (
	"errors"
	"fmt"
)

// ============================================================
// Message
// ============================================================

// Message is the unit of communication between services.
//
// Fields are intentionally minimal for V1.
//   - From identifies the sender by its registered name.
//   - To identifies the target service by its registered name.
//   - Body carries the payload. The type is any (alias for interface{}) so
//     the Bus remains payload-agnostic. The Boundary Engine is responsible
//     for inspecting and validating the body's actual content.
//
// Why not add a correlation ID, timestamp, or version field now?
// V1 is synchronous. The caller already holds the call stack. None of those
// fields serve a V1 purpose. They can be added later without breaking
// existing senders because Message is passed by value, not by pointer,
// so new fields default to their zero value — existing callers compile
// and behave identically after the addition.
type Message struct {
	From string
	To   string
	Body any
}

// ============================================================
// Handler
// ============================================================

// Handler is the contract a service must satisfy to receive messages.
//
// A service may exist in the Registry without implementing Handler.
// Such services can send messages but cannot receive them.
// This separation keeps the Registry generic — it stores any Component —
// while the Bus enforces that receivers must actively opt in.
//
// The return type is (any, error) for the same reason Body is any:
// the Bus does not know or care what a handler produces. Callers
// type-assert the returned value to the concrete type they expect.
// In a future version this could be made generic (Handler[Req, Resp any])
// without changing the Bus's routing logic.
type Handler interface {
	Handle(msg Message) (any, error)
}

// ============================================================
// Validator
// ============================================================

// Validator is the contract the Boundary Engine must satisfy.
//
// The Bus calls Validate before forwarding any message. What constitutes
// a valid message — allowed body types, required fields, schema rules —
// is entirely the Boundary Engine's concern. The Bus only cares that the
// call succeeds or fails.
//
// Placing this interface here (rather than in a boundary package) keeps
// the dependency arrow correct: the Bus imports nothing from the boundary
// package. The boundary package will import this interface (or satisfy it
// implicitly via Go's structural typing) and implement it.
type Validator interface {
	Validate(msg Message) error
}

// ============================================================
// Authorizer
// ============================================================

// Authorizer is the contract the Permission Engine must satisfy.
//
// The Bus calls Authorize(caller, target) before forwarding any message.
// Whether a given caller may reach a given target — and under what
// conditions — is entirely the Permission Engine's concern.
//
// The same inward-dependency rule applies: the permission package will
// implement this interface; the Bus imports nothing from it.
type Authorizer interface {
	Authorize(caller, target string) error
}

// ============================================================
// Locator
// ============================================================

// Locator is the contract the Bus requires from the Service Registry.
//
// The Bus needs exactly one capability from the Registry: resolve a name
// to a Component. Depending on this narrow interface instead of the
// concrete *Registry type means:
//   - The Bus evolves independently of the Registry's internal structure.
//   - A future registry implementation (e.g. distributed, versioned) can
//     satisfy this interface without changing the Bus.
//   - Tests can supply a minimal stub without constructing a full Registry.
//
// *Registry satisfies Locator implicitly via its existing Lookup method.
// No change to registry.go is required.
type Locator interface {
	Lookup(name string) (Component, error)
}

// ============================================================
// Bus
// ============================================================

// Bus routes synchronous messages between registered services.
//
// All three dependencies are required at construction time.
// A Bus with a missing dependency must never be used — fail early.
type Bus struct {
	registry   Locator
	validator  Validator
	authorizer Authorizer
}

// NewBus constructs a Bus with all required dependencies.
//
// Returns an error if any dependency is nil. All three are mandatory:
//   - Without the registry, the Bus cannot locate targets.
//   - Without the validator, messages bypass the Boundary Engine entirely.
//   - Without the authorizer, messages bypass the Permission Engine entirely.
//
// Skipping validation or authorization is not a valid V1 configuration.
// A future "pass-through" implementation (that always returns nil) is the
// correct way to run without real validation — not a nil pointer.
func NewBus(registry Locator, validator Validator, authorizer Authorizer) (*Bus, error) {
	if registry == nil {
		return nil, errors.New("bus: NewBus failed — registry must not be nil")
	}
	if validator == nil {
		return nil, errors.New("bus: NewBus failed — validator must not be nil")
	}
	if authorizer == nil {
		return nil, errors.New("bus: NewBus failed — authorizer must not be nil")
	}

	return &Bus{
		registry:   registry,
		validator:  validator,
		authorizer: authorizer,
	}, nil
}

// ============================================================
// Component interface compliance
// ============================================================

// Name satisfies kernel.Component so the Kernel's Boot function can accept
// a *Bus in Config.Bus without any changes to kernel.go.
func (b *Bus) Name() string {
	return "CommunicationBus"
}

// ============================================================
// Send
// ============================================================

// Send routes msg to the service named by msg.To, synchronously.
//
// Steps (in order):
//  1. Boundary validation — the Validator checks message structure.
//  2. Authorization       — the Authorizer checks caller permissions.
//  3. Registry lookup     — locate the target service by name.
//  4. Handler check       — the target must implement Handler.
//  5. Delivery            — call the handler and return its result.
//
// All errors are wrapped with context so the caller can trace exactly
// which step failed. The %w verb preserves the original error for
// errors.Is and errors.As inspection.
func (b *Bus) Send(msg Message) (any, error) {
	// Step 1 — message structure validation (Boundary Engine responsibility).
	if err := b.validator.Validate(msg); err != nil {
		return nil, fmt.Errorf("bus: Send failed — validation: %w", err)
	}

	// Step 3 — authorization (Permission Engine responsibility).
	// Authorization happens before registry lookup so that the existence
	// of a service is not revealed to unauthorized callers.
	if err := b.authorizer.Authorize(msg.From, msg.To); err != nil {
		return nil, fmt.Errorf("bus: Send failed — authorization: %w", err)
	}

	// Step 4 — locate the target service.
	svc, err := b.registry.Lookup(msg.To)
	if err != nil {
		return nil, fmt.Errorf("bus: Send failed — %w", err)
	}

	// Step 5 — confirm the target can receive messages.
	// A service that does not implement Handler exists in the Registry
	// but has not opted in to message delivery. This is a caller error.
	handler, ok := svc.(Handler)
	if !ok {
		return nil, fmt.Errorf("bus: Send failed — %q is not a Handler", msg.To)
	}

	// Step 6 — deliver.
	return handler.Handle(msg)
}
