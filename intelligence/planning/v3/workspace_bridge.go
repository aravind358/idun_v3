package v3

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"idun/core/storage"
	"idun/intelligence/communication"
	reasoningv3 "idun/intelligence/reasoning/v3"
)

// WorkspacePublisher defines the interface to emit Envelopes.
type WorkspacePublisher interface {
	Publish(ctx context.Context, env communication.Envelope) error
}

// WorkspaceSubscription represents an active subscription.
type WorkspaceSubscription interface {
	Cancel() error
}

// WorkspaceSubscriber defines the interface to subscribe to topics.
type WorkspaceSubscriber interface {
	Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error)
}

// PayloadStorer defines the CAS interface.
type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
	StoreArtifact(ctx context.Context, artifactID string, meta storage.ArtifactIndexMeta, data []byte) (string, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// WorkspaceBridge wraps the V3 Planning Orchestrator.
type WorkspaceBridge struct {
	orchestrator *Orchestrator
	storer       PayloadStorer
	publisher    WorkspacePublisher
	subscriber   WorkspaceSubscriber
	name         string

	mu  sync.Mutex
	sub WorkspaceSubscription
}

// NewWorkspaceBridge creates a bridge for Planning V3.
func NewWorkspaceBridge(orchestrator *Orchestrator, storer PayloadStorer, publisher WorkspacePublisher, subscriber WorkspaceSubscriber) *WorkspaceBridge {
	return &WorkspaceBridge{
		orchestrator: orchestrator,
		storer:       storer,
		publisher:    publisher,
		subscriber:   subscriber,
		name:         "Intelligence.PlanningV3",
	}
}

// Start subscribes the Planning module to TopicActiveGoals.
func (b *WorkspaceBridge) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriber != nil && b.sub == nil {
		sub, err := b.subscriber.Subscribe(communication.TopicActiveGoals, b.name, b.HandleActiveGoal)
		if err != nil {
			return err
		}
		b.sub = sub
	}
	return nil
}

// Close unsubscribes the Planning module.
func (b *WorkspaceBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.sub != nil {
		_ = b.sub.Cancel()
		b.sub = nil
	}
	return nil
}

// HandleActiveGoal processes an incoming ReasoningContext from TopicActiveGoals.
func (b *WorkspaceBridge) HandleActiveGoal(ctx context.Context, env communication.Envelope) error {
	fmt.Printf(">>> Planning V3 HandleActiveGoal invoked! Envelope ID: %s\n", env.ID)
	if env.PayloadRef == "" {
		return fmt.Errorf("planning_v3: payload ref missing")
	}

	data, err := b.storer.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return err
	}

	var rContexts []*reasoningv3.ReasoningContext
	if err := json.Unmarshal(data, &rContexts); err != nil {
		// Fallback for single object during migration
		var singleCtx reasoningv3.ReasoningContext
		if err2 := json.Unmarshal(data, &singleCtx); err2 == nil {
			rContexts = append(rContexts, &singleCtx)
		} else {
			return err
		}
	}

	// In the event-driven architecture, the SemanticInterpretation isn't explicitly
	// linked in the ReasoningContext. For integration purposes, we pass nil to
	// orchestrator.Plan as the mock V3 implementation does not consume it.
	execPlan, err := b.orchestrator.Plan(ctx, nil, rContexts)
	if err != nil {
		return err
	}

	resData, err := json.Marshal(execPlan)
	if err != nil {
		return err
	}

	meta := storage.ArtifactIndexMeta{
		Type:    "ExecutionPlan",
		Version: "3.0",
	}
	ref, err := b.storer.StoreArtifact(ctx, string(execPlan.ArtifactID()), meta, resData)
	if err != nil {
		return err
	}

	parentRef := env.ParentRef
	if parentRef == "" {
		parentRef = env.ID
	}

	outEnv := communication.Envelope{
		ID:         string(execPlan.ArtifactID()),
		Topic:      communication.TopicCandidatePlans,
		ParentRef:  parentRef,
		CreatedAt:  time.Now().UTC(),
		PayloadRef: ref,
		Source:     b.name,
	}

	return b.publisher.Publish(ctx, outEnv)
}
