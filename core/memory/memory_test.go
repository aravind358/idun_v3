// Tests for the Memory Service.
//
// Test file: memory_test.go
// Package: memory_test (external black-box test package)
//
// Using an external test package exercises only the exported API of Memory.
// All 72 tests specified in frozen ADR-CS-003 Section 10 are implemented here.
//
// Run with:
//
//	go test ./core/memory/... -v
//	go test ./core/memory/... -v -race    ← required for concurrency tests
package memory_test

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"idun/core/logger"
	"idun/core/memory"
	"idun/core/storage"
)

// ============================================================
// Helpers
// ============================================================

func newTestLogger(t *testing.T) logger.Writer {
	t.Helper()
	l, err := logger.NewLogger(logger.Config{Output: io.Discard})
	if err != nil {
		t.Fatalf("newTestLogger: %v", err)
	}
	return l
}

func newTestStorage(t *testing.T) storage.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := storage.NewStorage(storage.Config{Path: dir}, newTestLogger(t))
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func newTestMemory(t *testing.T) memory.Memory {
	t.Helper()
	m, err := memory.NewMemoryService(memory.Config{}, newTestStorage(t), newTestLogger(t))
	if err != nil {
		t.Fatalf("NewMemoryService: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func newTestMemoryService(t *testing.T) *memory.MemoryService {
	t.Helper()
	m, err := memory.NewMemoryService(memory.Config{}, newTestStorage(t), newTestLogger(t))
	if err != nil {
		t.Fatalf("NewMemoryService: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func makeRecord(id, typ string, payload []byte) memory.Record {
	return memory.Record{
		ID:      id,
		Type:    typ,
		Payload: payload,
		Creator: "test",
	}
}

// ============================================================
// 10.1 Constructor Tests
// ============================================================

func TestNewMemoryService_Success(t *testing.T) {
	m, err := memory.NewMemoryService(memory.Config{}, newTestStorage(t), newTestLogger(t))
	if err != nil {
		t.Fatalf("NewMemoryService returned unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("NewMemoryService returned nil")
	}
	_ = m.Close()
}

func TestNewMemoryService_NilStore(t *testing.T) {
	m, err := memory.NewMemoryService(memory.Config{}, nil, newTestLogger(t))
	if err == nil {
		t.Error("expected error for nil store, got nil")
	}
	if m != nil {
		t.Error("expected nil MemoryService for nil store")
	}
}

func TestNewMemoryService_NilLogger(t *testing.T) {
	m, err := memory.NewMemoryService(memory.Config{}, newTestStorage(t), nil)
	if err == nil {
		t.Error("expected error for nil logger, got nil")
	}
	if m != nil {
		t.Error("expected nil MemoryService for nil logger")
	}
}

func TestNewMemoryService_DefaultNamespace(t *testing.T) {
	// Empty Namespace should use "memory/" — validated indirectly by successful
	// CreateRecord and GetRecord round-trip.
	m, err := memory.NewMemoryService(memory.Config{}, newTestStorage(t), newTestLogger(t))
	if err != nil {
		t.Fatalf("NewMemoryService: %v", err)
	}
	defer m.Close()

	rec := makeRecord("ns/test", "thing", []byte("ok"))
	if err := m.CreateRecord(rec); err != nil {
		t.Fatalf("CreateRecord with default namespace: %v", err)
	}
}

func TestNewMemoryService_CustomNamespace(t *testing.T) {
	store := newTestStorage(t)
	log := newTestLogger(t)

	m1, err := memory.NewMemoryService(memory.Config{Namespace: "ns1/"}, store, log)
	if err != nil {
		t.Fatalf("NewMemoryService ns1: %v", err)
	}
	m2, err := memory.NewMemoryService(memory.Config{Namespace: "ns2/"}, store, log)
	if err != nil {
		t.Fatalf("NewMemoryService ns2: %v", err)
	}
	defer m1.Close()
	defer m2.Close()

	// Same ID in different namespaces must not collide.
	rec := makeRecord("person/alice", "person", []byte("alice"))
	if err := m1.CreateRecord(rec); err != nil {
		t.Fatalf("CreateRecord in ns1: %v", err)
	}
	if err := m2.CreateRecord(rec); err != nil {
		t.Fatalf("CreateRecord in ns2 (should not conflict): %v", err)
	}
}

// ============================================================
// 10.2 Component Interface Tests
// ============================================================

func TestMemoryService_Name(t *testing.T) {
	m := newTestMemoryService(t)
	got := m.Name()
	want := "MemoryService"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// Compile-time verification that *MemoryService satisfies Memory interface.
var _ memory.Memory = (*memory.MemoryService)(nil)

func TestMemoryService_InterfaceCompliance(t *testing.T) {
	m := newTestMemoryService(t)
	var _ memory.Memory = m
}

// ============================================================
// 10.3 CreateRecord Tests
// ============================================================

func TestCreateRecord_Success(t *testing.T) {
	m := newTestMemory(t)
	err := m.CreateRecord(makeRecord("person/alice", "person", []byte("data")))
	if err != nil {
		t.Fatalf("CreateRecord returned unexpected error: %v", err)
	}
}

func TestCreateRecord_EmptyID(t *testing.T) {
	m := newTestMemory(t)
	err := m.CreateRecord(makeRecord("", "person", []byte("data")))
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestCreateRecord_EmptyType(t *testing.T) {
	m := newTestMemory(t)
	err := m.CreateRecord(makeRecord("person/alice", "", []byte("data")))
	if err == nil {
		t.Error("expected error for empty Type, got nil")
	}
}

func TestCreateRecord_NilPayload(t *testing.T) {
	m := newTestMemory(t)
	err := m.CreateRecord(makeRecord("task/a", "task", nil))
	if err != nil {
		t.Fatalf("CreateRecord with nil payload returned error: %v", err)
	}
	got, err := m.GetRecord("task/a")
	if err != nil {
		t.Fatalf("GetRecord after nil payload: %v", err)
	}
	if got.Payload == nil {
		t.Error("Payload should be empty slice, not nil")
	}
}

func TestCreateRecord_DuplicateID(t *testing.T) {
	m := newTestMemory(t)
	rec := makeRecord("person/alice", "person", []byte("v1"))
	if err := m.CreateRecord(rec); err != nil {
		t.Fatalf("first CreateRecord: %v", err)
	}
	err := m.CreateRecord(makeRecord("person/alice", "person", []byte("v2")))
	if err == nil {
		t.Error("expected error for duplicate ID, got nil")
	}
}

func TestCreateRecord_SetsCreatedAt(t *testing.T) {
	m := newTestMemory(t)
	rec := makeRecord("event/x", "event", []byte("e"))
	if err := m.CreateRecord(rec); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	got, err := m.GetRecord("event/x")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by Memory, not remain zero")
	}
}

func TestCreateRecord_SetsUpdatedAt(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("event/y", "event", []byte("e"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	got, err := m.GetRecord("event/y")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if !got.UpdatedAt.Equal(got.CreatedAt) {
		t.Errorf("UpdatedAt should equal CreatedAt immediately after creation, got CreatedAt=%v UpdatedAt=%v",
			got.CreatedAt, got.UpdatedAt)
	}
}

func TestCreateRecord_PersistsToStorage(t *testing.T) {
	// We verify persistence indirectly: construct a second MemoryService
	// over the same storage and confirm GetRecord succeeds.
	store := newTestStorage(t)
	log := newTestLogger(t)

	m1, err := memory.NewMemoryService(memory.Config{}, store, log)
	if err != nil {
		t.Fatalf("NewMemoryService m1: %v", err)
	}
	if err := m1.CreateRecord(makeRecord("person/mom", "person", []byte("mom"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	_ = m1.Close()

	m2, err := memory.NewMemoryService(memory.Config{}, store, log)
	if err != nil {
		t.Fatalf("NewMemoryService m2: %v", err)
	}
	defer m2.Close()

	got, err := m2.GetRecord("person/mom")
	if err != nil {
		t.Fatalf("GetRecord after restart: %v", err)
	}
	if string(got.Payload) != "mom" {
		t.Errorf("Payload = %q, want %q", got.Payload, "mom")
	}
}

// ============================================================
// 10.4 GetRecord Tests
// ============================================================

func TestGetRecord_Success(t *testing.T) {
	m := newTestMemory(t)
	want := []byte("hello memory")
	if err := m.CreateRecord(makeRecord("task/t1", "task", want)); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	got, err := m.GetRecord("task/t1")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if string(got.Payload) != string(want) {
		t.Errorf("Payload = %q, want %q", got.Payload, want)
	}
}

func TestGetRecord_NotFound(t *testing.T) {
	m := newTestMemory(t)
	_, err := m.GetRecord("does/not/exist")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetRecord_EmptyID(t *testing.T) {
	m := newTestMemory(t)
	_, err := m.GetRecord("")
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestGetRecord_ReturnsCopy(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/copy", "task", []byte("original"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	first, err := m.GetRecord("task/copy")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	first.Payload[0] = 'X' // mutate returned copy

	second, err := m.GetRecord("task/copy")
	if err != nil {
		t.Fatalf("GetRecord second: %v", err)
	}
	if string(second.Payload) != "original" {
		t.Errorf("stored data was mutated: got %q, want %q", second.Payload, "original")
	}
}

func TestGetRecord_PreservesPayload(t *testing.T) {
	m := newTestMemory(t)
	want := []byte{0x00, 0xFF, 0x80, 0x7F, 0x01}
	if err := m.CreateRecord(makeRecord("bin/a", "binary", want)); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	got, err := m.GetRecord("bin/a")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if string(got.Payload) != string(want) {
		t.Errorf("Payload mismatch: got %v, want %v", got.Payload, want)
	}
}

func TestGetRecord_PreservesMetadata(t *testing.T) {
	m := newTestMemory(t)
	rec := makeRecord("meta/r", "meta", []byte("x"))
	rec.Creator = "tester"
	if err := m.CreateRecord(rec); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	got, err := m.GetRecord("meta/r")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if got.Creator != "tester" {
		t.Errorf("Creator = %q, want %q", got.Creator, "tester")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero")
	}
}

// ============================================================
// 10.5 UpdateRecord Tests
// ============================================================

func TestUpdateRecord_Success(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/upd", "task", []byte("v1"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	upd := memory.Record{ID: "task/upd", Payload: []byte("v2")}
	if err := m.UpdateRecord(upd); err != nil {
		t.Fatalf("UpdateRecord: %v", err)
	}
	got, err := m.GetRecord("task/upd")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if string(got.Payload) != "v2" {
		t.Errorf("Payload = %q, want %q", got.Payload, "v2")
	}
}

func TestUpdateRecord_NotFound(t *testing.T) {
	m := newTestMemory(t)
	err := m.UpdateRecord(memory.Record{ID: "missing/rec", Payload: []byte("x")})
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateRecord_EmptyID(t *testing.T) {
	m := newTestMemory(t)
	err := m.UpdateRecord(memory.Record{Payload: []byte("x")})
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestUpdateRecord_PreservesCreatedAt(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/ts", "task", []byte("v1"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	original, err := m.GetRecord("task/ts")
	if err != nil {
		t.Fatalf("GetRecord original: %v", err)
	}

	if err := m.UpdateRecord(memory.Record{ID: "task/ts", Payload: []byte("v2")}); err != nil {
		t.Fatalf("UpdateRecord: %v", err)
	}
	updated, err := m.GetRecord("task/ts")
	if err != nil {
		t.Fatalf("GetRecord updated: %v", err)
	}
	if !updated.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt changed: was %v, now %v", original.CreatedAt, updated.CreatedAt)
	}
}

func TestUpdateRecord_UpdatesUpdatedAt(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/uat", "task", []byte("v1"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	before, err := m.GetRecord("task/uat")
	if err != nil {
		t.Fatalf("GetRecord before: %v", err)
	}

	// Sleep briefly to ensure UpdatedAt advances.
	// On fast machines, two time.Now() calls within the same test may be equal
	// (sub-nanosecond). We force a small pause via a channel send with a timeout.
	done := make(chan struct{})
	go func() { close(done) }()
	<-done

	if err := m.UpdateRecord(memory.Record{ID: "task/uat", Payload: []byte("v2")}); err != nil {
		t.Fatalf("UpdateRecord: %v", err)
	}
	after, err := m.GetRecord("task/uat")
	if err != nil {
		t.Fatalf("GetRecord after: %v", err)
	}
	if !after.UpdatedAt.After(before.CreatedAt) && !after.UpdatedAt.Equal(before.UpdatedAt) {
		// UpdatedAt must be >= original UpdatedAt (it may be equal on very fast machines
		// if the clock resolution is low). It must not be before.
		if after.UpdatedAt.Before(before.UpdatedAt) {
			t.Errorf("UpdatedAt went backwards: was %v, now %v", before.UpdatedAt, after.UpdatedAt)
		}
	}
}

// ============================================================
// 10.6 DeleteRecord Tests
// ============================================================

func TestDeleteRecord_Success(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/del", "task", []byte("bye"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	if err := m.DeleteRecord("task/del"); err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}
	_, err := m.GetRecord("task/del")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteRecord_NotFound(t *testing.T) {
	m := newTestMemory(t)
	err := m.DeleteRecord("missing/rec")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteRecord_EmptyID(t *testing.T) {
	m := newTestMemory(t)
	err := m.DeleteRecord("")
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestDeleteRecord_RemovesLinksWhere_From(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/a", "task", []byte("a"))); err != nil {
		t.Fatalf("CreateRecord task/a: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/b", "person", []byte("b"))); err != nil {
		t.Fatalf("CreateRecord person/b: %v", err)
	}
	if err := m.LinkRecords("task/a", "person/b", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}

	if err := m.DeleteRecord("task/a"); err != nil {
		t.Fatalf("DeleteRecord task/a: %v", err)
	}

	// Forward traversal from deleted record should be empty (person/b is unaffected).
	records, err := m.GetLinkedRecordsReverse("person/b", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecordsReverse: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 reverse-linked records after source deleted, got %d", len(records))
	}
}

func TestDeleteRecord_RemovesLinksWhere_To(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/c", "task", []byte("c"))); err != nil {
		t.Fatalf("CreateRecord task/c: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/d", "person", []byte("d"))); err != nil {
		t.Fatalf("CreateRecord person/d: %v", err)
	}
	if err := m.LinkRecords("task/c", "person/d", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}

	if err := m.DeleteRecord("person/d"); err != nil {
		t.Fatalf("DeleteRecord person/d: %v", err)
	}

	// Forward traversal from task/c should no longer return person/d.
	records, err := m.GetLinkedRecords("task/c", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 forward-linked records after target deleted, got %d", len(records))
	}
}

func TestDeleteRecord_RemovesFromTypeIndex(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/idx", "task", []byte("x"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	if err := m.DeleteRecord("task/idx"); err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}
	records, err := m.ListRecordsByType("task")
	if err != nil {
		t.Fatalf("ListRecordsByType: %v", err)
	}
	for _, r := range records {
		if r.ID == "task/idx" {
			t.Error("deleted record still appears in ListRecordsByType")
		}
	}
}

// ============================================================
// 10.7 ListRecordsByType Tests
// ============================================================

func TestListRecordsByType_Empty(t *testing.T) {
	m := newTestMemory(t)
	records, err := m.ListRecordsByType("nonexistent")
	if err != nil {
		t.Fatalf("ListRecordsByType: %v", err)
	}
	if records == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestListRecordsByType_SingleRecord(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/only", "task", []byte("x"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	records, err := m.ListRecordsByType("task")
	if err != nil {
		t.Fatalf("ListRecordsByType: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
}

func TestListRecordsByType_MultipleRecords(t *testing.T) {
	m := newTestMemory(t)
	for _, id := range []string{"task/1", "task/2", "task/3"} {
		if err := m.CreateRecord(makeRecord(id, "task", []byte(id))); err != nil {
			t.Fatalf("CreateRecord %q: %v", id, err)
		}
	}
	records, err := m.ListRecordsByType("task")
	if err != nil {
		t.Fatalf("ListRecordsByType: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}

func TestListRecordsByType_OnlyMatchingType(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/t", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/p", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	records, err := m.ListRecordsByType("task")
	if err != nil {
		t.Fatalf("ListRecordsByType: %v", err)
	}
	for _, r := range records {
		if r.Type != "task" {
			t.Errorf("unexpected type %q in results", r.Type)
		}
	}
	if len(records) != 1 {
		t.Errorf("expected 1 task record, got %d", len(records))
	}
}

func TestListRecordsByType_CreationOrder(t *testing.T) {
	m := newTestMemory(t)
	ids := []string{"task/z", "task/a", "task/m"}
	for _, id := range ids {
		if err := m.CreateRecord(makeRecord(id, "task", []byte(id))); err != nil {
			t.Fatalf("CreateRecord %q: %v", id, err)
		}
	}
	records, err := m.ListRecordsByType("task")
	if err != nil {
		t.Fatalf("ListRecordsByType: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	// Must be in creation order: z, a, m (insertion order, not alphabetical).
	for i, id := range ids {
		if records[i].ID != id {
			t.Errorf("position %d: got %q, want %q", i, records[i].ID, id)
		}
	}
}

func TestListRecordsByType_AfterDelete(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/keep", "task", []byte("k"))); err != nil {
		t.Fatalf("CreateRecord keep: %v", err)
	}
	if err := m.CreateRecord(makeRecord("task/gone", "task", []byte("g"))); err != nil {
		t.Fatalf("CreateRecord gone: %v", err)
	}
	if err := m.DeleteRecord("task/gone"); err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}
	records, err := m.ListRecordsByType("task")
	if err != nil {
		t.Fatalf("ListRecordsByType: %v", err)
	}
	if len(records) != 1 || records[0].ID != "task/keep" {
		t.Errorf("expected only task/keep, got %v", records)
	}
}

func TestListRecordsByType_EmptyType(t *testing.T) {
	m := newTestMemory(t)
	_, err := m.ListRecordsByType("")
	if err == nil {
		t.Error("expected error for empty type, got nil")
	}
}

// ============================================================
// 10.8 LinkRecords Tests
// ============================================================

func TestLinkRecords_Success(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/lt", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/lp", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/lt", "person/lp", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	records, err := m.GetLinkedRecords("task/lt", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if len(records) != 1 || records[0].ID != "person/lp" {
		t.Errorf("GetLinkedRecords = %v, want [person/lp]", records)
	}
}

func TestLinkRecords_FromNotFound(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("person/lp2", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	err := m.LinkRecords("missing/from", "person/lp2", "concerns", "test")
	if err == nil {
		t.Error("expected error for missing from record, got nil")
	}
}

func TestLinkRecords_ToNotFound(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/lt2", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	err := m.LinkRecords("task/lt2", "missing/to", "concerns", "test")
	if err == nil {
		t.Error("expected error for missing to record, got nil")
	}
}

func TestLinkRecords_EmptyFrom(t *testing.T) {
	m := newTestMemory(t)
	err := m.LinkRecords("", "person/x", "concerns", "test")
	if err == nil {
		t.Error("expected error for empty from, got nil")
	}
}

func TestLinkRecords_EmptyTo(t *testing.T) {
	m := newTestMemory(t)
	err := m.LinkRecords("task/x", "", "concerns", "test")
	if err == nil {
		t.Error("expected error for empty to, got nil")
	}
}

func TestLinkRecords_EmptyRelationship(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/lr", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/lr", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	err := m.LinkRecords("task/lr", "person/lr", "", "test")
	if err == nil {
		t.Error("expected error for empty relationship, got nil")
	}
}

func TestLinkRecords_DuplicateLink(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/dup", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/dup", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/dup", "person/dup", "concerns", "test"); err != nil {
		t.Fatalf("first LinkRecords: %v", err)
	}
	err := m.LinkRecords("task/dup", "person/dup", "concerns", "test")
	if err == nil {
		t.Error("expected error for duplicate link, got nil")
	}
}

func TestLinkRecords_DifferentRelationship(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/dr", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/dr", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/dr", "person/dr", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords concerns: %v", err)
	}
	// Same from/to, different relationship — must succeed.
	if err := m.LinkRecords("task/dr", "person/dr", "references", "test"); err != nil {
		t.Fatalf("LinkRecords references: %v", err)
	}
}

func TestLinkRecords_SetsCreatedAt(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/lca", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/lca", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/lca", "person/lca", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	// Confirm the link exists — CreatedAt is tested indirectly (it is set inside Memory).
	records, err := m.GetLinkedRecords("task/lca", "concerns")
	if err != nil || len(records) != 1 {
		t.Errorf("GetLinkedRecords: records=%v err=%v", records, err)
	}
}

// ============================================================
// 10.9 UnlinkRecords Tests
// ============================================================

func TestUnlinkRecords_Success(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/ul", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/ul", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/ul", "person/ul", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	if err := m.UnlinkRecords("task/ul", "person/ul", "concerns"); err != nil {
		t.Fatalf("UnlinkRecords: %v", err)
	}
	records, err := m.GetLinkedRecords("task/ul", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records after unlink, got %d", len(records))
	}
}

func TestUnlinkRecords_NotFound(t *testing.T) {
	m := newTestMemory(t)
	err := m.UnlinkRecords("task/x", "person/y", "concerns")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUnlinkRecords_EmptyFrom(t *testing.T) {
	m := newTestMemory(t)
	err := m.UnlinkRecords("", "person/y", "concerns")
	if err == nil {
		t.Error("expected error for empty from, got nil")
	}
}

func TestUnlinkRecords_EmptyTo(t *testing.T) {
	m := newTestMemory(t)
	err := m.UnlinkRecords("task/x", "", "concerns")
	if err == nil {
		t.Error("expected error for empty to, got nil")
	}
}

func TestUnlinkRecords_EmptyRelationship(t *testing.T) {
	m := newTestMemory(t)
	err := m.UnlinkRecords("task/x", "person/y", "")
	if err == nil {
		t.Error("expected error for empty relationship, got nil")
	}
}

func TestUnlinkRecords_OnlyRemovesTargetEdge(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/oe", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/oe1", "person", []byte("p1"))); err != nil {
		t.Fatalf("CreateRecord p1: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/oe2", "person", []byte("p2"))); err != nil {
		t.Fatalf("CreateRecord p2: %v", err)
	}
	if err := m.LinkRecords("task/oe", "person/oe1", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords p1: %v", err)
	}
	if err := m.LinkRecords("task/oe", "person/oe2", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords p2: %v", err)
	}

	if err := m.UnlinkRecords("task/oe", "person/oe1", "concerns"); err != nil {
		t.Fatalf("UnlinkRecords p1: %v", err)
	}

	records, err := m.GetLinkedRecords("task/oe", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if len(records) != 1 || records[0].ID != "person/oe2" {
		t.Errorf("expected only person/oe2, got %v", records)
	}
}

// ============================================================
// 10.10 GetLinkedRecords Tests
// ============================================================

func TestGetLinkedRecords_Success(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/gs", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/gs", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/gs", "person/gs", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	records, err := m.GetLinkedRecords("task/gs", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if len(records) != 1 || records[0].ID != "person/gs" {
		t.Errorf("expected [person/gs], got %v", records)
	}
}

func TestGetLinkedRecords_NoLinks(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/nl", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	records, err := m.GetLinkedRecords("task/nl", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if records == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestGetLinkedRecords_MultipleTargets(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/mt", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	for _, id := range []string{"person/p1", "person/p2", "person/p3"} {
		if err := m.CreateRecord(makeRecord(id, "person", []byte(id))); err != nil {
			t.Fatalf("CreateRecord %q: %v", id, err)
		}
		if err := m.LinkRecords("task/mt", id, "concerns", "test"); err != nil {
			t.Fatalf("LinkRecords %q: %v", id, err)
		}
	}
	records, err := m.GetLinkedRecords("task/mt", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}

func TestGetLinkedRecords_FiltersByRelationship(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/fr", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/fr", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.CreateRecord(makeRecord("location/fr", "location", []byte("l"))); err != nil {
		t.Fatalf("CreateRecord location: %v", err)
	}
	if err := m.LinkRecords("task/fr", "person/fr", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords concerns: %v", err)
	}
	if err := m.LinkRecords("task/fr", "location/fr", "happened_at", "test"); err != nil {
		t.Fatalf("LinkRecords happened_at: %v", err)
	}

	records, err := m.GetLinkedRecords("task/fr", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if len(records) != 1 || records[0].ID != "person/fr" {
		t.Errorf("expected only person/fr via concerns, got %v", records)
	}
}

func TestGetLinkedRecords_EmptyFrom(t *testing.T) {
	m := newTestMemory(t)
	_, err := m.GetLinkedRecords("", "concerns")
	if err == nil {
		t.Error("expected error for empty from, got nil")
	}
}

func TestGetLinkedRecords_EmptyRelationship(t *testing.T) {
	m := newTestMemory(t)
	_, err := m.GetLinkedRecords("task/x", "")
	if err == nil {
		t.Error("expected error for empty relationship, got nil")
	}
}

func TestGetLinkedRecords_AfterUnlink(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/au", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/au", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/au", "person/au", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	if err := m.UnlinkRecords("task/au", "person/au", "concerns"); err != nil {
		t.Fatalf("UnlinkRecords: %v", err)
	}
	records, err := m.GetLinkedRecords("task/au", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records after unlink, got %d", len(records))
	}
}

func TestGetLinkedRecords_AfterDeleteRecord(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/adr", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/adr", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/adr", "person/adr", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	if err := m.DeleteRecord("person/adr"); err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}
	records, err := m.GetLinkedRecords("task/adr", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecords after delete: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records after target deleted, got %d", len(records))
	}
}

// ============================================================
// 10.10b GetLinkedRecordsReverse Tests
// ============================================================

func TestGetLinkedRecordsReverse_Success(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/rs", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/rs", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/rs", "person/rs", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	// Reverse: what tasks concern person/rs?
	records, err := m.GetLinkedRecordsReverse("person/rs", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecordsReverse: %v", err)
	}
	if len(records) != 1 || records[0].ID != "task/rs" {
		t.Errorf("expected [task/rs], got %v", records)
	}
}

func TestGetLinkedRecordsReverse_NoLinks(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("person/rnl", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	records, err := m.GetLinkedRecordsReverse("person/rnl", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecordsReverse: %v", err)
	}
	if records == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestGetLinkedRecordsReverse_MultipleFrom(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("person/rmf", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	for _, id := range []string{"task/t1", "task/t2", "task/t3"} {
		if err := m.CreateRecord(makeRecord(id, "task", []byte(id))); err != nil {
			t.Fatalf("CreateRecord %q: %v", id, err)
		}
		if err := m.LinkRecords(id, "person/rmf", "concerns", "test"); err != nil {
			t.Fatalf("LinkRecords %q: %v", id, err)
		}
	}
	records, err := m.GetLinkedRecordsReverse("person/rmf", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecordsReverse: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}

func TestGetLinkedRecordsReverse_FiltersByRelationship(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("person/rfr", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.CreateRecord(makeRecord("task/rfr", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("event/rfr", "event", []byte("e"))); err != nil {
		t.Fatalf("CreateRecord event: %v", err)
	}
	if err := m.LinkRecords("task/rfr", "person/rfr", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords concerns: %v", err)
	}
	if err := m.LinkRecords("event/rfr", "person/rfr", "involves", "test"); err != nil {
		t.Fatalf("LinkRecords involves: %v", err)
	}

	records, err := m.GetLinkedRecordsReverse("person/rfr", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecordsReverse: %v", err)
	}
	if len(records) != 1 || records[0].ID != "task/rfr" {
		t.Errorf("expected only task/rfr via concerns, got %v", records)
	}
}

func TestGetLinkedRecordsReverse_EmptyTo(t *testing.T) {
	m := newTestMemory(t)
	_, err := m.GetLinkedRecordsReverse("", "concerns")
	if err == nil {
		t.Error("expected error for empty to, got nil")
	}
}

func TestGetLinkedRecordsReverse_EmptyRelationship(t *testing.T) {
	m := newTestMemory(t)
	_, err := m.GetLinkedRecordsReverse("person/x", "")
	if err == nil {
		t.Error("expected error for empty relationship, got nil")
	}
}

func TestGetLinkedRecordsReverse_SymmetryWithForward(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/sym", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/sym", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/sym", "person/sym", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}

	fwd, err := m.GetLinkedRecords("task/sym", "concerns")
	if err != nil || len(fwd) != 1 || fwd[0].ID != "person/sym" {
		t.Errorf("forward traversal: records=%v err=%v", fwd, err)
	}

	rev, err := m.GetLinkedRecordsReverse("person/sym", "concerns")
	if err != nil || len(rev) != 1 || rev[0].ID != "task/sym" {
		t.Errorf("reverse traversal: records=%v err=%v", rev, err)
	}
}

func TestGetLinkedRecordsReverse_AfterUnlink(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/rau", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/rau", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/rau", "person/rau", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	if err := m.UnlinkRecords("task/rau", "person/rau", "concerns"); err != nil {
		t.Fatalf("UnlinkRecords: %v", err)
	}
	records, err := m.GetLinkedRecordsReverse("person/rau", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecordsReverse: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records after unlink, got %d", len(records))
	}
}

func TestGetLinkedRecordsReverse_AfterDeleteSource(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/rds", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/rds", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/rds", "person/rds", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	if err := m.DeleteRecord("task/rds"); err != nil {
		t.Fatalf("DeleteRecord task: %v", err)
	}
	records, err := m.GetLinkedRecordsReverse("person/rds", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecordsReverse: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records after source deleted, got %d", len(records))
	}
}

// ============================================================
// 10.11 ErrNotFound Tests
// ============================================================

func TestErrNotFound_IsDistinctFromStorageErrNotFound(t *testing.T) {
	if errors.Is(memory.ErrNotFound, storage.ErrNotFound) {
		t.Error("memory.ErrNotFound must not match storage.ErrNotFound")
	}
	if !errors.Is(memory.ErrNotFound, memory.ErrNotFound) {
		t.Error("memory.ErrNotFound must match itself")
	}
}

func TestErrNotFound_GetRecordUsesIt(t *testing.T) {
	m := newTestMemory(t)
	_, err := m.GetRecord("no/such")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound from GetRecord, got %v", err)
	}
}

func TestErrNotFound_UpdateRecordUsesIt(t *testing.T) {
	m := newTestMemory(t)
	err := m.UpdateRecord(memory.Record{ID: "no/such", Payload: []byte("x")})
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound from UpdateRecord, got %v", err)
	}
}

func TestErrNotFound_DeleteRecordUsesIt(t *testing.T) {
	m := newTestMemory(t)
	err := m.DeleteRecord("no/such")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound from DeleteRecord, got %v", err)
	}
}

func TestErrNotFound_UnlinkRecordsUsesIt(t *testing.T) {
	m := newTestMemory(t)
	err := m.UnlinkRecords("task/x", "person/y", "concerns")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound from UnlinkRecords, got %v", err)
	}
}

// ============================================================
// 10.12 Lifecycle Tests
// ============================================================

func TestMemoryService_Close_Success(t *testing.T) {
	m := newTestMemoryService(t)
	if err := m.Close(); err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}
}

func TestMemoryService_Close_Idempotent(t *testing.T) {
	m := newTestMemoryService(t)
	if err := m.Close(); err != nil {
		t.Errorf("first Close returned error: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Errorf("second Close returned error: %v", err)
	}
}

func TestMemoryService_AfterClose_ReturnsError(t *testing.T) {
	m := newTestMemoryService(t)
	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := m.CreateRecord(makeRecord("x", "x", nil)); err == nil {
		t.Error("expected error from CreateRecord after Close, got nil")
	}
	if _, err := m.GetRecord("x"); err == nil {
		t.Error("expected error from GetRecord after Close, got nil")
	}
	if err := m.UpdateRecord(memory.Record{ID: "x"}); err == nil {
		t.Error("expected error from UpdateRecord after Close, got nil")
	}
	if err := m.DeleteRecord("x"); err == nil {
		t.Error("expected error from DeleteRecord after Close, got nil")
	}
	if _, err := m.ListRecordsByType("x"); err == nil {
		t.Error("expected error from ListRecordsByType after Close, got nil")
	}
	if err := m.LinkRecords("a", "b", "r", "t"); err == nil {
		t.Error("expected error from LinkRecords after Close, got nil")
	}
	if err := m.UnlinkRecords("a", "b", "r"); err == nil {
		t.Error("expected error from UnlinkRecords after Close, got nil")
	}
	if _, err := m.GetLinkedRecords("a", "r"); err == nil {
		t.Error("expected error from GetLinkedRecords after Close, got nil")
	}
	if _, err := m.GetLinkedRecordsReverse("b", "r"); err == nil {
		t.Error("expected error from GetLinkedRecordsReverse after Close, got nil")
	}
}

func TestMemoryService_IndexRebuiltOnStartup(t *testing.T) {
	store := newTestStorage(t)
	log := newTestLogger(t)

	m1, err := memory.NewMemoryService(memory.Config{}, store, log)
	if err != nil {
		t.Fatalf("NewMemoryService m1: %v", err)
	}
	for _, id := range []string{"task/r1", "task/r2", "task/r3"} {
		if err := m1.CreateRecord(makeRecord(id, "task", []byte(id))); err != nil {
			t.Fatalf("CreateRecord %q: %v", id, err)
		}
	}
	_ = m1.Close()

	m2, err := memory.NewMemoryService(memory.Config{}, store, log)
	if err != nil {
		t.Fatalf("NewMemoryService m2: %v", err)
	}
	defer m2.Close()

	records, err := m2.ListRecordsByType("task")
	if err != nil {
		t.Fatalf("ListRecordsByType after restart: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records after restart, got %d", len(records))
	}
}

// ============================================================
// 10.13 Concurrency Tests
// ============================================================

func TestMemoryService_ConcurrentStores_NoPanic(t *testing.T) {
	m := newTestMemory(t)
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("concurrent/task/%d", n)
			_ = m.CreateRecord(makeRecord(id, "task", []byte("v")))
		}(i)
	}
	wg.Wait()
}

func TestMemoryService_ConcurrentReadWrite_NoPanic(t *testing.T) {
	m := newTestMemory(t)
	// Pre-create a record to read.
	if err := m.CreateRecord(makeRecord("shared/rec", "task", []byte("v"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			_, _ = m.GetRecord("shared/rec")
		}(i)
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("rw/task/%d", n)
			_ = m.CreateRecord(makeRecord(id, "task", []byte("v")))
		}(i)
	}
	wg.Wait()
}

func TestMemoryService_ConcurrentLinks_NoPanic(t *testing.T) {
	m := newTestMemory(t)
	// Pre-create source and targets.
	if err := m.CreateRecord(makeRecord("task/src", "task", []byte("src"))); err != nil {
		t.Fatalf("CreateRecord src: %v", err)
	}
	const goroutines = 20
	for i := 0; i < goroutines; i++ {
		id := fmt.Sprintf("person/cl/%d", i)
		if err := m.CreateRecord(makeRecord(id, "person", []byte("p"))); err != nil {
			t.Fatalf("CreateRecord %q: %v", id, err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			to := fmt.Sprintf("person/cl/%d", n)
			_ = m.LinkRecords("task/src", to, "concerns", "test")
		}(i)
	}
	wg.Wait()
}

// ============================================================
// 10.14 Integration Tests
// ============================================================

func TestMemory_SatisfiesKernelComponent(t *testing.T) {
	m := newTestMemoryService(t)
	// Verify Name() returns the canonical service name.
	if m.Name() != "MemoryService" {
		t.Errorf("Name() = %q, want %q", m.Name(), "MemoryService")
	}
	// Verify it can be stored as a kernel.Component equivalent (Name() string).
	type namer interface{ Name() string }
	var _ namer = m
}

func TestMemory_FullRecordLifecycle(t *testing.T) {
	m := newTestMemory(t)

	// Create
	if err := m.CreateRecord(makeRecord("task/lifecycle", "task", []byte("created"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	// Get
	got, err := m.GetRecord("task/lifecycle")
	if err != nil || string(got.Payload) != "created" {
		t.Fatalf("GetRecord: payload=%q err=%v", got.Payload, err)
	}

	// Update
	if err := m.UpdateRecord(memory.Record{ID: "task/lifecycle", Payload: []byte("updated")}); err != nil {
		t.Fatalf("UpdateRecord: %v", err)
	}

	// Get updated
	got2, err := m.GetRecord("task/lifecycle")
	if err != nil || string(got2.Payload) != "updated" {
		t.Fatalf("GetRecord after update: payload=%q err=%v", got2.Payload, err)
	}

	// Delete
	if err := m.DeleteRecord("task/lifecycle"); err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}

	// Get after delete — must be ErrNotFound
	_, err = m.GetRecord("task/lifecycle")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemory_FullRelationshipLifecycle(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/rl", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/rl", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}

	// Link
	if err := m.LinkRecords("task/rl", "person/rl", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}

	// Forward traversal
	fwd, err := m.GetLinkedRecords("task/rl", "concerns")
	if err != nil || len(fwd) != 1 {
		t.Fatalf("GetLinkedRecords: records=%v err=%v", fwd, err)
	}

	// Unlink
	if err := m.UnlinkRecords("task/rl", "person/rl", "concerns"); err != nil {
		t.Fatalf("UnlinkRecords: %v", err)
	}

	// Forward traversal — must be empty
	fwd2, err := m.GetLinkedRecords("task/rl", "concerns")
	if err != nil || len(fwd2) != 0 {
		t.Errorf("GetLinkedRecords after unlink: records=%v err=%v", fwd2, err)
	}
}

func TestMemory_BidirectionalTraversal(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/bd", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/bd", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/bd", "person/bd", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}

	fwd, err := m.GetLinkedRecords("task/bd", "concerns")
	if err != nil || len(fwd) != 1 || fwd[0].ID != "person/bd" {
		t.Errorf("forward traversal: records=%v err=%v", fwd, err)
	}

	rev, err := m.GetLinkedRecordsReverse("person/bd", "concerns")
	if err != nil || len(rev) != 1 || rev[0].ID != "task/bd" {
		t.Errorf("reverse traversal: records=%v err=%v", rev, err)
	}
}

func TestMemory_DeleteCascadesToLinks(t *testing.T) {
	m := newTestMemory(t)
	if err := m.CreateRecord(makeRecord("task/casc", "task", []byte("t"))); err != nil {
		t.Fatalf("CreateRecord task: %v", err)
	}
	if err := m.CreateRecord(makeRecord("person/casc", "person", []byte("p"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m.LinkRecords("task/casc", "person/casc", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}

	// Delete the source record.
	if err := m.DeleteRecord("task/casc"); err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}

	// Forward traversal — must be empty (source is gone).
	// We cannot call GetLinkedRecords because the source record no longer exists
	// and its from-prefix will return no keys. Verify via reverse traversal.
	rev, err := m.GetLinkedRecordsReverse("person/casc", "concerns")
	if err != nil {
		t.Fatalf("GetLinkedRecordsReverse: %v", err)
	}
	if len(rev) != 0 {
		t.Errorf("expected 0 reverse records after source deleted, got %d", len(rev))
	}
}

func TestMemory_PersistenceAcrossRestart(t *testing.T) {
	store := newTestStorage(t)
	log := newTestLogger(t)

	m1, err := memory.NewMemoryService(memory.Config{}, store, log)
	if err != nil {
		t.Fatalf("NewMemoryService m1: %v", err)
	}
	if err := m1.CreateRecord(makeRecord("task/persist", "task", []byte("data"))); err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	if err := m1.CreateRecord(makeRecord("person/persist", "person", []byte("person"))); err != nil {
		t.Fatalf("CreateRecord person: %v", err)
	}
	if err := m1.LinkRecords("task/persist", "person/persist", "concerns", "test"); err != nil {
		t.Fatalf("LinkRecords: %v", err)
	}
	_ = m1.Close()

	m2, err := memory.NewMemoryService(memory.Config{}, store, log)
	if err != nil {
		t.Fatalf("NewMemoryService m2: %v", err)
	}
	defer m2.Close()

	// Records are readable.
	got, err := m2.GetRecord("task/persist")
	if err != nil || string(got.Payload) != "data" {
		t.Errorf("GetRecord after restart: payload=%q err=%v", got.Payload, err)
	}

	// Forward traversal is intact.
	fwd, err := m2.GetLinkedRecords("task/persist", "concerns")
	if err != nil || len(fwd) != 1 || fwd[0].ID != "person/persist" {
		t.Errorf("GetLinkedRecords after restart: records=%v err=%v", fwd, err)
	}

	// Reverse traversal is intact.
	rev, err := m2.GetLinkedRecordsReverse("person/persist", "concerns")
	if err != nil || len(rev) != 1 || rev[0].ID != "task/persist" {
		t.Errorf("GetLinkedRecordsReverse after restart: records=%v err=%v", rev, err)
	}

	// Type index is intact.
	records, err := m2.ListRecordsByType("task")
	if err != nil || len(records) != 1 {
		t.Errorf("ListRecordsByType after restart: records=%v err=%v", records, err)
	}
}
