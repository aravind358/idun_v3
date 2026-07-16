package learning

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/core/memory"
)

// mockMemory is a thread-safe mock implementation of memory.Memory for testing aggregation.
type mockMemory struct {
	mu      sync.RWMutex
	records []memory.Record
}

func newMockMemory() *mockMemory {
	return &mockMemory{records: make([]memory.Record, 0)}
}

func (m *mockMemory) CreateRecord(rec memory.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec.ID == "" {
		rec.ID = fmt.Sprintf("rec-%d", len(m.records)+1)
	}
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now()
	}
	m.records = append(m.records, rec)
	return nil
}

func (m *mockMemory) GetRecord(id string) (memory.Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.records {
		if r.ID == id {
			return r, nil
		}
	}
	return memory.Record{}, errors.New("not found")
}

func (m *mockMemory) UpdateRecord(rec memory.Record) error { return nil }
func (m *mockMemory) DeleteRecord(id string) error                 { return nil }

func (m *mockMemory) ListRecordsByType(recordType string) ([]memory.Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []memory.Record
	for _, r := range m.records {
		if r.Type == recordType {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *mockMemory) LinkRecords(from, to, relationship, creator string) error { return nil }
func (m *mockMemory) UnlinkRecords(from, to, relationship string) error        { return nil }
func (m *mockMemory) GetLinkedRecords(from, relationship string) ([]memory.Record, error) {
	return nil, nil
}
func (m *mockMemory) GetLinkedRecordsReverse(to, relationship string) ([]memory.Record, error) {
	return nil, nil
}

func TestAggregatorPartitionEnforcement(t *testing.T) {
	ctx := context.Background()
	store := newMockMemory()
	agg := NewDefaultAggregator(store)

	now := time.Now()
	// Add valid COGNITIVE_PERFORMANCE records
	rec1 := memory.Record{
		ID:        "rec-cog-1",
		Type:      "idun.reasoning.trace.v1",
		Payload:   []byte(`{"category":"COGNITIVE_PERFORMANCE","score":0.95}`),
		CreatedAt: now.Add(-1 * time.Hour),
	}
	// Add invalid LEARNING_DIAGNOSTICS record
	rec2 := memory.Record{
		ID:        "rec-diag-2",
		Type:      "idun.reasoning.trace.v1",
		Payload:   []byte(`{"category":"LEARNING_DIAGNOSTICS","metrics":"sensitive"}`),
		CreatedAt: now.Add(-30 * time.Minute),
	}
	// Add record with TopicLearningTraces type
	rec3 := memory.Record{
		ID:        "rec-diag-3",
		Type:      string(TopicLearningTraces),
		Payload:   []byte(`{"category":"COGNITIVE_PERFORMANCE"}`), // Type should still block it
		CreatedAt: now.Add(-15 * time.Minute),
	}
	store.mu.Lock()
	store.records = append(store.records, rec1, rec2, rec3)
	store.mu.Unlock()

	req := &LearningRequest{
		RequestID:         "req-agg-1",
		DomainSchemaID:    "idun.reasoning.trace.v1",
		PolicyFingerprint: "fp-test-123",
		TimeWindowStart:   now.Add(-2 * time.Hour),
		TimeWindowEnd:     now.Add(1 * time.Hour),
	}
	snap := &LearningStrategySnapshot{
		SnapshotID:        "snap-1",
		SchemaVersion:     SchemaVersion,
		ActiveProfile:     DefaultLearningPolicyProfile(),
		Capabilities:      DefaultLearningCapabilities(),
		AggregationPolicy: DefaultAggregationPolicy(),
	}

	summary, lineage, err := agg.AggregateWindow(ctx, req, snap)
	if err != nil {
		t.Fatalf("AggregateWindow failed: %v", err)
	}

	if summary.TotalArtifactsIngested != 1 {
		t.Errorf("expected exactly 1 record ingested (rec-cog-1), got %d", summary.TotalArtifactsIngested)
	}
	if len(summary.Records) != 1 || summary.Records[0].ID != "rec-cog-1" {
		t.Errorf("expected rec-cog-1 retained, got %v", summary.Records)
	}
	if summary.SourceArtifactHash == "" || lineage.SourceArtifactHash != summary.SourceArtifactHash {
		t.Errorf("lineage source hash mismatch or empty: %s vs %s", lineage.SourceArtifactHash, summary.SourceArtifactHash)
	}
}

func TestAggregatorOrderingAndBounds(t *testing.T) {
	ctx := context.Background()
	store := newMockMemory()
	agg := NewDefaultAggregator(store)

	now := time.Now()
	r1 := memory.Record{ID: "r1", Type: "idun.reasoning.trace.v1", Payload: []byte(`{"score":1}`), CreatedAt: now.Add(-3 * time.Hour)}
	r2 := memory.Record{ID: "r2", Type: "idun.reasoning.trace.v1", Payload: []byte(`{"score":2}`), CreatedAt: now.Add(-1 * time.Hour)}
	r3 := memory.Record{ID: "r3", Type: "idun.reasoning.trace.v1", Payload: []byte(`{"score":3}`), CreatedAt: now.Add(-2 * time.Hour)}
	store.mu.Lock()
	store.records = append(store.records, r1, r2, r3)
	store.mu.Unlock()

	req := &LearningRequest{
		RequestID:         "req-ord",
		DomainSchemaID:    "idun.reasoning.trace.v1",
		PolicyFingerprint: "fp-test-123",
		TimeWindowStart:   now.Add(-4 * time.Hour),
		TimeWindowEnd:     now.Add(1 * time.Hour),
	}

	// 1. Chronological Ascending with max artifacts = 2
	policy := DefaultAggregationPolicy()
	policy.OrderingStrategy = OrderingStrategyChronologicalAsc
	policy.MaximumArtifacts = 2

	snap := &LearningStrategySnapshot{
		SnapshotID:        "snap-ord",
		SchemaVersion:     SchemaVersion,
		ActiveProfile:     DefaultLearningPolicyProfile(),
		Capabilities:      DefaultLearningCapabilities(),
		AggregationPolicy: policy,
	}

	summary, _, err := agg.AggregateWindow(ctx, req, snap)
	if err != nil {
		t.Fatalf("AggregateWindow asc failed: %v", err)
	}
	if len(summary.Records) != 2 {
		t.Fatalf("expected 2 records bounded, got %d", len(summary.Records))
	}
	if summary.Records[0].ID != "r1" || summary.Records[1].ID != "r3" {
		t.Errorf("expected [r1, r3] ascending chronological, got [%s, %s]", summary.Records[0].ID, summary.Records[1].ID)
	}

	// 2. Chronological Descending
	policy.OrderingStrategy = OrderingStrategyChronologicalDesc
	policy.MaximumArtifacts = 10
	summaryDesc, _, err := agg.AggregateWindow(ctx, req, snap)
	if err != nil {
		t.Fatalf("AggregateWindow desc failed: %v", err)
	}
	if len(summaryDesc.Records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(summaryDesc.Records))
	}
	if summaryDesc.Records[0].ID != "r2" || summaryDesc.Records[2].ID != "r1" {
		t.Errorf("expected [r2, r3, r1] descending chronological, got [%s, %s, %s]", summaryDesc.Records[0].ID, summaryDesc.Records[1].ID, summaryDesc.Records[2].ID)
	}
}
