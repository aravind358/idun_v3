// This file implements the Boundary Engine — Kernel component five of five.
//
// Responsibility: validate message structure before routing.
//
// The Boundary Engine is the sole owner of message structure validation
// in Kernel Version 1. It checks that a Message is well-formed so that
// the Communication Bus can act as a pure orchestrator.
//
// What this file intentionally does NOT do in Version 1:
//   - Schema or payload type validation (V2+)
//   - Size limits or rate limits (V2+)
//   - Wildcard matching or configuration loading
//
// The engine satisfies two interfaces implicitly:
//   - kernel.Component (via Name()) — required by kernel.Boot
//   - kernel.Validator (via Validate()) — required by the Bus
package kernel

import "errors"

// ============================================================
// BoundaryEngine
// ============================================================

// BoundaryEngine is a synchronous, stateless message structure validator.
type BoundaryEngine struct{}

// NewBoundaryEngine constructs a BoundaryEngine ready for message validation.
func NewBoundaryEngine() *BoundaryEngine {
	return &BoundaryEngine{}
}

// ============================================================
// Component interface compliance
// ============================================================

// Name satisfies kernel.Component so the Kernel's Boot function can accept
// a *BoundaryEngine in Config.Boundary without any changes to kernel.go.
func (b *BoundaryEngine) Name() string {
	return "BoundaryEngine"
}

// ============================================================
// Validator interface compliance
// ============================================================

// Validate checks whether msg is structurally valid.
//
// In Version 1, a message is structurally valid if and only if both its
// From and To fields are non-empty strings.
//
// Returns an error describing the structural violation if validation fails,
// or nil if the message is valid.
func (b *BoundaryEngine) Validate(msg Message) error {
	if msg.From == "" {
		return errors.New("boundary: Validate failed — From must not be empty")
	}
	if msg.To == "" {
		return errors.New("boundary: Validate failed — To must not be empty")
	}
	return nil
}
