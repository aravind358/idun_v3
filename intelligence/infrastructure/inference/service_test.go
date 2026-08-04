package inference_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"idun/core/storage"
	"idun/intelligence/infrastructure/inference"
	"idun/intelligence/infrastructure/registry"
)

type memStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newMemStore() *memStore {
	return &memStore{data: make(map[string][]byte)}
}

func (m *memStore) Write(key string, val []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]byte, len(val))
	copy(cp, val)
	m.data[key] = cp
	return nil
}

func (m *memStore) Read(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return nil, storage.ErrNotFound
	}
	cp := make([]byte, len(val))
	copy(cp, val)
	return cp, nil
}

func (m *memStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *memStore) Exists(key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}

func (m *memStore) List(prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []string
	for k := range m.data {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out, nil
}

func (m *memStore) StoreArtifact(artifactID string, meta storage.ArtifactIndexMeta, payloadRef string, data []byte) error {
	return m.Write(payloadRef, data)
}

func (m *memStore) LookupArtifact(artifactID string) (storage.ArtifactIndexMeta, error) {
	return storage.ArtifactIndexMeta{}, nil
}

func setupRegistry(t *testing.T) *registry.Service {
	reg := registry.NewService()
	_ = reg.Start()
	bd := registry.BackendDescriptor{
		ID:             "local-infer-01",
		DriverScheme:   "local-bin",
		Endpoint:       "bin/infer",
		Version:        "1.0",
		MaxConcurrency: 4,
	}
	if err := reg.Register(context.Background(), "language-reasoner", bd); err != nil {
		t.Fatalf("setup registry: %v", err)
	}
	return reg
}

func TestLifecycle(t *testing.T) {
	svc := inference.NewService()
	if svc.Name() != "Intelligence.Infrastructure.Inference" {
		t.Fatalf("unexpected name: %s", svc.Name())
	}
	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := svc.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	_, err := svc.Execute(context.Background(), inference.InferenceRequest{ModelID: "m1", InputRef: "in1"})
	if !errors.Is(err, inference.ErrServiceClosed) {
		t.Fatalf("expected ErrServiceClosed after close, got: %v", err)
	}
}

func TestExecuteAndCache(t *testing.T) {
	reg := setupRegistry(t)
	st := newMemStore()
	svc := inference.NewService(
		inference.WithResolver(reg),
		inference.WithStorage(st),
	)

	ctx := context.Background()
	req := inference.InferenceRequest{
		ModelID:  "language-reasoner",
		InputRef: "storage://input/item-1",
		Modality: inference.ModalityText,
		Budget:   "STANDARD",
	}

	res1, err := svc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if res1.Cached {
		t.Fatalf("expected initial execution to not be cached")
	}

	// Second execution with same parameters should hit exact cache
	res2, err := svc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("second Execute failed: %v", err)
	}
	if !res2.Cached {
		t.Fatalf("expected second execution to be cached")
	}
	if res1.OutputRef != res2.OutputRef {
		t.Fatalf("cache returned different OutputRef: %s vs %s", res1.OutputRef, res2.OutputRef)
	}
}

func TestExecuteInvalidAndCancelled(t *testing.T) {
	reg := setupRegistry(t)
	svc := inference.NewService(inference.WithResolver(reg))

	ctx := context.Background()
	_, err := svc.Execute(ctx, inference.InferenceRequest{})
	if !errors.Is(err, inference.ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest on empty request, got: %v", err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = svc.Execute(cancelCtx, inference.InferenceRequest{ModelID: "language-reasoner", InputRef: "in1"})
	if !errors.Is(err, inference.ErrInferenceCancelled) {
		t.Fatalf("expected ErrInferenceCancelled, got: %v", err)
	}
}

func TestExecuteStream(t *testing.T) {
	reg := setupRegistry(t)
	svc := inference.NewService(inference.WithResolver(reg))

	ch := make(chan inference.StreamChunk, 2)
	req := inference.InferenceRequest{
		ModelID:  "language-reasoner",
		InputRef: "in-stream",
		Budget:   "STANDARD",
	}

	if err := svc.ExecuteStream(context.Background(), req, ch); err != nil {
		t.Fatalf("ExecuteStream failed: %v", err)
	}

	chunk1 := <-ch
	if chunk1.ChunkRef == "" || chunk1.Done {
		t.Fatalf("unexpected chunk1: %+v", chunk1)
	}
	chunk2 := <-ch
	if !chunk2.Done {
		t.Fatalf("expected chunk2.Done=true: %+v", chunk2)
	}
}

func TestConcurrentInference(t *testing.T) {
	reg := setupRegistry(t)
	st := newMemStore()
	svc := inference.NewService(
		inference.WithResolver(reg),
		inference.WithStorage(st),
	)

	var wg sync.WaitGroup
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			budget := "STANDARD"
			if idx%2 == 0 {
				budget = "REFLEXIVE"
			}
			req := inference.InferenceRequest{
				ModelID:  "language-reasoner",
				InputRef: "storage://input/item",
				Budget:   budget,
			}
			_, _ = svc.Execute(ctx, req)
		}(i)
	}
	wg.Wait()

	snap := svc.GetTelemetry()
	if snap.TotalExecutions < 1 {
		t.Fatalf("expected positive TotalExecutions, got %d", snap.TotalExecutions)
	}
}
