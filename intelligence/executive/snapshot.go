package executive

import (
	"fmt"
	"sync/atomic"
)

// PolicySnapshotHolder provides lock-free, atomic reading and updating of
// immutable ExecutivePolicyProfile snapshots.
// Executive consumes snapshots via Load() without mutex contention during high-frequency cycles.
type PolicySnapshotHolder struct {
	ptr atomic.Pointer[ExecutivePolicyProfile]
}

// NewPolicySnapshotHolder initializes a snapshot holder with an initial policy profile.
// The initial profile is validated before storage.
func NewPolicySnapshotHolder(initial *ExecutivePolicyProfile) (*PolicySnapshotHolder, error) {
	if initial == nil {
		return nil, fmt.Errorf("%w: initial policy profile is nil", ErrInvalidPolicy)
	}
	if err := initial.Validate(); err != nil {
		return nil, fmt.Errorf("executive: failed to validate initial policy: %w", err)
	}
	h := &PolicySnapshotHolder{}
	h.ptr.Store(initial)
	return h, nil
}

// Load atomically retrieves the current immutable ExecutivePolicyProfile snapshot.
func (h *PolicySnapshotHolder) Load() *ExecutivePolicyProfile {
	return h.ptr.Load()
}

// Store validates and atomically swaps the active policy profile.
// Only validated profiles from Learning or SystemBoot may enter the snapshot holder.
func (h *PolicySnapshotHolder) Store(profile *ExecutivePolicyProfile) error {
	if profile == nil {
		return fmt.Errorf("%w: candidate policy profile is nil", ErrInvalidPolicy)
	}
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("executive: candidate policy failed validation: %w", err)
	}
	h.ptr.Store(profile)
	return nil
}

// CapabilitiesSnapshotHolder provides lock-free, atomic reading and updating of
// immutable ExecutiveCapabilities snapshots.
type CapabilitiesSnapshotHolder struct {
	ptr atomic.Pointer[ExecutiveCapabilities]
}

// NewCapabilitiesSnapshotHolder initializes a capabilities snapshot holder.
func NewCapabilitiesSnapshotHolder(initial *ExecutiveCapabilities) (*CapabilitiesSnapshotHolder, error) {
	if initial == nil {
		return nil, fmt.Errorf("%w: initial capabilities are nil", ErrInvalidCapabilities)
	}
	if err := initial.Validate(); err != nil {
		return nil, fmt.Errorf("executive: failed to validate initial capabilities: %w", err)
	}
	h := &CapabilitiesSnapshotHolder{}
	h.ptr.Store(initial)
	return h, nil
}

// Load atomically retrieves the current immutable ExecutiveCapabilities snapshot.
func (h *CapabilitiesSnapshotHolder) Load() *ExecutiveCapabilities {
	return h.ptr.Load()
}

// Store validates and atomically swaps the deployment capabilities snapshot.
func (h *CapabilitiesSnapshotHolder) Store(caps *ExecutiveCapabilities) error {
	if caps == nil {
		return fmt.Errorf("%w: candidate capabilities are nil", ErrInvalidCapabilities)
	}
	if err := caps.Validate(); err != nil {
		return fmt.Errorf("executive: candidate capabilities failed validation: %w", err)
	}
	h.ptr.Store(caps)
	return nil
}
