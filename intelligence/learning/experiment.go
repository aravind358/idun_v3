package learning

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// DefaultExperimentManager is the concrete thread-safe implementation of ExperimentManager,
// responsible for tracking and bounding shadow and canary evaluations.
type DefaultExperimentManager struct {
	mu            sync.RWMutex
	active        map[string]*ExperimentProfile
	maxShadow     int
	maxCanary     int
	policyProvider StrategyProvider
}

// NewDefaultExperimentManager creates a new ExperimentManager with standard bounds.
func NewDefaultExperimentManager(provider StrategyProvider) *DefaultExperimentManager {
	return &DefaultExperimentManager{
		active:         make(map[string]*ExperimentProfile),
		maxShadow:      3,
		maxCanary:      1,
		policyProvider: provider,
	}
}

// StartExperiment initiates and registers a bounded shadow or canary experiment.
// It enforces concurrent cardinality limits and ensures isolation from active production routing.
func (m *DefaultExperimentManager) StartExperiment(ctx context.Context, profile *ExperimentProfile) error {
	if profile == nil {
		return fmt.Errorf("%w: experiment profile cannot be nil", ErrValidationFailed)
	}
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid experiment profile: %w", err)
	}
	if profile.ShadowRatio+profile.CanaryRatio > 1.0 {
		return fmt.Errorf("%w: combined shadow and canary ratio exceeds 1.0", ErrValidationFailed)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.active[profile.ExperimentID]; exists {
		return fmt.Errorf("%w: experiment %q already active", ErrValidationFailed, profile.ExperimentID)
	}

	// Dynamic check against governing policy limits if provider is present
	maxShadow := m.maxShadow
	maxCanary := m.maxCanary
	if m.policyProvider != nil {
		if snap := m.policyProvider.ActiveSnapshot(); snap != nil && snap.ActiveProfile != nil {
			if limit, ok := snap.ActiveProfile.ExperimentLimits["max_concurrent_shadow"]; ok && limit > 0 {
				maxShadow = limit
			}
			if limit, ok := snap.ActiveProfile.ExperimentLimits["max_concurrent_canary"]; ok && limit > 0 {
				maxCanary = limit
			}
		}
	}

	var currShadow, currCanary int
	for _, activeProf := range m.active {
		if activeProf.ShadowRatio > 0 {
			currShadow++
		}
		if activeProf.CanaryRatio > 0 {
			currCanary++
		}
	}

	if profile.ShadowRatio > 0 && currShadow >= maxShadow {
		return fmt.Errorf("%w: max concurrent shadow experiments limit (%d) exceeded", ErrCardinalityExceeded, maxShadow)
	}
	if profile.CanaryRatio > 0 && currCanary >= maxCanary {
		return fmt.Errorf("%w: max concurrent canary experiments limit (%d) exceeded", ErrCardinalityExceeded, maxCanary)
	}

	m.active[profile.ExperimentID] = profile
	return nil
}

// StopExperiment cleanly halts an active experiment and removes it from tracking.
func (m *DefaultExperimentManager) StopExperiment(ctx context.Context, experimentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.active[experimentID]; !exists {
		return fmt.Errorf("%w: experiment %q not found", ErrNotFound, experimentID)
	}
	delete(m.active, experimentID)
	return nil
}

// GetActiveExperiment retrieves an active experiment profile by ID without locking.
func (m *DefaultExperimentManager) GetActiveExperiment(ctx context.Context, experimentID string) (*ExperimentProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prof, exists := m.active[experimentID]
	if !exists {
		return nil, fmt.Errorf("%w: experiment %q not found", ErrNotFound, experimentID)
	}
	return prof, nil
}

// ListActiveExperiments returns a slice of all currently active experiments.
func (m *DefaultExperimentManager) ListActiveExperiments(ctx context.Context) []*ExperimentProfile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*ExperimentProfile, 0, len(m.active))
	for _, prof := range m.active {
		list = append(list, prof)
	}
	return list
}

// ListActiveExperimentsPrioritized returns active experiments sorted by descending Priority,
// breaking ties deterministically by combined evaluation ratio and ExperimentID.
func (m *DefaultExperimentManager) ListActiveExperimentsPrioritized(ctx context.Context) []*ExperimentProfile {
	list := m.ListActiveExperiments(ctx)
	sort.Slice(list, func(i, j int) bool {
		if list[i].Priority != list[j].Priority {
			return list[i].Priority > list[j].Priority
		}
		ratioI := list[i].ShadowRatio + list[i].CanaryRatio
		ratioJ := list[j].ShadowRatio + list[j].CanaryRatio
		if ratioI != ratioJ {
			return ratioI > ratioJ
		}
		return list[i].ExperimentID < list[j].ExperimentID
	})
	return list
}
