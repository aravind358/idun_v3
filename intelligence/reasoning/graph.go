package reasoning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/understanding"
)

var (
	ErrGraphNodeLimitExceeded  = errors.New("reasoning: graph node limit exceeded")
	ErrGraphEdgeLimitExceeded  = errors.New("reasoning: graph edge limit exceeded")
	ErrGraphDepthLimitExceeded = errors.New("reasoning: graph traversal depth limit exceeded")
	ErrGraphNodeNotFound       = errors.New("reasoning: graph node not found")
)

// GraphNode represents an ephemeral node in a session-scoped working graph.
type GraphNode struct {
	ID       string
	NodeType string
	Label    string
}

// GraphEdge represents a directional relationship edge in a session-scoped working graph.
type GraphEdge struct {
	FromNodeID string
	ToNodeID   string
	RelType    string
	Weight     float64
}

// GraphPath represents a bounded multi-hop traversal path across the working graph.
type GraphPath struct {
	Nodes  []*GraphNode
	Edges  []*GraphEdge
	Weight float64
}

// SessionGraph implements an ephemeral, session-scoped working graph for Stage S2.
//
// MANDATORY INVARIANTS:
// 1. Exists ONLY during a single reasoning episode.
// 2. NEVER written to Memory or Storage.
// 3. Enforces hard spatial bounds (maxNodes, maxEdges, maxDepth).
// 4. Must be explicitly cleared/destroyed upon episode completion.
type SessionGraph struct {
	mu       sync.RWMutex
	maxNodes int
	maxEdges int
	maxDepth int

	nodes      map[string]*GraphNode
	edges      []*GraphEdge
	adjForward map[string][]*GraphEdge
}

// NewSessionGraph constructs a bounded, ephemeral session working graph.
func NewSessionGraph(maxNodes, maxEdges, maxDepth int) *SessionGraph {
	if maxNodes <= 0 {
		maxNodes = 500
	}
	if maxEdges <= 0 {
		maxEdges = 2000
	}
	if maxDepth <= 0 {
		maxDepth = 3
	}
	return &SessionGraph{
		maxNodes:   maxNodes,
		maxEdges:   maxEdges,
		maxDepth:   maxDepth,
		nodes:      make(map[string]*GraphNode),
		edges:      make([]*GraphEdge, 0, 64),
		adjForward: make(map[string][]*GraphEdge),
	}
}

// AddNode adds a node to the ephemeral session graph.
// Returns ErrGraphNodeLimitExceeded if maxNodes limit would be breached.
func (g *SessionGraph) AddNode(id, nodeType, label string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[id]; exists {
		return nil
	}
	if len(g.nodes) >= g.maxNodes {
		return ErrGraphNodeLimitExceeded
	}
	g.nodes[id] = &GraphNode{
		ID:       id,
		NodeType: nodeType,
		Label:    label,
	}
	return nil
}

// AddEdge creates a directional relationship edge between two nodes.
// Returns ErrGraphEdgeLimitExceeded if maxEdges limit would be breached.
func (g *SessionGraph) AddEdge(fromID, toID, relType string, weight float64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[fromID]; !exists {
		return fmt.Errorf("%w: fromNode %q", ErrGraphNodeNotFound, fromID)
	}
	if _, exists := g.nodes[toID]; !exists {
		return fmt.Errorf("%w: toNode %q", ErrGraphNodeNotFound, toID)
	}
	if len(g.edges) >= g.maxEdges {
		return ErrGraphEdgeLimitExceeded
	}

	edge := &GraphEdge{
		FromNodeID: fromID,
		ToNodeID:   toID,
		RelType:    relType,
		Weight:     weight,
	}
	g.edges = append(g.edges, edge)
	g.adjForward[fromID] = append(g.adjForward[fromID], edge)
	return nil
}

// ExtractRelationships returns all outgoing edges from a specific node ID.
func (g *SessionGraph) ExtractRelationships(fromID string) ([]*GraphEdge, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[fromID]; !exists {
		return nil, fmt.Errorf("%w: %q", ErrGraphNodeNotFound, fromID)
	}
	edges := g.adjForward[fromID]
	out := make([]*GraphEdge, len(edges))
	copy(out, edges)
	return out, nil
}

// Prune removes edges below minWeight and any orphaned nodes without incident edges.
func (g *SessionGraph) Prune(minWeight float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	prunedEdges := make([]*GraphEdge, 0, len(g.edges))
	adjForward := make(map[string][]*GraphEdge)
	connectedNodes := make(map[string]bool)

	for _, e := range g.edges {
		if e.Weight >= minWeight {
			prunedEdges = append(prunedEdges, e)
			adjForward[e.FromNodeID] = append(adjForward[e.FromNodeID], e)
			connectedNodes[e.FromNodeID] = true
			connectedNodes[e.ToNodeID] = true
		}
	}

	for id := range g.nodes {
		if !connectedNodes[id] {
			delete(g.nodes, id)
		}
	}
	g.edges = prunedEdges
	g.adjForward = adjForward
}

// TraverseBounded performs deterministic depth-bounded multi-hop exploration
// starting from startID up to maxDepth hops.
func (g *SessionGraph) TraverseBounded(startID string, maxDepth int) ([]GraphPath, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	startNode, ok := g.nodes[startID]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrGraphNodeNotFound, startID)
	}
	if maxDepth > g.maxDepth {
		maxDepth = g.maxDepth
	}

	var results []GraphPath
	var dfs func(curr *GraphNode, visited map[string]bool, nodes []*GraphNode, edges []*GraphEdge, weight float64)

	dfs = func(curr *GraphNode, visited map[string]bool, nodes []*GraphNode, edges []*GraphEdge, weight float64) {
		if len(edges) > 0 {
			pathNodes := make([]*GraphNode, len(nodes))
			copy(pathNodes, nodes)
			pathEdges := make([]*GraphEdge, len(edges))
			copy(pathEdges, edges)
			results = append(results, GraphPath{
				Nodes:  pathNodes,
				Edges:  pathEdges,
				Weight: weight,
			})
		}
		if len(edges) >= maxDepth {
			return
		}

		for _, e := range g.adjForward[curr.ID] {
			if visited[e.ToNodeID] {
				continue
			}
			nxt, exists := g.nodes[e.ToNodeID]
			if !exists {
				continue
			}
			visited[e.ToNodeID] = true
			dfs(nxt, visited, append(nodes, nxt), append(edges, e), weight*e.Weight)
			visited[e.ToNodeID] = false
		}
	}

	visited := map[string]bool{startID: true}
	dfs(startNode, visited, []*GraphNode{startNode}, nil, 1.0)
	return results, nil
}

// Clear destroys all ephemeral session graph nodes, edges, and index structures immediately.
// This guarantees zero state persistence across reasoning episodes.
func (g *SessionGraph) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.nodes = nil
	g.edges = nil
	g.adjForward = nil
}

// NodeCount returns the current number of nodes in the working graph.
func (g *SessionGraph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the current number of edges in the working graph.
func (g *SessionGraph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.edges)
}

// RelationalGraphSpecialist implements Stage S2 Relational Graph Reasoning.
type RelationalGraphSpecialist struct{}

// NewRelationalGraphSpecialist constructs an initialized RelationalGraphSpecialist.
func NewRelationalGraphSpecialist() *RelationalGraphSpecialist {
	return &RelationalGraphSpecialist{}
}

// ID returns StageS2RelationalGraph.
func (s *RelationalGraphSpecialist) ID() StageIdentifier {
	return StageS2RelationalGraph
}

// Evaluate constructs an ephemeral session graph, populates it from slots and context,
// discovers multi-hop relational paths, and destroys the graph immediately before returning.
func (s *RelationalGraphSpecialist) Evaluate(
	ctx context.Context,
	perceptionEnv communication.Envelope,
	frame *understanding.SemanticFrame,
	memContext []memory.Record,
	spec StrategySpec,
) ([]ReasoningHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	maxNodes := spec.MaxGraphNodes
	maxEdges := spec.MaxGraphEdges
	maxDepth := spec.MaxGraphDepth
	if maxNodes <= 0 {
		maxNodes = 500
	}
	if maxEdges <= 0 {
		maxEdges = 2000
	}
	if maxDepth <= 0 {
		maxDepth = 3
	}

	graph := NewSessionGraph(maxNodes, maxEdges, maxDepth)
	// MANDATORY INVARIANT: Destroy graph immediately upon return
	defer graph.Clear()

	rootID := "root:intent"
	_ = graph.AddNode(rootID, "intent", perceptionEnv.ID)

	if frame != nil && frame.PrimaryHypothesis.Intent != "" {
		_ = graph.AddNode(rootID, "intent", frame.PrimaryHypothesis.Intent)
		for i, slot := range frame.PrimaryHypothesis.Slots {
			slotNodeID := fmt.Sprintf("slot:%d:%s", i, slot.Name)
			_ = graph.AddNode(slotNodeID, "slot", slot.Value)
			_ = graph.AddEdge(rootID, slotNodeID, "has_argument", 0.95)
		}
	}

	for _, rec := range memContext {
		nodeID := fmt.Sprintf("mem:%s", rec.ID)
		label := rec.ID
		if (rec.Type == "semantic_goal" || rec.Type == "goal" || rec.Type == "consequent" || rec.Type == "target_state") && len(rec.Payload) > 0 {
			label = string(rec.Payload)
		}
		_ = graph.AddNode(nodeID, rec.Type, label)
		_ = graph.AddEdge(rootID, nodeID, "context_link", 0.85)
	}

	paths, err := graph.TraverseBounded(rootID, maxDepth)
	if err != nil {
		return nil, err
	}

	hyps := make([]ReasoningHypothesis, 0, len(paths))
	for i, path := range paths {
		if len(path.Nodes) < 2 {
			continue
		}
		endNode := path.Nodes[len(path.Nodes)-1]
		conclusion := fmt.Sprintf("Relational path discovered from %s to %s via %d hops", rootID, endNode.ID, len(path.Edges))

		var proposedGoal *SemanticGoal
		var missingInfoReason string
		for _, node := range path.Nodes {
			if node.NodeType == "semantic_goal" || node.NodeType == "consequent" || node.NodeType == "goal" || node.NodeType == "target_state" {
				var goal SemanticGoal
				if err := json.Unmarshal([]byte(node.Label), &goal); err == nil && goal.Validate() == nil {
					proposedGoal = goal.Clone()
					break
				}
			}
		}
		if proposedGoal == nil {
			missingInfoReason = "Relational working graph path contains only opaque node IDs and relationships without structured GoalKind/Intent/Target/DesiredState; insufficient semantic information to derive ProposedGoal without fabrication"
		}

		premises := []string{fmt.Sprintf("path_hops=%d", len(path.Edges))}
		if missingInfoReason != "" {
			premises = append(premises, fmt.Sprintf("missing_goal_semantics=%s", missingInfoReason))
		}

		hyps = append(hyps, ReasoningHypothesis{
			ID:                  fmt.Sprintf("s2-hyp-%s-%d", perceptionEnv.ID, i),
			Type:                HypothesisRelation,
			Conclusion:          conclusion,
			ProposedGoal:        proposedGoal,
			ReasoningConfidence: path.Weight,
			ContributingStages:  []StageIdentifier{StageS2RelationalGraph},
			SupportingPremises:  premises,
			EvidenceTrace:       "Evaluated via S2 Relational Graph multi-hop traversal",
		})
	}

	return hyps, nil
}
