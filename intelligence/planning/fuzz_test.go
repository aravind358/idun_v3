package planning

import (
	"encoding/json"
	"testing"
	"time"
)

// FuzzPlanValidation fuzzes Plan deserialization and structural validation against arbitrary bytes.
func FuzzPlanValidation(f *testing.F) {
	// Seed with valid and malformed JSON
	validPlan := Plan{
		PlanID:             "p-1",
		SchemaVersion:      SchemaVersion2_0_0,
		StrategySnapshotID: "snap-1",
		Goal:               "Test Goal",
		TraceID:            "trace-1",
		EstimatedCost:      10.0,
		Subgoals:           []Subgoal{{SubgoalID: "sg-1", Title: "Step 1"}},
	}
	rawValid, _ := json.Marshal(validPlan)
	f.Add(string(rawValid))
	f.Add("")
	f.Add("not-json")
	f.Add(`{"plan_id":"p-1","schema_version":"invalid"}`)
	f.Add(`{"plan_id":"p-1","schema_version":"2.0.0-FROZEN","estimated_cost":-50.0}`)

	f.Fuzz(func(t *testing.T, data string) {
		var plan Plan
		if err := json.Unmarshal([]byte(data), &plan); err == nil {
			_ = plan.Validate()
		}
	})
}

// FuzzPlanningTraceValidation fuzzes PlanningTrace structural validation and bounds checking.
func FuzzPlanningTraceValidation(f *testing.F) {
	validTrace := PlanningTrace{
		TraceID:            "trace-1",
		SchemaVersion:      SchemaVersion2_0_0,
		PlanID:             "p-1",
		StrategySnapshotID: "snap-1",
		PolicyFingerprint:  "pol-fp",
		TerminationReason:  TerminationGoalFound,
		QualityMetrics:     QualityMetrics{Completeness: 0.8, Efficiency: 0.9},
	}
	rawValid, _ := json.Marshal(validTrace)
	f.Add(string(rawValid))
	f.Add(`{"trace_id":"trace-1","schema_version":"2.0.0-FROZEN","quality_metrics":{"completeness":1.5}}`)
	f.Add(`{"trace_id":"trace-1","schema_version":"2.0.0-FROZEN","confidence_profile":{"overall_confidence":-0.1}}`)

	f.Fuzz(func(t *testing.T, data string) {
		var trace PlanningTrace
		if err := json.Unmarshal([]byte(data), &trace); err == nil {
			_ = trace.Validate()
		}
	})
}

// FuzzPlanningRequest fuzzes PlanningRequest validation against arbitrary string inputs.
func FuzzPlanningRequest(f *testing.F) {
	f.Add("req-1", "Decompose goal", "General", int64(500), 0.8)
	f.Add("", "Decompose goal", "General", int64(500), 0.8)
	f.Add("req-1", "", "General", int64(500), 0.8)
	f.Add("req-1", "Goal", "General", int64(-10), 1.5)

	f.Fuzz(func(t *testing.T, reqID, goal, domain string, budgetMs int64, floor float64) {
		req := PlanningRequest{
			RequestID:          reqID,
			Goal:               goal,
			Domain:             domain,
			MaxExecutionBudget: time.Duration(budgetMs) * time.Millisecond,
			MinConfidenceFloor: floor,
		}
		_ = req.Validate()
	})
}

// FuzzReplayMetadata fuzzes ReplayMetadata validation against random fidelity tags and IDs.
func FuzzReplayMetadata(f *testing.F) {
	f.Add("snap-1", "EXACT", uint64(12345), "hash-abc")
	f.Add("", "EXACT", uint64(12345), "hash-abc")
	f.Add("snap-1", "INVALID_FIDELITY", uint64(12345), "hash-abc")

	f.Fuzz(func(t *testing.T, snapID, fidelity string, seed uint64, memHash string) {
		rm := ReplayMetadata{
			StrategySnapshotID: snapID,
			ReplayFidelity:     fidelity,
			ReplaySeed:         seed,
			WorkingMemoryHash:  memHash,
		}
		_ = rm.Validate()
	})
}

// FuzzPlanningCapabilities fuzzes PlanningCapabilities bounds checking.
func FuzzPlanningCapabilities(f *testing.F) {
	f.Add(uint16(8), uint16(10), uint16(50), true, true)
	f.Add(uint16(0), uint16(0), uint16(0), false, false)
	f.Add(uint16(65535), uint16(1000), uint16(1000), true, true)

	f.Fuzz(func(t *testing.T, workers uint16, depth, alts uint16, htn, goap bool) {
		caps := PlanningCapabilities{
			MaxParallelWorkers:       workers,
			MaxPlanningDepth:         depth,
			MaxSupportedAlternatives: alts,
			SupportsHTN:              htn,
			SupportsGOAP:             goap,
		}
		_ = caps.Validate()
	})
}

// FuzzPlanningSearchStrategy fuzzes PlanningSearchStrategy structural rules and budget bounds.
func FuzzPlanningSearchStrategy(f *testing.F) {
	f.Add("search-1", "v1.0", "fp-abc", uint32(5), uint32(100), uint32(3), true)
	f.Add("", "v1.0", "fp-abc", uint32(5), uint32(100), uint32(3), true)
	f.Add("search-1", "", "fp-abc", uint32(0), uint32(0), uint32(0), false)

	f.Fuzz(func(t *testing.T, id, ver, fp string, depth, nodes, beam uint32, parallel bool) {
		strat := PlanningSearchStrategy{
			SearchID:               id,
			SearchVersion:          ver,
			SearchFingerprint:      fp,
			SearchType:             "BEAM_SEARCH",
			MaxDepth:               depth,
			MaxNodes:               nodes,
			BeamWidth:              beam,
			AllowParallelExpansion: parallel,
		}
		_ = strat.Validate()
	})
}

// FuzzSpecialistUsage fuzzes PlanningSpecialistUsage scoring and bounds checking.
func FuzzSpecialistUsage(f *testing.F) {
	f.Add("spec-htn", float32(0.8), string(SkipNone), uint64(50), uint64(1500), true, true)
	f.Add("spec-goap", float32(1.5), string(SkipDomainMismatch), uint64(0), uint64(0), false, false)
	f.Add("", float32(-0.2), string(SkipBudgetExceeded), uint64(10), uint64(500), true, false)

	f.Fuzz(func(t *testing.T, specID string, score float32, skip string, nodes uint64, timeUs uint64, success, invoked bool) {
		usage := PlanningSpecialistUsage{
			SpecialistID:      specID,
			Invoked:           invoked,
			ContributionScore: score,
			SkipReason:        SpecialistSkipReason(skip),
			NodesExpanded:     nodes,
			ExecutionTimeUs:   timeUs,
			Success:           success,
		}
		_ = usage.Validate()
	})
}

// FuzzFingerprints fuzzes canonical plan and capability fingerprint calculation against arbitrary inputs.
func FuzzFingerprints(f *testing.F) {
	fpGen := NewDefaultPlanFingerprinter()
	f.Add("plan-1", "Goal 1", "General", "TACTICAL", "sg-1", "Step 1")
	f.Add("", "", "", "", "", "")

	f.Fuzz(func(t *testing.T, planID, goal, domain, tier, sgID, sgTitle string) {
		plan := &Plan{
			PlanID:        planID,
			SchemaVersion: SchemaVersion2_0_0,
			CreatedAt:     time.Now().UTC(),
			Goal:          goal,
			Domain:        domain,
			SourceTier:    tier,
		}
		if sgID != "" && sgTitle != "" {
			plan.Subgoals = append(plan.Subgoals, Subgoal{SubgoalID: sgID, Title: sgTitle})
		}
		_, _ = fpGen.ComputeFingerprint(plan)
	})
}
