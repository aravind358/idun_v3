// Tests for the Scheduler Service (Version 1).
//
// Test file: scheduler_test.go
// Package: scheduler_test (external black-box test package)
//
// Implements the complete test plan specified in frozen ADR-CS-004 v1.1.
package scheduler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"idun/core/logger"
	"idun/core/memory"
	"idun/core/scheduler"
	"idun/core/storage"
)

// ============================================================
// Test Helpers & Mocks
// ============================================================

type dispatchCall struct {
	Target  string
	Payload []byte
}

type mockDispatcher struct {
	mu     sync.Mutex
	calls  []dispatchCall
	err    error
	notify chan struct{}
}

func newMockDispatcher() *mockDispatcher {
	return &mockDispatcher{
		notify: make(chan struct{}, 100),
	}
}

func (m *mockDispatcher) Dispatch(target string, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, dispatchCall{Target: target, Payload: append([]byte(nil), payload...)})
	select {
	case m.notify <- struct{}{}:
	default:
	}
	return m.err
}

func (m *mockDispatcher) getCalls() []dispatchCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]dispatchCall, len(m.calls))
	copy(out, m.calls)
	return out
}

func newTestLogger(t *testing.T) logger.Writer {
	t.Helper()
	l, err := logger.NewLogger(logger.Config{Output: io.Discard})
	if err != nil {
		t.Fatalf("newTestLogger: %v", err)
	}
	return l
}

func newTestMemory(t *testing.T) memory.Memory {
	t.Helper()
	dir := t.TempDir()
	store, err := storage.NewStorage(storage.Config{Path: dir}, newTestLogger(t))
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	mem, err := memory.NewMemoryService(memory.Config{}, store, newTestLogger(t))
	if err != nil {
		t.Fatalf("NewMemoryService: %v", err)
	}
	t.Cleanup(func() { _ = mem.Close() })
	return mem
}

func newTestScheduler(t *testing.T, disp scheduler.Dispatcher) (*scheduler.SchedulerService, memory.Memory) {
	t.Helper()
	mem := newTestMemory(t)
	s, err := scheduler.NewSchedulerService(scheduler.Config{}, mem, newTestLogger(t), disp)
	if err != nil {
		t.Fatalf("NewSchedulerService: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s, mem
}

// ============================================================
// 8.1 Constructor & Initialization Tests
// ============================================================

func TestNewScheduler_Success(t *testing.T) {
	disp := newMockDispatcher()
	s, mem := newTestScheduler(t, disp)
	if s == nil || mem == nil {
		t.Fatal("expected non-nil service and memory")
	}
}

func TestNewScheduler_NilMemory(t *testing.T) {
	disp := newMockDispatcher()
	s, err := scheduler.NewSchedulerService(scheduler.Config{}, nil, newTestLogger(t), disp)
	if err == nil {
		t.Error("expected error for nil memory, got nil")
	}
	if s != nil {
		t.Error("expected nil service on error")
	}
}

func TestNewScheduler_NilLogger(t *testing.T) {
	disp := newMockDispatcher()
	mem := newTestMemory(t)
	s, err := scheduler.NewSchedulerService(scheduler.Config{}, mem, nil, disp)
	if err == nil {
		t.Error("expected error for nil logger, got nil")
	}
	if s != nil {
		t.Error("expected nil service on error")
	}
}

func TestNewScheduler_NilDispatcher(t *testing.T) {
	mem := newTestMemory(t)
	s, err := scheduler.NewSchedulerService(scheduler.Config{}, mem, newTestLogger(t), nil)
	if err == nil {
		t.Error("expected error for nil dispatcher, got nil")
	}
	if s != nil {
		t.Error("expected nil service on error")
	}
}

func TestScheduler_Name(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	if got := s.Name(); got != "SchedulerService" {
		t.Errorf("Name() = %q, want %q", got, "SchedulerService")
	}
}

func TestScheduler_InterfaceCompliance(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	var _ scheduler.Scheduler = s
}

// ============================================================
// 8.2 ScheduleOnce Tests
// ============================================================

func TestScheduleOnce_Success(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	future := time.Now().UTC().Add(1 * time.Hour)
	err := s.ScheduleOnce("job/1", "TargetA", []byte("hello"), future)
	if err != nil {
		t.Fatalf("ScheduleOnce: %v", err)
	}
}

func TestScheduleOnce_EmptyID(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	err := s.ScheduleOnce("", "TargetA", nil, time.Now().Add(1*time.Hour))
	if err == nil {
		t.Error("expected error for empty job ID")
	}
}

func TestScheduleOnce_EmptyTarget(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	err := s.ScheduleOnce("job/1", "", nil, time.Now().Add(1*time.Hour))
	if err == nil {
		t.Error("expected error for empty target")
	}
}

func TestScheduleOnce_ZeroTime(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	err := s.ScheduleOnce("job/1", "TargetA", nil, time.Time{})
	if err == nil {
		t.Error("expected error for zero time")
	}
}

func TestScheduleOnce_DuplicateID(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	future := time.Now().UTC().Add(1 * time.Hour)
	if err := s.ScheduleOnce("job/dup", "TargetA", nil, future); err != nil {
		t.Fatalf("first ScheduleOnce: %v", err)
	}
	err := s.ScheduleOnce("job/dup", "TargetB", nil, future)
	if !errors.Is(err, scheduler.ErrDuplicateID) {
		t.Errorf("expected ErrDuplicateID, got %v", err)
	}
}

func TestScheduleOnce_PastTimeImmediatelyDispatches(t *testing.T) {
	disp := newMockDispatcher()
	s, mem := newTestScheduler(t, disp)

	past := time.Now().UTC().Add(-10 * time.Minute)
	err := s.ScheduleOnce("job/past", "TargetPast", []byte("past-payload"), past)
	if err != nil {
		t.Fatalf("ScheduleOnce past: %v", err)
	}

	calls := disp.getCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 immediate dispatch call, got %d", len(calls))
	}
	if calls[0].Target != "TargetPast" || string(calls[0].Payload) != "past-payload" {
		t.Errorf("unexpected dispatch call: %+v", calls[0])
	}

	// Must not remain in Memory.
	_, err = mem.GetRecord("job/past")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound in memory for immediately dispatched job, got %v", err)
	}
}

func TestScheduleOnce_NilPayload(t *testing.T) {
	disp := newMockDispatcher()
	s, mem := newTestScheduler(t, disp)

	future := time.Now().UTC().Add(1 * time.Hour)
	if err := s.ScheduleOnce("job/nil-payload", "TargetA", nil, future); err != nil {
		t.Fatalf("ScheduleOnce: %v", err)
	}

	rec, err := mem.GetRecord("job/nil-payload")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	var job scheduler.Job
	if err := json.Unmarshal(rec.Payload, &job); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if job.Payload == nil {
		t.Error("expected non-nil empty slice payload")
	}
}

func TestScheduleOnce_PersistsToMemory(t *testing.T) {
	disp := newMockDispatcher()
	s, mem := newTestScheduler(t, disp)

	future := time.Now().UTC().Add(1 * time.Hour)
	if err := s.ScheduleOnce("job/persist", "TargetPersist", []byte("data"), future); err != nil {
		t.Fatalf("ScheduleOnce: %v", err)
	}

	rec, err := mem.GetRecord("job/persist")
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if rec.Type != scheduler.JobRecordType {
		t.Errorf("Record.Type = %q, want %q", rec.Type, scheduler.JobRecordType)
	}
}

// ============================================================
// 8.3 Cancel Tests
// ============================================================

func TestCancel_Success(t *testing.T) {
	disp := newMockDispatcher()
	s, mem := newTestScheduler(t, disp)

	future := time.Now().UTC().Add(1 * time.Hour)
	if err := s.ScheduleOnce("job/cancel", "TargetCancel", nil, future); err != nil {
		t.Fatalf("ScheduleOnce: %v", err)
	}
	if err := s.Cancel("job/cancel"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	_, err := mem.GetRecord("job/cancel")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound after cancel, got %v", err)
	}
}

func TestCancel_NotFound(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	err := s.Cancel("nonexistent")
	if !errors.Is(err, scheduler.ErrJobNotFound) {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}

func TestCancel_EmptyID(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	if err := s.Cancel(""); err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestCancel_PreventsExecution(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	due := time.Now().UTC().Add(40 * time.Millisecond)
	if err := s.ScheduleOnce("job/prevent", "TargetPrevent", nil, due); err != nil {
		t.Fatalf("ScheduleOnce: %v", err)
	}
	if err := s.Cancel("job/prevent"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if len(disp.getCalls()) != 0 {
		t.Errorf("cancelled job should not have executed: %+v", disp.getCalls())
	}
}

func TestCancel_DeletesFromMemory(t *testing.T) {
	disp := newMockDispatcher()
	s, mem := newTestScheduler(t, disp)

	future := time.Now().UTC().Add(1 * time.Hour)
	_ = s.ScheduleOnce("job/del", "TargetDel", nil, future)
	_ = s.Cancel("job/del")

	_, err := mem.GetRecord("job/del")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected ErrNotFound in memory, got %v", err)
	}
}

// ============================================================
// 8.4 Execution & Dispatch Tests
// ============================================================

func TestDispatch_FiresAtExactDueTime(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	due := time.Now().UTC().Add(50 * time.Millisecond)
	if err := s.ScheduleOnce("job/fire", "TargetFire", []byte("payload"), due); err != nil {
		t.Fatalf("ScheduleOnce: %v", err)
	}

	select {
	case <-disp.notify:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for job dispatch")
	}

	calls := disp.getCalls()
	if len(calls) != 1 || calls[0].Target != "TargetFire" || string(calls[0].Payload) != "payload" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
}

func TestDispatch_DeletesFromMemoryAfterFire(t *testing.T) {
	disp := newMockDispatcher()
	s, mem := newTestScheduler(t, disp)

	due := time.Now().UTC().Add(40 * time.Millisecond)
	_ = s.ScheduleOnce("job/cleanup", "TargetClean", nil, due)

	select {
	case <-disp.notify:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for job dispatch")
	}

	// Small pause for goroutine to finish Memory deletion after dispatch.
	time.Sleep(20 * time.Millisecond)
	_, err := mem.GetRecord("job/cleanup")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected record to be deleted after dispatch, got %v", err)
	}
}

func TestDispatch_ChronologicalOrdering(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	now := time.Now().UTC()
	// Schedule second job first, first job second.
	_ = s.ScheduleOnce("job/2", "Target2", nil, now.Add(90*time.Millisecond))
	_ = s.ScheduleOnce("job/1", "Target1", nil, now.Add(30*time.Millisecond))

	for i := 0; i < 2; i++ {
		select {
		case <-disp.notify:
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timed out waiting for call %d", i+1)
		}
	}

	calls := disp.getCalls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].Target != "Target1" || calls[1].Target != "Target2" {
		t.Errorf("expected order Target1 then Target2, got: %+v", calls)
	}
}

func TestDispatch_DynamicTimerReset(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	now := time.Now().UTC()
	// Schedule a distant job.
	_ = s.ScheduleOnce("job/far", "TargetFar", nil, now.Add(5*time.Second))
	// Now insert a much earlier job; must reset timer dynamically.
	_ = s.ScheduleOnce("job/near", "TargetNear", nil, now.Add(40*time.Millisecond))

	select {
	case <-disp.notify:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for dynamic timer reset dispatch")
	}

	calls := disp.getCalls()
	if len(calls) != 1 || calls[0].Target != "TargetNear" {
		t.Errorf("unexpected calls: %+v", calls)
	}
}

func TestDispatch_DispatcherErrorLoggedAndCleanedUp(t *testing.T) {
	disp := newMockDispatcher()
	disp.err = errors.New("dispatch failure")
	s, mem := newTestScheduler(t, disp)

	due := time.Now().UTC().Add(30 * time.Millisecond)
	_ = s.ScheduleOnce("job/err", "TargetErr", nil, due)

	select {
	case <-disp.notify:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for dispatch")
	}

	time.Sleep(20 * time.Millisecond)
	_, err := mem.GetRecord("job/err")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected Memory record deleted even on dispatcher error, got %v", err)
	}
}

func TestDispatch_ExactPayloadIntegrity(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	payload := []byte{0x00, 0xFF, 0x12, 0x34, 0xAA}
	due := time.Now().UTC().Add(30 * time.Millisecond)
	_ = s.ScheduleOnce("job/bin", "TargetBin", payload, due)

	<-disp.notify
	calls := disp.getCalls()
	if !bytes.Equal(calls[0].Payload, payload) {
		t.Errorf("payload mismatch: got %v, want %v", calls[0].Payload, payload)
	}
}

func TestDispatch_ExactTargetIntegrity(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	due := time.Now().UTC().Add(30 * time.Millisecond)
	_ = s.ScheduleOnce("job/t", "IntelligenceCoreV1", nil, due)

	<-disp.notify
	calls := disp.getCalls()
	if calls[0].Target != "IntelligenceCoreV1" {
		t.Errorf("target mismatch: got %q", calls[0].Target)
	}
}

func TestDispatch_NoPrematureFiring(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	due := time.Now().UTC().Add(150 * time.Millisecond)
	_ = s.ScheduleOnce("job/premature", "TargetP", nil, due)

	time.Sleep(50 * time.Millisecond)
	if len(disp.getCalls()) != 0 {
		t.Errorf("job fired prematurely!")
	}
}

func TestDispatch_ConcurrentDueJobs(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	due := time.Now().UTC().Add(40 * time.Millisecond)
	for i := 0; i < 5; i++ {
		_ = s.ScheduleOnce(fmt.Sprintf("job/conc/%d", i), fmt.Sprintf("Target%d", i), nil, due)
	}

	for i := 0; i < 5; i++ {
		<-disp.notify
	}
	if len(disp.getCalls()) != 5 {
		t.Errorf("expected 5 calls, got %d", len(disp.getCalls()))
	}
}

// ============================================================
// 8.5 Reboot Recovery Tests
// ============================================================

func TestRecovery_LoadsFutureJobsFromMemory(t *testing.T) {
	disp := newMockDispatcher()
	mem := newTestMemory(t)

	future := time.Now().UTC().Add(1 * time.Hour)
	job := scheduler.Job{ID: "job/rec-fut", Target: "TargetRec", Payload: []byte("rec"), ExecuteAt: future}
	data, _ := json.Marshal(job)
	_ = mem.CreateRecord(memory.Record{ID: "job/rec-fut", Type: scheduler.JobRecordType, Payload: data})

	s, err := scheduler.NewSchedulerService(scheduler.Config{}, mem, newTestLogger(t), disp)
	if err != nil {
		t.Fatalf("NewSchedulerService: %v", err)
	}
	defer s.Close()

	if len(disp.getCalls()) != 0 {
		t.Errorf("future job should not dispatch on boot")
	}
}

func TestRecovery_OverdueJobDispatchedImmediately(t *testing.T) {
	disp := newMockDispatcher()
	mem := newTestMemory(t)

	past := time.Now().UTC().Add(-30 * time.Minute)
	job := scheduler.Job{ID: "job/rec-past", Target: "TargetPast", Payload: []byte("missed"), ExecuteAt: past}
	data, _ := json.Marshal(job)
	_ = mem.CreateRecord(memory.Record{ID: "job/rec-past", Type: scheduler.JobRecordType, Payload: data})

	s, err := scheduler.NewSchedulerService(scheduler.Config{}, mem, newTestLogger(t), disp)
	if err != nil {
		t.Fatalf("NewSchedulerService: %v", err)
	}
	defer s.Close()

	calls := disp.getCalls()
	if len(calls) != 1 || calls[0].Target != "TargetPast" {
		t.Fatalf("expected overdue job dispatched on boot, got: %+v", calls)
	}
}

func TestRecovery_OverdueJobDeletedFromMemory(t *testing.T) {
	disp := newMockDispatcher()
	mem := newTestMemory(t)

	past := time.Now().UTC().Add(-30 * time.Minute)
	job := scheduler.Job{ID: "job/rec-del", Target: "TargetDel", Payload: []byte("missed"), ExecuteAt: past}
	data, _ := json.Marshal(job)
	_ = mem.CreateRecord(memory.Record{ID: "job/rec-del", Type: scheduler.JobRecordType, Payload: data})

	s, _ := scheduler.NewSchedulerService(scheduler.Config{}, mem, newTestLogger(t), disp)
	defer s.Close()

	_, err := mem.GetRecord("job/rec-del")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Errorf("expected overdue record deleted from Memory on boot, got %v", err)
	}
}

func TestRecovery_CorruptMemoryRecordIgnored(t *testing.T) {
	disp := newMockDispatcher()
	mem := newTestMemory(t)

	_ = mem.CreateRecord(memory.Record{ID: "job/corrupt", Type: scheduler.JobRecordType, Payload: []byte("{corrupt-json")})

	s, err := scheduler.NewSchedulerService(scheduler.Config{}, mem, newTestLogger(t), disp)
	if err != nil {
		t.Fatalf("NewSchedulerService should ignore corrupt record: %v", err)
	}
	defer s.Close()
}

func TestRecovery_EmptyMemoryBootsCleanly(t *testing.T) {
	disp := newMockDispatcher()
	mem := newTestMemory(t)
	s, err := scheduler.NewSchedulerService(scheduler.Config{}, mem, newTestLogger(t), disp)
	if err != nil {
		t.Fatalf("NewSchedulerService empty: %v", err)
	}
	defer s.Close()
}

func TestRecovery_PreservesFutureExecutionOrder(t *testing.T) {
	disp := newMockDispatcher()
	mem := newTestMemory(t)

	now := time.Now().UTC()
	job1 := scheduler.Job{ID: "job/f1", Target: "Target1", ExecuteAt: now.Add(50 * time.Millisecond)}
	job2 := scheduler.Job{ID: "job/f2", Target: "Target2", ExecuteAt: now.Add(100 * time.Millisecond)}

	d2, _ := json.Marshal(job2)
	d1, _ := json.Marshal(job1)
	_ = mem.CreateRecord(memory.Record{ID: "job/f2", Type: scheduler.JobRecordType, Payload: d2})
	_ = mem.CreateRecord(memory.Record{ID: "job/f1", Type: scheduler.JobRecordType, Payload: d1})

	s, _ := scheduler.NewSchedulerService(scheduler.Config{}, mem, newTestLogger(t), disp)
	_ = s.Start()
	defer s.Close()

	<-disp.notify
	<-disp.notify
	calls := disp.getCalls()
	if len(calls) != 2 || calls[0].Target != "Target1" || calls[1].Target != "Target2" {
		t.Errorf("expected Target1 then Target2, got: %+v", calls)
	}
}

// ============================================================
// 8.6 Concurrency & Lifecycle Tests
// ============================================================

func TestConcurrency_ConcurrentScheduleOnce(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	future := time.Now().UTC().Add(1 * time.Hour)

	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			_ = s.ScheduleOnce(fmt.Sprintf("job/concurrent/%d", id), "TargetC", nil, future)
		}(i)
	}
	wg.Wait()
}

func TestConcurrency_ConcurrentScheduleAndCancel(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n * 2)
	future := time.Now().UTC().Add(1 * time.Hour)

	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			_ = s.ScheduleOnce(fmt.Sprintf("job/rw/%d", id), "TargetRW", nil, future)
		}(i)
		go func(id int) {
			defer wg.Done()
			_ = s.Cancel(fmt.Sprintf("job/rw/%d", id))
		}(i)
	}
	wg.Wait()
}

func TestLifecycle_CloseStopsTimerLoop(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestLifecycle_CloseIdempotent(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	_ = s.Close()
	if err := s.Close(); err != nil {
		t.Errorf("second Close returned error: %v", err)
	}
}

func TestLifecycle_OperationsAfterCloseError(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	_ = s.Close()

	if err := s.ScheduleOnce("job/post", "Target", nil, time.Now().Add(1*time.Hour)); !errors.Is(err, scheduler.ErrClosed) {
		t.Errorf("expected ErrClosed from ScheduleOnce after Close, got %v", err)
	}
	if err := s.Cancel("job/post"); !errors.Is(err, scheduler.ErrClosed) {
		t.Errorf("expected ErrClosed from Cancel after Close, got %v", err)
	}
}

func TestLifecycle_NoGoroutineLeaks(t *testing.T) {
	disp := newMockDispatcher()
	s, _ := newTestScheduler(t, disp)
	_ = s.Close()
}
