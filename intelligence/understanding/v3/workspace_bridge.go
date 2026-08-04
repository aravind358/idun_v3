package v3

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"idun/boundary/perception"
	"idun/intelligence/communication"
)

// WorkspacePublisher defines the interface required to emit Envelopes to the Global Workspace.
type WorkspacePublisher interface {
	Publish(ctx context.Context, env communication.Envelope) error
}

// WorkspaceSubscription represents an active subscription.
type WorkspaceSubscription interface {
	Cancel() error
}

// WorkspaceSubscriber defines the interface for subscribing to Workspace topics.
type WorkspaceSubscriber interface {
	Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error)
}

// PayloadStorer defines the interface required to persist payloads to CAS storage.
type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// WorkspaceBridge wraps the synchronous Understanding V3 Orchestrator to participate
// in the asynchronous Global Workspace publish/subscribe flow.
// 
// Documentation on why this new bridge is required:
// The legacy V1 `understanding.Service` cannot be safely reused because it is tightly
// coupled to the deprecated `SemanticFrame` schema, V1 internal specialist interfaces,
// and V1 telemetry. Hijacking it to output V3 `SemanticInterpretation` would require
// rewriting the V1 service entirely, violating the "minimize new code" and "do not
// redesign" rules. This minimal adapter fulfills the requirement to bridge the synchronous
// V3 Orchestrator to the asynchronous Global Workspace.
type WorkspaceBridge struct {
	orchestrator *Orchestrator
	storer       PayloadStorer
	publisher    WorkspacePublisher
	subscriber   WorkspaceSubscriber
	name         string
	
	mu  sync.Mutex
	sub WorkspaceSubscription
}

// NewWorkspaceBridge creates a new WorkspaceBridge for Understanding V3.
func NewWorkspaceBridge(orchestrator *Orchestrator, storer PayloadStorer, publisher WorkspacePublisher, subscriber WorkspaceSubscriber) *WorkspaceBridge {
	return &WorkspaceBridge{
		orchestrator: orchestrator,
		storer:       storer,
		publisher:    publisher,
		subscriber:   subscriber,
		name:         "Intelligence.UnderstandingV3",
	}
}

// Start initiates the subscription to TopicPerception.
func (b *WorkspaceBridge) Start() error {
	fmt.Println(">>> Understanding V3 WorkspaceBridge Start() is called!")
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscriber != nil && b.sub == nil {
		sub, err := b.subscriber.Subscribe(communication.TopicPerception, b.name, b.HandlePerception)
		if err != nil {
			fmt.Printf(">>> Subscribe error: %v\n", err)
			return fmt.Errorf("understanding_v3: failed to subscribe to TopicPerception: %w", err)
		}
		fmt.Println(">>> Subscribed to TopicPerception successfully.")
		b.sub = sub
	} else {
		fmt.Printf(">>> Did not subscribe. b.subscriber=%v, b.sub=%v\n", b.subscriber != nil, b.sub != nil)
	}
	return nil
}

// Close cancels the active subscription.
func (b *WorkspaceBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.sub != nil {
		_ = b.sub.Cancel()
		b.sub = nil
	}
	return nil
}

// HandlePerception is the subscription callback for TopicPerception.
func (b *WorkspaceBridge) HandlePerception(ctx context.Context, env communication.Envelope) error {
	if env.ID == "" || env.Topic != communication.TopicPerception || env.PayloadRef == "" {
		return fmt.Errorf("understanding_v3: invalid perception envelope")
	}

	var rawText string
	if b.storer != nil {
		data, err := b.storer.Retrieve(ctx, env.PayloadRef)
		if err == nil && len(data) > 0 {
			rawText = string(data)
		} else {
			rawText = env.PayloadRef
		}
	} else {
		rawText = env.PayloadRef
	}

	fmt.Printf(">>> HandlePerception invoked! RawText: %q\n", rawText)

	// Convert communication.Envelope to perception.PerceptionEnvelope
	perceptionEnv, err := perception.NewBuilder().
		EnvelopeID(env.ID).
		ArtifactID(env.ID).
		ParentArtifactID(env.ParentRef).
		Version("3.0").
		Timestamp(time.Now()).
		InputSource(env.Source).
		InputType(env.PayloadModality).
		RawInput(rawText).
		Build()
	if err != nil {
		return fmt.Errorf("understanding_v3: failed to build perception envelope: %w", err)
	}

	// Execute synchronous V3 Orchestrator
	batch, err := b.orchestrator.Analyze(ctx, perceptionEnv)
	if err != nil {
		return fmt.Errorf("understanding_v3: orchestrator analysis failed: %w", err)
	}

	// Serialize UnderstandingBatch
	payloadBytes, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("understanding_v3: failed to marshal batch: %w", err)
	}

	payloadRef := string(payloadBytes)
	if b.storer != nil {
		if key, storeErr := b.storer.Store(ctx, payloadBytes); storeErr == nil {
			payloadRef = key
		}
	}

	parentID := env.ParentRef
	if parentID == "" {
		parentID = env.ID
	}

	// Publish to TopicUserIntent
	pubEnv := communication.Envelope{
		ID:              fmt.Sprintf("interp-%s", env.ID),
		ParentRef:       parentID,
		Source:          b.name,
		Topic:           communication.TopicUserIntent,
		RawConfidence:   1.0, // Batch confidence could be aggregated, assuming 1.0 for now
		PayloadRef:      payloadRef,
		PayloadModality: "semantic-interpretation",
		CreatedAt:       time.Now().UTC(),
	}

	if b.publisher != nil {
		if err := b.publisher.Publish(ctx, pubEnv); err != nil {
			return fmt.Errorf("understanding_v3: failed to publish interpretation: %w", err)
		}
	}

	return nil
}
