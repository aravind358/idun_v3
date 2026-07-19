// Package planning provides foundational data structures and domain logic for IDUN V3's
// planning capabilities. This file implements the OPEN priority queue and bounded beam width
// pruning mechanism for the Beam A* strategic search engine.
package planning

import (
	"container/heap"
	"sort"
	"sync"
)

// nodeHeap implements container/heap.Interface ordered ascending by EvaluationScore f(n),
// where lower score represents higher priority (min-heap for path cost + heuristic).
type nodeHeap []*SearchNode

func (h nodeHeap) Len() int { return len(h) }
func (h nodeHeap) Less(i, j int) bool {
	return nodeLess(h[i], h[j])
}
func (h nodeHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// nodeLess compares two SearchNodes, ordering primarily by EvaluationScore f(n) ascending.
// Ties are broken deterministically by EpistemicConfidence descending, then SimulatedClock ascending, then NodeID ascending.
func nodeLess(a, b *SearchNode) bool {
	if a == nil || b == nil {
		return a != nil
	}
	if a.CostProfile.EvaluationScore != b.CostProfile.EvaluationScore {
		return a.CostProfile.EvaluationScore < b.CostProfile.EvaluationScore
	}
	if a.State != nil && b.State != nil {
		if a.State.EpistemicConfidence != b.State.EpistemicConfidence {
			return a.State.EpistemicConfidence > b.State.EpistemicConfidence
		}
		if a.State.SimulatedClock != b.State.SimulatedClock {
			return a.State.SimulatedClock < b.State.SimulatedClock
		}
	}
	return a.NodeID < b.NodeID
}

func (h *nodeHeap) Push(x interface{}) {
	node := x.(*SearchNode)
	*h = append(*h, node)
}

func (h *nodeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	node := old[n-1]
	*h = old[0 : n-1]
	return node
}

// OpenQueue implements a thread-safe priority queue ordered by EvaluationScore f(n)
// with built-in support for bounded beam-width pruning.
type OpenQueue struct {
	mu   sync.Mutex
	heap nodeHeap
}

// NewOpenQueue constructs an empty OpenQueue ready for node insertion.
func NewOpenQueue() *OpenQueue {
	q := &OpenQueue{
		heap: make(nodeHeap, 0),
	}
	heap.Init(&q.heap)
	return q
}

// Push inserts a SearchNode into the priority queue and re-establishes the min-heap invariant.
func (q *OpenQueue) Push(node *SearchNode) {
	if node == nil {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(&q.heap, node)
}

// Pop removes and returns the SearchNode with the lowest EvaluationScore f(n).
// Returns nil if the queue is empty.
func (q *OpenQueue) Pop() *SearchNode {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.heap) == 0 {
		return nil
	}
	return heap.Pop(&q.heap).(*SearchNode)
}

// Peek returns the SearchNode with the lowest EvaluationScore f(n) without removing it.
// Returns nil if the queue is empty.
func (q *OpenQueue) Peek() *SearchNode {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.heap) == 0 {
		return nil
	}
	return q.heap[0]
}

// Len returns the number of nodes currently awaiting expansion in the queue.
func (q *OpenQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.heap)
}

// PruneToBeam enforces the configured beam width K by sorting all nodes currently in the queue
// by EvaluationScore f(n) ascending, retaining the best K nodes, and marking any discarded nodes
// with NodeStatusPrunedBudget. Returns the list of pruned nodes and the count.
func (q *OpenQueue) PruneToBeam(beamWidth int) ([]*SearchNode, int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if beamWidth <= 0 || len(q.heap) <= beamWidth {
		return nil, 0
	}

	// Sort nodes directly by priority ascending with deterministic tie-breaking
	sort.SliceStable(q.heap, func(i, j int) bool {
		return nodeLess(q.heap[i], q.heap[j])
	})

	retained := q.heap[:beamWidth]
	discarded := q.heap[beamWidth:]

	prunedNodes := make([]*SearchNode, len(discarded))
	for i, node := range discarded {
		node.Status = NodeStatusPrunedBudget
		prunedNodes[i] = node
	}

	q.heap = retained
	heap.Init(&q.heap)

	return prunedNodes, len(prunedNodes)
}

// Snapshot returns an ordered slice copy of all nodes currently in the queue sorted by priority.
func (q *OpenQueue) Snapshot() []*SearchNode {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]*SearchNode, len(q.heap))
	copy(out, q.heap)
	sort.SliceStable(out, func(i, j int) bool {
		return nodeLess(out[i], out[j])
	})
	return out
}
