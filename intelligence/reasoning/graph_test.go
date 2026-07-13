package reasoning

import (
	"context"
	"errors"
	"testing"

	"idun/core/memory"
	"idun/intelligence/communication"
)

func TestSessionGraph_LifecycleAndLimits(t *testing.T) {
	g := NewSessionGraph(3, 2, 2)

	if err := g.AddNode("node1", "concept", "First"); err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}
	if err := g.AddNode("node2", "concept", "Second"); err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}
	if err := g.AddNode("node3", "concept", "Third"); err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}
	// Exceeding limit
	if err := g.AddNode("node4", "concept", "Fourth"); !errors.Is(err, ErrGraphNodeLimitExceeded) {
		t.Fatalf("expected ErrGraphNodeLimitExceeded, got %v", err)
	}

	if err := g.AddEdge("node1", "node2", "relates_to", 0.9); err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	if err := g.AddEdge("node2", "node3", "relates_to", 0.8); err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	// Exceeding edge limit
	if err := g.AddEdge("node1", "node3", "relates_to", 0.7); !errors.Is(err, ErrGraphEdgeLimitExceeded) {
		t.Fatalf("expected ErrGraphEdgeLimitExceeded, got %v", err)
	}

	paths, err := g.TraverseBounded("node1", 2)
	if err != nil {
		t.Fatalf("TraverseBounded failed: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("expected discovered paths")
	}

	// Verify pruning
	g.Prune(0.85)
	if g.EdgeCount() != 1 {
		t.Errorf("expected 1 edge after pruning < 0.85, got %d", g.EdgeCount())
	}

	// MANDATORY INVARIANT: Clear destroys all graph state
	g.Clear()
	if g.NodeCount() != 0 || g.EdgeCount() != 0 {
		t.Errorf("expected 0 nodes and 0 edges after Clear(), got %d nodes / %d edges", g.NodeCount(), g.EdgeCount())
	}
}

func TestRelationalGraphSpecialist_EvaluateAndDestroy(t *testing.T) {
	specialist := NewRelationalGraphSpecialist()
	if specialist.ID() != StageS2RelationalGraph {
		t.Errorf("expected stage ID S2, got %s", specialist.ID())
	}

	env := communication.Envelope{ID: "env-graph-test"}
	memRecords := []memory.Record{
		{ID: "ctx-node-1", Type: "fact"},
		{ID: "ctx-node-2", Type: "belief"},
	}

	spec := StrategySpec{
		StrategyID:    StrategyGraphDeliberative,
		MaxGraphNodes: 100,
		MaxGraphEdges: 200,
		MaxGraphDepth: 2,
	}

	hyps, err := specialist.Evaluate(context.Background(), env, nil, memRecords, spec)
	if err != nil {
		t.Fatalf("expected relational evaluation to succeed, got %v", err)
	}
	if len(hyps) == 0 {
		t.Fatalf("expected multi-hop hypotheses generated")
	}
	for _, hyp := range hyps {
		if err := hyp.Validate(); err != nil {
			t.Fatalf("generated relational hypothesis failed validation: %v", err)
		}
	}
}
