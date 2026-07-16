package learning

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"idun/core/memory"
)

// AggregationStrategy defines a pluggable, deterministic strategy for selecting,
// filtering, and ordering historical experiences within a bounded time window.
type AggregationStrategy interface {
	StrategyName() string
	Aggregate(ctx context.Context, store memory.Memory, start, end time.Time, schemaIDs []string) ([]memory.Record, error)
}

// ExperienceAggregator defines the gateway interface for harvesting, firewalling,
// and bounding historical experience records before learner synthesis.
type ExperienceAggregator interface {
	RegisterStrategy(strategy AggregationStrategy) error
	AggregateWindow(
		ctx context.Context,
		req *LearningRequest,
		snapshot *LearningStrategySnapshot,
	) (*AggregationSummary, ReplayMetadata, error)
}

// DefaultAggregator is the concrete thread-safe implementation of ExperienceAggregator.
type DefaultAggregator struct {
	mu         sync.RWMutex
	strategies map[string]AggregationStrategy
	store      memory.Memory
}

// NewDefaultAggregator initializes a new DefaultAggregator with default strategies registered.
func NewDefaultAggregator(store memory.Memory) *DefaultAggregator {
	a := &DefaultAggregator{
		strategies: make(map[string]AggregationStrategy),
		store:      store,
	}
	_ = a.RegisterStrategy(&CognitivePerformanceStrategy{})
	_ = a.RegisterStrategy(&RecentWindowStrategy{})
	_ = a.RegisterStrategy(&DomainSchemaStrategy{})
	return a
}

// RegisterStrategy registers or replaces a pluggable aggregation strategy.
func (a *DefaultAggregator) RegisterStrategy(strategy AggregationStrategy) error {
	if strategy == nil || strategy.StrategyName() == "" {
		return fmt.Errorf("%w: invalid strategy", ErrValidationFailed)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.strategies[strategy.StrategyName()] = strategy
	return nil
}

// AggregateWindow fetches, filters, bounds, and orders historical experiences for learning.
// It strictly enforces the COGNITIVE_PERFORMANCE partition, rejecting LEARNING_DIAGNOSTICS in code.
func (a *DefaultAggregator) AggregateWindow(
	ctx context.Context,
	req *LearningRequest,
	snapshot *LearningStrategySnapshot,
) (*AggregationSummary, ReplayMetadata, error) {
	if req == nil || snapshot == nil {
		return nil, ReplayMetadata{}, fmt.Errorf("%w: req and snapshot cannot be nil", ErrValidationFailed)
	}
	if err := req.Validate(); err != nil {
		return nil, ReplayMetadata{}, fmt.Errorf("invalid learning request: %w", err)
	}
	if err := snapshot.Validate(); err != nil {
		return nil, ReplayMetadata{}, fmt.Errorf("invalid strategy snapshot: %w", err)
	}

	// Determine governing policy parameters
	strategyName := "COGNITIVE_PERFORMANCE_DEFAULT"
	var policyID string
	maxArtifacts := uint32(500)
	maxMemoryBytes := uint64(MaxPayloadBytes)
	ordering := OrderingStrategyChronologicalAsc

	if snapshot.AggregationPolicy != nil {
		strategyName = snapshot.AggregationPolicy.Strategy
		policyID = snapshot.AggregationPolicy.PolicyID
		maxArtifacts = snapshot.AggregationPolicy.MaximumArtifacts
		maxMemoryBytes = snapshot.AggregationPolicy.MaximumMemoryBytes
		ordering = snapshot.AggregationPolicy.OrderingStrategy
	}

	a.mu.RLock()
	strategy, ok := a.strategies[strategyName]
	if !ok {
		// Fallback to default if custom strategy not found
		strategy, ok = a.strategies["COGNITIVE_PERFORMANCE_DEFAULT"]
	}
	a.mu.RUnlock()

	if !ok || strategy == nil {
		return nil, ReplayMetadata{}, fmt.Errorf("%w: aggregation strategy %q not found", ErrCapabilityUnavailable, strategyName)
	}

	schemaIDs := []string{req.DomainSchemaID}
	switch req.DomainSchemaID {
	case "idun.reasoning.strategy.v1":
		schemaIDs = append(schemaIDs, "idun.reasoning.trace.v1", "idun.reflection.report.v1")
	case "idun.planning.strategy.v1":
		schemaIDs = append(schemaIDs, "idun.planning.trace.v1", "idun.reflection.report.v1")
	case "idun.decision.weights.v1":
		schemaIDs = append(schemaIDs, "idun.decision.trace.v1", "idun.reflection.report.v1")
	case "idun.episodic.consolidation.v1":
		schemaIDs = append(schemaIDs, "idun.episodic.trace.v1", "idun.reflection.report.v1")
	}
	var rawRecords []memory.Record
	var err error

	if a.store != nil {
		rawRecords, err = strategy.Aggregate(ctx, a.store, req.TimeWindowStart, req.TimeWindowEnd, schemaIDs)
		if err != nil {
			return nil, ReplayMetadata{}, fmt.Errorf("strategy aggregation failed: %w", err)
		}
	}

	// =========================================================================
	// MANDATORY INGESTION FIREWALL
	// Strictly filter out any record whose type or payload indicates LEARNING_DIAGNOSTICS.
	// Only ingest COGNITIVE_PERFORMANCE.
	// =========================================================================
	cleanRecords := make([]memory.Record, 0, len(rawRecords))
	var currentBytes uint64

	for _, rec := range rawRecords {
		if isLearningDiagnostics(rec) {
			continue
		}
		// Enforce memory bounds
		recSize := uint64(len(rec.Payload))
		if currentBytes+recSize > maxMemoryBytes {
			break
		}
		cleanRecords = append(cleanRecords, rec)
		currentBytes += recSize
	}

	// =========================================================================
	// DETERMINISTIC ORDERING
	// =========================================================================
	switch ordering {
	case OrderingStrategyChronologicalDesc:
		sort.SliceStable(cleanRecords, func(i, j int) bool {
			if cleanRecords[i].CreatedAt.Equal(cleanRecords[j].CreatedAt) {
				return cleanRecords[i].ID > cleanRecords[j].ID
			}
			return cleanRecords[i].CreatedAt.After(cleanRecords[j].CreatedAt)
		})
	case OrderingStrategyDomainPriority:
		domainPrio := make(map[string]float64)
		if snapshot.AggregationPolicy != nil {
			domainPrio = snapshot.AggregationPolicy.DomainPriorities
		}
		sort.SliceStable(cleanRecords, func(i, j int) bool {
			wI := domainPrio[cleanRecords[i].Type]
			wJ := domainPrio[cleanRecords[j].Type]
			if wI == wJ {
				if cleanRecords[i].CreatedAt.Equal(cleanRecords[j].CreatedAt) {
					return cleanRecords[i].ID < cleanRecords[j].ID
				}
				return cleanRecords[i].CreatedAt.Before(cleanRecords[j].CreatedAt)
			}
			return wI > wJ
		})
	default: // OrderingStrategyChronologicalAsc
		sort.SliceStable(cleanRecords, func(i, j int) bool {
			if cleanRecords[i].CreatedAt.Equal(cleanRecords[j].CreatedAt) {
				return cleanRecords[i].ID < cleanRecords[j].ID
			}
			return cleanRecords[i].CreatedAt.Before(cleanRecords[j].CreatedAt)
		})
	}

	// Enforce artifact count bounds
	if uint32(len(cleanRecords)) > maxArtifacts {
		cleanRecords = cleanRecords[:maxArtifacts]
	}

	// =========================================================================
	// CRYPTOGRAPHIC LINEAGE (SourceArtifactHash)
	// SHA-256 Merkle/root calculation across sorted ingested records
	// =========================================================================
	hasher := sha256.New()
	for _, rec := range cleanRecords {
		hasher.Write([]byte(rec.ID))
		hasher.Write([]byte(rec.Type))
		hasher.Write(rec.Payload)
	}
	sourceHash := SourceArtifactHash(hex.EncodeToString(hasher.Sum(nil)))

	summary := &AggregationSummary{
		SummaryID:              fmt.Sprintf("sum-%d", time.Now().UnixNano()),
		TimeWindowStart:        req.TimeWindowStart,
		TimeWindowEnd:          req.TimeWindowEnd,
		TotalArtifactsIngested: len(cleanRecords),
		SourceArtifactHash:     sourceHash,
		DomainSchemaIDs:        schemaIDs,
		AggregationPolicyID:    policyID,
		Records:                cleanRecords,
	}

	lineage := ReplayMetadata{
		LearningFingerprint: LearningFingerprint(snapshot.Capabilities.CapabilityFingerprint),
		PolicyFingerprint:   PolicyFingerprint(snapshot.ActiveProfile.PolicyFingerprint),
		SourceArtifactHash:  sourceHash,
	}

	return summary, lineage, nil
}

// isLearningDiagnostics strictly checks if a record belongs to LEARNING_DIAGNOSTICS.
func isLearningDiagnostics(rec memory.Record) bool {
	// 1. Check record type
	if rec.Type == string(TopicLearningTraces) || rec.Type == "LEARNING_DIAGNOSTICS" || rec.Type == "idun.learning.trace.v1" {
		return true
	}
	// 2. Inspect payload JSON if present
	if len(rec.Payload) > 0 {
		var meta struct {
			Category string `json:"category"`
		}
		if err := json.Unmarshal(rec.Payload, &meta); err == nil {
			if meta.Category == "LEARNING_DIAGNOSTICS" {
				return true
			}
		}
	}
	return false
}

// =============================================================================
// Pluggable Strategies
// =============================================================================

// CognitivePerformanceStrategy collects records within the window and enforces COGNITIVE_PERFORMANCE.
type CognitivePerformanceStrategy struct{}

func (s *CognitivePerformanceStrategy) StrategyName() string {
	return "COGNITIVE_PERFORMANCE_DEFAULT"
}

func (s *CognitivePerformanceStrategy) Aggregate(ctx context.Context, store memory.Memory, start, end time.Time, schemaIDs []string) ([]memory.Record, error) {
	if store == nil {
		return nil, nil
	}
	var out []memory.Record
	for _, schemaID := range schemaIDs {
		records, err := store.ListRecordsByType(schemaID)
		if err != nil {
			continue
		}
		for _, rec := range records {
			if (rec.CreatedAt.Equal(start) || rec.CreatedAt.After(start)) && (rec.CreatedAt.Equal(end) || rec.CreatedAt.Before(end)) {
				if !isLearningDiagnostics(rec) {
					out = append(out, rec)
				}
			}
		}
	}
	return out, nil
}

// RecentWindowStrategy prioritizes newest experiences within the window.
type RecentWindowStrategy struct{}

func (s *RecentWindowStrategy) StrategyName() string {
	return "RECENT_WINDOW"
}

func (s *RecentWindowStrategy) Aggregate(ctx context.Context, store memory.Memory, start, end time.Time, schemaIDs []string) ([]memory.Record, error) {
	if store == nil {
		return nil, nil
	}
	var out []memory.Record
	for _, schemaID := range schemaIDs {
		records, err := store.ListRecordsByType(schemaID)
		if err != nil {
			continue
		}
		for _, rec := range records {
			if (rec.CreatedAt.Equal(start) || rec.CreatedAt.After(start)) && (rec.CreatedAt.Equal(end) || rec.CreatedAt.Before(end)) {
				if !isLearningDiagnostics(rec) {
					out = append(out, rec)
				}
			}
		}
	}
	// Sort newest first
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

// DomainSchemaStrategy queries strictly by requested schema ID.
type DomainSchemaStrategy struct{}

func (s *DomainSchemaStrategy) StrategyName() string {
	return "DOMAIN_SCHEMA"
}

func (s *DomainSchemaStrategy) Aggregate(ctx context.Context, store memory.Memory, start, end time.Time, schemaIDs []string) ([]memory.Record, error) {
	return (&CognitivePerformanceStrategy{}).Aggregate(ctx, store, start, end, schemaIDs)
}
