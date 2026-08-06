package v3

import "idun/core/foundation"

// Builder implements the Builder pattern for ExecutionPlan to enforce immutability
// and require valid initialization (including acyclic DAG rules) before publishing.
type Builder struct {
	obj *ExecutionPlan
}

func NewBuilder() *Builder {
	return &Builder{
		obj: &ExecutionPlan{
			specVersion: SpecVersion,
		},
	}
}

func (b *Builder) ArtifactID(id foundation.ArtifactID) *Builder {
	b.obj.artifactID = id
	return b
}
func (b *Builder) ParentArtifactID(id foundation.ParentArtifactID) *Builder {
	b.obj.parentArtifactID = id
	return b
}
func (b *Builder) EnvelopeID(id foundation.EnvelopeID) *Builder {
	b.obj.envelopeID = id
	return b
}
func (b *Builder) Timestamp(t foundation.Timestamp) *Builder {
	b.obj.timestamp = t
	return b
}
func (b *Builder) PlanIntent(intent string) *Builder {
	b.obj.planIntent = intent
	return b
}
func (b *Builder) Nodes(nodes []PlanNode) *Builder {
	b.obj.nodes = make([]PlanNode, len(nodes))
	copy(b.obj.nodes, nodes)
	return b
}
func (b *Builder) Edges(edges []Dependency) *Builder {
	b.obj.edges = make([]Dependency, len(edges))
	copy(b.obj.edges, edges)
	return b
}

func (b *Builder) Build() (*ExecutionPlan, error) {
	if err := b.obj.Validate(); err != nil {
		return nil, err
	}
	return b.obj, nil
}
