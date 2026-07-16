package executive

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	maxRollingHistory = 16
)

// Sentinel errors for EpisodeManager operations.
var (
	ErrEpisodeNotFound      = errors.New("episode_manager: episode not found")
	ErrEpisodeAlreadyExists = errors.New("episode_manager: episode already exists")
	ErrInvalidStatusChange  = errors.New("episode_manager: illegal FSM transition requested")
	ErrTerminalEpisode      = errors.New("episode_manager: cannot transition out of terminal status")
	ErrCheckpointNotFound   = errors.New("episode_manager: checkpoint not found")
)

// EpisodeManager manages episode storage, lifecycle transitions, rolling histories, and checkpoint recovery.
// Executive remains the content-blind coordinator; EpisodeManager encapsulates episode state transitions.
type EpisodeManager interface {
	// CreateEpisode registers a newly built and validated episode.
	CreateEpisode(ep *ExecutiveEpisode) error

	// GetEpisode retrieves an episode by its unique EpisodeID.
	GetEpisode(id EpisodeID) (*ExecutiveEpisode, bool)

	// ListEpisodes returns all active or archived episodes matching an optional status filter.
	ListEpisodes(filterStatus *EpisodeStatus) []*ExecutiveEpisode

	// TransitionStatus validates and executes a lifecycle state change on an episode.
	TransitionStatus(id EpisodeID, toStatus EpisodeStatus, outcome EpisodeOutcome, pauseReason PauseReason, resumeReason ResumeReason) error

	// UpdatePriority records a factual priority change into the bounded rolling history.
	UpdatePriority(id EpisodeID, newPriority PriorityBand, reason PriorityTransitionReason) error

	// UpdateBudget records a factual budget tier change into the bounded rolling history.
	UpdateBudget(id EpisodeID, newBudget BudgetTier, reason BudgetTransitionReason) error

	// CreateCheckpoint generates, validates, and stores an immutable EpisodeCheckpoint snapshot.
	CreateCheckpoint(id EpisodeID, reason CheckpointReason) (*EpisodeCheckpoint, error)

	// GetCheckpoint retrieves an immutable checkpoint by ID.
	GetCheckpoint(checkpointID string) (*EpisodeCheckpoint, bool)

	// RestoreFromCheckpoint restores runtime execution consistently from an immutable checkpoint and definition.
	RestoreFromCheckpoint(cp *EpisodeCheckpoint, def *ExecutiveEpisodeDefinition) (*ExecutiveEpisode, error)
}

type defaultEpisodeManager struct {
	mu          sync.RWMutex
	episodes    map[EpisodeID]*ExecutiveEpisode
	checkpoints map[string]*EpisodeCheckpoint
}

// NewEpisodeManager creates a thread-safe defaultEpisodeManager.
func NewEpisodeManager() EpisodeManager {
	return &defaultEpisodeManager{
		episodes:    make(map[EpisodeID]*ExecutiveEpisode),
		checkpoints: make(map[string]*EpisodeCheckpoint),
	}
}

func (m *defaultEpisodeManager) CreateEpisode(ep *ExecutiveEpisode) error {
	if ep == nil {
		return errors.New("episode_manager: cannot create nil episode")
	}
	if err := ep.Validate(); err != nil {
		return fmt.Errorf("episode_manager: validation failed before creation: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.episodes[ep.Definition.EpisodeID]; exists {
		return ErrEpisodeAlreadyExists
	}
	m.episodes[ep.Definition.EpisodeID] = ep
	return nil
}

func (m *defaultEpisodeManager) GetEpisode(id EpisodeID) (*ExecutiveEpisode, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ep, exists := m.episodes[id]
	return ep, exists
}

func (m *defaultEpisodeManager) ListEpisodes(filterStatus *EpisodeStatus) []*ExecutiveEpisode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*ExecutiveEpisode, 0, len(m.episodes))
	for _, ep := range m.episodes {
		if filterStatus != nil && ep.Runtime.Status != *filterStatus {
			continue
		}
		out = append(out, ep)
	}
	return out
}

// ValidateTransition checks whether moving from -> to is a valid FSM transition.
func ValidateTransition(from, to EpisodeStatus) error {
	if from == to {
		return nil
	}
	switch from {
	case EpisodeStatusCreated:
		if to == EpisodeStatusWaiting || to == EpisodeStatusRunning || to == EpisodeStatusCancelled {
			return nil
		}
	case EpisodeStatusWaiting:
		if to == EpisodeStatusRunning || to == EpisodeStatusPaused || to == EpisodeStatusCancelled {
			return nil
		}
	case EpisodeStatusRunning:
		if to == EpisodeStatusWaiting || to == EpisodeStatusPaused || to == EpisodeStatusSuspended || to == EpisodeStatusCompleted || to == EpisodeStatusFailed || to == EpisodeStatusCancelled {
			return nil
		}
	case EpisodeStatusPaused:
		if to == EpisodeStatusRunning || to == EpisodeStatusWaiting || to == EpisodeStatusSuspended || to == EpisodeStatusCancelled {
			return nil
		}
	case EpisodeStatusSuspended:
		if to == EpisodeStatusRunning || to == EpisodeStatusWaiting || to == EpisodeStatusCancelled {
			return nil
		}
	case EpisodeStatusCompleted, EpisodeStatusFailed, EpisodeStatusCancelled:
		return ErrTerminalEpisode
	}
	return fmt.Errorf("%w: from %s to %s", ErrInvalidStatusChange, from, to)
}

func (m *defaultEpisodeManager) TransitionStatus(id EpisodeID, toStatus EpisodeStatus, outcome EpisodeOutcome, pauseReason PauseReason, resumeReason ResumeReason) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ep, exists := m.episodes[id]
	if !exists {
		return ErrEpisodeNotFound
	}

	if err := ValidateTransition(ep.Runtime.Status, toStatus); err != nil {
		return err
	}

	now := time.Now().UTC()
	ep.Runtime.Status = toStatus
	ep.Runtime.UpdatedAt = now

	if pauseReason != "" && toStatus == EpisodeStatusPaused {
		ep.Runtime.PauseReason = pauseReason
	}
	if resumeReason != "" && (toStatus == EpisodeStatusRunning || toStatus == EpisodeStatusWaiting) {
		ep.Runtime.ResumeReason = resumeReason
		ep.Runtime.PauseReason = ""
	}

	if toStatus == EpisodeStatusCompleted || toStatus == EpisodeStatusFailed || toStatus == EpisodeStatusCancelled {
		ep.Runtime.CompletedAt = &now
		if outcome != "" {
			ep.Runtime.Outcome = outcome
		} else {
			if toStatus == EpisodeStatusCompleted {
				ep.Runtime.Outcome = EpisodeOutcomeSuccess
			} else if toStatus == EpisodeStatusFailed {
				ep.Runtime.Outcome = EpisodeOutcomeFailed
			} else {
				ep.Runtime.Outcome = EpisodeOutcomeCancelled
			}
		}
		ep.Runtime.TerminationSummary.Completed = (toStatus == EpisodeStatusCompleted)
		ep.Runtime.TerminationSummary.Cancelled = (toStatus == EpisodeStatusCancelled)
	}

	return ep.Validate()
}

func (m *defaultEpisodeManager) UpdatePriority(id EpisodeID, newPriority PriorityBand, reason PriorityTransitionReason) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ep, exists := m.episodes[id]
	if !exists {
		return ErrEpisodeNotFound
	}
	if ep.Runtime.CurrentPriority == newPriority && reason == PriorityReasonDefault {
		return nil
	}

	now := time.Now().UTC()
	transition := PriorityTransition{
		FromPriority: ep.Runtime.CurrentPriority,
		ToPriority:   newPriority,
		Reason:       reason,
		Timestamp:    now,
	}
	ep.Runtime.CurrentPriority = newPriority
	ep.Runtime.UpdatedAt = now
	ep.Runtime.PriorityHistory = append(ep.Runtime.PriorityHistory, transition)

	// Enforce bounded rolling history limit (Max 16)
	if len(ep.Runtime.PriorityHistory) > maxRollingHistory {
		ep.Runtime.PriorityHistory = ep.Runtime.PriorityHistory[len(ep.Runtime.PriorityHistory)-maxRollingHistory:]
	}
	return nil
}

func (m *defaultEpisodeManager) UpdateBudget(id EpisodeID, newBudget BudgetTier, reason BudgetTransitionReason) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ep, exists := m.episodes[id]
	if !exists {
		return ErrEpisodeNotFound
	}
	if ep.Runtime.CurrentBudget == newBudget && reason == BudgetReasonInitialAssignment {
		return nil
	}

	now := time.Now().UTC()
	transition := BudgetTransition{
		FromBudget: ep.Runtime.CurrentBudget,
		ToBudget:   newBudget,
		Reason:     reason,
		Timestamp:  now,
	}
	ep.Runtime.CurrentBudget = newBudget
	ep.Runtime.UpdatedAt = now
	ep.Runtime.BudgetHistory = append(ep.Runtime.BudgetHistory, transition)

	// Enforce bounded rolling history limit (Max 16)
	if len(ep.Runtime.BudgetHistory) > maxRollingHistory {
		ep.Runtime.BudgetHistory = ep.Runtime.BudgetHistory[len(ep.Runtime.BudgetHistory)-maxRollingHistory:]
	}
	return nil
}

func (m *defaultEpisodeManager) CreateCheckpoint(id EpisodeID, reason CheckpointReason) (*EpisodeCheckpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ep, exists := m.episodes[id]
	if !exists {
		return nil, ErrEpisodeNotFound
	}

	now := time.Now().UTC()
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s|%s|%d|%v", ep.Runtime.EpisodeID, ep.Runtime.Status, ep.Runtime.CurrentHorizon, ep.Runtime.UpdatedAt)))
	rtFingerprint := hex.EncodeToString(hasher.Sum(nil))

	cpID := fmt.Sprintf("cp-%s-%d", id, now.UnixNano())
	cp := &EpisodeCheckpoint{
		CheckpointID:       cpID,
		EpisodeID:          id,
		RuntimeFingerprint: rtFingerprint,
		WorkspaceReference: ep.Definition.ContextReference.WorkspaceReference,
		AttentionReference: ep.Definition.ContextReference.AttentionReference,
		Timestamp:          now,
		ReplayMetadata:     ep.Definition.ReplayMetadata,
	}
	if err := cp.Validate(); err != nil {
		return nil, fmt.Errorf("episode_manager: checkpoint validation failed: %w", err)
	}
	m.checkpoints[cpID] = cp
	return cp, nil
}

func (m *defaultEpisodeManager) GetCheckpoint(checkpointID string) (*EpisodeCheckpoint, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp, exists := m.checkpoints[checkpointID]
	return cp, exists
}

func (m *defaultEpisodeManager) RestoreFromCheckpoint(cp *EpisodeCheckpoint, def *ExecutiveEpisodeDefinition) (*ExecutiveEpisode, error) {
	if cp == nil || def == nil {
		return nil, errors.New("episode_manager: checkpoint or definition is nil")
	}
	if err := cp.Validate(); err != nil {
		return nil, fmt.Errorf("episode_manager: invalid checkpoint: %w", err)
	}
	if err := def.Validate(); err != nil {
		return nil, fmt.Errorf("episode_manager: invalid definition: %w", err)
	}
	if cp.EpisodeID != def.EpisodeID {
		return nil, errors.New("episode_manager: checkpoint and definition episode_id mismatch")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	runtime := &ExecutiveEpisodeRuntime{
		EpisodeID:          def.EpisodeID,
		Status:             EpisodeStatusPaused,
		Outcome:            EpisodeOutcomePending,
		CurrentPriority:    PriorityBand2Interactive,
		CurrentBudget:      BudgetStandard,
		RemainingCostUnits: 100,
		PriorityHistory:    make([]PriorityTransition, 0, 16),
		BudgetHistory:      make([]BudgetTransition, 0, 16),
		PauseReason:        PauseReasonCheckpointing,
		CreatedAt:          cp.Timestamp,
		UpdatedAt:          now,
	}

	ep := &ExecutiveEpisode{
		Definition: def,
		Runtime:    runtime,
	}
	if err := ep.Validate(); err != nil {
		return nil, fmt.Errorf("episode_manager: restored episode failed validation: %w", err)
	}
	m.episodes[def.EpisodeID] = ep
	return ep, nil
}
