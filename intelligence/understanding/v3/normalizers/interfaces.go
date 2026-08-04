package normalizers

import (
	"idun/intelligence/understanding/v3"
)

// Runner implements v3.NormalizerRunner
type Runner interface {
	Run(b *v3.Builder)
}

type deterministicNormalizers struct {
	temporal TemporalNormalizer
}

func NewDeterministicNormalizers(temporal TemporalNormalizer) Runner {
	return &deterministicNormalizers{
		temporal: temporal,
	}
}

func (n *deterministicNormalizers) Run(b *v3.Builder) {
	anchors := b.GetTemporalAnchors()
	if len(anchors) > 0 {
		normalized := n.temporal.Normalize(anchors)
		b.TemporalAnchors(normalized)
	}
}
