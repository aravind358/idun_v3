package decision

import (
	"errors"
	"testing"
)

func TestCandidateSet_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cs        CandidateSet
		wantErr   bool
		errTarget error
	}{
		{
			name:      "Empty candidate set fails validation",
			cs:        CandidateSet{EpisodeID: "ep-1", Candidates: nil},
			wantErr:   true,
			errTarget: ErrEmptyCandidateSet,
		},
		{
			name: "Candidate set overflow (>16) fails validation",
			cs: CandidateSet{
				EpisodeID:  "ep-2",
				Candidates: make([]Candidate, 17),
			},
			wantErr:   true,
			errTarget: ErrCandidateSetOverflow,
		},
		{
			name: "Valid candidate set (1 candidate)",
			cs: CandidateSet{
				EpisodeID: "ep-3",
				Candidates: []Candidate{
					{ID: "cand-1", Description: "Primary path"},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid candidate set (16 candidates)",
			cs: CandidateSet{
				EpisodeID:  "ep-4",
				Candidates: make([]Candidate, 16),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cs.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errTarget != nil && !errors.Is(err, tt.errTarget) {
				t.Errorf("Validate() error = %v, expected target %v", err, tt.errTarget)
			}
		})
	}
}

func TestDecisionStrategySnapshot_Validate(t *testing.T) {
	var nilSnap *DecisionStrategySnapshot
	if err := nilSnap.Validate(); !errors.Is(err, ErrInvalidStrategySnapshot) {
		t.Errorf("nil snapshot expected ErrInvalidStrategySnapshot, got %v", err)
	}

	emptyVer := &DecisionStrategySnapshot{StrategyVersion: ""}
	if err := emptyVer.Validate(); !errors.Is(err, ErrInvalidStrategySnapshot) {
		t.Errorf("empty version expected ErrInvalidStrategySnapshot, got %v", err)
	}

	valid := &DecisionStrategySnapshot{
		StrategyVersion:           "v1.0",
		EscalationConfidenceFloor: 0.60,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid snapshot returned unexpected error: %v", err)
	}
}
