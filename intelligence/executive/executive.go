// Package executive implements IDUN's Intelligence Pillar Executive Functions.
package executive

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"idun/core/logger"
)

// ============================================================================
// Service Configuration
// ============================================================================

// Config configures the ExecutiveService at construction time.
type Config struct {
	// Logger is the structured logger sink. Optional (nil-safe).
	Logger logger.Writer

	// IdleThreshold defines the duration of zero external salience before
	// homeostatic consolidation workflows may be scheduled.
	IdleThreshold time.Duration

	// DefaultMaxFuel defines the default maximum transition count per workflow.
	DefaultMaxFuel int

	// DefaultMaxReflection defines the default maximum recursive reflection depth.
	DefaultMaxReflection int
}

// ============================================================================
// Concrete ExecutiveService Implementation
// ============================================================================

// ExecutiveService is the thread-safe concrete implementation of Executive Functions.
// It implements the Executive composite interface and Kernel Component convention.
type ExecutiveService struct {
	mu            sync.RWMutex
	log           logger.Writer
	closed        bool
	idleThreshold time.Duration

	// Attentional & Goal state
	activeGoal ActiveGoalContext

	// Priority queues (5 bands: 0..4)
	queues [5][]*WorkflowGraph

	// Active task cancellation registry
	cancelFuncs map[string]context.CancelFunc

	// Registered Cognitive Ability Drivers
	drivers map[CognitiveAbility]AbilityDriver

	// Homeostasis state
	lastActivity time.Time

	// Default execution limits
	defaultMaxFuel       int
	defaultMaxReflection int
}

// NewExecutiveService creates and validates a new ExecutiveService instance.
func NewExecutiveService(cfg Config) *ExecutiveService {
	idleThreshold := cfg.IdleThreshold
	if idleThreshold <= 0 {
		idleThreshold = 30 * time.Second
	}
	defaultFuel := cfg.DefaultMaxFuel
	if defaultFuel <= 0 {
		defaultFuel = 16
	}
	defaultReflect := cfg.DefaultMaxReflection
	if defaultReflect <= 0 {
		defaultReflect = 2
	}

	return &ExecutiveService{
		log:                  cfg.Logger,
		idleThreshold:        idleThreshold,
		cancelFuncs:          make(map[string]context.CancelFunc),
		drivers:              make(map[CognitiveAbility]AbilityDriver),
		lastActivity:         time.Now(),
		defaultMaxFuel:       defaultFuel,
		defaultMaxReflection: defaultReflect,
	}
}

// ============================================================================
// Kernel Component Lifecycle
// ============================================================================

// Name returns the canonical Kernel Component name.
func (e *ExecutiveService) Name() string {
	return "Intelligence.Executive"
}

// Start boots the Executive Service.
func (e *ExecutiveService) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrServiceClosed
	}
	e.lastActivity = time.Now()
	if e.log != nil {
		e.log.Info("ExecutiveService started", logger.Field{Key: "component", Value: e.Name()})
	}
	return nil
}

// Close gracefully shuts down the Executive Service and cancels all running tasks.
func (e *ExecutiveService) Close() error {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return nil
	}
	e.closed = true
	cancelMap := make(map[string]context.CancelFunc, len(e.cancelFuncs))
	for k, v := range e.cancelFuncs {
		cancelMap[k] = v
	}
	e.cancelFuncs = make(map[string]context.CancelFunc)
	e.mu.Unlock()

	for _, cancel := range cancelMap {
		cancel()
	}

	if e.log != nil {
		e.log.Info("ExecutiveService closed", logger.Field{Key: "component", Value: e.Name()})
	}
	return nil
}

// ============================================================================
// AttentionGate Capability Implementation
// ============================================================================

// Evaluate inspects a Stimulus against current ActiveGoalContext and assigns triage salience.
func (e *ExecutiveService) Evaluate(s Stimulus) (SalienceDecision, PriorityBand) {
	e.RecordActivity()
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Hardware safety or constitutional tripwires immediately route to Band 0 Focus
	if s.SafetyFlag {
		return SalienceFocusImmediately, PriorityBand0CriticalSafety
	}

	score := s.SalienceScore
	if score >= 85 {
		return SalienceFocusImmediately, PriorityBand1RealTime
	} else if score >= 50 {
		return SalienceFocusImmediately, PriorityBand2Interactive
	} else if score >= 20 {
		return SalienceSchedule, PriorityBand3Background
	}
	return SalienceFilter, PriorityBand4Idle
}

// SetActiveGoal updates the lightweight active goal header reference.
func (e *ExecutiveService) SetActiveGoal(goal ActiveGoalContext) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activeGoal = goal
	if e.log != nil {
		e.log.Info("Active goal updated",
			logger.Field{Key: "goal_id", Value: goal.ID},
			logger.Field{Key: "summary", Value: goal.Summary},
		)
	}
}

// GetActiveGoal returns the current active goal header reference.
func (e *ExecutiveService) GetActiveGoal() ActiveGoalContext {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.activeGoal
}

// ============================================================================
// PriorityEngine Capability Implementation
// ============================================================================

// Enqueue schedules a workflow into the specified PriorityBand queue.
func (e *ExecutiveService) Enqueue(wg *WorkflowGraph) error {
	if wg == nil {
		return errors.New("executive: nil workflow graph")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrServiceClosed
	}

	band := wg.Priority
	if band < 0 || band > 4 {
		band = PriorityBand2Interactive
	}
	e.queues[band] = append(e.queues[band], wg)
	return nil
}

// Dequeue returns the highest priority workflow currently pending execution.
func (e *ExecutiveService) Dequeue() (*WorkflowGraph, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for band := 0; band < 5; band++ {
		if len(e.queues[band]) > 0 {
			wg := e.queues[band][0]
			e.queues[band] = e.queues[band][1:]
			return wg, true
		}
	}
	return nil, false
}

// Preempt signals immediate interruption of lower-priority active tasks when Band 0/1 arrives.
func (e *ExecutiveService) Preempt(ctx context.Context, incomingBand PriorityBand) error {
	if incomingBand > PriorityBand1RealTime {
		return nil
	}
	e.mu.RLock()
	cancelMap := make(map[string]context.CancelFunc, len(e.cancelFuncs))
	for k, v := range e.cancelFuncs {
		cancelMap[k] = v
	}
	e.mu.RUnlock()

	for _, cancel := range cancelMap {
		cancel()
	}
	return nil
}

// ============================================================================
// BudgetManager Capability Implementation
// ============================================================================

// AssignBudget determines the initial BudgetTier for a stimulus or workflow.
func (e *ExecutiveService) AssignBudget(salience SalienceDecision, priority PriorityBand) BudgetTier {
	if priority == PriorityBand0CriticalSafety || salience == SalienceFocusImmediately && priority == PriorityBand1RealTime {
		return BudgetReflexive
	}
	if priority == PriorityBand4Idle {
		return BudgetDeliberative
	}
	return BudgetStandard
}

// EvaluateEscalation decides whether an escalation request from a cognitive ability is granted.
func (e *ExecutiveService) EvaluateEscalation(current BudgetTier, priority PriorityBand) (BudgetTier, bool) {
	if current < BudgetDeliberative {
		return current + 1, true
	}
	return current, false
}

// ============================================================================
// CancellationCoordinator Capability Implementation
// ============================================================================

// RegisterTask registers a workflow context for potential preemption or cancellation.
func (e *ExecutiveService) RegisterTask(workflowID string, cancelFunc context.CancelFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if cancelFunc != nil {
		e.cancelFuncs[workflowID] = cancelFunc
	}
}

// CancelTask cancels a specific running workflow by ID.
func (e *ExecutiveService) CancelTask(workflowID string) error {
	e.mu.Lock()
	cancel, ok := e.cancelFuncs[workflowID]
	if ok {
		delete(e.cancelFuncs, workflowID)
	}
	e.mu.Unlock()

	if !ok {
		return ErrWorkflowCancelled
	}
	cancel()
	return nil
}

// CancelAll cancels all currently active workflows.
func (e *ExecutiveService) CancelAll() {
	e.mu.Lock()
	cancelMap := make(map[string]context.CancelFunc, len(e.cancelFuncs))
	for k, v := range e.cancelFuncs {
		cancelMap[k] = v
	}
	e.cancelFuncs = make(map[string]context.CancelFunc)
	e.mu.Unlock()

	for _, cancel := range cancelMap {
		cancel()
	}
}

// ============================================================================
// HomeostasisController Capability Implementation
// ============================================================================

// RecordActivity records that external stimulus or reactive processing occurred.
func (e *ExecutiveService) RecordActivity() {
	e.mu.Lock()
	e.lastActivity = time.Now()
	e.mu.Unlock()
}

// ShouldConsolidate returns true if the system has been idle longer than the idle threshold.
func (e *ExecutiveService) ShouldConsolidate() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return time.Since(e.lastActivity) >= e.idleThreshold
}

// IdleDuration returns the duration since last external activity.
func (e *ExecutiveService) IdleDuration() time.Duration {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return time.Since(e.lastActivity)
}

// ============================================================================
// AbilityRegistry Capability Implementation
// ============================================================================

// RegisterDriver registers an AbilityDriver for its declared CognitiveAbility.
func (e *ExecutiveService) RegisterDriver(driver AbilityDriver) error {
	if driver == nil {
		return errors.New("executive: nil ability driver")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrServiceClosed
	}
	e.drivers[driver.Ability()] = driver
	if e.log != nil {
		e.log.Info("Registered cognitive ability driver", logger.Field{Key: "ability", Value: string(driver.Ability())})
	}
	return nil
}

// GetDriver returns the registered driver for a given CognitiveAbility.
func (e *ExecutiveService) GetDriver(ability CognitiveAbility) (AbilityDriver, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	driver, ok := e.drivers[ability]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrAbilityNotRegistered, ability)
	}
	return driver, nil
}

// ListAbilities returns all currently registered cognitive abilities.
func (e *ExecutiveService) ListAbilities() []CognitiveAbility {
	e.mu.RLock()
	defer e.mu.RUnlock()
	abilities := make([]CognitiveAbility, 0, len(e.drivers))
	for ability := range e.drivers {
		abilities = append(abilities, ability)
	}
	return abilities
}

// ============================================================================
// WorkflowCoordinator Capability Implementation
// ============================================================================

// Execute runs a WorkflowGraph through registered cognitive ability drivers until completion or termination.
func (e *ExecutiveService) Execute(ctx context.Context, wg *WorkflowGraph) ExecutionResult {
	startTime := time.Now()
	e.RecordActivity()

	if wg == nil {
		return ExecutionResult{Error: errors.New("executive: nil workflow graph")}
	}

	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	e.RegisterTask(wg.ID, cancel)
	defer func() {
		e.mu.Lock()
		delete(e.cancelFuncs, wg.ID)
		e.mu.Unlock()
	}()

	currentNodeID := wg.StartNodeID
	currentPayloadRef := ""
	transitions := 0

	fuel := wg.MaxFuel
	if fuel <= 0 {
		fuel = e.defaultMaxFuel
	}
	maxReflect := wg.MaxReflection
	if maxReflect <= 0 {
		maxReflect = e.defaultMaxReflection
	}
	currentBudget := wg.Budget

	for currentNodeID != "" {
		if execCtx.Err() != nil {
			return ExecutionResult{
				WorkflowID:       wg.ID,
				FinalStatus:      StatusUnresolvableContradiction,
				Error:            ErrWorkflowCancelled,
				TransitionsCount: transitions,
				Duration:         time.Since(startTime),
			}
		}

		if fuel <= 0 {
			return ExecutionResult{
				WorkflowID:       wg.ID,
				FinalStatus:      StatusUnresolvableContradiction,
				Error:            ErrMaxFuelExceeded,
				TransitionsCount: transitions,
				Duration:         time.Since(startTime),
			}
		}

		node, ok := wg.Nodes[currentNodeID]
		if !ok {
			return ExecutionResult{
				WorkflowID:       wg.ID,
				Error:            fmt.Errorf("executive: node %s not found in workflow", currentNodeID),
				TransitionsCount: transitions,
				Duration:         time.Since(startTime),
			}
		}

		driver, err := e.GetDriver(node.Ability)
		if err != nil {
			return ExecutionResult{
				WorkflowID:       wg.ID,
				Error:            err,
				TransitionsCount: transitions,
				Duration:         time.Since(startTime),
			}
		}

		inputRef := node.InputRef
		if currentPayloadRef != "" {
			inputRef = currentPayloadRef
		}

		status, outputRef, execErr := driver.ExecuteTask(execCtx, wg.ID, node.ID, currentBudget, inputRef)
		transitions++
		fuel--

		if execErr != nil {
			if errors.Is(execErr, context.Canceled) || errors.Is(execErr, context.DeadlineExceeded) || execCtx.Err() != nil {
				execErr = ErrWorkflowCancelled
			}
			return ExecutionResult{
				WorkflowID:       wg.ID,
				Error:            execErr,
				TransitionsCount: transitions,
				Duration:         time.Since(startTime),
			}
		}

		switch status {
		case StatusConfident:
			currentPayloadRef = outputRef
			currentNodeID = node.NextNodeID

		case StatusEscalationRequired:
			// Explicit budget escalation request arbitration
			upgraded, ok := e.EvaluateEscalation(currentBudget, wg.Priority)
			if ok {
				currentBudget = upgraded
				// Re-attempt current node with upgraded budget
				continue
			}
			return ExecutionResult{
				WorkflowID:       wg.ID,
				FinalStatus:      StatusEscalationRequired,
				Error:            ErrBudgetExhausted,
				TransitionsCount: transitions,
				Duration:         time.Since(startTime),
			}

		case StatusUnsureConflicting:
			// Contradiction detected: route to Reflection unless MaxReflection exceeded
			if wg.ReflectionDepth >= maxReflect {
				return ExecutionResult{
					WorkflowID:       wg.ID,
					FinalStatus:      StatusUnresolvableContradiction,
					Error:            ErrMaxReflectionExceeded,
					TransitionsCount: transitions,
					Duration:         time.Since(startTime),
				}
			}
			wg.ReflectionDepth++

			reflectDriver, rErr := e.GetDriver(AbilityReflection)
			if rErr == nil {
				_, refOut, _ := reflectDriver.ExecuteTask(execCtx, wg.ID, "reflection_gate", currentBudget, outputRef)
				currentPayloadRef = refOut
				continue
			}
			return ExecutionResult{
				WorkflowID:       wg.ID,
				FinalStatus:      StatusUnsureConflicting,
				TransitionsCount: transitions,
				Duration:         time.Since(startTime),
			}

		case StatusUnsureConstitutionalRisk:
			// Safety intercept: route to CognitiveAbility.Value
			valueDriver, vErr := e.GetDriver(AbilityValue)
			if vErr != nil {
				return ExecutionResult{
					WorkflowID:       wg.ID,
					Error:            vErr,
					TransitionsCount: transitions,
					Duration:         time.Since(startTime),
				}
			}
			valStatus, _, _ := valueDriver.ExecuteTask(execCtx, wg.ID, "constitutional_gate", BudgetStandard, outputRef)
			if valStatus != StatusConfident {
				return ExecutionResult{
					WorkflowID:       wg.ID,
					FinalStatus:      StatusUnsureConstitutionalRisk,
					Error:            ErrConstitutionalVeto,
					TransitionsCount: transitions,
					Duration:         time.Since(startTime),
				}
			}
			currentPayloadRef = outputRef
			currentNodeID = node.NextNodeID

		default:
			return ExecutionResult{
				WorkflowID:       wg.ID,
				FinalStatus:      status,
				OutputRef:        outputRef,
				TransitionsCount: transitions,
				Duration:         time.Since(startTime),
			}
		}
	}

	return ExecutionResult{
		WorkflowID:       wg.ID,
		FinalStatus:      StatusConfident,
		OutputRef:        currentPayloadRef,
		TransitionsCount: transitions,
		Duration:         time.Since(startTime),
	}
}
