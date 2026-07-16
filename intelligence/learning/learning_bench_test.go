package learning

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/core/memory"
)

type benchMemory struct {
	mu      sync.RWMutex
	records []memory.Record
}

func newBenchMemory(n int) *benchMemory {
	m := &benchMemory{records: make([]memory.Record, 0, n)}
	now := time.Now()
	for i := 0; i < n; i++ {
		m.records = append(m.records, memory.Record{
			ID:        fmt.Sprintf("rec-cog-%d", i),
			Type:      "idun.reasoning.trace.v1",
			Payload:   []byte(`{"category":"COGNITIVE_PERFORMANCE","score":0.95}`),
			CreatedAt: now.Add(-30 * time.Minute),
		})
	}
	return m
}

func (m *benchMemory) CreateRecord(rec memory.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = append(m.records, rec)
	return nil
}
func (m *benchMemory) GetRecord(id string) (memory.Record, error)           { return memory.Record{}, nil }
func (m *benchMemory) UpdateRecord(rec memory.Record) error                 { return nil }
func (m *benchMemory) DeleteRecord(id string) error                         { return nil }
func (m *benchMemory) ListRecordsByType(t string) ([]memory.Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.records, nil
}
func (m *benchMemory) LinkRecords(from, to, relationship, creator string) error { return nil }
func (m *benchMemory) UnlinkRecords(from, to, relationship string) error        { return nil }
func (m *benchMemory) GetLinkedRecords(from, rel string) ([]memory.Record, error) { return nil, nil }
func (m *benchMemory) GetLinkedRecordsReverse(to, rel string) ([]memory.Record, error) { return nil, nil }

func BenchmarkAggregatorWindowing(b *testing.B) {
	store := newBenchMemory(50)
	agg := NewDefaultAggregator(store)
	ctx := context.Background()
	req := &LearningRequest{
		RequestID:         "bench-req-1",
		DomainSchemaID:    "idun.reasoning.trace.v1",
		PolicyFingerprint: "policy-fp-bench",
		TimeWindowStart:   time.Now().Add(-1 * time.Hour),
		TimeWindowEnd:     time.Now().Add(1 * time.Hour),
	}
	snap := &LearningStrategySnapshot{
		SnapshotID:        "snap-bench",
		SchemaVersion:     SchemaVersion,
		ActiveProfile:     DefaultLearningPolicyProfile(),
		Capabilities:      DefaultLearningCapabilities(),
		AggregationPolicy: DefaultAggregationPolicy(),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = agg.AggregateWindow(ctx, req, snap)
	}
}

func BenchmarkValidationPipelineBifurcated(b *testing.B) {
	vp := NewDefaultValidationPipeline(nil)
	ctx := context.Background()
	cand := &CandidateSnapshot{
		SnapshotID: "cand-bench-1",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Payload:    []byte(`{"weights":{"alpha":0.8}}`),
		Lineage: ReplayMetadata{
			PolicyFingerprint:  "policy-fp-bench",
			SourceArtifactHash: "hash-bench-summary",
		},
	}
	summary := &AggregationSummary{
		TotalArtifactsIngested: 15,
		SourceArtifactHash:     "hash-bench-summary",
	}
	profile := DefaultLearningPolicyProfile()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = vp.ValidateCandidate(ctx, cand, summary, profile)
	}
}

func BenchmarkCandidateRankingEngine(b *testing.B) {
	ranking := NewDefaultCandidateRankingEngine()
	ctx := context.Background()
	cands := make([]*CandidateSnapshot, 10)
	for i := 0; i < 10; i++ {
		cands[i] = &CandidateSnapshot{
			SnapshotID: fmt.Sprintf("cand-bench-%d", i),
			SchemaID:   "idun.reasoning.strategy.v1",
			Lineage: ReplayMetadata{
				GenerationDepth: uint32(i % 3),
			},
			ValidationRecords: []ValidationResult{
				{CheckID: "SCHEMA_PAYLOAD_CHECK", Passed: true, Score: 0.8 + float64(i)*0.01},
			},
		}
	}
	summary := &AggregationSummary{}
	profile := DefaultLearningPolicyProfile()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ranking.RankCandidates(ctx, cands, summary, profile)
	}
}

func BenchmarkExperimentManagerStartStop(b *testing.B) {
	sp, _ := NewDefaultStrategyProvider(nil)
	em := NewDefaultExperimentManager(sp)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		expID := fmt.Sprintf("bench-exp-%d", i)
		profile, err := NewExperimentProfileBuilder(expID, "idun.reasoning.strategy.v1", "snap-bench-target").
			WithRatios(0.10, 0.01).
			Build()
		if err == nil {
			_ = em.StartExperiment(ctx, profile)
			_ = em.StopExperiment(ctx, expID)
		}
	}
}

func BenchmarkServiceRunCycleEndToEnd(b *testing.B) {
	store := newBenchMemory(20)
	agg := NewDefaultAggregator(store)
	s, err := NewService(
		WithAggregator(agg),
		WithLearnerRegistry(NewDefaultLearnerRegistry()),
	)
	if err != nil {
		b.Fatalf("failed to init service: %v", err)
	}
	_ = s.Start()
	defer s.Close()

	ctx := context.Background()
	req := &LearningRequest{
		RequestID:         "bench-runcycle-req",
		DomainSchemaID:    "idun.reasoning.trace.v1",
		PolicyFingerprint: DefaultLearningPolicyProfile().PolicyFingerprint,
		TimeWindowStart:   time.Now().Add(-2 * time.Hour),
		TimeWindowEnd:     time.Now().Add(1 * time.Hour),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = s.RunCycle(ctx, req)
	}
}
