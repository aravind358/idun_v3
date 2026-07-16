package learning

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrLearnerAlreadyRegistered = errors.New("learning: learner already registered")
	ErrLearnerNotFound          = errors.New("learning: learner not found")
	ErrSnapshotNotFound         = errors.New("learning: snapshot not found")
)

// LearnerRegistry manages open signature-based registration (`Consumes()` / `Produces()`)
// of specialized learning algorithms without requiring core orchestration modifications.
type LearnerRegistry struct {
	mu         sync.RWMutex
	learners   map[string]Learner
	byConsumes map[string][]Learner
	byProduces map[string][]Learner
}

// NewLearnerRegistry initializes an empty LearnerRegistry skeleton.
func NewLearnerRegistry() *LearnerRegistry {
	return &LearnerRegistry{
		learners:   make(map[string]Learner),
		byConsumes: make(map[string][]Learner),
		byProduces: make(map[string][]Learner),
	}
}

// NewDefaultLearnerRegistry creates a LearnerRegistry populated with the default
// domain, statistical, and cross-domain synthesis learners.
func NewDefaultLearnerRegistry() *LearnerRegistry {
	reg := NewLearnerRegistry()
	_ = reg.Register(NewReasoningLearner())
	_ = reg.Register(NewPlanningLearner())
	_ = reg.Register(NewDecisionLearner())
	_ = reg.Register(NewThresholdOptimizationEngine())
	_ = reg.Register(NewWeightOptimizationEngine())
	_ = reg.Register(NewCalibrationOptimizationEngine())
	_ = reg.Register(NewConfidenceOptimizationEngine())
	_ = reg.Register(NewPreferenceOptimizationEngine())
	_ = reg.Register(NewCrossDomainLearner())
	return reg
}

// Register registers a Learner by its unique ID and signature schemas.
func (r *LearnerRegistry) Register(learner Learner) error {
	if learner == nil {
		return fmt.Errorf("%w: cannot register nil learner", ErrValidationFailed)
	}
	id := learner.LearnerID()
	if id == "" {
		return fmt.Errorf("%w: learner ID cannot be empty", ErrMissingID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.learners[id]; exists {
		return fmt.Errorf("%w: %q", ErrLearnerAlreadyRegistered, id)
	}

	r.learners[id] = learner
	for _, consumeID := range learner.Consumes() {
		r.byConsumes[consumeID] = append(r.byConsumes[consumeID], learner)
	}
	for _, produceID := range learner.Produces() {
		r.byProduces[produceID] = append(r.byProduces[produceID], learner)
	}
	return nil
}

// Get returns a registered Learner by its ID.
func (r *LearnerRegistry) Get(learnerID string) (Learner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	l, ok := r.learners[learnerID]
	return l, ok
}

// LookupByConsumes returns all registered learners that consume the given artifact schema ID.
func (r *LearnerRegistry) LookupByConsumes(artifactSchemaID string) []Learner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.byConsumes[artifactSchemaID]
	out := make([]Learner, len(list))
	copy(out, list)
	return out
}

// LookupByProduces returns all registered learners that synthesize the given snapshot schema ID.
func (r *LearnerRegistry) LookupByProduces(snapshotSchemaID string) []Learner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.byProduces[snapshotSchemaID]
	out := make([]Learner, len(list))
	copy(out, list)
	return out
}

// ListLearners returns a snapshot slice of all currently registered learners.
func (r *LearnerRegistry) ListLearners() []Learner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Learner, 0, len(r.learners))
	for _, l := range r.learners {
		out = append(out, l)
	}
	return out
}

// DefaultSnapshotRegistry implements SnapshotRegistry to store and pointer-switch immutable candidates.
type DefaultSnapshotRegistry struct {
	mu      sync.RWMutex
	active  map[string]*CandidateSnapshot   // Keyed by SchemaID
	history map[string][]*CandidateSnapshot // Keyed by SchemaID
}

// NewDefaultSnapshotRegistry constructs a new DefaultSnapshotRegistry.
func NewDefaultSnapshotRegistry() *DefaultSnapshotRegistry {
	return &DefaultSnapshotRegistry{
		active:  make(map[string]*CandidateSnapshot),
		history: make(map[string][]*CandidateSnapshot),
	}
}

// Publish registers and stores an active candidate snapshot.
func (sr *DefaultSnapshotRegistry) Publish(ctx context.Context, candidate *CandidateSnapshot) error {
	if candidate == nil {
		return fmt.Errorf("%w: cannot publish nil candidate", ErrValidationFailed)
	}
	if err := candidate.Validate(); err != nil {
		return fmt.Errorf("candidate validation failed: %w", err)
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	schemaID := candidate.SchemaID
	if current, exists := sr.active[schemaID]; exists {
		sr.history[schemaID] = append(sr.history[schemaID], current)
	}
	sr.active[schemaID] = candidate
	return nil
}

// GetActive retrieves the active CandidateSnapshot for the specified domain schema.
func (sr *DefaultSnapshotRegistry) GetActive(ctx context.Context, schemaID string) (*CandidateSnapshot, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	cand, exists := sr.active[schemaID]
	if !exists {
		return nil, fmt.Errorf("%w: no active snapshot for schema %q", ErrSnapshotNotFound, schemaID)
	}
	return cand, nil
}

// Rollback atomically reverts the active candidate snapshot pointer for schemaID to targetVersion.
func (sr *DefaultSnapshotRegistry) Rollback(ctx context.Context, schemaID string, targetVersion string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	history, exists := sr.history[schemaID]
	if !exists || len(history) == 0 {
		return fmt.Errorf("%w: no rollback history for schema %q", ErrSnapshotNotFound, schemaID)
	}

	for i := len(history) - 1; i >= 0; i-- {
		cand := history[i]
		if cand.SemVer == targetVersion || cand.SnapshotID == targetVersion {
			if current, activeExists := sr.active[schemaID]; activeExists {
				sr.history[schemaID] = append(sr.history[schemaID], current)
			}
			sr.active[schemaID] = cand
			return nil
		}
	}
	return fmt.Errorf("%w: version %q not found in history for schema %q", ErrSnapshotNotFound, targetVersion, schemaID)
}
