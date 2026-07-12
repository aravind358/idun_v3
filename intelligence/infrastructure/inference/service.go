// Package inference implements the Shared Inference Service for IDUN Intelligence Infrastructure.
package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"idun/core/logger"
	"idun/core/storage"
	"idun/intelligence/infrastructure/registry"
)

// Service is the concrete, thread-safe implementation of InferenceService.
type Service struct {
	mu       sync.RWMutex
	closed   bool
	resolver registry.Resolver
	store    storage.Store
	log      logger.Writer

	// Telemetry counters
	activeWorkers    int64
	totalExecutions  int64
	failedExecutions int64
	cacheHits        int64

	// Priority queue depths
	reflexiveQueueDepth    int64
	standardQueueDepth     int64
	deliberativeQueueDepth int64
}

// Option configures functional options for Service construction.
type Option func(*Service)

// WithResolver injects a model registry Resolver.
func WithResolver(r registry.Resolver) Option {
	return func(s *Service) {
		s.resolver = r
	}
}

// WithStorage injects a storage Store for content-addressed payloads.
func WithStorage(st storage.Store) Option {
	return func(s *Service) {
		s.store = st
	}
}

// WithLogger injects a structured logger.
func WithLogger(log logger.Writer) Option {
	return func(s *Service) {
		s.log = log
	}
}

// NewService constructs a new InferenceService instance.
func NewService(opts ...Option) *Service {
	s := &Service{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Name returns the canonical Kernel Component name.
func (s *Service) Name() string {
	return "Intelligence.Infrastructure.Inference"
}

// Start boots the Inference Service.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrServiceClosed
	}
	if s.log != nil {
		s.log.Info("Inference service started", logger.Field{Key: "component", Value: s.Name()})
	}
	return nil
}

// Close gracefully shuts down the Inference Service.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.log != nil {
		s.log.Info("Inference service shut down", logger.Field{Key: "component", Value: s.Name()})
	}
	return nil
}

// budgetTimeout returns SLA timeout duration for a budget tier.
func budgetTimeout(budget string) time.Duration {
	switch budget {
	case "REFLEXIVE":
		return 50 * time.Millisecond
	case "STANDARD":
		return 500 * time.Millisecond
	case "DELIBERATIVE":
		return 10 * time.Second
	default:
		return 500 * time.Millisecond
	}
}

// incrementQueueDepth increments queue depth counter for budget tier.
func (s *Service) incrementQueueDepth(budget string, delta int64) {
	switch budget {
	case "REFLEXIVE":
		atomic.AddInt64(&s.reflexiveQueueDepth, delta)
	case "STANDARD":
		atomic.AddInt64(&s.standardQueueDepth, delta)
	case "DELIBERATIVE":
		atomic.AddInt64(&s.deliberativeQueueDepth, delta)
	default:
		atomic.AddInt64(&s.standardQueueDepth, delta)
	}
}

// cacheKey computes a deterministic hash key for exact content-addressed caching.
func cacheKey(req InferenceRequest) string {
	raw := fmt.Sprintf("inf|%s|%s|%s|%f|%d|%s",
		req.ModelID, req.InputRef, req.Modality,
		req.Hints.ExploratoryVariance, req.Hints.ComputeBudgetUnits, req.Hints.OutputDetailHint,
	)
	hash := sha256.Sum256([]byte(raw))
	return "inference/cache/" + hex.EncodeToString(hash[:])
}

// Execute runs synchronous or bounded inference and returns immutable storage references.
func (s *Service) Execute(ctx context.Context, req InferenceRequest) (InferenceResult, error) {
	if ctx == nil {
		return InferenceResult{}, ErrInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		atomic.AddInt64(&s.failedExecutions, 1)
		return InferenceResult{}, ErrInferenceCancelled
	}

	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()
	if closed {
		atomic.AddInt64(&s.failedExecutions, 1)
		return InferenceResult{}, ErrServiceClosed
	}

	if req.ModelID == "" || req.InputRef == "" {
		atomic.AddInt64(&s.failedExecutions, 1)
		return InferenceResult{}, ErrInvalidRequest
	}

	// Apply SLA timeout budget
	sla := budgetTimeout(req.Budget)
	execCtx, cancel := context.WithTimeout(ctx, sla)
	defer cancel()

	s.incrementQueueDepth(req.Budget, 1)
	defer s.incrementQueueDepth(req.Budget, -1)

	// Check content-addressed cache
	ck := cacheKey(req)
	if s.store != nil {
		if cachedOut, err := s.store.Read(ck); err == nil && len(cachedOut) > 0 {
			atomic.AddInt64(&s.cacheHits, 1)
			atomic.AddInt64(&s.totalExecutions, 1)
			return InferenceResult{
				OutputRef:         string(cachedOut),
				ModelID:           req.ModelID,
				BackendID:         "cache",
				ComputeUnits:      0,
				ExecutionDuration: 0,
				Cached:            true,
			}, nil
		}
	}

	// Resolve backend from registry
	if s.resolver == nil {
		atomic.AddInt64(&s.failedExecutions, 1)
		return InferenceResult{}, ErrBackendFailure
	}

	bd, err := s.resolver.Resolve(execCtx, registry.ModelID(req.ModelID))
	if err != nil {
		atomic.AddInt64(&s.failedExecutions, 1)
		return InferenceResult{}, fmt.Errorf("resolve backend: %w", err)
	}

	atomic.AddInt64(&s.activeWorkers, 1)
	defer atomic.AddInt64(&s.activeWorkers, -1)

	start := time.Now()

	// Execute simulated backend computation on input payload
	var inputBytes []byte
	if s.store != nil {
		inputBytes, _ = s.store.Read(req.InputRef)
	}
	if len(inputBytes) == 0 {
		inputBytes = []byte(req.InputRef)
	}

	// Verify cancellation before completion
	if err := execCtx.Err(); err != nil {
		atomic.AddInt64(&s.failedExecutions, 1)
		if errors.Is(err, context.DeadlineExceeded) {
			return InferenceResult{}, ErrInferenceTimeout
		}
		return InferenceResult{}, ErrInferenceCancelled
	}

	// Generate deterministic output representation
	outHash := sha256.Sum256(append(inputBytes, []byte(bd.ID)...))
	outputRef := "storage://inference/output/" + hex.EncodeToString(outHash[:])

	if s.store != nil {
		_ = s.store.Write(outputRef, []byte(fmt.Sprintf("inference_output:%s:%s", req.ModelID, string(inputBytes))))
		_ = s.store.Write(ck, []byte(outputRef))
	}

	dur := time.Since(start)
	atomic.AddInt64(&s.totalExecutions, 1)

	if s.log != nil {
		s.log.Info("Executed inference request",
			logger.Field{Key: "modelID", Value: req.ModelID},
			logger.Field{Key: "backendID", Value: bd.ID},
			logger.Field{Key: "outputRef", Value: outputRef},
			logger.Field{Key: "duration_ms", Value: dur.Milliseconds()},
		)
	}

	return InferenceResult{
		OutputRef:         outputRef,
		ModelID:           req.ModelID,
		BackendID:         bd.ID,
		ComputeUnits:      1,
		ExecutionDuration: dur,
		Cached:            false,
	}, nil
}

// ExecuteStream runs streaming inference, delivering output chunks to stream channel.
func (s *Service) ExecuteStream(ctx context.Context, req InferenceRequest, stream chan<- StreamChunk) error {
	if stream == nil {
		return ErrInvalidRequest
	}
	res, err := s.Execute(ctx, req)
	if err != nil {
		stream <- StreamChunk{Error: err, Done: true}
		return err
	}
	stream <- StreamChunk{ChunkRef: res.OutputRef, Done: false}
	stream <- StreamChunk{Done: true}
	return nil
}

// GetTelemetry returns operational telemetry snapshot.
func (s *Service) GetTelemetry() TelemetrySnapshot {
	qDepths := map[string]int{
		"REFLEXIVE":    int(atomic.LoadInt64(&s.reflexiveQueueDepth)),
		"STANDARD":     int(atomic.LoadInt64(&s.standardQueueDepth)),
		"DELIBERATIVE": int(atomic.LoadInt64(&s.deliberativeQueueDepth)),
	}
	return TelemetrySnapshot{
		ActiveWorkers:    int(atomic.LoadInt64(&s.activeWorkers)),
		QueueDepths:      qDepths,
		TotalExecutions:  atomic.LoadInt64(&s.totalExecutions),
		FailedExecutions: atomic.LoadInt64(&s.failedExecutions),
		CacheHits:        atomic.LoadInt64(&s.cacheHits),
	}
}

// Ensure Service implements InferenceService and TelemetryProvider.
var (
	_ InferenceService  = (*Service)(nil)
	_ TelemetryProvider = (*Service)(nil)
)
