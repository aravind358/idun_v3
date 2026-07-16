package executive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ExecutiveEpisodeBuilder constructs an ExecutiveEpisode with automatic fingerprinting and validation firewalls.
type ExecutiveEpisodeBuilder struct {
	epID                EpisodeID
	epType              EpisodeType
	intent              EpisodeIntent
	origin              EpisodeOrigin
	parentID            EpisodeID
	rootID              EpisodeID
	hierarchyRef        string
	dependencyRef       string
	ctxRef              EpisodeContext
	capabilities        EpisodeCapabilities
	replayMeta          ReplayMetadata
	priority            PriorityBand
	horizon             int
	budget              BudgetTier
	costUnits           int
	executorID          string
	err                 error
}

// NewExecutiveEpisodeBuilder initializes a new fluent builder for ExecutiveEpisode.
func NewExecutiveEpisodeBuilder(id EpisodeID, epType EpisodeType, intent EpisodeIntent, origin EpisodeOrigin) *ExecutiveEpisodeBuilder {
	return &ExecutiveEpisodeBuilder{
		epID:      id,
		epType:    epType,
		intent:    intent,
		origin:    origin,
		rootID:    id,
		priority:  PriorityBand2Interactive,
		budget:    BudgetStandard,
		costUnits: 100,
	}
}

// WithParent sets the parent and root episode references.
func (b *ExecutiveEpisodeBuilder) WithParent(parentID, rootID EpisodeID) *ExecutiveEpisodeBuilder {
	b.parentID = parentID
	if rootID != "" {
		b.rootID = rootID
	}
	return b
}

// WithHierarchyReference sets the Workspace hierarchy graph reference.
func (b *ExecutiveEpisodeBuilder) WithHierarchyReference(ref string) *ExecutiveEpisodeBuilder {
	b.hierarchyRef = ref
	return b
}

// WithDependencyReference sets the Workspace dependency graph reference.
func (b *ExecutiveEpisodeBuilder) WithDependencyReference(ref string) *ExecutiveEpisodeBuilder {
	b.dependencyRef = ref
	return b
}

// WithContextReference sets the Hybrid EpisodeContext reference.
func (b *ExecutiveEpisodeBuilder) WithContextReference(ctxRef EpisodeContext) *ExecutiveEpisodeBuilder {
	b.ctxRef = ctxRef
	return b
}

// WithCapabilities sets the orchestration capabilities supported by this episode.
func (b *ExecutiveEpisodeBuilder) WithCapabilities(caps EpisodeCapabilities) *ExecutiveEpisodeBuilder {
	b.capabilities = caps
	return b
}

// WithReplayMetadata sets historical replay headers.
func (b *ExecutiveEpisodeBuilder) WithReplayMetadata(meta ReplayMetadata) *ExecutiveEpisodeBuilder {
	b.replayMeta = meta
	return b
}

// WithPriorityAndBudget sets initial priority band and budget tier.
func (b *ExecutiveEpisodeBuilder) WithPriorityAndBudget(priority PriorityBand, budget BudgetTier, costUnits int) *ExecutiveEpisodeBuilder {
	b.priority = priority
	b.budget = budget
	b.costUnits = costUnits
	return b
}

// WithExecutorID sets the initial execution node identifier.
func (b *ExecutiveEpisodeBuilder) WithExecutorID(executorID string) *ExecutiveEpisodeBuilder {
	b.executorID = executorID
	return b
}

// Build computes the SHA-256 definition fingerprint, constructs the episode, and runs validation firewalls.
func (b *ExecutiveEpisodeBuilder) Build() (*ExecutiveEpisode, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.epID == "" {
		return nil, ErrInvalidEpisodeID
	}
	if err := b.ctxRef.Validate(); err != nil {
		return nil, fmt.Errorf("executive_builder: context validation failed: %w", err)
	}

	// Compute deterministic SHA-256 fingerprint over immutable fields
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		b.epID, b.epType, b.intent, b.origin, b.parentID, b.rootID,
		b.hierarchyRef, b.dependencyRef, b.ctxRef.WorkspaceReference,
		b.ctxRef.AttentionReference, b.ctxRef.GoalReference,
	)))
	fingerprint := hex.EncodeToString(hasher.Sum(nil))

	now := time.Now().UTC()
	def := &ExecutiveEpisodeDefinition{
		EpisodeID:           b.epID,
		EpisodeType:         b.epType,
		EpisodeIntent:       b.intent,
		EpisodeOrigin:       b.origin,
		ParentEpisodeID:     b.parentID,
		RootEpisodeID:       b.rootID,
		SchemaVersion:       "2.0.0",
		HierarchyReference:  b.hierarchyRef,
		DependencyReference: b.dependencyRef,
		ContextReference:    b.ctxRef,
		Capabilities:        b.capabilities,
		EpisodeFingerprint:  fingerprint,
		ReplayMetadata:      b.replayMeta,
	}

	runtime := &ExecutiveEpisodeRuntime{
		EpisodeID:          b.epID,
		Status:             EpisodeStatusCreated,
		Outcome:            EpisodeOutcomePending,
		CurrentPriority:    b.priority,
		CurrentHorizon:     b.horizon,
		CurrentBudget:      b.budget,
		RemainingCostUnits: b.costUnits,
		PriorityHistory:    make([]PriorityTransition, 0, 16),
		BudgetHistory:      make([]BudgetTransition, 0, 16),
		ExecutorID:         b.executorID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Record initial priority and budget assignments into bounded histories
	runtime.PriorityHistory = append(runtime.PriorityHistory, PriorityTransition{
		FromPriority: b.priority,
		ToPriority:   b.priority,
		Reason:       PriorityReasonDefault,
		Timestamp:    now,
	})
	runtime.BudgetHistory = append(runtime.BudgetHistory, BudgetTransition{
		FromBudget: b.budget,
		ToBudget:   b.budget,
		Reason:       BudgetReasonInitialAssignment,
		Timestamp:    now,
	})

	episode := &ExecutiveEpisode{
		Definition: def,
		Runtime:    runtime,
	}
	if err := episode.Validate(); err != nil {
		return nil, fmt.Errorf("executive_builder: validation firewall rejected built episode: %w", err)
	}
	return episode, nil
}

// EpisodeCheckpointBuilder constructs an EpisodeCheckpoint artifact with validation firewalls.
type EpisodeCheckpointBuilder struct {
	cpID          string
	epID          EpisodeID
	rtFingerprint string
	wsRef         string
	attRef        string
	timestamp     time.Time
	replayMeta    ReplayMetadata
}

// NewEpisodeCheckpointBuilder initializes a new builder for EpisodeCheckpoint.
func NewEpisodeCheckpointBuilder(cpID string, epID EpisodeID) *EpisodeCheckpointBuilder {
	return &EpisodeCheckpointBuilder{
		cpID:      cpID,
		epID:      epID,
		timestamp: time.Now().UTC(),
	}
}

// WithRuntimeFingerprint sets the runtime execution fingerprint at checkpoint time.
func (cb *EpisodeCheckpointBuilder) WithRuntimeFingerprint(fp string) *EpisodeCheckpointBuilder {
	cb.rtFingerprint = fp
	return cb
}

// WithReferences sets the workspace and attention references.
func (cb *EpisodeCheckpointBuilder) WithReferences(wsRef, attRef string) *EpisodeCheckpointBuilder {
	cb.wsRef = wsRef
	cb.attRef = attRef
	return cb
}

// WithReplayMetadata sets the replay provenance header.
func (cb *EpisodeCheckpointBuilder) WithReplayMetadata(meta ReplayMetadata) *EpisodeCheckpointBuilder {
	cb.replayMeta = meta
	return cb
}

// Build constructs and validates the immutable EpisodeCheckpoint.
func (cb *EpisodeCheckpointBuilder) Build() (*EpisodeCheckpoint, error) {
	cp := &EpisodeCheckpoint{
		CheckpointID:       cb.cpID,
		EpisodeID:          cb.epID,
		RuntimeFingerprint: cb.rtFingerprint,
		WorkspaceReference: cb.wsRef,
		AttentionReference: cb.attRef,
		Timestamp:          cb.timestamp,
		ReplayMetadata:     cb.replayMeta,
	}
	if err := cp.Validate(); err != nil {
		return nil, fmt.Errorf("checkpoint_builder: validation firewall rejected checkpoint: %w", err)
	}
	return cp, nil
}
