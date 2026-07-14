package decision

import (
	"math"
	"testing"
)

func FuzzCalibrateConfidence(f *testing.F) {
	f.Add(0.85, 0.10, 0.90, 0.50)
	f.Add(1.50, -0.20, 0.0, 0.0)
	f.Add(-0.10, 2.00, 1.5, 1.0)

	f.Fuzz(func(t *testing.T, rawConf, topMargin, attrCoverage, riskTol float64) {
		calibrated, _ := CalibrateConfidence(rawConf, topMargin, attrCoverage, riskTol)
		if math.IsNaN(calibrated) || math.IsInf(calibrated, 0) {
			t.Errorf("CalibrateConfidence produced non-finite result: %f", calibrated)
		}
		if calibrated < 0.01 || calibrated > 0.9999 {
			t.Errorf("CalibrateConfidence out of bounds [0.01, 0.9999]: %f", calibrated)
		}
	})
}

func FuzzCandidateSetValidate(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(16)
	f.Add(17)

	f.Fuzz(func(t *testing.T, numCands int) {
		if numCands < 0 || numCands > 100 {
			return
		}
		cands := make([]Candidate, numCands)
		cs := CandidateSet{
			EpisodeID:  "ep-fuzz",
			Candidates: cands,
		}
		err := cs.Validate()
		if numCands >= 1 && numCands <= 16 {
			if err != nil {
				t.Errorf("expected valid CandidateSet for count %d, got err: %v", numCands, err)
			}
		} else {
			if err == nil {
				t.Errorf("expected error for CandidateSet count %d, got nil", numCands)
			}
		}
	})
}
