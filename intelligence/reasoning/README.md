# IDUN Intelligence Pillar: Reasoning Subsystem (`idun/intelligence/reasoning`)

**Architecture Version:** `2.0.0-FROZEN`  
**Classification:** Primary Cognitive Ability Implementation — Phases 1, 2, 3, 4 & 5 Complete  
**Status:** FULLY IMPLEMENTED, PRODUCTION HARDENED & PERMANENTLY FROZEN (`-race` CLEAN)

---

## 1. Overview & Architectural Principles

The `idun/intelligence/reasoning` package implements **CognitiveAbility.Reasoning** under the frozen Version 2.0.0 architecture specification. Reasoning acts as an arbitrated, confidence-gated neuro-symbolic cascade operating over ephemeral session graphs and communicating strictly via Global Workspace envelopes.

### Core Architectural Invariants
1. **Executive Content-Blindness:** `Executive` schedules Reasoning purely from envelope metadata (`cost_tier`, `priority`).
2. **Global Workspace Communication Only:** Reasoning consumes `communication.Envelope` messages and publishes `ReasoningResult` envelopes. It never invokes adjacent cognitive abilities (`Planning`, `Decision`, `Reflection`, `Learning`) directly.
3. **Single-Owner Confidence Principle:**
   - **Reasoning subsystem solely owns `ReasoningConfidence`:** Computed purely from intra-episode evidence (symbolic slots, multi-hop graph weights, analogical case similarity, log-odds fusion).
   - **Calibration subsystem solely owns `CalibratedConfidence`:** Stage $S_7$ (`CalibrationSpecialist`) delegates cross-episode historical trust adjustment to `idun/intelligence/calibration.CalibrationService`.
4. **Session-Scoped Working Graphs:** Any relational graph constructed during evaluation ($S_2$) is strictly bounded by hard spatial limits ($N_{\max} \le 500, E_{\max} \le 2000, D_{\max} \le 3$) and destroyed upon episode completion (`Clear()`).
5. **Structured Learning Compilation (`CompilationCandidate`):** Deliberative inferences expose a structured AST representation allowing `Learning` to compile fast $S_1$ production rules without parsing natural language prose.

---

## 2. Implemented Cascade Stages ($S_0$ – $S_{10}$)

### Stage S0: Context & Strategy Assembly (`ContextAssembler`)
- Queries bounded slices of Memory records ($k \le 20$) without loading full history.

### Stage S1: Symbolic Fast Path (`SymbolicSpecialist`)
- Implements forward-chaining symbolic rule evaluation over structured `SemanticFrame` slots and retrieved Memory records (<2ms). Populates `ReasoningConfidence`.

### Stage S2: Relational Graph Reasoning (`RelationalGraphSpecialist`, `SessionGraph`)
- Constructs an ephemeral, session-scoped working graph in memory (`SessionGraph`) bounded by configurable limits (`MaxGraphNodes`, `MaxGraphEdges`, `MaxGraphDepth`).
- Performs multi-hop relational path discovery (`TraverseBounded`).
- **Mandatory Invariant:** Immediately destroyed (`Clear()`) upon episode completion. Never persisted to Memory or Storage.

### Stage S3: Constraint Consistency Check (`CSPCheckSpecialist`)
- Evaluates candidate hypotheses against retrieved Memory records for logical contradictions (`[]ContradictionFlag`).

### Stage S4: Bayesian Evidence Fusion (`BayesianFusionSpecialist`)
- Combines multiple independent supporting/conflicting evidence sources within the episode and updates `ReasoningConfidence` via log-odds Bayesian updating $[0.01, 0.99]$.

### Stage S5: Case-Based / Analogical Reasoning (`CaseAnalogySpecialist`)
- Retrieves structurally or semantically similar past experiences (`rec.Type == "case" || "episode"`) and computes metric similarity (via `EmbeddingService` or slot matching).

### Stage S6: Multi-Hypothesis Beam Selection (`BeamSelectionSpecialist`)
- Sorts hypotheses by `ReasoningConfidence` descending.
- Selects the primary winner and preserves an intentional ambiguity set of runner-up hypotheses within `ambiguityThreshold` (up to `MaxBeamWidth = 3`). Never collapses close hypotheses prematurely.

### Stage S7: Calibration Integration (`CalibrationSpecialist`)
- Delegates to `idun/intelligence/calibration.CalibrationService`.
- Sole authorized writer of `CalibratedConfidence`. Transforms `ReasoningConfidence` $\to$ `CalibratedConfidence` for primary and beam hypotheses.

### Stage S8: Deliberative LLM Reasoning (`DeliberativeSpecialist`)
- Reuses shared `idun/intelligence/infrastructure/inference.InferenceService`.
- Escalates only when primary `ReasoningConfidence` falls below `spec.EscalationThreshold` (`< 0.65`). Emits structured `ReasoningHypothesis` (`Type: DELIBERATIVE`); never injects unstructured natural language prose into Workspace.

### Stage S9: Constitution Integration (`ConstitutionSpecialist`)
- Submits every completed `ReasoningResult` envelope to `idun/intelligence/constitution.ActionGate` before publication.
- Annotates constitutional approval (`CONSTITUTION_APPROVED`) or vetoes invalid reasoning results (`CONSTITUTION_VETOED`).

### Stage S10: Final Assembly & Workspace Publication (`Service.ReasonEnvelope`)
- Assembles immutable `ReasoningResult` containing primary hypothesis, ambiguity beam, telemetry, and trace logs.
- Publishes structured envelope to Global Workspace (`TopicActiveGoals`).
- Releases all ephemeral resources and guarantees zero state survives beyond the reasoning episode.

---

## 3. Operational Telemetry & Production Hardening

### Operational Telemetry (`TelemetrySnapshot`)
Exposed via `service.GetTelemetry() TelemetrySnapshot` strictly for Host/Kernel monitoring without revealing semantic content:
- Tracks total reasoning episodes, stage hits, deliberative escalations, beam selections, calibration calls, constitution evaluations, timeout/cancellation counts, validation failures, average duration, and average beam width.

### Benchmark Suite Performance (Ryzen 5 7530U)
| Benchmark | ns/op | B/op | allocs/op |
| :--- | :--- | :--- | :--- |
| `BenchmarkSymbolicReasoning` | ~617.5 ns | 424 B | 13 |
| `BenchmarkSessionGraphOperations` | ~1,458 ns | 2,128 B | 26 |
| `BenchmarkBayesianEvidenceFusion` | ~582.7 ns | 432 B | 4 |
| `BenchmarkCaseAnalogy` | ~1,032 ns | 648 B | 13 |
| `BenchmarkBeamSelection` | ~498.5 ns | 968 B | 5 |
| `BenchmarkCalibrationIntegration` | ~174.6 ns | 160 B | 3 |
| `BenchmarkDeliberativeWorkerMocked` | ~890.1 ns | 520 B | 12 |
| `BenchmarkConstitutionIntegrationMocked` | ~186.7 ns | 142 B | 2 |
| `BenchmarkReasonEnvelope` (Full Cascade) | ~3,358 ns | 2,153 B | 32 |

---

## 4. Declaration of Permanent Freeze

```
IDUN V3 Reasoning
Version 2.0.0-FROZEN
APPROVED FOR PERMANENT FREEZE
```
