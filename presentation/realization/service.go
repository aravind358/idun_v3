package realization

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"idun/boundary"
	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/inference"
)

var (
	ErrServiceClosed    = errors.New("realization: service is closed")
	ErrWorkspaceMissing = errors.New("realization: workspace cannot be nil")
	ErrInferenceMissing = errors.New("realization: inference service cannot be nil")
	ErrStorerMissing    = errors.New("realization: payload storer cannot be nil")
)

type WorkspaceSubscriberPublisher interface {
	Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error)
	Publish(ctx context.Context, env communication.Envelope) error
}

type WorkspaceSubscription interface {
	Cancel() error
}

type PayloadStorer interface {
	Store(ctx context.Context, data []byte) (string, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// Service implements the stateless presentation realization subsystem.
type Service struct {
	mu      sync.RWMutex
	closed  atomic.Bool
	started atomic.Bool

	ws     WorkspaceSubscriberPublisher
	inf    inference.InferenceService
	storer PayloadStorer
	cfg    Config

	sub       WorkspaceSubscription
	svcCtx    context.Context    // lifecycle context — cancelled on Close
	svcCancel context.CancelFunc // cancels svcCtx
}

// NewService constructs a new stateless realization service instance.
func NewService(ws WorkspaceSubscriberPublisher, inf inference.InferenceService, storer PayloadStorer, cfg Config) (*Service, error) {
	if ws == nil {
		return nil, ErrWorkspaceMissing
	}
	if inf == nil {
		return nil, ErrInferenceMissing
	}
	if storer == nil {
		return nil, ErrStorerMissing
	}

	return &Service{
		ws:     ws,
		inf:    inf,
		storer: storer,
		cfg:    cfg,
	}, nil
}

// Name returns the canonical component name.
func (s *Service) Name() string {
	return "Presentation.LanguageRealization"
}

// Start boots the realization service and subscribes to TopicActionExecution.
// The provided ctx is the kernel lifecycle context (context.Background() in production).
// It is stored so that inference HTTP requests are scoped to the service lifetime,
// independent of the cognitive pipeline's per-request deadlines.
func (s *Service) Start(ctx context.Context) error {
	if s.closed.Load() {
		return ErrServiceClosed
	}
	if s.started.Swap(true) {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	svcCtx, svcCancel := context.WithCancel(ctx)

	sub, err := s.ws.Subscribe(
		communication.TopicActionExecution,
		s.Name(),
		s.handleExecutionEnvelope,
	)
	if err != nil {
		svcCancel()
		s.started.Store(false)
		return fmt.Errorf("realization: failed to subscribe to TopicActionExecution: %w", err)
	}

	s.mu.Lock()
	s.sub = sub
	s.svcCtx = svcCtx
	s.svcCancel = svcCancel
	s.mu.Unlock()

	return nil
}

// Close gracefully shuts down the realization service and unsubscribes.
func (s *Service) Close() error {
	if s.closed.Swap(true) {
		return nil
	}

	s.mu.Lock()
	sub := s.sub
	s.sub = nil
	cancel := s.svcCancel
	s.svcCancel = nil
	s.mu.Unlock()

	// Cancel the lifecycle context first so any in-flight inference call is unblocked.
	if cancel != nil {
		cancel()
	}
	if sub != nil {
		_ = sub.Cancel()
	}
	return nil
}

// handleExecutionEnvelope processes inbound envelopes on TopicActionExecution,
// runs them through local model realization via InferenceService, and publishes
// the polished text back to the Workspace without retaining any memory.
func (s *Service) handleExecutionEnvelope(ctx context.Context, env communication.Envelope) error {
	if s.closed.Load() || env.Source == s.Name() || env.PayloadRef == "" {
		return nil
	}

	devLog("Language Realization", "Received CommunicationMessage payload")

	data, err := s.storer.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return fmt.Errorf("realization: failed to retrieve payload %s: %w", env.PayloadRef, err)
	}

	commMsg, err := boundary.Unmarshal(data)
	if err != nil || commMsg == nil {
		return fmt.Errorf("realization: failed to unmarshal CommunicationMessage from payload: %w", err)
	}

	if commMsg.ParentRef == "" {
		commMsg.ParentRef = env.ParentRef
	}
	if commMsg.ParentRef == "" {
		commMsg.ParentRef = env.ID
	}
	if commMsg.ResponseID == "" {
		commMsg.ResponseID = "realized-resp-" + env.ID
	}

	devLog("Language Realization", "Building realization prompt")
	prompt := BuildRealizationPrompt(commMsg)

	// Resolve the service lifecycle context before any I/O. The pipeline ctx
	// carries Planning's 500ms per-request SLA, which will be expired long
	// before Ollama returns. All operations in this handler must use baseCtx.
	s.mu.RLock()
	baseCtx := s.svcCtx
	s.mu.RUnlock()
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	promptRef, err := s.storer.Store(baseCtx, []byte(prompt))
	if err != nil {
		return fmt.Errorf("realization: failed to store prompt: %w", err)
	}

	// Derive the inference context from the service lifecycle context, not from the
	// cognitive pipeline envelope context. The envelope ctx carries Planning's short
	// per-request SLA (500ms) which is already expired by this point. The service
	// lifecycle context is cancelled only when the service itself shuts down.
	execCtx, cancel := context.WithTimeout(baseCtx, s.cfg.Timeout)
	defer cancel()

	req := inference.InferenceRequest{
		ModelID:  s.cfg.ModelID,
		InputRef: promptRef,
		Modality: inference.ModalityText,
		Budget:   "REFLEXIVE",
		Hints: inference.ExecutionHints{
			OutputDetailHint: "standard",
			BypassCache:      s.cfg.BypassCache,
		},
		CallerID: s.Name(),
	}

	devLog("Language Realization", "Calling InferenceService")
	res, err := s.inf.Execute(execCtx, req)
	if err != nil {
		devLog("Language Realization", "Inference execution failed", fmt.Sprintf("%v", err))
		return fmt.Errorf("realization: inference execution failed: %w", err)
	}

	realizedText := res.OutputRef
	if outBytes, readErr := s.storer.Retrieve(baseCtx, res.OutputRef); readErr == nil && len(outBytes) > 0 {
		realizedText = string(outBytes)
	}

	realized := RealizedOutput{
		OutputID:         "rlz-" + time.Now().UTC().Format("20060102150405.000000"),
		SourceResponseID: commMsg.ResponseID,
		ParentRef:        commMsg.ParentRef,
		RealizedText:     realizedText,
		CreatedAt:        time.Now().UTC(),
	}

	realizedBytes, err := json.Marshal(realized)
	if err != nil {
		return fmt.Errorf("realization: failed to marshal output: %w", err)
	}

	realizedRef, err := s.storer.Store(baseCtx, realizedBytes)
	if err != nil {
		return fmt.Errorf("realization: failed to store realized output: %w", err)
	}

	outEnv, err := communication.NewEnvelopeBuilder().
		WithSource(s.Name()).
		WithTopic(communication.TopicActionExecution).
		WithParentRef(commMsg.ParentRef).
		WithPayloadRef(realizedRef).
		WithModality("text").
		Build()
	if err != nil {
		return fmt.Errorf("realization: failed to build output envelope: %w", err)
	}

	err = s.ws.Publish(baseCtx, outEnv)
	if err == nil {
		devLog("Language Realization", "Published realized output")
	} else {
		devLog("Language Realization", "Publish failed", fmt.Sprintf("%v", err))
	}
	return err
}
