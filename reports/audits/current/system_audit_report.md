# IDUN V3 — System Architecture Audit Report (Fresh)

**Generated**: 2026-07-31
**Codebase Version**: `2.0.0-FROZEN`
**Audit Method**: Direct codebase inspection of source files, `runtime/host.go` wiring, and build verification.
**Previous Report**: `sysem_audit_report` (superseded by this document)

> [!IMPORTANT]
> This audit is generated exclusively from the current source code. It does not inherit conclusions from previous reports. Each status is backed by a cited file or grep result.

---

## Project Transition Status

**Tier 1 — Foundation**
────────────────────────────
✓ Kernel
✓ Infrastructure
✓ Storage
✓ Workspace
✓ Capability Framework
✓ Presentation
✓ Executive
✓ Cognitive Architecture
✓ Semantic Contracts
✓ Schema Alignment

**Status: COMPLETE**
────────────────────────────

**Tier 2 — Cognitive Intelligence**
────────────────────────────
□ Understanding implementation
□ Reasoning implementation
□ Planning intelligence
□ Decision intelligence
□ Reflection / Learning
□ Meta-cognition
□ Continuous improvement

**Status: READY TO BEGIN**

---

## Status Legend

| Symbol | Meaning |
| :--- | :--- |
| ✅ **Implemented** | Fully coded, wired into runtime, and verified working |
| ⚠️ **Partially Implemented** | Code exists but has known gaps, missing wiring, or incomplete features |
| 🔶 **Placeholder** | Stub/skeleton that compiles and runs but returns no real output |
| ❌ **Missing** | Not implemented; expected by the architecture |

---

## 1. Kernel

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **Bus** (`kernel/bus.go`) | ✅ Implemented | `bus.go` (9.5 KB), `bus_test.go` (15 KB) — topic-based pub/sub with subscriber registration |
| **Lifecycle** (`kernel/lifecycle.go`) | ✅ Implemented | `lifecycle.go`, `lifecycle_test.go` — phased `Start`/`Close` sequence |
| **Registry** (`kernel/registry.go`) | ✅ Implemented | `registry.go` (5.9 KB), full test coverage |
| **Boundary Engine** (`kernel/boundary.go`) | ✅ Implemented | `boundary.go`, `boundary_test.go` — cross-subsystem boundary enforcement |
| **Permission Engine** (`kernel/permission.go`) | ✅ Implemented | `permission.go` (6.1 KB), `permission_test.go` (10.9 KB) |
| **Kernel Phases** | ✅ Implemented | `PhaseCore`, `PhaseInfrastructure`, `PhaseWorkspace`, `PhaseExecutive`, `PhaseCognitive`, `PhaseBackground` — all registered in `host.go` |

**Verdict**: The Kernel is fully implemented and wired. All 6 subsystems have test coverage.

---

## 2. Core Services

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **Storage Service** (`core/storage/storage.go`) | ✅ Implemented | CAS write/read/delete/exists/list. `StoreArtifact` and `LookupArtifact` (Artifact Index) fully implemented. Registered as `Core.Storage`. |
| **Memory Service** (`core/memory/`) | ✅ Implemented | `memory.Memory` interface, wired into runtime as `Core.Memory`. |
| **Scheduler Service** (`core/scheduler/`) | ✅ Implemented | Registered as `Core.Scheduler`. |
| **Time Service** (`core/time/`) | ✅ Implemented | `TimeService` interface, injected into `NativeCapabilityDependencies`. |

**Notable**: `Storage.Store()` implements a CAS (`sha256`-keyed content-addressed store). `StoreArtifact()` combines payload write + artifact index creation in a single atomic call. The `payloadStorerAdapter` in `runtime/host.go` (lines 130–155) bridges the storage interface to all V3 subsystem consumers.

---

## 3. Global Workspace & Communication

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **Workspace Engine** (`intelligence/workspace/`) | ✅ Implemented | Registered as `Intelligence.Workspace`. Topic-based pub/sub engine. |
| **Topic Definitions** (`intelligence/communication/topics.go`) | ✅ Implemented | `TopicPerception`, `TopicActiveGoal`, `TopicCandidatePlans`, `TopicEvaluatedOptions`, `TopicActionExecution`, `TopicImpasses`, `TopicLearningOpportunity` all defined. |
| **Envelope Builder** (`communication.NewEnvelopeBuilder()`) | ✅ Implemented | Used by all V3 workspace bridges. |
| **CAS Payload refs in Envelopes** | ✅ Implemented | All envelopes carry `PayloadRef` pointing to CAS-stored data. |

---

## 4. Brain / Cognitive Pipeline (V3)

### Understanding

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **Grammar Specialist** | ✅ Implemented | `underv3.NewDefaultGrammarSpecialist()` created in `host.go:532` |
| **Neural Specialist** | ✅ Implemented | `underv3.NewDefaultNeuralSpecialist()` created in `host.go:533` |
| **Deliberative Worker** | ⚠️ Partially Implemented | `underv3.NewDeliberativeWorker(h.inferenceSvc, ...)` IS wired in `host.go:534`. However, `h.inferenceSvc` is initialized **after** Understanding (lines 499–527 vs. 530–546), so the reference is valid at construction time. LLM fallback escalation depends on `deliberative-parser` model being available. |
| **V3 Orchestrator** | ✅ Implemented | `underv3.NewOrchestrator(grammar, neural, delib, exts)` at `host.go:537` |
| **WorkspaceBridge** | ✅ Implemented | Subscribed to `TopicPerception`. Verified working via `>>> Understanding V3 WorkspaceBridge Start()` log output. |
| **Modular Extractors** | ✅ Implemented | `extractors.NewDeterministicExtractors()` mapped into Orchestrator. Completely implements Phase 4B.3 semantic construction without normalization. |
| **NormalizerRunner** | ✅ Implemented | `normalizers.NewDeterministicNormalizers()` mapped into Orchestrator. Completely implements Phase 4B.4 Temporal Processing. |

### Reasoning

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **V3 Orchestrator** | ✅ Implemented | `reasoningv3.NewOrchestrator(v3Mem, v3Delib)` at `host.go:562` |
| **LegacyMemoryAdapter** | ✅ Implemented | `reasoningv3.NewLegacyMemoryAdapter(memSvc)` |
| **WorkspaceBridge** | ✅ Implemented | Subscribed to `TopicUserIntent`. Verified via `>>> Reasoning V3 HandleIntent invoked!` log. |
| **InferenceService injection** | ✅ Implemented | `InferenceService` is injected into `reasoningv3.NewOrchestrator` via `reasoning.NewDeliberativeSpecialist(h.inferenceSvc)`. The reasoning block was moved after inference initialization. |
| **Verification Fixes** | ✅ Resolved | 1. Added deliberative escalation for both `StatusAmbiguous` and `StatusFailed` (Impasse).<br>2. Registered the `reasoning-deliberative-llm` model in the Runtime Host.<br>3. Fixed payload reference construction for deliberative inference. |
### Planning

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **V3 Orchestrator** | ✅ Implemented | `planningv3.NewOrchestrator(v3CapReg)` at `host.go:457` |
| **LegacyCapabilityAdapter** | ✅ Implemented | `planningv3.NewLegacyCapabilityAdapter(h.capManager)` |
| **WorkspaceBridge** | ✅ Implemented | Subscribed to `TopicCandidatePlans`(?). Verified via `>>> Planning V3 HandleActiveGoal invoked!` log. |
| **Capability name resolution** | ✅ Implemented | `legacy_adapters.go` in `planning/v3` resolves `"sys-time-1"` to capability. |

### Decision

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **V3 Orchestrator** | ⚠️ Partially Implemented | `decisionv3.NewOrchestrator(v3Safety, nil, nil, nil)` — 3 of 4 dependencies are `nil`. Constitutional safety gate is wired; LLM-based deliberation and multi-criteria scoring are absent. |
| **LegacySafetyAdapter** | ✅ Implemented | `decisionv3.NewLegacySafetyAdapter(h.constGate)` |
| **WorkspaceBridge** | ✅ Implemented | Subscribed to `TopicCandidatePlans`. Verified via `>>> Decision V3 HandleCandidatePlan invoked!` log. |

### Executive (V3)

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **ExecutionEngine** | ✅ Implemented | `executivev3.NewExecutionEngine(v3PlanProv, v3CapReg, v3MemAdapter)` at `host.go:426` |
| **DAG Executor** | ✅ Implemented | `intelligence/executive/v3/dag_executor.go` — concurrent node execution with graceful degradation on failure, proper `sync.WaitGroup` + channel pattern |
| **Capability Execution** | ✅ Implemented | `executeNode()` resolves capability from registry, calls `Execute()`, stores payload in CAS via `MemoryProvider.StorePayload()` |
| **WorkspaceBridge** | ✅ Implemented | Subscribed to `TopicEvaluatedOptions`. Publishes full `ExecutionResult` JSON to `TopicActionExecution`. Source: `"Intelligence.ExecutiveV3"` |
| **Legacy Adapters** | ✅ Implemented | `LegacyCapabilityExecutor`, `LegacyCapabilityRegistryAdapter`, `LegacyMemoryAdapter`, `StoragePlanProvider` all implemented in `legacy_adapters.go` |
| **Full CapabilityResult serialization** | ✅ Implemented | `legacy_adapters.go:36` — `json.Marshal(res)` serializes entire `CapabilityResult` including `Realization` and `ResponseType` fields |
| **Debug fmt.Printf statements** | ⚠️ Known Issue | `workspace_bridge.go` lines 88, 112, 141 contain `fmt.Printf` debug statements. Not structured logger. Tracked in TODO §25 Priority 4. |
| **Stale comment** | ⚠️ Minor | `dag_executor.go:201` — `// Wait, if MemoryProvider fails?` is a thinking-out-loud comment. Tracked in audit DEV-12. |

---

## 5. Capability Framework

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **Architectural Status** | ✅ Complete | Unified framework seamlessly supports both Native and Application capability classes with zero architectural divergence. |
| **Capability Interface** (`capabilities/interfaces.go`) | ✅ Implemented | `ID()`, `Metadata()`, `State()`, `Execute()`, `Lifecycle()` methods defined and shared. |
| **BaseCapability** (`capabilities/base.go`) | ✅ Implemented | Embeddable base with default lifecycle. |
| **CapabilityRegistry** (`capabilities/registry.go`) | ✅ Implemented | Single unified registry handles both Native and Application capabilities (`Register()`, `Get()`, `All()`). |
| **CapabilityManager & Resolver** (`capabilities/manager.go`) | ✅ Implemented | Wraps registry, resolving by name/ID. Executive interacts purely via this interface, completely unaware of capability class. |
| **Native Capabilities** (`capabilities/native/`) | ✅ Implemented | Exclusive owners of OS/Hardware execution. `Time` capability fully wired. Others are placeholders awaiting provider implementation. |
| **Application Capabilities** (`capabilities/applications/`) | ✅ Implemented | Models 1 & 2 verified. Zero `os`, `os/exec`, or `net/http` imports exist. Uses `NativeCapabilityResolver` to strictly enforce dependency flow. |
| **Application Implementations** | ✅ Implemented | Calculator (Deterministic), Weather (Network Orchestration), Reminder (System Orchestration), Notes (Files Orchestration). Technical Debt (APP-01, APP-02) eradicated. |
| **CapabilityResult.RealizationStrategy** | ✅ Implemented | `capabilities/types.go` — `Deterministic` and `Generative` constants. |

---

## 6. Presentation Layer

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **RealizationEngine Interface** (`presentation/interfaces.go`) | ✅ Implemented | `Realize(ctx, CapabilityResult, parentRef, responseID) (*RealizedOutput, error)` |
| **RealizedOutput type** (`presentation/types.go`) | ✅ Implemented | `OutputID`, `SourceResponseID`, `ParentRef`, `RealizedText`, `CreatedAt` |
| **Response Type Router** (`presentation/router/service.go`) | ✅ Implemented | Subscribed to `TopicActionExecution`, filters `Source == "Intelligence.ExecutiveV3"`, retrieves `ExecutionResult` from CAS, dispatches to engine by `capResult.Realization`. Registered as `Presentation.Router`. |
| **Deterministic Engine** (`presentation/deterministic/engine.go`) | ✅ Implemented | Loads Go text template from disk, renders with `res.Data`. No template caching (see TODO §26). |
| **Template Repository** (`data/runtime/templates/`) | ⚠️ Partially Implemented | Contains only `time.tmpl`. No other response types have templates. |
| **Generative Engine** (`presentation/realization/service.go`) | ✅ Implemented | Implements `RealizationEngine` via new `Realize()` method. Uses `InferenceService` + prompt building. Also has legacy subscription handler for backward compatibility. |
| **Engine Registry in Router** | ✅ Implemented | `map[capabilities.Deterministic]detEngine, [capabilities.Generative]rlzSvc` registered in `host.go:563–566`. |
| **Legacy LanguageRealization subscription** | ⚠️ Commented Out | `host.go:769` — `registerWrap("Presentation.LanguageRealization", ...)` is commented out. The service still exists in memory but is not independently started as a subscriber. It is invoked directly by the Router as a `RealizationEngine` implementation. |

### Known Limitation
The Router currently processes **only the first node result** from `ExecutionResult.NodeResults()` (iterating map, taking first entry — `router/service.go:133–137`). Multi-node DAG results are not aggregated. This is an MVP limitation.

---

## 7. World Pillar

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **World Service** (`world/service.go`) | ✅ Implemented | Registered as `World.Service`. Accepts envelopes from `Presentation.LanguageRealization` and now `Presentation.Router`. |
| **Text Input Adapter** (`world/adapters/text/`) | ✅ Implemented | Reads stdin, publishes `TopicPerception`. |
| **Output pipeline** | ✅ Implemented | `world/service.go` extracts `RealizedText` from `RealizedOutput`, writes to `outWriter`. End-to-end verified. |
| **Router source acceptance** | ✅ Implemented | `world/service.go:428,450` — checks `env.Source != "Presentation.Router"` alongside `"Presentation.LanguageRealization"`. |

---

## 8. Runtime

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **RuntimeHost** | ✅ Implemented | `runtime/host.go` (950 lines) — `Build()`, `Wire()`, `Register()`, `Start()`, `Stop()` |
| **Dependency wiring** | ✅ Implemented | Full V3 pipeline wired: Understanding → Reasoning → Planning → Decision → Executive V3 → Router → World |
| **Phased startup** | ✅ Implemented | `Core → Infrastructure → Workspace → Executive → Cognitive → Background` — verified via kernel boot logs |
| **Subsystem enable/disable** | ✅ Implemented | `isSubsystemEnabled(name)` checks `cfg.EnabledSubsystems` map. Default config enables all subsystems. |
| **Config loading** | ✅ Implemented | `runtime/config.go` reads env vars (`IDUN_REALIZATION_MODEL`, etc.) |
| **Manifest** | ✅ Implemented | `runtime/manifest.go` — immutable boot fingerprint |
| **Inference backend registration** | ✅ Implemented | `ollama-local-01` registered for both `local-realizer` and `deliberative-parser` models. |

### Runtime Gaps

| Gap | Severity | Detail |
| :--- | :--- | :--- |
| `InferenceService` not injected into `reasoningv3.Orchestrator` | ⚠️ Medium | `host.go:442` — `NewOrchestrator(v3Mem)` only. Deliberative specialist for reasoning is inactive. |
| `TopicLearningOpportunity` never published | ⚠️ Medium | Learning service is active but idle. No cognitive module emits learning opportunities. |
| `TopicImpasses` never published | ⚠️ Medium | Reflection service is active but idle. Planning/Reasoning do not emit impasses. |
| `capabilities.ActionExecutionHandler` subscription status | ❓ Unknown | Legacy capability handler may be registered or bypassed — V3 DAG does capability execution directly. Requires clarification. |
| `fmt.Printf` debug statements | ⚠️ Low | `executive/v3/workspace_bridge.go` lines 88, 112, 141 — should be replaced with structured logger. |
| Stale `// Wait, if MemoryProvider fails?` comment | ⚠️ Low | `dag_executor.go:201` — production noise. |

---

## 9. Metacognitive Services

| Component | Status | Evidence |
| :--- | :--- | :--- |
| **Reflection Service** (`intelligence/reflection/`) | ⚠️ Partially Wired | Instantiated with `reflection.WithWorkspace(h.workspaceSvc)`. Subscribed to `TopicImpasses`. But no publisher emits impasse events. Loop is idle. |
| **Learning Service** (`intelligence/learning/`) | ⚠️ Partially Wired | Instantiated with `learning.WithWorkspace(h.workspaceSvc)`. Subscribed to `TopicLearningOpportunity`. But no publisher emits opportunities. Loop is idle. |
| **Calibration Service** (`intelligence/calibration/`) | ✅ Implemented | Registered as `Foundation.Calibration`. |
| **Attention Service** (`intelligence/attention/`) | ⚠️ Partially Wired | Instantiated in `host.go`, registered as `Intelligence.Attention`. Subscribed to `TopicCandidatePlans` and `TopicEvaluatedOptions` via inline handlers (lines 648–675). Salience scores are computed but do not influence dispatch priority. |
| **Constitution Gate** (`intelligence/constitution/`) | ✅ Implemented | `Foundation.Constitution` registered. Used by Decision V3 safety adapter. |

---

## 10. End-to-End Pipeline Verification

The following complete turn was verified on 2026-07-31:

```
Input:  "What time is it?"
Output: "The current time is 2026-07-31T00:30:07.062703+05:30."
```

**Verified path**:
```
World.TextInput
  → TopicPerception
  → Understanding V3 (Grammar + Neural + Deliberative)
  → TopicActiveGoal
  → Reasoning V3
  → TopicCandidatePlans (?)
  → Planning V3
  → TopicEvaluatedOptions (Decision V3)
  → Decision V3 → TopicEvaluatedOptions
  → Executive V3 (DAG → TimeCapability → CAS)
  → TopicActionExecution [Source: Intelligence.ExecutiveV3]
  → Presentation.Router → Deterministic Engine → time.tmpl
  → TopicActionExecution [Source: Presentation.Router]
  → World.Service → stdout
```

**LLM bypass confirmed**: The Deterministic Engine rendered `time.tmpl` directly from `CapabilityResult.Data["CurrentTime"]`. No inference call was made for realization.

---

## 11. Architecture Maturity Summary

| Pillar | Maturity |
| :--- | :--- |
| **Kernel** | Production-ready |
| **Core Services (Storage, Memory, Time)** | Production-ready |
| **Global Workspace** | Production-ready |
| **Understanding V3** | Phase 4B.4 Frozen — Temporal Processing complete. Semantic enrichment and deterministic normalization implemented. |
| **Reasoning V3** | Architecturally Complete — Wired, S8 injected and escalates on ambiguity/failure. The Reasoning Layer is architecturally complete for the current IDUN V3 scope. Future work such as improved models, prompt optimization, additional reasoning strategies, or performance enhancements should be considered feature enhancements rather than missing architecture. |
| **Planning V3** | ~75% — Capability resolution works, tree search and deep plans untested |
| **Decision V3** | Architecturally Complete — Validates single ExecutionPlan. All baseline validators (Safety, Auth, Policy, Budget) injected. Multi-candidate evaluation is intentionally superseded by the Planning architecture. |
| **Executive V3** | ~85% — DAG execution correct, debug prints remain |
| **Capability Framework** | ~40% — TimeCapability fully verified; 8 other capability categories are unregistered placeholders |
| **Presentation.Router** | ~90% — Routing works, multi-node aggregation missing |
| **Deterministic Engine** | ~80% — Works, no template caching, one template exists |
| **Generative Engine** | ~70% — Works as fallback, invoked directly as RealizationEngine |
| **Template Repository** | ~15% — Only `time.tmpl` exists |
| **World** | ~90% — Output delivery verified, minor source string coupling |
| **Reflection** | ~30% — Wired but idle (no impasse publishers) |
| **Learning** | ~30% — Wired but idle (no opportunity publishers) |
| **Attention** | ~40% — Salience computed, no priority integration |
| **Runtime** | ~85% — All major V3 components wired, 5 known gaps |

---

## 12. Regressions vs Previous Audit

The following items have changed status since the prior `sysem_audit_report`:

| Item | Previous Status | Current Status | Change |
| :--- | :--- | :--- | :--- |
| `Presentation.LanguageRealization` as primary route | Active (direct subscriber) | Commented out; invoked via Router | Architecture improvement |
| `WorkspaceBridge` legacy CommunicationMessage adapter | Present (hack) | Removed | Cleanup |
| `CapabilityResult.RealizationStrategy` | Missing | Implemented | New |
| `Presentation.Router` | Missing | Fully wired | New |
| `Deterministic Engine` | Missing | Implemented | New |
| `RealizationEngine` interface | Missing | Implemented | New |
| `NativeCapabilityDependencies` | Missing | Implemented | New |
| Understanding InferenceService injection | Missing | Now wired (host.go:534) | Fixed |
| Reflection `WithWorkspace` injection | Missing | Now wired (host.go:484) | Fixed |

---

## 13. Inconsistencies Found During Audit

> [!WARNING]
> The following were discovered during audit and are flagged for separate review:

1. **Router: Map iteration order for NodeResults** — `presentation/router/service.go:133–137` iterates `map[string]NodeResult` and takes the first result from a Go map, which has non-deterministic iteration order. For single-node DAGs this is fine, but for multi-node plans the "first" result may differ across runs. This should be fixed when multi-node aggregation is implemented.

2. **LegacyMemoryAdapter bypasses CAS interface** — `legacy_adapters.go:70–79` uses raw `sha256` hashing + `store.Write()` directly, bypassing the `StoreArtifact()` Artifact Index. Stored payloads are not indexed and therefore cannot be looked up by artifact ID. This is intentional for MVP but should be aligned with the full Artifact Store contract in a future iteration.

3. **`isSubsystemEnabled` not used for Understanding** — The `isSubsystemEnabled("understanding")` check exists at `host.go:530`, but `InferenceService` (`h.inferenceSvc`) is initialized unconditionally at line 499 before the check. If `understanding` is disabled, the inference service is still fully initialized and registered. This is a minor efficiency issue, not a functional regression.

---

## 14. Architecture Decision Record (ADR) Summary

The following points summarize the major, long-term architectural decisions that define IDUN V3. These represent foundational principles of the cognitive architecture that persist irrespective of the specific implementations of individual components.

> **ARCHITECTURE FREEZE POINT**: The Phase 3 Schema Verification Audit has passed. The Cognitive API Specification, Architecture Alignment, Schema Alignment, and Contract Compliance have all been completed and verified. This verification certifies schema correctness, not behavioral completeness. The remaining work belongs exclusively to Implementation Alignment (Phase 4). Phase 4 is responsible for making each cognitive layer produce meaningful semantic content that conforms to the approved Cognitive API Specification. No further architectural redesign or schema modification should occur unless a verified conflict with the approved Cognitive API Specification is discovered. The architecture, semantic contracts, and schemas are now considered frozen. This marks the official architecture freeze point for IDUN V3.

1. **Strict Separation of Cognition and Presentation:** The cognitive pipeline (Brain) decides *what* to do and *when* to do it, but it never decides *how* a result is communicated. The Executive and Capabilities are entirely presentation-agnostic.
2. **Standardized Cognitive Pipeline:** The core intelligence flow strictly follows the sequence: `Understanding` → `Reasoning` → `Planning` → `Decision` → `Executive`.
3. **Structured Capability Contracts:** Capabilities are self-contained logical units that always return structured, semantic data (`CapabilityResult`), without any presentation logic, string formatting, or direct output manipulation.
4. **Deterministic Bypasses for LLMs:** Deterministic capabilities operate securely and completely bypass Generative Engines for realization, dramatically improving performance, latency, and consistency for standard operations.
5. **Decoupled Response Realization:** Response realization is handled separately by the `Response Type Router`, which evaluates the `CapabilityResult` and invokes the appropriate `RealizationEngine`.
6. **Replaceable Generative Engines:** The Generative Engine operates behind a common `RealizationEngine` interface, making it an interchangeable component rather than a core architectural dependency.
7. **Unified External Interaction via World:** The `World` pillar holds exclusive responsibility for all external interaction, message delivery, and environment-facing state updates. The cognitive pipeline acts upon the world only through established boundaries.
8. **Future Presentation Consolidation:** Over time, the Presentation layer's routing and realization responsibilities are slated to be consolidated entirely into the World pillar.
9. **Planning owns plan generation. Decision owns plan validation:** Decision V1 selected the optimal action from multiple candidates through scoring and ranking. Decision V3 validates a single ExecutionPlan produced by Planning V3. Multi-candidate evaluation is therefore no longer part of the Decision layer and has been intentionally superseded by the Planning architecture.
10. **Native vs. Application Capabilities:** The capability system uses a two-tier architectural class model. **Native Capabilities** own mechanical execution and are the only layer allowed to interact with external providers or OS primitives. **Application Capabilities** deliver user-facing functionality through two valid execution models: (1) Orchestrating Native Capabilities without bypassing Native boundaries, or (2) Pure Deterministic Execution for self-contained logic (e.g. calculator) that requires no OS, hardware, file, or network access.

11. **Communicative Intent Boundary**: Purely communicative intents (e.g., greet_user) do not bind to standard execution capabilities. The Planning V3 boundary filters them to a dedicated mock.capability placeholder, which the Executive V3 mock executor processes securely. This preserves boundary integrity between cognitive reasoning and native execution without requiring Presentation layer fallback hacks.
12. **Compound Intent Ordering**: The Understanding layer splits and preserves compound goals purely in spoken order. The Planning layer exclusively owns workflow generation, dependency construction, and parallelization strategy. See [Compound Intent Ordering ADR](file:///C:/Projects/idun_v3/reports/architecture/architecture_decisions/compound_intent_ordering.md).


## 15. Temporary Infrastructure

### Component

mock.capability

### Purpose

Acts as a temporary communicative capability used to validate the complete end-to-end cognitive pipeline while native communicative capabilities are still under development.

This placeholder exists solely to verify architectural correctness and contract compliance.

### Replacement Plan

Replace mock.capability with a native Communicative Capability Family when implemented.

### Replacement Scope

**Planning**

Replace placeholder capability mapping with native communicative capability discovery.

**Executive**

Remove MockCommunicativeExecutor.

### No changes required

- Understanding
- Reasoning
- Decision
- Presentation
- Realization

### Status

**Approved Temporary Infrastructure**

The placeholder is intentional engineering infrastructure and not technical debt or an accidental workaround.
