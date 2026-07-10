// This file implements the Permission Engine — Kernel component four of five.
//
// Responsibility: decide whether a named caller may send a message to a
// named target.
//
// The Permission Engine answers exactly one question at runtime:
// "Is caller permitted to reach target?" It does nothing else.
//
// What this file intentionally does NOT do:
//   - Authenticate callers (verifying identity is out of scope for V1)
//   - Inspect message content (that is the Boundary Engine's job)
//   - Support wildcards, roles, or groups (V2+ features, no current caller)
//   - Provide Revoke() (no V1 runtime caller; will be added in V2 when needed)
//   - Persist permissions across restarts (in-memory only, V1)
//   - Use sync.RWMutex (V1 is synchronous, consistent with Registry)
//
// Default policy: DENY.
// A freshly constructed PermissionEngine blocks all communication.
// Callers must explicitly call Allow() for every permitted path.
// This is the correct security posture — start closed, open deliberately.
//
// In V1, Allow() is called by the Host (main.go) during the wiring phase,
// before kernel.Boot(). No component calls Allow() at runtime.
//
// The engine satisfies two interfaces implicitly:
//   - kernel.Component  (via Name()) — required by kernel.Boot
//   - kernel.Authorizer (via Authorize()) — required by the Bus
package kernel

import (
	"errors"
	"fmt"
)

// ============================================================
// PermissionEngine
// ============================================================

// PermissionEngine is a synchronous, in-memory permission store.
//
// rules is a nested map: outer key is the caller name, inner key is the
// target name, value is an empty struct (Go's zero-allocation set member).
//
//	rules["IntelligenceCore"]["MemoryService"] = struct{}{}
//	// means: IntelligenceCore is permitted to send to MemoryService
//
// Why a nested map rather than a flat map with a composite key?
// A composite key (e.g. "caller:target") requires choosing a separator
// that is guaranteed not to appear in service names. A nested map has no
// such requirement and is self-documenting.
//
// Do not construct a PermissionEngine with a struct literal.
// Always use NewPermissionEngine so the internal map is initialised.
type PermissionEngine struct {
	rules map[string]map[string]struct{}
}

// NewPermissionEngine constructs a PermissionEngine with no permissions.
//
// The engine starts in a deny-all state. Every Authorize call will fail
// until Allow is called for the specific (caller, target) pair.
func NewPermissionEngine() *PermissionEngine {
	return &PermissionEngine{
		rules: make(map[string]map[string]struct{}),
	}
}

// ============================================================
// Component interface compliance
// ============================================================

// Name satisfies kernel.Component so the Kernel's Boot function can accept
// a *PermissionEngine in Config.Permission without any changes to kernel.go.
func (p *PermissionEngine) Name() string {
	return "PermissionEngine"
}

// ============================================================
// Allow
// ============================================================

// Allow registers a permission for caller to send messages to target.
//
// Validation rules:
//   - caller must not be empty.
//   - target must not be empty.
//   - The (caller, target) pair must not already be permitted.
//
// Returning an error for a duplicate (rather than silently succeeding) is
// deliberate and consistent with the Registry's Register behaviour.
// If the Host registers the same permission twice, that is likely a bug
// in the wiring code — it should be surfaced, not hidden.
//
// In V1, Allow is called by the Host during the wiring phase before Boot.
// It is not called at runtime by any Kernel component.
func (p *PermissionEngine) Allow(caller, target string) error {
	if caller == "" {
		return errors.New("permission: Allow failed — caller must not be empty")
	}
	if target == "" {
		return errors.New("permission: Allow failed — target must not be empty")
	}

	// Check for duplicate before creating the inner map.
	if targets, exists := p.rules[caller]; exists {
		if _, already := targets[target]; already {
			return fmt.Errorf("permission: Allow failed — %q → %q is already permitted", caller, target)
		}
	} else {
		// Lazily initialise the inner map only when the caller is seen for
		// the first time. This avoids allocating maps for callers that are
		// never granted any permission.
		p.rules[caller] = make(map[string]struct{})
	}

	p.rules[caller][target] = struct{}{}
	return nil
}

// ============================================================
// Authorize
// ============================================================

// Authorize returns nil if caller is permitted to send to target.
// Returns an error if the permission has not been granted.
//
// Validation rules:
//   - caller must not be empty.
//   - target must not be empty.
//
// This is the method the Bus calls on every Send, after content validation
// and before the registry lookup. The Bus treats any non-nil error as a
// denial — the error message is forwarded to the caller as-is.
//
// The error message deliberately does not distinguish between
// "caller has no permissions at all" and "caller cannot reach this specific
// target". Both are represented as a single denial. This prevents callers
// from inferring the permission topology by probing different targets.
func (p *PermissionEngine) Authorize(caller, target string) error {
	if caller == "" {
		return errors.New("permission: Authorize failed — caller must not be empty")
	}
	if target == "" {
		return errors.New("permission: Authorize failed — target must not be empty")
	}

	targets, callerKnown := p.rules[caller]
	if !callerKnown {
		return fmt.Errorf("permission: Authorize failed — %q is not permitted to send to %q", caller, target)
	}

	if _, ok := targets[target]; !ok {
		return fmt.Errorf("permission: Authorize failed — %q is not permitted to send to %q", caller, target)
	}

	return nil
}
