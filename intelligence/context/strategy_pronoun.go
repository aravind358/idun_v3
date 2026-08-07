package context

import (
	"context"

	underv3 "idun/intelligence/understanding/v3"
	"idun/intelligence/understanding/v3/ontology"
)

// PronounStrategy resolves explicit pronouns (it, that, them) mapped to references.
type PronounStrategy struct{}

func (s *PronounStrategy) Execute(ctx context.Context, orig *underv3.SemanticInterpretation, builder *underv3.Builder, state DialogueStateReader, resolvedEntities map[string]string) (bool, ResolutionStatus) {
	handled := false
	status := StatusContextUnnecessary

	refs := orig.References()
	var newRefs []underv3.Reference

	for _, ref := range refs {
		if ref.Type() == ontology.RefPronoun || ref.Type() == ontology.RefDemonstrative {
			if ref.Resolved() {
				newRefs = append(newRefs, ref)
				continue
			}

			handled = true
			candidates := state.GetRecentCandidates(ref.AnchorHint(), 2)
			
			// Also check our rolling resolved entities if candidates are empty.
			// (e.g. if previous intent in the same batch resolved something we can use)
			// For simplicity in U7.5, if we found it in resolvedEntities we could use it, 
			// but state.GetRecentCandidates is the canonical way if we update the state reader.
			
			if len(candidates) == 1 {
				// We found a single explicit resolution
				newRef := underv3.NewReference(
					ref.Surface(), ref.Type(), ref.AnchorHint(),
					candidates[0], true, ref.Confidence(),
				)
				newRefs = append(newRefs, newRef)
				resolvedEntities[ref.Surface()] = candidates[0] // Track for rolling resolution
				status = StatusResolved
			} else if len(candidates) > 1 {
				// Cross-domain ambiguity detected
				newRefs = append(newRefs, ref) // Keep unresolved
				status = StatusAmbiguous
			} else {
				// Could not resolve the pronoun
				newRefs = append(newRefs, ref) // Keep unresolved
				status = StatusFailed
				// Don't return early here, evaluate all pronouns to completion
			}
		} else {
			newRefs = append(newRefs, ref)
		}
	}

	if handled {
		builder.References(newRefs)
	}

	return handled, status
}
