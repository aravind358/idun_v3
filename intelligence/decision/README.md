# IDUN Intelligence Pillar: Decision Subsystem (`idun/intelligence/decision`)

**Architecture Version:** `2.0.0-FROZEN`  
**Classification:** Core Cognitive Ability Specification — Commitment Under Uncertainty  
**Status:** PERMANENT ARCHITECTURE SPECIFICATION FROZEN BEFORE IMPLEMENTATION  

---

## 1. Overview & Core Architectural Identity

Every cognitive architecture reaches branching points where multiple courses of action are simultaneously valid. **Decision** (`CognitiveAbility.Decision`) is the sole cognitive ability authorized to perform **Commitment Under Uncertainty**—collapsing a space of live candidate alternatives ($C = \{c_1, c_2, \dots, c_n\}$) into a single committed outcome ($c^*$) or an explicit epistemic non-choice (`DEFER`, `ABSTAIN`, `REQUEST_MORE_CANDIDATES`, `REQUEST_ADDITIONAL_INFO`).

### 1.1 Defining Cognitive Question
* **Understanding:** *"What does this mean?"*
* **Reasoning:** *"What conclusions can I derive?"*
* **Reflection:** *"How well did I think?"*
* **Decision:** *"Given these valid alternatives, which one do we commit to — and why?"*

---

## 2. Immutable Responsibility Boundaries

### 2.1 What Decision Must Own (Unique Verb: **SELECT / COMMIT**)
1. Accept an externally generated candidate set $C$ ($1 \le |C| \le 16$).
2. Accept applicable constitutional constraints, operational goals, risk preferences, and a deliberation budget.
3. Apply strict Tier 1 hard constitutional gates before Tier 2 objective scoring.
4. Select exactly one terminal outcome and emit an auditable `DecisionRecord`.

### 2.2 What Decision Never Owns
* **Option Generation (`Reasoning`):** Never invents candidate options from nothing.
* **Execution & Enactment (`Executive`):** Never carries out physical or external actions.
* **Temporal Plan Sequencing (`Planning`):** Never sequences actions across multi-step time horizons.
* **Self-Evaluation & Policy Update (`Reflection` / `Learning`):** Never evaluates its own historical performance or mutates its runtime strategy weights.
* **Input Interpretation (`Understanding`):** Never converts raw sensory/textual input into meaning.
* **Goal Definition (`Constitution` / `Executive`):** Never sets top-level goals.

---

## 3. Tiered Input Evaluation Hierarchy

Decision evaluates candidates through a strict, immutable 4-tier pipeline:

```
Tier 1: Hard Constitutional Gates (Non-Negotiable Binary Filter)
  ├── Input: Constitution constraints & safety rules
  └── Action: Immediate disqualification of violating candidates into RejectedAlternatives.
  
Tier 4: Deliberation Budget Bounds (Computational Expenditure Gate)
  ├── Input: Time available, depth tier (REFLEXIVE_MICRO vs DELIBERATIVE_MACRO)
  └── Action: Selects algorithm complexity (Linear utility dot-product vs MCDA / Pareto search).
  
Tier 3: Epistemic Context & Calibration (Confidence Modulators)
  ├── Input: Situational meaning, historical memory precedents, CalibrationService trust reports
  └── Action: Modulates candidate confidence intervals and risk weights before scoring.
  
Tier 2: Objective Function Weighing (Utility & Risk Scoring)
  ├── Input: Executive goals, host preferences, risk profile
  └── Action: Scores surviving candidates and selects optimal commitment c*.
```

---

## 4. Dual-Mode Polymorphic Execution Model

To serve both high-frequency micro-choices (e.g., 200–300 branch prunings during a `Reasoning` episode) and high-stakes strategic forks without creating split subsystems or message-bus saturation, Decision implements a **Unified Polymorphic Engine** with two execution surfaces:

### 4.1 Execution Surface Comparison Matrix

| Architectural Dimension | Reflexive Micro-Decision (`DEPTH_REFLEXIVE`) | Deliberative Macro-Decision (`DEPTH_DELIBERATIVE`) |
| :--- | :--- | :--- |
| **Primary Caller** | Intra-episode abilities (`Reasoning`, `Planning`) | Control plane (`Executive`, High-stakes `Planning`) |
| **Invocation Transport** | Direct synchronous library call (`EvaluateReflexive`) | Global Workspace envelope (`communication.Envelope`) |
| **Latency Budget** | $< 2\text{ ms}$ (Target $< 0.5\text{ ms}$) | $50\text{--}500\text{ ms}$ |
| **Constitution Tier 1 Gate** | **MANDATORY** (Compiled bitmask / fast-path gate) | **MANDATORY** (Full Constitutional Action Gate check) |
| **Strategy Snapshot** | Shared immutable `DecisionStrategySnapshot` | Shared immutable `DecisionStrategySnapshot` |
| **Scoring Algorithm** | Calibrated linear utility score: $U(c_i) = \mathbf{w}^T \mathbf{x}_i$ | Multi-Criteria Decision Analysis (MCDA) / Pareto trade-off |
| **Telemetry Emission** | Compressed into episode `ReflexiveDecisionTrace` | Full immediate `DecisionRecord` published to Workspace |
| **Global Workspace Broadcast** | **SKIPPED** (Prevent bus flood & deadlock) | **PUBLISHED** (`TopicDecisionRecord`) |

---

## 5. Public Output Contract: `DecisionRecord`

Every decision, regardless of mode, conforms to the standardized output contract:

```go
package decision

import (
	"context"
	"time"
)

type OutcomeType string

const (
	OutcomeCommit                OutcomeType = "COMMIT"
	OutcomeDefer                 OutcomeType = "DEFER"
	OutcomeAbstain               OutcomeType = "ABSTAIN"
	OutcomeRequestCandidates     OutcomeType = "REQUEST_MORE_CANDIDATES"
	OutcomeRequestAdditionalInfo OutcomeType = "REQUEST_ADDITIONAL_INFO"
	OutcomeEscalateToDeliberative OutcomeType = "ESCALATE_TO_DELIBERATIVE"
)

type DeliberationDepth string

const (
	DepthReflexive    DeliberationDepth = "REFLEXIVE_MICRO"
	DepthDeliberative DeliberationDepth = "DELIBERATIVE_MACRO"
)

type InformationGap struct {
	CandidateID      string `json:"candidate_id"`
	MissingAttribute string `json:"missing_attribute"`
	Reason           string `json:"reason"`
	ImpactOnChoice   string `json:"impact_on_choice"`
	TargetProvider   string `json:"target_provider"` // "UNDERSTANDING", "MEMORY", "HOST_INPUT"
}

type RejectedAlternative struct {
	CandidateID    string  `json:"candidate_id"`
	RejectionStage string  `json:"rejection_stage"` // "TIER_1_CONSTITUTION", "TIER_2_SCORING"
	PrimaryReason  string  `json:"primary_reason"`
	ScoreDelta     float64 `json:"score_delta"`     // Normalized distance from winning choice
}

type EscalationRecommendation struct {
	TriggeredDimensions []string `json:"triggered_dimensions"` // e.g. "CONFIDENCE_DROP", "AMBIGUITY_MARGIN", "TAIL_RISK"
	ConfidenceDelta     float64  `json:"confidence_delta"`
	UtilityScoreMargin  float64  `json:"utility_score_margin"`
	Reason              string   `json:"reason"`
}

type DecisionRecord struct {
	DecisionID               string                    `json:"decision_id"`
	EpisodeID                string                    `json:"episode_id"`
	Timestamp                time.Time                 `json:"timestamp"`
	StrategyVersion          string                    `json:"strategy_version"`
	DeliberationDepth        DeliberationDepth         `json:"deliberation_depth"`
	SelectedOutcome          OutcomeType               `json:"selected_outcome"`
	SelectedCandidateID      string                    `json:"selected_candidate_id,omitempty"`
	Confidence               float64                   `json:"confidence"`
	Rationale                string                    `json:"rationale"`
	RejectedCandidates       []RejectedAlternative     `json:"rejected_candidates"`
	ConstraintsApplied       []string                  `json:"constraints_applied"`
	InformationGaps          []InformationGap          `json:"information_gaps,omitempty"`
	FlaggedAssumptions       []string                  `json:"flagged_assumptions"`
	EscalationRecommendation *EscalationRecommendation `json:"escalation_recommendation,omitempty"`
	TradeoffMatrix           map[string]map[string]float64 `json:"tradeoff_matrix,omitempty"`
}
```

---

## 6. Reflexive → Deliberative Escalation Workflow

When `Decision` operates at `REFLEXIVE_MICRO` depth ($< 2\text{ ms}$ budget), it must never automatically force an escalation nor execute Deliberative MCDA synchronously. Instead, if confidence, ambiguity margin, or risk thresholds are breached, `Decision` emits an **Escalation Recommendation** (`SelectedOutcome = OutcomeEscalateToDeliberative`).

```
Reflexive Decision
        ↓
Outcome: ESCALATE_TO_DELIBERATIVE
        ↓
Reasoning / Planning / Executive
        ↓
Caller decides whether to spend additional computational budget
        ↓
If approved
        ↓
Decision is invoked again using DELIBERATIVE_MACRO
```

This preserves strict responsibility boundaries:
* **Decision** recommends escalation under uncertainty.
* **Executive** (or calling cognitive ability) owns computational budget and control flow.
* **Decision** never dictates how the caller must proceed.

---

## 7. Long-Term Evolution & Bounded O(1) Telemetry Contract

1. **Immutable Strategy Snapshots:** At episode start, Decision acquires a read-only `DecisionStrategySnapshot` (versioned weights and policies) published by `Learning`. Both Reflexive and Deliberative surfaces execute against this identical snapshot.
2. **Fixed-Memory ($O(1)$) `ReflexiveDecisionTrace` ABI:** To support billions of micro-decisions across a 20–30 year lifecycle without storage write amplification or memory leaks, `ReflexiveDecisionTrace` is strictly bounded to $O(1)$ memory consumption ($< 4\text{ KB}$ footprint per episode). Raw micro-decisions are **never** persisted except under structural anomaly conditions.

```go
package decision

import "time"

// ReflexiveDecisionTrace is an O(1) memory-bounded episode accumulator.
type ReflexiveDecisionTrace struct {
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
	MeanConfidence float64    `json:"mean_confidence"`
	VarianceConf   float64    `json:"variance_conf"` // Welford's online algorithm

	// 3. Margin & Ambiguity Telemetry
	MeanTopTwoMargin float64 `json:"mean_top_two_margin"`
	NearTieCount     uint32  `json:"near_tie_count"`

	// 4. Constitutional & Safety Gate Telemetry
	Tier1Rejections uint32            `json:"tier1_rejections"`
	RejectionByRule map[string]uint32 `json:"rejection_by_rule"`

	// 5. Hardware Latency Distribution (Microseconds)
	LatencyP50Us uint32 `json:"latency_p50_us"`
	LatencyP95Us uint32 `json:"latency_p95_us"`
	LatencyP99Us uint32 `json:"latency_p99_us"`

	// 6. Bounded Ring Buffer of Structural Anomalies (Max 16 records)
	Anomalies []MicroDecisionAnomaly `json:"anomalies"`
}

type MicroDecisionAnomaly struct {
	DecisionID      string    `json:"decision_id"`
	Timestamp       time.Time `json:"timestamp"`
	AnomalyType     string    `json:"anomaly_type"` // "CONSTITUTIONAL_TENSION", "ESCALATED", "CALIBRATION_FAULT"
	TopCandidateID  string    `json:"top_candidate_id"`
	TriggeringRule  string    `json:"triggering_rule,omitempty"`
	ConfidenceScore float64   `json:"confidence_score"`
}
```

3. **Decade-Safe Adaptation:** `Learning` updates internal scoring weights over years without modifying the public `DecisionAbility` interface or violating Tier 1 Constitutional safety gates.

### 7.1 ReflexiveDecisionTrace Ownership & Write/Read Isolation Boundaries (Anti-Memory-Leak Invariant)

To guarantee that `ReflexiveDecisionTrace` never accidentally evolves into a second memory subsystem or violates the ownership boundaries of `idun/core/memory`, `idun/intelligence/learning`, or `idun/intelligence/reflection`, the following **5 Immutable Boundary Invariants** are enforced:

1. **Strictly Bounded Observational Contents:** `ReflexiveDecisionTrace` contains only fixed-size ($O(1)$) statistical summaries and a bounded anomaly ring buffer (`len(Anomalies) <= 16`). Nominal raw micro-decisions are discarded immediately after aggregation.
2. **Zero Decision Read-Back:** `Decision` never queries, reads, or consults previous or current `ReflexiveDecisionTrace` instances when evaluating future decisions. Decision evaluation (`EvaluateReflexive`, `EvaluateDeliberative`) depends exclusively on the input `CandidateSet` and the read-only `DecisionStrategySnapshot`.
3. **Exclusive Long-Term Adaptation Consumer (`Learning`):** `Learning` (`idun/intelligence/learning`) is the solely authorized consumer for inter-episode policy distillation and weight calibration.
4. **Exclusive Metacognitive Audit Consumer (`Reflection`):** `Reflection` (`idun/intelligence/reflection`) is the solely authorized consumer for episode-end self-critique, overconfidence detection, and contradiction auditing.
5. **Purely Observational Episode Artifact:** `Decision` treats `ReflexiveDecisionTrace` purely as a write-only observational telemetry accumulator during an episode and an immutable episode artifact upon closure—never as episodic or semantic storage.

### 7.2 Decision Policy Profiles (Versioned Strategy Snapshots)

Rather than assuming one universal objective-weight configuration, `DecisionStrategySnapshot` supports versioned **Decision Policy Profiles** (`DecisionPolicyProfile`). Each profile contains cohesive decision parameters tailored to a specific operating mode or risk regime:
* **Utility weights** (`FeatureWeights`)
* **Risk tolerance** (`RiskTolerance`)
* **Escalation thresholds** (`EscalationConfidenceFloor`, `EscalationAmbiguityMargin`)
* **Objective priorities** (`ObjectivePriorities`)
* **Resource preferences** (`MaxReflexiveLatencyUs`)

```go
type DecisionPolicyProfile struct {
	ProfileID                 string             `json:"profile_id"`
	PolicyVersion             string             `json:"policy_version"`
	PolicySource              string             `json:"policy_source"`
	PolicyFingerprint         string             `json:"policy_fingerprint"`
	Description               string             `json:"description"`
	FeatureWeights            map[string]float64 `json:"feature_weights"`
	RiskTolerance             float64            `json:"risk_tolerance"`
	EscalationConfidenceFloor float64            `json:"escalation_confidence_floor"`
	EscalationAmbiguityMargin float64            `json:"escalation_ambiguity_margin"`
	ObjectivePriorities       []string           `json:"objective_priorities"`
	MaxReflexiveLatencyUs     uint32             `json:"max_reflexive_latency_us"`
}
```

#### Strict Passive Consumption Contract
* **Complete Passivity:** `Decision` remains completely passive. It never determines which profile is active, nor does it generate, modify, or optimize profiles.
* **Responsibility Invariance:** `Decision` simply consumes the active immutable profile contained in `DecisionStrategySnapshot` to execute its sole responsibility: **Select / Commit under uncertainty**.

---

## 8. Production Hardening Invariants (Phase 5 Contract)

To ensure decades of auditable, deterministic, and privacy-preserving operation across all cognitive pipelines, `Decision` enforces five strict production invariants:

1. **Immutable DecisionRecord (Single-Writer Principle):**  
   Once published to the Global Workspace, a `DecisionRecord` is permanently immutable. No subsystem may rewrite historical decision records. Reevaluations emit a new `DecisionRecord` with a distinct `DecisionID`.
2. **Schema Versioning (`SchemaVersion`):**  
   Every `DecisionRecord` explicitly records `SchemaVersion` (e.g., `"2.0.0-FROZEN"`) to guarantee correct historical interpretation by `Reflection` and `Learning` over decades.
3. **Telemetry Privacy Invariant (`ReflexiveDecisionTrace`):**  
   `ReflexiveDecisionTrace` is strictly operational telemetry. It never records raw conversation text, prompts, host messages, PII, or semantic memory contents—only bounded counters, decile distributions, and anonymized anomaly identifiers.
4. **Validation Firewall (`DecisionRecord.Validate()`):**  
   Every `DecisionRecord` must pass strict structural validation prior to publication (`DecisionID`, `EpisodeID`, `SchemaVersion`, `[0,1]` confidence bounds, depth validity, and candidate selection integrity). Invalid records are discarded immediately.
5. **Deterministic Replay Invariant (`ReplaySeed`):**  
   Given identical inputs (`CandidateSet`, `DecisionStrategySnapshot`, `DecisionPolicyProfile`, and `PolicyFingerprint`), evaluation produces an identical `DecisionRecord`. Any stochastic sampling records explicit provenance (`ReplaySeed`) to enable deterministic replay for scientific debugging and reflection.



