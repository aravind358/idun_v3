// This file implements the Service Registry — Kernel component two of five.
//
// Responsibility: let callers register, deregister, look up, and list
// services by name at runtime.
//
// What this file intentionally does NOT do:
//   - Health checking (not a registry concern)
//   - Permission filtering (that is the Permission Engine's job)
//   - Concurrency safety (V1 is synchronous; a mutex can be added later
//     without changing any public signature)
//   - Dependency injection or lifecycle management
//
// The Registry satisfies the Component interface so the Kernel's Boot
// function can accept it without modification.
package kernel

import (
	"errors"
	"fmt"
	"sort"
)

// ============================================================
// Registry
// ============================================================

// Registry is a synchronous, in-memory service store.
//
// In Go, unexported fields cannot be accessed from outside the package.
// The map is the only state this type holds; there are no hidden caches,
// counters, or auxiliary structures.
//
// Do not construct a Registry with a struct literal ({}).
// Always use NewRegistry so the internal map is initialised.
// An uninitialised map panics on write — NewRegistry prevents that class
// of bug by construction.
type Registry struct {
	services map[string]Component
}

// NewRegistry constructs an empty Registry.
//
// The naming convention "New<Type>" is idiomatic Go for constructors.
// It returns a pointer so that Register, Deregister, and their callers
// all operate on the same underlying map.
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]Component),
	}
}

// ============================================================
// Component interface compliance
// ============================================================

// Name returns the canonical identifier of this component.
//
// This method makes *Registry satisfy the kernel.Component interface,
// which is the contract the Kernel's Boot function requires.
// The name is fixed — it is not configurable — because the Registry
// is a singular, well-defined kernel component, not a generic data store.
func (r *Registry) Name() string {
	return "ServiceRegistry"
}

// ============================================================
// Register
// ============================================================

// Register adds svc to the registry under svc.Name().
//
// Validation rules:
//   - svc must not be nil.
//   - svc.Name() must not be empty.
//   - No service with the same name may already be registered.
//
// These three rules cover every foreseeable invalid input without
// introducing speculative validation that V1 does not need.
//
// Returning an error (rather than silently overwriting) is deliberate.
// Silent overwrites hide bugs; an explicit error forces the caller
// to acknowledge the conflict and decide what to do.
func (r *Registry) Register(svc Component) error {
	if svc == nil {
		return errors.New("registry: Register failed — service must not be nil")
	}

	name := svc.Name()
	if name == "" {
		return errors.New("registry: Register failed — service name must not be empty")
	}

	if _, exists := r.services[name]; exists {
		return fmt.Errorf("registry: Register failed — %q is already registered", name)
	}

	r.services[name] = svc
	return nil
}

// ============================================================
// Deregister
// ============================================================

// Deregister removes the service registered under name.
//
// Validation rules:
//   - name must not be empty.
//   - A service with that name must currently be registered.
//
// Returning an error for a missing name (rather than silently doing
// nothing) is intentional. Silent no-ops hide bugs in callers that
// assume a service was registered when it was not.
func (r *Registry) Deregister(name string) error {
	if name == "" {
		return errors.New("registry: Deregister failed — name must not be empty")
	}

	if _, exists := r.services[name]; !exists {
		return fmt.Errorf("registry: Deregister failed — %q is not registered", name)
	}

	delete(r.services, name)
	return nil
}

// ============================================================
// Lookup
// ============================================================

// Lookup returns the service registered under name.
//
// Validation rules:
//   - name must not be empty.
//   - A service with that name must currently be registered.
//
// The caller receives the Component interface, not a concrete type.
// If the caller needs the underlying concrete type it uses a type
// assertion: svc.(*MyService). This is the standard Go pattern and
// requires no generics in V1.
func (r *Registry) Lookup(name string) (Component, error) {
	if name == "" {
		return nil, errors.New("registry: Lookup failed — name must not be empty")
	}

	svc, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("registry: Lookup failed — %q is not registered", name)
	}

	return svc, nil
}

// ============================================================
// List
// ============================================================

// List returns the names of all currently registered services,
// sorted alphabetically.
//
// Why sort? Map iteration in Go is intentionally randomised to prevent
// callers from depending on insertion order. Sorting makes the output
// deterministic, which matters for logging, diagnostics, and tests.
//
// List never returns nil. An empty registry returns an empty slice,
// not nil. Callers should not need to nil-check a slice returned from
// a well-designed function.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
