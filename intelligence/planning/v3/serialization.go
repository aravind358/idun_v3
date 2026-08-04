package v3

import (
	"encoding/json"
	"fmt"
	"idun/core/foundation"
	"time"
)

type jsonPlanNode struct {
	NodeID      string         `json:"NodeID"`
	Capability  string         `json:"Capability"`
	BoundParams map[string]any `json:"BoundParams"`
	UserImpact  string         `json:"UserImpact"`
}

type jsonDependency struct {
	FromNodeID string `json:"FromNodeID"`
	ToNodeID   string `json:"ToNodeID"`
}

type jsonExecutionPlan struct {
	SpecVersion      string           `json:"SpecVersion"`
	ArtifactID       string           `json:"ArtifactID"`
	ParentArtifactID string           `json:"ParentArtifactID"`
	EnvelopeID       string           `json:"EnvelopeID"`
	Timestamp        time.Time        `json:"Timestamp"`
	PlanIntent       string           `json:"PlanIntent"`
	ExecutionClass   string           `json:"ExecutionClass"`
	Nodes            []jsonPlanNode   `json:"Nodes"`
	Edges            []jsonDependency `json:"Edges"`
}

func (p *ExecutionPlan) MarshalJSON() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("cannot marshal invalid ExecutionPlan: %w", err)
	}

	j := jsonExecutionPlan{
		SpecVersion:      p.specVersion,
		ArtifactID:       string(p.artifactID),
		ParentArtifactID: string(p.parentArtifactID),
		EnvelopeID:       string(p.envelopeID),
		Timestamp:        time.Time(p.timestamp),
		PlanIntent:       p.planIntent,
		ExecutionClass:   p.executionClass,
	}

	for _, n := range p.nodes {
		j.Nodes = append(j.Nodes, jsonPlanNode{
			NodeID:      n.nodeID,
			Capability:  string(n.capability),
			BoundParams: n.boundParams, // map handles json internally
			UserImpact:  n.userImpact,
		})
	}
	for _, e := range p.edges {
		j.Edges = append(j.Edges, jsonDependency{
			FromNodeID: e.fromNodeID,
			ToNodeID:   e.toNodeID,
		})
	}

	return json.Marshal(j)
}

func (p *ExecutionPlan) UnmarshalJSON(data []byte) error {
	var j jsonExecutionPlan
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	p.specVersion = j.SpecVersion
	p.artifactID = foundation.ArtifactID(j.ArtifactID)
	p.parentArtifactID = foundation.ParentArtifactID(j.ParentArtifactID)
	p.envelopeID = foundation.EnvelopeID(j.EnvelopeID)
	p.timestamp = foundation.Timestamp(j.Timestamp)
	p.planIntent = j.PlanIntent
	p.executionClass = j.ExecutionClass

	p.nodes = make([]PlanNode, len(j.Nodes))
	for i, n := range j.Nodes {
		p.nodes[i] = NewPlanNode(n.NodeID, CapabilityID(n.Capability), n.BoundParams, n.UserImpact)
	}

	p.edges = make([]Dependency, len(j.Edges))
	for i, e := range j.Edges {
		p.edges[i] = NewDependency(e.FromNodeID, e.ToNodeID)
	}

	return p.Validate()
}
