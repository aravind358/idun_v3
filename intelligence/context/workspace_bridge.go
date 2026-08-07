package context

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

// WorkspaceBridge wraps the Context Resolver for the Workspace.
type WorkspaceBridge struct {
	resolver    ContextResolver
	stateReader DialogueStateReader
	storer      PayloadStorer
	publisher   WorkspacePublisher
	subscriber  WorkspaceSubscriber
	name        string

	mu  sync.Mutex
	sub WorkspaceSubscription
}

// NewWorkspaceBridge creates a bridge for Context Resolver.
func NewWorkspaceBridge(resolver ContextResolver, stateReader DialogueStateReader, storer PayloadStorer, publisher WorkspacePublisher, subscriber WorkspaceSubscriber) *WorkspaceBridge {
	return &WorkspaceBridge{
		resolver:    resolver,
		stateReader: stateReader,
		storer:      storer,
		publisher:   publisher,
		subscriber:  subscriber,
		name:        "Intelligence.ContextResolver",
	}
}

func (b *WorkspaceBridge) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriber != nil && b.sub == nil {
		sub, err := b.subscriber.Subscribe(communication.TopicUserIntent, b.name, b.HandleIntent)
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
	if env.PayloadRef == "" {
		return fmt.Errorf("context_resolver: payload ref missing")
	}

	data, err := b.storer.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return err
	}

	var batch underv3.UnderstandingBatch
	err = json.Unmarshal(data, &batch)
	if err != nil {
		return fmt.Errorf("context_resolver: failed to unmarshal UnderstandingBatch: %w", err)
	}

	resolvedBatch, err := b.resolver.Resolve(ctx, &batch, b.stateReader)
	if err != nil {
		return err
	}

	if resolvedBatch == nil {
		return fmt.Errorf("context_resolver: resolved batch is nil")
	}

	resData, err := json.Marshal(resolvedBatch)
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

	outEnv := communication.Envelope{
		ID:         fmt.Sprintf("resolved-%s", env.ID),
		Topic:      communication.TopicResolvedIntent,
		ParentRef:  parentRef,
		CreatedAt:  time.Now().UTC(),
		PayloadRef: ref,
		Source:     b.name,
	}

	return b.publisher.Publish(ctx, outEnv)
}
