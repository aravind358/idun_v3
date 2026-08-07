package v3

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"idun/intelligence/communication"
	underv3 "idun/intelligence/understanding/v3"
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
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// WorkspaceBridge wraps the V3 Reasoning Orchestrator.
type WorkspaceBridge struct {
	orchestrator *Orchestrator
	storer       PayloadStorer
	publisher    WorkspacePublisher
	subscriber   WorkspaceSubscriber
	name         string

	mu  sync.Mutex
	sub WorkspaceSubscription
}

// NewWorkspaceBridge creates a bridge for Reasoning V3.
func NewWorkspaceBridge(orchestrator *Orchestrator, storer PayloadStorer, publisher WorkspacePublisher, subscriber WorkspaceSubscriber) *WorkspaceBridge {
	return &WorkspaceBridge{
		orchestrator: orchestrator,
		storer:       storer,
		publisher:    publisher,
		subscriber:   subscriber,
		name:         "Intelligence.ReasoningV3",
	}
}

func (b *WorkspaceBridge) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriber != nil && b.sub == nil {
		sub, err := b.subscriber.Subscribe(communication.TopicResolvedIntent, b.name, b.HandleIntent)
		if err != nil {
			return err
		}
		b.sub = sub
	}
	return nil
}

func (b *WorkspaceBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.sub != nil {
		_ = b.sub.Cancel()
		b.sub = nil
	}
	return nil
}

func (b *WorkspaceBridge) HandleIntent(ctx context.Context, env communication.Envelope) error {
	fmt.Printf(">>> Reasoning V3 HandleIntent invoked! Envelope ID: %s\n", env.ID)
	if env.PayloadRef == "" {
		return fmt.Errorf("reasoning_v3: payload ref missing")
	}

	data, err := b.storer.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return err
	}

	batch := &underv3.UnderstandingBatch{}
	if err := json.Unmarshal(data, batch); err != nil {
		// Fallback for backward compatibility during rollout if needed, but per spec it's strictly a batch now.
		return err
	}
	
	var contexts []*ReasoningContext
	for _, interp := range batch.Interpretations() {
		rContext, err := b.orchestrator.Reason(ctx, interp)
		if err != nil {
			return err
		}
		contexts = append(contexts, rContext)
	}

	resData, err := json.Marshal(contexts)
	if err != nil {
		return err
	}

	ref, err := b.storer.Store(ctx, resData)
	if err != nil {
		return err
	}

	parentRef := env.ParentRef
	if parentRef == "" {
		parentRef = env.ID
	}
	
	// Ensure we have a valid ID for the outgoing envelope (using the batch's envelope ID or first context)
	outID := env.ID
	if len(contexts) > 0 {
		outID = string(contexts[0].ArtifactID())
	}

	outEnv := communication.Envelope{
		ID:         outID,
		Topic:      communication.TopicActiveGoals,
		ParentRef:  parentRef,
		CreatedAt:  time.Now().UTC(),
		PayloadRef: ref,
		Source:     b.name,
	}

	return b.publisher.Publish(ctx, outEnv)
}
