package decision

import (
	"testing"
)

func TestCalibrateConfidence(t *testing.T) {
	// 1. High coverage, high top-margin -> minimal adjustment
	conf, flags := CalibrateConfidence(0.90, 0.15, 1.0, 0.50)
	if len(flags) != 0 {
		t.Errorf("expected 0 flags for clean calibration, got %v", flags)
	}
	if conf != 0.90 {
		t.Errorf("expected 0.90 confidence, got %f", conf)
	}

	// 2. Low coverage and tight ambiguity margin -> penalize and flag
	confLow, flagsLow := CalibrateConfidence(0.90, 0.02, 0.80, 0.50)
	if len(flagsLow) != 2 {
		t.Errorf("expected 2 uncertainty flags, got %v", flagsLow)
	}
	if confLow >= 0.90 {
		t.Errorf("expected confidence to be deflated below 0.90, got %f", confLow)
	}
}

func TestIdentifyInformationGaps(t *testing.T) {
	candidates := []Candidate{
		{
			ID: "cand-partial",
			Attributes: map[string]float64{
				"utility": 0.8,
			},
		},
	}

	expected := map[string]float64{
		"utility": 1.0,
		"safety":  1.5,
	}

	gaps := IdentifyInformationGaps(candidates, expected)
	if len(gaps) != 1 {
		t.Fatalf("expected 1 information gap for missing safety attribute, got %d", len(gaps))
	}
	if gaps[0].MissingAttribute != "safety" {
		t.Errorf("expected missing attribute 'safety', got %s", gaps[0].MissingAttribute)
	}
}
