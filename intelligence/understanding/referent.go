package understanding

import (
	"strings"
)

// ReferentCandidate represents an active entity snapshot from working memory
// provided via TopicEpisodicContext or Envelope metadata.
type ReferentCandidate struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
	Role string `json:"Role"`
}

// ReferentBinder defines the capability to ground pronouns and entity mentions
// against active working memory referent candidates before semantic parsing.
type ReferentBinder interface {
	BindReferents(norm NormalizedText, candidates []ReferentCandidate) []Slot
}

// DefaultReferentBinder implements deterministic pre-parse entity and pronoun grounding.
type DefaultReferentBinder struct{}

// NewDefaultReferentBinder constructs a new DefaultReferentBinder.
func NewDefaultReferentBinder() *DefaultReferentBinder {
	return &DefaultReferentBinder{}
}

// BindReferents inspects normalized tokens and binds pronouns/names to candidate entities.
func (b *DefaultReferentBinder) BindReferents(norm NormalizedText, candidates []ReferentCandidate) []Slot {
	if len(candidates) == 0 {
		return nil
	}

	var slots []Slot
	boundIDs := make(map[string]struct{})

	// 1. Explicit name matching
	for _, cand := range candidates {
		nameLower := strings.ToLower(strings.TrimSpace(cand.Name))
		if nameLower != "" && strings.Contains(norm.Cleaned, nameLower) {
			if _, exists := boundIDs[cand.ID]; !exists {
				slots = append(slots, Slot{
					Name:        "entity_reference",
					Value:       cand.Name,
					GroundingID: cand.ID,
					Confidence:  0.98,
				})
				boundIDs[cand.ID] = struct{}{}
			}
		}
	}

	// 2. Pronoun / deictic anchor binding (binds to most recent candidate if not already bound)
	if len(norm.PronounsFound) > 0 && len(boundIDs) == 0 {
		primaryCand := candidates[0]
		for _, pronoun := range norm.PronounsFound {
			slots = append(slots, Slot{
				Name:        "pronominal_referent",
				Value:       pronoun,
				GroundingID: primaryCand.ID,
				Confidence:  0.92,
			})
			boundIDs[primaryCand.ID] = struct{}{}
			break
		}
	}

	return slots
}

// Ensure DefaultReferentBinder implements ReferentBinder.
var _ ReferentBinder = (*DefaultReferentBinder)(nil)
