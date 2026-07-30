package v3

import (
	"errors"
	"fmt"
	"idun/core/foundation"
)

var ErrValidation = errors.New("planning validation error")

const SpecVersion = "3.0"

// CapabilityID represents an abstract capability identifier (e.g. "urn:capability:device.lights.on").
type CapabilityID string

// PlanNode represents a single task in the execution plan graph.
type PlanNode struct {
	nodeID      string
	capability  CapabilityID
	boundParams map[string]any
}

func NewPlanNode(nodeID string, capID CapabilityID, params map[string]any) PlanNode {
	cp := make(map[string]any)
	for k, v := range params {
		cp[k] = v
	}
	return PlanNode{nodeID: nodeID, capability: capID, boundParams: cp}
}

func (p PlanNode) NodeID() string { return p.nodeID }
func (p PlanNode) Capability() CapabilityID { return p.capability }
func (p PlanNode) BoundParams() map[string]any {
	cp := make(map[string]any)
	for k, v := range p.boundParams {
		cp[k] = v
	}
	return cp
}

// Dependency represents a directed edge in the execution DAG.
type Dependency struct {
	fromNodeID string // The node that must complete first.
	toNodeID   string // The node that waits.
}

func NewDependency(from, to string) Dependency {
	return Dependency{fromNodeID: from, toNodeID: to}
}

func (d Dependency) FromNodeID() string { return d.fromNodeID }
func (d Dependency) ToNodeID() string   { return d.toNodeID }

// ExecutionPlan represents the definitive output of the Planning layer.
type ExecutionPlan struct {
	// Lineage
	specVersion      string
	artifactID       foundation.ArtifactID
	parentArtifactID foundation.ParentArtifactID
	envelopeID       foundation.EnvelopeID
	timestamp        foundation.Timestamp

	// DAG
	nodes []PlanNode
	edges []Dependency
}

// CognitiveArtifact implementation
func (p *ExecutionPlan) IsImmutable() bool                                 { return true }
func (p *ExecutionPlan) ArtifactID() foundation.ArtifactID                 { return p.artifactID }
func (p *ExecutionPlan) ParentArtifactID() foundation.ParentArtifactID     { return p.parentArtifactID }
func (p *ExecutionPlan) EnvelopeID() foundation.EnvelopeID                 { return p.envelopeID }
func (p *ExecutionPlan) Timestamp() foundation.Timestamp                   { return p.timestamp }
func (p *ExecutionPlan) Version() foundation.Version                       { return foundation.Version(p.specVersion) }

// Field Getters
func (p *ExecutionPlan) Nodes() []PlanNode {
	cp := make([]PlanNode, len(p.nodes))
	copy(cp, p.nodes)
	return cp
}

func (p *ExecutionPlan) Edges() []Dependency {
	cp := make([]Dependency, len(p.edges))
	copy(cp, p.edges)
	return cp
}

// Validate enforces structural rules: lineage, unique nodes, valid edges, and acyclic DAG.
func (p *ExecutionPlan) Validate() error {
	if p.specVersion != SpecVersion {
		return fmt.Errorf("%w: expected %q, got %q", ErrValidation, SpecVersion, p.specVersion)
	}
	if p.artifactID == "" {
		return fmt.Errorf("%w: ArtifactID cannot be empty", ErrValidation)
	}
	if p.parentArtifactID == "" {
		return fmt.Errorf("%w: ParentArtifactID cannot be empty", ErrValidation)
	}
	if p.envelopeID == "" {
		return fmt.Errorf("%w: EnvelopeID cannot be empty", ErrValidation)
	}

	// Unique Node IDs
	nodeSet := make(map[string]bool)
	for _, n := range p.nodes {
		if n.nodeID == "" {
			return fmt.Errorf("%w: empty node ID found", ErrValidation)
		}
		if nodeSet[n.nodeID] {
			return fmt.Errorf("%w: duplicate node ID %q", ErrValidation, n.nodeID)
		}
		nodeSet[n.nodeID] = true
	}

	// Valid edges
	adjList := make(map[string][]string)
	for _, e := range p.edges {
		if !nodeSet[e.fromNodeID] {
			return fmt.Errorf("%w: edge references unknown from-node %q", ErrValidation, e.fromNodeID)
		}
		if !nodeSet[e.toNodeID] {
			return fmt.Errorf("%w: edge references unknown to-node %q", ErrValidation, e.toNodeID)
		}
		adjList[e.fromNodeID] = append(adjList[e.fromNodeID], e.toNodeID)
	}

	// Acyclic DAG check (Kahn's or DFS)
	visited := make(map[string]int) // 0: unvisited, 1: visiting, 2: visited
	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		if visited[node] == 1 {
			return true // back edge -> cycle
		}
		if visited[node] == 2 {
			return false
		}
		visited[node] = 1
		for _, neighbor := range adjList[node] {
			if hasCycle(neighbor) {
				return true
			}
		}
		visited[node] = 2
		return false
	}

	for _, n := range p.nodes {
		if visited[n.nodeID] == 0 {
			if hasCycle(n.nodeID) {
				return fmt.Errorf("%w: graph contains a cycle", ErrValidation)
			}
		}
	}

	return nil
}
