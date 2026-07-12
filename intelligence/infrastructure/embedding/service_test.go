package embedding_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"idun/intelligence/infrastructure/embedding"
	"idun/intelligence/infrastructure/registry"
)

func setupRegistry(t *testing.T) *registry.Service {
	reg := registry.NewService()
	_ = reg.Start()
	bd := registry.BackendDescriptor{
		ID:             "embed-backend-01",
		DriverScheme:   "local-bin",
		Endpoint:       "bin/embed",
		Version:        "1.0",
		MaxConcurrency: 4,
	}
	if err := reg.Register(context.Background(), "canonical-embedder", bd); err != nil {
		t.Fatalf("setup registry: %v", err)
	}
	return reg
}

func TestLifecycle(t *testing.T) {
	svc := embedding.NewService()
	if svc.Name() != "Intelligence.Infrastructure.Embedding" {
		t.Fatalf("unexpected name: %s", svc.Name())
	}
	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := svc.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	_, err := svc.Embed(context.Background(), embedding.EmbeddingRequest{ContentRef: "c1"})
	if !errors.Is(err, embedding.ErrServiceClosed) {
		t.Fatalf("expected ErrServiceClosed after close, got: %v", err)
	}
}

func TestEmbedAndCache(t *testing.T) {
	reg := setupRegistry(t)
	svc := embedding.NewService(embedding.WithResolver(reg))

	ctx := context.Background()
	req := embedding.EmbeddingRequest{
		ContentRef: "storage://doc/item-1",
		Modality:   "text",
	}

	res1, err := svc.Embed(ctx, req)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if res1.VectorRef == "" {
		t.Fatalf("empty VectorRef returned")
	}
	if res1.Cached {
		t.Fatalf("expected initial embed not cached")
	}

	res2, err := svc.Embed(ctx, req)
	if err != nil {
		t.Fatalf("second Embed failed: %v", err)
	}
	if !res2.Cached {
		t.Fatalf("expected second embed cached")
	}
	if res1.VectorRef != res2.VectorRef {
		t.Fatalf("cache returned different VectorRef: %s vs %s", res1.VectorRef, res2.VectorRef)
	}
}

func TestEmbedBatch(t *testing.T) {
	reg := setupRegistry(t)
	svc := embedding.NewService(embedding.WithResolver(reg))

	ctx := context.Background()
	reqs := []embedding.EmbeddingRequest{
		{ContentRef: "doc1", Modality: "text"},
		{ContentRef: "doc2", Modality: "text"},
		{ContentRef: "doc3", Modality: "text"},
	}

	results, err := svc.EmbedBatch(ctx, reqs)
	if err != nil {
		t.Fatalf("EmbedBatch failed: %v", err)
	}
	if len(results) != len(reqs) {
		t.Fatalf("expected %d results, got %d", len(reqs), len(results))
	}
}

func TestSimilarity(t *testing.T) {
	svc := embedding.NewService()
	ctx := context.Background()

	scoreIdentity, err := svc.Similarity(ctx, "vector-A", "vector-A")
	if err != nil {
		t.Fatalf("Similarity failed: %v", err)
	}
	if scoreIdentity != 1.0 {
		t.Fatalf("expected 1.0 for identical vectors, got %f", scoreIdentity)
	}

	scoreDiff, err := svc.Similarity(ctx, "vector-A", "vector-B")
	if err != nil {
		t.Fatalf("Similarity diff failed: %v", err)
	}
	if scoreDiff < 0.0 || scoreDiff > 1.0 {
		t.Fatalf("expected similarity score in [0.0, 1.0], got %f", scoreDiff)
	}
}

func TestConcurrentEmbedding(t *testing.T) {
	reg := setupRegistry(t)
	svc := embedding.NewService(embedding.WithResolver(reg))

	var wg sync.WaitGroup
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				_, _ = svc.Embed(ctx, embedding.EmbeddingRequest{ContentRef: "doc-concurrency"})
			} else {
				_, _ = svc.Similarity(ctx, "v1", "v2")
			}
		}(i)
	}
	wg.Wait()

	snap := svc.GetTelemetry()
	if snap.TotalEmbeddings < 1 || snap.TotalSimilarityQueries < 1 {
		t.Fatalf("unexpected telemetry snapshot: %+v", snap)
	}
}
