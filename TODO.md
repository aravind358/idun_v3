# IDUN V3 Master Development Roadmap (`TODO.md`)

This document serves as the authoritative master development roadmap, technical debt ledger, unfinished runtime wiring manifest, and architectural refinement tracker for IDUN V3 (`2.0.0-FROZEN`).

> [!IMPORTANT]
> **Single Source of Truth & Zero-Memory Rule**: All future development, planned improvements, technical debt, and architectural refactoring must be tracked here so that future development does not rely on memory or hidden assumptions.
> When a TODO is completed, remove it from this document and record the implementation details in the appropriate package `README.md` or project `CHANGELOG.md`.

---

## Engineering Assessment: Runtime Investigation Verdict

```
Runtime Investigation Verdict
✅ Root cause analysis completed.
✅ Cache investigation completed.
✅ Runtime repair completed.
⚠️ Language Realization contains temporary conversational policy.
📋 Future improvement: Move conversational content generation into the cognitive layer while keeping Language Realization responsible only for natural-language expression.
```

## 0. Future — Dangerous Operations Mode
- **Priority**: Future Backlog
- **Category**: Testing / Acceptance
- **Background**: Commands like `shutdown`, `restart`, `sleep`, or `delete system files` should not execute during normal acceptance testing. 
- **Goal**: Add a dedicated Dangerous Operations Test Mode that:
  1. Requires explicit opt-in.
  2. Confirms the action before execution.
  3. Can simulate execution (dry-run) or perform the real operation.
  4. Is excluded from routine acceptance runs.
- **Benefit**: Verify dangerous capabilities safely without risking the development machine.

---

## 1. Architecture & System-Wide Pipelines

### Introduce Structured Semantic Response Object
- **Priority**: High
- **Status**: Planned
- **Category**: Architecture, Presentation, Decision, Understanding
- **Background**: The Executive Output Audit revealed that the cognitive system currently loses semantic information before reaching Language Realization.
  Current flow:
  ```
  Understanding
          ↓
  Reasoning
          ↓
  Planning
          ↓
  Decision
          ↓
  ExecutionResponse
          ↓
  Language Realization
  ```
  Current `ExecutionResponse` contains only: `ResponseID`, `ParentRef`, `FinalizedContent`, `Tone`, and `Language`.
  Most semantic information is flattened into a single string (`FinalizedContent`), causing Language Realization to reconstruct meaning using prompts instead of receiving structured semantic data. This creates unnecessary prompt complexity and prevents deterministic response generation.
- **Goal**: Introduce a structured semantic response object that preserves cognitive meaning through the pipeline until Language Realization. Executive should remain completely content-blind and continue forwarding payloads unchanged. Language Realization should receive structured semantic information rather than reconstructing intent from text.
- **Current implementation**: `Decision` flattens cognitive candidate descriptions into a raw string inside `ExecutionResponse.FinalizedContent`. `Language Realization` parses this text using LLM prompt instructions (`prompt.go`).
- **Planned architecture**: Upstream cognitive modules (`Understanding`, `Reasoning`, `Decision`) produce and propagate a structured semantic object across the Global Workspace. `Executive` forwards `PayloadRef` content-blindly. `Language Realization` unmarshals the structured semantic response and renders natural human speech without inferring intent from flat text.
- **Detailed TODO Breakdown**:
  - **SR-001**: Audit current semantic information produced by Understanding (`SemanticFrame`, `Intent`, `Entities`, `Slots`, `Confidence`, `Context`, `Metadata`). Document which fields are already available. (**Status**: Planned)
  - **SR-002**: Audit semantic information produced by Reasoning (`Goal`, `Meaning`, `Explanation`, `Constraints`, `Candidate responses`, `Internal reasoning artifacts`). Determine which should survive downstream. (**Status**: Planned)
  - **SR-003**: Audit Decision output. Determine exactly where semantic information is flattened into `FinalizedContent`. Document required modifications without implementation. (**Status**: Planned)
  - **SR-004**: Design the new semantic response object. This should contain only information required by downstream presentation (e.g., `semantic type`, `semantic meaning`, `intent`, `slots/entities`, `tone`, `language`, `confidence`). Do not finalize the schema yet; this task is investigation and design only. (**Status**: Planned)
  - **SR-005**: Update Language Realization to consume structured semantic information instead of reconstructing meaning from `FinalizedContent`. This is a future implementation task. (**Status**: Planned)
  - **SR-006**: Eventually replace prompt-based conversational policies with deterministic semantic rendering. Prompt instructions should become minimal. Language Realization should become primarily a realization layer rather than a reasoning layer. (**Status**: Technical Debt)
- **Notes**: Do not implement code changes until subtasks SR-001 through SR-004 (investigation and design) are completed and reviewed.

---

## 1. Runtime

### Wire Deliberative LLM Stages (`Understanding` & `Reasoning`) into Runtime
- **Priority**: High
- **Status**: Unfinished Runtime Work
- **Reason**: Both `understanding.DeliberativeWorker` and `reasoning.DeliberativeSpecialist` (`StageS8DeliberativeLLM`) are fully implemented in code (`deliberative.go`), but `InferenceService` (`inference.Service`) is not injected into these services during `RuntimeHost.Build()` (`runtime/host.go`).
- **Current implementation**: `runtime/host.go` instantiates `Understanding` and `Reasoning` without passing `WithInferenceService(infSvc)`. Consequently, when speculative NLU or deterministic reasoning falls below confidence thresholds, these modules cannot escalate to local LLM fallbacks during production turns.
- **Desired implementation**: Pass `InferenceService` to `understanding.NewService` and `reasoning.NewService` in `RuntimeHost.Build()`, ensuring `DeliberativeWorker` and `DeliberativeSpecialist` are active and ready for fallback escalation.
- **Notes**: Ensure appropriate SLA timeouts and circuit breakers are configured so deliberative escalation does not violate overall turn latency budgets.

### Wire Attention Subsystem into Production Runtime
- **Priority**: Medium
- **Status**: Unfinished Runtime Work
- **Reason**: The `attention` package (`idun/intelligence/attention`) is fully implemented with salience scoring (`service.go`) and workspace subscription adapters (`workspace_bridge.go`), but `RuntimeHost.Build()` does not instantiate or connect it.
- **Current implementation**: `runtime/host.go` skips `attention.Service` initialization entirely.
- **Desired implementation**: Instantiate `attention.NewService()` in `runtime/host.go` and subscribe it to `TopicCandidatePlans` and `TopicEvaluatedOptions` to actively regulate working memory salience and prioritize high-urgency envelopes.
- **Notes**: Verify `-race` clean execution under high concurrency once wired into the Global Workspace bus.

---

## 2. Understanding

### Evaluate Understanding for Structured Semantic Output
- **Priority**: Medium
- **Status**: Investigation
- **Category**: Understanding
- **Description**: Determine whether Understanding already produces enough structured semantic information to support the future Semantic Response architecture. Specifically investigate whether the following already exist internally: `Intent`, `Entities`, `Slots`, `Confidence`, `Context`, `SemanticFrame`, `Metadata`.
- **Current implementation**: `SpeculativeEvaluator` and `GrammarSpecialist` (`speculative.go`) parse user text into `SemanticFrame` structures containing `Intent`, `Slots`, and confidence scores during initial perception.
- **Investigation**: Identify what already exists inside `SemanticFrame` and perception structures, what is missing, what requires only wiring, and what requires new implementation. The goal is to maximize reuse of the existing Understanding architecture rather than redesigning it. Do not modify code. Do not implement any new features. Produce recommendations only.
- [ ] Add rigorous benchmark suite for new Neural Classifier model (`understanding/v3/neural.go`)
- [ ] Implement `EntityTask` semantic mapping for future reminder/task integration
- [ ] **app-calc-1**: Extend `app-calc-1` to support complex arithmetic expressions while preserving the existing semantic contract. (Future capability enhancement, not a Phase 4 restoration defect)
- [ ] **app-reminder-1**: Extend `sys-native-1` scheduler to support non-RFC3339 temporal formats if needed, though temporal normalization should ideally handle all conversions.
- **Planned architecture**: Upstream `Understanding` outputs cleanly structured semantic attributes (`Intent`, `Entities`, `Slots`, `Context`) that flow through `Reasoning` and `Decision` into the eventual Semantic Response object without data loss.
- **Future implementation**: Connect validated `Understanding` semantic outputs to downstream structures once investigation recommendations are approved.
- **Notes**: Maintain strict documentation rules: clearly distinguish between what currently exists (`SemanticFrame`), what is under investigation, and what is planned for future implementation. Do not describe planned architecture as if it already exists.

### Local LLM Fallback (`DeliberativeWorker`) Escalation Pipeline
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: When user input contains complex idioms, typos, or out-of-grammar phrasing, deterministic pattern matching (`GrammarSpecialist`) yields low confidence or unparsed slots.
- **Current implementation**: `SpeculativeEvaluator` runs parallel grammar and neural heuristic scoring. If unparsed, the system lacks active runtime connection to `DeliberativeWorker`.
- **Desired implementation**: Once `InferenceService` is wired into `Understanding`, verify that `DeliberativeWorker` (`ModelID: "nlu-deliberative-parser"`) automatically intercepts low-confidence frames, extracts structured slots/intents, and outputs validated `SemanticFrame` envelopes.
- **Notes**: Must enforce strict schema validation so `DeliberativeWorker` never emits malformed JSON or free-form text.

### Multimodal Perception Specialists & Continuous Embeddings
- **Priority**: Low
- **Status**: Future Work
- **Reason**: `Understanding` currently operates exclusively on text token streams (`world/adapters/text`). Future iterations require processing audio, visual, and structured sensor streams alongside continuous vector similarity.
- **Current implementation**: Text-only `TextInputAdapter` and `SemanticFrame` slot extraction.
- **Desired implementation**: Implement multimodal perception specialists that extract unified `SemanticFrame` structures from audio/visual inputs, and integrate continuous vector embedding lookups for semantic entity resolution.
- **Notes**: Requires multimodal backend support in `idun/intelligence/infrastructure/inference` and `registry`.

### Deep Intent DAG Construction
- **Priority**: Medium
- **Status**: Planned Improvement
- **Category**: Understanding
- **Description**: Enhance speculative.go to construct deep Intent DAGs during simultaneous slot matching, rather than mapping the primary hypothesis to an isolated root node.
- **Reason**: Full structural parsing for complex, multi-action conversational inputs requires linked intent graphs rather than flattened lists.
- **Affected Component**: `intelligence/understanding/evaluator.go`, `intelligence/understanding/speculative.go`
- **Recommended Future Phase**: Phase 2C / 3A

### Deliberative LLM Ambiguity and Assumption Extraction
- **Priority**: Low
- **Status**: Planned Improvement
- **Category**: Understanding
- **Description**: Expand the Deliberative LLM parsing prompt to naturally extract and populate the `Ambiguity` and `Assumption` arrays in the `SemanticFrame`.
- **Reason**: While deterministic specialists struggle with open-ended ambiguity detection, deliberative LLM workers inherently detect ambiguities and make assumptions; these should be serialized into the Phase 2B models rather than flattened.
- **Affected Component**: `intelligence/understanding/deliberative.go`
- **Recommended Future Phase**: Phase 2C

### Implement Deferred Normalizers
- **Priority**: Medium
- **Status**: Future Work
- **Category**: Understanding
- **Description**: Implement `number.go` and `unit.go` normalizers.
- **Reason**: These were deferred during Phase 4B.4 (Temporal Processing) to maintain focus on temporal objects, but will be needed for comprehensive semantic object normalization.
- **Affected Component**: `intelligence/understanding/v3/normalizers/`
- **Recommended Future Phase**: Phase 4B.5

---

## 3. Reasoning

### Activate CSP Checking (`S3`) & Case Analogy (`S5`) in Production Profiles
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: `CSPCheckSpecialist` (`S3`) and `CaseAnalogySpecialist` (`S5`) are implemented and wired into `Registry`, but are disabled by default (`IsStageEnabled` returns `false` under default `StrategySpec` / `DefaultConfig`).
- **Current implementation**: Production reasoning relies on `SymbolicSpecialist` ($S_1$), `RelationalGraphSpecialist` ($S_2$), and `BayesianFusionSpecialist` ($S_4$).
- **Desired implementation**: Conduct performance and latency benchmarking across domain profiles. Enable `StageS3CSPChecking` for constraint-heavy planning tasks and `StageS5CaseAnalogy` for historical case matching where appropriate.
- **Notes**: Ensure exact timeout bounding so adding $S_3$ and $S_5$ does not breach the 50ms reflexive reasoning budget.

### Neural Relational Reasoning & Theorem Proving
- **Priority**: Low
- **Status**: Future Research
- **Reason**: Deterministic symbolic forward-chaining and relational graph traversals excel at explicit rules but struggle with dense, implicit multi-hop relational inference.
- **Current implementation**: Symbolic rule engine ($S_1$) and directed graph traversals ($S_2$).
- **Desired implementation**: Research Graph Neural Networks (GNNs) and neural theorem proving modules integrated into `Reasoning` to evaluate complex multi-hop hypotheses with probabilistic confidence guarantees.
- **Notes**: Long-term research initiative; must maintain constitutional safety invariants and verification traces.

---

## 4. Planning

### Deliberative Tree Search (`S4`) Activation in Normal Dialogue
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: Multi-alternative tree search (`tree_search.go` / `StageS4DeliberativeTreeSearch`) provides robust contingency generation and multi-branch exploration, but standard user turns default to `DepthTactical` (HTN/GOAP).
- **Current implementation**: `handleActiveGoal` invokes `executePlanningEpisode(ctx, req, DepthTactical)`. `DepthStrategic` (which triggers `tree_search.go`) runs only on explicit budget escalation.
- **Desired implementation**: Implement dynamic depth selection in `handleActiveGoal` where high-uncertainty goals or tasks marked with high risk automatically elevate to `DepthStrategic` / multi-alternative tree search ($S_4$).
- **Notes**: Monitor memory footprint and execution duration when expanding wide decision trees.

### Monte Carlo Tree Search (MCTS) & Neural Policy/Value Networks
- **Priority**: Low
- **Status**: Future Research
- **Reason**: For complex multi-step strategic workflows, HTN and GOAP require explicit domain schemas and goal postconditions.
- **Current implementation**: Domain-weighted HTN, GOAP, and A* heuristic scoring.
- **Desired implementation**: Integrate MCTS guided by learned neural value and policy networks (`Policy/Value networks`) to dynamically explore and prune expansive action state spaces.
- **Notes**: Requires alignment with `Learning` subsystem to train value models over historical `PlanningTrace` outcomes.

---

## 5. Decision

### Reflexive Evaluation (`EvaluateReflexive`) Fast-Path Routing
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: Micro-decisions and low-risk candidate selections do not require multi-criteria Pareto evaluation (`EvaluateDeliberative`), which consumes additional computation and time.
- **Current implementation**: `handleCandidatePlans` (`decision/service.go:229`) unconditionally invokes `EvaluateDeliberative(ctx, cs)` for every incoming `TopicCandidatePlans` envelope.
- **Desired implementation**: Implement a classification check on candidate urgency, cost, and budget tier. For reflexive budget envelopes or single-candidate confirmations, route directly to `EvaluateReflexive` (<2ms SLA) and publish appropriate outcome envelopes.
- **Notes**: Ensure `ShouldPublishToWorkspace` rules are respected or updated cleanly if reflexive decisions need downstream visibility.

### Reinforcement Learning Policy Optimization
- **Priority**: Low
- **Status**: Future Research
- **Reason**: Utility scoring weights across confidence, cost, and risk criteria are currently static or adjusted through explicit statistical formulas.
- **Current implementation**: Multi-criteria utility scoring and Pareto frontier analysis after Tier 1 Constitutional Hard Gate filtering.
- **Desired implementation**: Implement offline/online reinforcement learning policy optimization to dynamically fine-tune candidate attribute weights based on historical trace success rates and user feedback.
- **Notes**: Must operate downstream of the non-negotiable Tier 1 Constitutional Hard Gate.

---

## 6. Executive

### Distributed Multi-Node Episode Execution
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: IDUN currently executes all cognitive episodes and worker threads locally on a single runtime host.
- **Current implementation**: Single-process `ServiceV2` managing local concurrency slots, worker pools, and dynamic SLA budgets.
- **Desired implementation**: Extend `Executive` with distributed coordination capabilities (e.g., gRPC/raft or distributed workspace bridge) to dispatch and arbitrate cognitive bids across multi-node compute clusters.
- **Notes**: Preserve the content-blind arbitration contract and centralized Pre-Broadcast Constitutional Gate across all remote execution nodes.

### Predictive Preemption & Resource Forecasting
- **Priority**: Low
- **Status**: Future Work
- **Reason**: Currently, budget checks and constitutional gate evaluations occur right when an action bid is arbitrated or submitted (`SubmitBid` / `ArbitrateCompetition`).
- **Current implementation**: Reactive SLA timeout enforcement (`context.WithTimeout`) and budget unit deduction during competition arbitration.
- **Desired implementation**: Add predictive resource forecasting that monitors active queue depths across `Understanding`, `Reasoning`, and `Planning`, preemptively throttling or canceling low-priority speculative branches before resource starvation occurs.
- **Notes**: Integrate with `inference.Service.GetTelemetry()` queue depth metrics.

---

## 7. Attention

### Salience Regulation & Priority Queue Integration
- **Priority**: Medium
- **Status**: Unfinished Runtime Work / Planned Improvement
- **Reason**: Once `attention.Service` is wired into `runtime/host.go`, its salience scores (`SalienceScore`) must actively influence envelope processing priority across the Global Workspace.
- **Current implementation**: `attention.Service` computes salience and urgency adjustments (`workspace_bridge.go`), but the Global Workspace (`workspace.Service`) delivers envelopes in standard subscriber registration/FIFO order.
- **Desired implementation**: Integrate `attention.Service` salience thresholds directly into `workspace.Service` dispatch queues, ensuring envelopes with `Urgency >= 80` or high salience interrupt lower-priority background tasks immediately.
- **Notes**: Ensure priority inversion protection and deadlock prevention in the workspace message broker.

### Neural Attention Mechanism Models & Vector Salience
- **Priority**: Low
- **Status**: Future Research
- **Reason**: Hard-coded salience heuristics and rule-based urgency modifiers can miss nuanced context shifts across long-horizon dialogues.
- **Current implementation**: Rule-based scoring functions and explicit urgency weighting.
- **Desired implementation**: Integrate learned neural attention mechanism models and continuous vector salience embeddings to dynamically highlight critical context across multi-turn episodes.
- **Notes**: Must remain lightweight and memory-efficient ($O(1)$ retention bounding).

---

## 8. Learning

### Connect Learning Opportunity Emitting across Cognitive Modules
- **Priority**: Medium
- **Status**: Unfinished Runtime Work
- **Reason**: `learning.Service` (`service.go`) is active in `runtime/host.go` and subscribed to `TopicLearningOpportunity`, but no cognitive subsystem currently publishes `TopicLearningOpportunity` envelopes during normal turn flow.
- **Current implementation**: The `learning` service sits idle waiting for `TopicLearningOpportunity` envelopes.
- **Desired implementation**: Update `Reasoning` (when detecting novel symbolic rules or calibration errors) and `Decision` (when observing unexpected candidate rejection patterns) to construct and emit structured `TopicLearningOpportunity` envelopes.
- **Notes**: Enforce bounded memory invariants so learning trace ingestion never causes memory leaks or unbounded trace growth.

### Online Reinforcement & Gradient Updating
- **Priority**: Low
- **Status**: Future Research
- **Reason**: Continuous adaptation to user preferences across sessions requires online parameter updates and procedural skill synthesis.
- **Current implementation**: Statistical models, bounded campaign summaries (`LearningCampaignSummary`), and drift tracking (`TraceStatisticalSummary`).
- **Desired implementation**: Develop safe online reinforcement learning and gradient-updating mechanisms constrained by constitutional bounds to refine cognitive strategy profiles (`PlanningPolicyProfile`).
- **Notes**: Must undergo strict pre-broadcast constitutional validation before committing profile changes.

---

## 9. Reflection

### Connect Impasse Publishing across Reasoning & Planning
- **Priority**: Medium
- **Status**: Unfinished Runtime Work
- **Reason**: `reflection.Service` is instantiated in `runtime/host.go` and subscribed to `TopicImpasses`, but `Reasoning` and `Planning` do not publish `TopicImpasses` when deadlocks or plan infeasibility occur during standard turns.
- **Current implementation**: When `Planning` exhausts its budget or finds no valid plan (`TerminationNoValidPlan` / `PlanStatusInfeasible`), it returns an error or partial result to the caller without publishing to `TopicImpasses`.
- **Desired implementation**: Instrument `Planning` (`service.go:421-430`) and `Reasoning` (`deliberative.go`) to emit `TopicImpasses` envelopes when all candidates are disqualified or budgets expire, triggering `reflection.Service` to produce structured `ReflectionReport` critique artifacts.
- **Notes**: Ensure `ReflectionReport` recommendations flow cleanly into `Executive` or `Learning` without creating infinite retry loops.

### Metacognitive Critique Networks & Autoencoders
- **Priority**: Low
- **Status**: Future Research
- **Reason**: Rule-based impasse analysis (`AnalyzeTrends`) identifies explicit threshold breaches but may overlook subtle multi-episode performance drift.
- **Current implementation**: Pattern matching, rule engines, and historical summary comparative checks.
- **Desired implementation**: Research neural metacognitive critique networks and variational autoencoders trained over historical `LearningTrace` summaries to detect behavioral anomalies and structural cognitive inefficiencies.
- **Notes**: Keep analysis read-only (`HistoricalSummaryRequest`) to maintain absolute separation from live transaction paths.

---

## 10. Calibration

### Bayesian Hierarchical Calibration & Neural Adaptors
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: Current calibration (`calibration.Specialist`) uses flat Platt scaling and Beta calibration over primary/beam hypotheses, which can over- or under-calibrate when switching rapidly across diverse problem domains.
- **Current implementation**: Statistical scaling functions applied uniformly within `Reasoning` stage evaluation.
- **Desired implementation**: Implement Bayesian hierarchical calibration models and lightweight neural adaptors that dynamically adjust confidence profiles based on domain metadata and historical specialist accuracy.
- **Notes**: Ensure calibration execution remains deterministic and well within the allotted statistical calculation budget (<5ms).

---

## 11. Constitution

### Constitutional AI Critique & Formal Verification
- **Priority**: Low
- **Status**: Future Work
- **Reason**: While hard-coded safety rules (`ActionGate`) and HMAC-SHA256 tokens (`ActionApprovalToken`) guarantee absolute protection against known invariant violations, nuanced ethical trade-offs benefit from deliberative constitutional evaluation.
- **Current implementation**: Deterministic safety rule engine (`gate.go`) filtering `TopicActionExecution` and intermediate hypotheses.
- **Desired implementation**: Add a secondary Constitutional AI self-critique pass (via local LLM fallback) for complex boundary actions, alongside formal mathematical verification proofs for critical kernel safety invariants.
- **Notes**: The deterministic hard-coded safety gate must always take precedence over any probabilistic LLM critique.

---

## 12. Memory

### Automated CAS Garbage Collection & Generational Pruning
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: Content-Addressed Storage (`core/storage` & `payloadStorerAdapter`) persists envelopes, traces, and intermediate reasoning artifacts indefinitely, which can cause disk space accumulation over long periods.
- **Current implementation**: Append-only / store-and-retrieve CAS storage (`storage.Service`) without automated expiration.
- **Desired implementation**: Implement automated TTL-based pruning and generational garbage collection policies that archive or delete transient reflexive traces (`TraceID`, `ExecutionResponse`) while preserving long-term `HistoricalSummary` contracts and constitutional audit logs.
- **Notes**: Ensure concurrent read safety (`-race` clean) during background pruning sweeps.

---

## 13. Language Realization

### Remove Conversational Policy from Language Realization
- **Priority**: Medium
- **Status**: Technical Debt
- **Reason**: The runtime repair correctly prevents internal reasoning leakage and restores natural conversation (`I am IDUN...`, `Hello! How can I assist you today?`, etc.). However, Language Realization currently contains small amounts of conversational policy (e.g., greetings, identity responses, farewells, wellbeing responses) inside the realization prompt (`prompt.go`). Architecturally, these behaviors belong in the cognitive layer. Language Realization should remain responsible only for expressing approved semantic meaning naturally.
- **Current implementation**: The realization prompt (`presentation/realization/prompt.go`) maps specific intents (`greet_user`, `query_identity`, `query_wellbeing`, `farewell_user`) to conversational responses via system prompt instructions.
- **Desired implementation**: Future cognitive modules (`Reasoning` / `Decision`) should produce structured semantic responses such as:
  ```json
  {
    "Type": "Greeting",
    "Meaning": "Hello",
    "Tone": "Friendly"
  }
  ```
  or
  ```json
  {
    "Type": "Identity",
    "Meaning": "I am IDUN, your personal AI assistant.",
    "Tone": "Professional"
  }
  ```
  Language Realization should simply realize these semantic structures into natural language without explicit intent-matching rules inside the realization prompt.
- **Notes**: **Do not implement this now.** The current runtime repair is correct and should remain in place until the cognitive semantic-response pipeline matures. Implement after `Reasoning` and `Decision` natively emit `SemanticResponse` envelopes.

### Multi-Turn Dialogue Model & Multilingual Styling Engines
- **Priority**: Low
- **Status**: Future Work
- **Reason**: Surface realization currently executes single-turn stateless phrasing (`BuildRealizationPrompt`) targeting `en-US` with simple tone rules (`ToneProfessional`, `ToneConversational`).
- **Current implementation**: Template prompt formatting submitted to shared `InferenceService` (`local-realizer` / `qwen2.5:1.5b`).
- **Desired implementation**: Integrate specialized multi-turn dialogue realization models (`Dialogue Model`) that maintain conversational cadence over extended sessions, along with dynamic multilingual styling engines for localized idioms (`ja-JP`, `de-DE`, `es-ES`, etc.).
- **Notes**: Must strictly preserve the stateless realization guarantee (no NLU or decision comparison inside the presentation layer).

---

## 14. Infrastructure

### Elastic Worker Auto-Scaling & Telemetry Dashboard Export
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: `InferenceService` (`inference/service.go`) tracks queue depths (`reflexiveQueueDepth`, `standardQueueDepth`, `deliberativeQueueDepth`) and active workers via `GetTelemetry()`, but worker pools have fixed concurrency boundaries (`MaxConcurrency: 4` in `host.go`).
- **Current implementation**: Fixed concurrency per backend registered in `ModelRegistry` (`registry/service.go`).
- **Desired implementation**: Implement elastic worker auto-scaling that dynamically adjusts backend concurrency based on real-time SLA queue depth pressure, and create an OpenTelemetry/Prometheus export adapter to visualize live cognitive pipeline metrics on external monitoring dashboards.
- **Notes**: Ensure concurrency limits respect underlying hardware GPU/VRAM constraints (`ollama-local-01`).

---

## 15. Testing

### Comprehensive End-to-End Multimodal & Escalation Benchmarks
- **Priority**: Medium
- **Status**: Planned Improvement
- **Reason**: The repository maintains rigorous `-race` clean unit tests (`service_test.go`) and component benchmark suites, but complex multi-turn escalation flows ($S_1 \to S_8$ and HTN tree search $S_4$) need continuous integration validation under simulated user loads.
- **Current implementation**: Independent package test suites (`intelligence/understanding/...`, `presentation/realization/...`, `runtime/...`).
- **Desired implementation**: Build an automated end-to-end integration test harness (`tests/e2e`) that simulates full multi-turn user dialogues, injecting ambiguous prompts to verify automatic fallback escalation (`DeliberativeWorker` / `DeliberativeSpecialist`), constitutional gate arbitration, and accurate natural language realization without reasoning leakage.
- **Notes**: Integrate into CI pipelines with automated latency threshold checks.

---

## 16. Documentation

### Automated Architecture Status Verification Checks
- **Priority**: Low
- **Status**: Planned Improvement
- **Reason**: As IDUN evolves toward future phases, documentation tables in `intelligence/README.md` (`Code Exists`, `Wired into Runtime`, `Used in Production`) must stay precisely aligned with actual source code wiring.
- **Current implementation**: Comprehensive, exact manual verification and status audits (`README.md`).
- **Desired implementation**: Create a CI verification linter/script (`scripts/audit_architecture.go`) that parses `runtime/host.go` and `topics.go` AST definitions to automatically verify that documented runtime wiring and topic subscriptions match the code base exactly.
- **Notes**: Prevents architectural drift across long-term collaborative engineering cycles.

---

## 17. Future Research

### Self-Improving Autonomous Cognitive Loop
- **Priority**: Low
- **Status**: Future Research
- **Reason**: Achieving true autonomous agentic intelligence requires IDUN to continuously refine its own cognitive strategies and heuristics based on empirical outcome evidence over time.
- **Current implementation**: Static and statistical strategy definitions (`PlanningPolicyProfile`, `StrategySnapshot`) managed through explicit component configs.
- **Desired implementation**: Research a closed-loop autonomous improvement cycle where `Reflection`, `Learning`, and `Calibration` collaboratively propose, simulate, and deploy updated cognitive strategy snapshots (`StrategySnapshot`) in real time, strictly guarded by immutable `Constitution` safety firewalls.
- **Notes**: Requires mathematical safety verification and cryptographic rollback checkpoints before deployment.

---

## 18. Phase 2A Technical Debt & Enhancements

### Remove Legacy Struct Unpacking in Workspace Bridge
- **Priority**: Medium
- **Status**: Future Work
- **Reason**: The `decision/workspace_bridge.go` still contains `bridgeReasoningResultPayload` and `bridgeSemanticFramePayload` to support backwards compatibility parsing of legacy envelopes.
- **Desired implementation**: Completely remove these bridge payloads once all subsystems cleanly emit strict boundaries and legacy traces are cleared or ignored.

### Consolidate Candidate Metadata Mapping
- **Priority**: Low
- **Status**: Future Work
- **Reason**: `decision.Candidate` uses a stringly typed map (`map[string]string`) to pass `ResolvedGoal` and `PresentationDirectives` JSON strings down to the `WorkspaceBridge`.
- **Desired implementation**: Refactor `Candidate` metadata to support typed any-values or natively embed presentation hints directly into a dedicated Phase 2B candidate structure to avoid marshal/unmarshal overhead.

---

## 19. Phase 3B Capability Framework Engineering Improvements

### Capability Blueprint
- **Priority**: Low
- **Status**: Future Work
- **Reason**: The standard internal structure of every capability needs to be formally documented to assist developers.
- **Desired implementation**: Create a formal architecture blueprint describing the standard internal structure of every capability.

### Engineering Standard v1.1
- **Priority**: Low
- **Status**: Future Work
- **Reason**: The Capability Conformance Checklist requires periodic refinement.
- **Desired implementation**: Update the Engineering Standard to include a Mandatory Provider Architecture section, Mandatory Capability Metrics section, Engineering Quality section, Observability section, and an Updated Version History.

### Capability Generator
- **Priority**: Low (after Phase 3B)
- **Status**: Future Enhancement
- **Reason**: Copying the template directory manually is repetitive and prone to search/replace errors.
- **Desired implementation**: Create a command-line tool (e.g. `idun create capability Files`) that automatically creates a new capability from the Native Capability Template, generating the directory with all required scaffolding, tests, providers, metadata, metrics, lifecycle, permissions, and normalization already in place.

---

## 20. Native Files Capability Roadmap

### Streaming File Operations
- **Priority**: Medium
- **Status**: Deferred
- **Reason**: The current Files Capability loads file data directly during execution. Future versions should support efficient streaming for large files to avoid unnecessary memory consumption.
- **Desired implementation**: Add Read Stream, Write Stream, Chunk Reader, and Chunk Writer. Future goals include supporting extremely large files, minimizing memory usage, progressive reads/writes, context cancellation, and configurable chunk sizes. Do NOT implement during Phase 3B.3.

---

## 21. Native Media Capability Roadmap

### Media Session Manager
- **Priority**: Medium
- **Status**: Deferred
- **Reason**: The current Media Capability executes playback and recording as independent operations. Future versions should support persistent media sessions.
- **Desired implementation**: Add persistent playback and recording sessions, session lifecycle management, state tracking, concurrent sessions, and graceful cleanup of abandoned sessions. Do NOT implement during Phase 3B.6.

### Hardware Capability Detection
- **Priority**: Medium
- **Status**: Deferred
- **Reason**: Future versions of the Media Provider should expose the multimedia capabilities supported by the host hardware (e.g. microphones, hardware encoders/decoders).
- **Desired implementation**: Add hardware capability discovery, runtime capability reporting, supported codec enumeration, and device feature reporting. Do NOT implement during Phase 3B.6.

---

## 22. Native Devices Capability Roadmap

### Device Permission Cache
- **Priority**: Low
- **Status**: Deferred
- **Reason**: Some operating systems repeatedly prompt users for permissions such as Camera, Microphone, Bluetooth, and Location. Future versions should maintain a temporary cache of permission states.
- **Desired implementation**: Add caching of granted/denied permissions, expiration support, refresh mechanism, platform-specific adapters, and thread-safe cache. This must never bypass OS security. Do NOT implement during Phase 3B.7.

---

## 23. Phase 3B Capability Roadmap

The sequence of implementation for the Native Capability Framework:
1. **Phase 3B.1: Native System** (Completed)
2. **Phase 3B.2: Native Files** (Completed)
3. **Phase 3B.3: Native Communication** (Completed)
4. **Phase 3B.4: Native Network** (Completed)
5. **Phase 3B.5: Native Media** (Completed)
6. **Phase 3B.6: Native Devices** (Completed)
7. **Phase 3B.7: Native Automation** (Completed)

*(Note: CategoryExternalServices remains for future plugins, and CategoryLocation is merged into CategoryDevicesSensors).*

---

## 24. Native Capability Framework Status

### Platform Adapter Layer
- **Priority**: Medium
- **Status**: Deferred
- **Reason**: The current Native Providers intentionally expose a platform-independent interface while many implementations remain stubbed. Future versions should isolate OS-specific implementations into dedicated layers (e.g. `providers/windows/`, `providers/linux/`).
- **Desired implementation**: Separate platform-specific code from shared logic, minimize conditional compilation, simplify OS support, and allow independent evolution. This is a long-term enhancement and does NOT modify the current architecture. Do NOT begin implementation without authorization.

### Layer Status
- **Architecture**: FROZEN
- **Implementation**: COMPLETE
- **Modification Policy**:
  - Bug fixes: Allowed
  - Performance improvements: Allowed
  - Platform adapters: Allowed
  - New provider implementations: Allowed
  - Deferred TODO implementation: Allowed
  - Architectural redesign: Not allowed without a major version revision

---

## 25. Phase 6 Executive Runtime Improvements

### Expand DAG Traversal Test Cases
- **Priority**: Low
- **Status**: Future Work
- **Reason**: The concurrent `DAGExecutor` requires exhaustive parallel testing.
- **Desired implementation**: Create more exhaustive test cases in `engine_test.go` using manually constructed `ExecutionPlan` graphs to verify complex parallel traversal, dependencies, and graceful degradation on physical failure without full `planning/v3` mocking overhead.

---

## 26. Presentation Layer Improvements

### Template Caching (Deterministic Engine)
- **Priority**: Medium
- **Status**: Future Work
- **Reason**: `presentation/deterministic/engine.go` currently calls `template.ParseFiles(tmplPath)` on every `Realize()` invocation, reading from disk on each request. For high-frequency queries this causes avoidable filesystem I/O on every realization cycle.
- **Current implementation**: `Engine.Realize()` → `template.ParseFiles(tmplPath)` → render → return. No caching exists.
- **Desired implementation**: Load and parse all templates once during `NewEngine()` startup (or an explicit `Load()` call). Cache the parsed `*template.Template` values in a `map[string]*template.Template` protected by a read lock. On `Realize()`, look up the pre-parsed template from the map instead of reading disk.
- **Notes**: The cache is read-only after startup — no locking overhead during execution. Preserve the ability to reload templates for development hot-reload scenarios via an explicit method.

### ResponseType Constants
- **Priority**: Medium
- **Status**: Future Work
- **Reason**: String literals such as `"time"`, `"calculator"`, `"notes"` are currently scattered across capability `Execute()` functions. Typos in any one site silently break the deterministic routing path without compile-time detection.
- **Current implementation**: `ResponseType: "time"` hardcoded in `capabilities/native/time/capability.go`.
- **Desired implementation**: Introduce a shared constants package (e.g., `capabilities/responsetypes/constants.go`) defining named constants:
  ```go
  const (
      ResponseTypeTime       = "time"
      ResponseTypeCalculator = "calculator"
      ResponseTypeNotes      = "notes"
  )
  ```
  All capabilities and template files should reference these constants.
- **Notes**: Prevents typos, enables IDE navigation, and centralizes response type definitions across the codebase.

### Output Formatting Helpers
- **Priority**: Medium
- **Status**: Future Work
- **Reason**: Capabilities return raw structured data (e.g., `time.Time`, `int64`, `float64`). Currently templates must format these values inline using Go's default `{{.Field}}` rendering which produces machine-formatted strings (e.g., RFC3339 timestamps, raw float literals).
- **Current implementation**: `time.tmpl` receives `CurrentTime` as a raw `time.Time` value and renders it using Go's default formatter.
- **Desired implementation**: Introduce a `presentation/deterministic/helpers.go` package that registers reusable Go template functions:
  - `formatTime` — locale-aware time display
  - `formatDate` — day/month/year display
  - `formatDuration` — human-readable duration (e.g., "2 hours 15 minutes")
  - `formatFileSize` — byte count to human string (e.g., "1.4 MB")
  - `formatNumber` — locale-aware number formatting
  Capabilities continue returning raw structured data. All formatting logic belongs in the realization layer.
- **Notes**: Template functions must be registered on the `template.FuncMap` before `Execute()` is called.

---

## 27. Localization & Architecture Evolution

### Template Localization
- **Priority**: Low
- **Status**: Future Work
- **Reason**: The current template repository (`presentation/router/templates/`) stores a single flat directory of response templates with no language awareness. Future users in non-English locales would receive `en-US`-formatted responses.
- **Current implementation**: Single template directory, e.g., `data/runtime/templates/time.tmpl`.
- **Desired implementation**: Restructure templates into language-tagged subdirectories:
  ```
  data/runtime/templates/
      en/
          time.tmpl
          calculator.tmpl
      hi/
          time.tmpl
      te/
          time.tmpl
  ```
  The Deterministic Engine should accept a `language string` parameter and resolve the correct subdirectory. Capabilities remain completely unmodified.
- **Notes**: Implement after at least 3 deterministic capabilities are stable.

### World Pillar Refactor (Realization Consolidation)
- **Priority**: Low
- **Status**: Future Architecture Work
- **Reason**: The `presentation/` layer (Router, Deterministic Engine, Generative Engine, interfaces, types) currently exists as a separate pillar for implementation convenience. Architecturally, these components are responsible for delivering final output to the user, which is the stated purpose of the World pillar.
- **Current implementation**: `presentation/` and `world/` are separate packages. World only accepts finalized `RealizedOutput` envelopes.
- **Desired implementation**: Once the overall presentation architecture is stable and at least 4 capabilities have been verified end-to-end, evaluate migrating `presentation/` components into the `world/` pillar as a structural consolidation. This is a source reorganization only — no behavioral changes.
- **Notes**: **Do not implement now.** Do not merge until architecture is validated across multiple capability types. This is a structural cleanup, not a feature.

---

## Future Ontology Evolution (Post Phase 4B.3)

These items represent architectural improvements for future phases or ontology evolution. They should not be implemented during Phase 4B.3.

### TODO 1 — Move Beyond Slot-Name Classification
- **Goal**: Classify semantic objects using grammar metadata (e.g., Semantic Hints attached to Grammar Rules) rather than relying on slot names.
- **Current State**: Extractors use `switch slot.Name()`.
- **Future State**: `Grammar Rule -> Semantic Hint -> Semantic Object`.

### TODO 2 — Introduce EntityExpression
- **Goal**: Map mathematical expressions to `EntityExpression` rather than `EntityNumber`.
- **Current State**: `expression` slot maps to `EntityNumber`.
- **Future State**: `Expression -> Reasoning / Calculator -> Result -> EntityNumber`.
- **Reason**: `decision.Candidate` uses a stringly typed map (`map[string]string`) to pass `ResolvedGoal` and `PresentationDirectives` JSON strings down to the `WorkspaceBridge`.
- **Desired implementation**: Refactor `Candidate` metadata to support typed any-values or natively embed presentation hints directly into a dedicated Phase 2B candidate structure to avoid marshal/unmarshal overhead.

---

## 19. Phase 3B Capability Framework Engineering Improvements

### Capability Blueprint
- **Priority**: Low
- **Status**: Future Work
- **Reason**: The standard internal structure of every capability needs to be formally documented to assist developers.
- **Desired implementation**: Create a formal architecture blueprint describing the standard internal structure of every capability.

### Engineering Standard v1.1
- **Priority**: Low
- **Status**: Future Work
- **Reason**: The Capability Conformance Checklist requires periodic refinement.
- **Desired implementation**: Update the Engineering Standard to include a Mandatory Provider Architecture section, Mandatory Capability Metrics section, Engineering Quality section, Observability section, and an Updated Version History.

### Capability Generator
- **Priority**: Low (after Phase 3B)
- **Status**: Future Enhancement
- **Reason**: Copying the template directory manually is repetitive and prone to search/replace errors.
- **Desired implementation**: Create a command-line tool (e.g. `idun create capability Files`) that automatically creates a new capability from the Native Capability Template, generating the directory with all required scaffolding, tests, providers, metadata, metrics, lifecycle, permissions, and normalization already in place.

---

## 20. Native Files Capability Roadmap

### Streaming File Operations
- **Priority**: Medium
- **Status**: Deferred
- **Reason**: The current Files Capability loads file data directly during execution. Future versions should support efficient streaming for large files to avoid unnecessary memory consumption.
- **Desired implementation**: Add Read Stream, Write Stream, Chunk Reader, and Chunk Writer. Future goals include supporting extremely large files, minimizing memory usage, progressive reads/writes, context cancellation, and configurable chunk sizes. Do NOT implement during Phase 3B.3.

---

## 21. Native Media Capability Roadmap

### Media Session Manager
- **Priority**: Medium
- **Status**: Deferred
- **Reason**: The current Media Capability executes playback and recording as independent operations. Future versions should support persistent media sessions.
- **Desired implementation**: Add persistent playback and recording sessions, session lifecycle management, state tracking, concurrent sessions, and graceful cleanup of abandoned sessions. Do NOT implement during Phase 3B.6.

### Hardware Capability Detection
- **Priority**: Medium
- **Status**: Deferred
- **Reason**: Future versions of the Media Provider should expose the multimedia capabilities supported by the host hardware (e.g. microphones, hardware encoders/decoders).
- **Desired implementation**: Add hardware capability discovery, runtime capability reporting, supported codec enumeration, and device feature reporting. Do NOT implement during Phase 3B.6.

---

## 22. Native Devices Capability Roadmap

### Device Permission Cache
- **Priority**: Low
- **Status**: Deferred
- **Reason**: Some operating systems repeatedly prompt users for permissions such as Camera, Microphone, Bluetooth, and Location. Future versions should maintain a temporary cache of permission states.
- **Desired implementation**: Add caching of granted/denied permissions, expiration support, refresh mechanism, platform-specific adapters, and thread-safe cache. This must never bypass OS security. Do NOT implement during Phase 3B.7.

---

## 23. Phase 3B Capability Roadmap

The sequence of implementation for the Native Capability Framework:
1. **Phase 3B.1: Native System** (Completed)
2. **Phase 3B.2: Native Files** (Completed)
3. **Phase 3B.3: Native Communication** (Completed)
4. **Phase 3B.4: Native Network** (Completed)
5. **Phase 3B.5: Native Media** (Completed)
6. **Phase 3B.6: Native Devices** (Completed)
7. **Phase 3B.7: Native Automation** (Completed)

*(Note: CategoryExternalServices remains for future plugins, and CategoryLocation is merged into CategoryDevicesSensors).*

---

## 24. Native Capability Framework Status

### Platform Adapter Layer
- **Priority**: Medium
- **Status**: Deferred
- **Reason**: The current Native Providers intentionally expose a platform-independent interface while many implementations remain stubbed. Future versions should isolate OS-specific implementations into dedicated layers (e.g. `providers/windows/`, `providers/linux/`).
- **Desired implementation**: Separate platform-specific code from shared logic, minimize conditional compilation, simplify OS support, and allow independent evolution. This is a long-term enhancement and does NOT modify the current architecture. Do NOT begin implementation without authorization.

### Layer Status
- **Architecture**: FROZEN
- **Implementation**: COMPLETE
- **Modification Policy**:
  - Bug fixes: Allowed
  - Performance improvements: Allowed
  - Platform adapters: Allowed
  - New provider implementations: Allowed
  - Deferred TODO implementation: Allowed
  - Architectural redesign: Not allowed without a major version revision

---

## 25. Phase 6 Executive Runtime Improvements

### Expand DAG Traversal Test Cases
- **Priority**: Low
- **Status**: Future Work
- **Reason**: The concurrent `DAGExecutor` requires exhaustive parallel testing.
- **Desired implementation**: Create more exhaustive test cases in `engine_test.go` using manually constructed `ExecutionPlan` graphs to verify complex parallel traversal, dependencies, and graceful degradation on physical failure without full `planning/v3` mocking overhead.

---

## 26. Presentation Layer Improvements

### Template Caching (Deterministic Engine)
- **Priority**: Medium
- **Status**: Future Work
- **Reason**: `presentation/deterministic/engine.go` currently calls `template.ParseFiles(tmplPath)` on every `Realize()` invocation, reading from disk on each request. For high-frequency queries this causes avoidable filesystem I/O on every realization cycle.
- **Current implementation**: `Engine.Realize()` → `template.ParseFiles(tmplPath)` → render → return. No caching exists.
- **Desired implementation**: Load and parse all templates once during `NewEngine()` startup (or an explicit `Load()` call). Cache the parsed `*template.Template` values in a `map[string]*template.Template` protected by a read lock. On `Realize()`, look up the pre-parsed template from the map instead of reading disk.
- **Notes**: The cache is read-only after startup — no locking overhead during execution. Preserve the ability to reload templates for development hot-reload scenarios via an explicit method.

### ResponseType Constants
- **Priority**: Medium
- **Status**: Future Work
- **Reason**: String literals such as `"time"`, `"calculator"`, `"notes"` are currently scattered across capability `Execute()` functions. Typos in any one site silently break the deterministic routing path without compile-time detection.
- **Current implementation**: `ResponseType: "time"` hardcoded in `capabilities/native/time/capability.go`.
- **Desired implementation**: Introduce a shared constants package (e.g., `capabilities/responsetypes/constants.go`) defining named constants:
  ```go
  const (
      ResponseTypeTime       = "time"
      ResponseTypeCalculator = "calculator"
      ResponseTypeNotes      = "notes"
  )
  ```
  All capabilities and template files should reference these constants.
- **Notes**: Prevents typos, enables IDE navigation, and centralizes response type definitions across the codebase.

### Output Formatting Helpers
- **Priority**: Medium
- **Status**: Future Work
- **Reason**: Capabilities return raw structured data (e.g., `time.Time`, `int64`, `float64`). Currently templates must format these values inline using Go's default `{{.Field}}` rendering which produces machine-formatted strings (e.g., RFC3339 timestamps, raw float literals).
- **Current implementation**: `time.tmpl` receives `CurrentTime` as a raw `time.Time` value and renders it using Go's default formatter.
- **Desired implementation**: Introduce a `presentation/deterministic/helpers.go` package that registers reusable Go template functions:
  - `formatTime` — locale-aware time display
  - `formatDate` — day/month/year display
  - `formatDuration` — human-readable duration (e.g., "2 hours 15 minutes")
  - `formatFileSize` — byte count to human string (e.g., "1.4 MB")
  - `formatNumber` — locale-aware number formatting
  Capabilities continue returning raw structured data. All formatting logic belongs in the realization layer.
- **Notes**: Template functions must be registered on the `template.FuncMap` before `Execute()` is called.

---

## 27. Localization & Architecture Evolution

### Template Localization
- **Priority**: Low
- **Status**: Future Work
- **Reason**: The current template repository (`presentation/router/templates/`) stores a single flat directory of response templates with no language awareness. Future users in non-English locales would receive `en-US`-formatted responses.
- **Current implementation**: Single template directory, e.g., `data/runtime/templates/time.tmpl`.
- **Desired implementation**: Restructure templates into language-tagged subdirectories:
  ```
  data/runtime/templates/
      en/
          time.tmpl
          calculator.tmpl
      hi/
          time.tmpl
      te/
          time.tmpl
  ```
  The Deterministic Engine should accept a `language string` parameter and resolve the correct subdirectory. Capabilities remain completely unmodified.
- **Notes**: Implement after at least 3 deterministic capabilities are stable.

### World Pillar Refactor (Realization Consolidation)
- **Priority**: Low
- **Status**: Future Architecture Work
- **Reason**: The `presentation/` layer (Router, Deterministic Engine, Generative Engine, interfaces, types) currently exists as a separate pillar for implementation convenience. Architecturally, these components are responsible for delivering final output to the user, which is the stated purpose of the World pillar.
- **Current implementation**: `presentation/` and `world/` are separate packages. World only accepts finalized `RealizedOutput` envelopes.
- **Desired implementation**: Once the overall presentation architecture is stable and at least 4 capabilities have been verified end-to-end, evaluate migrating `presentation/` components into the `world/` pillar as a structural consolidation. This is a source reorganization only — no behavioral changes.
- **Notes**: **Do not implement now.** Do not merge until architecture is validated across multiple capability types. This is a structural cleanup, not a feature.

---

## Future Ontology Evolution (Post Phase 4B.3)

These items represent architectural improvements for future phases or ontology evolution. They should not be implemented during Phase 4B.3.

### TODO 1 — Move Beyond Slot-Name Classification
- **Goal**: Classify semantic objects using grammar metadata (e.g., Semantic Hints attached to Grammar Rules) rather than relying on slot names.
- **Current State**: Extractors use `switch slot.Name()`.
- **Future State**: `Grammar Rule -> Semantic Hint -> Semantic Object`.

### TODO 2 — Introduce EntityExpression
- **Goal**: Map mathematical expressions to `EntityExpression` rather than `EntityNumber`.
- **Current State**: `expression` slot maps to `EntityNumber`.
- **Future State**: `Expression -> Reasoning / Calculator -> Result -> EntityNumber`.

### TODO 3 — Introduce EntityCommand
- **Goal**: Map executable commands (e.g., shutdown, restart, delete, copy, move, lock) to `EntityCommand`.
- **Current State**: `operation` slot maps to `EntityUnknown`.

### TODO 4 — Refine Document Representation
- **Goal**: Evolve `EntityDocument` to explicitly handle structured documents.
- **Future State**: `EntityDocument` containing `EntityDocumentTitle` and `EntityDocumentBody`.
- **Current State**: Both `title` and `content` map directly to `EntityDocument`.

### TODO 5 — Implement Additional Normalizers
- **Goal**: Expand the `normalizers` package beyond temporal normalization.
- **Planned Files**: 
  - `normalizers/number.go` (normalize text to numeric values)
  - `normalizers/unit.go` (normalize "kilometers" to standard SI units)
  - `normalizers/expression.go` (normalize math expressions)
- **Status**: Deferred to post-Phase 4B.4.

### Replace mock.capability with True Communicative Capabilities
- **Priority**: Low
- **Status**: Future Architecture Work
- **Reason**: The Planning and Executive V3 boundaries currently rely on a mock.capability placeholder to route purely communicative intents (e.g., greet_user, query_identity) to the generative realization engine without triggering standard OS execution flows. 
- **Desired implementation**: In future phases (e.g., Chit-Chat Module / Chit-Chat Capability), replace the mock.capability injection in planning/v3/legacy_adapters.go with actual semantic mappings to a dedicated communicative capability. This capability will naturally produce the capabilities.Generative realization response currently spoofed by MockCommunicativeExecutor.
- **Notes**: Must not change Understanding, Planning, or Realization layers. This is purely an Executive/Capability-level migration.

## Post-Restoration Governance Enhancements (Phase 5.x)

### Phase 5.x — Restoration Traceability Matrix
**Objective**: Introduce a formal Traceability Matrix linking every restored artifact across the architecture.

**Tasks**:
- Generate a Traceability Matrix linking:
  - Grammar Rules
  - Intents
  - Planning Mappings
  - Application Capabilities
  - Native Capabilities
  - Verification Status
- Detect orphaned or undocumented restoration artifacts.
- Use the matrix for future architectural audits and regression analysis.

### Phase 5.x — Architectural Exception Register
**Objective**: Maintain a centralized register of intentional architectural limitations.

**Tasks**:
For each exception, record:
- Component
- Reason
- Classification
- Future Phase
- Status

*Example entries*:
- Complex calculator expressions
- Reminder scheduler RFC3339 limitation
- Future temporal enhancements
- Future Unified Policy Engine migration

This register should clearly distinguish intentional MVP limitations from implementation defects.

### Phase 5.x — Repository Certification Baseline
**Objective**: Strengthen long-term auditability by introducing repository-level certification baselines.

**Tasks**:
- Create permanent Git tags for certified architectural baselines.
- Record repository commit hashes in certification artifacts.
- Allow future audits to compare against exact certified repository snapshots.

### Phase 5.x — Change Control Framework
**Objective**: Formalize architectural governance after restoration.

**Tasks**:
Introduce documented change-control procedures covering:
- Architectural review requirements
- Component impact analysis
- Baseline modification process
- Certification document revision policy

This expands the current Change Control Statement into a complete governance framework.

### Phase 5.x — Long-Term Certification Metrics
**Objective**: Extend architectural certification beyond the restoration effort.

Future certification metrics may include:
- Traceability completeness
- Architectural debt
- Documentation coverage
- Engineering rule coverage
- Governance compliance
- Long-term architectural drift analysis

### Deferred Rationale
These governance enhancements improve long-term maintainability, auditability, and architectural evolution. However, they are not prerequisites for certifying the deterministic restoration completed during Phase 4.

Sprint 7 already verifies:
- Restoration correctness
- Architecture compliance
- Runtime behavior
- Documentation consistency
- Engineering rule compliance
- **Notes**: Must not change Understanding, Planning, or Realization layers. This is purely an Executive/Capability-level migration.

## Post-Restoration Governance Enhancements (Phase 5.x)

### Phase 5.x — Restoration Traceability Matrix
**Objective**: Introduce a formal Traceability Matrix linking every restored artifact across the architecture.

**Tasks**:
- Generate a Traceability Matrix linking:
  - Grammar Rules
  - Intents
  - Planning Mappings
  - Application Capabilities
  - Native Capabilities
  - Verification Status
- Detect orphaned or undocumented restoration artifacts.
- Use the matrix for future architectural audits and regression analysis.

### Phase 5.x — Architectural Exception Register
**Objective**: Maintain a centralized register of intentional architectural limitations.

**Tasks**:
For each exception, record:
- Component
- Reason
- Classification
- Future Phase
- Status

*Example entries*:
- Complex calculator expressions
- Reminder scheduler RFC3339 limitation
- Future temporal enhancements
- Future Unified Policy Engine migration

This register should clearly distinguish intentional MVP limitations from implementation defects.

### Phase 5.x — Repository Certification Baseline
**Objective**: Strengthen long-term auditability by introducing repository-level certification baselines.

**Tasks**:
- Create permanent Git tags for certified architectural baselines.
- Record repository commit hashes in certification artifacts.
- Allow future audits to compare against exact certified repository snapshots.

### Phase 5.x — Change Control Framework
**Objective**: Formalize architectural governance after restoration.

**Tasks**:
Introduce documented change-control procedures covering:
- Architectural review requirements
- Component impact analysis
- Baseline modification process
- Certification document revision policy

This expands the current Change Control Statement into a complete governance framework.

### Phase 5.x — Long-Term Certification Metrics
**Objective**: Extend architectural certification beyond the restoration effort.

Future certification metrics may include:
- Traceability completeness
- Architectural debt
- Documentation coverage
- Engineering rule coverage
- Governance compliance
- Long-term architectural drift analysis

### Deferred Rationale
These governance enhancements improve long-term maintainability, auditability, and architectural evolution. However, they are not prerequisites for certifying the deterministic restoration completed during Phase 4.

Sprint 7 already verifies:
- Restoration correctness
- Architecture compliance
- Runtime behavior
- Documentation consistency
- Engineering rule compliance
- Security boundaries
- Regression stability

That evidence is sufficient to certify and freeze the Phase 4 restoration baseline.

These governance enhancements are therefore intentionally scheduled for Phase 5.x, where the project transitions from restoration to long-term architectural evolution.

---

## Phase 5.x — Presentation Architecture Evolution

These enhancements improve the long-term extensibility and modularity of the Presentation layer. They are intentionally deferred and must not modify the certified Phase 4 baseline.

### Phase 5.x — PresentationContext Builder Simplification
- **Priority**: Medium
- **Status**: Future Work
- **Objective**: Further reduce coupling between the Presentation layer and Capability internals.
- **Current state**: `PresentationContextBuilder` exposes `FromCapabilityResult(capabilities.CapabilityResult)`, which means the builder API surface requires knowledge of the capabilities package. The builder is currently the only controlled boundary where `CapabilityResult` enters the Presentation layer.
- **Desired implementation**: Simplify the construction API so the caller provides only presentation-level values, for example:
  - `presentation.NewPresentationContext(responseType, strategy, intent, parentRef)`, or
  - a builder that accepts individual fields without importing `capabilities.CapabilityResult` directly.
  The internal translation of `capabilities.RealizationStrategy` → `policy.RealizationStrategy` should be hidden inside an adapter or factory, not visible in the public API.
- **Goal**: The Presentation layer should be fully isolated from future Capability contract changes. If `CapabilityResult` gains or loses fields, the Presentation layer must not require updates.
- **Deferred because**: The current `PresentationContextBuilder` already isolates the Router and Policy from `CapabilityResult`. The simplification is a UX improvement on the builder API, not a functional requirement.

### Phase 5.x — RealizationPlan
- **Priority**: Medium
- **Status**: Future Work
- **Objective**: Introduce an intermediate `RealizationPlan` object between `RealizationPolicy` and `RealizationEngine`.
- **Current architecture**:
  ```
  PresentationContext → RealizationPolicy → RealizationEngine
  ```
- **Future architecture**:
  ```
  PresentationContext → RealizationPolicy → RealizationPlan → RealizationEngine
  ```
- **Motivation**: Today the policy selects only an engine. A `RealizationPlan` allows the policy to encode richer presentation instructions alongside the engine reference, without changing the `RealizationEngine` interface. A plan may carry:
  - Realization engine reference
  - Response style (professional, conversational, concise)
  - Verbosity level
  - Formatting preferences (markdown, plain text, rich text)
  - GUI presentation hints
  - Voice / TTS configuration
  - Streaming options
  - Adaptive presentation metadata for future learned models
- **Benefit**: Engine selection and presentation planning remain separately evolvable. The Router is unchanged. Adding a new presentation dimension requires only extending `RealizationPlan`, not modifying the `RealizationPolicy` interface.
- **Deferred because**: The current pipeline produces correct output without a plan object. This is an extensibility enhancement for Phase 5+ multi-modal or adaptive realization scenarios.

### Phase 5.x — Learned Realization Policy
- **Priority**: Low-Medium
- **Status**: Future Work
- **Objective**: Replace `DeterministicRealizationPolicy` with a learned realization selector without modifying the Router.
- **Current implementation**: `DeterministicRealizationPolicy` uses a static two-level rule table (ResponseType → engine, Strategy fallback → engine).
- **Future implementation**: `AdaptiveRealizationPolicy` (or `NeuralRealizationPolicy`) using a decision model that accepts richer inputs than response type alone. Possible future inputs include:
  - Response type and semantic richness
  - Originating intent
  - Conversation history and state
  - User preferences and accessibility settings
  - Confidence score from the understanding layer
  - Latency and execution budget
  - Device capabilities and rendering environment
  - Future optimization criteria
- **Constraint**: The `RealizationPolicy` interface (`Select(ctx, PresentationContext) (RealizationEngine, error)`) and the Router must remain unchanged when this replacement occurs. Only the injected policy implementation is replaced.
- **Deferred because**: The deterministic rule table is sufficient for the certified Phase 4 capabilities. A learned selector requires a training corpus and evaluation framework that does not yet exist.

### Phase 5.x — PresentationContext Immutability
- **Objective**: Prevent accidental mutation of `PresentationContext` after construction.
- **Future Tasks**:
  - Introduce an immutable `PresentationContext` (e.g. expose only getter methods).
  - Restrict mutation exclusively to the `PresentationContextBuilder`.
  - Treat the built context as read-only throughout the Presentation pipeline.
- **Deferred Rationale**: The current builder pattern is sufficient for the certified architecture. Immutability is a future maintainability improvement.

### Phase 5.x — Presentation View Model
- **Objective**: Introduce an explicit Presentation View Model between semantic data and template rendering.
- **Future Tasks**:
  - Add a Presentation View Model layer.
  - Allow deterministic templates to consume only view models instead of raw semantic data.
  - Keep semantic contracts and presentation models independently evolvable.
- **Deferred Rationale**: The current merge of semantic data with presentation context is appropriate for Phase 5.0. A dedicated view model would improve long-term maintainability but is not required for adopting the Presentation pipeline.

### Deferred Rationale

These enhancements improve long-term extensibility and maintainability of the Presentation architecture but are not required for the certified deterministic restoration.

The current implementation (`DeterministicRealizationPolicy`, `PresentationContextBuilder`, direct engine injection) already satisfies the Phase 4 architecture and certification requirements. The `RealizationPolicy` interface is stable and the Router is already decoupled from selection logic.

These items are therefore intentionally scheduled for Phase 5.x, where the project extends the certified baseline with adaptive, multi-modal, and learned presentation capabilities.

---

## Phase 5.x — Runtime Acceptance Evolution (Completed)

✅ **Suggested Ownership & File Mapping**: Implemented in `cmd/runtime_acceptance/main.go`. The test harness now maps failing categories to their likely owner layer and source file, reducing mean time to investigation.

✅ **Runtime Acceptance Coverage**: Implemented. The test harness now dynamically tracks the universe of supported capabilities and reports total coverage percentage to prevent un-tested capabilities from being released.

✅ **Behavioral Validation Upgrade (Phase 5.2.x)**: Implemented. The Runtime Acceptance Test Harness was completely overhauled to validate user-facing behavior first and explicitly classify failing layers using structured metadata from the runtime.

 # #   T e c h n i c a l   D e b t 
 
 # # #   U 8 . 5      R a w   U s e r   I n p u t   P r e s e r v a t i o n 
 * * S t a t u s : * *   D e f e r r e d 
 
 * * D e s c r i p t i o n : * * 
 C u r r e n t l y   t h e   U n d e r s t a n d i n g   l a y e r   p u b l i s h e s   o n l y   t h e   S e m a n t i c F r a m e . 
 T h e   o r i g i n a l   r a w   u s e r   u t t e r a n c e   i s   d i s c a r d e d   b e f o r e   d o w n s t r e a m   c o g n i t i v e   s y s t e m s   r e c e i v e   i t . 
 
 F u t u r e   r e a s o n i n g ,   k n o w l e d g e ,   i n t e r n e t ,   l e a r n i n g ,   a n d   c o n v e r s a t i o n a l   m e m o r y   s h o u l d   h a v e   a c c e s s   t o   b o t h : 
 -   O r i g i n a l   r a w   u s e r   i n p u t 
 -   S e m a n t i c F r a m e 
 w i t h o u t   c h a n g i n g   t h e   d e t e r m i n i s t i c   U n d e r s t a n d i n g   p i p e l i n e . 
 
 * * R e a s o n : * * 
 F u t u r e   c o g n i t i v e   m o d u l e s   s h o u l d   r e a s o n   o v e r   n a t u r a l   l a n g u a g e ,   n o t   o n l y   n o r m a l i z e d   s e m a n t i c   r e p r e s e n t a t i o n s . 
 
 # # #   N o r m a l i z e r   C o u p l i n g   A u d i t 
 * * S t a t u s : * *   M i n o r   T e c h n i c a l   D e b t 
 
 * * D e s c r i p t i o n : * * 
 S o m e   g r a m m a r   r u l e s   c u r r e n t l y   d e p e n d   o n   n o r m a l i z a t i o n   b e h a v i o r   ( f o r   e x a m p l e   c o n v e r s a t i o n a l   f i l l e r   r e m o v a l ) . 
 T h i s   i s   a c c e p t a b l e   f o r   U 6   a n d   s h o u l d   r e m a i n   u n c h a n g e d . 
 R e v i e w   d u r i n g   f u t u r e   U n d e r s t a n d i n g   e v o l u t i o n   t o   r e d u c e   c o u p l i n g   w h e r e   p r a c t i c a l   w i t h o u t   c h a n g i n g   d e t e r m i n i s t i c   b e h a v i o r . 
  
 
 # #   P h a s e   U 7      C o n t e x t u a l   U n d e r s t a n d i n g 
 
 -   [   ]   * * U 7 . 1   C o m p o n e n t   E x p a n s i o n * * :   A d d   t h e   c o n t e x t - s p e c i f i c   s e m a n t i c   f r a g m e n t s   ( C o n c e p t A n a p h o r a ,   C o n n e c t o r E l l i p s i s ,   C o n c e p t A f f i r m a t i o n ,   C o n c e p t N e g a t i o n )   t o   g r a m m a r _ c o m p o n e n t s . g o . 
 -   [   ]   * * U 7 . 2   C o n t e x t   B u i l d e r s * * :   I n t r o d u c e   a   n e w   B u i l d C o n t e x t R u l e   f u n c t i o n   i n   g r a m m a r _ b u i l d e r s . g o   s p e c i f i c a l l y   d e s i g n e d   f o r   e l l i p t i c a l   s t r u c t u r e s . 
 -   [   ]   * * U 7 . 3   C o n t e x t   D o m a i n   L o a d e r * * :   C r e a t e   g r a m m a r _ c o n t e x t . g o   t o   h o u s e   t h e   n e w   r u l e . c o n t e x t . *   d e f i n i t i o n s . 
 -   [   ]   * * U 7 . 4   C o n t e x t   B i n d e r   R e f a c t o r i n g * * :   D e p r e c a t e   t h e   n a i v e   p r o n o u n   b i n d i n g   i n   D e f a u l t R e f e r e n t B i n d e r   i n   f a v o r   o f   r e l y i n g   o n   t h e   G r a m m a r   S p e c i a l i s t   e x t r a c t i n g   t h e   p r o n o u n   a s   a   n a t i v e   s l o t   ( S l o t A n a p h o r a ) . 
 -   [   ]   * * U 7 . 5   C e r t i f i c a t i o n * * :   E x p a n d   e v a l _ p h r a s e s . g o   t o   t e s t   c o n v e r s a t i o n a l   f o l l o w - u p s   a n d   e n s u r e   n o   r e g r e s s i o n s   o n   t h e   1 1 5   b a s e l i n e   p h r a s e s . 
  
 
 -   [   ]   * * M i n o r   I m p r o v e m e n t :   I n t e r f a c e   S e g r e g a t i o n   f o r   S t a t e * * :   R e p l a c e   t h e   m o n o l i t h i c   D i a l o g u e S t a t e   c o n c e p t   w i t h   i s o l a t e d   i n t e r f a c e s   ( M e m o r y R e a d e r ,   A c t i v e G o a l R e a d e r ,   e t c . )   f o r   C o n t e x t   R e s o l v e r   d e p e n d e n c i e s . 
 -   [   ]   * * M i n o r   I m p r o v e m e n t :   T e m p o r a l   A n c h o r i n g * * :   T r a c k   t e m p o r a l   r e s o l u t i o n   ( " n o w " )   a s   a   f o r m a l   c o n t e x t u a l   r e s p o n s i b i l i t y   i n   t h e   R e s o l v e r   i n s t e a d   o f   a d - h o c   r e s o l u t i o n . 
 -   [   ]   * * M i n o r   I m p r o v e m e n t :   C r o s s - D o m a i n   C l a s h   H a n d l i n g * * :   I m p l e m e n t   e x p l i c i t   c o n f l i c t   r e s o l u t i o n   l o g i c   f o r   a m b i g u o u s   p r o n o u n s   ( e . g .   i f   " d e l e t e   i t "   c o u l d   r e f e r   t o   a   F i l e   o r   a n   A l a r m ) . 
  
 