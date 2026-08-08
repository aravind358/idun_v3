package world

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/executive/v3"
	"idun/intelligence/workspace"
)

// ============================================================================
// telemetryCollector — internal, thread-safe summary bookkeeping
// ============================================================================

type telemetryCollector struct {
	mu                  sync.Mutex
	totalInteractions   int64
	successfulResponses int64
	failedResponses     int64
	timeoutCount        int64
	droppedInputCount   int64
	totalInputLen       int64
	totalOutputLen      int64
	totalLatencyNs      int64
}

func (t *telemetryCollector) recordInput(inputLen int) {
	t.mu.Lock()
	t.totalInteractions++
	t.totalInputLen += int64(inputLen)
	t.mu.Unlock()
}

func (t *telemetryCollector) recordDropped() {
	t.mu.Lock()
	t.droppedInputCount++
	t.mu.Unlock()
}

func (t *telemetryCollector) recordResponse(status ResponseStatus, outputLen int, latency time.Duration) {
	t.mu.Lock()
	switch status {
	case ResponseStatusSuccess:
		t.successfulResponses++
	case ResponseStatusError, ResponseStatusTimeout:
		t.failedResponses++
		if status == ResponseStatusTimeout {
			t.timeoutCount++
		}
	}
	t.totalOutputLen += int64(outputLen)
	t.totalLatencyNs += latency.Nanoseconds()
	t.mu.Unlock()
}

func (t *telemetryCollector) snapshot() WorldSummary {
	t.mu.Lock()
	defer t.mu.Unlock()

	var avgInput, avgOutput float64
	var avgLatency time.Duration
	if t.totalInteractions > 0 {
		avgInput = float64(t.totalInputLen) / float64(t.totalInteractions)
	}
	successful := t.successfulResponses
	total := t.totalInteractions
	responded := successful + t.failedResponses
	if responded > 0 {
		avgOutput = float64(t.totalOutputLen) / float64(responded)
		avgLatency = time.Duration(t.totalLatencyNs / responded)
	}
	return WorldSummary{
		TotalInteractions:   total,
		SuccessfulResponses: t.successfulResponses,
		FailedResponses:     t.failedResponses,
		AverageLatency:      avgLatency,
		AverageInputLength:  avgInput,
		AverageOutputLength: avgOutput,
		TimeoutCount:        t.timeoutCount,
		DroppedInputCount:   t.droppedInputCount,
	}
}

// ============================================================================
// pendingEntry — tracks in-flight interactions
// ============================================================================

type pendingEntry struct {
	interaction *Interaction
	startTime   time.Time
}

// ============================================================================
// Service
// ============================================================================

// Service implements WorldService and is the sole concrete implementation of the
// World subsystem. It is fully event-driven: HandleInteraction publishes to the
// Global Workspace and returns immediately. The Workspace subscription callback
// delivers responses via OutputAdapter when Executive publishes to TopicActionExecution.
type Service struct {
	mu           sync.RWMutex
	cfg          Config
	policy       *WorldPolicyProfile
	capabilities *WorldCapabilities
	input        InputAdapter
	output       OutputAdapter // retained for Close() cleanup; egress is via dispatchOutput
	storer       PayloadStorer
	ws           workspace.Workspace
	telemetry    *telemetryCollector

	// pending maps EnvelopeID (= InteractionID) → in-flight entry
	pending map[string]*pendingEntry

	// subscriptions holds active Workspace subscriptions (for cleanup on Close)
	subscriptions []workspace.Subscription
	cancelFunc    context.CancelFunc
	doneCh        chan struct{}

	started atomic.Bool
	closed  atomic.Bool

	dispatchOutput func(ctx context.Context, interaction *Interaction, execResult *v3.ExecutionResult) error
}

// ServiceOption configures a Service instance at initialization time.
type ServiceOption func(*Service)

// WithConfig sets custom operational configuration for the World service.
func WithConfig(cfg Config) ServiceOption {
	return func(s *Service) {
		s.cfg = cfg
	}
}

// WithPolicy sets a custom WorldPolicyProfile for the World service.
// This is owned externally (by Runtime or Executive); World only consumes it read-only.
func WithPolicy(profile *WorldPolicyProfile) ServiceOption {
	return func(s *Service) {
		if profile != nil {
			s.policy = profile
		}
	}
}

// WithCapabilities sets custom WorldCapabilities for this deployment.
func WithCapabilities(caps *WorldCapabilities) ServiceOption {
	return func(s *Service) {
		if caps != nil {
			s.capabilities = caps
		}
	}
}

// SetOutputDispatcher sets the callback to dispatch raw ExecutionResults to the OutputManager.
// This is injected to avoid cyclic imports between world and world/output.
func (s *Service) SetOutputDispatcher(fn func(ctx context.Context, interaction *Interaction, execResult *v3.ExecutionResult) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dispatchOutput = fn
}


// NewService constructs a new World Service with all required dependencies.
//
// Parameters:
//   - ws: the Global Workspace for publishing Interaction envelopes and subscribing to responses
//   - input: the InputAdapter for receiving external input
//   - output: the OutputAdapter for presenting responses
//   - storer: the PayloadStorer for persisting large payloads to Core.Storage
//   - opts: functional options for Config, Policy, and Capabilities overrides
func NewService(
	ws workspace.Workspace,
	input InputAdapter,
	output OutputAdapter,
	storer PayloadStorer,
	opts ...ServiceOption,
) (*Service, error) {
	if ws == nil {
		return nil, ErrNilWorkspace
	}
	if input == nil {
		return nil, ErrNilInputAdapter
	}
	if output == nil {
		return nil, ErrNilOutputAdapter
	}
	if storer == nil {
		return nil, ErrNilPayloadStorer
	}

	svc := &Service{
		cfg:          DefaultConfig(),
		policy:       DefaultWorldPolicyProfile(),
		capabilities: DefaultWorldCapabilities(),
		input:        input,
		output:       output,
		storer:       storer,
		ws:           ws,
		telemetry:    &telemetryCollector{},
		pending:      make(map[string]*pendingEntry),
		doneCh:       make(chan struct{}),
	}
	for _, opt := range opts {
		opt(svc)
	}
	if err := svc.cfg.Validate(); err != nil {
		return nil, fmt.Errorf("world: invalid config: %w", err)
	}
	return svc, nil
}

// Name returns the canonical Kernel component name.
func (s *Service) Name() string {
	return "World.Service"
}

// GetPolicyProfile returns the immutable WorldPolicyProfile governing this service instance.
func (s *Service) GetPolicyProfile() *WorldPolicyProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.policy
}

// GetCapabilities returns the immutable WorldCapabilities for this deployment.
func (s *Service) GetCapabilities() *WorldCapabilities {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.capabilities
}

// GetSummary returns a bounded statistical snapshot of World interaction telemetry.
func (s *Service) GetSummary() WorldSummary {
	return s.telemetry.snapshot()
}

// Start boots the World Service lifecycle and wires Workspace subscriptions.
// Start is idempotent on the second call; it returns nil if already started.
func (s *Service) Start(ctx context.Context) error {
	if s.closed.Load() {
		return ErrServiceClosed
	}
	if s.started.Swap(true) {
		return nil // already started
	}

	sub, err := s.ws.Subscribe(
		communication.TopicActionExecution,
		"World.Service",
		s.handleResponseEnvelope,
	)
	if err != nil {
		s.started.Store(false)
		return fmt.Errorf("world: failed to subscribe to TopicActionExecution: %w", err)
	}

	ctxCancel, cancel := context.WithCancel(ctx)

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.cancelFunc = cancel
	s.mu.Unlock()

	go s.readLoop(ctxCancel)

	return nil
}

// Close gracefully shuts down the World Service, cancels Workspace subscriptions,
// and closes both input and output adapters.
func (s *Service) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil // already closed
	}

	s.mu.Lock()
	subs := s.subscriptions
	s.subscriptions = nil
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	if s.doneCh != nil {
		select {
		case <-s.doneCh:
		default:
			close(s.doneCh)
		}
	}
	s.mu.Unlock()

	for _, sub := range subs {
		_ = sub.Cancel()
	}

	_ = s.input.Close()
	_ = s.output.Close()
	return nil
}

func (s *Service) readLoop(ctx context.Context) {
	for !s.closed.Load() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		interaction, err := s.input.Receive(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return
			}
			s.telemetry.recordDropped()
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if interaction == nil {
			continue
		}
		trimmed := strings.ToLower(strings.TrimSpace(interaction.NormalizedInput))
		if trimmed == "exit" || trimmed == "quit" {
			_ = s.Close()
			return
		}
		fullInt, err := s.CreateInteraction(ctx, interaction.OriginalInput, interaction.SessionID, interaction.Origin, interaction.Modality)
		if err != nil {
			s.telemetry.recordDropped()
			continue
		}
		_ = s.HandleInteraction(ctx, fullInt)
	}
}

// Done returns a read-only channel that is closed when the World Service closes.
func (s *Service) Done() <-chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doneCh
}

// HandleInteraction is the primary event entry point for the World service.
// It normalizes the interaction, stores the payload, publishes an Envelope to the
// Global Workspace on TopicPerception, and returns immediately without blocking.
//
// Architecture: World is fully event-driven (Refinement 11). It never waits for
// Executive to respond. When Executive publishes a response on TopicActionExecution,
// handleResponseEnvelope is invoked asynchronously via the Workspace subscription.
func (s *Service) HandleInteraction(ctx context.Context, interaction *Interaction) error {
	if s.closed.Load() {
		return ErrServiceClosed
	}
	if !s.started.Load() {
		return ErrServiceNotStarted
	}
	if interaction == nil {
		return ErrNilInteraction
	}
	if err := interaction.Validate(); err != nil {
		s.telemetry.recordDropped()
		return fmt.Errorf("world: HandleInteraction validation failed: %w", err)
	}

	s.telemetry.recordInput(len(interaction.NormalizedInput))

	devLog("World", "Received input", fmt.Sprintf("%q", interaction.NormalizedInput))

	// Store the pending interaction for async response matching.
	s.mu.Lock()
	s.pending[interaction.InteractionID] = &pendingEntry{
		interaction: interaction,
		startTime:   time.Now(),
	}
	s.mu.Unlock()

	// Publish the interaction as a perception envelope on the Global Workspace.
	// World is content-blind: it does not interpret whether this is a user intent,
	// a tool request, or a system event. That interpretation belongs to Understanding.
	//
	// Note on TopicInteraction (Refinement 13):
	// The preferred future topic for Layer 3 would be TopicInteraction, allowing
	// Understanding to explicitly subscribe to all World inputs without sharing
	// the generic TopicPerception channel. This is documented as a post-Layer-1
	// evolution; for Layer 1 we use TopicPerception as the semantically closest
	// existing topic for raw external world input.
	env, err := communication.NewEnvelopeBuilder().
		WithID(interaction.InteractionID).
		WithSource("World.Service").
		WithTopic(communication.TopicPerception).
		WithPayloadRef(interaction.PayloadRef).
		WithModality(string(interaction.Modality)).
		WithConfidence(1.0).
		WithUrgency(0).
		Build()
	if err != nil {
		s.mu.Lock()
		delete(s.pending, interaction.InteractionID)
		s.mu.Unlock()
		s.telemetry.recordDropped()
		return fmt.Errorf("world: failed to build perception envelope: %w", err)
	}

	if err := s.ws.Publish(ctx, env); err != nil {
		s.mu.Lock()
		delete(s.pending, interaction.InteractionID)
		s.mu.Unlock()
		s.telemetry.recordDropped()
		return fmt.Errorf("world: failed to publish interaction to Workspace: %w", err)
	}

	devLog("World", "Published TopicPerception")

	return nil
}

// handleResponseEnvelope is the asynchronous Workspace subscription callback.
// It is invoked by the Workspace when Executive publishes a response on TopicActionExecution.
//
// O4 Architecture: World is the sole egress boundary.
// All output flows through the OutputManager pipeline (dispatchOutput).
// The legacy s.output.Send() path has been removed.
//
// This handler must be fast and non-blocking per Workspace contract.
func (s *Service) handleResponseEnvelope(ctx context.Context, env communication.Envelope) error {
	if s.closed.Load() {
		return nil
	}

	// World only handles responses from the Executive V3 pipeline.
	// Envelopes from other sources (Presentation.Router, etc.) are ignored now that
	// the Presentation subsystem has been disconnected.
	if env.Source != "Intelligence.ExecutiveV3" {
		return nil
	}

	// Match response to pending interaction via ParentRef.
	parentID := env.ParentRef
	if parentID == "" {
		return nil
	}

	s.mu.Lock()
	entry, ok := s.pending[parentID]
	if ok {
		delete(s.pending, parentID)
	}
	s.mu.Unlock()

	if !ok {
		// Not our interaction; another subscriber may own this.
		return nil
	}

	latency := time.Since(entry.startTime)
	interaction := entry.interaction

	devLog("World", "Received Executive response — dispatching to OutputManager")

	// Dispatch to the World Output pipeline (non-blocking — runs in a goroutine inside Dispatch).
	if s.dispatchOutput != nil && env.PayloadRef != "" {
		if raw, err := s.storer.Retrieve(ctx, env.PayloadRef); err == nil && len(raw) > 0 {
			var execResult v3.ExecutionResult
			if err := json.Unmarshal(raw, &execResult); err == nil {
				_ = s.dispatchOutput(ctx, interaction, &execResult)
			}
		}
	}

	s.telemetry.recordResponse(ResponseStatusSuccess, 0, latency)
	devLog("World", "Response dispatched to OutputManager.")
	return nil
}


// CreateInteraction is a helper that builds a validated Interaction from raw adapter input,
// applying the current WorldPolicyProfile normalization rules and computing all
// fingerprints and replay metadata. This is called by adapter drivers or test code.
func (s *Service) CreateInteraction(
	ctx context.Context,
	rawInput string,
	sessionID string,
	origin InteractionOrigin,
	modality Modality,
) (*Interaction, error) {
	s.mu.RLock()
	policy := s.policy
	caps := s.capabilities
	s.mu.RUnlock()

	if sessionID == "" {
		sessionID = s.cfg.DefaultSessionID
	}

	// Apply policy normalization.
	normalized, err := ApplyPolicy(rawInput, policy)
	if err != nil {
		s.telemetry.recordDropped()
		return nil, fmt.Errorf("world: CreateInteraction policy rejected input: %w", err)
	}

	// Compute deterministic interaction fingerprint (Refinement 9).
	fingerprint := ComputeInteractionFingerprint(rawInput, normalized, modality, policy.PolicyFingerprint)

	// Store payload via content-addressed storage (Refinement 12).
	// U8.5: Preserve raw input using Claim Check Pattern.
	payloadRef, storeErr := s.storer.Store(ctx, []byte(rawInput))
	if storeErr != nil {
		// Fall back to fingerprint as ref when storage fails.
		payloadRef = fingerprint
	}

	replayMeta := buildWorldReplayMetadata(fingerprint, policy.PolicyFingerprint, caps.CapabilityFingerprint, 0)

	return NewInteractionBuilder().
		WithSessionID(sessionID).
		WithOrigin(origin).
		WithModality(modality).
		WithOriginalInput(rawInput).
		WithNormalizedInput(normalized).
		WithPayloadRef(payloadRef).
		WithReplayMetadata(replayMeta).
		Build()
}
