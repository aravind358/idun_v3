// Package kernel is the foundation of the IDUN system.
//
// The Kernel holds all five components together.
// Nothing in IDUN operates unless the Kernel is running.
//
// Kernel V1 contains five components:
//   - Boot Mechanism   (this file — lifecycle only)
//   - Service Registry (registers and resolves services)
//   - Communication Bus (routes messages)
//   - Boundary Engine  (validates messages)
//   - Permission Engine (authorizes callers)
//
// This file implements the Boot Mechanism only.
// The other four components are defined by the Component interface below.
// Their implementations will be added in separate files as V1 progresses.
package kernel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
)

// ============================================================
// Component Interface
// ============================================================

// Component is the contract every Kernel component must satisfy.
//
// In Go, an interface is a set of method signatures.
// Any type that implements all the methods automatically satisfies the interface.
// You do not declare "implements" — it is implicit.
//
// Every Kernel component must be able to report its own name.
// This serves two practical purposes:
//  1. Boot logs every component that is registered, making startup transparent.
//  2. When something fails, the name tells you exactly which component is involved.
//
// This is the minimal interface. It will not grow unless there is a strong reason.
type Component interface {
	// Name returns the human-readable name of the component.
	// Example: "ServiceRegistry", "CommunicationBus"
	Name() string
}

// ============================================================
// Kernel
// ============================================================

// Kernel is the central structure of the IDUN system.
//
// In Go, a struct is a collection of named fields.
// This struct holds one reference to each of the four runtime components.
// (The fifth component, Boot Mechanism, is the code in this file itself.)
//
// All fields are unexported (lowercase). This means only code inside
// the kernel package can read or write them directly.
// External code interacts with the Kernel only through its public methods.
// This is Go's way of enforcing encapsulation without access modifiers.
type Kernel struct {
	registry   Component
	bus        Component
	boundary   Component
	permission Component

	// running tracks whether the Kernel is active.
	// It starts true after a successful Boot and becomes false after Shutdown.
	running bool

	// started holds all Lifecycle-enabled components that successfully started,
	// ordered exactly by their startup sequence.
	started []Lifecycle
}

// ============================================================
// Config
// ============================================================

// Config carries the four components the Host must supply before booting.
//
// The Host (main.go) is responsible for constructing each component
// and passing them here. The Kernel does not create components itself.
//
// Why a Config struct instead of four function parameters?
//   - As V1 grows, adding a new required component is one field addition.
//   - The caller's code stays readable: cfg.Registry = ..., cfg.Bus = ...
//   - It is immediately obvious what the Kernel needs to start.
//
// All fields are exported (uppercase) because the Host lives in a different package
// and must be able to set them.
type Config struct {
	Registry   Component
	Bus        Component
	Boundary   Component
	Permission Component
}

// ============================================================
// Boot
// ============================================================

// Boot is the only way to create a Kernel.
//
// In Go, returning a pointer (*Kernel) and an error is the standard pattern
// for constructors that can fail. The caller must check the error before using
// the returned value. If err is not nil, the Kernel pointer will always be nil.
//
// Boot validates that every required component is present.
// If any component is missing, Boot returns an error immediately.
// A Kernel with missing components must never start — fail loudly, fail early.
//
// Boot does NOT:
//   - Start goroutines
//   - Open files or network connections
//   - Do any work beyond initialization and validation
//
// This keeps Boot fast, predictable, and easy to test.
func Boot(cfg Config) (*Kernel, error) {
	if cfg.Registry == nil {
		return nil, errors.New("kernel: Boot failed — Registry is required")
	}
	if cfg.Bus == nil {
		return nil, errors.New("kernel: Boot failed — Bus is required")
	}
	if cfg.Boundary == nil {
		return nil, errors.New("kernel: Boot failed — Boundary Engine is required")
	}
	if cfg.Permission == nil {
		return nil, errors.New("kernel: Boot failed — Permission Engine is required")
	}

	k := &Kernel{
		registry:   cfg.Registry,
		bus:        cfg.Bus,
		boundary:   cfg.Boundary,
		permission: cfg.Permission,
		running:    true,
	}

	// Discover all unique components across core infrastructure and registry.
	components := make(map[Component]bool)
	for _, c := range []Component{cfg.Registry, cfg.Bus, cfg.Boundary, cfg.Permission} {
		if c != nil {
			components[c] = true
		}
	}
	type serviceLister interface {
		List() []string
		Lookup(name string) (Component, error)
	}
	if lister, ok := cfg.Registry.(serviceLister); ok {
		for _, name := range lister.List() {
			if c, err := lister.Lookup(name); err == nil && c != nil {
				components[c] = true
			}
		}
	}

	// Group Lifecycle-enabled components by their topological BootPhase.
	phasedComponents := make(map[Phase][]Component)
	for c := range components {
		if _, isLifecycle := c.(Lifecycle); isLifecycle {
			phase := PhaseCognitive // Safe default phase for services
			if p, ok := c.(Phased); ok {
				phase = p.BootPhase()
			} else if c == cfg.Registry || c == cfg.Bus || c == cfg.Boundary || c == cfg.Permission {
				phase = PhaseCore
			}
			phasedComponents[phase] = append(phasedComponents[phase], c)
		}
	}

	// Execute startup across Phases 1 through 6 sequentially.
	ctx := context.Background()
	for p := PhaseCore; p <= PhaseBackground; p++ {
		list := phasedComponents[p]
		sort.Slice(list, func(i, j int) bool {
			return list[i].Name() < list[j].Name()
		})
		for _, c := range list {
			lc := c.(Lifecycle)
			log.Printf("[Kernel] Starting component: %s (Phase %d)", c.Name(), p)
			if err := lc.Start(ctx); err != nil {
				k.running = false
				for i := len(k.started) - 1; i >= 0; i-- {
					_ = k.started[i].Close()
				}
				k.started = nil
				return nil, fmt.Errorf("kernel: Boot failed starting component %q (phase %d): %w", c.Name(), p, err)
			}
			k.started = append(k.started, lc)
		}
	}

	// Log every registered component so startup is fully transparent.
	// The format is intentionally aligned for easy reading in a terminal.
	log.Println("[Kernel] Boot successful")
	log.Printf("[Kernel]   Registry  : %s", k.registry.Name())
	log.Printf("[Kernel]   Bus       : %s", k.bus.Name())
	log.Printf("[Kernel]   Boundary  : %s", k.boundary.Name())
	log.Printf("[Kernel]   Permission: %s", k.permission.Name())

	return k, nil
}

// ============================================================
// Shutdown
// ============================================================

// Shutdown stops the Kernel cleanly.
//
// After Shutdown, IsRunning returns false.
//
// Calling Shutdown on an already-stopped Kernel is safe.
// It returns immediately without doing anything.
// This is called "idempotent" behaviour — applying the operation
// more than once has the same effect as applying it once.
// Idempotent shutdown is important because real programs often
// call Shutdown in multiple places (signal handlers, deferred calls, etc.).
//
// In Go, methods are defined on a type using a receiver: (k *Kernel).
// The * means this method operates on a pointer to Kernel,
// so any changes it makes to k affect the actual Kernel in memory.
func (k *Kernel) Shutdown() {
	if !k.running {
		return
	}
	k.running = false
	for i := len(k.started) - 1; i >= 0; i-- {
		_ = k.started[i].Close()
	}
	k.started = nil
	log.Println("[Kernel] Shutdown complete")
}

// ============================================================
// IsRunning
// ============================================================

// IsRunning reports whether the Kernel is currently active.
//
// Returns true after a successful Boot.
// Returns false after Shutdown is called.
//
// Other components can use this to decide whether to accept new requests.
// In V1 this is a simple boolean field read.
// In future versions it could reflect richer lifecycle states (paused, draining, etc.)
// without changing the method signature — that is the extension point.
func (k *Kernel) IsRunning() bool {
	return k.running
}

// StartedCount returns the number of Lifecycle components that successfully started.
func (k *Kernel) StartedCount() int {
	return len(k.started)
}

// Registry returns the underlying service registry.
func (k *Kernel) Registry() Component {
	return k.registry
}
