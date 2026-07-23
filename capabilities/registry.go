package capabilities

import (
	"errors"
	"sync"
)

// DefaultRegistry provides a thread-safe implementation of CapabilityRegistry.
type DefaultRegistry struct {
	mu           sync.RWMutex
	capabilities map[string]Capability
}

func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		capabilities: make(map[string]Capability),
	}
}

func (r *DefaultRegistry) Register(cap Capability) error {
	if cap == nil {
		return errors.New("cannot register nil capability")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	id := cap.ID()
	if _, exists := r.capabilities[id]; exists {
		return errors.New("capability already registered: " + id)
	}
	r.capabilities[id] = cap
	return nil
}

func (r *DefaultRegistry) Deregister(capabilityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.capabilities[capabilityID]; !exists {
		return errors.New("capability not found: " + capabilityID)
	}
	delete(r.capabilities, capabilityID)
	return nil
}

func (r *DefaultRegistry) Get(capabilityID string) (Capability, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cap, exists := r.capabilities[capabilityID]
	return cap, exists
}

func (r *DefaultRegistry) List() []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Capability, 0, len(r.capabilities))
	for _, cap := range r.capabilities {
		list = append(list, cap)
	}
	return list
}

func (r *DefaultRegistry) FindByName(name string) []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []Capability
	for _, cap := range r.capabilities {
		if cap.Metadata().Name == name {
			list = append(list, cap)
		}
	}
	return list
}
