package reasoning

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestReasoningResultValidation_Success(t *testing.T) {
	result := ReasoningResult{
		SchemaVersion: SchemaVersion,
		EnvelopeID:    "env-001",
		SourceFrameID: "frame-001",
		Status:        StatusUnambiguousSolved,
		StrategyUsed:  StrategySymbolicFast,
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                   "hyp-001",
			Type:                 HypothesisInference,
			Conclusion:           "Subject is authorized",
			CalibratedConfidence: 0.95,
		},
		AmbiguitySet: []ReasoningHypothesis{
			{
				ID:                   "hyp-002",
				Type:                 HypothesisInference,
				Conclusion:           "Subject requires elevation",
				CalibratedConfidence: 0.82,
			},
		},
		ContradictionsFlagged: []ContradictionFlag{
			{
				BeliefID:         "bel-101",
				ConflictingClaim: "Subject was revoked",
				Confidence:       0.90,
				DetectedAtStage:  StageS3CSPCheck,
			},
		},
		ProposedBeliefUpdates: []BeliefUpdateProposal{
			{
				Subject:            "user:42",
				Predicate:          "status",
				Object:             "authorized",
				ProposedConfidence: 0.95,
				SourceHypothesisID: "hyp-001",
			},
		},
		CompilationCandidate: &CompilationCandidate{
			EligibleForLearning:  true,
			SourceStage:          StageS8DeliberativeLLM,
			StrategyUsed:         StrategyGraphDeliberative,
			CalibratedConfidence: 0.94,
			Antecedents: []CompilationCondition{
				{Slot: "role", Operator: "EQ", Value: "admin"},
			},
			DerivedConsequent: CompilationConsequent{
				Intent:     "grant_access",
				RuleAction: "FAST_ALLOW",
			},
		},
		StrategyTelemetry: StrategyTelemetry{
			EpisodeID:            "ep-001",
			StrategySelected:     StrategySymbolicFast,
			SpecialistsExecuted:  []StageIdentifier{StageS0ContextAssembly, StageS1SymbolicFast},
			ExecutionDurationMs:  0.8,
			CalibratedConfidence: 0.95,
			ResourceCostTier:     "LOCAL_FAST",
			EscalatedToLLM:       false,
			OutcomeStatus:        StatusUnambiguousSolved,
		},
	}

	if err := result.Validate(); err != nil {
		t.Fatalf("expected valid result, got %v", err)
	}
}

func TestReasoningResultValidation_Failures(t *testing.T) {
	validHyp := ReasoningHypothesis{
		ID:                   "hyp-1",
		Type:                 HypothesisInference,
		Conclusion:           "conclusion",
		CalibratedConfidence: 0.8,
	}

	tests := []struct {
		name    string
		mutate  func(*ReasoningResult)
		wantErr error
	}{
		{
			name: "invalid schema version",
			mutate: func(r *ReasoningResult) {
				r.SchemaVersion = "1.0"
			},
			wantErr: ErrInvalidSchemaVersion,
		},
		{
			name: "missing envelope ID",
			mutate: func(r *ReasoningResult) {
				r.EnvelopeID = ""
			},
			wantErr: ErrMissingEnvelopeID,
		},
		{
			name: "missing source frame ID",
			mutate: func(r *ReasoningResult) {
				r.SourceFrameID = ""
			},
			wantErr: ErrMissingSourceFrameID,
		},
		{
			name: "beam overflow (> MaxBeamWidth)",
			mutate: func(r *ReasoningResult) {
				r.AmbiguitySet = []ReasoningHypothesis{validHyp, validHyp, validHyp} // 1 primary + 3 runners-up = 4 > 3
			},
			wantErr: ErrBeamOverflow,
		},
		{
			name: "invalid confidence out of bounds",
			mutate: func(r *ReasoningResult) {
				r.PrimaryHypothesis.CalibratedConfidence = 1.5
			},
			wantErr: ErrInvalidConfidence,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := ReasoningResult{
				SchemaVersion:     SchemaVersion,
				EnvelopeID:        "env-1",
				SourceFrameID:     "frame-1",
				Status:            StatusUnambiguousSolved,
				PrimaryHypothesis: validHyp,
				StrategyTelemetry: StrategyTelemetry{
					EpisodeID:            "ep-1",
					CalibratedConfidence: 0.8,
				},
			}
			tc.mutate(&r)
			if err := r.Validate(); !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestReasoningResult_CloneAndImmutability(t *testing.T) {
	orig := ReasoningResult{
		SchemaVersion: SchemaVersion,
		EnvelopeID:    "env-orig",
		SourceFrameID: "frame-orig",
		Status:        StatusUnambiguousSolved,
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                 "hyp-orig",
			Type:               HypothesisInference,
			Conclusion:         "Original",
			ContributingStages: []StageIdentifier{StageS1SymbolicFast},
		},
		AmbiguitySet: []ReasoningHypothesis{
			{
				ID:         "hyp-amb",
				Conclusion: "Ambiguous",
			},
		},
		CompilationCandidate: &CompilationCandidate{
			EligibleForLearning: true,
			Antecedents: []CompilationCondition{
				{Slot: "A", Operator: "EQ", Value: "1"},
			},
		},
	}

	clone := orig.Clone()
	clone.EnvelopeID = "env-mutated"
	clone.PrimaryHypothesis.Conclusion = "Mutated"
	clone.PrimaryHypothesis.ContributingStages[0] = StageS8DeliberativeLLM
	clone.CompilationCandidate.Antecedents[0].Value = "999"

	if orig.EnvelopeID == clone.EnvelopeID {
		t.Errorf("EnvelopeID was mutated on original")
	}
	if orig.PrimaryHypothesis.Conclusion == clone.PrimaryHypothesis.Conclusion {
		t.Errorf("PrimaryHypothesis.Conclusion was mutated on original")
	}
	if orig.PrimaryHypothesis.ContributingStages[0] == clone.PrimaryHypothesis.ContributingStages[0] {
		t.Errorf("ContributingStages slice was mutated on original")
	}
	if orig.CompilationCandidate.Antecedents[0].Value == "999" {
		t.Errorf("CompilationCandidate.Antecedents slice was mutated on original")
	}
}

func TestReasoningResult_JSONSerialization(t *testing.T) {
	orig := ReasoningResult{
		SchemaVersion: SchemaVersion,
		EnvelopeID:    "env-json",
		SourceFrameID: "frame-json",
		Status:        StatusUnambiguousSolved,
		StrategyUsed:  StrategySymbolicFast,
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                   "hyp-1",
			Type:                 HypothesisInference,
			Conclusion:           "test conclusion",
			CalibratedConfidence: 0.9,
		},
		StrategyTelemetry: StrategyTelemetry{
			EpisodeID:            "ep-json",
			CalibratedConfidence: 0.9,
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("failed to marshal ReasoningResult: %v", err)
	}

	var decoded ReasoningResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ReasoningResult: %v", err)
	}

	if err := decoded.Validate(); err != nil {
		t.Fatalf("decoded object failed validation: %v", err)
	}
	if decoded.EnvelopeID != orig.EnvelopeID || decoded.PrimaryHypothesis.Conclusion != orig.PrimaryHypothesis.Conclusion {
		t.Errorf("decoded content mismatch: got %+v", decoded)
	}
}

func TestStage1_SemanticGoalCreationAndValidation(t *testing.T) {
	// 1. SemanticGoal creation and validation
	validGoal := &SemanticGoal{
		Kind:          GoalKindCommunicative,
		Intent:        "greet_user",
		Target:        "user",
		DesiredState:  map[string]string{"acknowledged": "true"},
		OperationHint: "greeting",
		Constraints:   map[string]string{"read_only": "true"},
	}
	if err := validGoal.Validate(); err != nil {
		t.Fatalf("expected valid SemanticGoal to pass, got: %v", err)
	}

	// Nil goal is valid/optional
	var nilGoal *SemanticGoal
	if err := nilGoal.Validate(); err != nil {
		t.Fatalf("expected nil SemanticGoal validation to pass, got: %v", err)
	}

	// Missing intent
	invalidGoal := &SemanticGoal{
		Kind:         GoalKindCommunicative,
		Target:       "user",
		DesiredState: map[string]string{"k": "v"},
	}
	if err := invalidGoal.Validate(); err == nil {
		t.Fatalf("expected error for missing intent, got nil")
	}

	// Missing target
	invalidGoal = &SemanticGoal{
		Kind:         GoalKindCommunicative,
		Intent:       "greet_user",
		DesiredState: map[string]string{"k": "v"},
	}
	if err := invalidGoal.Validate(); err == nil {
		t.Fatalf("expected error for missing target, got nil")
	}

	// Missing desired_state map
	invalidGoal = &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
	}
	if err := invalidGoal.Validate(); err == nil {
		t.Fatalf("expected error for nil desired_state, got nil")
	}
}

func TestStage1_GoalKindValidation(t *testing.T) {
	// 2. Every valid GoalKind
	validKinds := []GoalKind{
		GoalKindCommunicative,
		GoalKindStateChange,
		GoalKindToolExecution,
		GoalKindInformationRetrieval,
	}
	for _, k := range validKinds {
		if !k.IsValid() {
			t.Errorf("expected GoalKind %q to be valid", k)
		}
		g := &SemanticGoal{
			Kind:         k,
			Intent:       "test_intent",
			Target:       "test_target",
			DesiredState: map[string]string{"status": "ok"},
		}
		if err := g.Validate(); err != nil {
			t.Errorf("expected SemanticGoal with kind %q to validate cleanly, got: %v", k, err)
		}
	}

	// 3. Invalid GoalKind rejection
	invalidKind := GoalKind("INVALID_KIND_XYZ")
	if invalidKind.IsValid() {
		t.Errorf("expected %q to be invalid", invalidKind)
	}
	invalidGoal := &SemanticGoal{
		Kind:         invalidKind,
		Intent:       "test_intent",
		Target:       "test_target",
		DesiredState: map[string]string{"status": "ok"},
	}
	if err := invalidGoal.Validate(); err == nil {
		t.Fatalf("expected error for invalid GoalKind, got nil")
	}
}

func TestStage1_JSONRoundTrips(t *testing.T) {
	// 4. SemanticGoal JSON round-trip
	goal := &SemanticGoal{
		Kind:          GoalKindToolExecution,
		Intent:        "set_alarm",
		Target:        "alarm_service",
		DesiredState:  map[string]string{"exists": "true", "time": "07:00"},
		OperationHint: "schedule",
		Constraints:   map[string]string{"override": "false"},
	}
	data, err := json.Marshal(goal)
	if err != nil {
		t.Fatalf("failed to marshal SemanticGoal: %v", err)
	}
	var decodedGoal SemanticGoal
	if err := json.Unmarshal(data, &decodedGoal); err != nil {
		t.Fatalf("failed to unmarshal SemanticGoal: %v", err)
	}
	if !reflect.DeepEqual(*goal, decodedGoal) {
		t.Errorf("SemanticGoal round-trip mismatch. got %+v, expected %+v", decodedGoal, *goal)
	}

	// 5. PresentationDirectives JSON round-trip
	directives := &PresentationDirectives{
		Tone:      "conversational",
		Verbosity: "brief",
		Language:  "en-US",
	}
	dData, err := json.Marshal(directives)
	if err != nil {
		t.Fatalf("failed to marshal PresentationDirectives: %v", err)
	}
	var decodedDirectives PresentationDirectives
	if err := json.Unmarshal(dData, &decodedDirectives); err != nil {
		t.Fatalf("failed to unmarshal PresentationDirectives: %v", err)
	}
	if !reflect.DeepEqual(*directives, decodedDirectives) {
		t.Errorf("PresentationDirectives round-trip mismatch. got %+v, expected %+v", decodedDirectives, *directives)
	}
}

func TestStage1_ReasoningHypothesisProposedGoal(t *testing.T) {
	// 6. ReasoningHypothesis with nil ProposedGoal
	hypNil := ReasoningHypothesis{
		ID:                   "hyp-nil",
		Type:                 HypothesisInference,
		Conclusion:           "diagnostic conclusion",
		CalibratedConfidence: 0.85,
		ProposedGoal:         nil,
	}
	if err := hypNil.Validate(); err != nil {
		t.Fatalf("expected valid hypothesis with nil ProposedGoal, got: %v", err)
	}

	// 7. ReasoningHypothesis with valid ProposedGoal
	hypValid := ReasoningHypothesis{
		ID:                   "hyp-valid",
		Type:                 HypothesisInference,
		Conclusion:           "diagnostic conclusion",
		CalibratedConfidence: 0.90,
		ProposedGoal: &SemanticGoal{
			Kind:         GoalKindCommunicative,
			Intent:       "greet_user",
			Target:       "user",
			DesiredState: map[string]string{"acknowledged": "true"},
		},
	}
	if err := hypValid.Validate(); err != nil {
		t.Fatalf("expected valid hypothesis with ProposedGoal, got: %v", err)
	}

	// Hypothesis with invalid ProposedGoal should fail validation
	hypInvalid := ReasoningHypothesis{
		ID:                   "hyp-invalid",
		Type:                 HypothesisInference,
		Conclusion:           "diagnostic conclusion",
		CalibratedConfidence: 0.90,
		ProposedGoal: &SemanticGoal{
			Kind: GoalKind("BAD_KIND"),
		},
	}
	if err := hypInvalid.Validate(); err == nil {
		t.Fatalf("expected validation error when ProposedGoal is invalid, got nil")
	}
}

func TestStage1_ReasoningResultResolvedGoalAndDirectives(t *testing.T) {
	validHyp := ReasoningHypothesis{
		ID:                   "hyp-base",
		Type:                 HypothesisInference,
		Conclusion:           "base conclusion",
		CalibratedConfidence: 0.88,
	}

	// 8. ReasoningResult with nil ResolvedGoal
	resNil := ReasoningResult{
		SchemaVersion:          SchemaVersion,
		EnvelopeID:             "env-nil",
		SourceFrameID:          "frame-nil",
		Status:                 StatusUnambiguousSolved,
		StrategyUsed:           StrategySymbolicFast,
		PrimaryHypothesis:      validHyp,
		ResolvedGoal:           nil,
		PresentationDirectives: nil,
		StrategyTelemetry: StrategyTelemetry{
			EpisodeID:            "ep-nil",
			CalibratedConfidence: 0.88,
		},
	}
	if err := resNil.Validate(); err != nil {
		t.Fatalf("expected valid result with nil ResolvedGoal, got: %v", err)
	}

	// 9. ReasoningResult with valid ResolvedGoal
	resValid := ReasoningResult{
		SchemaVersion:     SchemaVersion,
		EnvelopeID:        "env-valid",
		SourceFrameID:     "frame-valid",
		Status:            StatusUnambiguousSolved,
		StrategyUsed:      StrategySymbolicFast,
		PrimaryHypothesis: validHyp,
		ResolvedGoal: &SemanticGoal{
			Kind:         GoalKindInformationRetrieval,
			Intent:       "diagnose_system",
			Target:       "local_host",
			DesiredState: map[string]string{"root_cause_identified": "true"},
		},
		PresentationDirectives: &PresentationDirectives{
			Tone:      "professional",
			Verbosity: "standard",
			Language:  "en-US",
		},
		StrategyTelemetry: StrategyTelemetry{
			EpisodeID:            "ep-valid",
			CalibratedConfidence: 0.88,
		},
	}
	if err := resValid.Validate(); err != nil {
		t.Fatalf("expected valid result with ResolvedGoal, got: %v", err)
	}

	// Invalid ResolvedGoal should fail validation of ReasoningResult
	resInvalid := resValid.Clone()
	resInvalid.ResolvedGoal.Intent = "" // Make invalid
	if err := resInvalid.Validate(); err == nil {
		t.Fatalf("expected validation error when ResolvedGoal is invalid, got nil")
	}
}

func TestStage1_LegacyDeserializationAndDeterministicSerialization(t *testing.T) {
	// 10. Legacy ReasoningResult JSON deserialization (no resolved_goal, proposed_goal, presentation_directives)
	legacyJSON := []byte(`{
		"schema_version": "2.0",
		"envelope_id": "env-legacy",
		"source_frame_id": "frame-legacy",
		"status": "UNAMBIGUOUS_SOLVED",
		"strategy_used": "STRATEGY_SYMBOLIC_FAST",
		"primary_hypothesis": {
			"id": "hyp-legacy",
			"type": "INFERENCE",
			"conclusion": "Derived symbolic conclusion for intent \"greet_user\"",
			"calibrated_confidence": 0.95
		},
		"strategy_telemetry": {
			"episode_id": "ep-legacy",
			"calibrated_confidence": 0.95
		}
	}`)

	var decodedLegacy ReasoningResult
	if err := json.Unmarshal(legacyJSON, &decodedLegacy); err != nil {
		t.Fatalf("failed to unmarshal legacy JSON: %v", err)
	}
	if err := decodedLegacy.Validate(); err != nil {
		t.Fatalf("legacy ReasoningResult failed validation: %v", err)
	}
	if decodedLegacy.ResolvedGoal != nil || decodedLegacy.PresentationDirectives != nil || decodedLegacy.PrimaryHypothesis.ProposedGoal != nil {
		t.Errorf("expected optional Stage 1 pointer fields to be nil when unmarshaling legacy JSON")
	}

	// 11. Deterministic serialization behavior (omitempty ensures nil pointers omit fields)
	marshaledLegacy, err := json.Marshal(decodedLegacy)
	if err != nil {
		t.Fatalf("failed to marshal decoded legacy result: %v", err)
	}
	var reDecoded map[string]interface{}
	if err := json.Unmarshal(marshaledLegacy, &reDecoded); err != nil {
		t.Fatalf("failed to unmarshal into map: %v", err)
	}
	if _, exists := reDecoded["resolved_goal"]; exists {
		t.Errorf("expected resolved_goal to be omitted from JSON when nil")
	}
	if _, exists := reDecoded["presentation_directives"]; exists {
		t.Errorf("expected presentation_directives to be omitted from JSON when nil")
	}
	hypMap := reDecoded["primary_hypothesis"].(map[string]interface{})
	if _, exists := hypMap["proposed_goal"]; exists {
		t.Errorf("expected proposed_goal to be omitted from hypothesis JSON when nil")
	}
}

func TestStage1_ConcurrentRaceSafeImmutabilityAndCloning(t *testing.T) {
	// 12. Concurrent/race-safe use of the new immutable data contracts
	orig := ReasoningResult{
		SchemaVersion: SchemaVersion,
		EnvelopeID:    "env-concurrent",
		SourceFrameID: "frame-concurrent",
		Status:        StatusUnambiguousSolved,
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                   "hyp-1",
			Type:                 HypothesisInference,
			Conclusion:           "concurrent test",
			CalibratedConfidence: 0.9,
			ProposedGoal: &SemanticGoal{
				Kind:         GoalKindToolExecution,
				Intent:       "set_alarm",
				Target:       "alarm_service",
				DesiredState: map[string]string{"time": "07:00"},
			},
		},
		ResolvedGoal: &SemanticGoal{
			Kind:         GoalKindToolExecution,
			Intent:       "set_alarm",
			Target:       "alarm_service",
			DesiredState: map[string]string{"time": "07:00"},
			Constraints:  map[string]string{"priority": "high"},
		},
		PresentationDirectives: &PresentationDirectives{
			Tone:      "conversational",
			Verbosity: "brief",
			Language:  "en-US",
		},
		StrategyTelemetry: StrategyTelemetry{
			EpisodeID:            "ep-conc",
			CalibratedConfidence: 0.9,
		},
	}

	var wg sync.WaitGroup
	workers := 20
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			// Clone the result concurrently
			clone := orig.Clone()
			if clone.ResolvedGoal == nil || clone.PresentationDirectives == nil || clone.PrimaryHypothesis.ProposedGoal == nil {
				t.Errorf("worker %d: cloned pointer fields are unexpectedly nil", workerID)
				return
			}
			// Mutate clone to verify deep copy isolation
			clone.ResolvedGoal.DesiredState["time"] = "08:00"
			clone.ResolvedGoal.Constraints["priority"] = "low"
			clone.PresentationDirectives.Tone = "concise"
			clone.PrimaryHypothesis.ProposedGoal.DesiredState["time"] = "09:00"
		}(i)
	}

	wg.Wait()

	// Verify original values remained intact
	if orig.ResolvedGoal.DesiredState["time"] != "07:00" {
		t.Errorf("original ResolvedGoal.DesiredState was mutated concurrently")
	}
	if orig.ResolvedGoal.Constraints["priority"] != "high" {
		t.Errorf("original ResolvedGoal.Constraints was mutated concurrently")
	}
	if orig.PresentationDirectives.Tone != "conversational" {
		t.Errorf("original PresentationDirectives.Tone was mutated concurrently")
	}
	if orig.PrimaryHypothesis.ProposedGoal.DesiredState["time"] != "07:00" {
		t.Errorf("original PrimaryHypothesis.ProposedGoal.DesiredState was mutated concurrently")
	}
}

func TestSemanticGoal_Fingerprint_DeterminismAndMapOrder(t *testing.T) {
	// Proves: Same semantic goal + different map insertion order -> same fingerprint
	goalA := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
			"locale":       "en-US",
			"timestamp":    "morning",
		},
		Constraints: map[string]string{
			"read_only": "true",
			"priority":  "high",
		},
	}

	goalB := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"timestamp":    "morning",
			"acknowledged": "true",
			"locale":       "en-US",
		},
		Constraints: map[string]string{
			"priority":  "high",
			"read_only": "true",
		},
	}

	fpA := goalA.Fingerprint()
	fpB := goalB.Fingerprint()

	if fpA == "" || fpB == "" {
		t.Fatalf("expected non-empty fingerprints, got %q and %q", fpA, fpB)
	}
	if fpA != fpB {
		t.Errorf("expected identical fingerprints despite map insertion order differences, got %q vs %q", fpA, fpB)
	}
}

func TestSemanticGoal_Fingerprint_DistinctFields(t *testing.T) {
	base := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
		Constraints: map[string]string{
			"read_only": "true",
		},
		OperationHint: "template-v1",
	}
	baseFP := base.Fingerprint()

	// 1. Different DesiredState -> different fingerprint
	diffDS := base.Clone()
	diffDS.DesiredState["acknowledged"] = "false"
	if diffDS.Fingerprint() == baseFP {
		t.Errorf("expected different fingerprint for different DesiredState")
	}

	// 2. Different Target -> different fingerprint
	diffTarget := base.Clone()
	diffTarget.Target = "admin"
	if diffTarget.Fingerprint() == baseFP {
		t.Errorf("expected different fingerprint for different Target")
	}

	// 3. Different Intent -> different fingerprint
	diffIntent := base.Clone()
	diffIntent.Intent = "farewell_user"
	if diffIntent.Fingerprint() == baseFP {
		t.Errorf("expected different fingerprint for different Intent")
	}

	// 4. Different GoalKind -> different fingerprint
	diffKind := base.Clone()
	diffKind.Kind = GoalKindStateChange
	if diffKind.Fingerprint() == baseFP {
		t.Errorf("expected different fingerprint for different GoalKind")
	}

	// 5. Different meaningful Constraints -> different fingerprint
	diffConstraints := base.Clone()
	diffConstraints.Constraints["read_only"] = "false"
	if diffConstraints.Fingerprint() == baseFP {
		t.Errorf("expected different fingerprint for different Constraints")
	}

	// 6. Different OperationHint only -> SAME fingerprint
	diffHint := base.Clone()
	diffHint.OperationHint = "template-v2"
	if diffHint.Fingerprint() != baseFP {
		t.Errorf("expected identical fingerprint when only OperationHint differs, got %q vs %q", diffHint.Fingerprint(), baseFP)
	}
}

func TestSemanticGoal_Fingerprint_SafetyAndEdgeCases(t *testing.T) {
	// Nil SemanticGoal -> safe behavior ("")
	var nilGoal *SemanticGoal
	if fp := nilGoal.Fingerprint(); fp != "" {
		t.Errorf("expected empty string for nil goal, got %q", fp)
	}

	// Invalid SemanticGoal -> safe behavior ("")
	invalidGoal := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "", // missing intent -> invalid
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
	}
	if fp := invalidGoal.Fingerprint(); fp != "" {
		t.Errorf("expected empty string for invalid goal, got %q", fp)
	}
}
