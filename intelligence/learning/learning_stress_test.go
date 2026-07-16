package learning

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/core/memory"
)

type stressMemory struct {
	mu      sync.RWMutex
	records map[string]memory.Record
}

func newStressMemory() *stressMemory {
	return &stressMemory{records: make(map[string]memory.Record)}
}

func (m *stressMemory) CreateRecord(rec memory.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec.ID == "" {
		rec.ID = fmt.Sprintf("rec-%d-%d", time.Now().UnixNano(), len(m.records))
	}
	m.records[rec.ID] = rec
	return nil
}
func (m *stressMemory) GetRecord(id string) (memory.Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.records[id]
	if !ok {
		return memory.Record{}, ErrNotFound
	}
	return r, nil
}
func (m *stressMemory) UpdateRecord(rec memory.Record) error { return nil }
func (m *stressMemory) DeleteRecord(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.records, id)
	return nil
}
func (m *stressMemory) ListRecordsByType(t string) ([]memory.Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []memory.Record
	for _, r := range m.records {
		if r.Type == t {
			out = append(out, r)
		}
	}
	return out, nil
}
func (m *stressMemory) LinkRecords(from, to, relationship, creator string) error { return nil }
func (m *stressMemory) UnlinkRecords(from, to, relationship string) error        { return nil }
func (m *stressMemory) GetLinkedRecords(from, rel string) ([]memory.Record, error) { return nil, nil }
func (m *stressMemory) GetLinkedRecordsReverse(to, rel string) ([]memory.Record, error) {
	return nil, nil
}

func TestStressConcurrentRunCycleAndAggregator(t *testing.T) {
	store := newStressMemory()
	agg := NewDefaultAggregator(store)
	s, err := NewService(
		WithAggregator(agg),
		WithLearnerRegistry(NewDefaultLearnerRegistry()),
	)
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}
	_ = s.Start()
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	const numGoroutines = 24

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			schemaID := "idun.reasoning.trace.v1"
			if idx%2 == 1 {
				schemaID = "idun.planning.task.v1"
			}
			for ctx.Err() == nil {
				_ = store.CreateRecord(memory.Record{
					ID:        fmt.Sprintf("stress-rec-%d-%d", idx, time.Now().UnixNano()),
					Type:      schemaID,
					Payload:   []byte(fmt.Sprintf(`{"category":"COGNITIVE_PERFORMANCE","score":0.88,"idx":%d}`, idx)),
					CreatedAt: time.Now().Add(-30 * time.Minute),
				})
				req := &LearningRequest{
					RequestID:         fmt.Sprintf("req-stress-%d-%d", idx, time.Now().UnixNano()),
					DomainSchemaID:    schemaID,
					PolicyFingerprint: DefaultLearningPolicyProfile().PolicyFingerprint,
					TimeWindowStart:   time.Now().Add(-1 * time.Hour),
					TimeWindowEnd:     time.Now().Add(1 * time.Hour),
				}
				_, _ = s.RunCycle(ctx, req)
			}
		}(i)
	}

	wg.Wait()
}

func TestStressConcurrentLearnerAndSnapshotRegistry(t *testing.T) {
	reg := NewDefaultSnapshotRegistry()
	lReg := NewLearnerRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for ctx.Err() == nil {
				lnr := &mockStressLearner{
					id:  fmt.Sprintf("stress-lnr-%d", idx%5),
					ver: "1.0.0",
					fp:  fmt.Sprintf("fp-stress-%d", idx%5),
					in:  []string{"idun.reasoning.trace.v1"},
					out: []string{"idun.reasoning.strategy.v1"},
				}
				_ = lReg.Register(lnr)
				_ = lReg.ListLearners()

				snap := &CandidateSnapshot{
					SnapshotID: fmt.Sprintf("snap-stress-%d-%d", idx, time.Now().UnixNano()),
					SemVer:     "1.0.0",
					SchemaID:   "idun.reasoning.strategy.v1",
					Lifecycle:  LifecycleValidated,
					Payload:    []byte(`{"weights":{"a":1}}`),
					Lineage: ReplayMetadata{
						PolicyFingerprint:  "fp-pol",
						SourceArtifactHash: "hash-src",
					},
				}
				_ = reg.Publish(ctx, snap)
				_, _ = reg.GetActive(ctx, "idun.reasoning.strategy.v1")
			}
		}(i)
	}

	wg.Wait()
}

type mockStressLearner struct {
	id  string
	ver string
	fp  string
	in  []string
	out []string
}

func (m *mockStressLearner) LearnerID() string          { return m.id }
func (m *mockStressLearner) LearnerVersion() string     { return m.ver }
func (m *mockStressLearner) LearnerFingerprint() string { return m.fp }
func (m *mockStressLearner) Consumes() []string         { return m.in }
func (m *mockStressLearner) Produces() []string         { return m.out }
func (m *mockStressLearner) Generate(ctx context.Context, sum *AggregationSummary) ([]*CandidateSnapshot, error) {
	return nil, nil
}

func TestStressConcurrentExperimentAndTelemetryQueries(t *testing.T) {
	sp, _ := NewDefaultStrategyProvider(nil)
	em := NewDefaultExperimentManager(sp)
	s, err := NewService(
		WithLearnerRegistry(NewDefaultLearnerRegistry()),
	)
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}
	_ = s.Start()
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for ctx.Err() == nil {
				expID := fmt.Sprintf("stress-exp-%d", idx)
				prof, _ := NewExperimentProfileBuilder(expID, "idun.reasoning.strategy.v1", "snap-target").Build()
				_ = em.StartExperiment(ctx, prof)
				_, _ = em.GetActiveExperiment(ctx, expID)
				_ = em.ListActiveExperiments(ctx)
				_ = em.StopExperiment(ctx, expID)

				_ = s.ListLearnerPerformanceSummaries(ctx)
				_, _ = s.GetLearnerPerformanceSummary(ctx, "learner-reasoning-heuristics-v1")
			}
		}(i)
	}

	wg.Wait()
}
