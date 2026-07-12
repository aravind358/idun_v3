// Package registry implements the Model & Capability Registry for IDUN Intelligence Infrastructure.
package registry

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"idun/core/logger"
)

// Service is the thread-safe, production-grade implementation of ModelRegistry.
type Service struct {
	mu      sync.RWMutex
	closed  bool
	log     logger.Writer
	active  map[ModelID]BackendDescriptor
	history map[ModelID][]BackendDescriptor

	// Telemetry counters
	totalResolutions  int64
	failedResolutions int64
}

// Option configures functional options for Service construction.
type Option func(*Service)

// WithLogger injects a structured logger into the registry service.
func WithLogger(log logger.Writer) Option {
	return func(s *Service) {
		s.log = log
	}
}

// NewService constructs a new ModelRegistry service instance.
func NewService(opts ...Option) *Service {
	s := &Service{
		active:  make(map[ModelID]BackendDescriptor),
		history: make(map[ModelID][]BackendDescriptor),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Name returns the canonical Kernel Component name.
func (s *Service) Name() string {
	return "Intelligence.Infrastructure.Registry"
}

// Start boots the Registry Service.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrRegistryClosed
	}
	if s.log != nil {
		s.log.Info("Registry service started", logger.Field{Key: "component", Value: s.Name()})
	}
	return nil
}

// Close gracefully shuts down the Registry Service.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.log != nil {
		s.log.Info("Registry service shut down", logger.Field{Key: "component", Value: s.Name()})
	}
	return nil
}

// checkContext validates whether the caller's context is cancelled.
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

// validateDescriptor checks required fields on a BackendDescriptor.
func validateDescriptor(bd BackendDescriptor) error {
	if bd.ID == "" || bd.DriverScheme == "" || bd.Endpoint == "" || bd.Version == "" {
		return ErrInvalidDescriptor
	}
	if bd.MaxConcurrency < 1 {
		return ErrInvalidDescriptor
	}
	return nil
}

// Register binds a logical ModelID to a physical BackendDescriptor atomically.
func (s *Service) Register(ctx context.Context, modelID ModelID, backend BackendDescriptor) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if modelID == "" {
		return ErrInvalidDescriptor
	}
	if err := validateDescriptor(backend); err != nil {
		return err
	}

	bd := backend.Clone()
	if bd.Health == "" {
		bd.Health = HealthHealthy
	}
	if bd.RegisteredAt.IsZero() {
		bd.RegisteredAt = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrRegistryClosed
	}

	s.active[modelID] = bd
	s.history[modelID] = append(s.history[modelID], bd)

	if s.log != nil {
		s.log.Info("Registered model backend",
			logger.Field{Key: "modelID", Value: string(modelID)},
			logger.Field{Key: "backendID", Value: bd.ID},
			logger.Field{Key: "scheme", Value: bd.DriverScheme},
			logger.Field{Key: "version", Value: bd.Version},
		)
	}

	return nil
}

// Deregister removes a logical ModelID registration.
func (s *Service) Deregister(ctx context.Context, modelID ModelID) error {
	if err := checkContext(ctx); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrRegistryClosed
	}

	if _, exists := s.active[modelID]; !exists {
		return ErrModelNotFound
	}

	delete(s.active, modelID)
	delete(s.history, modelID)

	if s.log != nil {
		s.log.Info("Deregistered model", logger.Field{Key: "modelID", Value: string(modelID)})
	}
	return nil
}

// Resolve returns the active, healthy BackendDescriptor for a logical ModelID.
func (s *Service) Resolve(ctx context.Context, modelID ModelID) (BackendDescriptor, error) {
	if err := checkContext(ctx); err != nil {
		atomic.AddInt64(&s.failedResolutions, 1)
		return BackendDescriptor{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		atomic.AddInt64(&s.failedResolutions, 1)
		return BackendDescriptor{}, ErrRegistryClosed
	}

	bd, exists := s.active[modelID]
	if !exists {
		atomic.AddInt64(&s.failedResolutions, 1)
		return BackendDescriptor{}, ErrModelNotFound
	}

	if bd.Health == HealthUnhealthy {
		atomic.AddInt64(&s.failedResolutions, 1)
		return BackendDescriptor{}, ErrBackendUnavailable
	}

	atomic.AddInt64(&s.totalResolutions, 1)
	return bd.Clone(), nil
}

// Rollback reverts the active backend descriptor for modelID to a specific historical version.
func (s *Service) Rollback(ctx context.Context, modelID ModelID, version string) error {
	if err := checkContext(ctx); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrRegistryClosed
	}

	versions, exists := s.history[modelID]
	if !exists || len(versions) == 0 {
		return ErrModelNotFound
	}

	var target *BackendDescriptor
	for i := len(versions) - 1; i >= 0; i-- {
		if versions[i].Version == version {
			c := versions[i].Clone()
			target = &c
			break
		}
	}

	if target == nil {
		return ErrVersionNotFound
	}

	target.RegisteredAt = time.Now().UTC()
	s.active[modelID] = *target
	s.history[modelID] = append(s.history[modelID], *target)

	if s.log != nil {
		s.log.Info("Rolled back model backend",
			logger.Field{Key: "modelID", Value: string(modelID)},
			logger.Field{Key: "version", Value: version},
		)
	}

	return nil
}

// SetHealth updates the runtime health status of a registered model ID.
func (s *Service) SetHealth(ctx context.Context, modelID ModelID, health BackendHealth, reason string) error {
	if err := checkContext(ctx); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrRegistryClosed
	}

	bd, exists := s.active[modelID]
	if !exists {
		return ErrModelNotFound
	}

	bd.Health = health
	bd.HealthReason = reason
	s.active[modelID] = bd

	if s.log != nil {
		s.log.Info("Updated backend health",
			logger.Field{Key: "modelID", Value: string(modelID)},
			logger.Field{Key: "health", Value: string(health)},
			logger.Field{Key: "reason", Value: reason},
		)
	}

	return nil
}

// ListModels returns a snapshot of all currently active logical ModelIDs and descriptors.
func (s *Service) ListModels() map[ModelID]BackendDescriptor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[ModelID]BackendDescriptor, len(s.active))
	for id, bd := range s.active {
		out[id] = bd.Clone()
	}
	return out
}

// ListVersions returns all historical versions registered for a logical ModelID.
func (s *Service) ListVersions(modelID ModelID) []BackendDescriptor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions := s.history[modelID]
	out := make([]BackendDescriptor, len(versions))
	for i, bd := range versions {
		out[i] = bd.Clone()
	}
	return out
}

// GetTelemetry returns a snapshot of operational telemetry metrics.
func (s *Service) GetTelemetry() TelemetrySnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.active)
	healthy := 0
	unhealthy := 0
	for _, bd := range s.active {
		if bd.Health == HealthUnhealthy {
			unhealthy++
		} else {
			healthy++
		}
	}

	return TelemetrySnapshot{
		TotalRegisteredModels: total,
		HealthyModels:         healthy,
		UnhealthyModels:       unhealthy,
		TotalResolutions:      atomic.LoadInt64(&s.totalResolutions),
		FailedResolutions:     atomic.LoadInt64(&s.failedResolutions),
	}
}

// Ensure Service implements ModelRegistry at compile time.
var _ ModelRegistry = (*Service)(nil)
