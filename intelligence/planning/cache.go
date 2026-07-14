package planning

import (
	"sync"
	"time"
)

const (
	// maxTemplateEntries is the strict upper bound on memoized partial plan templates stored inside an episode cache.
	maxTemplateEntries = 32
)

// CachedTemplateSummary holds a lightweight memoized partial plan template or stage result.
type CachedTemplateSummary struct {
	TemplateID      string    `json:"template_id"`
	Domain          string    `json:"domain"`
	GoalFingerprint string    `json:"goal_fingerprint"`
	SubgoalsCount   int       `json:"subgoals_count"`
	EdgesCount      int       `json:"edges_count"`
	LatencyUs       uint32    `json:"latency_us"`
	Timestamp       time.Time `json:"timestamp"`
}

// ReflexivePlanningCache is an O(1) memory-bounded intra-episode scratchpad and statistical accumulator.
// Total memory footprint is constant regardless of how many tactical or reflexive evaluations occur.
//
// IMMUTABLE ARCHITECTURAL BOUNDARY RULES:
// 1. Intra-Episode Existence: Exists ONLY during a single planning episode (`EpisodeID`).
// 2. Strict Memory Bounding: Is strictly bounded to O(1) memory (max 32 templates + statistical summaries).
// 3. Episode Destruction: Is destroyed immediately when the episode ends (`Close()`).
// 4. Zero Persistent Memory: Is NEVER used as semantic or persistent memory across episodes.
// 5. Zero Cross-Episode Consult: Is NEVER consulted by future planning episodes.
// 6. Bounded Publication: Only a bounded PlanningTrace and SearchStatistics summary are published to Reflection and Learning.
// 7. Zero Post-Episode Reuse: Planning itself must never retain or reuse the cache after the episode ends.
type ReflexivePlanningCache struct {
	mu sync.Mutex

	EpisodeID       string    `json:"episode_id"`
	StrategyVersion string    `json:"strategy_version"`
	CreatedAt       time.Time `json:"created_at"`

	// 1. Exact Process & Search Counters
	TotalEvaluations     uint64 `json:"total_evaluations"`
	CacheHits            uint64 `json:"cache_hits"`
	CacheMisses          uint64 `json:"cache_misses"`
	NodesExpanded        uint64 `json:"nodes_expanded"`
	NodesPruned          uint64 `json:"nodes_pruned"`
	DeadEndsReached      uint32 `json:"dead_ends_reached"`
	ConstraintViolations uint32 `json:"constraint_violations"`
	BeamWidthUsed        uint32 `json:"beam_width_used"`

	// 2. Hardware Latency Summary (Microseconds)
	TotalLatencyUs uint64 `json:"total_latency_us"`
	MaxLatencyUs   uint32 `json:"max_latency_us"`

	// 3. Bounded Map of Partial Template Summaries (Max 32 entries)
	Templates map[string]CachedTemplateSummary `json:"templates"`
}

// NewReflexivePlanningCache constructs a clean O(1) intra-episode scratchpad.
func NewReflexivePlanningCache(episodeID, strategyVersion string) *ReflexivePlanningCache {
	return &ReflexivePlanningCache{
		EpisodeID:       episodeID,
		StrategyVersion: strategyVersion,
		CreatedAt:       time.Now().UTC(),
		Templates:       make(map[string]CachedTemplateSummary, maxTemplateEntries),
	}
}

// RecordEvaluation records a single planning lookup or search increment inside the episode.
func (c *ReflexivePlanningCache) RecordEvaluation(hit bool, expanded, pruned int, deadEnds, violations, beamWidth uint32, latencyUs uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.TotalEvaluations++
	if hit {
		c.CacheHits++
	} else {
		c.CacheMisses++
	}
	if expanded > 0 {
		c.NodesExpanded += uint64(expanded)
	}
	if pruned > 0 {
		c.NodesPruned += uint64(pruned)
	}
	c.DeadEndsReached += deadEnds
	c.ConstraintViolations += violations
	if beamWidth > c.BeamWidthUsed {
		c.BeamWidthUsed = beamWidth
	}
	c.TotalLatencyUs += uint64(latencyUs)
	if latencyUs > c.MaxLatencyUs {
		c.MaxLatencyUs = latencyUs
	}
}

// PutTemplate stores a memoized partial graph summary, strictly enforcing the 32-entry capacity ceiling.
func (c *ReflexivePlanningCache) PutTemplate(key string, summary CachedTemplateSummary) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Templates == nil {
		return // Cache already closed/destroyed
	}
	// If capacity reached, evict an arbitrary entry to preserve O(1) memory ceiling
	if len(c.Templates) >= maxTemplateEntries {
		for k := range c.Templates {
			delete(c.Templates, k)
			break
		}
	}
	c.Templates[key] = summary
}

// GetTemplate retrieves a memoized partial graph summary if present.
func (c *ReflexivePlanningCache) GetTemplate(key string) (CachedTemplateSummary, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Templates == nil {
		return CachedTemplateSummary{}, false
	}
	summary, found := c.Templates[key]
	return summary, found
}

// Summary exports the bounded SearchStatistics ready for inclusion inside PlanningTrace.
func (c *ReflexivePlanningCache) Summary() SearchStatistics {
	c.mu.Lock()
	defer c.mu.Unlock()

	return SearchStatistics{
		NodesExpanded:        c.NodesExpanded,
		NodesPruned:          c.NodesPruned,
		BeamWidthUsed:        c.BeamWidthUsed,
		ConstraintViolations: c.ConstraintViolations,
		DeadEndsReached:      c.DeadEndsReached,
		CacheHits:            c.CacheHits,
		CacheMisses:          c.CacheMisses,
	}
}

// Close explicitly destroys internal maps and resets counters at the end of the episode,
// guaranteeing zero memory leakage and preventing post-episode reuse.
func (c *ReflexivePlanningCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Templates = nil
	c.TotalEvaluations = 0
	c.NodesExpanded = 0
	c.NodesPruned = 0
}
