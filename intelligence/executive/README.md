# IDUN Intelligence Pillar: Executive Functions Version 2.0 (`idun/intelligence/executive`)

**Architecture Version:** `2.0.0-FROZEN`  
**Classification:** Core Executive Control Plane — Phase 5 Upgrade  
**Status:** FULLY IMPLEMENTED & TESTED (`-race` CLEAN)

---

## 1. Purpose & Version 1 → Version 2 Upgrade Summary
The `executive` package coordinates the seven immutable cognitive abilities.
- **Version 1 Heritage Preserved:** All existing V1 structures (`Service`, `WorkflowGraph`, `AttentionGate`, `PriorityEngine`, `BudgetManager`, `AbilityRegistry`) remain intact and are embedded within `ServiceV2`.
- **Version 2.0 Capabilities Added:**
  - **Global Workspace Integration (`Workspace()`):** Symmetric publish/subscribe arbitration across leveled topic channels.
  - **Epistemic Calibration Integration (`Calibration()`):** Evaluates candidate bids using Calibrated Effective Priority ($P_{\text{eff}}$), discounting overconfident modules.
  - **Constitutional Gate Integration (`Constitution()`):** Routes actions targeting `TopicActionExecution` through the Pre-Broadcast Constitutional Action Gate before physical execution.
  - **Multi-Horizon Scheduling (`Horizon`):** Decouples Reflexive ($<15\text{ms}$), Deliberative ($100\text{--}500\text{ms}$), and Background scheduling.
  - **SOAR-Style Content-Blind Impasse Emission:** Automatically publishes `TopicImpasses` events when no candidate bid meets admission threshold $\tau$.

---

## 2. Strict Content Blindness & Separation of Responsibilities
Executive Functions Version 2.0 inspects **only** control-plane Envelope metadata (`Source`, `Topic`, `RawConfidence`, `Urgency`, `CostEstimateUnits`, `PayloadRef`).
Executive **never** dereferences `PayloadRef`, **never** parses language, and **never** performs domain reasoning or planning.

---

## 3. Phase 3 Episode Orchestration (`2.0.0-FROZEN`)
Executive Phase 3 introduces **Episode-Based Orchestration**, dividing historical identity from runtime execution state:
- **`ExecutiveEpisodeDefinition` (Immutable):** Stores `EpisodeID`, `EpisodeType`, `EpisodeIntent`, `EpisodeOrigin`, `HierarchyReference` (`Workspace` owned), `DependencyReference` (`Workspace` owned), `EpisodeFingerprint` (`SHA-256`), and `ReplayMetadata`.
- **`ExecutiveEpisodeRuntime` (Mutable):** Tracks FSM Factual transitions (`EpisodeStatus` vs. `EpisodeOutcome`), `ExecutorID` (migration location), and bounded rolling histories (`PriorityHistory`, `BudgetHistory` bounded to max 16 items).
- **`Hybrid EpisodeContext`:** Strongly typed core (`WorkspaceReference`, `AttentionReference`, `GoalReference`) plus an extensible `ModuleReferences map[string]string` registry (`planning://...`, `vision://...`, `robotics://...`, `skills://...`).
- **`EpisodeOrchestrator`:** Event-driven engine reacting to asynchronous signals (`EventDependencyResolved`, `EventPlanningCompleted`, `EventDecisionCompleted`, `EventAttentionChanged`, `EventEpisodeCancelled`) and coordinating cognitive waking (`Reflection`, `Learning`, `Strategy Activation`) without evaluating or generating domain content.
- **`EpisodeCheckpoint` (Minimal Layer 1 Snapshot):** Immutable recovery snapshot containing `CheckpointID`, `EpisodeID`, `RuntimeFingerprint`, `WorkspaceReference`, `AttentionReference`, `Timestamp`, and `ReplayMetadata`.

---

## 4. Future Executive Enhancements (Post Layer 1)
The following features are documented as future design ideas for expansion **beyond** Layer 1 (`2.0.0-FROZEN`) and are **not currently implemented**:

### `EpisodeCheckpoint v2`
Possible future additions to recovery snapshots:
- `CheckpointVersion`
- `CheckpointFingerprint`
- `CreatedByExecutor`
- `CheckpointReason`

### `EpisodeCapabilities Extensions`
Possible future additions to capability advertising:
- `SupportsNestedEpisodes`
- `SupportsDistributedExecution`

### `EpisodeMetrics`
Potential future telemetry tracking:
- `CPUTime`
- `WallTime`
- `MemoryUsed`
- `PauseCount`
- `ResumeCount`
- `MigrationCount`
- `CheckpointCount`

### `EpisodeSubsystemUsage Enhancements`
Possible future additions to subsystem monitoring:
- Invocation counters across all cognitive turns
- Aggregate execution statistics and latency percentiles across distributed worker nodes
