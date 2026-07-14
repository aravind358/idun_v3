package decision

import (
	"math"
	"sync"
	"time"
)

const (
	// maxAnomalyBuffer is the strict upper bound on structural anomalies retained per episode.
	maxAnomalyBuffer = 16
)

// MicroDecisionAnomaly captures exceptional structural events during reflexive execution.
type MicroDecisionAnomaly struct {
	DecisionID      string    `json:"decision_id"`
	Timestamp       time.Time `json:"timestamp"`
	AnomalyType     string    `json:"anomaly_type"` // "CONSTITUTIONAL_TENSION", "ESCALATED", "CALIBRATION_FAULT"
	TopCandidateID  string    `json:"top_candidate_id"`
	TriggeringRule  string    `json:"triggering_rule,omitempty"`
	ConfidenceScore float64   `json:"confidence_score"`
}

// ReflexiveDecisionTrace is an O(1) memory-bounded episode accumulator.
// Total memory footprint is constant regardless of whether 10 or 10,000,000 micro-decisions occur.
//
// IMMUTABLE ARCHITECTURAL BOUNDARY RULES (Section 7.1):
// 1. Strictly Observational: Contains only bounded statistical summaries and max 16 anomaly records.
// 2. Zero Decision Read-Back: Decision NEVER consults previous or current trace instances during evaluation.
// 3. Exclusive Learning Consumer: Only idun/intelligence/learning uses traces for long-term adaptation.
// 4. Exclusive Reflection Consumer: Only idun/intelligence/reflection evaluates traces for metacognitive audit.
// 5. Immutable Episode Artifact: Treated purely as an immutable observational artifact, never as persistent memory.
type ReflexiveDecisionTrace struct {
	mu sync.Mutex

	EpisodeID       string `json:"episode_id"`
	StrategyVersion string `json:"strategy_version"`

	// 1. Exact Volume Counters
	TotalEvaluated uint64 `json:"total_evaluated"`
	CommitCount    uint64 `json:"commit_count"`
	DeferCount     uint64 `json:"defer_count"`
	AbstainCount   uint64 `json:"abstain_count"`
	EscalateCount  uint64 `json:"escalate_count"`

	// 2. Fixed-Bin Confidence Distribution (10 decile bins: [0-0.1), ..., [0.9-1.0])
	ConfidenceBins [10]uint32 `json:"confidence_bins"`

	// Online statistics using Welford's algorithm
	MeanConfidence float64 `json:"mean_confidence"`
	m2Confidence   float64
	VarianceConf   float64 `json:"variance_conf"`

	// 3. Margin & Ambiguity Telemetry
	MeanTopTwoMargin float64 `json:"mean_top_two_margin"`
	NearTieCount     uint32  `json:"near_tie_count"`

	// 4. Constitutional & Safety Gate Telemetry
	Tier1Rejections uint32            `json:"tier1_rejections"`
	RejectionByRule map[string]uint32 `json:"rejection_by_rule"`

	// 5. Hardware Latency Telemetry (Microseconds summary)
	TotalLatencyUs uint64 `json:"total_latency_us"`
	MaxLatencyUs   uint32 `json:"max_latency_us"`

	// 6. Bounded Ring Buffer of Structural Anomalies (Max 16 records)
	Anomalies []MicroDecisionAnomaly `json:"anomalies"`
}

// NewReflexiveDecisionTrace constructs a clean O(1) accumulator for an episode.
func NewReflexiveDecisionTrace(episodeID, strategyVersion string) *ReflexiveDecisionTrace {
	return &ReflexiveDecisionTrace{
		EpisodeID:       episodeID,
		StrategyVersion: strategyVersion,
		RejectionByRule: make(map[string]uint32),
		Anomalies:       make([]MicroDecisionAnomaly, 0, maxAnomalyBuffer),
	}
}

// RecordDecision ingests a completed reflexive decision into O(1) statistical summaries.
// Raw nominal micro-decisions are discarded immediately after aggregation.
func (t *ReflexiveDecisionTrace) RecordDecision(rec *DecisionRecord, latencyUs uint32, topTwoMargin float64, isNearTie bool, anomaly *MicroDecisionAnomaly) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.TotalEvaluated++

	switch rec.SelectedOutcome {
	case OutcomeCommit:
		t.CommitCount++
	case OutcomeDefer:
		t.DeferCount++
	case OutcomeAbstain:
		t.AbstainCount++
	case OutcomeEscalateToDeliberative:
		t.EscalateCount++
	}

	// Update confidence decile bin
	conf := math.Max(0.0, math.Min(0.9999, rec.Confidence))
	binIndex := int(conf * 10)
	if binIndex >= 0 && binIndex < 10 {
		t.ConfidenceBins[binIndex]++
	}

	// Update Welford's online mean and variance for confidence
	delta := rec.Confidence - t.MeanConfidence
	t.MeanConfidence += delta / float64(t.TotalEvaluated)
	delta2 := rec.Confidence - t.MeanConfidence
	t.m2Confidence += delta * delta2
	if t.TotalEvaluated > 1 {
		t.VarianceConf = t.m2Confidence / float64(t.TotalEvaluated-1)
	}

	// Update top-two margin mean
	marginDelta := topTwoMargin - t.MeanTopTwoMargin
	t.MeanTopTwoMargin += marginDelta / float64(t.TotalEvaluated)
	if isNearTie {
		t.NearTieCount++
	}

	// Record Tier 1 rejections
	for _, rej := range rec.RejectedCandidates {
		if rej.RejectionStage == "TIER_1_CONSTITUTION" {
			t.Tier1Rejections++
			t.RejectionByRule[rej.PrimaryReason]++
		}
	}

	// Update latency summary
	t.TotalLatencyUs += uint64(latencyUs)
	if latencyUs > t.MaxLatencyUs {
		t.MaxLatencyUs = latencyUs
	}

	// If anomaly occurred, append to bounded ring buffer (dropping oldest if > 16)
	if anomaly != nil {
		if len(t.Anomalies) >= maxAnomalyBuffer {
			// Shift left by 1
			copy(t.Anomalies, t.Anomalies[1:])
			t.Anomalies[maxAnomalyBuffer-1] = *anomaly
		} else {
			t.Anomalies = append(t.Anomalies, *anomaly)
		}
	}
}

// Snapshot returns a thread-safe copy of the current accumulator state.
func (t *ReflexiveDecisionTrace) Snapshot() ReflexiveDecisionTrace {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := *t
	// Copy map and slice
	out.RejectionByRule = make(map[string]uint32, len(t.RejectionByRule))
	for k, v := range t.RejectionByRule {
		out.RejectionByRule[k] = v
	}
	out.Anomalies = make([]MicroDecisionAnomaly, len(t.Anomalies))
	copy(out.Anomalies, t.Anomalies)
	return out
}
