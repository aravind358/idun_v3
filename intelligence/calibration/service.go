package calibration

import (
	"sync"
	"time"

	"idun/intelligence/communication"
)

const defaultMaxHistoryPerKey = 200

// Service is the concrete, thread-safe implementation of the Epistemic Calibration System.
type Service struct {
	mu            sync.RWMutex
	closed        bool
	strategy      WeightStrategy
	maxHistory    int
	records       map[string][]AuditRecord
	weights       map[string]float64
	lastAuditedAt map[string]time.Time
}

// Option configures functional options for Service construction.
type Option func(*Service)

// WithMaxHistoryPerKey configures the maximum audit record history retained per source/topic key.
func WithMaxHistoryPerKey(limit int) Option {
	return func(s *Service) {
		if limit > 0 {
			s.maxHistory = limit
		}
	}
}

// WithStrategy configures a custom initial WeightStrategy.
func WithStrategy(strategy WeightStrategy) Option {
	return func(s *Service) {
		if strategy != nil {
			s.strategy = strategy
		}
	}
}

// NewService constructs a new Epistemic Calibration service.
func NewService(opts ...Option) *Service {
	s := &Service{
		strategy:      NewDefaultWeightStrategy(),
		maxHistory:    defaultMaxHistoryPerKey,
		records:       make(map[string][]AuditRecord),
		weights:       make(map[string]float64),
		lastAuditedAt: make(map[string]time.Time),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Name returns the canonical Kernel component name.
func (s *Service) Name() string {
	return "Intelligence.Calibration"
}

// Start boots the Calibration Service.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrServiceClosed
	}
	return nil
}

// Close gracefully shuts down the Calibration Service.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	s.records = make(map[string][]AuditRecord)
	s.weights = make(map[string]float64)
	s.lastAuditedAt = make(map[string]time.Time)
	return nil
}

// GetWeight retrieves the current calibration trust multiplier W_calib in [0.1, 1.5].
func (s *Service) GetWeight(source string, topic communication.TopicID) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := makeKey(source, topic)
	w, ok := s.weights[key]
	if !ok {
		return defaultWeight
	}
	return w
}

// CalibrateConfidence returns calibrated confidence clamped to [0.0, 1.0].
func (s *Service) CalibrateConfidence(source string, topic communication.TopicID, rawConfidence float64) float64 {
	w := s.GetWeight(source, topic)
	calibrated := rawConfidence * w
	if calibrated < 0.0 {
		return 0.0
	} else if calibrated > 1.0 {
		return 1.0
	}
	return calibrated
}

// CalibrateEnvelope computes the Calibrated Effective Priority (P_eff) for an Envelope.
func (s *Service) CalibrateEnvelope(env communication.Envelope, alpha, beta float64, totalBudget int) float64 {
	w := s.GetWeight(env.Source, env.Topic)
	return env.EffectivePriority(w, alpha, beta, totalBudget)
}

// RecordAudit records an empirical accuracy audit from Reflection or Learning.
func (s *Service) RecordAudit(record AuditRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrServiceClosed
	}

	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now().UTC()
	}

	key := makeKey(record.Source, record.Topic)
	history := s.records[key]
	history = append(history, record)
	if len(history) > s.maxHistory {
		history = history[len(history)-s.maxHistory:]
	}
	s.records[key] = history
	s.lastAuditedAt[key] = record.Timestamp

	// Recompute weight using current strategy
	s.weights[key] = s.strategy.ComputeWeight(record.Source, record.Topic, history)
	return nil
}

// SetWeightStrategy dynamically upgrades the weight calculation algorithm.
func (s *Service) SetWeightStrategy(strategy WeightStrategy) error {
	if strategy == nil {
		return ErrNilStrategy
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrServiceClosed
	}

	s.strategy = strategy
	for key, history := range s.records {
		if len(history) > 0 {
			rec := history[0]
			s.weights[key] = s.strategy.ComputeWeight(rec.Source, rec.Topic, history)
		}
	}
	return nil
}

// GetSnapshot retrieves the current calibration summary for a source and topic.
func (s *Service) GetSnapshot(source string, topic communication.TopicID) CalibrationSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := makeKey(source, topic)
	w, ok := s.weights[key]
	if !ok {
		w = defaultWeight
	}
	n := int64(len(s.records[key]))
	last := s.lastAuditedAt[key]

	return CalibrationSnapshot{
		Source:      source,
		Topic:       topic,
		Weight:      w,
		TotalAudits: n,
		LastAudited: last,
	}
}

func makeKey(source string, topic communication.TopicID) string {
	return source + "|" + string(topic)
}

// Ensure Service implements CalibrationService.
var _ CalibrationService = (*Service)(nil)
