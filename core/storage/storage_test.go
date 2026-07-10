// Tests for the Storage Service.
//
// Test file: storage_test.go
// Package: storage_test (external black-box test package)
//
// Using an external test package exercises only the exported API of Storage.
// All 39 tests specified in frozen ADR-CS-002 Section 15 are implemented here.
//
// Run with:
//
//	go test ./core/storage/... -v
//	go test ./core/storage/... -v -race    ← required for concurrency tests
package storage_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"

	"idun/core/logger"
	"idun/core/storage"
)

// ============================================================
// Helpers
// ============================================================

func newTestLogger(t *testing.T) logger.Writer {
	t.Helper()
	l, err := logger.NewLogger(logger.Config{Output: io.Discard})
	if err != nil {
		t.Fatalf("newTestLogger failed: %v", err)
	}
	return l
}

func newTestStorage(t *testing.T) *storage.Storage {
	t.Helper()
	dir := t.TempDir()
	s, err := storage.NewStorage(storage.Config{Path: dir}, newTestLogger(t))
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})
	return s
}

// ============================================================
// 15.1 Constructor Tests
// ============================================================

func TestNewStorage_Success(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.NewStorage(storage.Config{Path: dir}, newTestLogger(t))
	if err != nil {
		t.Fatalf("NewStorage returned unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("NewStorage returned nil Storage")
	}
}

func TestNewStorage_EmptyPath(t *testing.T) {
	s, err := storage.NewStorage(storage.Config{Path: ""}, newTestLogger(t))
	if err == nil {
		t.Error("expected error for empty Path, got nil")
	}
	if s != nil {
		t.Error("expected nil Storage for empty Path")
	}
}

func TestNewStorage_NilLogger(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.NewStorage(storage.Config{Path: dir}, nil)
	if err == nil {
		t.Error("expected error for nil logger, got nil")
	}
	if s != nil {
		t.Error("expected nil Storage for nil logger")
	}
}

// ============================================================
// 15.2 Name / Component Interface Tests
// ============================================================

func TestStorage_Name(t *testing.T) {
	s := newTestStorage(t)
	got := s.Name()
	want := "StorageService"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// Compile-time verification that *Storage implements Store interface.
var _ storage.Store = (*storage.Storage)(nil)

func TestStorage_StoreInterfaceCompliance(t *testing.T) {
	s := newTestStorage(t)
	var _ storage.Store = s
}

// ============================================================
// 15.3 Write Tests
// ============================================================

func TestWrite_Success(t *testing.T) {
	s := newTestStorage(t)
	err := s.Write("test/key", []byte("hello"))
	if err != nil {
		t.Fatalf("Write returned unexpected error: %v", err)
	}
}

func TestWrite_EmptyKey(t *testing.T) {
	s := newTestStorage(t)
	err := s.Write("", []byte("hello"))
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
}

func TestWrite_NilValue(t *testing.T) {
	s := newTestStorage(t)
	err := s.Write("key/nil", nil)
	if err != nil {
		t.Fatalf("Write(nil) returned unexpected error: %v", err)
	}
	exists, err := s.Exists("key/nil")
	if err != nil || !exists {
		t.Errorf("Exists after Write(nil) = %v, %v; want true, nil", exists, err)
	}
}

func TestWrite_EmptyValue(t *testing.T) {
	s := newTestStorage(t)
	err := s.Write("key/empty", []byte{})
	if err != nil {
		t.Fatalf("Write([]byte{}) returned unexpected error: %v", err)
	}
	data, err := s.Read("key/empty")
	if err != nil {
		t.Fatalf("Read after Write empty returned error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected 0 bytes, got %d", len(data))
	}
}

func TestWrite_Overwrite(t *testing.T) {
	s := newTestStorage(t)
	if err := s.Write("k", []byte("first")); err != nil {
		t.Fatalf("first Write failed: %v", err)
	}
	if err := s.Write("k", []byte("second")); err != nil {
		t.Fatalf("second Write failed: %v", err)
	}
	got, err := s.Read("k")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(got) != "second" {
		t.Errorf("Read = %q, want %q", got, "second")
	}
}

// ============================================================
// 15.4 Read Tests
// ============================================================

func TestRead_Success(t *testing.T) {
	s := newTestStorage(t)
	want := []byte("hello storage")
	if err := s.Write("docs/readme", want); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	got, err := s.Read("docs/readme")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("Read = %q, want %q", got, want)
	}
}

func TestRead_EmptyKey(t *testing.T) {
	s := newTestStorage(t)
	data, err := s.Read("")
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
	if data != nil {
		t.Errorf("expected nil data for empty key, got %v", data)
	}
}

func TestRead_NotFound(t *testing.T) {
	s := newTestStorage(t)
	data, err := s.Read("non/existent/key")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if data != nil {
		t.Errorf("expected nil data on ErrNotFound, got %v", data)
	}
}

func TestRead_ReturnsCopy(t *testing.T) {
	s := newTestStorage(t)
	orig := []byte("original")
	if err := s.Write("key", orig); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	first, err := s.Read("key")
	if err != nil {
		t.Fatalf("first Read failed: %v", err)
	}
	first[0] = 'X' // Mutate returned copy

	second, err := s.Read("key")
	if err != nil {
		t.Fatalf("second Read failed: %v", err)
	}
	if string(second) != "original" {
		t.Errorf("stored data was mutated: got %q, want %q", second, "original")
	}
}

// ============================================================
// 15.5 Delete Tests
// ============================================================

func TestDelete_Success(t *testing.T) {
	s := newTestStorage(t)
	if err := s.Write("temp/file", []byte("data")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := s.Delete("temp/file"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err := s.Read("temp/file")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Read after Delete error = %v, want ErrNotFound", err)
	}
}

func TestDelete_EmptyKey(t *testing.T) {
	s := newTestStorage(t)
	err := s.Delete("")
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
}

func TestDelete_NotFound(t *testing.T) {
	s := newTestStorage(t)
	err := s.Delete("missing/key")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ============================================================
// 15.6 Exists Tests
// ============================================================

func TestExists_True(t *testing.T) {
	s := newTestStorage(t)
	if err := s.Write("exist/key", []byte("yes")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	exists, err := s.Exists("exist/key")
	if err != nil {
		t.Fatalf("Exists returned unexpected error: %v", err)
	}
	if !exists {
		t.Error("Exists = false, want true")
	}
}

func TestExists_False(t *testing.T) {
	s := newTestStorage(t)
	exists, err := s.Exists("no/such/key")
	if err != nil {
		t.Fatalf("Exists returned unexpected error: %v", err)
	}
	if exists {
		t.Error("Exists = true, want false")
	}
}

func TestExists_EmptyKey(t *testing.T) {
	s := newTestStorage(t)
	exists, err := s.Exists("")
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
	if exists {
		t.Error("expected false for empty key")
	}
}

func TestExists_AfterDelete(t *testing.T) {
	s := newTestStorage(t)
	if err := s.Write("key", []byte("data")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := s.Delete("key"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	exists, err := s.Exists("key")
	if err != nil {
		t.Fatalf("Exists after Delete returned unexpected error: %v", err)
	}
	if exists {
		t.Error("Exists after Delete = true, want false")
	}
}

// ============================================================
// 15.7 List Tests
// ============================================================

func TestList_Empty(t *testing.T) {
	s := newTestStorage(t)
	keys, err := s.List("")
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if keys == nil {
		t.Fatal("List returned nil slice, expected non-nil empty slice")
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestList_AllKeys(t *testing.T) {
	s := newTestStorage(t)
	inputs := []string{"a/1", "b/1", "c/1"}
	for _, k := range inputs {
		if err := s.Write(k, []byte("val")); err != nil {
			t.Fatalf("Write(%q) failed: %v", k, err)
		}
	}
	keys, err := s.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if !reflect.DeepEqual(keys, inputs) {
		t.Errorf("List = %v, want %v", keys, inputs)
	}
}

func TestList_PrefixFilter(t *testing.T) {
	s := newTestStorage(t)
	for _, k := range []string{"a/1", "a/2", "b/1"} {
		if err := s.Write(k, []byte("val")); err != nil {
			t.Fatalf("Write(%q) failed: %v", k, err)
		}
	}
	keys, err := s.List("a/")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	want := []string{"a/1", "a/2"}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("List = %v, want %v", keys, want)
	}
}

func TestList_Sorted(t *testing.T) {
	s := newTestStorage(t)
	for _, k := range []string{"c/3", "a/1", "b/2"} {
		if err := s.Write(k, []byte("val")); err != nil {
			t.Fatalf("Write(%q) failed: %v", k, err)
		}
	}
	keys, err := s.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	want := []string{"a/1", "b/2", "c/3"}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("List = %v, want %v", keys, want)
	}
}

func TestList_NonMatchingPrefix(t *testing.T) {
	s := newTestStorage(t)
	if err := s.Write("a/1", []byte("val")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	keys, err := s.List("z/")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if keys == nil || len(keys) != 0 {
		t.Errorf("expected non-nil empty slice, got %v", keys)
	}
}

func TestList_ExactKeyAsPrefix(t *testing.T) {
	s := newTestStorage(t)
	for _, k := range []string{"exact/key", "exact/key_child"} {
		if err := s.Write(k, []byte("val")); err != nil {
			t.Fatalf("Write(%q) failed: %v", k, err)
		}
	}
	keys, err := s.List("exact/key")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	want := []string{"exact/key", "exact/key_child"}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("List = %v, want %v", keys, want)
	}
}

// ============================================================
// 15.8 ErrNotFound Tests
// ============================================================

func TestErrNotFound_IsDistinctFromIOError(t *testing.T) {
	if errors.Is(storage.ErrNotFound, os.ErrPermission) {
		t.Error("ErrNotFound should not match os.ErrPermission")
	}
	if !errors.Is(storage.ErrNotFound, storage.ErrNotFound) {
		t.Error("ErrNotFound should match itself")
	}
}

func TestErrNotFound_ReadUsesIt(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.Read("missing")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound from Read, got %v", err)
	}
}

func TestErrNotFound_DeleteUsesIt(t *testing.T) {
	s := newTestStorage(t)
	err := s.Delete("missing")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound from Delete, got %v", err)
	}
}

// ============================================================
// 15.9 Round-Trip Tests
// ============================================================

func TestRoundTrip_BinaryData(t *testing.T) {
	s := newTestStorage(t)
	want := []byte{0x00, 0xFF, 0x01, 0xFE, 0x80, 0x7F}
	if err := s.Write("binary/data", want); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	got, err := s.Read("binary/data")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("round trip mismatch: got %v, want %v", got, want)
	}
}

func TestRoundTrip_LargeValue(t *testing.T) {
	s := newTestStorage(t)
	const size = 1 * 1024 * 1024 // 1 MB
	want := make([]byte, size)
	for i := 0; i < size; i++ {
		want[i] = byte(i % 256)
	}
	if err := s.Write("large/value", want); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	got, err := s.Read("large/value")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Error("round trip mismatch for 1MB payload")
	}
}

// ============================================================
// 15.10 Goroutine Safety Tests
// ============================================================

func TestStorage_ConcurrentWrites_NoPanic(t *testing.T) {
	s := newTestStorage(t)
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			_ = s.Write("concurrent/key", []byte("val"))
		}(i)
	}
	wg.Wait()
}

func TestStorage_ConcurrentReadWrite_NoPanic(t *testing.T) {
	s := newTestStorage(t)
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = s.Write("shared/key", []byte("val"))
		}()
		go func() {
			defer wg.Done()
			_, _ = s.Read("shared/key")
		}()
	}
	wg.Wait()
}

// ============================================================
// 15.11 Lifecycle Tests
// ============================================================

func TestStorage_Close_Success(t *testing.T) {
	s := newTestStorage(t)
	if err := s.Close(); err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}
}

func TestStorage_Close_Idempotent(t *testing.T) {
	s := newTestStorage(t)
	if err := s.Close(); err != nil {
		t.Errorf("first Close returned error: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("second Close returned error: %v", err)
	}
}

func TestStorage_AfterClose_ReturnsError(t *testing.T) {
	s := newTestStorage(t)
	if err := s.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if err := s.Write("k", []byte("v")); err == nil {
		t.Error("expected error from Write after Close, got nil")
	}
	if _, err := s.Read("k"); err == nil {
		t.Error("expected error from Read after Close, got nil")
	}
	if err := s.Delete("k"); err == nil {
		t.Error("expected error from Delete after Close, got nil")
	}
	if _, err := s.Exists("k"); err == nil {
		t.Error("expected error from Exists after Close, got nil")
	}
	if _, err := s.List(""); err == nil {
		t.Error("expected error from List after Close, got nil")
	}
}

// ============================================================
// 15.12 Integration Tests
// ============================================================

type stubRegistry struct {
	components []interface{ Name() string }
}

func (r *stubRegistry) register(c interface{ Name() string }) {
	r.components = append(r.components, c)
}

func TestStorage_RegistersInKernelRegistry(t *testing.T) {
	s := newTestStorage(t)
	reg := &stubRegistry{}
	reg.register(s)

	if len(reg.components) != 1 || reg.components[0].Name() != "StorageService" {
		t.Errorf("expected StorageService registered, got %v", reg.components)
	}
}

func TestStorage_MemoryServiceCanUseStoreInterface(t *testing.T) {
	s := newTestStorage(t)
	var store storage.Store = s // Caller holds only capability interface

	if err := store.Write("mem/rec1", []byte("data")); err != nil {
		t.Fatalf("Write via Store interface failed: %v", err)
	}
	got, err := store.Read("mem/rec1")
	if err != nil || string(got) != "data" {
		t.Fatalf("Read via Store interface = %q, %v", got, err)
	}
	exists, err := store.Exists("mem/rec1")
	if err != nil || !exists {
		t.Fatalf("Exists via Store interface = %v, %v", exists, err)
	}
	keys, err := store.List("mem/")
	if err != nil || !reflect.DeepEqual(keys, []string{"mem/rec1"}) {
		t.Fatalf("List via Store interface = %v, %v", keys, err)
	}
	if err := store.Delete("mem/rec1"); err != nil {
		t.Fatalf("Delete via Store interface failed: %v", err)
	}
}
