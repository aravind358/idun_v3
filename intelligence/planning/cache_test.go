package planning

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestReflexivePlanningCache_LifecycleAndBounding(t *testing.T) {
	cache := NewReflexivePlanningCache("ep-100", "2.0")
	if cache.EpisodeID != "ep-100" || cache.StrategyVersion != "2.0" {
		t.Fatalf("unexpected cache initialization: %+v", cache)
	}

	// Record evaluations and verify counters
	cache.RecordEvaluation(true, 10, 5, 1, 0, 2, 150)
	cache.RecordEvaluation(false, 20, 15, 0, 1, 3, 300)

	summary := cache.Summary()
	if summary.CacheHits != 1 || summary.CacheMisses != 1 {
		t.Errorf("unexpected hits/misses: %d / %d", summary.CacheHits, summary.CacheMisses)
	}
	if summary.NodesExpanded != 30 || summary.NodesPruned != 20 {
		t.Errorf("unexpected search counts: %d / %d", summary.NodesExpanded, summary.NodesPruned)
	}
	if summary.BeamWidthUsed != 3 {
		t.Errorf("expected max beam width 3, got %d", summary.BeamWidthUsed)
	}

	// Verify template bounding (capacity ceiling is 32)
	for i := 0; i < 50; i++ {
		cache.PutTemplate(fmt.Sprintf("key-%d", i), CachedTemplateSummary{
			TemplateID:    fmt.Sprintf("t-%d", i),
			Domain:        "Coding",
			SubgoalsCount: 3,
			Timestamp:     time.Now(),
		})
	}

	cache.mu.Lock()
	mapLen := len(cache.Templates)
	cache.mu.Unlock()

	if mapLen > maxTemplateEntries {
		t.Fatalf("cache exceeded max capacity ceiling %d, got %d", maxTemplateEntries, mapLen)
	}

	// Verify GetTemplate
	cache.PutTemplate("special-key", CachedTemplateSummary{TemplateID: "special-1"})
	val, found := cache.GetTemplate("special-key")
	if !found || val.TemplateID != "special-1" {
		t.Errorf("failed to retrieve template: %+v found=%v", val, found)
	}

	// Verify Close / destruction
	cache.Close()
	_, foundAfterClose := cache.GetTemplate("special-key")
	if foundAfterClose {
		t.Error("expected cache to be destroyed after Close(), but template was found")
	}
	cache.PutTemplate("after-close", CachedTemplateSummary{TemplateID: "should-not-store"})
	if cache.Templates != nil {
		t.Error("expected Templates map to remain nil after Close()")
	}
}

func TestReflexivePlanningCache_ConcurrentAccess(t *testing.T) {
	cache := NewReflexivePlanningCache("ep-concurrent", "2.0")
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hit := idx%2 == 0
			cache.RecordEvaluation(hit, 5, 2, 0, 0, 2, uint32(100+idx))
			cache.PutTemplate(fmt.Sprintf("k-%d", idx%40), CachedTemplateSummary{
				TemplateID: fmt.Sprintf("t-%d", idx),
			})
			_ = cache.Summary()
		}(i)
	}
	wg.Wait()

	summary := cache.Summary()
	if summary.NodesExpanded != 500 {
		t.Errorf("expected 500 total nodes expanded under concurrency, got %d", summary.NodesExpanded)
	}
}
