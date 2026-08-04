package router

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"idun/capabilities"
	"idun/intelligence/communication"
	"idun/intelligence/executive/v3"
	"idun/presentation"
)

type WorkspaceSubscriberPublisher interface {
	Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error)
	Publish(ctx context.Context, env communication.Envelope) error
}

type WorkspaceSubscription interface {
	Cancel() error
}

type PayloadStorer interface {
	Retrieve(ctx context.Context, key string) ([]byte, error)
	Store(ctx context.Context, data []byte) (string, error)
}

// Service is the Response Type Router.
type Service struct {
	mu      sync.RWMutex
	closed  atomic.Bool
	started atomic.Bool

	ws      WorkspaceSubscriberPublisher
	storer  PayloadStorer
	engines map[capabilities.RealizationStrategy]presentation.RealizationEngine

	sub       WorkspaceSubscription
	svcCtx    context.Context
	svcCancel context.CancelFunc
}

// NewService creates a new Response Type Router.
func NewService(ws WorkspaceSubscriberPublisher, storer PayloadStorer, engines map[capabilities.RealizationStrategy]presentation.RealizationEngine) *Service {
	return &Service{
		ws:      ws,
		storer:  storer,
		engines: engines,
	}
}

func (s *Service) Name() string {
	return "Presentation.Router"
}

func (s *Service) Start(ctx context.Context) error {
	if s.closed.Load() {
		return fmt.Errorf("router: service is closed")
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
		return fmt.Errorf("router: failed to subscribe to TopicActionExecution: %w", err)
	}

	s.mu.Lock()
	s.sub = sub
	s.svcCtx = svcCtx
	s.svcCancel = svcCancel
	s.mu.Unlock()

	return nil
}

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

	if cancel != nil {
		cancel()
	}
	if sub != nil {
		_ = sub.Cancel()
	}
	return nil
}

func (s *Service) handleExecutionEnvelope(ctx context.Context, env communication.Envelope) error {
	if s.closed.Load() || env.Source != "Intelligence.ExecutiveV3" || env.PayloadRef == "" {
		return nil
	}

	data, err := s.storer.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return fmt.Errorf("router: failed to retrieve payload %s: %w", env.PayloadRef, err)
	}

	var execResult v3.ExecutionResult
	if err := json.Unmarshal(data, &execResult); err != nil {
		return fmt.Errorf("router: failed to unmarshal ExecutionResult: %w", err)
	}

	// For MVP, we process the first node result. 
	// A robust implementation would handle multiple nodes or aggregate them.
	if len(execResult.NodeResults()) == 0 {
		return nil
	}
	var nr v3.NodeResult
	for _, res := range execResult.NodeResults() {
		nr = res
		break
	}

	if nr.OutputRef == "" {
		return nil
	}

	outData, err := s.storer.Retrieve(ctx, nr.OutputRef)
	if err != nil {
		return fmt.Errorf("router: failed to retrieve capability result %s: %w", nr.OutputRef, err)
	}

	var capResult capabilities.CapabilityResult
	if err := json.Unmarshal(outData, &capResult); err != nil {
		// Might not be a CapabilityResult JSON (e.g. legacy fallback)
		return nil
	}

	engine, ok := s.engines[capResult.Realization]
	if !ok {
		return fmt.Errorf("router: no realization engine registered for strategy %v", capResult.Realization)
	}

	responseID := "resp-" + env.ID
	parentRef := env.ParentRef
	if parentRef == "" {
		parentRef = env.ID
	}

	realizedOut, err := engine.Realize(ctx, capResult, parentRef, responseID)
	if err != nil {
		return fmt.Errorf("router: realization engine failed: %w", err)
	}

	outBytes, err := json.Marshal(realizedOut)
	if err != nil {
		return fmt.Errorf("router: failed to marshal realized output: %w", err)
	}

	s.mu.RLock()
	baseCtx := s.svcCtx
	s.mu.RUnlock()
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	outRef, err := s.storer.Store(baseCtx, outBytes)
	if err != nil {
		return fmt.Errorf("router: failed to store realized output: %w", err)
	}

	outEnv, err := communication.NewEnvelopeBuilder().
		WithSource(s.Name()).
		WithTopic(communication.TopicActionExecution).
		WithParentRef(parentRef).
		WithPayloadRef(outRef).
		WithModality("text").
		Build()
	if err != nil {
		return fmt.Errorf("router: failed to build output envelope: %w", err)
	}

	return s.ws.Publish(baseCtx, outEnv)
}
