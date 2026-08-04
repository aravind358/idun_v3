package capabilities

import (
	"context"
	"errors"
	"fmt"
)

// PermissionManager gates capability execution based on authorization policies.
type PermissionManager interface {
	CheckPermission(ctx context.Context, category CapabilityCategory, requirementID string) error
}

type DefaultPermissionManager struct {
	// In a real system, this connects to the constitutional/security layers.
}

func NewPermissionManager() *DefaultPermissionManager {
	return &DefaultPermissionManager{}
}

func (p *DefaultPermissionManager) CheckPermission(ctx context.Context, category CapabilityCategory, requirementID string) error {
	// For V1 placeholder, we allow all except explicitly blocked categories.
	// V2+ will enforce strict user RBAC policies per Category.
	if category == "" {
		return errors.New("missing capability category")
	}
	return nil
}

// CapabilityLoader discovers and registers capability plugins/implementations.
type CapabilityLoader interface {
	Load(registry CapabilityRegistry) error
}

type DefaultLoader struct{}

func NewLoader() *DefaultLoader {
	return &DefaultLoader{}
}

func (l *DefaultLoader) Load(registry CapabilityRegistry) error {
	// In V1, this initializes the core native categories.
	return nil
}

// HealthMonitor periodically updates the OperationalStatus of capabilities.
type HealthMonitor interface {
	Start() error
	Stop() error
	Check(capabilityID string) (OperationalStatus, error)
}

type DefaultHealthMonitor struct {
	registry CapabilityRegistry
}

func NewHealthMonitor(registry CapabilityRegistry) *DefaultHealthMonitor {
	return &DefaultHealthMonitor{registry: registry}
}

func (h *DefaultHealthMonitor) Start() error { return nil }
func (h *DefaultHealthMonitor) Stop() error  { return nil }
func (h *DefaultHealthMonitor) Check(capabilityID string) (OperationalStatus, error) {
	cap, exists := h.registry.Get(capabilityID)
	if !exists {
		return StatusUnavailable, fmt.Errorf("capability not found")
	}
	return cap.State().Operational, nil
}

// DefaultResultNormalizer standardizes payloads.
type DefaultResultNormalizer struct{}

func NewResultNormalizer() *DefaultResultNormalizer {
	return &DefaultResultNormalizer{}
}

func (n *DefaultResultNormalizer) Normalize(raw interface{}) (CapabilityResult, error) {
	// Simplistic V1 normalization
	return CapabilityResult{
		Success:     true,
		Realization: Generative,
		Data:        map[string]interface{}{"raw": raw},
	}, nil
}
