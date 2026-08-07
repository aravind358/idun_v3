package understanding

import (
	"context"
	"sort"
	"sync"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
)

// SpeculativeEvaluator coordinates bounded parallel evaluation across deterministic
// and probabilistic specialists, integrates epistemic calibration, applies slot-aware
// hypothesis merging, and executes bounded beam pruning (K <= 3, delta <= 0.15).
type SpeculativeEvaluator struct {
	deltaThreshold float64
}

// NewSpeculativeEvaluator constructs a new speculative evaluator.
func NewSpeculativeEvaluator(deltaThreshold float64) *SpeculativeEvaluator {
	if deltaThreshold <= 0 {
		deltaThreshold = DefaultAmbiguityDelta
	}
	return &SpeculativeEvaluator{
		deltaThreshold: deltaThreshold,
	}
}

// EvaluateParallel executes GrammarSpecialist and NeuralSpecialist concurrently,
// calibrates candidate confidence scores, performs slot-aware merging for matching
// intents, and selects primary + ambiguity beam.
func (e *SpeculativeEvaluator) EvaluateParallel(
	ctx context.Context,
	norm NormalizedText,
	grammar GrammarSpecialist,
	neural NeuralSpecialist,
	calibrator calibration.CalibrationService,
) (Hypothesis, []Hypothesis, error) {
	var (
		mu         sync.Mutex
		candidates []Hypothesis
		wg         sync.WaitGroup
	)

	// 1. Parallel execution: Reflexive Grammar Specialist (deterministic)
	if grammar != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if hyp, matched := grammar.Evaluate(norm); matched {
				mu.Lock()
				candidates = append(candidates, hyp)
				mu.Unlock()
			}
		}()
	}

	// 2. Parallel execution: Neural Specialist (probabilistic)
	if neural != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if hyps, err := neural.Evaluate(norm); err == nil {
				mu.Lock()
				candidates = append(candidates, hyps...)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(candidates) == 0 {
		return Hypothesis{}, nil, nil
	}

	// 3. Calibration integration
	for i := range candidates {
		if calibrator != nil {
			calibrated := calibrator.CalibrateConfidence(
				string(candidates[i].SourceLayer),
				communication.TopicUserIntent,
				candidates[i].CalibratedConfidence,
			)
			candidates[i].CalibratedConfidence = calibrated
		}
		if candidates[i].CalibratedConfidence < 0.0 {
			candidates[i].CalibratedConfidence = 0.0
		} else if candidates[i].CalibratedConfidence > 1.0 {
			candidates[i].CalibratedConfidence = 1.0
		}
	}

	// 4. Slot-Aware Hypothesis Merging by Intent
	mergedList := MergeHypothesesByIntent(candidates)

	// 5. Sort descending by calibrated confidence
	sort.SliceStable(mergedList, func(i, j int) bool {
		return mergedList[i].CalibratedConfidence > mergedList[j].CalibratedConfidence
	})

	primary := mergedList[0]
	primary.DeltaFromPrimary = 0.0

	var ambiguitySet []Hypothesis
	for i := 1; i < len(mergedList); i++ {
		cand := mergedList[i]
		delta := primary.CalibratedConfidence - cand.CalibratedConfidence
		cand.DeltaFromPrimary = delta

		// Enforce bounded beam: delta <= threshold AND runner-up count < MaxBeamWidth - 1
		if delta <= e.deltaThreshold && len(ambiguitySet) < MaxBeamWidth-1 {
			ambiguitySet = append(ambiguitySet, cand)
		}
	}

	return primary, ambiguitySet, nil
}

// MergeHypothesesByIntent groups candidate hypotheses by Intent and performs a
// slot-aware merge. Complementary slots across specialists are combined. When two
// specialists extract a slot with the same Name, the slot with higher Confidence wins.
func MergeHypothesesByIntent(candidates []Hypothesis) []Hypothesis {
	grouped := make(map[string][]Hypothesis)
	for _, c := range candidates {
		grouped[c.Intent] = append(grouped[c.Intent], c)
	}

	merged := make([]Hypothesis, 0, len(grouped))
	for intent, group := range grouped {
		if len(group) == 1 {
			merged = append(merged, group[0])
			continue
		}

		// Sort group descending by hypothesis CalibratedConfidence
		sort.SliceStable(group, func(i, j int) bool {
			return group[i].CalibratedConfidence > group[j].CalibratedConfidence
		})

		base := group[0]
		slotMap := make(map[string]Slot)

		// Populate slots starting from highest-confidence hypothesis downwards
		for _, h := range group {
			for _, slot := range h.Slots {
				existing, found := slotMap[slot.Name]
				if !found || slot.Confidence > existing.Confidence {
					slotMap[slot.Name] = slot
				}
			}
		}

		// Convert slot map to sorted slice for deterministic ordering
		slotNames := make([]string, 0, len(slotMap))
		for name := range slotMap {
			slotNames = append(slotNames, name)
		}
		sort.Strings(slotNames)

		mergedSlots := make([]Slot, 0, len(slotNames))
		for _, name := range slotNames {
			mergedSlots = append(mergedSlots, slotMap[name])
		}

		base.Intent = intent
		base.Slots = mergedSlots
		merged = append(merged, base)
	}

	return merged
}
