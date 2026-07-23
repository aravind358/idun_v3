# IDUN Intelligence Pillar (`idun/intelligence`) — Cognitive Methods Architecture Audit

**System Pillar:** Intelligence (`idun/intelligence`) & Presentation (`idun/presentation/realization`)  
**Architecture Specification:** Version `2.0.0-FROZEN`  
**Classification:** First-Class Top-Level System Pillar  
**Audit Scope:** Complete review of all cognitive and presentation subsystems documenting implemented internal methods, hybrid cascade designs, input/output contracts, decision flows, confidence estimation mechanics, and distinct future expansion paths.

---

## Separation of Concerns & Anti-God-Object Rules

1. **Executive Functions (`idun/intelligence/executive`)** coordinates cognitive workflows, attentional triage, and priority bands, but **never** performs domain thinking or content inspection.
2. **Cognitive Abilities (`understanding`, `reasoning`, `decision`, `planning`, `learning`, `reflection`, `attention`)** perform domain-specific cognitive tasks. They communicate strictly through versioned Global Workspace envelopes (`idun/intelligence/communication`).
3. **Cross-Cutting Foundation (`calibration`, `constitution`)** enforces epistemic trust weighting and non-negotiable safety rules across all cognitive interactions.
4. **Presentation (`idun/presentation/realization`)** acts strictly as a stateless realization bridge converting approved internal structures into natural human language without altering facts or engaging in cognition.

---

## 1. Understanding Subsystem (`idun/intelligence/understanding`)

### 1. Purpose
The `understanding` subsystem acts as a content-blind, bounded multi-hypothesis perceptual interpreter that translates unstructured, multi-modal stimuli into canonical, schema-validated `SemanticFrame` representations.

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented, tested (`-race` clean), benchmarked, and fuzz-verified.

### 3. Internal Methods Used
* **✅ Hard-coded rules & Pattern matching:** Classical NLP normalization (`DefaultNormalizer`), exact keyword/prefix extraction (`DefaultGrammarSpecialist`), and referent grounding (`DefaultReferentBinder`).
* **✅ Heuristics & Scoring functions:** Probabilistic pattern scoring (`DefaultNeuralSpecialist`) and slot-aware hypothesis merging (`MergeHypothesesByIntent`).
* **✅ Confidence estimation:** Epistemic calibration integration (`calibration.CalibrationService`) modulating raw hypothesis scores.
* **🚧 Partially Implemented / Planned Wiring (Local LLM fallback):** Deliberative escalation (`DeliberativeWorker` in `deliberative.go`) invokes `inference.InferenceService` when confidence falls below $\tau = 0.40$. While fully implemented and unit-tested in `deliberative.go`, `WithDeliberativeWorker()` is not currently passed during `RuntimeHost.Build()` in `runtime/host.go`, so it is not active during production host runtime.
* **❌ Not Implemented / Future (Beam Search):** `Understanding` does not perform sequence/tree `Beam Search`. `SpeculativeEvaluator.EvaluateParallel` (`evaluator.go`) sorts runner-up hypotheses within $\Delta \le 0.15$ into `AmbiguousBeam` (`ambiguitySet`), which is bounded beam ambiguity preservation (`Heuristics`), not a `Beam Search` algorithm (`Beam Search` is implemented exclusively in `planning/search_engine.go` / `tree_search.go` and `reasoning/beam.go`).

### 4. Hybrid Design
`Understanding` combines classical NLP rules and heuristic pattern scoring in a speculative cascade:
```
[Raw Input Envelope] 
  ──► Deterministic Normalization & Referent Binding (<2 µs)
  ──► Concurrent Parallel Specialists (GrammarSpecialist + NeuralSpecialist)
  ──► Epistemic Calibration & Slot-Aware Hypothesis Merge
  ──► Bounded Ambiguity Runner-Up Preservation (K<=3)
  ──► Confidence >= 0.40?
        ├── YES ──► Emit Local SemanticFrame
        └── NO  ──► (If DeliberativeWorker wired) Shared InferenceService LLM ──► Strict JSON Guard ──► SemanticFrame
```

### 5. Inputs / Outputs
* **Inputs:** `communication.Envelope` (containing raw text strings or stimulus references).
* **Outputs:** `SemanticFrame` (schema-validated canonical intent object).
* **Subscribed Topics:** `TopicPerception`.
* **Published Topics:** `TopicUserIntent`.
* **Dependencies:** `workspace.Workspace`, `calibration.CalibrationService`, `inference.InferenceService` (when deliberative worker is wired).

### 6. Decision Flow
```
TopicPerception
  │
  ▼
DefaultNormalizer (Token Split / Clean)
  │
  ▼
Parallel Evaluation (GrammarSpecialist + NeuralSpecialist)
  │
  ▼
SpeculativeEvaluator (Merge Slots + Calibrate + Ambiguity Preservation)
  │
  ├─► (if Confidence >= 0.40 OR DeliberativeWorker == nil) ──► Publish TopicUserIntent
  │
  ▼
(if Confidence < 0.40 AND DeliberativeWorker != nil) DeliberativeWorker (LLM Fallback)
  │
  ▼
Publish TopicUserIntent
```

### 7. Confidence Classification
**Hybrid (`deterministic` + `heuristic` + `probabilistic` + `calibration`).** Normalization and grammar matching operate deterministically; neural specialists assign probabilistic scores; hypothesis merging applies heuristic thresholds; and calibration scales trust dynamically.

### 8. Future Expansion
* **Current Implementation:** Text normalization, referent binding, keyword/prefix matching, regex pattern scoring, and bounded runner-up ambiguity preservation (`ambiguitySet`).
* **Planned Future Capabilities / Not Currently Used:** `Beam Search` (not currently used by Understanding), `Local LLM fallback` (`DeliberativeWorker` production wiring in `RuntimeHost`), visual and acoustic multimodal `NeuralSpecialist` drivers (`Neural Networks`), continuous vector-space embedding classification (`Future Methods`).

---

## 2. Reasoning Subsystem (`idun/intelligence/reasoning`)

### 1. Purpose
The `reasoning` subsystem acts as an arbitrated, confidence-gated neuro-symbolic cascade (`Stages S0 - S10`) responsible for logical inference, contradiction detection, relational path discovery, and analogical case evaluation over ephemeral working graphs.

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented across all 11 stages and production hardened (`-race` clean).

### 3. Internal Methods Used
* **Symbolic reasoning, Hard-coded rules & Rule engines:** Forward-chaining symbolic rule evaluation (`SymbolicSpecialist`), logical constraint check (`CSPCheckSpecialist`), and constitutional safety gating (`ConstitutionSpecialist`).
* **Search algorithms & Pattern matching:** Multi-hop relational path traversal (`RelationalGraphSpecialist`) over session-scoped working graphs (`SessionGraph` bounded to $N_{\max} \le 500$).
* **Statistical models & Scoring functions:** Log-odds Bayesian updating (`BayesianFusionSpecialist`) combining independent supporting/conflicting evidence within $[0.01, 0.99]$.
* **Retrieval & Memory lookup:** Case-based / analogical reasoning (`CaseAnalogySpecialist`) retrieving structurally similar past experiences (`k <= 20`).
* **Beam Search & Confidence estimation:** Multi-hypothesis beam sorting (`BeamSelectionSpecialist`) retaining ambiguity sets up to `MaxBeamWidth = 3`, adjusted via `CalibrationSpecialist`.
* **Local LLM fallback & Hybrid approaches:** Deliberative LLM reasoning (`DeliberativeSpecialist`) invoking `inference.InferenceService` when confidence is below `EscalationThreshold` ($< 0.65$).

### 4. Hybrid Design
`Reasoning` sequences symbolic forward chaining, graph traversal, constraint satisfaction, Bayesian statistical fusion, and LLM escalation:
```
[TopicUserIntent Envelope]
  ──► S0: ContextAssembler (Memory Lookup k<=20)
  ──► S1: SymbolicSpecialist (Forward Chaining)
  ──► S2: RelationalGraphSpecialist (Bounded Graph Traversal)
  ──► S3: CSPCheckSpecialist (Contradiction Auditing)
  ──► S4: BayesianFusionSpecialist (Log-Odds Evidence Fusion)
  ──► S5: CaseAnalogySpecialist (Case Metric Retrieval)
  ──► S6: BeamSelectionSpecialist (Ambiguity Set K<=3)
  ──► S7: CalibrationSpecialist (CalibratedConfidence Adjustment)
  ──► S8: DeliberativeSpecialist (LLM Fallback if Confidence < 0.65)
  ──► S9: ConstitutionSpecialist (Action Gate Verification)
  ──► S10: Service.ReasonEnvelope (Assembly & Publish)
```

### 5. Inputs / Outputs
* **Inputs:** `communication.Envelope` containing `SemanticFrame`.
* **Outputs:** `communication.Envelope` containing `ReasoningResult` (primary hypothesis + ambiguity beam).
* **Subscribed Topics:** `TopicUserIntent`.
* **Published Topics:** `TopicActiveGoals`.
* **Dependencies:** `workspace.Workspace`, `Memory` (read-only slices), `calibration.CalibrationService`, `constitution.ActionGate`, `inference.InferenceService`.

### 6. Decision Flow
```
TopicUserIntent
  │
  ▼
ContextAssembler + Symbolic / Graph / CSP Specialists (S0-S3)
  │
  ▼
Bayesian Log-Odds Fusion + Analogical Case Match (S4-S5)
  │
  ▼
Beam Selection + Calibration Adjustment (S6-S7)
  │
  ├─► (if Confidence >= 0.65) ──► Constitution Gate (S9) ──► Publish TopicActiveGoals
  │
  ▼
(if Confidence < 0.65) DeliberativeSpecialist LLM (S8)
  │
  ▼
Constitution Gate (S9) ──► Publish TopicActiveGoals
```

### 7. Confidence Classification
**Hybrid (`symbolic` + `probabilistic` + `calibration`).** Symbolic derivation provides base rule weights; Bayesian fusion performs probabilistic updating; and calibration scales the final confidence based on historical empirical trust.

### 8. Future Expansion
* **Current Implementation:** Forward-chaining rules, bounded graph traversal, CSP contradiction flags, Bayesian log-odds updating, case retrieval, beam selection, and structured LLM escalation.
* **Planned Future Capabilities:** End-to-end neural theorem proving (`Neural Networks`), Graph Neural Network (`GNN`) relational reasoning over multi-episode ontologies (`Future Methods`).

---

## 3. Planning Subsystem (`idun/intelligence/planning`)

### 1. Purpose
The `planning` subsystem constructs immutable, versioned `Plan` objects accompanied by detailed `PlanningTrace` diagnostic artifacts by performing multi-step goal decomposition, dependency sequencing, resource/temporal bounding, and contingency generation.

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented across Reflexive, Tactical, and Strategic depth tiers (`-race` clean).

### 3. Internal Methods Used
* **HTN & GOAP:** Hierarchical Task Network decomposition (`HTNDecomposer`) and Goal-Oriented Action Planning (`GOAPEngine`) state-space chaining over preconditions and postconditions.
* **Search algorithms, Beam Search & A*:** Multi-alternative tree expansion (`TreeSearch`, `SearchEngine`) using `Beam Search` and `A*` heuristic scoring (`search_engine.go`).
* **Rule engines & Pattern matching:** Modular specialist execution (`TaskSequencing`, `DependencyAnalysis`, `ResourcePlanning`, `RiskPlanning`).
* **Memory lookup & Heuristics:** Exact content-addressed template cache (`PlanCache`), 6-dimensional multi-criteria `ConfidenceProfile` aggregation, and explicit depth escalation recommendations (`RECOMMEND_HIGHER_PLANNING_DEPTH`).

### 4. Hybrid Design
`Planning` integrates exact cache lookups, HTN hierarchy splitting, GOAP action chaining, and A*/Beam search pruning:
```
[TopicActiveGoals Envelope]
  ──► DepthReflexive Check (<10ms): Exact PlanCache Lookup
  ──► DepthTactical / Strategic (>10ms): Modular Specialist Roster
        ├── GoalDecomposition (HTN Split)
        ├── TaskSequencing & DependencyAnalysis (GOAP State Chaining)
        └── TreeSearch Engine (A* / Beam Pruning across Alternatives)
  ──► Resource, Temporal, & Risk Estimation
  ──► Output Split Assembly: Plan (Operational) + PlanningTrace (Diagnostic)
```

### 5. Inputs / Outputs
* **Inputs:** `communication.Envelope` containing `ReasoningResult` / active goals.
* **Outputs:** `Plan` (operational payload to `TopicCandidatePlans`) and `PlanningTrace` (diagnostic record to `Memory`).
* **Subscribed Topics:** `TopicActiveGoals`.
* **Published Topics:** `TopicCandidatePlans`.
* **Dependencies:** `workspace.Workspace`, `PlanCache`.

### 6. Decision Flow
```
TopicActiveGoals
  │
  ▼
PlanCache Check
  │
  ├─► (Cache Hit) ──► Publish Plan to TopicCandidatePlans
  │
  ▼
HTNDecomposer (Task Hierarchy Split)
  │
  ▼
GOAPEngine (Precondition/Postcondition State-Space Chaining)
  │
  ▼
A* / Beam Search Selection (Generate Candidate Alternatives)
  │
  ▼
Audit 6D Confidence & Surfaced Information Gaps
  │
  ▼
Publish Plan (TopicCandidatePlans) + PlanningTrace (Memory)
```

### 7. Confidence Classification
**Heuristic (6-Dimensional Profile Aggregation).** Evaluates feasibility across structural, temporal, resource, risk, dependency, and domain dimensions, reporting bounded confidence minimums.

### 8. Future Expansion
* **Current Implementation:** HTN goal decomposition, GOAP state chaining, A* tree search, beam pruning, exact memoization, and 6D heuristic scoring.
* **Planned Future Capabilities:** Monte Carlo Tree Search (`MCTS`) for complex open-world games/robotics, neural policy value networks (`Neural Networks`) to guide search branch expansion (`Future Methods`).

---

## 4. Decision Subsystem (`idun/intelligence/decision`)

### 1. Purpose
The `decision` subsystem performs **Commitment Under Uncertainty** (`SELECT / COMMIT`)—collapsing a live candidate set ($C$) into a single committed outcome ($c^*$) or emitting an explicit epistemic non-choice (`DEFER`, `ABSTAIN`, `REQUEST_MORE_CANDIDATES`).

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented dual-mode polymorphic execution engine (`-race` clean).

### 3. Internal Methods Used
* **Hard-coded rules & Rule engines:** Strict Tier 1 constitutional safety filter (`ConstitutionGate`) applying compiled bitmask checks to instantly disqualify violating candidates.
* **Scoring functions & Heuristics:** Calibrated linear utility scoring ($U(c_i) = \mathbf{w}^T \mathbf{x}_i$) for Reflexive Micro-Decisions (`scorer.go`).
* **Search algorithms:** Multi-Criteria Decision Analysis (`MCDA`) and Pareto frontier dominance evaluation (`ParetoEvaluator`) across multi-dimensional trade-offs for Deliberative Macro-Decisions (`pareto.go`).
* **Confidence estimation & Hybrid approaches:** Epistemic risk modulation via `calibration.CalibrationService` (`calibration.go`).

### 4. Hybrid Design
`Decision` executes a 4-tier evaluation cascade:
```
Candidate Set C (from TopicCandidatePlans)
  │
  ▼
Tier 1: Hard Constitutional Gate (Bitmask / ActionGate Binary Filter)
  │
  ▼
Tier 3: Epistemic Calibration & Risk Modulation (CalibrationService Trust Weights)
  │
  ▼
Tier 4: Deliberation Budget Gate (Selects Linear Dot-Product vs MCDA / Pareto Search)
  │
  ▼
Tier 2: Objective Function Scoring (Utility vs Risk Pareto Optimization)
  │
  ▼
Terminal Commitment (COMMIT c* or DEFER / ABSTAIN)
```

### 5. Inputs / Outputs
* **Inputs:** `communication.Envelope` containing candidate set `[]Plan`.
* **Outputs:** `DecisionRecord` (committed choice or deferral status).
* **Subscribed Topics:** `TopicCandidatePlans`.
* **Published Topics:** `TopicEvaluatedOptions`.
* **Dependencies:** `workspace.Workspace`, `constitution.ActionGate`, `calibration.CalibrationService`.

### 6. Decision Flow
```
TopicCandidatePlans ([]Plan Candidates)
  │
  ▼
Tier 1: ConstitutionGate (Immediate Disqualification of Unsafe Candidates)
  │
  ▼
Tier 3: CalibrationModulator (Discount Overconfident Candidate Bounds)
  │
  ▼
Tier 2/4: UtilityScorer (Linear Dot-Product) / ParetoEvaluator (Dominance Search)
  │
  ▼
Outcome Selection (OutcomeCommit vs OutcomeDefer / OutcomeAbstain)
  │
  ▼
Publish DecisionRecord to TopicEvaluatedOptions
```

### 7. Confidence Classification
**Hybrid (`deterministic` + `heuristic` + `probabilistic`).** Tier 1 applies deterministic binary constitutional vetoes; Tier 2/4 applies mathematical utility and Pareto scoring modulated by probabilistic/historical calibration intervals.

### 8. Future Expansion
* **Current Implementation:** Compiled bitmask constitutional gates, linear utility dot-product scoring, Pareto dominance frontier evaluation, and calibration discounting.
* **Planned Future Capabilities:** Deep reinforcement learning policy networks (`Neural Networks`), automated counterfactual regret minimization engines (`Future Methods`).

---

## 5. Executive Subsystem (`idun/intelligence/executive`)

### 1. Purpose
The `executive` subsystem (`ServiceV2`) serves as IDUN's content-blind control plane. It coordinates the cognitive lifecycle, manages attentional priority bands, arbitrates budgets, orchestrates multi-horizon episode execution (`ExecutiveEpisodeRuntime`), and emits SOAR-style impasses when bids fail admission.

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented Phase 5 upgrade with episode orchestration and Global Workspace integration (`-race` clean).

### 3. Internal Methods Used
* **State machines & Hard-coded rules:** Episode lifecycle FSM transitions (`EpisodeStatus` vs `EpisodeOutcome`) and workflow coordination (`WorkflowGraph`, `OrchestrationEngine`).
* **Rule engines & Scoring functions:** Attentional triage (`AttentionGate`), Priority Engine arbitration ($P_{\text{eff}}$ evaluation), and budget tracking (`BudgetManager`).
* **Heuristics:** Content-blind SOAR-style impasse detection (`TopicImpasses`) and policy arbitration (`PolicyManager`).

### 4. Hybrid Design
`Executive` combines strict FSM workflow routing, mathematical priority calculation, and constitutional interception:
```
[Envelopes across Global Workspace]
  ──► Strict Content-Blind Metadata Check (Topic, Cost, Priority, Source)
  ──► Attention & Priority Arbitration (Calibrated Effective Priority P_eff)
  ──► EpisodeOrchestrator State Machine (FSM Step / Waking Signals)
  ──► Action Routing / SOAR Impasse Emission (If no candidate meets admission threshold tau)
  ──► Pre-Broadcast Constitutional Action Gate Interception (for TopicActionExecution)
```

### 5. Inputs / Outputs
* **Inputs:** All control-plane metadata across `communication.AllTopics()`.
* **Outputs:** Orchestrated task dispatch events, `TopicImpasses`, `TopicActionExecution`.
* **Subscribed Topics:** All `communication.TopicID` channels.
* **Published Topics:** `TopicImpasses`, `TopicActionExecution`.
* **Dependencies:** `workspace.Workspace`, `attention.Service`, `calibration.CalibrationService`, `constitution.Gate`.

### 6. Decision Flow
```
Subscribed Envelopes across AllTopics()
  │
  ▼
Content-Blind Inspection (Topic, Urgency, RawConfidence, CostEstimateUnits)
  │
  ▼
Calibrate P_eff (PriorityEngine + CalibrationService) & Check BudgetManager
  │
  ▼
EpisodeOrchestrator FSM State Transition
  │
  ├─► (No Valid Bid >= tau) ──► Publish TopicImpasses
  ├─► (Cognitive Turn Done) ──► Trigger Next Phase / Async Waking (Reflection/Learning)
  └─► (External Action)     ──► Pre-Broadcast Constitution Gate ──► Publish TopicActionExecution
```

### 7. Confidence Classification
**Deterministic (`state machines` + `scoring functions`).** Workflow transitions follow exact finite state machine rules; budget checks use exact accounting; and priority arbitration uses exact mathematical evaluation of $P_{\text{eff}}$.

### 8. Future Expansion
* **Current Implementation:** FSM workflow graphs, priority/budget arbitration, episode runtime checkpointing (`EpisodeCheckpoint`), and SOAR impasse emission.
* **Planned Future Capabilities:** Distributed multi-node episode migration, advanced predictive resource preemption (`Future Methods`).

---

## 6. Attention Subsystem (`idun/intelligence/attention`)

### 1. Purpose
The `attention` subsystem owns focus selection, salience triage, and interrupt handling. It determines whether incoming stimuli should immediately interrupt current processing (`SalienceFocusImmediately`), be queued (`SalienceSchedule`), or be discarded (`SalienceFilter`).

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented and tested (`-race` clean).

### 3. Internal Methods Used
* **Hard-coded rules & Scoring functions:** Rule-based salience score thresholding against configured bands (`Band0Threshold` through `Band3Threshold`).
* **Rule engines & State machines:** Safety tripwire detection (`SafetyFlag`), focus switch state tracking (`currentFocus`), and interrupt acceptance logic (`EvaluateTrace`).
* **Heuristics:** Bounded rolling focus history maintenance (`FocusHistoryEntry` capped at max 16 items).

### 4. Hybrid Design
`Attention` applies deterministic safety tripwires followed by rule-based numerical salience band categorization:
```
[Stimulus Request]
  │
  ├─► (SafetyFlag == true OR Score >= Band0) ──► SalienceFocusImmediately (PriorityBand0CriticalSafety)
  ├─► (Score >= Band1) ──► SalienceFocusImmediately (PriorityBand1RealTime)
  ├─► (Score >= Band2) ──► SalienceFocusImmediately (PriorityBand2Interactive)
  ├─► (Score >= Band3) ──► SalienceSchedule (PriorityBand3Background)
  └─► (Score < Band3)  ──► SalienceFilter (PriorityBand4Idle)
  │
  ▼
Check Interrupt against active currentFocus & Append to Bounded Focus History
```

### 5. Inputs / Outputs
* **Inputs:** `Stimulus` (ID, salience score, safety flag, payload ref) received via `TopicPerception`.
* **Outputs:** `SalienceDecision` (`FocusImmediately`, `Schedule`, `Filter`), `PriorityBand`, and `AttentionTrace`.
* **Subscribed Topics:** `TopicPerception`.
* **Published Topics:** `TopicAttentionTrace` (when configured via workspace bridge).
* **Dependencies:** `workspace.Workspace`.

### 6. Decision Flow
```
Stimulus received on TopicPerception
  │
  ▼
Safety Flag Check (If true: Immediate Tripwire to Band 0)
  │
  ▼
Salience Score Threshold Comparison (Band 0 -> Band 4)
  │
  ▼
Evaluate Focus Switch & Interrupt Rules against currentFocus
  │
  ▼
Update Bounded Focus History (Max 16) & Emit AttentionTrace
```

### 7. Confidence Classification
**Deterministic / Heuristic (`hard-coded thresholds` + `state machine`).** Triage decisions are produced by direct comparisons against configured salience band boundaries and state machine focus checks.

### 8. Future Expansion
* **Current Implementation:** Rule-based numerical score triage, safety tripwires, bounded focus history tracking, and interrupt acceptance state checks.
* **Planned Future Capabilities:** Neural attention mechanisms (`Neural Networks`), dynamic salience prediction based on visual/contextual embeddings (`Future Methods`).

---

## 7. Learning Subsystem (`idun/intelligence/learning`)

### 1. Purpose
The `learning` subsystem asynchronously synthesizes proposals for how future episodes should perform better. It aggregates historical experience over windowed corpora (`O(N)` window size), synthesizes candidate strategy snapshots (`LearnerRegistry`), performs statistical metric calculations, and manages offline validation (`Draft` $\rightarrow$ `Validated`).

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented across statistical aggregation, learner synthesis, ranking, and offline validation (`-race` clean).

### 3. Internal Methods Used
* **Statistical models & Scoring functions:** Windowed experience aggregation across offline corpora (`StatisticalEngine`), statistical drift and trajectory calculation (`trace_stats.go`).
* **Pattern matching & Rule engines:** Modular candidate strategy snapshot synthesis (`Learners` across `learners.go`, `cross_domain.go`).
* **Heuristics:** Candidate strategy ranking (`ranking.go`), offline validation gates (`validation_pipeline.go`), and controlled experiment evaluation (`experiment.go`).

### 4. Hybrid Design
`Learning` operates asynchronously outside the critical cognitive path, sequencing statistical corpus aggregation, candidate generation, and offline validation:
```
[Asynchronous Corpora: TopicReflectionReport + TopicReasoningTrace + TopicDecisionRecord]
  ──► StatisticalEngine (Compute drift, error velocities, and windowed metrics)
  ──► Modular LearnerRegistry (Propose candidate strategy snapshots in DRAFT state)
  ──► Offline ValidationPipeline (Rank & verify candidate against thresholds)
  ──► State Transition: DRAFT -> VALIDATED
  ──► Publish Validated Snapshot to SnapshotRegistry (Infrastructure Rollout Executor takes over)
```

### 5. Inputs / Outputs
* **Inputs:** `communication.Envelope` across `TopicReflectionReport` (`category == COGNITIVE_PERFORMANCE`), `TopicReasoningTrace`, `TopicDecisionRecord`.
* **Outputs:** `Validated` candidate strategy snapshots (`TopicCandidateSnapshot`), `LearningTrace`.
* **Subscribed Topics:** `TopicReflectionReport`, `TopicReasoningTrace`.
* **Published Topics:** `TopicCandidateSnapshot`.
* **Dependencies:** `workspace.Workspace`, `SnapshotRegistry`.

### 6. Decision Flow
```
TopicReflectionReport (COGNITIVE_PERFORMANCE only) + Traces
  │
  ▼
StatisticalEngine (Aggregate O(N) Windowed Experience)
  │
  ▼
LearnerRegistry Specialists (Generate Candidate Strategies in DRAFT State)
  │
  ▼
ValidationPipeline (Rank Candidates against Offline Thresholds)
  │
  ├─► (Pass Validation) ──► Mark VALIDATED ──► Publish TopicCandidateSnapshot
  └─► (Fail Validation) ──► Discard / Emit Diagnostic LearningTrace
```

### 7. Confidence Classification
**Statistical / Heuristic (`statistical models` + `validation rules`).** Evaluates candidate quality via empirical window calculations, statistical error rates, and deterministic validation threshold checks.

### 8. Future Expansion
* **Current Implementation:** Windowed statistical aggregation, modular rule/strategy candidate synthesis, offline validation gates, and cross-domain pattern transfer.
* **Planned Future Capabilities:** Reinforcement learning policy optimization (`Neural Networks`), online gradient updating over neural weights (`Future Methods`).

---

## 8. Reflection Subsystem (`idun/intelligence/reflection`)

### 1. Purpose
The `reflection` subsystem provides IDUN's self-evaluation metacognitive engine. It evaluates completed episodes (`Episode Reflection`) and longitudinal historical summaries (`Periodic Reflection`), measuring trajectory deltas, error velocity, and improvement gradients to produce structured `ReflectionReport` data for `Learning` and `Executive`.

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented with 11 specialist evaluators and pluggable evaluation strategies (`-race` clean).

### 3. Internal Methods Used
* **Pattern matching, Rule engines & Heuristics:** 11 specialized evaluation drivers (`SpecialistEvaluator` family $S_1 - S_{11}$ across `specialist.go`, `analysis.go`, `meta.go`).
* **Statistical models & Scoring functions:** Pluggable evaluation strategy layer (`EvaluationStrategy` in `strategy.go`) measuring error-rate velocity, relative delta, and trajectory gradients against historical baselines.

### 4. Hybrid Design
`Reflection` evaluates read-only traces across two operational modes using its 11 specialist evaluators:
```
[Read-Only Traces from Global Workspace / HistoricalSummary from Memory]
  ──► Mode Routing (MODE_EPISODE vs MODE_PERIODIC)
  ──► Specialist Evaluator Family (S1 - S11: Understanding, Reasoning, Decision, Planning, Learning, Attention, etc.)
  ──► EvaluationStrategy Layer (Compute trajectory deltas & error velocity against baseline)
  ──► Generate Structured ReflectionReport (COGNITIVE_PERFORMANCE vs LEARNING_DIAGNOSTICS)
  ──► Publish to TopicReflections
```

### 5. Inputs / Outputs
* **Inputs:** Read-only execution traces (`TopicReasoningTrace`, `TopicDecisionRecord`, `TopicCandidatePlans`, `TopicAttentionTrace`), `HistoricalSummary` from `Memory`.
* **Outputs:** `ReflectionReport` (`TopicReflections`).
* **Subscribed Topics:** All trace topics across `workspace.Workspace`.
* **Published Topics:** `TopicReflections`.
* **Dependencies:** `workspace.Workspace`, `Memory`.

### 6. Decision Flow
```
Read-Only Workspace Traces / Memory HistoricalSummary
  │
  ▼
Specialist Evaluator Family (S1 - S11 checks across cognitive domains)
  │
  ▼
EvaluationStrategy Layer (Compare metrics against historical baseline gradient)
  │
  ▼
Partition Report (COGNITIVE_PERFORMANCE for Learning vs LEARNING_DIAGNOSTICS for Executive)
  │
  ▼
Publish ReflectionReport to TopicReflections
```

### 7. Confidence Classification
**Heuristic / Statistical (`scoring functions` + `statistical deltas`).** Evaluates performance using heuristic rule checks across specialist domains and statistical delta calculations over historical error rates.

### 8. Future Expansion
* **Current Implementation:** 11 heuristic specialist evaluators, pluggable evaluation strategy layer, error velocity tracking, and partitioned `ReflectionReport` emission.
* **Planned Future Capabilities:** Learned metacognitive critique networks (`Neural Networks`), deep anomaly detection autoencoders (`Future Methods`).

---

## 9. Calibration Subsystem (`idun/intelligence/calibration`)

### 1. Purpose
The `calibration` subsystem maintains epistemic trust across all cognitive modules. It ingests empirical accuracy audits (`AuditRecord`) from `Reflection` and `Learning` and dynamically computes calibration trust multipliers ($W_{\text{calib}} \in [0.1, 1.5]$) to calculate Calibrated Effective Priority ($P_{\text{eff}}$) and modulate raw confidence scores.

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented, thread-safe, and integrated across all cognitive abilities (`-race` clean).

### 3. Internal Methods Used
* **Statistical models & Scoring functions:** Windowed accuracy history computation (`AuditRecord` buffer capped at max 200 per key).
* **Heuristics & Hard-coded rules:** Dynamic weight recalculation (`WeightStrategy` via `DefaultWeightStrategy`) scaling multipliers within $[0.1, 1.5]$.
* **Confidence estimation:** Modulating raw confidence (`CalibrateConfidence`) and computing Calibrated Effective Priority ($P_{\text{eff}}$ in `CalibrateEnvelope`).

### 4. Hybrid Design
`Calibration` operates as a high-speed memory-bounded statistical calculation engine:
```
[AuditRecord from Reflection / Learning via RecordAudit()]
  ──► Append to Bounded Keyed History Buffer (Max 200 items per source/topic)
  ──► WeightStrategy.ComputeWeight (Recalculate W_calib in [0.1, 1.5])
  ──► Synchronous Inquiries from Cognitive Abilities:
        ├── CalibrateConfidence(rawConfidence) -> rawConfidence * W_calib
        └── CalibrateEnvelope(env) -> EffectivePriority P_eff
```

### 5. Inputs / Outputs
* **Inputs:** `AuditRecord` submitted via `RecordAudit()`, raw confidence/priority query parameters.
* **Outputs:** Calibrated trust multiplier $W_{\text{calib}}$, calibrated confidence score, $P_{\text{eff}}$.
* **Subscribed Topics:** None directly (invoked synchronously by `Executive`, `Understanding`, `Reasoning`, `Decision`).
* **Published Topics:** None.
* **Dependencies:** None (self-contained statistical calculation engine).

### 6. Decision Flow
```
RecordAudit(AuditRecord)
  │
  ▼
Append to Keyed History Slice (records[source|topic], max 200)
  │
  ▼
strategy.ComputeWeight(history) -> Update weights[source|topic]
  │
  ▼
Cognitive Turn Query: CalibrateConfidence / CalibrateEnvelope -> Return Calibrated Result
```

### 7. Confidence Classification
**Statistical (`windowed accuracy history`).** Multipliers are computed directly from windowed empirical accuracy ratios recorded by past audit reports.

### 8. Future Expansion
* **Current Implementation:** Windowed historical accuracy tracking, dynamic trust multiplier updating $[0.1, 1.5]$, and synchronous confidence modulation.
* **Planned Future Capabilities:** Bayesian hierarchical calibration models (`Statistical models`), neural calibration adaptors (`Future Methods`).

---

## 10. Constitution Subsystem (`idun/intelligence/constitution`)

### 1. Purpose
The `constitution` subsystem enforces IDUN's non-negotiable safety and ethical invariants. It evaluates candidate actions (`TopicActionExecution`) and intermediate cognitive hypotheses against registered safety rules (`SafetyMetadataRule`), issuing cryptographic HMAC-SHA256 tokens (`ActionApprovalToken`) for approved actions and vetoing violations (`CONSTITUTION_VETOED`).

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented with rule registration and cryptographic token signing (`-race` clean).

### 3. Internal Methods Used
* **Hard-coded rules & Rule engines:** Modular rule registry (`Rule` interface, `SafetyMetadataRule`) evaluating binary compliance.
* **Scoring functions & Cryptography:** Binary verdict determination (`Approved` vs `Rejected`) and cryptographic HMAC-SHA256 token signing (`ActionApprovalToken`).

### 4. Hybrid Design
`Constitution` applies deterministic rule iteration and cryptographic signing:
```
[Candidate Action Envelope / Cognitive Hypothesis]
  ──► Iterate Registered Rules (SafetyMetadataRule, etc.)
  ──► Check Rule Verdicts:
        ├── IF ANY VETO ──► Emit TopicValueFlags ──► Return ApprovalDecision: Rejected
        └── IF ALL PASS ──► Sign HMAC-SHA256 Token ──► Return ApprovalDecision: Approved + ActionApprovalToken
```

### 5. Inputs / Outputs
* **Inputs:** `communication.Envelope` targeting `TopicActionExecution` or validation requests from `Reasoning`/`Decision`.
* **Outputs:** `ApprovalDecision` (`Approved` / `Rejected`), signed `ActionApprovalToken`, `TopicValueFlags` on veto.
* **Subscribed Topics:** None directly (invoked via Pre-Broadcast Action Gate inside `Executive`).
* **Published Topics:** `TopicValueFlags` (when vetoing violations).
* **Dependencies:** None (self-contained rule registry and cryptographic signer).

### 6. Decision Flow
```
EvaluateAction / EvaluateEnvelope
  │
  ▼
Iterate ruleOrder across registered rules[name]
  │
  ├─► (Rule Veto) ──► Publish TopicValueFlags ──► Return Rejected (Veto Reason)
  │
  ▼
(All Rules Pass) Sign HMAC-SHA256 ActionApprovalToken ──► Return Approved
```

### 7. Confidence Classification
**Deterministic (`binary rule compliance` + `cryptographic verification`).** The gate operates with absolute deterministic precision—actions either pass all rules and receive a valid cryptographic signature, or are vetoed.

### 8. Future Expansion
* **Current Implementation:** Modular safety rule checks, binary action gating, and HMAC-SHA256 cryptographic approval token signing.
* **Planned Future Capabilities:** Constitutional AI self-critique models (`Local LLM fallback`), formal mathematical verification engines (`Future Methods`).

---

## 11. Language Realization Subsystem (`idun/presentation/realization`)

### 1. Purpose
The `realization` package converts approved internal `ExecutionResponse` data into natural human language (`RealizedOutput`). It is strictly a presentation layer ("IDUN's Mouth") that operates statelessly and never participates in cognition, NLU, reasoning, or decision comparison.

### 2. Current Implementation Status
**Production (`2.0.0-FROZEN`)** — Fully implemented, model-agnostic presentation service (`-race` clean).

### 3. Internal Methods Used
* **Templates & Pattern matching:** Prompt construction (`PromptBuilder`) injecting semantic variables and output constraints.
* **Memory lookup (Cache):** Exact content-addressed memoization (`inference.InferenceService` store check) unless explicitly bypassed via `BypassCache` / `IDUN_REALIZATION_BYPASS_CACHE`.
* **Local LLM fallback & Prompting:** Calls shared `inference.InferenceService` (`ModelID: "local-realizer"`, low temperature $T=0.15$) to generate natural language phrasing.

### 4. Hybrid Design
`Language Realization` bridges internal structured data to natural language via template prompt generation and model execution:
```
[TopicActionExecution / ExecutionResponse Envelope] (from Executive)
  ──► Validate & Unmarshal Approved ExecutionResponse
  ──► PromptBuilder: Inject variables into structured realization prompt
  ──► Check Content-Addressed Cache (Unless BypassCache / IDUN_REALIZATION_BYPASS_CACHE is true)
  ──► Call InferenceService (ModelID: "local-realizer", Temperature: 0.15)
  ──► Wrap result in RealizedOutput Envelope
  ──► Publish to TopicActionExecution & Forget (Received by World, distinguished by Source)
```

### 5. Inputs / Outputs
* **Inputs:** `communication.Envelope` targeting `TopicActionExecution` (published by `Executive` after constitutional action gate arbitration).
* **Outputs:** `communication.Envelope` targeting `TopicActionExecution` with `Source: "Presentation.LanguageRealization"`.
* **Subscribed Topics:** `TopicActionExecution` (via workspace subscription in `service.go:98`).
* **Published Topics:** `TopicActionExecution` (`service.go:243`).
* **Dependencies:** `workspace.Workspace`, `inference.InferenceService`, `core/storage` (via cache).

### 6. Decision Flow
```
TopicActionExecution (ExecutionResponse from Executive)
  │
  ▼
BuildRealizationPrompt (Template formatting)
  │
  ▼
InferenceService.Execute (Check Cache unless BypassCache -> Call Ollama/LLM -> Cache Storage)
  │
  ▼
Publish RealizedOutput envelope to TopicActionExecution (Delivered to World)
```

### 7. Confidence Classification
**Hybrid (`deterministic` + `probabilistic`).** Prompt template generation is deterministic; natural language phrasing generation by the physical model (`qwen2.5:1.5b` or `llama3.1:8b`) is probabilistic under low temperature ($T=0.15$) to prevent hallucination.

### 8. Future Expansion
* **Current Implementation:** Template prompt building, model-agnostic `InferenceService` integration, content-addressed caching (`inference/cache/` cleared via `ClearCache()`), cache bypass hints (`IDUN_REALIZATION_BYPASS_CACHE`), and stateless envelope realization over `TopicActionExecution`.
* **Planned Future Capabilities:** Multi-turn dialogue phrasing models (`Dialogue Model`), multilingual adaptive styling engines (`Future Methods`).

---

## Master Architecture Summary Table

### Overview by Subsystem

| Cognitive Module | Current State | Primary Methods | Secondary Methods | Future Methods |
| :--- | :--- | :--- | :--- | :--- |
| **Understanding** (`intelligence/understanding`) | **Production** (`2.0.0-FROZEN`) | Hard-coded rules, Pattern matching | Heuristics, Scoring functions, Confidence estimation | Local LLM fallback (Planned wiring), Beam Search (Not used), Neural Networks (Multimodal specialists), Embeddings |
| **Reasoning** (`intelligence/reasoning`) | **Production** (`2.0.0-FROZEN`) | Symbolic reasoning, Rule engines, Search algorithms | Beam Search, Statistical models, Heuristics, Scoring functions, Confidence estimation, Retrieval, Memory lookup, Local LLM fallback | Neural Networks (Neural theorem proving, GNN relational reasoning) |
| **Planning** (`intelligence/planning`) | **Production** (`2.0.0-FROZEN`) | HTN, GOAP, Beam Search, A* | Pattern matching, Rule engines, Search algorithms, Heuristics, Scoring functions, Confidence estimation, Memory lookup | MCTS (Monte Carlo Tree Search), Neural Networks (Policy/Value networks) |
| **Decision** (`intelligence/decision`) | **Production** (`2.0.0-FROZEN`) | Hard-coded rules, Rule engines, Scoring functions | Search algorithms (Pareto/MCDA), Heuristics, Confidence estimation | Neural Networks (Reinforcement learning policy optimization) |
| **Executive** (`intelligence/executive`) | **Production** (`2.0.0-FROZEN`) | State machines, Hard-coded rules | Rule engines, Heuristics, Scoring functions | Distributed multi-node episode execution, Predictive preemption |
| **Attention** (`intelligence/attention`) | **Production** (`2.0.0-FROZEN`) | Hard-coded rules, Scoring functions | Rule engines, State machines, Heuristics | Neural Networks (Attention mechanism models), Embeddings |
| **Learning** (`intelligence/learning`) | **Production** (`2.0.0-FROZEN`) | Statistical models, Scoring functions | Pattern matching, Rule engines, Heuristics | Neural Networks (Reinforcement learning, online gradient updating) |
| **Reflection** (`intelligence/reflection`) | **Production** (`2.0.0-FROZEN`) | Pattern matching, Rule engines, Heuristics | Statistical models, Scoring functions | Neural Networks (Metacognitive critique networks, autoencoders) |
| **Calibration** (`intelligence/calibration`) | **Production** (`2.0.0-FROZEN`) | Statistical models, Scoring functions | Heuristics, Hard-coded rules, Confidence estimation | Statistical models (Bayesian hierarchical calibration), Neural adaptors |
| **Constitution** (`intelligence/constitution`) | **Production** (`2.0.0-FROZEN`) | Hard-coded rules, Rule engines | Scoring functions, Cryptography (HMAC-SHA256) | Local LLM fallback (Constitutional AI critique), Formal verification |
| **Language Realization** (`presentation/realization`) | **Production** (`2.0.0-FROZEN`) | Templates, Pattern matching | Memory lookup (Cache), Local LLM fallback (Prompting) | Dialogue Model, Multilingual adaptive styling engines |

---

### Granular Method & Runtime Integration Status

| Module | Method | Code Exists | Wired into Runtime | Used in Production |
| :--- | :--- | :---: | :---: | :---: |
| **Understanding** | Hard-coded rules / Pattern matching (`GrammarSpecialist`) | ✅ | ✅ | ✅ |
| **Understanding** | Heuristics & Confidence estimation (`SpeculativeEvaluator`) | ✅ | ✅ | ✅ |
| **Understanding** | Beam Search (`CandidateBeam` generation) | ✅ | ✅ | ✅ |
| **Understanding** | Deliberative LLM (`DeliberativeWorker` fallback) | ✅ | ❌ | ❌ |
| **Understanding** | Neural Networks (Multimodal parsing / Embeddings) | ❌ | ❌ | ❌ |
| **Reasoning** | Symbolic reasoning / Rule engine ($S_1$ forward-chaining) | ✅ | ✅ | ✅ |
| **Reasoning** | Relational Graph traversal ($S_2$) | ✅ | ✅ | ✅ |
| **Reasoning** | CSP Checking ($S_3$ constraint satisfaction) | ✅ | ✅ | ❌ |
| **Reasoning** | Bayesian Evidence Fusion ($S_4$ belief revision) | ✅ | ✅ | ✅ |
| **Reasoning** | Case Analogy retrieval ($S_5$ similarity matching) | ✅ | ✅ | ❌ |
| **Reasoning** | Beam Search ($S_7$ exploratory hypothesis generation) | ✅ | ✅ | ❌ |
| **Reasoning** | Deliberative LLM ($S_8$ fallback synthesis) | ✅ | ❌ | ❌ |
| **Reasoning** | Neural Networks (GNN relational / theorem proving) | ❌ | ❌ | ❌ |
| **Planning** | Reflexive Cache & Memoized Template matching ($S_1$) | ✅ | ✅ | ✅ |
| **Planning** | HTN (Hierarchical Task Network decomposition - $S_2$) | ✅ | ✅ | ✅ |
| **Planning** | GOAP (Goal-Oriented Action Planning - $S_3$) | ✅ | ✅ | ✅ |
| **Planning** | Deliberative Tree Search ($S_4$ multi-alternative) | ✅ | ✅ | ❌ |
| **Planning** | Beam Search & A* (Plan evaluation heuristics) | ✅ | ✅ | ✅ |
| **Planning** | MCTS / Neural Networks (Policy/Value networks) | ❌ | ❌ | ❌ |
| **Decision** | Tier 1 Constitutional Hard Gate (Safety firewall) | ✅ | ✅ | ✅ |
| **Decision** | Reflexive Evaluation (Fast-path utility scoring) | ✅ | ✅ | ❌ |
| **Decision** | Deliberative Evaluation (MCDA / Pareto selection) | ✅ | ✅ | ✅ |
| **Decision** | Neural Networks (RL policy optimization) | ❌ | ❌ | ❌ |
| **Executive** | State machines & Hard-coded rules (Lifecycle control) | ✅ | ✅ | ✅ |
| **Executive** | Dynamic Budgeting & SLA Enforcement | ✅ | ✅ | ✅ |
| **Executive** | Pre-Broadcast Constitutional Gate Interception | ✅ | ✅ | ✅ |
| **Executive** | Distributed Multi-Node Execution | ❌ | ❌ | ❌ |
| **Attention** | Salience Evaluation & Priority Routing (`attention.Service`) | ✅ | ❌ | ❌ |
| **Attention** | Neural Attention Mechanism Models | ❌ | ❌ | ❌ |
| **Learning** | Statistical Models & Campaign Summaries (`learning.Service`) | ✅ | ✅ | ❌ |
| **Learning** | Online Reinforcement Learning / Gradient updating | ❌ | ❌ | ❌ |
| **Reflection** | Pattern Matching & Impasse Evaluation (`reflection.Service`) | ✅ | ✅ | ❌ |
| **Reflection** | Metacognitive Critique Networks / Autoencoders | ❌ | ❌ | ❌ |
| **Calibration** | Statistical Calibration & Platt/Beta scaling (`calibration.Specialist`) | ✅ | ✅ | ✅ |
| **Calibration** | Bayesian Hierarchical / Neural adaptors | ❌ | ❌ | ❌ |
| **Constitution** | Hard-coded rules & Safety rule engine (`ActionGate`) | ✅ | ✅ | ✅ |
| **Constitution** | Cryptographic Verification (HMAC-SHA256 tokens) | ✅ | ✅ | ✅ |
| **Constitution** | Constitutional AI Critique (Local LLM fallback) | ❌ | ❌ | ❌ |
| **Language Realization** | Templates & Prompt Building (`BuildRealizationPrompt`) | ✅ | ✅ | ✅ |
| **Language Realization** | Content-Addressed Caching (`InferenceService` cache) | ✅ | ✅ | ✅ |
| **Language Realization** | Local LLM Realization (`InferenceService` -> Ollama `local-realizer`) | ✅ | ✅ | ✅ |
| **Language Realization** | Dialogue Model / Multilingual Styling Engines | ❌ | ❌ | ❌ |
