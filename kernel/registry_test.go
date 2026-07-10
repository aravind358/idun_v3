// Tests for the Service Registry.
//
// All tests live in package kernel (white-box) so they share the same
// package as the implementation and can exercise it without any exported
// constructor tricks. This mirrors the style of kernel_test.go.
//
// Test naming convention: Test<Type>_<Scenario>
//   - Happy paths use "Success" or a short description.
//   - Error paths describe the invalid condition.
//
// Table-driven tests are used wherever multiple similar inputs produce
// similar outcomes. Standalone tests are used for behaviours that are
// meaningfully distinct enough to deserve their own name.
package kernel

import (
	"testing"
)

// ============================================================
// Helpers
// ============================================================

// newRegistry returns a fresh, empty Registry for use in tests.
// It is a thin wrapper around NewRegistry that keeps test bodies short.
func newRegistry() *Registry {
	return NewRegistry()
}

// ============================================================
// NewRegistry
// ============================================================

// TestNewRegistry verifies that NewRegistry returns a non-nil, usable Registry.
func TestNewRegistry(t *testing.T) {
	r := NewRegistry()

	if r == nil {
		t.Fatal("NewRegistry() returned nil, expected a valid *Registry")
	}

	// An empty registry must return an empty, non-nil slice from List.
	names := r.List()
	if names == nil {
		t.Error("List() on a new Registry returned nil, expected an empty slice")
	}
	if len(names) != 0 {
		t.Errorf("List() on a new Registry returned %d entries, expected 0", len(names))
	}
}

// ============================================================
// Name
// ============================================================

// TestRegistry_Name verifies that the Registry reports the correct component name.
// This also confirms it satisfies the Component interface at compile time.
func TestRegistry_Name(t *testing.T) {
	r := newRegistry()

	// Compile-time check: *Registry must satisfy Component.
	// This assignment will not compile if the interface is not satisfied.
	var _ Component = r

	got := r.Name()
	want := "ServiceRegistry"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// ============================================================
// Register
// ============================================================

// TestRegister_Success verifies that a valid service registers without error.
func TestRegister_Success(t *testing.T) {
	r := newRegistry()
	svc := newStub("Alpha")

	if err := r.Register(svc); err != nil {
		t.Fatalf("Register() returned unexpected error: %v", err)
	}

	// Confirm the service is now visible via List.
	names := r.List()
	if len(names) != 1 || names[0] != "Alpha" {
		t.Errorf("after Register, List() = %v, want [Alpha]", names)
	}
}

// TestRegister_InvalidInputs verifies that Register rejects every class of bad input.
func TestRegister_InvalidInputs(t *testing.T) {
	cases := []struct {
		name string
		svc  Component
	}{
		{
			name: "nil service",
			svc:  nil,
		},
		{
			// emptyNameComponent is defined at package level below;
			// it returns "" from Name() to exercise the empty-name guard.
			name: "empty service name",
			svc:  &emptyNameComponent{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newRegistry()
			err := r.Register(tc.svc)
			if err == nil {
				t.Errorf("Register(%v) should have returned an error", tc.svc)
			}
		})
	}
}

// emptyNameComponent is a Component that intentionally returns an empty name.
// It is defined at package level so any test in this file can reference it.
type emptyNameComponent struct{}

func (e *emptyNameComponent) Name() string { return ""}

// TestRegister_Duplicate verifies that registering the same name twice returns an error.
func TestRegister_Duplicate(t *testing.T) {
	r := newRegistry()

	first := newStub("Bravo")
	second := newStub("Bravo") // same name, different instance

	if err := r.Register(first); err != nil {
		t.Fatalf("first Register() returned unexpected error: %v", err)
	}

	err := r.Register(second)
	if err == nil {
		t.Error("second Register() with duplicate name should have returned an error")
	}

	// Registry must still contain exactly one entry.
	names := r.List()
	if len(names) != 1 {
		t.Errorf("after duplicate Register, List() returned %d entries, want 1", len(names))
	}
}

// ============================================================
// Deregister
// ============================================================

// TestDeregister_Success verifies that a registered service can be removed.
func TestDeregister_Success(t *testing.T) {
	r := newRegistry()
	svc := newStub("Charlie")

	_ = r.Register(svc)

	if err := r.Deregister("Charlie"); err != nil {
		t.Fatalf("Deregister() returned unexpected error: %v", err)
	}

	// The service must no longer appear in List.
	if len(r.List()) != 0 {
		t.Error("after Deregister, List() should be empty")
	}

	// Lookup must fail after deregistration.
	if _, err := r.Lookup("Charlie"); err == nil {
		t.Error("Lookup() should fail after Deregister")
	}
}

// TestDeregister_InvalidInputs verifies that Deregister rejects bad inputs.
func TestDeregister_InvalidInputs(t *testing.T) {
	cases := []struct {
		name    string
		regName string // name to attempt to deregister
	}{
		{
			name:    "empty name",
			regName: "",
		},
		{
			name:    "name not registered",
			regName: "Ghost",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newRegistry()
			// "Ghost" is never registered, so both cases test missing names.
			err := r.Deregister(tc.regName)
			if err == nil {
				t.Errorf("Deregister(%q) should have returned an error", tc.regName)
			}
		})
	}
}

// ============================================================
// Lookup
// ============================================================

// TestLookup_Success verifies that a registered service can be retrieved.
func TestLookup_Success(t *testing.T) {
	r := newRegistry()
	original := newStub("Delta")
	_ = r.Register(original)

	found, err := r.Lookup("Delta")
	if err != nil {
		t.Fatalf("Lookup() returned unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("Lookup() returned nil service, expected a valid Component")
	}
	if found.Name() != "Delta" {
		t.Errorf("Lookup() returned service with name %q, want %q", found.Name(), "Delta")
	}
	// The returned value must be the exact instance that was registered.
	if found != original {
		t.Error("Lookup() returned a different instance than the one registered")
	}
}

// TestLookup_InvalidInputs verifies that Lookup rejects bad inputs.
func TestLookup_InvalidInputs(t *testing.T) {
	cases := []struct {
		name       string
		lookupName string
	}{
		{
			name:       "empty name",
			lookupName: "",
		},
		{
			name:       "name not registered",
			lookupName: "Echo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newRegistry()
			svc, err := r.Lookup(tc.lookupName)
			if err == nil {
				t.Errorf("Lookup(%q) should have returned an error", tc.lookupName)
			}
			if svc != nil {
				t.Errorf("Lookup(%q) on error should return nil service, got %v", tc.lookupName, svc)
			}
		})
	}
}

// ============================================================
// List
// ============================================================

// TestList_Empty verifies that List returns an empty, non-nil slice when
// nothing is registered.
func TestList_Empty(t *testing.T) {
	r := newRegistry()
	names := r.List()

	if names == nil {
		t.Error("List() returned nil on empty registry, expected an empty slice")
	}
	if len(names) != 0 {
		t.Errorf("List() returned %d entries on empty registry, expected 0", len(names))
	}
}

// TestList_Single verifies that List works correctly with one registered service.
func TestList_Single(t *testing.T) {
	r := newRegistry()
	_ = r.Register(newStub("Foxtrot"))

	names := r.List()
	if len(names) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(names))
	}
	if names[0] != "Foxtrot" {
		t.Errorf("List()[0] = %q, want %q", names[0], "Foxtrot")
	}
}

// TestList_Sorted verifies that List returns names in alphabetical order
// regardless of the registration order.
func TestList_Sorted(t *testing.T) {
	r := newRegistry()

	// Register in reverse alphabetical order.
	_ = r.Register(newStub("Zulu"))
	_ = r.Register(newStub("Alpha"))
	_ = r.Register(newStub("Mike"))

	names := r.List()
	want := []string{"Alpha", "Mike", "Zulu"}

	if len(names) != len(want) {
		t.Fatalf("List() returned %d entries, want %d", len(names), len(want))
	}
	for i, name := range names {
		if name != want[i] {
			t.Errorf("List()[%d] = %q, want %q", i, name, want[i])
		}
	}
}

// TestList_AfterDeregister verifies that List reflects deregistration immediately.
func TestList_AfterDeregister(t *testing.T) {
	r := newRegistry()

	_ = r.Register(newStub("November"))
	_ = r.Register(newStub("Oscar"))
	_ = r.Deregister("November")

	names := r.List()
	if len(names) != 1 || names[0] != "Oscar" {
		t.Errorf("after Deregister, List() = %v, want [Oscar]", names)
	}
}

// ============================================================
// Register → Deregister → Register cycle
// ============================================================

// TestRegister_AfterDeregister verifies that a name can be reused after
// its previous service has been deregistered.
//
// This matters because the duplicate check must look at what is currently
// registered, not at historical registrations.
func TestRegister_AfterDeregister(t *testing.T) {
	r := newRegistry()

	first := newStub("Papa")
	_ = r.Register(first)
	_ = r.Deregister("Papa")

	second := newStub("Papa")
	if err := r.Register(second); err != nil {
		t.Fatalf("Register() after Deregister returned unexpected error: %v", err)
	}

	found, err := r.Lookup("Papa")
	if err != nil {
		t.Fatalf("Lookup() after re-registration returned error: %v", err)
	}
	// The registry must hold the new instance, not the old one.
	if found != second {
		t.Error("Lookup() returned the old instance; expected the newly registered one")
	}
}
