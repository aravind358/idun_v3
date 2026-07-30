package v3

import "idun/core/foundation"

// Builder implements the Builder pattern for DecisionRecord to enforce immutability
// and require valid initialization before publishing.
type Builder struct {
	obj *DecisionRecord
}

func NewBuilder() *Builder {
	return &Builder{
		obj: &DecisionRecord{
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
func (b *Builder) Resolution(r ResolutionStatus) *Builder {
	b.obj.resolution = r
	return b
}
func (b *Builder) Reason(r string) *Builder {
	b.obj.reason = r
	return b
}
func (b *Builder) EvaluationFlags(safety, policy, auth, budget bool) *Builder {
	b.obj.safetyPassed = safety
	b.obj.policyPassed = policy
	b.obj.permissionsPassed = auth
	b.obj.budgetPassed = budget
	return b
}
func (b *Builder) Findings(findings []DecisionFinding) *Builder {
	b.obj.findings = make([]DecisionFinding, len(findings))
	copy(b.obj.findings, findings)
	return b
}

func (b *Builder) Build() (*DecisionRecord, error) {
	if err := b.obj.Validate(); err != nil {
		return nil, err
	}
	return b.obj, nil
}
