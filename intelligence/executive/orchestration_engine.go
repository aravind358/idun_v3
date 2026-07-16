package executive

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"idun/intelligence/workspace"
)

// OrchestrationEventType defines event types that drive asynchronous episode coordination.
type OrchestrationEventType string

const (
	EventPlanningCompleted  OrchestrationEventType = "PLANNING_COMPLETED"
	EventDecisionCompleted  OrchestrationEventType = "DECISION_COMPLETED"
	EventWorkspaceReady     OrchestrationEventType = "WORKSPACE_READY"
	EventDependencyResolved OrchestrationEventType = "DEPENDENCY_RESOLVED"
	EventAttentionChanged   OrchestrationEventType = "ATTENTION_CHANGED"
	EventEpisodeCancelled   OrchestrationEventType = "EPISODE_CANCELLED"
)

// OrchestrationEvent carries asynchronous coordination signals to the Executive engine.
type OrchestrationEvent struct {
	Type        OrchestrationEventType `json:"type"`
	EpisodeID   EpisodeID              `json:"episode_id"`
	PayloadRef  string                 `json:"payload_ref,omitempty"`
	Priority    *PriorityBand          `json:"priority,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ReflectionCoordinator defines the capability contract for waking the Reflection subsystem.
type ReflectionCoordinator interface {
	WakeReflection(ctx context.Context, epID EpisodeID, reason string) error
}

// LearningCoordinator defines the capability contract for waking the Learning subsystem.
type LearningCoordinator interface {
	WakeLearning(ctx context.Context, epID EpisodeID, reason string) error
}

// StrategyActivationCoordinator defines the capability contract for coordinating strategy activation.
type StrategyActivationCoordinator interface {
	ActivateStrategySnapshot(ctx context.Context, snapshotRef string) error
}

// EpisodeOrchestrator coordinates event-driven transitions, background tasks, and cognitive waking.
// Executive remains strictly coordination-only; it never performs thinking, learning, or reflection.
type EpisodeOrchestrator interface {
	// HandleEvent processes an orchestration event and triggers appropriate FSM transitions.
	HandleEvent(ctx context.Context, ev OrchestrationEvent) error

	// ScheduleBackgroundEpisode schedules a non-blocking background episode (learning, maintenance, calibration, indexing).
	ScheduleBackgroundEpisode(ctx context.Context, ep *ExecutiveEpisode) error

	// CoordinateReflection wakes Reflection when policy conditions or contradictions require it.
	CoordinateReflection(ctx context.Context, epID EpisodeID, reason string) error

	// CoordinateLearning wakes Learning according to consolidation or skill training policies.
	CoordinateLearning(ctx context.Context, epID EpisodeID, reason string) error

	// CoordinateStrategyActivation requests activation of a validated strategy snapshot without generating or modifying it.
	CoordinateStrategyActivation(ctx context.Context, snapshotRef string) error
}

type defaultEpisodeOrchestrator struct {
	mu           sync.RWMutex
	manager      EpisodeManager
	ws           workspace.Workspace
	reflection   ReflectionCoordinator
	learning     LearningCoordinator
	strategyAct  StrategyActivationCoordinator
}

// NewEpisodeOrchestrator creates a new event-driven orchestration engine.
func NewEpisodeOrchestrator(
	manager EpisodeManager,
	ws workspace.Workspace,
	ref ReflectionCoordinator,
	learn LearningCoordinator,
	strat StrategyActivationCoordinator,
) EpisodeOrchestrator {
	return &defaultEpisodeOrchestrator{
		manager:     manager,
		ws:          ws,
		reflection:  ref,
		learning:    learn,
		strategyAct: strat,
	}
}

func (o *defaultEpisodeOrchestrator) HandleEvent(ctx context.Context, ev OrchestrationEvent) error {
	if ev.EpisodeID == "" {
		return ErrInvalidEpisodeID
	}
	ep, exists := o.manager.GetEpisode(ev.EpisodeID)
	if !exists {
		return ErrEpisodeNotFound
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	now := time.Now().UTC()
	switch ev.Type {
	case EventDependencyResolved, EventWorkspaceReady:
		if o.ws != nil {
			ready, err := o.ws.IsEpisodeReady(ctx, string(ev.EpisodeID))
			if err != nil {
				return fmt.Errorf("orchestrator: dependency check failed: %w", err)
			}
			if ready && (ep.Runtime.Status == EpisodeStatusWaiting || ep.Runtime.Status == EpisodeStatusPaused) {
				return o.manager.TransitionStatus(ev.EpisodeID, EpisodeStatusRunning, EpisodeOutcomePending, "", ResumeReasonDependencyResolved)
			}
		} else if ep.Runtime.Status == EpisodeStatusWaiting || ep.Runtime.Status == EpisodeStatusPaused {
			return o.manager.TransitionStatus(ev.EpisodeID, EpisodeStatusRunning, EpisodeOutcomePending, "", ResumeReasonDependencyResolved)
		}

	case EventPlanningCompleted:
		ep.Runtime.SubsystemUsage.Planning.Invoked = true
		ep.Runtime.SubsystemUsage.Planning.Success = true
		ep.Runtime.UpdatedAt = now

	case EventDecisionCompleted:
		ep.Runtime.SubsystemUsage.Decision.Invoked = true
		ep.Runtime.SubsystemUsage.Decision.Success = true
		ep.Runtime.UpdatedAt = now
		// If decision is done, transition episode to completed
		if ep.Runtime.Status == EpisodeStatusRunning {
			return o.manager.TransitionStatus(ev.EpisodeID, EpisodeStatusCompleted, EpisodeOutcomeSuccess, "", "")
		}

	case EventAttentionChanged:
		if ev.Priority != nil && *ev.Priority != ep.Runtime.CurrentPriority {
			return o.manager.UpdatePriority(ev.EpisodeID, *ev.Priority, PriorityReasonSalienceOverride)
		}

	case EventEpisodeCancelled:
		if ep.Runtime.Status != EpisodeStatusCompleted && ep.Runtime.Status != EpisodeStatusFailed && ep.Runtime.Status != EpisodeStatusCancelled {
			return o.manager.TransitionStatus(ev.EpisodeID, EpisodeStatusCancelled, EpisodeOutcomeCancelled, "", "")
		}

	default:
		return fmt.Errorf("orchestrator: unhandled orchestration event type: %s", ev.Type)
	}
	return nil
}

func (o *defaultEpisodeOrchestrator) ScheduleBackgroundEpisode(ctx context.Context, ep *ExecutiveEpisode) error {
	if ep == nil {
		return errors.New("orchestrator: cannot schedule nil background episode")
	}
	if ep.Definition.EpisodeType != EpisodeTypeBackground {
		ep.Definition.EpisodeType = EpisodeTypeBackground
	}
	ep.Runtime.CurrentPriority = PriorityBand3Background
	if err := o.manager.CreateEpisode(ep); err != nil {
		return err
	}
	// Background episodes start in WAITING state or transition immediately to RUNNING if unblocked
	if o.ws != nil {
		ready, err := o.ws.IsEpisodeReady(ctx, string(ep.Definition.EpisodeID))
		if err == nil && ready {
			return o.manager.TransitionStatus(ep.Definition.EpisodeID, EpisodeStatusRunning, EpisodeOutcomePending, "", ResumeReasonSchedulerWake)
		}
	}
	return o.manager.TransitionStatus(ep.Definition.EpisodeID, EpisodeStatusWaiting, EpisodeOutcomePending, "", "")
}

func (o *defaultEpisodeOrchestrator) CoordinateReflection(ctx context.Context, epID EpisodeID, reason string) error {
	if o.reflection == nil {
		return nil // No-op if not configured
	}
	ep, exists := o.manager.GetEpisode(epID)
	if exists {
		ep.Runtime.SubsystemUsage.Reflection.Invoked = true
		ep.Runtime.UpdatedAt = time.Now().UTC()
	}
	return o.reflection.WakeReflection(ctx, epID, reason)
}

func (o *defaultEpisodeOrchestrator) CoordinateLearning(ctx context.Context, epID EpisodeID, reason string) error {
	if o.learning == nil {
		return nil // No-op if not configured
	}
	return o.learning.WakeLearning(ctx, epID, reason)
}

func (o *defaultEpisodeOrchestrator) CoordinateStrategyActivation(ctx context.Context, snapshotRef string) error {
	if o.strategyAct == nil {
		return nil // No-op if not configured
	}
	if snapshotRef == "" {
		return errors.New("orchestrator: strategy snapshot reference cannot be empty")
	}
	return o.strategyAct.ActivateStrategySnapshot(ctx, snapshotRef)
}
