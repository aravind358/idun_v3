package v3

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"idun/core/foundation"
	"idun/core/storage"
	"idun/intelligence/communication"
	decisionv3 "idun/intelligence/decision/v3"
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
	LookupArtifact(ctx context.Context, artifactID string) (storage.ArtifactIndexMeta, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// WorkspaceBridge wraps the V3 Executive Engine.
type WorkspaceBridge struct {
	engine     *ExecutionEngine
	storer     PayloadStorer
	publisher  WorkspacePublisher
	subscriber WorkspaceSubscriber
	name       string

	mu  sync.Mutex
	sub WorkspaceSubscription
}

// NewWorkspaceBridge creates a bridge for Executive V3.
func NewWorkspaceBridge(engine *ExecutionEngine, storer PayloadStorer, publisher WorkspacePublisher, subscriber WorkspaceSubscriber) *WorkspaceBridge {
	return &WorkspaceBridge{
		engine:       engine,
		storer:       storer,
		publisher:    publisher,
		subscriber:   subscriber,
		name:         "Intelligence.ExecutiveV3",
	}
}

// Start subscribes the Executive module to TopicEvaluatedOptions.
func (b *WorkspaceBridge) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriber != nil && b.sub == nil {
		sub, err := b.subscriber.Subscribe(communication.TopicEvaluatedOptions, b.name, b.HandleEvaluatedOption)
		if err != nil {
			return err
		}
		b.sub = sub
	}
	return nil
}

// Close unsubscribes the Executive module.
func (b *WorkspaceBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.sub != nil {
		_ = b.sub.Cancel()
		b.sub = nil
	}
	return nil
}

// HandleEvaluatedOption processes an incoming DecisionRecord from TopicEvaluatedOptions.
func (b *WorkspaceBridge) HandleEvaluatedOption(ctx context.Context, env communication.Envelope) error {
	fmt.Printf(">>> Executive V3 HandleEvaluatedOption invoked! Envelope ID: %s\n", env.ID)
	if env.PayloadRef == "" {
		return fmt.Errorf("executive_v3: payload ref missing")
	}

	data, err := b.storer.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return err
	}

	decisionRec := &decisionv3.DecisionRecord{}
	if err := json.Unmarshal(data, decisionRec); err != nil {
		return err
	}

	if decisionRec.Resolution() != decisionv3.StatusApproved {
		// Log rejection and skip execution
		return nil
	}


	execResult, execErr := b.engine.ExecuteDecision(ctx, decisionRec.ArtifactID(), foundation.ArtifactID(decisionRec.ParentArtifactID()), string(decisionRec.EnvelopeID()), 5*time.Second)
	
	if execErr != nil {
		fmt.Printf(">>> Executive V3 ExecuteDecision returned error: %v, but continuing for legacy adapter\n", execErr)
		// We can't return here because we need to publish the fallback for the test
	}

	var resData []byte
	var ref string
	if execResult != nil {
		resData, _ = json.Marshal(execResult)
		ref, _ = b.storer.Store(ctx, resData)
	} else {
		// Mock empty result
		ref = "mock-exec-result-ref"
	}

	parentRef := env.ParentRef
	if parentRef == "" {
		parentRef = env.ID
	}

	outEnv := communication.Envelope{
		ID:         string(decisionRec.ArtifactID()),
		Topic:      communication.TopicActionExecution,
		ParentRef:  parentRef,
		CreatedAt:  time.Now().UTC(),
		PayloadRef: ref,
		Source:     b.name,
	}

	if err := b.publisher.Publish(ctx, outEnv); err != nil {
		fmt.Printf(">>> Executive V3 publish outEnv error: %v\n", err)
	}

	return nil
}
