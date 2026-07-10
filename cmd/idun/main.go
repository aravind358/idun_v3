// Command idun is the Host that starts and stops the IDUN Kernel.
//
// The Host is the entry point for the IDUN system.
// It is responsible for three things only:
//  1. Assembling the Kernel configuration (which components to use)
//  2. Calling kernel.Boot to start the Kernel
//  3. Calling Shutdown when finished
//
// The Host does NOT implement any Kernel logic.
// The Host does NOT know how the Kernel works internally.
// The Host is a thin wiring layer — nothing more.
//
// In V1, the Host uses stub components because no real implementations exist yet.
// Each stub will be replaced by a real component as V1 progresses.
// The Host code will change only in the wiring section, not in its structure.
package main

import (
	"idun/kernel"
	"log"
)

// ============================================================
// Stub Components
// ============================================================
//
// Stubs are minimal implementations of the Component interface.
// They exist only to let the Boot Mechanism run and be verified
// before the real components are built.
//
// A stub is NOT a placeholder that does nothing forever.
// It is PROOF that the Component interface is correct and usable.
// When a real component is implemented, its stub is deleted from here.
//
// In Go, a struct with no fields is perfectly valid.
// It allocates no memory and satisfies interfaces purely through its methods.

// stubRegistry stands in for the real Service Registry.
// It will be replaced when kernel/registry.go is implemented.
type stubRegistry struct{}

func (s *stubRegistry) Name() string { return "StubRegistry" }

// stubBus stands in for the real Communication Bus.
// It will be replaced when kernel/bus.go is implemented.
type stubBus struct{}

func (s *stubBus) Name() string { return "StubBus" }

// stubBoundary stands in for the real Boundary Engine.
// It will be replaced when kernel/boundary.go is implemented.
type stubBoundary struct{}

func (s *stubBoundary) Name() string { return "StubBoundary" }

// stubPermission stands in for the real Permission Engine.
// It will be replaced when kernel/permission.go is implemented.
type stubPermission struct{}

func (s *stubPermission) Name() string { return "StubPermission" }

// ============================================================
// main
// ============================================================

func main() {
	log.Println("[Host] IDUN starting...")

	// Assemble the Kernel configuration.
	// The Host provides all components. The Kernel does not create them itself.
	// This is Dependency Injection: the Kernel depends on abstractions (Component),
	// not on concrete types. The Host decides which concrete types to supply.
	cfg := kernel.Config{
		Registry:   &stubRegistry{},
		Bus:        &stubBus{},
		Boundary:   &stubBoundary{},
		Permission: &stubPermission{},
	}

	// Boot the Kernel.
	// Boot returns a *Kernel and an error.
	// In Go, you must always check the error before using the returned value.
	// If err is not nil, k will be nil and using k would panic.
	k, err := kernel.Boot(cfg)
	if err != nil {
		// log.Fatalf prints the error and calls os.Exit(1).
		// This is the correct way to terminate when startup fails.
		log.Fatalf("[Host] Kernel boot failed: %v", err)
	}

	log.Printf("[Host] Kernel running: %v", k.IsRunning())

	// --- Future work happens here ---
	//
	// In the next steps, this section will:
	//   - Register a real Service into the Registry
	//   - Send a Message through the Bus
	//   - Receive a Response
	//
	// For now it is empty because the Boot Mechanism is all that exists.
	// Empty is correct. Do not add speculative code here.

	// Shut down the Kernel.
	// In production, Shutdown would be called in response to an OS signal (SIGINT, SIGTERM).
	// In V1, we call it directly after the work is done.
	k.Shutdown()

	log.Printf("[Host] Kernel running: %v", k.IsRunning())
	log.Println("[Host] IDUN stopped.")
}
