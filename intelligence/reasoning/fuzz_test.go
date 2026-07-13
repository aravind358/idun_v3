package reasoning

import (
	"testing"
)

func FuzzReasoningResultValidation(f *testing.F) {
	f.Add("env-1", "frame-1", "strategy-1", 0.85, 0.2)
	f.Add("", "", "", -1.5, 5.0)
	f.Add("🧨 emoji env", "👾 frame", "🎲 strategy", 1.0, 0.0)

	f.Fuzz(func(t *testing.T, envID, frameID, strat string, conf, ambConf float64) {
		res := ReasoningResult{
			SchemaVersion: SchemaVersion,
			EnvelopeID:    envID,
			SourceFrameID: frameID,
			StrategyUsed:  StrategyIdentifier(strat),
			PrimaryHypothesis: ReasoningHypothesis{
				ID:                   "p1",
				ReasoningConfidence:  conf,
				CalibratedConfidence: conf,
			},
			AmbiguitySet: []ReasoningHypothesis{
				{
					ID:                   "b1",
					ReasoningConfidence:  ambConf,
					CalibratedConfidence: ambConf,
				},
			},
		}

		// Validate must never panic regardless of input
		_ = res.Validate()
	})
}

func FuzzBeamSelection(f *testing.F) {
	f.Add(0.9, 0.8, 0.3, 3, 0.25)
	f.Add(-10.0, 100.0, 0.0, -5, -1.0)
	f.Add(0.5, 0.5, 0.5, 10, 2.0)

	specialist := NewBeamSelectionSpecialist()

	f.Fuzz(func(t *testing.T, c1, c2, c3 float64, maxWidth int, threshold float64) {
		hyps := []ReasoningHypothesis{
			{ID: "h1", ReasoningConfidence: c1},
			{ID: "h2", ReasoningConfidence: c2},
			{ID: "h3", ReasoningConfidence: c3},
		}
		// SelectBeam must never panic
		_, _, _ = specialist.SelectBeam(hyps, maxWidth, threshold)
	})
}
