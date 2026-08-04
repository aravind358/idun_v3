package v3

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"idun/core/storage"
	"idun/intelligence/communication"
	planningv3 "idun/intelligence/planning/v3"
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

// WorkspaceBridge wraps the V3 Decision Orchestrator.
type WorkspaceBridge struct {
	orchestrator *Orchestrator
	storer       PayloadStorer
	publisher    WorkspacePublisher
	subscriber   WorkspaceSubscriber
	name         string

	mu  sync.Mutex
	sub WorkspaceSubscription
}

// NewWorkspaceBridge creates a bridge for Decision V3.
func NewWorkspaceBridge(orchestrator *Orchestrator, storer PayloadStorer, publisher WorkspacePublisher, subscriber WorkspaceSubscriber) *WorkspaceBridge {
	return &WorkspaceBridge{
		orchestrator: orchestrator,
		storer:       storer,
		publisher:    publisher,
		subscriber:   subscriber,
		name:         "Intelligence.DecisionV3",
	}
}

// Start subscribes the Decision module to TopicCandidatePlans.
func (b *WorkspaceBridge) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriber != nil && b.sub == nil {
		sub, err := b.subscriber.Subscribe(communication.TopicCandidatePlans, b.name, b.HandleCandidatePlan)
		if err != nil {
			return err
		}
		b.sub = sub
	}
	return nil
}

// Close unsubscribes the Decision module.
func (b *WorkspaceBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.sub != nil {
		_ = b.sub.Cancel()
		b.sub = nil
	}
	return nil
}

// HandleCandidatePlan processes an incoming ExecutionPlan from TopicCandidatePlans.
func (b *WorkspaceBridge) HandleCandidatePlan(ctx context.Context, env communication.Envelope) error {
	fmt.Printf(">>> Decision V3 HandleCandidatePlan invoked! Envelope ID: %s\n", env.ID)
	if env.PayloadRef == "" {
		return fmt.Errorf("decision_v3: payload ref missing")
	}

	data, err := b.storer.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return err
	}

	plan := &planningv3.ExecutionPlan{}
	if err := json.Unmarshal(data, plan); err != nil {
		return err
	}

	// For integration, interp and reasonCtx are passed as nil since the V3 Orchestrator
	// mock implementation only relies on the ExecutionPlan.
	decisionRec, err := b.orchestrator.Decide(ctx, nil, nil, plan)
	if err != nil {
		return err
	}

	resData, err := json.Marshal(decisionRec)
	if err != nil {
		return err
	}

	meta := storage.ArtifactIndexMeta{
		Type:    "DecisionRecord",
		Version: "3.0",
	}
	ref, err := b.storer.StoreArtifact(ctx, string(decisionRec.ArtifactID()), meta, resData)
	if err != nil {
		return err
	}

	parentRef := env.ParentRef
	if parentRef == "" {
		parentRef = env.ID
	}

	outEnv := communication.Envelope{
		ID:         string(decisionRec.ArtifactID()),
		Topic:      communication.TopicEvaluatedOptions,
		ParentRef:  parentRef,
		CreatedAt:  time.Now().UTC(),
		PayloadRef: ref,
		Source:     b.name,
	}

	return b.publisher.Publish(ctx, outEnv)
}
