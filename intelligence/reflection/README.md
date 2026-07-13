# IDUN V3 Reflection Cognitive Ability Architecture

**Architecture Version:** `2.0.0-FROZEN`  
**Package:** `idun/intelligence/reflection`  
**Classification:** Standalone Cognitive Ability Architecture Specification  
**Status:** PERMANENT ARCHITECTURE SPECIFICATION (FROZEN FOR 20–30 YEAR LIFECYCLE)

---

## 1. Purpose & Architectural Philosophy

Reflection is IDUN's complete self-evaluation metacognitive system. It exists to improve the quality of future cognition, not merely to judge past cognition. Its output is structured raw material for **Learning** — Reflection itself never learns, never decides, never plans, and never executes actions.

```
+-----------------------------------------------------------------------------------+
|                                 GLOBAL WORKSPACE                                  |
+-----------------------------------------------------------------------------------+
       | (Read-Only Traces)                          ^ (Structured ReflectionReport)
       v                                             |
+-----------------------------------------------------------------------------------+
|                           REFLECTION COGNITIVE ABILITY                            |
|                                                                                   |
|  +-----------------------------------------------------------------------------+  |
|  |                   Specialist Evaluator Family (S1 - S11)                    |  |
|  |  [Understanding] [Reasoning] [Decision] [Planning] [Learning] [Attention]   |  |
|  |  [Conversation] [Executive] [Overall Cognitive Assessment]                  |  |
|  |  [Trend Reflection] [Reflection-on-Reflection]                              |  |
|  +-----------------------------------------------------------------------------+  |
|                                     |                                             |
|                                     v                                             |
|                     EvaluationStrategy Abstraction Layer                          |
+-----------------------------------------------------------------------------------+
```

Reflection evaluates. Everything else about what happens next belongs to other specialized cognitive abilities.

---

## 2. Core Design Principles

### 2.1 Multi-Decade Architectural Invariance
Multi-decade operation without system redesign requires separating the **fixed structural boundary** from the **continuously evolving strategy layer**:
- **Fixed for the Life of the System:** Read-only trace ingestion contracts, versioned `ReflectionReport` schemas, versioned `HistoricalSummary` contracts, one-directional consumption boundaries, and optional non-blocking infrastructure isolation.
- **Continuously Evolving Layer:** Specialist evaluation logic sitting behind a stable `EvaluationStrategy` interface—starting as heuristic evaluators and evolving over decades into statistical, learned, or hybrid evaluation models.

### 2.2 Continuous Improvement Principle
> **Architectural Principle:** *Reflection does not seek perfection. Reflection seeks continuous improvement.*

Reflection optimizes future cognition rather than attempting to produce brittle or intractable absolute perfection proofs. Every `EvaluationStrategy` must evaluate cognitive performance by measuring **trajectory, relative delta, error-rate velocity, and gradient of improvement** against past baselines.

---

## 3. Operating Modes

Reflection executes in two distinct modes under identical structural boundaries:

### 3.1 Episode Reflection
- **Trigger:** Invoked asynchronously by `Executive` immediately after a cognitive episode closes.
- **Purpose:** Evaluate that single completed episode across all participating cognitive abilities.
- **Scope:** Read-only traces of that specific episode, plus lightweight verification of earlier predictions that resolved during the episode.

### 3.2 Periodic Reflection
- **Trigger:** Scheduled by `Executive` during idle periods, low-CPU budget windows, or maintenance cycles. Never executes on the critical cognitive path.
- **Purpose:** Detect longer-horizon longitudinal trends, strategy drift, recurring failure motifs, or structural improvements across thousands of episodes.
- **Scope:** Versioned historical summaries (`HistoricalSummary`) supplied read-only by `Memory`.

Both modes produce the unified `ReflectionReport` envelope structure, distinguished by the `Mode` field (`MODE_EPISODE` vs. `MODE_PERIODIC`).

---

## 4. Trace Ingestion & Immutable Trace Lineage

### 4.1 Read-Only Trace Isolation
Reflection has zero direct access to any adjacent cognitive ability's internal memory, parameters, or execution state. All participating abilities emit structured execution traces to the Global Workspace (`communication.Envelope`); Reflection consumes these traces strictly read-only.

### 4.2 Immutable Trace Lineage (`source_trace_refs[]`)
Every specialist evaluation finding must preserve immutable causal references back to the exact Workspace envelopes that triggered the finding:

```go
type TraceReference struct {
    EnvelopeID      string `json:"envelope_id"`
    SourceAbility   string `json:"source_ability"`
    TraceTimestamp  int64  `json:"trace_timestamp"`
    PayloadHashRef  string `json:"payload_hash_ref"`
}
```

This ensures that `Reflection-on-Reflection` can audit and verify the exact empirical evidence behind any historical evaluation years or decades later.

---

## 5. Specialist Architecture & Explicit Evaluation Verdicts

### 5.1 Specialist Evaluator Family
Reflection organizes evaluation across 11 specialized evaluators:
1. **Understanding Reflection Specialist:** Evaluates slot accuracy, ambiguity resolution, and semantic framing.
2. **Reasoning Reflection Specialist:** Evaluates cascade efficiency, hypothesis contradiction rates, and beam calibration.
3. **Decision Reflection Specialist:** Evaluates utility estimation, risk calibration, and action selection bias.
4. **Planning Reflection Specialist:** Evaluates DAG feasibility, step ordering, and resource bounding.
5. **Learning Reflection Specialist:** Evaluates rule compilation stability and parameter adaptation drift.
6. **Attention Reflection Specialist:** Evaluates salience filtering and context window relevance.
7. **Conversation Reflection Specialist:** Evaluates dialogue coherence, pragmatics, and user alignment.
8. **Executive Reflection Specialist:** Evaluates scheduler fairness, budget allocation, and deadline adherence.
9. **Overall Cognitive Assessment Specialist:** Synthesis specialist with full visibility across all episode specialist reports. Implements two distinct, separately-versioned capabilities:
   - *Cross-Cognitive Analysis:* Identifies inter-ability interaction issues (e.g., strong planning paired with poor decision execution handoff).
   - *Growth Potential Estimation:* Computes high-ROI learning targets across abilities.
10. **Trend Reflection Specialist (Periodic Mode):** Consumes `HistoricalSummary` contracts to identify longitudinal behavioral shifts.
11. **Reflection-on-Reflection Specialist:** Audits earlier `ReflectionReport` findings against subsequent long-term outcomes to evaluate whether Reflection's own judgments were well-calibrated.

### 5.2 Explicit Evaluation Verdicts (`ABSTAIN` / `INSUFFICIENT_DATA`)
Specialists must never be forced to fabricate evaluations when traces are sparse, abbreviated, or prematurely cancelled. Every specialist sub-report declares an explicit `EvaluationVerdict`:

```go
type EvaluationVerdict string

const (
    VerdictEvaluated        EvaluationVerdict = "EVALUATED"
    VerdictInsufficientData EvaluationVerdict = "INSUFFICIENT_DATA"
    VerdictAbstain          EvaluationVerdict = "ABSTAIN"
)
```

When evidence is genuinely insufficient, specialists return `VERDICT_INSUFFICIENT_DATA` or `VERDICT_ABSTAIN`. Downstream consumers (`Learning` and `Trend Reflection`) filter out abstained reports, preventing noise pollution and statistical distortion over decades.

---

## 6. Evaluation Strategy Abstraction & Decadal Self-Improvement Loop

### 6.1 `EvaluationStrategy` Interface
Each specialist delegates judgment logic to a stable, swappable interface:

```go
type EvaluationStrategy interface {
    Evaluate(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error)
    StrategyID() string
    Version() string
}
```

### 6.2 Closing the Loop Without Crossing Boundaries
Reflection never mutates its own strategy parameters. Decadal self-improvement operates through a strict three-phase causal chain:
1. **Reflection-on-Reflection** audits earlier `ReflectionReport` predictions against real-world outcomes and emits an outcome assessment within `ReflectionReport.self_calibration`.
2. **Learning** consumes `ReflectionReport.self_calibration`, evaluates evaluator accuracy, and computes optimal strategy retunings.
3. **Learning** updates the parameter weights of Reflection's `EvaluationStrategy` instances.

```
+--------------------------+       (Self-Calibration Data)       +--------------------------+
|  Reflection-on-Reflection| ----------------------------------> |         Learning         |
+--------------------------+                                     +--------------------------+
             ^                                                                     |
             | (Evaluates Past Reports)                 (Updates Strategy Weights) |
             |                                                                     v
+--------------------------+                                     +--------------------------+
|  Historical Workspace    |                                     |    EvaluationStrategy    |
+--------------------------+                                     +--------------------------+
```

---

## 7. Versioned Historical Summary Contract

To prevent structural coupling to `Memory`'s internal storage schemas over 20–30 years, periodic trend analysis is governed by formal versioned contracts:

### 7.1 `HistoricalSummaryRequest` Contract
Emitted by `Periodic Reflection` to request pre-aggregated longitudinal data:

```go
type HistoricalSummaryRequest struct {
    SchemaVersion    string            `json:"schema_version"`    // e.g., "2.0.0"
    RequestID        string            `json:"request_id"`
    TimeWindow       TimeWindowSpec    `json:"time_window"`
    TargetAbilities  []string          `json:"target_abilities"`  // e.g., ["reasoning", "planning"]
    AggregationLevel string            `json:"aggregation_level"` // "HOURLY" | "DAILY" | "WEEKLY"
    RequestedMetrics []string          `json:"requested_metrics"` // e.g., ["failure_rates", "confidence_drift"]
}
```

### 7.2 `HistoricalSummary` Contract
Returned read-only by `Memory` to `Trend Reflection`:

```go
type HistoricalSummary struct {
    SchemaVersion     string                    `json:"schema_version"` // "2.0.0"
    SummaryID         string                    `json:"summary_id"`
    GeneratedTimestamp int64                    `json:"generated_timestamp"`
    TimeWindow        TimeWindowSpec            `json:"time_window"`
    EpisodeCount      int64                     `json:"episode_count"`
    AverageScores     map[string]float64        `json:"average_scores"`
    TrendMetrics      map[string]float64        `json:"trend_metrics"`      // Gradient / trajectory slopes
    FailureRates      map[string]float64        `json:"failure_rates"`
    ImprovementRates  map[string]float64        `json:"improvement_rates"`  // Delta velocity
    SummaryConfidence float64                   `json:"summary_confidence"`
}
```

Reflection consumes only this versioned contract; `Memory` retains 100% freedom to upgrade backend index structures and storage engines.

---

## 8. Reflection Confidence & Cross-Cognitive Analysis

### 8.1 Independent Reflection Confidence
Every `SpecialistReport` carries a distinct `ReflectionConfidence` (`[0.0, 1.0]`) representing Reflection's epistemic certainty in its own critique. It is never merged or conflated with `ReasoningConfidence` or `CalibratedConfidence`.

### 8.2 Cross-Cognitive & Growth Potential Findings
Housed in `Overall Cognitive Assessment`:
- **Cross-Cognitive Findings:** Identifies inter-ability mismatches (`source_abilities: ["planning", "decision"]`).
- **Growth Potential Estimates:** Scores per-ability learning ROI (`CurrentQualitySignal`, `GrowthPotentialRating`) to guide Learning prioritization.

---

## 9. Stable Versioned `ReflectionReport` Schema & Cardinality Limits

### 9.1 Complete Canonical Schema
```go
type ReflectionReport struct {
    SchemaVersion                string                       `json:"schema_version"` // "2.0.0-FROZEN"
    ReportID                     string                       `json:"report_id"`
    EpisodeID                    string                       `json:"episode_id,omitempty"` // null in periodic mode
    Timestamp                    int64                        `json:"timestamp"`
    Mode                         string                       `json:"mode"` // "MODE_EPISODE" | "MODE_PERIODIC"
    SpecialistReports            []SpecialistReport           `json:"specialist_reports"`
    CrossCognitiveFindings       []CrossCognitiveFinding      `json:"cross_cognitive_findings"`
    GrowthPotentialEstimates     []GrowthPotentialEstimate    `json:"growth_potential_estimates"`
    TrendFindings                []TrendFinding               `json:"trend_findings,omitempty"`
    RecommendedLearningSignals   []RecommendedLearningSignal  `json:"recommended_learning_signals"`
    SessionNotes                 []string                     `json:"session_notes"`
    SelfCalibration              *SelfCalibrationReport       `json:"self_calibration,omitempty"`
}
```

### 9.2 Mandatory Cardinality & Bounding Limits (`Validate()`)
To prevent report bloat and Workspace memory exhaustion over decades, `ReflectionReport.Validate()` enforces strict upper bounds:

| Boundary Parameter | Maximum Allowed Limit | Architectural Rationale |
| :--- | :--- | :--- |
| `MaxSpecialistFindingsPerReport` | **20 findings** per specialist | Prevents runaway verbose heuristic outputs |
| `MaxCrossCognitiveFindings` | **10 findings** per report | Bounds inter-ability interaction summaries |
| `MaxTrendFindings` | **10 findings** per report | Bounds periodic longitudinal motifs |
| `MaxSessionNotes` | **15 items** per report | Bounds transient short-term context notes |
| `MaxRecommendations` | **10 signals** per report | Bounds proposed learning signals |
| `MaxStringLength` | **1,024 bytes** per string field | Prevents unstructured natural language bloat |

Any report exceeding these hard bounds is rejected before publication.

---

## 10. Consumption Contract & Non-Blocking Isolation

1. **One-Directional Consumption:** `Learning` subscribes to `ReflectionReport` envelopes via Workspace to guide rule compilation and strategy retuning. `Decision` may query recent reports. Reflection has zero visibility into downstream utilization.
2. **Infrastructure Isolation (Never a Dependency):** Reflection executes asynchronously after episode completion. If Reflection times out, crashes, or abstains, `Executive` and all primary cognitive abilities continue uninterrupted.

---

## 11. Final Architecture Verification & Permanent Freeze Declaration

### Verification Against Core Audit Questions
- **Does the architecture remain free of responsibility overlap?**  
  **Yes.** Reflection never mutates Memory, never selects actions, never schedules execution DAGs, and never executes parameter learning.
- **Does Reflection remain a pure evaluator?**  
  **Yes.** Its sole external output is the immutable, structured `ReflectionReport` envelope published to Global Workspace.
- **Can Reflection continue improving over the next 20–30 years without redesign?**  
  **Yes.** Via `EvaluationStrategy` abstraction and the three-phase self-improvement loop (`Reflection-on-Reflection` $\to$ `Learning` $\to$ strategy parameter retuning).
- **Does the addition of these refinements introduce any architectural complexity or technical debt?**  
  **No.** Formal contracts (`HistoricalSummary`), explicit enum verdicts (`ABSTAIN`), continuous improvement gradients, trace lineage references, and cardinality bounds eliminate decadal technical debt and prevent interface drift.

### Formal Freeze Declaration
```
APPROVED FOR PERMANENT FREEZE
```

---

## 12. Complete Implementation & Operational Readiness (Phases 1–5 Permanent Freeze)

The `idun/intelligence/reflection` package implements all 11 specialists and three evaluation modes defined by `Version 2.0.0-FROZEN`:

### 12.1 Operational Pipelines
1. **Episode Reflection (`ReflectEpisode`):** Evaluates single completed episodes from read-only Workspace traces, runs specialist evaluators, surfaces cross-cognitive interaction findings and growth potential estimates, validates report bounds, and publishes to Global Workspace (`TopicReflectionReports`).
2. **Periodic Reflection (`ReflectPeriodic`):** Consumes versioned `HistoricalSummary` contracts, performs longitudinal trend analysis (`AnalyzeTrends`), longitudinal cross-cognitive analysis, and long-term growth potential ROI estimation.
3. **Reflection-on-Reflection (`ReflectOnReflection`):** Compares prior `ReflectionReports` against actual historical outcomes (`CompareHistoricalOutcomes`), computes diagnostic self-calibration (`SelfCalibrationReport`), surfaces systematic bias indicators (`OVERESTIMATION_BIAS_...`), and emits recommended learning signals.

### 12.2 Operational Telemetry (`TelemetrySnapshot`)
The service exposes lock-free atomic counters via `Service.Telemetry() TelemetrySnapshot`:
- `TotalEpisodeReflections`, `TotalPeriodicReflections`, `TotalMetaReflections`
- `SpecialistEvaluations`, `TrendAnalyses`, `CrossCognitiveAnalyses`, `GrowthPotentialEstimations`, `SelfCalibrationRuns`
- `ValidationFailures`, `AbstainCount`, `TimeoutCount`, `CancellationCount`
- `AvgEpisodeDurationMs`, `AvgPeriodicDurationMs`, `AvgMetaDurationMs`

All telemetry is content-blind and thread-safe.

### 12.3 Strict Responsibility Boundaries
- **Reflection evaluates.** (Pure read-only evaluator; produces validated diagnostic reports).
- **Learning learns.** (Exclusive owner of model updates, strategy weights, prompts, and symbolic rules).
- **Decision decides.** (Exclusive owner of action selection and utility trade-offs).
- **Planning plans.** (Exclusive owner of HTN task decomposition and execution graphs).
- **Executive coordinates.** (Exclusive owner of lifecycle scheduling, budget assignment, and preemption).
- **Memory owns persistence.** (Exclusive owner of storage engines and historical indexing).
