package attention

import (
	"sync/atomic"
)

// PolicySnapshotHolder provides lock-free atomic reads and updates for AttentionPolicyProfile.
type PolicySnapshotHolder struct {
	ptr atomic.Pointer[AttentionPolicyProfile]
}

// NewPolicySnapshotHolder constructs a new holder initialized with the given profile.
func NewPolicySnapshotHolder(initial *AttentionPolicyProfile) *PolicySnapshotHolder {
	h := &PolicySnapshotHolder{}
	if initial != nil {
		h.ptr.Store(initial)
	} else {
		h.ptr.Store(DefaultAttentionPolicyProfile())
	}
	return h
}

// Load atomically returns the current AttentionPolicyProfile snapshot.
func (h *PolicySnapshotHolder) Load() *AttentionPolicyProfile {
	return h.ptr.Load()
}

// Store atomically replaces the active AttentionPolicyProfile snapshot.
func (h *PolicySnapshotHolder) Store(profile *AttentionPolicyProfile) {
	if profile != nil {
		h.ptr.Store(profile)
	}
}

// CapabilitiesSnapshotHolder provides lock-free atomic reads and updates for AttentionCapabilities.
type CapabilitiesSnapshotHolder struct {
	ptr atomic.Pointer[AttentionCapabilities]
}

// NewCapabilitiesSnapshotHolder constructs a new holder initialized with the given capabilities.
func NewCapabilitiesSnapshotHolder(initial *AttentionCapabilities) *CapabilitiesSnapshotHolder {
	h := &CapabilitiesSnapshotHolder{}
	if initial != nil {
		h.ptr.Store(initial)
	} else {
		h.ptr.Store(DefaultAttentionCapabilities())
	}
	return h
}

// Load atomically returns the current AttentionCapabilities snapshot.
func (h *CapabilitiesSnapshotHolder) Load() *AttentionCapabilities {
	return h.ptr.Load()
}

// Store atomically replaces the active AttentionCapabilities snapshot.
func (h *CapabilitiesSnapshotHolder) Store(caps *AttentionCapabilities) {
	if caps != nil {
		h.ptr.Store(caps)
	}
}
