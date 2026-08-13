# IDUN BLUEPRINT
## Engineering Master Plan: From V3 to the North Star

---

> **Document Classification:** Engineering Master Plan — Blueprint
> **Status:** UNDER ACTIVE ARCHITECTURAL REVIEW — Not frozen. Sections are progressively reviewed and locked.
> **Position in Hierarchy:** Below `IDUN_NORTH_STAR.md` · Above Architecture Specifications
> **Relationship to Existing Architecture:** Extends, does not replace, existing `2.0.0-FROZEN` contracts
> **Date:** 2026
> **Decision 0.1 Status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE — 6 implementation constraints added. (3 policy-level open questions remain; see §0.1.O)

```
IDUN_NORTH_STAR.md         Where IDUN ultimately needs to go
        │
        ▼
IDUN_BLUEPRINT.md          How IDUN must be designed to get there   ← THIS DOCUMENT
        │
        ▼
Architecture Specs         How individual subsystems are engineered
        │
        ▼
Implementation Plans       What gets built in each phase
        │
        ▼
Code                       The actual implementation
        │
        ▼
Tests / Validation         Verification that the implementation is correct
```

---

## Part 0 — Critical Architecture Decisions

Before defining the architecture, the following decisions from the Cognitive Operating Workflow are explicitly reviewed, challenged, and resolved.

---

### Decision 0.1 — Gap Classifier: Subsystem or Policy?

**Review Status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE
The core architectural decisions for Decision 0.1 are fully resolved (see §0.1.N). Six implementation constraints have been added based on final adversarial review. Three policy-level open questions remain (see §0.1.O). Decision is frozen.

---

#### 0.1.A — Original Proposal

The Cognitive Operating Workflow proposed a `Gap Classifier` as a distinct subsystem with its own package, goroutines, and lifecycle.

#### 0.1.B — Core Decision (Confirmed)

A dedicated `Gap Classifier` subsystem is **rejected**.

A standalone package with its own lifecycle:

```
GapClassifier/
    service.go
    lifecycle.go
    worker.go
```

is not justified. Gap classification is not a separate cognitive act — it is an inherent output of the subsystem that detects the gap. There is no cognitive work for a classifier to perform that has not already been performed by the detecting subsystem.

#### 0.1.C — Gap Classification Ownership (Confirmed)

The four gap types have **different owners** at their point of detection. Reasoning does not classify all gaps.

```
           GAP DETECTION AND CLASSIFICATION

  Understanding ──────► ClarificationGap
  (Context Resolver)          │
                              │
  Reasoning ──────────► KnowledgeGap
                              │
                              │
  Planning ───────────► SkillGap
                              │
                              │
  Constitutional Gate ─► AuthorizationGap
                              │
                              ▼
                          GapSignal
                    (typed, in contracts)
```

| Gap Type | Detecting / Classifying Owner | Meaning |
|:---------|:------------------------------|:--------|
| `ClarificationGap` | Context Resolver / Understanding | IDUN cannot determine what the Host meant |
| `KnowledgeGap` | Reasoning | IDUN lacks the knowledge required to proceed |
| `SkillGap` | Planning | IDUN knows what needs doing but lacks the required capability |
| `AuthorizationGap` | Constitutional Gate | IDUN understands the action but is not authorized to perform it |

Each gap type is emitted by the stage that has sufficient context to make the determination. No downstream classifier is needed.

> **RESOLVED (§0.1.J):** `ClarificationGap` always routes through Goal Manager. There is no bypass path. Goal Manager uses a two-tier wait model: blocking goals create a persisted `GapRecord`; non-blocking clarifications create a transient `PendingClarification` token in Working Memory. Conversation Planner generates the clarification question in both cases.

#### 0.1.D — Event Bus Remains Content-Blind (Confirmed)

The Event Bus transports gap signals. It does not inspect them, classify them, or route them based on their type.

```
  Gap-producing subsystem
          │
          ▼
       GapSignal
    (typed envelope)
          │
          ▼
  Workspace Event Bus
    (content-blind:
     transports, does not inspect)
          │
          ▼
     Goal Manager
  (authoritative recipient)
```

The Event Bus must NOT contain logic such as:

```
// PROHIBITED — cognitive routing in infrastructure
if signal.Type == KnowledgeGap {
    route to KnowledgeAcquisition
}
```

Infrastructure transports. Goal Manager decides.

#### 0.1.E — Goal Manager is the Gap Lifecycle Owner (Confirmed)

The Goal Manager is the authoritative owner of the complete gap resolution lifecycle.

```
                    Goal Manager
                         │
          ┌──────────────┼──────────────┐
          │              │              │
        PAUSE         DISPATCH       MONITOR
          │              │              │
          │              ▼              │
          │    Acquisition Subsystem    │
          │    (executes resolution)    │
          │              │              │
          │              ▼              │
          │          Resolution         │
          │          Outcome            │
          │              │              │
          └──────────────┴──────────────┘
                         │
                ┌────────┼────────┐
                ▼        ▼        ▼
             SUCCESS   FAILURE  TIMEOUT
                │        │        │
                ▼        ▼        ▼
             RESUME    RETRY/  CANCEL +
              GOAL      FAIL   ESCALATE
```

**Goal Manager owns:**
- Pausing the goal when a gap is detected
- Creating and persisting the `GapRecord`
- Dispatching the gap to the appropriate acquisition subsystem
- Monitoring the resolution deadline (Level 2 timeout)
- Receiving the resolution result from the acquisition subsystem
- Deciding retry / failure / escalation
- Cancelling in-flight acquisition when necessary
- Scheduling the Continuation Episode on success
- Resuming the goal
- Permanently failing the goal when all retries are exhausted
- Requesting Host intervention via the communication layer

**Goal Manager does NOT perform the acquisition itself.** It is a lifecycle conductor, not an acquisition engine.

**F1 — GapRecord write-before-acknowledge**
Goal Manager must persist the GapRecord successfully before the corresponding GapSignal is acknowledged/consumed from the Event Bus. This prevents a crash between signal consumption and durable state creation from permanently losing a gap.

```
  ═══════════════════════════════════════════════════════════════
  SAFE GAP SIGNAL PERSISTENCE ORDERING (Diagram A)
  ═══════════════════════════════════════════════════════════════
  GapSignal
      │
      ▼
  Goal Manager
      │
      ▼
  Persist GapRecord
      │
      ├── FAIL → do not acknowledge signal
      │
      └── SUCCESS
           │
           ▼
        Dispatch
           │
           ▼
      Acknowledge signal
```

**F11 — Multiple in-flight GapSignals**
Goal Manager must safely handle multiple GapSignals arriving for an already-paused goal. The implementation may queue or safely discard duplicates according to the eventual runtime policy, but it must never corrupt lifecycle state or create inconsistent duplicate GapRecords.

#### 0.1.F — No Dedicated Gap Router Subsystem (Confirmed)

A separate `GapRouter` service is not justified. Dispatch from Goal Manager to the appropriate acquisition subsystem is implemented as a **pure routing function embedded within Goal Manager**.

```
  Goal Manager
       │
       └── routeGap(GapType) DispatchTopic
                  │
          ┌───────┼────────┐
          ▼       ▼        ▼
  KnowledgeGap  SkillGap  AuthorizationGap
       │          │              │
  TopicKnow-  TopicSkill-  TopicAuthorization-
  ledgeGap    GapRequested  Required
  Requested
```

This routing function:
- Has no state
- Has no goroutines
- Has no lifecycle
- Does not perform acquisition
- Does not monitor timeouts
- Is a pure deterministic mapping (GapType → Topic)
- Is colocated in the `intelligence/goalmanager` package

Adding a new gap type requires: adding to the `contracts.GapType` enum, adding one case to this function, and creating the new acquisition service. Three isolated changes, no cascade.

#### 0.1.G — Five-Level Timeout and Deadline Model (Refined — OQ-0.1-3 resolved)

The earlier two-level model was confirmed for acquisition gaps. The architectural review of OQ-0.1-3 revealed three additional concepts that must be formally distinguished. There are now **five distinct timeout/deadline concepts** with **five distinct owners or policies**. None of these may be conflated.

```
  ═══════════════════════════════════════════════════════════════
  FIVE TIMEOUT / DEADLINE CONCEPTS — ALL DISTINCT
  ═══════════════════════════════════════════════════════════════

  Level 1   Operation Timeout
  ──────────────────────────
  Scope:     Single acquisition request (HTTP, tool, API)
  Owner:     Acquisition Subsystem (internal)
  Mechanism: Go context.WithDeadline — never visible to Goal Manager
  Applies:   KnowledgeGap, SkillGap acquisition only

  Level 2   Acquisition Resolution Timeout  ← GapRecord.Deadline
  ──────────────────────────────────────────
  Scope:     Entire gap resolution lifecycle for an acquisition gap
  Owner:     Goal Manager exclusively
  Mechanism: Persistent GapRecord.Deadline — survives restarts
  Applies:   KnowledgeGap, SkillGap ONLY
             NOT applicable to ClarificationGap or AuthorizationGap

  Level 3   Goal.ExpiresAt  (optional per-goal hard deadline)
  ──────────────────────────────────────────────────────────
  Scope:     Entire goal, any state
  Owner:     Goal Manager
  Mechanism: Optional field set at goal creation; Goal Manager
             archives goal when ExpiresAt is reached regardless of
             current state
  Example:   "Book a flight before 18:00" → ExpiresAt = 18:00
  Default:   None (most goals have no hard deadline)

  Level 4   Host Staleness Reminder  (policy-driven)
  ──────────────────────────────────────────────────
  Scope:     Goal in AWAITING_HOST_RESPONSE state
  Owner:     Goal Manager reads policy from SS-32 (Runtime Policy)
  Trigger:   ReminderInterval elapsed since last Host interaction
  Effect:    Re-emit the pending clarification/authorization request
  Default:   Configurable per priority band — not hardcoded

  Level 5   Host Staleness Archive  (policy-driven, optional)
  ──────────────────────────────────────────────────────────
  Scope:     Goal in AWAITING_HOST_RESPONSE state past stale threshold
  Owner:     Goal Manager reads policy from SS-32 (Runtime Policy)
  Trigger:   ArchiveThreshold elapsed since last Host interaction
  Effect:    Goal marked STALE → archived (does not auto-expire goal
             if ExpiresAt is not set)
  Default:   Configurable, OFF by default
```

| Level | Name | Owner | Applies To | Mechanism |
|:------|:-----|:------|:-----------|:----------|
| **1** | Operation Timeout | Acquisition Subsystem | Acquisition requests | `context.WithDeadline` (internal) |
| **2** | Acquisition Resolution Timeout | **Goal Manager** | KnowledgeGap, SkillGap | `GapRecord.Deadline` (persisted) |
| **3** | Goal Hard Deadline | **Goal Manager** | Any goal (optional) | `Goal.ExpiresAt` (optional field) |
| **4** | Staleness Reminder | **Goal Manager** + SS-32 policy | Goals in `AWAITING_HOST_RESPONSE` | Configurable interval per priority band |
| **5** | Staleness Archive | **Goal Manager** + SS-32 policy | Goals in `AWAITING_HOST_RESPONSE` | Configurable threshold, off by default |

**Key invariants:**

> Level 2 (`GapRecord.Deadline`) applies **only** to acquisition-type gaps (KnowledgeGap, SkillGap). ClarificationGap and AuthorizationGap wait for a human — they do not have an acquisition deadline.
>
> `AWAITING_HOST_RESPONSE` does NOT have a universal mandatory maximum lifetime. It is governed by the Host Staleness Policy (Levels 4–5), not a hard timeout.
>
> `Goal.ExpiresAt` and `GapRecord.Deadline` are orthogonal. A goal may have an ExpiresAt that fires while acquisition is still in progress; Goal Manager archives the goal and cancels the acquisition.

```
  AWAITING_HOST_RESPONSE — Lifecycle

  AWAITING_HOST_RESPONSE
         │
         ├─── Level 4: ReminderInterval elapsed
         │         │
         │         ▼
         │    Goal Manager re-emits pending request
         │    via Conversation Planner
         │
         ├─── Level 5: ArchiveThreshold elapsed (if configured)
         │         │
         │         ▼
         │    Goal marked STALE → optional archive
         │
         ├─── Level 3: Goal.ExpiresAt reached
         │         │
         │         ▼
         │    Goal archived regardless of AWAITING state
         │
         └─── Host responds
                   │
                   ▼
              RESOLVED → Continuation Episode (if goal-blocking)
              OR next-turn normal resolution (if transient)
```

**Policy values (genuinely open — do NOT invent defaults):**
- Exact `ReminderInterval` per priority band: policy decision, owned by SS-32
- Exact `ArchiveThreshold`: policy decision, owned by SS-32, OFF by default
- These must be resolved as configuration design decisions, not hardcoded here

#### 0.1.H — GapRecord State Machines — Two Branches (Refined — OQ-0.1-4 partially resolved)

The single generic GapRecord state machine is **architecturally incorrect** when applied uniformly to all gap types. ClarificationGap and AuthorizationGap wait for a human response — they must not enter the acquisition retry/timeout branch. The GapRecord state machine is split into two explicit branches.

**Branch A — Acquisition Gaps (KnowledgeGap, SkillGap)**

```
  CREATED
     │ Goal Manager dispatches via routeGap()
     ▼
  DISPATCHED
     │ Acquisition subsystem acknowledges
     ▼
  RESOLVING ─────────────────────────── Level 2 deadline active (GapRecord.Deadline)
     │              │                         │
     ▼              ▼                         │ deadline exceeded
  SUCCESS        FAILURE                      ▼
     │              │                   TIMED_OUT
     │       RetryCount < Max            │
     │              │            Goal Manager sends CancelGap
     │              ▼            to Acquisition subsystem
     │           RETRYING               │
     │           (backoff)              │ RetryCount < Max
     │              │                   │
     │              ▼◄──────────────────┘
     │           RESOLVING
     │           (loop)
     │
     │       RetryCount ≥ Max (or Goal.ExpiresAt reached)
     │              │
     │              ▼
     │    HOST_INTERVENTION_REQUIRED
     │    (Goal → AWAITING_HOST)
     │    (Level 4/5 staleness policy now applies)
     │
     ▼
  RESOLVED
  (Goal Manager schedules Continuation Episode)
  (Goal → ACTIVE resumed)

  ─────────────────────────────────────────────
  At any state, if goal is cancelled by Host:

  DISPATCHED / RESOLVING / RETRYING / TIMED_OUT
     │ Goal cancelled
     ▼
  CANCELLED
  (Goal Manager sends CancelGap to Acquisition)
  (Acquisition terminates cleanly)
  (No continuation episode scheduled)
```

**F12 — CANCELLED is terminal**
Once a Goal enters CANCELLED, it is terminal. Any subsequent GapResolved, including SUCCESS, must be ignored for that cancelled goal. A successful acquisition must never resurrect a Host-cancelled goal.

```
  ═══════════════════════════════════════════════════════════════
  CANCELLATION TERMINAL STATE (Diagram E)
  ═══════════════════════════════════════════════════════════════
  Goal
   │
   ├── CANCELLED ──► TERMINAL
   │                    │
   │                    └── later GapResolved → IGNORE
   │
   └── active ──► normal lifecycle
```

**On restart — Acquisition gap recovery (see 0.1.I for detail):**
```
  DISPATCHED or RESOLVING found in persisted state
     │
     ▼
  Treat as UNCERTAIN
     │
     ▼
  Re-dispatch same GapID to Acquisition
     │
     ▼
  Acquisition checks output store (idempotency)
     ├── Result exists → GapResolved(SUCCESS)
     └── Result absent → begin acquisition
```

**Branch B — Host-Response Gaps (ClarificationGap, AuthorizationGap)**

**F5 — AuthorizationGap path distinction**
The transient `PendingClarification` path is exclusive to ClarificationGap. AuthorizationGap always originates from the Constitutional Gate on an active committed goal and therefore always creates a persisted GapRecord.

```
  ═══════════════════════════════════════════════════════════════
  AUTHORIZATIONGAP VS CLARIFICATIONGAP (Diagram D)
  ═══════════════════════════════════════════════════════════════
  ClarificationGap ──► Goal Manager
                          │
                     ┌────┴────┐
                     ▼         ▼
               Persisted    PendingClarification
               GapRecord    (transient only)

  AuthorizationGap ──► Goal Manager
                          │
                          ▼
                     Persist GapRecord
                     (always persistent)
```

```
  CREATED
     │ Goal Manager decides: blocking active goal?
     │
     ├── YES: persist GapRecord
     │
     └── NO:  transient PendingClarification token in Working Memory
              (no persistent GapRecord; Working Memory owns token)
     │
     ▼ (persistent path only)
  AWAITING_HOST_RESPONSE
     │
     │ [No Level 2 acquisition timeout — does not apply]
     │ [No acquisition dispatch — does not apply]
     │ [No retry loop — does not apply]
     │ [Level 4 staleness reminder may apply]
     │ [Level 5 staleness archive may apply]
     │
     ├── Host responds
     │       │
     │       ▼
     │   RESOLVED
     │   (persistent path: Goal Manager schedules Continuation Episode)
     │   (transient path: Working Memory token consumed; next turn resolves)
     │
     └── Goal cancelled
             │
             ▼
         CANCELLED
         (No acquisition to cancel)
         (Working Memory token cleared if transient)
```

**Critical invariants for Branch B:**
- No acquisition subsystem is involved
- No Level 2 deadline is set
- No retry loop executes
- Conversation Planner generates the question/request
- On restart with persisted AWAITING_HOST_RESPONSE: Goal Manager re-emits the request via Conversation Planner (not a store check)

#### 0.1.I — Restart and Recovery (Refined — OQ-0.1-2 resolved)

Gap resolution state must survive IDUN process restarts. The recovery model is based on the **Saga pattern**: the output store (not the operation log) is the authoritative truth for completion.

**Minimum persistent state (stored in `core/storage`):**

| Record | Fields |
|:-------|:-------|
| `Goal` | `GoalID`, `State`, `Priority`, `CreatedAt`, `ExpiresAt` (optional) |
| `GapRecord` | `GapID`, `GoalID`, `Type`, `Domain`, `Status`, `StartedAt`, `Deadline` (acq. gaps only), `RetryCount` |
| `PlanCheckpoint` | `GoalID`, serialized paused plan state |

**Recovery procedure — Acquisition gaps (KnowledgeGap, SkillGap):**

```
  ═══════════════════════════════════════════════════════════════
  ACQUISITION GAP RESTART RECOVERY
  ═══════════════════════════════════════════════════════════════

  IDUN restarts
       │
       ▼
  Goal Manager loads persisted Goals + GapRecords
       │
       ▼
  Goal.ExpiresAt reached?
     /              \
   YES               NO
    │                 │
  Archive       For each GapRecord in {DISPATCHED, RESOLVING}:
                      │
                      ▼

**F4 — Goal expiry before gap recovery**
During restart recovery, Goal Manager must check `Goal.ExpiresAt` before attempting GapRecord recovery. If the goal has expired, archive/terminate it according to the existing goal lifecycle instead of restarting its gap acquisition.
  Treat as UNCERTAIN
  (do NOT trust operational status — trust output store)
       │
       ▼
  Re-dispatch same GapID to Acquisition subsystem
       │
       ▼
  Acquisition subsystem performs idempotency check:
  Check authoritative output store for this GapID/Domain
       │
       ├── Verified result already exists
       │       │
       │       ▼
       │   Emit GapResolved(SUCCESS) immediately
       │   (no redundant work performed)
       │
       └── Result absent
               │
               ▼
           Begin acquisition
           Atomic store write
           Emit GapResolved(SUCCESS/FAILURE)

  ─────────────────────────────────────────────────────────
  NOTE: Goal Manager does NOT query Knowledge Store or
  Skill Registry directly on restart.
  Re-dispatching to Acquisition and relying on idempotency
  is the correct pattern — it avoids coupling Goal Manager
  to acquisition-specific stores.
  ─────────────────────────────────────────────────────────

  After GapRecord.Deadline check:
  GapRecord.Deadline > now:  re-dispatch with original deadline
  GapRecord.Deadline < now:  RetryCount check → retry or HOST_INTERVENTION
```

**Recovery procedure — Host-Response gaps (ClarificationGap, AuthorizationGap):**

```
  ═══════════════════════════════════════════════════════════════
  HOST-RESPONSE GAP RESTART RECOVERY
  ═══════════════════════════════════════════════════════════════

  IDUN restarts
       │
       ▼
  Goal Manager finds GapRecord{Type: CLARIFICATION or AUTHORIZATION,
                               Status: AWAITING_HOST_RESPONSE}
       │
       ▼
  No output store to check — resolution comes from Host
       │
       ▼
  Goal Manager signals Conversation Planner to re-emit
  the pending clarification or authorization request
       │
       ▼
  System waits for Host response (AWAITING_HOST_RESPONSE)
  (Staleness policy applies — Level 4/5)
```

**Idempotency contract (required for all Acquisition subsystems):**

> Re-dispatching the same `GapID` must not create duplicate authoritative results. If a verified record already exists in the output store for this GapID/Domain, Acquisition reports SUCCESS immediately without performing any work.

This is stronger than "safe to run twice." The check must be against the authoritative output store, not an internal operation log. The acquisition subsystem owns this check; Goal Manager does not perform it.

**Atomic write contract (required for Knowledge Store and Skill Registry):**

> Knowledge Store and Skill Registry writes must be atomic. A partially-written record must not be queryable by the idempotency check. Partial writes must be rolled back or invisible to readers.

**F3 — Idempotent acquisition writes (insert-if-absent)**
Knowledge Store and Skill Registry must use insert-if-absent / write-once semantics keyed by GapID, not merely atomic writes. This prevents duplicate authoritative results during concurrent retries, restart recovery, or duplicate dispatch.

```
  ═══════════════════════════════════════════════════════════════
  ACQUISITION IDEMPOTENCY (Diagram C)
  ═══════════════════════════════════════════════════════════════
  GapID
    │
    ▼
  Acquisition
    │
    ▼
  insert-if-absent(GapID)
    │
    ├── already exists → use existing result
    │
    └── absent → write authoritative result
```

#### 0.1.J — ClarificationGap Two-Tier Lifecycle (Resolved — OQ-0.1-4)

All ClarificationGaps are routed to Goal Manager. Goal Manager is the single lifecycle owner. There is no bypass path to Conversation Planner.

Goal Manager uses a **two-tier wait model** based on whether an active goal is blocked:

```
  ═══════════════════════════════════════════════════════════════
  CLARIFICATIONGAP — TWO-TIER LIFECYCLE
  ═══════════════════════════════════════════════════════════════

  Context Resolver detects ambiguity
         │
         │ emits ClarificationGap signal
         ▼
   Workspace Event Bus
   (content-blind transport)
         │
         ▼
    Goal Manager
    (single authoritative recipient)
         │
    Is an active goal blocked by this clarification?
         │
        YES                              NO
         │                               │
         ▼                               ▼
    GapRecord created            PendingClarification
    (persisted in storage)       token created in
    Goal → PAUSED /              Working Memory
    AWAITING_CLARIFICATION       (transient, not persisted)
         │                               │
         └──────────────┬────────────────┘
                        │
                        ▼
              Conversation Planner
              generates clarification
              question (natural language)
                        │
                        ▼
              Language Realization
                        │
                        ▼
                  Host / World
                        │
                        │ Host responds
                        ▼
                 [new stimulus / turn]
                        │
                        ▼
                  Goal Manager
                        │
          PERSISTENT ───┴─── TRANSIENT
               │                   │
               ▼                   ▼
         GapRecord             Working Memory
         resolved              token consumed
         Continuation          Next turn resolves
         Episode               normally
         scheduled
```

**Ownership for ClarificationGap:**

| Responsibility | Owner |
|:---------------|:------|
| Detect ClarificationGap | Context Resolver / Understanding |
| Emit ClarificationGap signal | Context Resolver |
| Transport signal | Event Bus (content-blind) |
| Single recipient | Goal Manager |
| Decide persistent vs transient | Goal Manager (checks active goal state) |
| Create persisted GapRecord | Goal Manager (blocking-goal path) |
| Create transient PendingClarification token | Goal Manager → Working Memory |
| Generate clarification question | Conversation Planner |
| Receive Host response (blocking path) | Goal Manager |
| Schedule Continuation Episode | Goal Manager (blocking path only) |
| Consume transient token | Context Resolver (next turn) |
| Re-emit request on restart | Goal Manager → Conversation Planner |

**What ClarificationGap must NOT use:**
- Level 2 acquisition deadline (`GapRecord.Deadline` — acquisition gaps only)
- Acquisition retry loop (no acquisition subsystem involved)
- `routeGap()` dispatch to acquisition topic

#### 0.1.K — Gap Dependency and Parallel Gaps (Resolved — OQ-0.1-1)

The architectural review examined whether `GapSignal` should carry `DependsOn []GapID` to express ordering between dependent gaps.

**Decision: Do NOT add `GapSignal.DependsOn`.**

Established patterns (build systems, workflow DAGs, HTN planning, Temporal) consistently place dependency ordering in the *structure* (the plan, the workflow definition, the task network) rather than in individual units of work. Following this pattern:

**Sequential gap dependencies** are handled by the continuation-episode model:

```
  ═══════════════════════════════════════════════════════
  SEQUENTIAL GAP HANDLING — NO DependsOn REQUIRED
  ═══════════════════════════════════════════════════════

  Goal active
     │
     ▼
  Planning / Reasoning detects Gap A
     │
     ▼
  Goal Manager: pause goal, dispatch Gap A
     │
     ▼
  Gap A acquired → RESOLVED
     │
     ▼
  Goal Manager: schedule Continuation Episode
     │
     ▼
  Continuation Episode runs with enriched context
     │
     ▼
  Planning / Reasoning detects Gap B
  (Gap B can now be fully specified
   because Gap A's result is available)
     │
     ▼
  Goal Manager: pause goal, dispatch Gap B
     │
     ▼
  Gap B acquired → RESOLVED
     │
     ▼
  Goal Manager: schedule Continuation Episode
     │
     ▼
  Goal proceeds normally

  The ordering is in the PLAN STRUCTURE,
  not in GapSignal dependency pointers.
```

**Independent parallel gaps** (if needed in the future) use a `PendingCount` counter, not a dependency graph:

```
  ═══════════════════════════════════════════════════════
  PARALLEL INDEPENDENT GAPS — PendingCount, NO DependsOn
  ═══════════════════════════════════════════════════════

  Goal Manager dispatches Gap A and Gap B simultaneously
  PendingCount = 2

  Gap A resolves → PendingCount = 1
  Gap B resolves → PendingCount = 0
                         │
                         ▼
              Schedule Continuation Episode
              (both results available in
               Knowledge Store / Working Memory)
```

**Constraints:**
- `GapSignal.DependsOn` must NOT be added
- Parallel gap dispatch machinery must NOT be implemented until a demonstrated architectural need exists
- When parallel gaps are eventually needed, the batch-level `GapBatch.Strategy` flag is preferred over per-gap dependency pointers
- Default behavior: sequential dispatch (one gap at a time through continuation episodes)

#### 0.1.L — Gap Ownership and Routing Diagrams

**Diagram 1 — Gap Detection, Transport, and Ownership**

```
  ═══════════════════════════════════════════════════════════════
  GAP DETECTION AND ROUTING — ALL FOUR TYPES
  ═══════════════════════════════════════════════════════════════

  ┌──────────────────────┐
  │ CONTEXT RESOLVER /   │──► ClarificationGap ──────────────┐
  │ UNDERSTANDING        │    (before Reasoning)              │
  └──────────────────────┘                                    │
                                                              │
  ┌──────────────────────┐                                    │
  │ REASONING            │──► KnowledgeGap ──────────────────┤
  └──────────────────────┘                                    │
                                                              │
  ┌──────────────────────┐                                    │
  │ PLANNING             │──► SkillGap ───────────────────────┤
  └──────────────────────┘                                    │
                                                              │
  ┌──────────────────────┐                                    │
  │ CONSTITUTIONAL GATE  │──► AuthorizationGap ───────────────┤
  └──────────────────────┘    (after Decision)                │
                                                              │
                              Workspace Event Bus             │
                              (content-blind — ◄──────────────┘
                               transports only,
                               does NOT inspect)
                                    │
                                    ▼
                             Goal Manager
                         (single authoritative
                          lifecycle owner)
                                    │
              ┌─────────────────────┼──────────────────────┐
              │                     │                      │
              ▼                     ▼                      ▼
     KnowledgeGap /         ClarificationGap /     AuthorizationGap
       SkillGap              (blocking goal)        → Conversation
          │                        │                    Planner
          ▼                        ▼
    routeGap()              GapRecord (persisted)
     (pure function)         or PendingClarification
          │                   token (Working Memory)
          ├── KnowledgeGap → TopicKnowledgeGapRequested
          ├── SkillGap     → TopicSkillGapRequested
          │
          ▼
    Acquisition Subsystem
    (executes, checks store,
     reports outcome)
```

**Diagram 2 — Where Each Gap Type Originates in the Pipeline**

```
  STIMULUS → UNDERSTANDING → CONTEXT RESOLVER
                                    │
                    ┌───────────────┤
                    │ Ambiguous?    │ Clear?
                    ▼               ▼
            ClarificationGap    Continues
            → Goal Manager      to Working
            → [two-tier]        Memory write
                                    │
                                    ▼
                                REASONING
                                    │
                    ┌───────────────┤
                    │ KnowledgeGap? │ None?
                    ▼               ▼
            Goal Manager        GOAL PROPOSAL
            [acq. branch]       → GOAL MANAGER
                                    │
                                    ▼
                                PLANNING
                                    │
                    ┌───────────────┤
                    │ SkillGap?     │ None?
                    ▼               ▼
            Goal Manager        DECISION
            [acq. branch]           │
                                    ▼
                            CONSTITUTIONAL GATE
                                    │
                    ┌───────────────┤
                    │ AuthGap?      │ Approved?
                    ▼               ▼
            Goal Manager        EXECUTIVE
            → Conv. Planner
```

#### 0.1.M — Complete Ownership Table (Updated)

| Responsibility | Owner |
|:---------------|:------|
| Detect `ClarificationGap` | Context Resolver / Understanding |
| Detect `KnowledgeGap` | Reasoning |
| Detect `SkillGap` | Planning |
| Detect `AuthorizationGap` | Constitutional Gate |
| Emit `GapSignal` | Detecting subsystem |
| Transport `GapSignal` | Workspace Event Bus (content-blind) |
| Receive all `GapSignal` types | Goal Manager (single recipient) |
| Pause goal | Goal Manager |
| Decide persistent vs. transient clarification | Goal Manager |
| Create persisted `GapRecord` (blocking goals) | Goal Manager |
| Create transient `PendingClarification` token | Goal Manager → Working Memory |
| Store `PendingClarification` token | Working Memory |
| Consume `PendingClarification` token | Context Resolver (next turn) |
| Dispatch acquisition gap via `routeGap()` | Goal Manager |
| `routeGap()` pure routing function | Goal Manager (colocated, no separate service) |
| Execute acquisition | Knowledge Acquisition / Skill Acquisition |
| Level 1 operation timeout | Acquisition Subsystem (internal only) |
| Level 2 acquisition resolution timeout | **Goal Manager exclusively** |
| Level 3 goal hard deadline (`Goal.ExpiresAt`) | **Goal Manager** |
| Level 4 staleness reminder (AWAITING_HOST) | **Goal Manager** (reads policy from SS-32) |
| Level 5 staleness archive (AWAITING_HOST) | **Goal Manager** (reads policy from SS-32) |
| Staleness policy configuration | SS-32 — Runtime Policy |
| Acquisition idempotency check | Acquisition Subsystem |
| Atomic write to output store | Knowledge Store / Skill Registry |
| Report resolution outcome | Acquisition Subsystem |
| Re-dispatch uncertain GapRecord on restart | Goal Manager |
| Re-emit pending request on restart (ClarificationGap) | Goal Manager → Conversation Planner |
| Retry decision | Goal Manager |
| Cancel in-flight acquisition | Goal Manager (signal) → Acquisition (executes) |
| Generate clarification question | Conversation Planner |
| Generate authorization request | Conversation Planner |
| Schedule Continuation Episode | Goal Manager |
| Resume goal | Goal Manager |
| Request Host intervention | Goal Manager → Conversation Planner → Language Realization |
| Permanently fail goal | Goal Manager |
| Sequential gap ordering | Plan/Goal structure (continuation episodes) |
| Parallel gap counter (if needed) | Goal Manager `PendingCount` |
| **F1** GapRecord write-before-acknowledge | Goal Manager |
| **F3** Idempotent insert-if-absent | Knowledge Store / Skill Registry |
| **F4** Restart check: Goal.ExpiresAt first | Goal Manager |
| **F5** AuthorizationGap persistent path routing | Goal Manager |
| **F11** Queue/discard multiple in-flight GapSignals | Goal Manager |
| **F12** Ignore GapResolved if CANCELLED | Goal Manager |

---

#### 0.1.N — Confirmed Decisions Summary (Updated)

| # | Decision | Status | Source |
|:--|:---------|:-------|:-------|
| 1 | No dedicated Gap Classifier subsystem | ✅ Confirmed | Initial review |
| 2 | Classification at point of detection (four owners) | ✅ Confirmed | Initial review |
| 3 | Event Bus remains content-blind for gap signals | ✅ Confirmed | Initial review |
| 4 | Goal Manager owns complete gap resolution lifecycle | ✅ Confirmed | Initial review |
| 5 | Goal Manager owns Level 2 resolution timeout exclusively | ✅ Confirmed | Initial review |
| 6 | Acquisition Subsystem owns Level 1 operation timeout | ✅ Confirmed | Initial review |
| 7 | No dedicated Gap Router subsystem | ✅ Confirmed | Initial review |
| 8 | Goal Manager uses pure routing function for dispatch | ✅ Confirmed | Initial review |
| 9 | Acquisition subsystems must be idempotent on re-dispatch | ✅ Confirmed | Initial review |
| 10 | Gap state (GapRecord) must survive process restarts | ✅ Confirmed | Initial review |
| 11 | GapRecord state machine splits into two branches (Acquisition / Host-Response) | ✅ Confirmed | OQ-0.1-4 |
| 12 | ClarificationGap always routes through Goal Manager (no bypass) | ✅ Confirmed | OQ-0.1-4 |
| 13 | ClarificationGap uses two-tier wait model (persistent vs. transient) | ✅ Confirmed | OQ-0.1-4 |
| 14 | GapSignal.DependsOn must NOT be added | ✅ Confirmed | OQ-0.1-1 |
| 15 | Sequential gap deps handled by continuation-episode model | ✅ Confirmed | OQ-0.1-1 |
| 16 | Goal Manager re-dispatches uncertain state on restart; Acquisition owns idempotency | ✅ Confirmed | OQ-0.1-2 |
| 17 | Goal Manager does NOT query Knowledge Store directly on restart | ✅ Confirmed | OQ-0.1-2 |
| 18 | Knowledge Store writes must be atomic | ✅ Confirmed | OQ-0.1-2 |
| 19 | Five distinct timeout/deadline levels; none may be conflated | ✅ Confirmed | OQ-0.1-3 |
| 20 | AWAITING_HOST has no mandatory maximum lifetime; staleness policy applies | ✅ Confirmed | OQ-0.1-3 |
| 21 | Level 2 deadline does NOT apply to ClarificationGap or AuthorizationGap | ✅ Confirmed | OQ-0.1-3 + OQ-0.1-4 |

#### 0.1.O — Open Questions After Architectural Review

> **OPEN — Level 4/5 Policy Values**
> Exact `ReminderInterval` and `ArchiveThreshold` values per goal priority band are not hardcoded here. These are configuration policy decisions owned by SS-32 (Runtime Policy). Must be resolved before Goal Manager's staleness logic is implemented.

> **OPEN — Parallel Gap Execution Mechanism**
> When parallel independent gaps are genuinely needed (not yet demonstrated), the exact mechanism for `GapBatch.Strategy` and `PendingCount` tracking must be designed as a separate architectural decision. Do not implement until needed.

> **OPEN — Working Memory Slot: `PendingClarification` Token Lifetime**
> When a transient `PendingClarification` token is written to Working Memory for a non-blocking clarification, how long should the token persist? Does it survive the turn? Multiple turns? This must be resolved in the Working Memory design specification (Phase 2).

**Decision 0.1 freeze status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE — 6 implementation constraints added. The three remaining open questions above are policy/configuration decisions, not architectural decisions. The core architecture is resolved and frozen.

#### 0.1.P — Per-Gap-Type Detailed Traceability

This section upgrades each gap type to the same data-provenance standard used in Decision 0.3. Every major connection identifies source, producer, contract, consumer, read dependencies, write authority, persistence, lifecycle, and failure path.

---

##### 0.1.P.1 — KnowledgeGap: Complete Trace

**Legend**
```text
──────→  synchronous / data flow
- - - →  asynchronous / event flow
READ →   read dependency
WRITE →  write authority
[STORE]  persistent storage
[WM]     Working Memory slot
[EVENT]  event transport (content-blind)
[CONTRACT] typed contract in intelligence/types
⚠        failure / degraded path
```

**Origin and Detection**
```text
═══════════════════════════════════════════════════════════════
KNOWLEDGEGAP — ORIGIN AND DETECTION
═══════════════════════════════════════════════════════════════

REASONING (intelligence/reasoning)
  │
  │ READ → Working Memory (EPISODE scope)
  │          ActiveBeliefs, RetrievedMemories
  │ READ → Knowledge Store
  │          domain facts, acquired records
  │
  │ Determines: knowledge required to proceed does not exist
  │             in Working Memory or Knowledge Store
  ▼
KnowledgeGap [CONTRACT]
  │   fields: GapID, GoalID, EpisodeID, Domain,
  │           Query, Context, CorrelationToken
  │   producer: Reasoning
  │   lifecycle: EPISODE
  │   persistence: TRANSIENT (signal only; GapRecord created by Goal Manager)
  ▼
GapSignal [EVENT] [CONTRACT]
  │   transport: Workspace Event Bus (content-blind)
  │   producer: Reasoning
  │   consumer: Goal Manager (only)
  │   lifecycle: TRANSIENT — consumed on delivery
  │   ⚠ Event Bus failure → signal lost
  │     Goal Manager does not receive → gap not handled
  │     OPEN — dead-letter / retry on signal transport failure
  ▼
WORKSPACE EVENT BUS
  │ content-blind — does not inspect GapType
  ▼
GOAL MANAGER (intelligence/goalmanager)
```

**Goal Manager Processing**
```text
═══════════════════════════════════════════════════════════════
KNOWLEDGEGAP — GOAL MANAGER PROCESSING (F1 ordering enforced)
═══════════════════════════════════════════════════════════════

GOAL MANAGER receives GapSignal
  │
  │ READ → Goal record [STORE] (core/storage)
  │          GoalID, State, ExpiresAt
  │ VALIDATE: GoalID correlates to an active, non-cancelled goal
  │
  │ ExpiresAt check (F4)
  │   ├── expired → archive goal; ignore signal; ACK
  │   └── active → continue
  │
  ▼ (F1 — write-before-acknowledge)
Persist GapRecord [STORE] (core/storage)
  │   fields: GapID, GoalID, Type=KNOWLEDGE,
  │           Domain, Status=CREATED,
  │           StartedAt, Deadline (Level 2 timeout)
  │   WRITE → core/storage
  │   ⚠ FAIL → do NOT acknowledge GapSignal
  │             Signal remains on bus; will be redelivered
  └── SUCCESS
         │
         ▼
Pause Goal
  │   WRITE → Goal.State = PAUSED [STORE]
  ▼
routeGap(KnowledgeGap) — pure function, no goroutines
  │   maps: KnowledgeGap → TopicKnowledgeGapRequested
  ▼
Dispatch to Knowledge Acquisition
  │   [EVENT] → TopicKnowledgeGapRequested
  │   payload: GapID, GoalID, Domain, Query, Deadline
  │   ⚠ dispatch failure → GapRecord.Status = DISPATCH_FAILED
  │                        retry dispatch (Goal Manager policy)
  │
  ▼ (F1 complete)
Acknowledge GapSignal from Event Bus
```

**Acquisition Branch**
```text
═══════════════════════════════════════════════════════════════
KNOWLEDGEGAP — ACQUISITION BRANCH
═══════════════════════════════════════════════════════════════

KNOWLEDGE ACQUISITION (intelligence/acquisition)
  │ receives: GapID, GoalID, Domain, Query, Deadline
  │
  │ IDEMPOTENCY CHECK (F3 — insert-if-absent)
  │   READ → Knowledge Store [STORE]
  │   query: exists(GapID) in authoritative output store?
  │   ├── EXISTS → emit GapResolved(SUCCESS) immediately
  │   │             no acquisition work performed
  │   └── ABSENT → begin acquisition
  │
  │ Level 1 timeout: context.WithDeadline (internal only)
  │ Goal Manager never sees Level 1 details
  │
  │ ACQUISITION OPERATION
  │   ├── HTTP/tool call / LLM synthesis
  │   ├── Level 1 timeout fires → retry internally
  │   └── Max internal retries → emit GapResolved(FAILURE)
  │
  │ ATOMIC WRITE (if result obtained)
  │   WRITE → Knowledge Store [STORE] (insert-if-absent keyed by GapID)
  │   ⚠ partial write → rolled back / invisible to idempotency check
  │
  ▼
GapResolved [EVENT] [CONTRACT]
  │   fields: GapID, GoalID, Status (SUCCESS|FAILURE), StoreKey?
  │   producer: Knowledge Acquisition
  │   consumer: Goal Manager (only)
  │   transport: Workspace Event Bus (content-blind)
  ▼
GOAL MANAGER
  │ Correlate GapID → GapRecord
  │ Check: Goal.State == CANCELLED? (F12)
  │   └── YES → ignore; GapRecord closed; no continuation episode
  │
  ├── SUCCESS
  │     WRITE → GapRecord.Status = RESOLVED [STORE]
  │     WRITE → Goal.State = ACTIVE [STORE]
  │     Schedule Continuation Episode
  │
  ├── FAILURE
  │     GapRecord.RetryCount < Max?
  │     ├── YES → re-dispatch (GapRecord.Status = RETRYING)
  │     └── NO  → GapRecord.Status = HOST_INTERVENTION_REQUIRED
  │               Goal → AWAITING_HOST
  │               Conversation Planner generates intervention request
  │
  └── DEADLINE EXCEEDED (Level 2: GapRecord.Deadline < now)
        Goal Manager cancels in-flight acquisition
        → TIMED_OUT → retry or HOST_INTERVENTION per RetryCount
```

---

##### 0.1.P.2 — SkillGap: Complete Trace

**Origin and Detection**
```text
═══════════════════════════════════════════════════════════════
SKILLGAP — ORIGIN AND DETECTION
═══════════════════════════════════════════════════════════════

PLANNING (intelligence/planning)
  │
  │ READ → Working Memory (EPISODE scope)
  │          CurrentPlan, ActiveBeliefs
  │ READ → Skill Registry [STORE]
  │          available SkillCards
  │ READ → Goal Manager: active goals
  │
  │ Determines: plan step requires a skill that does not exist
  │             or is not in AVAILABLE state in Skill Registry
  ▼
SkillGap [CONTRACT]
  │   fields: GapID, GoalID, EpisodeID, SkillSpec,
  │           RequiredCapability, CorrelationToken
  │   producer: Planning
  │   lifecycle: EPISODE
  ▼
GapSignal [EVENT] → Workspace Event Bus → Goal Manager

(Same F1 persistence ordering, routeGap(), and Acquisition
 Branch as KnowledgeGap; replace Knowledge Acquisition with
 Skill Acquisition and Knowledge Store with Skill Registry)
```

**Key differences from KnowledgeGap:**
```text
  routeGap(SkillGap) → TopicSkillGapRequested

  Skill Acquisition performs:
    source / validate / sandbox / AVAILABLE transition
    writes to Skill Registry (insert-if-absent keyed by GapID)

  On SUCCESS:
    Skill Registry entry: Status = AVAILABLE
    Goal Manager schedules Continuation Episode
    Planning resumes with new SkillCard available

  Level 1 timeout: Skill Acquisition internal
  Level 2 deadline: GapRecord.Deadline (Goal Manager)
  Idempotency: same insert-if-absent pattern (F3)
```

---

##### 0.1.P.3 — ClarificationGap: Complete Trace

**Origin and Detection**
```text
═══════════════════════════════════════════════════════════════
CLARIFICATIONGAP — ORIGIN AND DETECTION
═══════════════════════════════════════════════════════════════

CONTEXT RESOLVER / UNDERSTANDING (intelligence/understanding)
  │
  │ READ → Working Memory (TURN/SESSION scope)
  │          ConversationTurns, PendingClarification
  │          ActiveTopic, CurrentEntities
  │ READ → UnderstandingBatch (current turn)
  │
  │ Determines: cannot resolve pronoun / ellipsis / temporal
  │             reference; ambiguity cannot be resolved from
  │             existing context in Working Memory
  ▼
ClarificationGap [CONTRACT]
  │   fields: GapID, GoalID?, EpisodeID?,
  │           AmbiguousFragment, CandidateInterpretations[],
  │           CorrelationToken
  │   producer: Context Resolver
  │   lifecycle: TURN (signal); then GOAL or TRANSIENT
  ▼
GapSignal [EVENT] → Workspace Event Bus → Goal Manager
```

**Two-Tier Goal Manager Processing**
```text
═══════════════════════════════════════════════════════════════
CLARIFICATIONGAP — TWO-TIER GOAL MANAGER PROCESSING
═══════════════════════════════════════════════════════════════

GOAL MANAGER receives GapSignal
  │
  │ DECISION: Is an active goal blocked by this clarification?
  │
  ├── YES (blocking-goal path)
  │     │
  │     │ (F1) Persist GapRecord [STORE] (core/storage)
  │     │       GapRecord.Type = CLARIFICATION
  │     │       GapRecord.Status = AWAITING_HOST_RESPONSE
  │     │       NO Deadline (Level 2 does NOT apply)
  │     │       NO acquisition dispatch
  │     │ ⚠ FAIL → do not acknowledge signal
  │     │
  │     │ Pause Goal → AWAITING_CLARIFICATION
  │     │ WRITE → Goal.State = PAUSED [STORE]
  │     │
  │     └── ACK signal
  │
  └── NO (transient path)
        │
        │ WRITE → Working Memory [WM]
        │          PendingClarification token
        │          (TURN/SESSION scope — NOT persisted)
        │ ACK signal
        │
        └── token lifetime: OPEN — IMPLEMENTATION POLICY REQUIRED
              (bounded by mechanical capacity eviction at minimum)

  (Both paths converge here)
  │
  ▼
Goal Manager signals Conversation Planner
  │ to generate clarification question
  ▼
CONVERSATION PLANNER (SS-20)
  │ Reads: ClarificationGap context (from GapSignal or WM token)
  │ Generates: ConversativeIntent(CLARIFY)
  ▼
CONSTITUTIONAL GATE → LANGUAGE REALIZATION → NOTIFICATION SERVICE → HOST

HOST RESPONDS (new Turn / stimulus)
  │
  ▼
UNDERSTANDING / CONTEXT RESOLVER
  │ READ → Working Memory [WM]
  │          PendingClarification (transient path)
  │          ConversationTurns (recent turns)
  │          GapRecord (blocking-goal path, via Goal Manager)
  │
  ├── BLOCKING-GOAL PATH:
  │     Goal Manager receives resolution signal
  │     WRITE → GapRecord.Status = RESOLVED [STORE]
  │     WRITE → Goal.State = ACTIVE [STORE]
  │     Schedule Continuation Episode
  │
  └── TRANSIENT PATH:
        Working Memory token consumed by Context Resolver
        Next turn resolves normally
        No Continuation Episode needed

RESTART RECOVERY (ClarificationGap — blocking path):
  Goal Manager finds GapRecord{Type=CLARIFICATION,
                               Status=AWAITING_HOST_RESPONSE}
  No output store to check (resolution comes from Host)
  Goal Manager re-emits clarification request via Conversation Planner
  System waits (staleness policy Levels 4/5 applies)

RESTART RECOVERY (transient path):
  PendingClarification token is NOT persisted
  Token is lost on restart
  Next Host input creates a new turn; Understanding processes fresh
```

---

##### 0.1.P.4 — AuthorizationGap: Complete Trace

**Origin and Detection**
```text
═══════════════════════════════════════════════════════════════
AUTHORIZATIONGAP — ORIGIN AND DETECTION
═══════════════════════════════════════════════════════════════

CONSTITUTIONAL GATE (SS-30 / intelligence/constitution)
  │
  │ READ → Decision record (approved plan for execution)
  │ READ → Constitution rules (HMAC-verified constraints)
  │ READ → Autonomy level (SS-27 / runtime config)
  │
  │ Determines: IDUN understands the action and the plan is
  │             approved by Decision, but the action exceeds
  │             the configured autonomy level OR requires
  │             explicit per-action Host authorization
  ▼
AuthorizationGap [CONTRACT]
  │   fields: GapID, GoalID, EpisodeID, PlannedAction,
  │           AuthorizationReason, CorrelationToken
  │   producer: Constitutional Gate
  │   lifecycle: EPISODE → GOAL (escalated to persistent)
  ▼
AuthorizationGap signal [EVENT]
  │   transport: Workspace Event Bus (content-blind)
  │   NOTE: AuthorizationGap ALWAYS uses the persistent path (F5)
  │         No transient PendingClarification variant
  ▼
WORKSPACE EVENT BUS → GOAL MANAGER
```

**Goal Manager Processing — Persistent Path Only (F5)**
```text
═══════════════════════════════════════════════════════════════
AUTHORIZATIONGAP — GOAL MANAGER PROCESSING
═══════════════════════════════════════════════════════════════

GOAL MANAGER receives signal
  │
  │ (F1) Persist GapRecord [STORE]
  │       GapRecord.Type = AUTHORIZATION
  │       GapRecord.Status = AWAITING_HOST_RESPONSE
  │       NO Deadline (Level 2 does NOT apply — F5)
  │       NO acquisition dispatch
  │ ⚠ FAIL → do not acknowledge; signal remains for redelivery
  │
  │ Pause Goal → AWAITING_HOST
  │ WRITE → Goal.State = AWAITING_HOST [STORE]
  │
  └── ACK signal

Goal Manager signals Conversation Planner
  │   to generate authorization request
  ▼
CONVERSATION PLANNER → CONSTITUTIONAL GATE → LANGUAGE REALIZATION
   → NOTIFICATION SERVICE → HOST

HOST RESPONDS (approves or denies)
  │
  ├── APPROVED:
  │     Understanding / Goal Manager receives approval
  │     WRITE → GapRecord.Status = RESOLVED [STORE]
  │     WRITE → Goal.State = ACTIVE [STORE]
  │     Goal Manager resumes: re-issue the HMAC-approved action
  │     Executive dispatches
  │
  └── DENIED:
        WRITE → GapRecord.Status = CANCELLED [STORE]
        WRITE → Goal.State = CANCELLED [STORE] (terminal — F12)
        Conversation Planner acknowledges denial to Host

RESTART RECOVERY (AuthorizationGap):
  Goal Manager finds GapRecord{Type=AUTHORIZATION,
                               Status=AWAITING_HOST_RESPONSE}
  No output store to check
  Goal Manager re-emits authorization request via Conversation Planner
  Staleness policy Levels 4/5 applies
```

---

#### 0.1.Q — Complete Annotated Flow Diagrams

**Legend (same as Decision 0.3)**
```text
──────→  synchronous / data flow
- - - →  asynchronous / event flow
READ →   read dependency
WRITE →  write authority
[STORE]  persistent storage
[WM]     Working Memory
[EVENT]  event transport
[CONTRACT] typed contract
⚠        failure / degraded path
```

**Diagram Q1 — Full Gap Lifecycle with Data Provenance**
```text
═══════════════════════════════════════════════════════════════
FULL GAP LIFECYCLE — DATA PROVENANCE ANNOTATED
═══════════════════════════════════════════════════════════════

  ┌────────────────────────────────────────────────────────┐
  │  GAP DETECTORS (per-stage, per-gap-type)               │
  │                                                        │
  │  Context Resolver / Understanding                      │
  │    READ → WM: ConversationTurns, PendingClarification  │
  │    READ → UnderstandingBatch (current turn)            │
  │    DETECTS: ambiguity → ClarificationGap               │
  │                                                        │
  │  Reasoning                                             │
  │    READ → WM: ActiveBeliefs, RetrievedMemories         │
  │    READ → Knowledge Store                              │
  │    DETECTS: missing fact/knowledge → KnowledgeGap      │
  │                                                        │
  │  Planning                                              │
  │    READ → WM: CurrentPlan, ActiveBeliefs               │
  │    READ → Skill Registry                               │
  │    DETECTS: missing capability → SkillGap              │
  │                                                        │
  │  Constitutional Gate                                   │
  │    READ → Constitution rules, Autonomy level           │
  │    READ → Decision record (approved plan)              │
  │    DETECTS: unauthorized action → AuthorizationGap     │
  └───────────────────────┬────────────────────────────────┘
                          │ produces: GapSignal [CONTRACT] [EVENT]
                          │   fields: GapID (UUID), GoalID, EpisodeID,
                          │           GapType, Domain/Spec, Context,
                          │           CorrelationToken
                          │ persistence: TRANSIENT (signal only)
                          ▼
  ┌────────────────────────────────────────────────────────┐
  │  WORKSPACE EVENT BUS                                   │
  │  content-blind — transports, does NOT inspect GapType  │
  │  ⚠ delivery failure: signal lost                       │
  │     OPEN — dead-letter / retry policy required         │
  └───────────────────────┬────────────────────────────────┘
                          │
                          ▼
  ┌────────────────────────────────────────────────────────┐
  │  GOAL MANAGER (intelligence/goalmanager)               │
  │  Single authoritative recipient of all GapSignals      │
  │                                                        │
  │  Step 1: Validate correlation                          │
  │    READ → Goal record [STORE]: GoalID exists? Active?  │
  │           ExpiresAt check (F4)                         │
  │    ├── expired → archive; ACK; done                    │
  │    └── active → continue                               │
  │                                                        │
  │  Step 2: Persist GapRecord (F1 — BEFORE ACK)           │
  │    WRITE → GapRecord [STORE] (core/storage)            │
  │    ⚠ FAIL → do NOT ACK signal; allow redelivery        │
  │                                                        │
  │  Step 3: Pause goal                                    │
  │    WRITE → Goal.State = PAUSED [STORE]                 │
  │                                                        │
  │  Step 4: Route (pure function — no separate service)   │
  │    routeGap(GapType) → dispatch topic                  │
  │                                                        │
  │  Step 5: Dispatch                                      │
  │    [EVENT] → Acquisition or Host-Response path         │
  │                                                        │
  │  Step 6: ACK signal (F1 complete)                      │
  └───────────────────────┬────────────────────────────────┘
                          │
              ┌───────────┴────────────────┐
              │                            │
              ▼                            ▼
  ACQUISITION PATH               HOST-RESPONSE PATH
  (KnowledgeGap, SkillGap)       (ClarificationGap, AuthorizationGap)
              │                            │
              ▼                            │
  Acquisition Subsystem          Goal Manager → Conv. Planner
    Level 1 timeout (internal)   → Constitutional Gate
    Idempotency check (F3)       → Language Realization
    Atomic write (F3)            → Notification Service → HOST
    GapResolved [EVENT]                    │
              │                            │ Host responds
              ▼                            ▼
  GOAL MANAGER                   GOAL MANAGER
    Correlate GapID              Correlate GapRecord
    CANCELLED? (F12)             Mark RESOLVED / CANCELLED
    → ignore if yes              Resume or terminate goal
    SUCCESS → resume
    FAILURE → retry or HOST_INTERVENTION
```

**Diagram Q2 — Multiple GapSignals and Cancellation**
```text
═══════════════════════════════════════════════════════════════
MULTIPLE GAPSIGNALS — CONCURRENT / OUT-OF-ORDER HANDLING
═══════════════════════════════════════════════════════════════

  ACTIVE GOAL
    │
    ├── GapSignal arrives (Gap A) → GapRecord created → DISPATCHED
    │
    ├── GapSignal arrives (Gap B) while PAUSED (F11):
    │     Options (implementation policy):
    │     ├── Queue: hold in Goal Manager internal queue;
    │     │          process after Gap A resolves
    │     └── Discard duplicate: if same GapID → drop safely
    │
    │   OPEN — IMPLEMENTATION POLICY: exact mechanism for
    │   concurrent GapSignals (queue depth, ordering guarantees)
    │   is not frozen. Must not corrupt lifecycle state.
    │
    ├── Gap A resolves (SUCCESS)
    │     → Goal resumes
    │     → Continuation Episode runs
    │     → Gap B (if queued) processed in new episode
    │
    └── Host cancels goal
          │
          ▼
      Goal.State = CANCELLED (F12 — TERMINAL)
          │
          ├── Gap A resolves later → IGNORED (F12)
          ├── Gap B resolves later → IGNORED (F12)
          └── No Continuation Episode scheduled

═══════════════════════════════════════════════════════════════
RESTART RECOVERY — ACQUISITION GAPS (Diagram Q3)
═══════════════════════════════════════════════════════════════

  SYSTEM RESTART
       │
       ▼
  GOAL MANAGER loads persisted Goals + GapRecords
       │
       ▼
  For each Goal: check Goal.ExpiresAt (F4)
       ├── expired → archive goal; do not recover acquisition
       └── active → continue
       │
       ▼
  For each GapRecord{Status ∈ {DISPATCHED, RESOLVING}}:
       │
       ▼
  TREAT AS UNCERTAIN
  (Do not trust operational status — trust output store)
       │
       ▼
  Re-dispatch same GapID → Acquisition Subsystem
       │
       ▼
  ACQUISITION SUBSYSTEM: idempotency check (F3)
       │   READ → authoritative output store (Knowledge Store / Skill Registry)
       │   query: exists(GapID)?
       ├── EXISTS (result already written before crash)
       │      │ emit GapResolved(SUCCESS)
       │      │ NO duplicate acquisition work
       │
       └── ABSENT
              │ begin acquisition
              │ atomic write (insert-if-absent)
              │ emit GapResolved(SUCCESS | FAILURE)
       │
       ▼ (GapRecord.Deadline check before re-dispatch)
  Deadline > now  → re-dispatch with original deadline
  Deadline < now  → RetryCount check → retry or HOST_INTERVENTION
```

---

#### 0.1.R — Implementation Wiring Map

```text
═══════════════════════════════════════════════════════════════
DECISION 0.1 — IMPLEMENTATION WIRING MAP
═══════════════════════════════════════════════════════════════

This map answers: when implementation begins, where do I connect this?

CONNECTION 1: Context Resolver → GapSignal emission
  Producer:   intelligence/understanding (Context Resolver)
  Contract:   ClarificationGap [intelligence/types]
  Transport:  Workspace Event Bus topic: TopicClarificationGapDetected
  Consumer:   intelligence/goalmanager
  State:      None (signal is transient)
  Dependency: intelligence/types (Tier 0, must exist first)

CONNECTION 2: Reasoning → GapSignal emission
  Producer:   intelligence/reasoning
  Contract:   KnowledgeGap [intelligence/types]
  Transport:  Workspace Event Bus topic: TopicKnowledgeGapRequested
  Consumer:   intelligence/goalmanager
  State:      None (signal is transient)
  Dependency: intelligence/types

CONNECTION 3: Planning → GapSignal emission
  Producer:   intelligence/planning
  Contract:   SkillGap [intelligence/types]
  Transport:  Workspace Event Bus topic: TopicSkillGapRequested
  Consumer:   intelligence/goalmanager
  State:      None
  Dependency: intelligence/types

CONNECTION 4: Constitutional Gate → GapSignal emission
  Producer:   intelligence/constitution (SS-30)
  Contract:   AuthorizationGap [intelligence/types]
  Transport:  Workspace Event Bus topic: TopicAuthorizationRequired
  Consumer:   intelligence/goalmanager
  State:      None
  Dependency: intelligence/types

CONNECTION 5: Goal Manager → GapRecord persistence
  Writer:     intelligence/goalmanager
  Contract:   GapRecord struct [intelligence/types]
  Destination: core/storage (persistence layer)
  Key:        GapID (UUID)
  Ordering:   F1 — persist BEFORE acknowledging signal
  ⚠ Failure: do NOT acknowledge; signal remains for redelivery

CONNECTION 6: Goal Manager → routeGap() dispatch
  Location:   intelligence/goalmanager (pure function, no separate package)
  Input:      GapType enum [intelligence/types]
  Output:     dispatch topic string
  State:      stateless
  Package:    PACKAGE LOCATION — within intelligence/goalmanager

CONNECTION 7: Goal Manager → Knowledge Acquisition dispatch
  Publisher:  intelligence/goalmanager
  Contract:   GapDispatchRequest [intelligence/types]
  Topic:      TopicKnowledgeGapRequested
  Consumer:   intelligence/acquisition/knowledge
  State:      GapRecord [STORE] (core/storage)

CONNECTION 8: Goal Manager → Skill Acquisition dispatch
  Publisher:  intelligence/goalmanager
  Contract:   GapDispatchRequest [intelligence/types]
  Topic:      TopicSkillGapRequested
  Consumer:   intelligence/acquisition/skill
  State:      GapRecord [STORE] (core/storage)

CONNECTION 9: Acquisition → GapResolved emission
  Producers:  intelligence/acquisition/knowledge,
              intelligence/acquisition/skill
  Contract:   GapResolved [intelligence/types]
              fields: GapID, GoalID, Status, StoreKey?
  Transport:  Workspace Event Bus
  Consumer:   intelligence/goalmanager

CONNECTION 10: Acquisition → authoritative output store write (F3)
  Writer:     intelligence/acquisition/knowledge → core/storage (Knowledge Store)
  Writer:     intelligence/acquisition/skill → core/storage (Skill Registry)
  Semantics:  insert-if-absent keyed by GapID
  Package:    PACKAGE LOCATION — TO BE DEFINED DURING IMPLEMENTATION
              (atomic write requirement applies regardless of storage tech)

CONNECTION 11: Goal Manager → Working Memory (PendingClarification)
  Writer:     intelligence/goalmanager
  Target:     Working Memory TURN scope [WM]
  Contract:   PendingClarification [intelligence/types]
  Via:        WorkingMemoryManager (write authority enforced)
  Lifetime:   OPEN — IMPLEMENTATION POLICY REQUIRED
              bounded at minimum by mechanical LRU eviction

CONNECTION 12: Goal Manager → Conversation Planner (Host-Response gaps)
  Signal:     Goal Manager triggers ClarificationGap / AuthorizationGap notification
  Path:       Goal Manager → Conversation Planner → Constitutional Gate
              → Language Realization → Notification Service → HOST
  Mechanism:  OPEN — exact invocation mechanism (direct call vs topic)
              TO BE DEFINED DURING IMPLEMENTATION
              Must respect the full proactive communication pipeline (Decision 0.3)
```

---




---

### Decision 0.2 — Working Memory: Service or Store?

**Review Status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE

**Proposal:** Working Memory as a first-class subsystem with its own goroutines and lifecycle.

**Challenge:** The word "subsystem" implies complexity. Working Memory is fundamentally a bounded, in-memory key-value store with expiration and priority rules. It does not need goroutines to exist. It needs a contract.

**Analysis:** Working Memory must be:
- Readable by any cognitive subsystem (Reasoning, Planning, Conversation Planner)
- Writable by a defined set of subsystems (Understanding, Execution layer, Goal Manager)
- Bounded in capacity
- Automatically expiring stale entries
- Accessible as a synchronous in-process read on the hot cognitive path

This is a **typed in-process store with a defined write contract**, not a service that publishes to the workspace bus.

**Decision:** Working Memory is a **Store** — a typed in-process structure with defined read access and write authority per slot. It is managed by a `WorkingMemoryManager` that enforces capacity, priority, expiration, checkpointing, and consolidation. It is injected into subsystems that need it. It does not communicate via the Workspace Event Bus.

**Impact:** Cleaner architecture. No message-passing overhead for context queries that happen on the critical path.

#### 0.2.A — Lifetime Scope Partitioning

Working Memory slots are partitioned into four explicit lifetime scopes. Each scope has defined persistence behavior.

```text
  ═══════════════════════════════════════════════════════════════
  WORKING MEMORY LIFETIME HIERARCHY
  ═══════════════════════════════════════════════════════════════

  PROCESS START
       │
       ├── SESSION (active host session; survives across turns and episodes)
       │     ConversationTurns, ActiveTopic, CurrentEntities, RecentCorrections
       │     Checkpoint: NOT persisted — restart = fresh session context
       │
       │     ├── GOAL (while a specific goal is active; may span sessions)
       │     │     ActiveGoalID, ActiveSubgoalIDs, ActiveGapIDs (ref-only)
       │     │     Checkpoint: REQUIRED — restored on restart from core/storage
       │     │
       │     │     ├── EPISODE (bounded cognitive work session)
       │     │     │     CurrentPlan, PlanCheckpoint, ActiveBeliefs,
       │     │     │     UnresolvedQuestions, RecentObservations
       │     │     │     Checkpoint: REQUIRED for PlanCheckpoint
       │     │     │                RECOMMENDED for ActiveBeliefs,
       │     │     │                             UnresolvedQuestions
       │     │     │     TemporaryResearch, RetrievedMemories:
       │     │     │                DO NOT checkpoint
       │     │     │
       │     │     └── TURN (single stimulus-response cycle)
       │     │           PendingClarification (transient path only)
       │     │           Checkpoint: NOT persisted
       │     │           Lifetime: Bounded by mechanical capacity eviction
       │     │
       │     └── (goal resolves → goal scope cleared; session scope retained)
       │
       └── (session ends → consolidation → Long-Term Memory)
```

**Restart rule:** SESSION-scoped slots start fresh after restart (stale session context is not injected into a new session). GOAL-scoped and required EPISODE-scoped slots (PlanCheckpoint) are restored from `core/storage` checkpoint. `TemporaryResearch` and `RetrievedMemories` are not checkpointed — they are re-acquired or re-fetched on resume.

#### 0.2.B — PendingGaps Removed; ActiveGapIDs Added

`PendingGaps` is removed from Working Memory. GapRecords are authoritative lifecycle state owned by Goal Manager and persisted in `core/storage`. Placing them in Working Memory creates dual authority and checkpoint inconsistency.

**Replacement:** `ActiveGapIDs []GapID` — a reference-only slot. Contains the IDs of gaps currently in-flight for the foreground goal. No GapRecord state lives here. Goal Manager remains the sole authoritative owner of GapRecord state.

#### 0.2.C — Memory Retrieval Layer Boundary

The `RetrievedMemories` slot is populated by a **Memory Retrieval Layer**, which queries Long-Term Memory for records relevant to the current episode context and writes them to Working Memory.

**Architectural boundaries:**
- Working Memory does NOT pull from Long-Term Memory directly (passive store).
- Goal Manager does NOT trigger or execute general memory retrieval (separation of concerns).
- Executive / WorkflowCoordinator triggers retrieval at episode start.
- `MemoryRetriever` is an interface/contract boundary (e.g., in `intelligence/interfaces`), abstracting LTM implementation (JSON/Vector/Graph).

```text
  ═══════════════════════════════════════════════════════════════
  MEMORY RETRIEVAL OWNERSHIP (Diagram 2)
  ═══════════════════════════════════════════════════════════════

         EPISODE START
               │
               ▼
      WorkflowCoordinator
          (Executive)
               │ (triggers retrieval via interface)
               ▼
        MemoryRetriever
          (contract)
               │ (queries using episode context)
               ▼
       Long-Term Memory
               │ (returns semantic knowledge)
               ▼
       Retrieved Context
               │ (pushes into)
               ▼
     WorkingMemoryManager
               │ (exposes for reading)
               ▼
           Reasoning
```

#### 0.2.D — Multi-Goal Architecture and Dormant Goal Contexts

Working Memory is partitioned into a global section and isolated per-goal sections. The number of dormant goal contexts held in RAM is bounded by a capacity-aware cache.

```text
  ═══════════════════════════════════════════════════════════════
  DORMANT GOAL LIFECYCLE (Diagram 3)
  ═══════════════════════════════════════════════════════════════

           ACTIVE GOALS
                │
                ▼
         Working Memory
                │
           Goal becomes
              dormant
                │
                ▼
       Dormant Goal Cache
                │
        Capacity exceeded?
           /          \
         NO            YES
         │              │
         ▼              ▼
       Keep       Priority-aware
                   LRU eviction
                        │
                        ▼
               Verify checkpoint
                 (re-checkpoint
                  if stale)
                        │
                        ▼
                  Evict from RAM
```

**Rule:** At any point in time, exactly one goal is the foreground goal. Background goals have their contexts managed by a priority-aware LRU cache. The exact capacity and eviction weights are runtime policies, not architectural constants.

**Rehydration:** When a dormant goal is reactivated: Restore checkpoint → Validate gmStateVersion → Retrieve current relevant memory → Re-dispatch uncertain gaps → Resume goal.

#### 0.2.E — Concurrency Model

```text
  ═══════════════════════════════════════════════════════════════
  WORKING MEMORY CONCURRENCY MODEL
  ═══════════════════════════════════════════════════════════════

  Global slots:
    shared RWMutex — multiple concurrent readers, exclusive writer

  Per-goal slots:
    per-goal RWMutex — multiple concurrent readers, exclusive writer
    only the subsystem processing the foreground goal may write
    background goal contexts are effectively read-only at runtime

  ForegroundGoalID:
    atomic pointer — Goal Manager swaps on goal context switch

  Write authority enforcement:
    checked INSIDE the write lock (not by convention)
    unauthorized write returns an error; never silently succeeds
```

#### 0.2.F — PendingClarification Architectural Role

PendingClarification in Working Memory applies only to the transient (non-blocking) path defined in Decision 0.1 §0.1.J.

```text
  ═══════════════════════════════════════════════════════════════
  PENDINGCLARIFICATION LIFECYCLE (Diagram 1)
  ═══════════════════════════════════════════════════════════════

               PendingClarification
                       │ (Transient path)
                       ▼
                Working Memory
                       │
          ┌────────────┼────────────┐
          │            │            │
     Host answer   Goal ends   Capacity pressure
          │            │            │
          ▼            ▼            ▼
      RESOLVED      INVALID    LRU EVICTION

  TRANSITION TO PERSISTENT GOAL:
  
  Transient Clarification
          │ (Host responds; creates a real goal)
          ▼
     Goal Manager
          │ (creates GapRecord; deduplicates token)
          ▼
  Persistent Goal / GapRecord
```

**Eviction Rule:** Transient tokens are bounded by mechanical capacity eviction (e.g., LRU), not semantic evaluation. Working Memory does NOT decide if a question is "stale." Exact capacity remains runtime configuration.

#### 0.2.G — Checkpoint Consistency Contract

Working Memory checkpoints must be atomic and consistent with Goal Manager state. **Goal Manager is the authoritative source of truth; Working Memory is a recoverable projection.**

```text
  ═══════════════════════════════════════════════════════════════
  CHECKPOINT CONSISTENCY ORDERING (Diagram 4)
  ═══════════════════════════════════════════════════════════════

                    Goal Manager
                         │
                         ▼
              Persist authoritative
                  goal state
                         │
                    SUCCESS?
                    /      \
                  NO        YES
                  │           │
                  ▼           ▼
                STOP      WMM checkpoint
                         (with gmStateVersion)
                              │
                         SUCCESS?
                         /     \
                       NO       YES
                       │          │
                       ▼          ▼
                 Recovery      Continue
                              / Pause
```

**Version Contract:** The Working Memory checkpoint must include a `gmStateVersion`. On restart, if the `gmStateVersion` in the checkpoint does not match the Goal Manager's authoritative state version (e.g., GM=17, WM=16), the Goal Manager's state wins, and the Working Memory context is re-derived or discarded.

#### 0.2.H — Consolidation for Long-Running Goals

Long-running goals span multiple sessions and maintain a clear separation between **Execution State** (owned by Goal Manager + Working Memory) and **Semantic Knowledge** (owned by Long-Term Memory).

```text
  ═══════════════════════════════════════════════════════════════
  LONG-RUNNING GOAL LIFECYCLE (Diagram 5)
  ═══════════════════════════════════════════════════════════════

                 LONG-RUNNING GOAL
                        │
              ┌─────────┴─────────┐
              │                   │
       Goal Execution State    Semantic Knowledge
              │                   │
       Goal Manager +          Long-Term Memory
       WM Checkpoint
              │                   │
              └─────────┬─────────┘
                        │
                  Future Session
                        │
                        ▼
                 Goal Rehydration
                        │
              ┌─────────┴─────────┐
              ▼                   ▼
        Restore checkpoint    Retrieve current
                              relevant knowledge
              │                   │
              └─────────┬─────────┘
                        ▼
                     Reasoning
                        │ (Old beliefs are TENTATIVE;
                        ▼  must be re-evaluated)
                  Continue Goal
```

**Critical Rules:**
- LTM does not own the goal lifecycle. If LTM holds an "ongoing goal" entry, it is purely a retrieval index/summary, not the authoritative state.
- Restored beliefs (`ActiveBeliefs`) from old checkpoints are **TENTATIVE** and must be reconsidered against current retrieved context by Reasoning at S0.

#### 0.2.I — Integrated Working Memory Architecture

```text
  ═══════════════════════════════════════════════════════════════
  COMPLETE WORKING MEMORY RELATIONSHIP (Diagram 6)
  ═══════════════════════════════════════════════════════════════

                    Goal Manager
                         │
                         ├──────────────► Goal State (core/storage)
                         │                 (Authoritative)
                         ▼
               WorkingMemoryManager
                         │
                         ├── GLOBAL (Session)
                         ├── GOAL   (Dormant Cache / LRU)
                         ├── EPISODE(Checkpoint Projection)
                         └── TURN   (Transient / Mechanical LRU)
                         │
                         ▼
    WorkflowCoordinator ─► MemoryRetriever
      (Episode Start)            │
                                 ▼
                         Long-Term Memory
                           (Semantic/Facts)
```

#### 0.2.J — Deferred Runtime Policies

The following details are explicitly deferred as runtime configuration / calibration policies (SS-32), rather than architectural invariants:
- **PendingClarification capacity:** Exact number of transient tokens held in memory.
- **Dormant-goal capacity:** Exact number of dormant goal contexts held in RAM before checkpoint eviction.
- **Eviction weights:** The priority weights vs recency factors used in LRU caches.
- **Storage implementation:** The exact underlying database technology (JSON/Vector/SQL) used for checkpoints or LTM.
- **Memory ranking:** The specific algorithms used to score relevance in `MemoryRetriever`.

**Decision 0.2 freeze status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE. The architecture establishes clear ownership, lifecycle, and recovery invariants. Configuration values are deferred to policy.

#### 0.2.K — Complete Slot Ownership Matrix

```text
LEGEND
  Scope:     GLOBAL | GOAL | EPISODE | TURN
  Producer:  subsystem that creates/populates the slot
  Writer:    who is authorized to write (enforced by WorkingMemoryManager)
  Readers:   subsystems that may read this slot
  Persist:   REQUIRED | RECOMMENDED | NOT PERSISTED
  Checkpoint: YES | NO | POLICY
```

| Slot | Scope | Producer | Writer Authority | Readers | Persist | Checkpoint |
|:-----|:------|:---------|:-----------------|:--------|:--------|:-----------|
| `ConversationTurns` | GLOBAL (session) | Understanding + Conversation Planner | Understanding, Conversation Planner (via WMM) | Understanding, Context Resolver, Conversation Planner | NOT PERSISTED | NO — fresh on restart |
| `ActiveTopic` | GLOBAL (session) | Understanding | Understanding (via WMM) | Reasoning, Planning, Conversation Planner, Attention | NOT PERSISTED | NO |
| `CurrentEntities` | GLOBAL (session) | Context Resolver | Context Resolver (via WMM) | Reasoning, Planning, Conversation Planner, Attention | NOT PERSISTED | NO |
| `RecentCorrections` | GLOBAL (session) | Context Resolver | Context Resolver (via WMM) | Understanding, Reasoning | NOT PERSISTED | NO |
| `HostPreferences` | GLOBAL (session) | OPEN — ownership contract required | OPEN | Conversation Planner | NOT PERSISTED | NO |
| `ForegroundGoalID` | GLOBAL | Goal Manager | Goal Manager (atomic pointer via WMM) | All cognitive subsystems | REQUIRED | YES — Goal Manager state |
| `ActiveGoalID` | GOAL | Goal Manager | Goal Manager (via WMM) | Reasoning, Planning, Decision, Conversation Planner, Attention | REQUIRED | YES |
| `ActiveSubgoalIDs` | GOAL | Goal Manager | Goal Manager (via WMM) | Planning, Reasoning | REQUIRED | YES |
| `ActiveGapIDs` | GOAL | Goal Manager | Goal Manager (via WMM) | OPEN — reference-only, visibility policy required | REQUIRED | YES — ref only, GapRecord authoritative in core/storage |
| `CurrentPlan` | EPISODE | Planning | Planning (via WMM) | Decision, Executive, Conversation Planner | RECOMMENDED | YES if mid-episode pause |
| `PlanCheckpoint` | EPISODE | Planning | Planning (via WMM) | Goal Manager (rehydration), Planning (resume) | REQUIRED | YES — required for continuation episode |
| `ActiveBeliefs` | EPISODE | Reasoning | Reasoning (via WMM) | Reasoning (next stage), Planning | RECOMMENDED | YES — marked TENTATIVE on restore |
| `UnresolvedQuestions` | EPISODE | Reasoning / Conversation Planner | OPEN — ownership contract required | Reasoning, Conversation Planner | RECOMMENDED | YES |
| `RecentObservations` | EPISODE | Executive | Executive (via WMM) | Reasoning, Conversation Planner | NOT PERSISTED | NO |
| `TemporaryResearch` | EPISODE | Knowledge Acquisition / MemoryRetriever | OPEN | Reasoning | NOT PERSISTED | NO — re-acquired on resume |
| `RetrievedMemories` | EPISODE | MemoryRetriever (triggered by Executive) | MemoryRetriever (via WMM) | Reasoning | NOT PERSISTED | NO — re-fetched on resume |
| `PendingClarification` | TURN | Goal Manager | Goal Manager (via WMM) | Context Resolver | NOT PERSISTED | NO |

> **OPEN items in this table** must be resolved as implementation contracts before the relevant subsystem is implemented. Do not invent ownership; mark it OPEN until the contract is established.

---

#### 0.2.L — Read/Write Architecture

```text
═══════════════════════════════════════════════════════════════
WORKING MEMORY READ/WRITE ARCHITECTURE
═══════════════════════════════════════════════════════════════

LEGEND
  ──────→  synchronous data flow
  READ →   read dependency (does not acquire write lock)
  WRITE →  write authority (enforced inside write lock)
  [WM]     Working Memory (in-process store)
  [WMM]    WorkingMemoryManager (policy/access manager)
  ⚠        unauthorized write → error; never silently succeeds

─────────────────────────────────────────────────────────────
READ PATH (all cognitive subsystems)
─────────────────────────────────────────────────────────────

  Reasoning / Planning / Decision / Conversation Planner /
  Attention / Executive / Context Resolver / Understanding
       │
       │ READ →
       ▼
  WorkingMemoryManager [WMM]
       │ acquires read lock (per-scope RWMutex)
       │ determines: which goal context is foreground?
       │   reads ForegroundGoalID (atomic pointer)
       │ returns: requested slot value for the foreground goal context
       ▼
  Working Memory Store [WM]
       │ (passive in-process store — does NOT reason)
       ▼
  returned value to caller

─────────────────────────────────────────────────────────────
WRITE PATH (authorized writers only)
─────────────────────────────────────────────────────────────

  AUTHORIZED WRITER
  (Understanding, Context Resolver, Goal Manager,
   Reasoning, Planning, Executive, MemoryRetriever,
   Conversation Planner — per slot authority matrix above)
       │
       │ WRITE →
       ▼
  WorkingMemoryManager [WMM]
       │ acquires write lock (exclusive, per-scope)
       │ checks: is caller authorized for this slot? (enforced)
       │   ⚠ unauthorized → return error; log; never silently succeed
       │ enforces: capacity limits
       │   ⚠ over capacity → evict lowest-priority item(s) first
       │ applies: scope-level eviction policy
       ▼
  Working Memory Store [WM]
       │ (slot updated atomically within write lock)
       ▼
  (if checkpoint required: WMM triggers checkpoint write → core/storage)

─────────────────────────────────────────────────────────────
FOREGROUND GOAL SWITCH (Goal Manager only)
─────────────────────────────────────────────────────────────

  Goal Manager
       │ determines: new foreground goal
       │ WRITE → ForegroundGoalID [WM]
       │   atomic pointer swap (not a slot write lock — see 0.2.E)
       │   all subsequent reads now return new goal context
       ▼
  Working Memory correctly reflects new foreground goal

─────────────────────────────────────────────────────────────
INVARIANT: WORKING MEMORY DOES NOT TRIGGER COGNITION
─────────────────────────────────────────────────────────────

  Working Memory is passive. It stores and returns data.
  It never:
    - initiates pipeline stages
    - subscribes to events
    - calls cognitive subsystems
    - decides what is relevant
  Events trigger cognition. Working Memory is read during cognition.
```

---

#### 0.2.M — Restart Recovery Detailed Diagram

```text
═══════════════════════════════════════════════════════════════
WORKING MEMORY RESTART RECOVERY
═══════════════════════════════════════════════════════════════

SYSTEM RESTART
     │
     ▼
Goal Manager
     │ READ → core/storage: all persisted Goals + GapRecords
     │                       + PlanCheckpoints
     ▼
Check Goal.ExpiresAt for each recovered Goal (Decision 0.1, F4)
     ├── expired → archive; do not restore WM context
     └── active → continue
     │
     ▼
Restore GOAL-scoped Working Memory from checkpoint
     │ (via WorkingMemoryManager — not directly)
     │ WRITE → Working Memory [WM]
     │          ActiveGoalID, ActiveSubgoalIDs, ActiveGapIDs
     │          (ref-only — authoritative GapRecord stays in core/storage)
     │
     │ Validate gmStateVersion:
     │   WM checkpoint version == Goal Manager state version?
     │   ├── MATCH → use restored GOAL context
     │   └── MISMATCH → Goal Manager state wins
     │                   discard/re-derive WM context
     ▼
Restore EPISODE-scoped Working Memory from checkpoint (REQUIRED slots)
     │ WRITE → Working Memory [WM]
     │          PlanCheckpoint (REQUIRED)
     │          ActiveBeliefs (if RECOMMENDED checkpoint exists)
     │            ⚠ MARK AS TENTATIVE — must be re-evaluated by Reasoning
     │          UnresolvedQuestions (if RECOMMENDED checkpoint exists)
     │
     │ DO NOT RESTORE:
     │          TemporaryResearch — re-acquired at episode start
     │          RetrievedMemories — re-fetched at episode start
     │          RecentObservations — NOT checkpointed
     ▼
GLOBAL / SESSION scope:
     │ ConversationTurns — NOT restored (session starts fresh)
     │ ActiveTopic       — NOT restored (session starts fresh)
     │ CurrentEntities   — NOT restored (session starts fresh)
     │ RecentCorrections — NOT restored (session starts fresh)
     │ HostPreferences   — NOT restored (OPEN — policy required)
     │
     │ RATIONALE: stale session context from a previous session
     │ must not pollute a new session. Context Resolver would
     │ resolve references against incorrect prior-session state.
     ▼
TURN scope:
     │ PendingClarification — NOT restored (transient; token lost)
     │   If a ClarificationGap GapRecord exists in AWAITING_HOST_RESPONSE:
     │     Goal Manager re-emits via Conversation Planner (see 0.1.P.3)
     ▼
Re-dispatch uncertain GapRecords (Decision 0.1, 0.1.I)
     │ For each GapRecord{Status ∈ {DISPATCHED, RESOLVING}}:
     │   re-dispatch → Acquisition → idempotency check (F3)
     ▼
Trigger MemoryRetriever at episode resume
     │ WorkflowCoordinator (Executive) triggers retrieval
     │ MemoryRetriever → Long-Term Memory
     │ WRITE → Working Memory [WM]: RetrievedMemories
     │   (this is fresh retrieval, not a checkpoint restore)
     ▼
Resume foreground goal

═══════════════════════════════════════════════════════════════
SUMMARY: WHAT IS RESTORED vs WHAT STARTS FRESH
═══════════════════════════════════════════════════════════════

  RESTORED from checkpoint:
    GOAL:    ActiveGoalID, ActiveSubgoalIDs, ActiveGapIDs (refs)
    EPISODE: PlanCheckpoint (required)
             ActiveBeliefs (if exists; TENTATIVE)
             UnresolvedQuestions (if exists)

  STARTS FRESH:
    GLOBAL/SESSION: ConversationTurns, ActiveTopic,
                    CurrentEntities, RecentCorrections
    TURN:           PendingClarification
    EPISODE (non-checkpointed): TemporaryResearch,
                    RetrievedMemories, RecentObservations

  REHYDRATED (re-fetched, not restored):
    EPISODE: RetrievedMemories (MemoryRetriever at episode resume)
```

---

#### 0.2.N — Working Memory ↔ Long-Term Memory Boundary

```text
═══════════════════════════════════════════════════════════════
WORKING MEMORY ↔ LONG-TERM MEMORY BOUNDARY
═══════════════════════════════════════════════════════════════

LEGEND
  ──────→  synchronous data flow
  READ →   read dependency
  WRITE →  write authority
  [WM]     Working Memory (in-process)
  [LTM]    Long-Term Memory (persistent, external store)
  [WMM]    WorkingMemoryManager
  ⚠        boundary violation / prohibited

─────────────────────────────────────────────────────────────
RETRIEVAL DIRECTION (episode start)
─────────────────────────────────────────────────────────────

  WorkflowCoordinator (Executive)
  │ triggers: MemoryRetriever at episode start
  │ provides: episode context (GoalID, EpisodeID, domain hints)
  ▼
  MemoryRetriever [CONTRACT: intelligence/interfaces]
  │ READ → Long-Term Memory [LTM]
  │         query: relevant records for current episode context
  │         returns: semantic knowledge, durable facts,
  │                  historical observations, past goal summaries
  │ (LTM implementation: JSON/Vector/Graph — OPEN — policy)
  ▼
  WorkingMemoryManager [WMM]
  │ WRITE → Working Memory [WM]: RetrievedMemories (EPISODE scope)
  │ lifecycle: EPISODE — evicted at episode close
  │ NOT checkpointed — re-fetched on resume
  ▼
  Reasoning / Planning
  │ READ → Working Memory [WM]: RetrievedMemories
  │ Uses retrieved facts to inform reasoning and planning

─────────────────────────────────────────────────────────────
CONSOLIDATION DIRECTION (session close / episode close)
─────────────────────────────────────────────────────────────

  Episode closes
  │
  ▼
  Reflection (ASYNC — see §2.2 Canonical Lifecycle)
  │ READ → Working Memory [WM]: episode trace (read-only consumption)
  │ READ → Workspace episode data
  │ Produces: ReflectionReport (to Workspace)
  │ Reflection does NOT write to Working Memory
  │ Reflection does NOT write to Long-Term Memory directly
  ▼
  Learning (ASYNC)
  │ Consumes: ReflectionReport
  │ Generates: CandidateSnapshot (knowledge domain updates)
  ▼
  Rollout Executor (ASYNC)
  │ Activates CandidateSnapshot → Knowledge Store (LTM layer)
  │ WRITE → Long-Term Memory [LTM]
  │   (strictly unidirectional: Reflection → Learning → LTM)

─────────────────────────────────────────────────────────────
  ALSO (Working Memory consolidation at session end):
─────────────────────────────────────────────────────────────

  WorkingMemoryManager (on session/goal close)
  │ consolidates: relevant GLOBAL/GOAL-scoped content
  │   (conversation preferences, goal outcomes, entity updates)
  │ recommendation comes from Reflection (OPEN — exact policy)
  │ WorkingMemoryManager executes the consolidation
  │ WRITE → Long-Term Memory [LTM]
  │ Reflection recommends; WorkingMemoryManager executes (approved Decision 0.2)

─────────────────────────────────────────────────────────────
WHAT BELONGS WHERE
─────────────────────────────────────────────────────────────

  WORKING MEMORY [WM]        LONG-TERM MEMORY [LTM]
  ─────────────────────      ─────────────────────────────
  Current execution state    Semantic knowledge
  Active goal context        Durable facts
  Current plan / beliefs     Past goal summaries / outcomes
  Retrieved episode context  Conversation preferences (durable)
  Transient clarifications   Skill registry (AVAILABLE skills)
  Pending gap references     Knowledge store (acquired records)
  Recent observations        Episodic memory (past episodes)

─────────────────────────────────────────────────────────────
PROHIBITED BOUNDARY VIOLATIONS
─────────────────────────────────────────────────────────────

  ⚠ Working Memory MUST NOT pull from LTM directly (passive store)
  ⚠ Goal Manager MUST NOT trigger general memory retrieval
  ⚠ LTM MUST NOT own the authoritative goal lifecycle state
     (If LTM holds a "goal summary", it is a retrieval index only;
      Goal Manager in core/storage is authoritative)
  ⚠ Reflection MUST NOT write to WM or LTM directly
     (Reflection → ReflectionReport → Learning → CandidateSnapshot
      → Rollout Executor → LTM activation)
  ⚠ Learning MUST NOT write active goal state
     (Learning produces CandidateSnapshot for knowledge domains only)
```

---

#### 0.2.O — Implementation Wiring Map

```text
═══════════════════════════════════════════════════════════════
DECISION 0.2 — IMPLEMENTATION WIRING MAP
═══════════════════════════════════════════════════════════════

This map answers: when implementation begins, where do I connect this?

CONNECTION 1: Working Memory Store — package location
  Package:    intelligence/workingmemory (already exists in v3)
  Contents:   typed in-process store; slot definitions; scope rules
  Lifecycle:  in-process; no goroutines required by the store itself
  Dependency: intelligence/types (slot type definitions)

CONNECTION 2: WorkingMemoryManager — package location
  Package:    intelligence/workingmemory (WorkingMemoryManager type)
  Responsibility: capacity enforcement, scope eviction, checkpoint,
                  write-authority enforcement, foreground goal pointer
  Injection:  injected into all cognitive subsystems that read or write WM
  Dependency: core/storage (checkpoint writes)

CONNECTION 3: Goal Manager → WorkingMemoryManager (GOAL scope writes)
  Writer:     intelligence/goalmanager
  Target:     Working Memory GOAL scope via WorkingMemoryManager
  Slots:      ActiveGoalID, ActiveSubgoalIDs, ActiveGapIDs, ForegroundGoalID
  Ordering:   GOAL scope checkpoint must succeed BEFORE Goal Manager
              persists the corresponding Goal state change (§0.2.G)
  ⚠ Failure: checkpoint fail → STOP; do not advance goal state

CONNECTION 4: Understanding → WorkingMemoryManager (SESSION scope writes)
  Writer:     intelligence/understanding
  Target:     Working Memory SESSION scope via WorkingMemoryManager
  Slots:      ConversationTurns (append), ActiveTopic, CurrentEntities,
              RecentCorrections
  Lifecycle:  SESSION — not checkpointed

CONNECTION 5: Reasoning → WorkingMemoryManager (EPISODE scope writes)
  Writer:     intelligence/reasoning
  Target:     Working Memory EPISODE scope via WorkingMemoryManager
  Slots:      ActiveBeliefs, UnresolvedQuestions
  Checkpoint: RECOMMENDED

CONNECTION 6: Planning → WorkingMemoryManager (EPISODE scope writes)
  Writer:     intelligence/planning
  Target:     Working Memory EPISODE scope via WorkingMemoryManager
  Slots:      CurrentPlan, PlanCheckpoint
  Checkpoint: REQUIRED for PlanCheckpoint

CONNECTION 7: Conversation Planner → WorkingMemoryManager (SESSION scope writes)
  Writer:     intelligence/conversation (SS-20)
  Target:     Working Memory SESSION scope via WorkingMemoryManager
  Slots:      ConversationTurns (append outbound ConversativeIntent)
  Timing:     BEFORE Language Realization (Decision 0.3, AI-20)
  Lifecycle:  SESSION — not checkpointed

CONNECTION 8: Executive / WorkflowCoordinator → MemoryRetriever trigger
  Triggerer:  runtime/executive (WorkflowCoordinator)
  Interface:  MemoryRetriever [intelligence/interfaces]
  Target:     Long-Term Memory (implementation: OPEN — JSON/Vector/Graph)
  Timing:     episode start (before Reasoning reads RetrievedMemories)
  Result:     WRITE → Working Memory [WM]: RetrievedMemories via WMM

CONNECTION 9: WorkingMemoryManager → core/storage (checkpoint writes)
  Writer:     intelligence/workingmemory (WorkingMemoryManager)
  Target:     core/storage
  Contents:   GOAL scope snapshot + gmStateVersion
              PlanCheckpoint (required EPISODE slot)
              ActiveBeliefs, UnresolvedQuestions (if RECOMMENDED)
  Ordering:   gmStateVersion included in every checkpoint
  ⚠ Version mismatch on restore → Goal Manager state wins

CONNECTION 10: Goal Manager restart recovery
  Reader:     intelligence/goalmanager
  Source:     core/storage: Goals + GapRecords + PlanCheckpoints
  Sequence:   (1) check Goal.ExpiresAt
              (2) restore GOAL WM context via WMM
              (3) restore required EPISODE WM context via WMM
              (4) re-dispatch uncertain GapRecords (Decision 0.1)
              (5) trigger MemoryRetriever at episode resume
  Package:    PACKAGE LOCATION — within intelligence/goalmanager

CONNECTION 11: Reflection → Consolidation boundary
  Reflection:     intelligence/reflection — READ-ONLY from WM episode trace
  Emits:          ReflectionReport → Workspace
  Learning:       intelligence/learning — consumes ReflectionReport
  Emits:          CandidateSnapshot
  Rollout:        PACKAGE LOCATION — TO BE DEFINED DURING IMPLEMENTATION
  Activates:      Long-Term Memory update
  ⚠ Reflection MUST NOT write to WM; Learning MUST NOT write active state

CONNECTION 12: PendingClarification token lifetime (OPEN)
  Writer:     intelligence/goalmanager
  Target:     Working Memory TURN scope
  Lifetime:   OPEN — IMPLEMENTATION POLICY REQUIRED
              minimum: mechanical LRU eviction
              maximum: OPEN (survives turn? session? policy decision)
  Eviction:   WorkingMemoryManager enforces capacity; does not
              semantically evaluate token staleness
```

---



---

### Decision 0.3 — Conversation Planner: Cognitive or Presentation?

**Status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE

**Proposal:** Conversation Planner as part of the cognitive layer, orchestrating proactive communication.

**Resolution:** Conversation Planner is a **cognitive component** that determines *what* to communicate and *why*. It does not decide *how* to realize it (Language Realization) nor *when/how* to mechanically deliver it (Notification Service).

**The Five Required Refinements (Group 1):**
1. **Constitutional interception:** Every proactive `ConversativeIntent` must pass through the Constitutional Gate before Language Realization or Notification.
2. **Language Realization:** `ConversativeIntent` must go through SS-21 Language Realization to be rendered into natural language before reaching the Notification Service.
3. **Semantic batching:** Attention owns semantic grouping of related events; Notification Service only performs mechanical queueing, retry, rate limiting, deduplication, and delivery.
4. **Working Memory:** Conversation Planner records proactive outbound communication in `ConversationTurns` so immediate Host replies have the necessary context.
5. **Contracts:** Uses `SystemEvent` as the generic event envelope and `SalientEvent` as Attention's output (containing the original event plus salience/urgency metadata). Preserves the already-approved `ConversativeIntent`.

#### Diagram 1 — Complete proactive communication flow

```text
┌──────────────────────┐
│  Data Provenance     │
│  Legend              │
├──────────────────────┤
│ ──► Data/Event Flow  │
│ ──► Read Access      │
│ ═►  Sync Call        │
│ [x] External         │
│ (W) Working Memory   │
└──────────────────────┘

[Executive]   [Scheduler]   [Goal Manager]
       │             │              │ (Owners)
       └──────┬──────┴──────────────┘
              ▼
         SystemEvent (Contract)
              │
              ▼
     WORKSPACE EVENT BUS (Content-Blind Transport)
              │
              ▼
  ┌─────────────────────────┐
  │ EVENT ROUTER (SS-27)    │
  │ Owner: Runtime          │
  │ Trans: Check Autonomy   │
  └───────────┬─────────────┘
              ▼
  ┌─────────────────────────┐     READ (W)
  │ ATTENTION (SS-02)       │◄─── ActiveGoals, CurrentEntities
  │ Owner: Cognitive        │     from WORKING MEMORY (SS-06)
  │ Trans: Score Salience,  │
  │        Semantic Batching│
  └───────────┬─────────────┘
              ▼
         SalientEvent (Contract)
              │
              ▼
  ┌─────────────────────────┐     READ & WRITE (W)
  │ CONVERSATION PLANNER    │◄─── Reads WM Context
  │ (SS-20)                 │───► Writes to ConversationTurns
  │ Owner: Cognitive        │     (Lifecycle: TURN)
  │ Trans: Plan Strategy    │
  └───────────┬─────────────┘
              ▼
      ConversativeIntent (Contract)
              │
              ▼
  ┌─────────────────────────┐
  │ CONSTITUTIONAL GATE     │
  │ (SS-30)                 │
  │ Owner: Policy           │
  │ Trans: Safety check     │
  │ Failure: Drop & Log     │
  └───────────┬─────────────┘
              ▼
  ┌─────────────────────────┐
  │ LANGUAGE REALIZATION    │
  │ (SS-21)                 │
  │ Owner: Presentation     │
  │ Trans: Render NLG       │
  └───────────┬─────────────┘
              ▼
      Rendered Message
              │
              ▼
  ┌─────────────────────────┐
  │ NOTIFICATION SERVICE    │
  │ (SS-28)                 │
  │ Owner: Presentation     │
  │ Trans: Mech Queue, Retry│
  │        Rate Limit       │
  └───────────┬─────────────┘
              ▼
             HOST
```

#### Diagram 2 — Mid-episode communication

```text
  Active Goal (Lifecycle: GOAL)
       │
  Running Episode (Lifecycle: EPISODE)
       │
 AUTONOMY POLICY EVALUATION                                  [POLICY]
For non-Host events: what is IDUN permitted to do?
  → Route to full cognitive pipeline (if autonomy permits action)
  → Queue for Host confirmation (if action exceeds autonomy level)
  → Discard (if irrelevant and no notification warranted)
Note: all proactive notification paths go through the cognitive pipeline
Note: all proactive notification paths flow through Attention → Conversation Planner.
Notification Service is never routed to directly from Event Router.
       │
       ▼
  ATTENTION TRIAGE                                            [PIPELINE]
Receives: SystemEvent from Event Router.
Reads: Working Memory (ActiveGoals, CurrentEntities, ActiveTopic) — contextual salience scoring.
Evaluates: salience, urgency, contextual relevance.
Groups semantically related events where appropriate (semantic grouping is cognitive).
Assigns priority band (0–4). Reserves budget for this priority level.
Higher-band events preempt lower-band work in progress.
Emits: SalientEvent (original SystemEvent + salience score + urgency band + semantic grouping).
       │
       ▼
  [...pipeline...]

  --- LATER / CONCURRENTLY ---

  Host responds: "Stop researching that, do X instead."
       │
       ▼
  Understanding / Context Resolver 
       │ ◄── READS ConversationTurns from Working Memory (SS-06)
       ▼
  Reasoning / Planning → New GoalProposal
       │
       ▼
  Goal Manager
  (Pauses/Cancels active Executive episode based on new intent)
```

#### Diagram 3 — Responsibility boundary

```text
  Event Router (SS-27)
      = autonomy permission check & routing policy

  Attention (SS-02)
      = salience scoring + urgency + semantic grouping of related events

  Conversation Planner (SS-20)
      = communication strategy (IntentType, tone, verbosity)

  Constitutional Gate (SS-30)
      = safety, authorization, anti-spam check

  Language Realization (SS-21)
      = natural-language realization (NLG)

  Notification Service (SS-28)
      = mechanical delivery, queuing, deduplication
```

#### Diagram 4 — Data/contract flow

```text
    SystemEvent
    (Produced by: Exec/GoalMgr/Scheduler | Consumed by: EventRouter/Attention)
         │
         ▼
    Attention (reads WM)
         │
         ▼
    SalientEvent
    (Produced by: Attention | Consumed by: Conversation Planner)
         │
         ▼
    Conversation Planner (reads WM)
         │
         ▼
    ConversativeIntent
    (Produced by: ConvPlanner | Consumed by: Const. Gate / LangRealization)
         │
         ▼
    Constitutional Gate
         │
         ▼
    Language Realization
         │
         ▼
    Rendered Message
    (Produced by: LangRealization | Consumed by: Notification Service)
         │
         ▼
    Notification Service
```

#### Diagram 5 — Semantic vs mechanical batching

```text
  SEMANTIC BATCHING (Cognitive)
  Related events
       │
       ▼
  Attention (SS-02)
       │ (Understands relationships, groups into single SalientEvent)
       ▼
  Conversation Planner
       │ (Generates one combined ConversativeIntent)
       ▼
  [...pipeline...]

  VERSUS

  MECHANICAL BATCHING/DEDUPLICATION (Delivery)
  Rendered Messages
       │
       ▼
  Notification Service (SS-28)
       │ (No cognitive understanding)
       │ (Mechanical deduplication: drop identical messages)
       │ (Queue / retry / rate limit)
       ▼
  Host
```

---

### Decision 0.4 — The Recursion Problem in Knowledge Acquisition

**Status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE

**Problem:** Reasoning detects a KnowledgeGap (or SkillGap) during plan execution. The naive solution — have Reasoning invoke itself recursively after acquisition — creates unbounded depth, unclear episode boundaries, and stack-overflow risk for chained acquisition.

**Decision:** Knowledge acquisition triggers a **Continuation Episode**. The original Reasoning does not recurse. The current Episode closes at the gap boundary. A new, bounded Episode starts after acquisition completes, restoring the execution state from a persisted checkpoint.

This is architecturally equivalent to a Saga step / Bounded Task in a durable workflow: the Goal is the long-running workflow; each Episode is a bounded task step.

---

#### 0.4.A — Episode Identity Across the Continuation Boundary

**Principle P3 (Acquisition is not recursion)** is implemented through strict Episode identity rules:

- A **Goal** is a long-lived intention that persists across any number of Episodes.
- An **Episode** is a *bounded* unit of cognitive work. It must have a clear start and a clear close.
- An Episode that would wait hours or days for external acquisition is no longer bounded. Therefore the gap boundary is the Episode boundary.

**Rules:**

1. `GoalID` remains unchanged across E1 and E2. The Goal is continuous.
2. `EpisodeID` is always new for the continuation (E2). E1's EpisodeID is never reused.
3. E1 closes when it emits the `GapSignal`. E2 begins when Goal Manager schedules the continuation after `GapResolved`.
4. Do NOT add a `ContinuationOf` field linking E2 to E1. Correlation is achieved through `GoalID` + `StartedAt` ordering, which is the standard Correlation ID pattern.
5. The sequence of Episodes under a Goal can be reconstructed by querying: all Episodes where `GoalID = G1`, ordered by `StartedAt`.

#### 0.4.B — Episode Closure at a Gap Boundary

When E1 closes due to a KnowledgeGap or SkillGap:

- `EpisodeRecord.Outcome = PARTIAL`
- `EpisodeRecord.TerminationReason = GAP_PENDING`

**Do NOT create a new `EpisodeOutcome` value** such as `PAUSED_GAP_PENDING`. The existing `Outcome` field describes the qualitative result of the work done; the new `TerminationReason` field describes the mechanical reason execution stopped. These are orthogonal concerns and must be separate fields.

**Reflection:** Reflection MUST run after a gap-closed Episode (Outcome=PARTIAL, TerminationReason=GAP_PENDING). Reflection should evaluate the cognitive work done up to the gap boundary. The ReflectionReport naturally scopes to the segment of work completed. Running Reflection on partial Episodes provides metacognitive oversight of the reasoning that led to the gap, which is valuable for learning and debugging.

#### 0.4.C — PlanCheckpoint and State Transfer

`PlanCheckpoint` remains **EPISODE-scoped** (Decision 0.2 is not changed). The checkpoint belongs to E1's execution state as a runtime artifact. However, its *persisted representation* can be transferred to E2's EPISODE scope through an explicit restore operation.

**Checkpoint transfer model:**

```text
═══════════════════════════════════════════════════════════════
DIAGRAM B — PLANCHECKPOINT TRANSFER (MODEL B)
═══════════════════════════════════════════════════════════════

EPISODE E1 (active)
  │
  │ Reasoning detects KnowledgeGap
  │
  │ Planning writes PlanCheckpoint into E1 EPISODE scope
  ▼
WorkingMemoryManager
  │ checkpoint persistence: Checkpoint(ctx, episodeID=E1)
  │ WRITE → core/storage (key: EpisodeID=E1_checkpoint)
  │ atomic write — all GOAL + required EPISODE slots
  ▼
core/storage [PERSISTED]
  │ E1 checkpoint stored
  │ E1 EPISODE scope drops from live RAM (episode closed)
  │
  │ ← acquisition runs (arbitrarily long) ←
  │
  │ E2 begins (Goal Manager schedules ContinuationRequest)
  ▼
WorkingMemoryManager
  │ restore: Restore(ctx, sourceEpisodeID=E1, targetEpisodeID=E2)
  │ READ ← core/storage (key: E1_checkpoint)
  │ WRITE → E2 EPISODE scope (PlanCheckpoint, ActiveBeliefs[TENTATIVE])
  ▼
EPISODE E2 (fresh EPISODE scope, plan checkpoint loaded)
  │
  ▼
Planning / Reasoning resume from PlanCheckpoint

KEY POINTS:
  - E1's live EPISODE Working Memory does NOT survive. Only the
    persisted checkpoint survives.
  - E2 gets a fresh EPISODE scope populated via explicit Restore().
  - PlanCheckpoint belongs to E2's EPISODE scope at runtime.
  - Goal scope (ActiveGoalID, ActiveSubgoalIDs) remains unchanged.
  - SESSION scope (ConversationTurns, ActiveTopic) is unaffected.
```

**WorkingMemoryManager interface update (Decision 0.2 synchronization):**
The existing `Restore(ctx, episodeID)` signature implies restoring state back into the same episode. The continuation model requires loading E1's checkpoint into E2's scope. The architectural intent is:

```
Restore(ctx, sourceEpisodeID, targetEpisodeID) error
```

This makes the cross-episode state transfer boundary explicit. The exact parameter types are an implementation specification.

#### 0.4.D — Acquired Knowledge Injection (Q0.4-5)

The Goal Manager must NOT interpret, summarize, classify, or semantically inject acquired knowledge. It carries only the correlation/reference information necessary to route execution to the right data.

Two acquisition paths exist based on the durability of the acquired result:

**Path A — Durable / Verified Knowledge:**
Knowledge Acquisition has high confidence in the result. It writes to the Knowledge Store (LTM) and emits a `StoreKey` reference.

```text
═══════════════════════════════════════════════════════════════
DIAGRAM C — KNOWLEDGE INJECTION PATHS
═══════════════════════════════════════════════════════════════

PATH A (Durable)                    PATH B (Transient)
                                    
Knowledge Acquisition               Knowledge Acquisition
        │                                   │
        │ WRITE → Knowledge Store           │ (no LTM write)
        │         (insert-if-absent, F3)    │
        │                                   │
        ▼                                   ▼
  GapResolved                         GapResolved
  {GapID, Status=SUCCESS,             {GapID, Status=SUCCESS,
   StoreKey="k1"}                      Payload=[opaque bytes]}
        │                                   │
        ▼                                   ▼
            GOAL MANAGER
            (does NOT read Payload or StoreKey semantics)
            Creates ContinuationRequest:
            {GoalID, SourceEpisodeID=E1,
             GapResolutionHint{GapID, StoreKey="k1" or nil},
             TransientPayload=[opaque] or nil}
                    │
                    ▼
                Executive (E2 init)
                /                \
               /                  \
              ▼                    ▼
     GapResolutionHint          TransientPayload
              │                    │
              ▼                    ▼
      MemoryRetriever          Inject directly
    (mandatory StoreKey         into E2 WM
     fetch + general            TemporaryResearch
     relevance query)           slot
              │                    │
              └──────────┬─────────┘
                         ▼
                  E2 Working Memory
                (RetrievedMemories +
                 TemporaryResearch)
                         │
                         ▼
                    Reasoning
```

**Path B — Transient / Unverified Result:**
Knowledge Acquisition has low confidence or acquired context that is inherently temporal (e.g., "current weather in Tokyo"). Writing this to LTM would pollute the durable semantic store with transient facts. Instead:

- The result is stored as an opaque transient payload in `GapResolved` and carried through `GapRecord` and `ContinuationRequest`.
- Goal Manager treats this payload as opaque bytes — it does not read or interpret it.
- Executive injects the payload directly into E2's `TemporaryResearch` slot.
- `TemporaryResearch` is EPISODE-scoped and NOT persisted — if E2 encounters another gap, the transient payload must be re-acquired.

**LTM boundary preservation:**
LTM = durable semantic knowledge.
GapRecord = durable workflow/execution state (carries transient payload tied to the Goal's lifecycle).
Working Memory = active cognitive context.

#### 0.4.E — GapResolutionHint: Preventing Gap Recurrence

Without a targeted retrieval hint, MemoryRetriever's general relevance query might not rank the newly acquired knowledge highly enough to include it in E2's `RetrievedMemories`. This would cause Reasoning to encounter the identical KnowledgeGap again — creating an infinite pause/acquire loop.

The `GapResolutionHint {GapID, StoreKey}` is a lightweight storage correlation reference (adapting the **Claim Check** industry pattern). It does not carry semantic meaning. It allows MemoryRetriever to guarantee inclusion of the gap-resolution record alongside its normal relevance retrieval.

Goal Manager already possesses `StoreKey` because it is included in `GapResolved`. Passing it forward as a hint does not require Goal Manager to understand the knowledge semantics.

#### 0.4.F — Continuation Episode Diagrams

```text
═══════════════════════════════════════════════════════════════
DIAGRAM A — FULL EPISODE LIFECYCLE (E1 → ACQUISITION → E2)
═══════════════════════════════════════════════════════════════

LEGEND
  ──────→  synchronous data flow
  - - - →  async / event
  [STORE]  persisted to core/storage or LTM
  [WM]     Working Memory (in-process)
  ★        new contract/field required

GOAL G1 (GoalID="g1", State=ACTIVE)
  │  Owner: Goal Manager
  ▼
EPISODE E1 (EpisodeID="e1", GoalID="g1")
  │  Owner: Executive
  │  StartedAt: T0
  │
  │ [PIPELINE: Understanding → Context → Reasoning → Planning]
  │
  │ Reasoning detects KnowledgeGap
  │   Producer: Reasoning
  │   Event: GapSignal → Event Bus → Goal Manager (F1: GapRecord written first)
  │
  │ Planning writes PlanCheckpoint
  │   Producer: Planning
  │   WRITE → Working Memory [WM]: PlanCheckpoint (EPISODE scope)
  │
  │ WorkingMemoryManager.Checkpoint(episodeID=E1)
  │   WRITE → core/storage: E1 checkpoint [STORE]
  │
  │ Goal.State transitions: ACTIVE → PAUSED
  │   Owner: Goal Manager
  │   WRITE → core/storage: Goal, GapRecord [STORE]
  │
  │ E1 CLOSES:
  │   CompletedAt: T1
  │   Outcome: PARTIAL
  │   TerminationReason: GAP_PENDING ★
  │   Producer: Executive
  │   WRITE → EpisodeRecord E1 [STORE]
  │
  │ Reflection(E1) — triggered async by Executive
  │   Reads: E1 episode trace (PARTIAL, GAP_PENDING)
  │   Emits: ReflectionReport for E1 to Workspace [STORE]
  ▼

ACQUISITION (async, arbitrarily long)
  │  Owner: Knowledge Acquisition
  │  Input: GapSignal {GapID, GoalID, Domain}
  │
  │  Idempotency check (F3): exists(GapID) in Knowledge Store?
  │    YES → emit GapResolved(SUCCESS) immediately
  │    NO  → acquire → evaluate → persist
  │
  │  Path A (Durable): WRITE → Knowledge Store [STORE] (insert-if-absent)
  │  Path B (Transient): no LTM write; result in GapResolved.Payload
  │
  ▼
GapResolved {GapID, GoalID, Status=SUCCESS, StoreKey?, Payload?}
  │  Producer: Knowledge Acquisition
  │  Consumer: Goal Manager
  │  Transport: Event Bus (content-blind)
  │
  │ Goal Manager checks Goal.State == CANCELLED? (F12)
  │   YES → ignore; no E2 scheduled
  │   NO  → continue
  │
  │ WRITE → GapRecord.Status = RESOLVED [STORE]
  │ WRITE → Goal.State = ACTIVE [STORE]
  ▼

ContinuationRequest ★ [CONTRACT REQUIRED]
  {GoalID="g1", SourceEpisodeID="e1",
   GapResolutionHint{GapID, StoreKey?},
   TransientPayload?}
  │  Producer: Goal Manager
  │  Consumer: Executive
  │  (mechanism: ARCHITECTURE boundary defined — implementation open)
  ▼

EPISODE E2 (EpisodeID="e2", GoalID="g1")
  │  Owner: Executive
  │  StartedAt: T2
  │
  │ E2 INITIALIZATION:
  │   (1) WorkingMemoryManager.Restore(source="e1", target="e2")
  │         READ ← core/storage: E1 checkpoint
  │         WRITE → E2 WM [WM]: PlanCheckpoint, ActiveBeliefs[TENTATIVE]
  │
  │   (2) MemoryRetriever invoked by Executive
  │         Input: GoalID, EpisodeID=E2, GapResolutionHint
  │         READ ← Knowledge Store (relevance query + mandatory StoreKey fetch)
  │         WRITE → E2 WM [WM]: RetrievedMemories (EPISODE scope)
  │
  │   (3) If TransientPayload present:
  │         Executive injects directly
  │         WRITE → E2 WM [WM]: TemporaryResearch (EPISODE scope)
  │
  │ [PIPELINE resumes: Planning reads PlanCheckpoint → Reasoning continues]
  │ → KnowledgeGap does NOT recur (GapResolutionHint guarantees inclusion)
  ▼
Continue Goal G1
```

#### 0.4.G — E1 → E2 State Survival Map

```text
═══════════════════════════════════════════════════════════════
DIAGRAM D — E1 → E2 SURVIVAL MAP
═══════════════════════════════════════════════════════════════

SCOPE      SLOT                SURVIVES E1→E2?    HOW
───────────────────────────────────────────────────────────────

SESSION    ConversationTurns   YES                Session continues; not episode-bound
(GLOBAL)   ActiveTopic         YES
           CurrentEntities     YES
           RecentCorrections   YES

───────────────────────────────────────────────────────────────

GOAL       ForegroundGoalID    YES                Goal still active; unchanged
           ActiveGoalID        YES
           ActiveSubgoalIDs    YES
           ActiveGapIDs        UPDATED            gap1 removed (RESOLVED); no new gaps at E2 start

───────────────────────────────────────────────────────────────

EPISODE    PlanCheckpoint      YES (via Restore)  Persisted in E1 checkpoint;
                                                   loaded into E2 EPISODE scope by WMM.Restore()
                                                   [EPISODE scope at runtime in E2]

           ActiveBeliefs       YES (TENTATIVE)    Restored from E1 checkpoint;
                                                   marked TENTATIVE for Reasoning re-evaluation

           UnresolvedQuestions YES (if checkpointed)

           CurrentPlan         NO → RE-DERIVED    Planning derives current plan from PlanCheckpoint

           RecentObservations  NO                 Episode-bounded; belongs historically to E1

           TemporaryResearch   NO → INJECTED      E1's slot dies on E1 close;
                                                   E2 gets new injection from TransientPayload
                                                   (if acquisition was transient/unverified)

           RetrievedMemories   NO → RE-FETCHED    Fresh MemoryRetriever fetch at E2 init
                                                   GapResolutionHint ensures gap result is included

───────────────────────────────────────────────────────────────

TURN       PendingClarification NO                TURN-scoped; never survives turn boundary

───────────────────────────────────────────────────────────────

NEW        GapResolutionHint   TRANSIENT          Passed from Goal Manager → Executive;
(transient) {GapID, StoreKey}                     consumed at E2 init; NOT stored as WM slot
```

#### 0.4.H — Failure and Recovery

```text
═══════════════════════════════════════════════════════════════
FAILURE/RECOVERY MATRIX
═══════════════════════════════════════════════════════════════

FAILURE MODE                 WHO DETECTS      STATE TRANSITION / RECOVERY
─────────────────────────────────────────────────────────────────────────────

Acquisition fails            Goal Manager     Level 2 timeout → retry (if RetryCount < Max)
(all retries exhausted)                       → Goal AWAITING_HOST → Conversation Planner
                                               notifies Host

Knowledge Store write fails  Acquisition      Does not emit GapResolved(SUCCESS)
(persistence failure)                         → GapRecord remains RESOLVING
                                               → Level 2 timeout → retry → idempotent re-acquire

Idun crash during E1         Goal Manager     E1 WM lost; GapRecord UNCERTAIN on restart
(checkpoint written)         on restart       → re-dispatch; idempotency (F3) safe

Idun crash during            Goal Manager     GapRecord UNCERTAIN → re-dispatch
acquisition                  on restart       → idempotency (F3) returns SUCCESS immediately
                                               if LTM write had already succeeded

Idun crash after LTM write   Goal Manager     GapRecord UNCERTAIN → re-dispatch
but before GapResolved        on restart       → idempotency (F3) returns SUCCESS; schedules E2

Idun crash before E2 starts  Goal Manager     GapRecord RESOLVED but no EpisodeRecord for E2
(after GapResolved received)  on restart       → Goal Manager re-schedules ContinuationRequest
                                               NOTE: ordering contract between GapRecord RESOLVED
                                               and E2 scheduling must be atomic or idempotent
                                               [IMPLEMENTATION CONTRACT REQUIRED]

MemoryRetriever fails        Executive        Retry E2 initialization; if exhausted →
at E2 init                                    escalate to Goal Manager → Goal AWAITING_HOST

Goal cancelled (F12)         Goal Manager     Goal.State == CANCELLED check on GapResolved
                                               → drop result; no E2; KnowledgeStore record remains

Goal expires (F4)            Goal Manager     ExpiresAt check at restart / deadline monitor
                                               → archive goal; no E2
```

#### 0.4.I — New Contracts Required

| Contract | Architectural Purpose | Status |
|:---------|:---------------------|:-------|
| `GapResolved` | Extended with: `StoreKey?` (reference to durable LTM record), `Payload?` (opaque transient acquisition result) | Update existing contract |
| `ContinuationRequest` | Carries: `GoalID`, `SourceEpisodeID`, `GapResolutionHint?{GapID, StoreKey}`, `TransientPayload?` | NEW — not yet defined |
| `GapResolutionHint` | Lightweight LTM reference enabling MemoryRetriever to guarantee inclusion of the gap-resolution record | NEW — not yet defined |
| `EpisodeRecord.TerminationReason` | New field on EpisodeRecord; value `GAP_PENDING` for gap-closed episodes | NEW field |
| `WMM.Restore(src, target)` | Two-argument restore enabling cross-episode checkpoint transfer | Update existing interface |

#### 0.4.J — Industry Pattern Mapping

| Idun Architectural Choice | Established Pattern | What It Solves | IDUN Adaptation |
|:---|:---|:---|:---|
| New EpisodeID | Saga Step / Bounded Task | Long-running workflows cannot hold compute resources during external I/O waits | Goal = Saga; Episode = bounded step |
| EPISODE checkpoint + Restore | State Rehydration / Snapshot | Workflow state is serialized at step closure and rehydrated at next step start | PlanCheckpoint persisted at E1 close; loaded into E2 EPISODE scope |
| GoalID + StartedAt correlation | Correlation ID | Avoids fragile linked lists; query-time reconstruction | Replaces `ContinuationOf` pointer |
| GapResolutionHint(StoreKey) | Claim Check | Passes a lightweight ticket through the orchestrator; redeemed at the retrieval layer | Goal Manager holds the key; MemoryRetriever redeems it |
| GapRecord.Payload (transient) | Saga State / Process Manager | Intermediate non-global step results held in orchestrator state, not in global DB | Transient knowledge stays out of LTM |
| LTM → MemoryRetriever → WM | Durable-State Rehydration | Pipeline reconstructs worldview from authoritative semantic database | Same pattern; GapResolutionHint adds mandatory-inclusion guarantee |


#### 0.4.K — Q0.4-1: Acquisition Execution Model (Bounded Worker)

Knowledge Acquisition is **not a cognitive Episode**. It is a **Bounded Worker** — a specialized service that receives a `GapSignal`, performs a fixed internal pipeline, and emits `GapResolved`. It does not instantiate the Executive, does not create an EpisodeRecord, and does not run any cognitive pipeline stage.

**Authoritative model:**

```text
═══════════════════════════════════════════════════════════════
DIAGRAM K1 — KNOWLEDGE ACQUISITION BOUNDED WORKER
═══════════════════════════════════════════════════════════════

LEGEND
  ──────→  synchronous data flow
  - - - →  async / event
  [STORE]  persisted to core/storage or LTM
  ⚠        boundary that must NOT be crossed

GOAL MANAGER
  │  Dispatches via routeGap()
  │  Event Bus (content-blind transport)
  ▼
KNOWLEDGE ACQUISITION (intelligence/knowledge/acquisition)
─────────────────────────────────────────────────────────
WHAT IT IS: Bounded Worker — specialized service
WHAT IT IS NOT: Cognitive Episode, Executive client, pipeline stage

Internal fixed pipeline:
  │
  ├── Receive GapSignal {GapID, GoalID, Domain, GapType}
  │     Producer: Goal Manager (via routeGap)
  │     Consumer: Knowledge Acquisition
  │     Transport: Event Bus (content-blind)
  │
  ├── Idempotency check (F3)
  │     READ → Knowledge Store: exists(GapID)?
  │       YES → emit GapResolved(SUCCESS, StoreKey) immediately
  │       NO  → continue
  │
  ├── Classify acquisition requirement
  │     factual-stable / factual-temporal / contextual / procedural / causal
  │
  ├── Select acquisition strategy
  │     Local:  Knowledge Store → Episodic Memory (analogy)
  │     Online: Tool use / Internet research
  │     Host:   ClarificationGap path (if automated acquisition insufficient)
  │
  ├── Acquire raw information from selected source
  │
  ├── Normalize and extract relevant claims
  │
  ├── Evaluate source quality (confidence score 0.0–1.0)
  │     Source type, freshness, cross-reference, internal consistency, provenance
  │
  ├── Classify result durability:
  │     VERIFIED durable   → persist to Knowledge Store (insert-if-absent, F3)
  │     UNVERIFIED durable → persist to Knowledge Store (UNVERIFIED flag)
  │     TRANSIENT          → do NOT persist to LTM; carry as opaque Payload
  │     CONFLICTED         → persist with ConflictFlag; notify Host
  │     FAILURE            → emit GapResolved(FAILURE); do not persist
  │
  ├── WRITE → Knowledge Store [STORE] (if durable)
  │     insert-if-absent keyed by GapID (F3)
  │     ⚠ partial write → rolled back; do NOT emit GapResolved(SUCCESS)
  │
  └── Emit GapResolved {GapID, GoalID, Status, StoreKey?, Payload?}
        Producer: Knowledge Acquisition
        Consumer: Goal Manager (only)
        Transport: Event Bus (content-blind)

─────────────────────────────────────────────────────────
STAGES KNOWLEDGE ACQUISITION DOES NOT USE:
  ⚠ Executive                (does not orchestrate episodes)
  ⚠ Working Memory (goal)    (does not need goal-scoped WM)
  ⚠ Understanding             (no natural-language input to parse)
  ⚠ Context Resolution        (no dialogue state to resolve)
  ⚠ Reasoning (cognitive)     (confidence scoring is algorithmic)
  ⚠ Planning                  (has fixed pipeline; not goal-directed)
  ⚠ Decision                  (no plan selection required)
  ⚠ Episode lifecycle         (is NOT an episode)
─────────────────────────────────────────────────────────
WHY NOT A COGNITIVE EPISODE:
  An Episode (Decision 0.6) = "bounded unit of cognitive work."
  Acquisition is bounded but not cognitive. It acquires facts;
  it does not reason about their meaning for the Goal.
  Industry pattern: Temporal/Saga Activity — a bounded task
  invoked by the orchestrator (Goal Manager), not a sub-workflow.
```

**Ownership:**
- **Owns:** Strategy selection, source evaluation, normalization, confidence scoring, cross-reference, Knowledge Store writes, contradiction detection, GapResolved emission.
- **Must NOT own:** Reasoning over acquired knowledge, deciding whether knowledge is useful for the Goal, creating Episodes, invoking the Executive, writing to goal-scoped Working Memory.
- **Must NOT become:** A cognitive subsystem that understands the meaning of what it acquired. It evaluates quality; Reasoning evaluates meaning.

**Industry pattern:**

| Pattern | Problem solved | IDUN mapping | IDUN adaptation |
|:--------|:--------------|:-------------|:----------------|
| Temporal Activity / Saga Step | Long-running external work should not block the orchestrator's state machine | Knowledge Acquisition = Activity; Goal Manager = orchestrator | LTM write authority stays with Acquisition (vs. passing payload through orchestrator for durable knowledge) |

---

#### 0.4.L — Q0.4-2: Goal Manager Signals, Executive Acts

**Architectural rule:**

Goal Manager owns the lifecycle. It determines *that* a continuation is needed. Executive owns execution. It determines *how* and *when* to run E2. These responsibilities must never cross.

**Goal Manager must NOT:**
- Directly invoke Executive methods (creates tight coupling and gives Goal Manager execution responsibility).
- Synchronously wait for E2 to complete (makes Goal Manager an execution monitor).
- Execute or manage E2 internally.
- Know the internal pipeline stages of E2.
- Become an execution engine.

**Executive must NOT:**
- Poll Goal Manager for continuation work (wasteful, introduces latency, couples Executive to Goal Manager's internal state).

```text
═══════════════════════════════════════════════════════════════
DIAGRAM L1 — CONTINUATION SIGNALING BOUNDARY
═══════════════════════════════════════════════════════════════

GOAL MANAGER
  │  Owner: intelligence/goalmanager
  │
  │  On GapResolved(SUCCESS):
  │    (1) Check Goal.State == CANCELLED (F12) → ignore if CANCELLED
  │    (2) Check Goal.ExpiresAt (F4)           → archive if expired
  │    (3) WRITE → GapRecord.Status = RESOLVED [STORE]
  │    (4) WRITE → GapRecord.ResolvedAt = now  [STORE]  ← required for restart recovery
  │    (5) WRITE → Goal.State = ACTIVE         [STORE]
  │    (6) Emit ContinuationRequest (see contract below)
  │
  │  The architectural boundary:
  │    Goal Manager SIGNALS that continuation is required.
  │    Goal Manager does NOT call Executive methods.
  │    Goal Manager does NOT wait for E2.
  │
  ▼
ContinuationRequest [CONTRACT]
  {
    GoalID          : identifies the continuing Goal
    SourceEpisodeID : E1 — the episode whose checkpoint to restore
    GapResolutionHint : { GapID, StoreKey? }
    TransientPayload  : opaque bytes (if acquisition was transient)
  }
  ⟵ Serialization format: IMPLEMENTATION SPECIFICATION
  ⟵ Transport mechanism: IMPLEMENTATION SPECIFICATION
     (Event Bus, internal async channel, or equivalent.
      The Blueprint defines the contract; the implementation defines the wire.)

  │
  ▼
[delivery mechanism — IMPLEMENTATION SPECIFICATION]
  │
  ▼
EXECUTIVE
  │  Owner: runtime/executive
  │  Receives ContinuationRequest
  │  Owns: creation of E2, full E2 lifecycle
  ▼
CREATE EPISODE E2 (see §0.4.M)
```

**Ownership table:**

| Component | Responsibility | Must NOT |
|:----------|:--------------|:---------|
| Goal Manager | Persist state change → Emit ContinuationRequest | Wait for E2 / invoke Executive / manage E2 |
| Executive | Receive ContinuationRequest → Create E2 → Run E2 | Poll Goal Manager / modify Goal state |

#### 0.4.L.1 — Restart Recovery: Lost Continuation

**The crash scenario:**

```text
GapResolved received
        │
        ▼
Goal Manager
        │
        ├── WRITE → GapRecord.Status = RESOLVED [STORE]  ← persisted ✓
        ├── WRITE → Goal.State = ACTIVE [STORE]           ← persisted ✓
        │
        X ← IDUN CRASHES HERE
        │
        │  Executive never receives ContinuationRequest
        │  E2 never created
        │
        ▼
SYSTEM RESTART

GOAL MANAGER RESTART RECOVERY:
        │
        ▼
Scan: all GapRecords where Status = RESOLVED
        │
        ▼
For each RESOLVED GapRecord:
        │
        ├── Does an EpisodeRecord exist where
        │     GoalID = this GapRecord.GoalID AND
        │     EpisodeType = CONTINUATION AND
        │     StartedAt > GapRecord.ResolvedAt?
        │
        ├── YES → continuation was started; no action needed
        │
        └── NO  → continuation was lost; re-emit ContinuationRequest
```

**This is a new restart invariant:** Goal Manager must re-emit `ContinuationRequest` for any `GapRecord{Status=RESOLVED}` that has no corresponding Continuation Episode. This prevents a Goal from remaining permanently stuck in ACTIVE with no continuation running.

**Implementation contract:** `GapRecord.ResolvedAt` (timestamp) is required to make the "started after resolution" check unambiguous. The exact field type is an implementation specification.

**Industry pattern:** This is standard Saga compensation / idempotent orchestration. The orchestrator (Goal Manager) is responsible for re-emitting its own outputs on restart if downstream processing was not confirmed.

---

#### 0.4.M — Q0.4-4: E2 Is a Normal Bounded Episode with Continuation Initialization

**Architectural rule:** E2 is a fully normal bounded Episode under Decision 0.6. It uses the standard Episode lifecycle machinery (EpisodeRecord, Working Memory, pipeline orchestration). The difference is not in the machinery — it is in the initialization: E2 is initialized from a `ContinuationRequest` and a persisted checkpoint rather than from a fresh Host stimulus.

**Key distinction:** E2 does not "jump into the middle of a program." E2 is initialized with the continuation state, and the normal cognitive pipeline proceeds from the state appropriate to that continuation. The pipeline is not altered — only the initialization inputs differ.

```text
═══════════════════════════════════════════════════════════════
DIAGRAM M1 — E2 INITIALIZATION AND PIPELINE ENTRY
═══════════════════════════════════════════════════════════════

EXECUTIVE receives ContinuationRequest
  │
  ▼
(1) CREATE EpisodeRecord E2
      EpisodeID:    new (never reuses E1)
      GoalID:       same as E1 (Goal G1 continues)
      EpisodeType:  CONTINUATION  ← metadata; no pipeline behavior change
      StartedAt:    now
      [WRITE → Episodic Memory store]

(2) WorkingMemoryManager.Restore(source=E1, target=E2)
      READ  ← core/storage: E1 checkpoint
      WRITE → E2 EPISODE WM: PlanCheckpoint
      WRITE → E2 EPISODE WM: ActiveBeliefs [TENTATIVE]
      WRITE → E2 EPISODE WM: UnresolvedQuestions (if checkpointed)

(3) MemoryRetriever invoked by Executive
      Inputs: GoalID, EpisodeID=E2, GapResolutionHint{GapID, StoreKey?}
      READ ← Knowledge Store:
        (a) general relevance query for this Goal/Episode context
        (b) mandatory StoreKey fetch (GapResolutionHint) — guaranteed inclusion
      WRITE → E2 EPISODE WM: RetrievedMemories

(4) If ContinuationRequest.TransientPayload present:
      Executive injects directly
      WRITE → E2 EPISODE WM: TemporaryResearch (EPISODE scope, NOT persisted)

────────────────────────────────────────────────────────────────
E2 COGNITIVE PIPELINE ENTRY
────────────────────────────────────────────────────────────────

  UNDERSTANDING:  ⊘ NOT EXECUTED
    Reason: No new Host stimulus exists. There is no natural-language
    input to parse. Running Understanding on a synthetic re-statement
    would introduce hallucinated restatement noise.

  CONTEXT RESOLUTION:  ⊘ NOT EXECUTED as fresh-input stage
    Reason: The continuation context already contains the relevant
    resolved state from E1's checkpoint. SESSION-scope WM (ConversationTurns,
    ActiveTopic, CurrentEntities) carries the session state forward normally.
    There is no new dialogue input to resolve.

  REASONING:  ✓ ENTRY POINT
    Reads:  E2 EPISODE WM: ActiveBeliefs [TENTATIVE]
            E2 EPISODE WM: RetrievedMemories (includes resolved knowledge)
            E2 EPISODE WM: TemporaryResearch (if transient payload)
            SESSION WM:    ConversationTurns, ActiveTopic, CurrentEntities
    Actions:
      - Re-evaluates TENTATIVE beliefs in the context of new knowledge
      - Verifies the original gap has been resolved
      - Updates beliefs: TENTATIVE → CONFIRMED | CONTRADICTED | UPDATED
      - Emits: ReasoningResult

  PLANNING:  ✓ NEXT STAGE
    Reads:  E2 EPISODE WM: PlanCheckpoint → deserializes to identify
                            next plan step
    Actions:
      - Resumes plan from the paused step identified by PlanCheckpoint
      - Does NOT re-plan from scratch
      - Generates: resumed CandidatePlan

  DECISION → EXECUTIVE → COMMUNICATION:  ✓ NORMAL PIPELINE
    All stages run normally from this point.

────────────────────────────────────────────────────────────────
E2 EPISODE TYPE: CONTINUATION
────────────────────────────────────────────────────────────────
  EpisodeType = CONTINUATION is metadata only.
  It does NOT change pipeline behavior.
  It enables: analytics, Reflection queries ("how many continuations
  did this goal require?"), Episodic Memory reconstruction.
```

**MemoryRetriever failure at E2 init:**
- Executive retries E2 initialization (locally, without Goal Manager involvement).
- If retries exhausted → Executive reports failure to Goal Manager.
- Goal Manager transitions Goal to `AWAITING_HOST`.
- Conversation Planner generates Host notification (following Decision 0.3 pipeline).

---

#### 0.4.N — Gap Recurrence Protection

If Reasoning in E2 encounters the same KnowledgeGap that E1 encountered (GapResolutionHint failed or retrieval was incomplete):

```text
E2 Reasoning
     │
     ▼
GapSignal{GapID=gap1}   ← same GapID as E1's gap
     │
     ▼
Goal Manager
     │
     ├── Check: has this GapID appeared in a previous
     ?          GapRecord for this GoalID?
     ?
     ??? YES: recurrence detected
     ?         Increment: GapRecord.RecurrenceCount
     ?         IF RecurrenceCount < policy threshold:
     ?           ? permitted: re-dispatch to Acquisition
     ?             (Acquisition idempotency: if already exists, SUCCESS immediately)
     ?         IF RecurrenceCount ? policy threshold:
     ?           ? escalate: Goal ? AWAITING_HOST
     ?              Conversation Planner notifies Host
     ?              (Decision 0.3 pipeline: ConversativeIntent ? Constitutional Gate
     ?               ? Language Realization ? Notification Service ? Host)
     ?
     ??? NO:  treat as new gap normally
`

**Architectural rule:** The existence of recurrence protection is architecture. The threshold value is **runtime policy** ? never frozen in the Blueprint.
  → GapRecord for this GoalID?
     │
     ├── YES: recurrence detected
     │         Increment: GapRecord.RecurrenceCount
     │         IF RecurrenceCount < policy threshold:
     │           → permitted: re-dispatch to Acquisition
     │             (Acquisition idempotency: if already exists, SUCCESS immediately)
     │         IF RecurrenceCount ≥ policy threshold:
     │           → escalate: Goal → AWAITING_HOST
     │              Conversation Planner notifies Host
     │              (Decision 0.3 pipeline: ConversativeIntent → Constitutional Gate
     │               → Language Realization → Notification Service → Host)
     │
     └── NO:  treat as new gap normally
```


**Why this matters:** Without recurrence protection, IDUN could enter an infinite `E1 → gap → E2 → same gap → E3 → same gap → ...` loop, consuming resources indefinitely while the Host is unaware.

---

#### 0.4.O — Complete Canonical Lifecycle Diagram

```text
═══════════════════════════════════════════════════════════════
DECISION 0.4 — COMPLETE CANONICAL LIFECYCLE
E1 → ACQUISITION → E2
═══════════════════════════════════════════════════════════════

LEGEND
  ──────→  synchronous data flow
  - - - →  async / event
  [STORE]  persisted to core/storage or LTM
  [WM]     Working Memory (in-process)
  ⚠        boundary; must NOT be crossed
  ★        new contract / field

─────────────────────────────────────────────────────────
GOAL G1 (GoalID="g1")
  State: ACTIVE
  Owner: Goal Manager
  Persisted: core/storage
─────────────────────────────────────────────────────────

  ▼
EPISODE E1 (EpisodeID="e1", GoalID="g1", EpisodeType=INITIAL)
  Owner: Executive
  StartedAt: T0

  [PIPELINE: Understanding → Context → Reasoning → Planning]

  Reasoning detects KnowledgeGap
    Producer: Reasoning (intelligence/reasoning)
    Event:    GapSignal{GapID="gap1", GoalID="g1", Domain}
    Transport: Event Bus → Goal Manager (F1: GapRecord written BEFORE ack)

  Planning writes PlanCheckpoint
    Producer: Planning
    WRITE → E1 WM [WM]: PlanCheckpoint (EPISODE scope)

  WorkingMemoryManager.Checkpoint(episodeID=E1)
    WRITE → core/storage: E1_checkpoint [STORE]
    (atomic: GOAL + required EPISODE slots)

  Goal Manager: ACTIVE → PAUSED
    WRITE → core/storage: Goal.State=PAUSED, GapRecord{GapID, Status=DISPATCHED} [STORE]

  E1 CLOSES
    CompletedAt: T1
    Outcome: PARTIAL
    TerminationReason: GAP_PENDING ★
    Producer: Executive
    WRITE → EpisodeRecord E1 [STORE]

  Reflection(E1) — triggered async by Executive
    Reads: E1 episode trace (Outcome=PARTIAL, TerminationReason=GAP_PENDING)
    Emits: ReflectionReport for E1 [STORE]
    (Non-blocking; E1's ReflectionReport scopes to the segment before the gap)

─────────────────────────────────────────────────────────
KNOWLEDGE ACQUISITION (BOUNDED WORKER — NOT AN EPISODE)
─────────────────────────────────────────────────────────

  Owner: intelligence/knowledge/acquisition
  Triggered by: GapSignal via routeGap()

  Idempotency check (F3): exists(GapID="gap1") in Knowledge Store?
    YES → emit GapResolved(SUCCESS, StoreKey) immediately
    NO  → proceed

  Internal pipeline:
    classify → select strategy → acquire → normalize →
    evaluate confidence → classify durability

  Path A (Durable / VERIFIED or UNVERIFIED domain knowledge):
    WRITE → Knowledge Store [STORE] (insert-if-absent, F3)
    key: GapID="gap1"  value: KnowledgeRecord{...}
    ⚠ write failure → do NOT emit GapResolved(SUCCESS) → Level 2 timeout applies

  Path B (Transient / contextual: e.g., "current weather"):
    NO write to LTM (preserves LTM integrity)
    result carried in GapResolved.Payload (opaque)

  Emit: GapResolved{GapID="gap1", GoalID="g1", Status=SUCCESS, StoreKey="k1"?, Payload?}
    Producer: Knowledge Acquisition
    Consumer: Goal Manager (only)
    Transport: Event Bus (content-blind)

─────────────────────────────────────────────────────────
GOAL MANAGER (on GapResolved)
─────────────────────────────────────────────────────────

  Owner: intelligence/goalmanager

  (1) Check Goal.State == CANCELLED (F12)?
        YES → ignore GapResolved; no E2 scheduled; KnowledgeStore record remains
  (2) Check Goal.ExpiresAt (F4)?
        Expired → archive Goal; no E2
  (3) WRITE → GapRecord.Status = RESOLVED [STORE]
  (4) WRITE → GapRecord.ResolvedAt = now   [STORE]  ★ (required for restart recovery)
  (5) WRITE → Goal.State = ACTIVE           [STORE]
  (6) Emit ContinuationRequest ★
        {GoalID="g1", SourceEpisodeID="e1",
         GapResolutionHint{GapID="gap1", StoreKey="k1"?},
         TransientPayload? (if Path B)}
        Transport: IMPLEMENTATION SPECIFICATION
          (Event Bus or internal async mechanism — not frozen by Blueprint)

  ⚠ Goal Manager does NOT:
      call Executive.StartContinuation() directly
      wait synchronously for E2
      manage E2's internal pipeline

─────────────────────────────────────────────────────────
EXECUTIVE (receives ContinuationRequest)
─────────────────────────────────────────────────────────

  Owner: runtime/executive

  CREATE EPISODE E2
    EpisodeID:   "e2" (new — E1 never reused)
    GoalID:      "g1" (same Goal)
    EpisodeType: CONTINUATION ★
    StartedAt:   T2
    WRITE → EpisodeRecord E2 (opened) [STORE]

  E2 INITIALIZATION:
    (1) WMM.Restore(source="e1", target="e2")
          READ ← core/storage: E1_checkpoint
          WRITE → E2 WM [WM]: PlanCheckpoint
          WRITE → E2 WM [WM]: ActiveBeliefs [TENTATIVE]
          WRITE → E2 WM [WM]: UnresolvedQuestions (if checkpointed)
    (2) MemoryRetriever(GoalID="g1", EpisodeID="e2", hint={GapID="gap1", StoreKey="k1"?})
          READ ← Knowledge Store: general relevance query
          READ ← Knowledge Store: mandatory StoreKey="k1" fetch (if durable)
          WRITE → E2 WM [WM]: RetrievedMemories (EPISODE scope)
    (3) If TransientPayload present:
          WRITE → E2 WM [WM]: TemporaryResearch (EPISODE scope, NOT persisted)

  E2 PIPELINE ENTRY:
    ⊘ Understanding:     NOT executed (no new Host stimulus)
    ⊘ Context Resolution: NOT executed as fresh-input stage
    ✓ Reasoning:         ENTRY POINT (re-evaluates TENTATIVE beliefs + new knowledge)
    ✓ Planning:          reads PlanCheckpoint → resumes paused step
    ✓ Decision → Executive → Communication: normal pipeline

─────────────────────────────────────────────────────────
E2 NORMAL LIFECYCLE → Goal G1 continues
─────────────────────────────────────────────────────────
  If E2 completes normally:
    Outcome: SUCCESS | PARTIAL | FAILURE
    Reflection(E2) triggered async

  If E2 encounters another gap:
    → Same pattern: new E3 with new EpisodeID
    → Goal Manager tracks gap recurrence (§0.4.N)
    → If same GapID recurs beyond policy threshold → AWAITING_HOST
```

---

#### 0.4.P — Ownership Table

| Component | Owns | Must NOT Own |
|:----------|:-----|:-------------|
| Goal Manager | Gap lifecycle (GapRecord state machine), continuation signaling, restart recovery, recurrence tracking | Cognitive execution, Executive invocation, knowledge semantics |
| Knowledge Acquisition | Information acquisition, source evaluation, confidence scoring, Knowledge Store writes, GapResolved emission | Cognitive episode lifecycle, goal lifecycle, Reasoning over acquired knowledge |
| Executive | Episode creation/execution (E1 and E2), E2 initialization, MemoryRetriever invocation at E2 start | Goal lifecycle, GapRecord management |
| WorkingMemoryManager | Checkpoint persistence, cross-episode Restore, scope enforcement | Reasoning, acquisition decisions |
| MemoryRetriever | Relevance-based LTM retrieval, mandatory StoreKey fetch | Goal lifecycle, episode scheduling |
| Reasoning | Belief re-evaluation in E2 (TENTATIVE → CONFIRMED/CONTRADICTED), gap detection | Goal lifecycle, acquisition lifecycle |
| Planning | Plan resumption from PlanCheckpoint, next-step identification | Acquisition lifecycle |

---

#### 0.4.Q — Architecture vs Implementation vs Runtime Policy

| Conclusion | Classification | Reason |
|:-----------|:--------------|:-------|
| Acquisition is a Bounded Worker, not a cognitive Episode | ARCHITECTURE | Defines subsystem boundary and responsibility |
| Goal Manager signals; Executive acts; no direct invocation | ARCHITECTURE | Defines control-flow boundary and coupling rule |
| Executive must not poll Goal Manager | ARCHITECTURE | Defines coupling constraint |
| Delivery mechanism (Event Bus vs internal queue) | IMPLEMENTATION SPECIFICATION | Does not alter the ownership boundary |
| E2 enters at Reasoning; skips Understanding/Context | ARCHITECTURE | Defines canonical lifecycle variation |
| EpisodeType = CONTINUATION | ARCHITECTURE (metadata) | Required for Reflection and Episodic Memory |
| Gap recurrence protection exists | ARCHITECTURE | Prevents infinite acquisition loops |
| GapRecord.RecurrenceCount tracking | ARCHITECTURE | Required for recurrence detection |
| Recurrence threshold value | RUNTIME POLICY | Exact count is configuration/calibration |
| GapRecord.ResolvedAt timestamp | IMPLEMENTATION SPECIFICATION | Required for restart recovery ordering |
| MemoryRetriever retry at E2 init | ARCHITECTURE (rule) | Executive retries; escalates if exhausted |

---

#### 0.4.R — Updated Contract Summary

| Contract | Fields (architectural) | Status |
|:---------|:-----------------------|:-------|
| `GapResolved` | `GapID`, `GoalID`, `Status` (SUCCESS/FAILURE), `StoreKey?` (durable result reference), `Payload?` (opaque transient result) | Updated — extends existing contract |
| `ContinuationRequest` | `GoalID`, `SourceEpisodeID`, `GapResolutionHint?{GapID, StoreKey}`, `TransientPayload?` | NEW — not yet implemented |
| `GapResolutionHint` | `GapID`, `StoreKey` | NEW — not yet implemented |
| `EpisodeRecord.TerminationReason` | `GAP_PENDING` value | NEW field on EpisodeRecord |
| `EpisodeRecord.EpisodeType` | `CONTINUATION` value | NEW value on existing field |
| `GapRecord.ResolvedAt` | Timestamp of RESOLVED state transition | NEW field for restart recovery |
| `GapRecord.RecurrenceCount` | Count of times this GapID has recurred across Episodes under the same Goal | NEW field for recurrence protection |
| `WMM.Restore(src, target)` | Two-argument: sourceEpisodeID, targetEpisodeID | Updated — already applied to Decision 0.2 |

---

#### 0.4.S — Industry Pattern Complete Mapping

| Idun Architectural Choice | Established Pattern | What It Solves | IDUN Adaptation |
|:---|:---|:---|:---|
| New EpisodeID per continuation | Saga Step / Bounded Task | Long-running workflows cannot hold runtime resources during external waits | Goal = Saga parent; Episode = bounded step |
| EPISODE checkpoint + Restore(src, target) | State Rehydration / Snapshot | Workflow step state is persisted at step close and rehydrated at next step start | PlanCheckpoint persisted at E1 close; loaded into E2 EPISODE scope |
| GoalID + StartedAt correlation | Correlation ID | Avoids fragile linked lists; query-time sequence reconstruction | Replaces `ContinuationOf` pointer |
| GapResolutionHint(StoreKey) | Claim Check | Passes a lightweight reference through the orchestrator; retrieved at the destination | Goal Manager holds the key; MemoryRetriever redeems it |
| GapRecord.Payload (transient) | Saga State / Process Manager | Intermediate non-global step results held in orchestrator state, not global DB (LTM) | Transient knowledge stays out of LTM |
| LTM → MemoryRetriever → WM | Durable-State Rehydration | Pipeline reconstructs worldview from authoritative semantic database | GapResolutionHint adds mandatory-inclusion guarantee |
| Knowledge Acquisition as Bounded Worker | Saga Activity / Workflow Task | Specialized bounded work should not instantiate orchestrator machinery | Acquisition = Activity; Goal Manager = orchestrator; no Episode needed |
| Restart recovery re-emit | Idempotent Orchestration | Orchestrator re-emits pending signals after crash to prevent stuck states | Goal Manager re-emits ContinuationRequest for RESOLVED gaps without E2 |
| Gap recurrence detection | Circuit Breaker | Prevent infinite retry loops in a distributed system | Recurrence tracking per GapID per Goal; escalation to Host when threshold exceeded |

**Decision 0.4 freeze status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE. All five questions (Q0.4-1 through Q0.4-5) are fully resolved. Implementation specifications (delivery mechanism, serialization format, recurrence threshold, exact timestamp types) are intentionally deferred.

---

### Decision 0.5 — Planning Imports Reasoning: Anti-Corruption Boundary

**Finding:** The original v1 `CandidatePlan` embedded Reasoning's internal `SemanticGoal`. The v3 cognitive pipeline allowed Planning to directly ingest `ReasoningContext` and Decision to ingest `ReasoningContext`, creating tight package coupling across the intelligence pillar. In `reasoning/interfaces.go`, the dependency inverted (Reasoning importing Executive).

**Decision:** The inter-stage dependency graph is strictly mediated by the Executive acting as an Anti-Corruption Layer (Ports and Adapters architecture). The proposed `intelligence/contracts` package is explicitly **REJECTED** in favor of Executive Mapping. A shared `intelligence/types` package is used **only** for legacy v1 structural compatibility and genuinely shared primitives.

#### 0.5.A — Dependency Architecture

The final topology strictly prohibits cognitive stages from importing each other. The Executive is the sole orchestrator.

```text
Understanding
      ↓
Context Resolution
      ↓
Reasoning
      ↓
ReasoningContext (Owned by Reasoning)
      ↓
Executive (Orchestrator)
      │
      ├── maps to ──► PlanningInput (Owned by Planning) ──► Planning
      │
      └── maps to ──► DecisionInput (Owned by Decision) ──► Decision
```

- **Reasoning MUST NOT** import or call Executive.
- **Planning MUST NOT** import Reasoning.
- **Decision MUST NOT** import Reasoning.
- **Executive** performs boundary mapping and invokes the stages.

#### 0.5.B — Executive Mapping / Anti-Corruption Boundary

The Executive translates output schemas into input schemas. 

**ALLOWED for Executive:**
- Structural field extraction (e.g., pulling `ResolvedIntent` from `ReasoningContext`).
- Boundary translation (e.g., mapping Reasoning's `EnrichedSlot` to Planning's parameter requirements).
- Orchestration and invoking the next subsystem.

**NOT ALLOWED for Executive:**
- Cognitive reasoning or logic derivation.
- Planning or Decision-making.
- Semantic reinterpretation of data.
- Hidden business logic.

#### 0.5.C — Shared Type Boundary

The proposal to create `intelligence/contracts` is rejected because it duplicates existing packages and encourages a monolithic design.

- **`intelligence/types`**: Used strictly for genuinely shared domain primitives (e.g., `PresentationDirectives`) and legacy v1 compatibility types (e.g., `ResolvedGoal`).
- **`intelligence/interfaces`**: Used for pillar-wide behavioral interfaces that multiple subsystems must agree on, where no single subsystem can canonically host the interface.

#### 0.5.D — Forbidden Shared-Package Responsibilities

Types placed in `intelligence/types` or `intelligence/interfaces` **MUST NOT** contain:
- Cognitive logic, algorithms, or behavioral orchestration.
- Subsystem state (e.g., `ReasoningEpisodeID`).
- Subsystem-specific internal types. 
- Reasoning implementation state, Planning implementation state, or Decision implementation state.
- Hidden dependencies between cognitive subsystems.

**Critical Rule:** Indirect coupling is forbidden. A type in `intelligence/types` cannot embed a pointer to a type defined in `reasoning` or `planning`.

#### 0.5.E — Semantic Ownership

Producer ≠ Semantic Owner. The semantic definition of boundary types is owned by the Domain or the consuming Pipeline, not the producing subsystem.

| Type | Physical Location | Semantic Owner | Producer | Consumers | Forbidden Contents |
|:---|:---|:---|:---|:---|:---|
| **SemanticGoal** (v1) | `reasoning/types.go` | Reasoning (internal) | Reasoning | Reasoning | Do not export across boundaries |
| **ResolvedGoal** (v1) | `intelligence/types` | Goal Manager | Reasoning | Planning (v1) | Reasoning internals (`ReasoningTrace`) |
| **ReasoningContext** | `reasoning/v3` | Reasoning | Reasoning | Executive | Downstream subsystem types |
| **PlanningInput** | `planning/types.go` | Planning | Executive | Planning | Reasoning internal fields |
| **DecisionInput** | `decision/types.go` | Decision | Executive | Decision | Reasoning internal fields |
| **PresentationDirectives** | `intelligence/types` | Communication | Reasoning | Comms, Planning | Cognitive state or logic |
| **EnrichedSlot** | `reasoning/v3` | Reasoning | Reasoning | Executive | Must not be shared globally |

#### 0.5.F — Planning Boundary

The boundary data crossing into Planning is strictly limited to what Planning requires.

```text
Reasoning
    ↓ (ReasoningContext)
Executive
    ↓ (PlanningInput)
Planning
```

Fields crossing the boundary into `PlanningInput`:
1. **ResolvedIntent**: Source = Reasoning. Transformation = None. Consumer = Planning. Purpose = Core goal/capability selector.
2. **ArtifactID**: Source = Reasoning. Transformation = None. Consumer = Planning. Purpose = Lineage tracking.
3. **EnvelopeID**: Source = Reasoning. Transformation = None. Consumer = Planning. Purpose = Event traceability.
4. **EnrichedSlots**: Source = Reasoning. Transformation = Executive maps `reasoning.EnrichedSlot` to a local `planning` parameter format. Consumer = Planning. Purpose = Capability parameter binding.
5. **Metadata**: Source = Reasoning. Transformation = None. Consumer = Planning. Purpose = Interaction constraints.

#### 0.5.G — Decision Boundary

Similarly, Decision defines its own input.

```text
Planning
    ↓ (CandidatePlan)
Executive
    ↓ (DecisionInput)
Decision
```

`DecisionInput` is strictly owned by Decision. Decision and Planning do not share a generic input envelope.

#### 0.5.H — Physical Residency / Migration Map

| Type | Current Location | Final Location | Owner | Producer | Consumers | Reason |
|:---|:---|:---|:---|:---|:---|:---|
| **ReasoningContext** | `reasoning/v3` | `reasoning/v3` | Reasoning | Reasoning | Executive | Clean output envelope. |
| **PlanningInput** | *None* | `planning/types.go` | Planning | Executive | Planning | Consumer owns its input. |
| **DecisionInput** | *None* | `decision/types.go` | Decision | Executive | Decision | Consumer owns its input. |
| **SemanticGoal (v1)** | `reasoning/types.go`| `reasoning/types.go`| Reasoning | Reasoning | Reasoning | Becomes strictly internal. |
| **ResolvedGoal (v1)** | *None* | `intelligence/types`| Goal Manager| Reasoning | Planning (v1)| Fixes v1 `CandidatePlan` embedding. |
| **PresentationDirectives**| `reasoning/types.go`| `intelligence/types`| Communication| Reasoning | Planning, Decision| True cross-cutting primitive. |
| **EnrichedSlot** | `reasoning/v3` | `reasoning/v3` | Reasoning | Reasoning | Executive | Remains local; Executive maps it. |

#### 0.5.I — Legacy v1 Compatibility

The v1 structure physically embedded `reasoning.SemanticGoal` in `CandidatePlan`. Because Executive mapping cannot fix a struct-field embedding, this is resolved by stripping Reasoning-internals from the type and moving it to a shared domain primitive.

**Before:**
`CandidatePlan` → `reasoning.SemanticGoal`

**After:**
`CandidatePlan` → `intelligence/types.ResolvedGoal`

#### 0.5.J — Dependency Inversion

**Before:**
`intelligence/reasoning/interfaces.go` imported `intelligence/executive` (Dependency cycle: Executive orchestrates Reasoning, but Reasoning imports Executive).

**After:**
`intelligence/executive` imports `intelligence/reasoning/interfaces`. The required interfaces are defined in `reasoning/interfaces.go` (or `intelligence/interfaces`), and the Executive fulfills them. Reasoning never imports Executive.

#### 0.5.K — Detailed Data-Provenance Diagram

```text
SOURCE          Understanding (Perception Envelope)
  │
CONTRACT        understanding.SemanticInterpretation
  │
PRODUCER        Context Resolver → Reasoning
  │
CONTRACT        reasoning.ReasoningContext (Owned by Reasoning)
  │
EXECUTIVE       [Anti-Corruption Layer / Mapping]
  │               (Extracts Intent, Slots, Lineage. Drops internal belief state)
  │               (Write Authority: Executive creates new envelope)
  │
CONTRACT        planning.PlanningInput (Owned by Planning)
  │               (Forbidden Data: Reasoning internals must not cross)
  │
CONSUMER        Planning
```

#### 0.5.L — Boundary Violation Diagram

**Old Architecture (Violates Boundaries):**
```text
Planning ─────────► Reasoning (Import cycle / tight coupling)
Decision ─────────► Reasoning (Import cycle / tight coupling)
Executive ◄───────► Reasoning (Bidirectional dependency cycle)
```

**Corrected Architecture (Hexagonal / Ports & Adapters):**
```text
Reasoning
    │  (returns ReasoningContext)
    ▼
Executive
    │  (maps and invokes)
    ▼
Planning
```

#### 0.5.M — Failure / Recovery Considerations

- **Reasoning fails:** Executive handles failure according to the canonical lifecycle. Planning is never invoked.
- **Boundary mapping fails:** If Reasoning produces data that the Executive cannot map into a valid `PlanningInput`, the Executive raises a pipeline fault. Planning must not receive malformed or incomplete input.

#### 0.5.N — Industry Pattern Mapping

This architecture corresponds to established industry patterns:
- **Hexagonal Architecture (Ports and Adapters):** Reasoning and Planning are isolated domain hexagons. `ReasoningContext` and `PlanningInput` are ports. Executive acts as the adapter.
- **Mediator/Orchestrator:** Executive mediates between decoupled cognitive stages.
- **Anti-Corruption Layer (ACL):** Executive prevents Reasoning's internal schema (`SemanticGoal`, traces) from corrupting Planning's domain model.
- **Dependency Inversion / Information Hiding:** Planning depends only on its own inputs. Reasoning's internals are completely hidden.

#### 0.5.O — Implementation Wiring Map

| Connection | Producer | Contract | Consumer | Transformation | Ownership |
|:---|:---|:---|:---|:---|:---|
| Reasoning → Exec | Reasoning | `reasoning.ReasoningContext` | Executive | None (Direct Return) | Reasoning |
| Exec → Planning | Executive | `planning.PlanningInput` | Planning | Executive (ACL Mapping) | Planning |
| Exec → Decision | Executive | `decision.DecisionInput` | Decision | Executive (ACL Mapping) | Decision |
| Planner (v1) → Storage | Planning | `types.ResolvedGoal` | Decision/Reflection | Direct Embedding | Goal Manager |

**Decision 0.5 freeze status:** ARCHITECTURE RESOLVED / SAFE TO FREEZE. Executive Mapping is established as the pipeline data flow standard. Legacy v1 coupling is resolved via `intelligence/types`.

---

### Decision 0.6 — Episode vs. Turn (Complete Lifecycle Classification)

**Finding:** The Cognitive Operating Workflow conflates two distinct concepts under "episode."

**Definitions:**
- **Turn:** A single stimulus-to-response cycle. "What time is it?" → answer. One turn. Typically no episode record created.
- **Episode:** A bounded unit of cognitive work with a meaningful goal, plan, and outcome.
- **Goal:** A long-lived intention that may span many episodes and many days.

Not every turn becomes an episode. Not every episode serves a single turn.

#### 0.6.A — Executive Owns Lifecycle Classification

The pipeline initially processes the stimulus through:

```text
SystemEvent
    ↓
Understanding
    ↓
Context Resolution
    ↓
Reasoning
    ↓
ReasoningContext / ReasoningResult
    ↓
Executive
    ↓
Turn OR Episode
```

**Responsibilities:**
- **Understanding:** Understands the incoming stimulus.
- **Context Resolution:** Resolves the stimulus against available Working Memory/context.
- **Reasoning:** Determines the semantic/cognitive meaning of the interaction and produces the information required for lifecycle classification. Reasoning does not allocate EpisodeID, create EpisodeRecord, classify lifecycle state, create GoalRecord, or take ownership of Episode lifecycle.
- **Executive:** Owns orchestration and therefore owns Turn/Episode lifecycle classification, Episode promotion, EpisodeID, EpisodeRecord, routing between the Turn and Episode execution paths.
- **Goal Manager:** Continues to own GoalID, GoalRecord, Goal lifecycle, GapRecord, Gap lifecycle. This must remain consistent with Decision 0.1.

#### 0.6.B — GoalProposal Must NOT Equal Episode

A GoalProposal is evidence/input to Executive classification, not an automatic Episode trigger.

The architectural distinction is:
`Goal exists conceptually ≠ Active Episode is required`

A Goal may be long-lived and span multiple Episodes. For example, a request may establish a persistent Goal while the current interaction itself remains a Turn.

Therefore:
`GoalProposal + Episode-level execution requirements → Executive classification`

#### 0.6.C — Controlled Turn → Episode Promotion

Turn → Episode promotion is an explicit architectural transition.

```text
TURN
  │
  │ Episode-level requirement discovered
  ▼
PROMOTION BOUNDARY
  │
  ├── Executive allocates EpisodeID
  ├── Executive creates/persists EpisodeRecord
  ├── Episode lineage associated with EnvelopeID
  ├── EPISODE Working Memory scope established
  └── Goal Manager establishes required Goal/Gap lifecycle
              │
              ▼
           EPISODE
```

Promotion is controlled and structural. Promotion occurs when an established architectural requirement requires Episode lifecycle capabilities. Examples include:
- Gap lifecycle
- persistent bounded cognitive work
- multi-step execution
- planning/execution requiring Episode lifecycle
- another explicitly established Episode-level requirement

The exact representation of those triggers in code remains an implementation concern.

#### 0.6.D — GapSignal Is a Hard Promotion Trigger

```text
TURN
  ↓
Reasoning
  ↓
GapSignal
  ↓
Executive
  ↓
PROMOTION
  ↓
Episode + Goal lifecycle
  ↓
Goal Manager
  ↓
GapRecord
```

A Turn does not have an EpisodeID or GoalID. However, Decision 0.1 requires the normal Gap lifecycle to be associated with the appropriate Goal lifecycle. Therefore a `GapSignal` discovered during Turn processing is a hard architectural trigger for promotion. The `GapSignal` must not become an orphaned gap.

The sequence is:
Turn → Reasoning → GapSignal → Executive promotion → EpisodeID established → Goal Manager establishes/finds GoalID → GapRecord persisted → existing Decision 0.1 gap lifecycle.

Do not create a second Gap lifecycle. Reuse Decision 0.1.

#### 0.6.E — Goal/Episode Ownership Must Remain Separate

```text
Executive
├── EpisodeID
└── EpisodeRecord

Goal Manager
├── GoalID
├── GoalRecord
└── GapRecord
```

The Executive may request/delegate Goal establishment during promotion, but it must not create or own Goal records. Executive owns Episode lifecycle and Goal Manager owns Goal/Gap lifecycle. This is an important Decision 0.1 invariant.

#### 0.6.F — Turn Fast Path

A Turn does not automatically enter Planning and Decision. Planning and Decision are Episode-level execution stages, not mandatory stages for every conversational interaction.

```text
                         HOST
                           │
                           ▼
                    SystemEvent
                           │
                           ▼
                    Understanding
                           │
                           ▼
                  Context Resolution
                           │
                           ▼
                       Reasoning
                           │
                           ▼
                      Executive
                           │
                    CLASSIFICATION
                           │
              ┌────────────┴────────────┐
              │                         │
            TURN                    EPISODE
              │                         │
              │                         ▼
              │                      Planning
              │                         │
              │                      Decision
              │                         │
              └────────────┬────────────┘
                           ▼
                  Communication Path
                           │
                           ▼
                  Constitutional Gate
                           │
                           ▼
                          HOST
```

Do not create a second independent pipeline to support the Turn fast path. It is simply an Executive routing branch within the existing architecture.

#### 0.6.G — MemoryRetriever Boundary

Episode-level MemoryRetriever initialization occurs at Episode initialization (Decision 0.4). Therefore:
`TURN → ConversationTurns, ActiveTopic, existing WM context` does not automatically perform Episode-level LTM retrieval.

If the Turn discovers durable/external knowledge is required:
`TURN → KnowledgeGap → Promotion → EPISODE → Acquisition / continuation → MemoryRetriever → Reasoning`

#### 0.6.H — Working Memory During Promotion

Promotion does not copy the entire Turn Working Memory into Episode scope. Decision 0.2 already defines `ConversationTurns` as an independent sliding-window scope.

```text
TURN WM
 ├── ConversationTurns
 └── ActiveTopic

          │
          │ PROMOTION
          ▼

EPISODE WM
 ├── ActiveBeliefs
 ├── PlanCheckpoint
 └── Episode-specific state
```

Do not invent a migration/copy operation unless already required elsewhere. The existing Turn context remains available through its established Working Memory scope.

#### 0.6.I — Turn Identity and Lineage

Do not introduce TurnID.

```text
SystemEvent
    │
    └── EnvelopeID
            │
            ▼
           TURN
            │
            │ promotion
            ▼
        EpisodeID
            │
            ▼
         EPISODE
```

Before promotion: `EnvelopeID` provides Turn lineage.
After promotion: `EpisodeID` identifies bounded cognitive execution and remains associated with the originating `EnvelopeID`.

Maintain semantic distinction between:
- `EnvelopeID` — originating event/stimulus lineage
- `ArtifactID` — internal artifact identity
- `EpisodeID` — bounded cognitive execution identity
- `GoalID` — persistent Goal identity

Do not collapse these identities into one identifier. Do not state `EnvelopeID = ArtifactID` unless an existing contract explicitly requires that relationship.

#### 0.6.J — Restart and Replay Safety

**Scenario A — crash before Goal creation**
```text
TURN
  ↓
Promotion
  ↓
EpisodeRecord persisted
  ↓
CRASH
  ↓
Event not ACKed
  ↓
SystemEvent replay
  ↓
Same EnvelopeID
  ↓
Reprocess
  ↓
New EpisodeID
```
**Recovery:** The orphaned EpisodeRecord is harmless under the existing architecture because it has no side effects and is not associated with an active cognitive Goal/Episode lifecycle. The unpersisted EpisodeID must not become a source of inconsistent lineage.

**Scenario B — crash after Goal creation but before Gap persistence**
```text
TURN
  ↓
Promotion
  ↓
EpisodeRecord
  ↓
GoalRecord
  ↓
CRASH
  ↓
Event replay
  ↓
Same EnvelopeID
  ↓
Idempotent Goal handling
```
**Recovery:** Handled cleanly by existing Decision 0.1 restart/idempotency rules. No new transaction mechanism is invented.

**Scenario C — crash after GapRecord persistence but before ACK**
```text
TURN
  ↓
GapSignal
  ↓
Promotion
  ↓
GoalRecord
  ↓
GapRecord persisted
  ↓
CRASH before ACK
  ↓
┌──────────────────────┐
│ SystemEvent replay   │
│ Goal Manager recovery│
└──────────┬───────────┘
           ↓
      Gap re-dispatch
           ↓
    Acquisition F3
    insert-if-absent
           ↓
      safe recovery
```
**Recovery:** Explicitly connected to the existing Decision 0.1 Saga/idempotency architecture. Does not introduce two-phase commit, distributed transactions, rollback architecture, or any new recovery subsystem.

If `EnvelopeID` is used as an implementation-level uniqueness/idempotency key for Goal creation, explicitly classify that as Implementation Specification, not as a new architecture.

#### 0.6.K — Cross-Decision Traceability

Decision 0.6 defines when an interaction remains a Turn and when it enters the Episode lifecycle. It is an architectural extension/reuse of the existing model:

- **Decision 0.1:** Provides Executive orchestration authority, Goal Manager lifecycle ownership, Gap lifecycle, restart/recovery/idempotency.
- **Decision 0.2:** Provides TURN Working Memory scope, EPISODE Working Memory scope, ConversationTurns, scope isolation.
- **Decision 0.3:** Provides Conversation Planner, communication path, Constitutional Gate.
- **Decision 0.4:** Provides Episode identity, Episode lifecycle, continuation Episodes, Episode-level MemoryRetriever, E1 → E2 model.
- **Decision 0.5:** Provides Executive Mapping, ReasoningContext, EnvelopeID lineage, boundary separation between cognitive subsystems.
- **North Star:** Provides Simplicity, Loose coupling, Restart safety. Decision 0.6 preserves these by explicitly rejecting TurnID, Turn Manager, Classification subsystems, and mandatory Planning/Decision for trivial Turns. It strictly reuses existing Goal/Gap lifecycle and Saga/idempotency recovery mechanisms while ensuring Executive remains the single orchestration boundary.

#### 0.6.L — Architecture vs Implementation vs Runtime Policy

| Classification | Examples |
|:---|:---|
| **Architecture** | Executive owns lifecycle classification. Turn → Episode promotion exists. GapSignal is a hard promotion trigger. Goal Manager owns Goal/Gap lifecycle. Executive owns Episode lifecycle. Turns can use the communication fast path. Planning/Decision are Episode-level stages. EnvelopeID provides pre-Episode lineage. No TurnID. |
| **Implementation Specification** | Exact trigger field names. Boolean versus enum representation. Episode database schema. Goal database uniqueness constraints. ID-generation implementation. Event bus/internal queue details. API method signatures. Serialization format. Exact persistence transaction mechanics. |
| **Runtime Policy** | Tunable thresholds. Timeouts. Retention/eviction values. Operational resource limits. |

#### 0.6.M — Implementation Wiring Map

Input → Understanding → Context Resolution → Reasoning → ReasoningContext → Executive Classification

| Boundary | Producer | Consumer | Contract | Identity | Persistence Requirement | Ownership | Failure Behavior | Source Decision |
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
| **TURN Fast Path** | Executive | Conversation Planner | ConversativeIntent | EnvelopeID | None (Volatile Turn) | Executive routes, CP processes | Replay SystemEvent | Decision 0.3, Decision 0.6 |
| **EPISODE Path** | Executive | Planning, Decision, Goal Manager | PlanningInput, DecisionInput | EpisodeID | EpisodeRecord | Executive owns Episode | Recovery via Decision 0.1 | Decision 0.4, Decision 0.6 |
| **Goal Req (Promo)** | Executive | Goal Manager | GapSignal, GoalProposal | EnvelopeID / GoalID | GoalRecord, GapRecord | Goal Manager owns Goal | Idempotent insertion | Decision 0.1, Decision 0.6 |

#### 0.6.N — Forbidden Interpretations

- Reasoning creates Episodes → **FORBIDDEN**
- GoalProposal automatically creates Episode → **FORBIDDEN**
- Executive creates GoalRecord → **FORBIDDEN**
- Goal Manager owns EpisodeID → **FORBIDDEN**
- Every Turn creates Episode → **FORBIDDEN**
- Every Turn requires TurnID → **FORBIDDEN**
- Every Turn enters Planning → **FORBIDDEN**
- Every Turn enters Decision → **FORBIDDEN**
- Turn bypasses Constitutional Gate → **FORBIDDEN**
- Turn automatically initializes LTM → **FORBIDDEN**
- New Turn Manager subsystem → **FORBIDDEN**
- New Classification subsystem → **FORBIDDEN**
- Second independent cognitive pipeline → **FORBIDDEN**

---

### Decision 0.7 — Two Distinct Write Paths: Acquisition vs. Learning

**Finding:** The existing Learning subsystem (`2.0.0-FROZEN`) already has a sophisticated rollout model: `Draft → Validated → Shadow → Canary → Active`. Learning produces `CandidateSnapshot` objects; a Rollout Executor (owned by the runtime/Executive layer) activates them.

**Architectural Correction:** 
The previous version of this decision broadly stated: *"New knowledge and skill updates must flow through the same CandidateSnapshot lifecycle."* This was too broad. It inadvertently routed factual knowledge acquisition through the `Shadow → Canary` rollout, which would block the Decision 0.1/0.4 `Continuation Episode` model, freezing the system during simple factual queries.

**Decision:** 
Learning-produced cognitive updates flow through the `CandidateSnapshot` lifecycle. Knowledge and skills acquired by the Acquisition subsystems follow their established P8/P9 validation and direct-write paths. Acquisition and Learning are separate state-change authorities. The `CandidateSnapshot` rollout lifecycle governs Learning-produced cognitive updates and must not be interpreted as a prerequisite for ordinary `KnowledgeGap`/`SkillGap` acquisition.

#### A. Acquisition → Knowledge Store Path
This path allows the system to fetch information from the world and use it immediately to resolve a pending Goal.

*   **Origin:** External source / acquisition target
*   **Trigger:** `KnowledgeGap`
*   **Producer:** Knowledge Acquisition
*   **Validation:** P8 (evaluates confidence, attaches provenance, establishes verification status)
*   **Resulting Artifact:** Verified Knowledge Record
*   **Destination:** Knowledge Store
*   **Write Authority:** Knowledge Acquisition (direct write)
*   **GapResolved:** `GapResolved{GapID, Status=SUCCESS, StoreKey}` is emitted immediately after the write.
*   **Consumer:** Goal Manager (ContinuationRequest)
*   **Continuation:** Episode E2 resumes the Goal using the newly acquired knowledge.

#### B. Acquisition → Skill Registry Path
This path allows the system to fetch discrete capabilities and make them available for planning.

*   **Origin:** External registry / tool library
*   **Trigger:** `SkillGap`
*   **Producer:** Skill Acquisition
*   **Validation:** P9 (validation, security check, sandbox execution)
*   **Security Boundary:** Sandbox
*   **Resulting Artifact:** Available Skill (`SkillCard`)
*   **Destination:** Skill Registry
*   **Write Authority:** Skill Acquisition (direct write)
*   **GapResolved:** `GapResolved` is emitted immediately after the write.
*   **Consumer:** Goal Manager / Planning (for continuation)

#### C. Learning → CandidateSnapshot Path
This path handles internal, systemic improvements to cognitive behavior and strategies.

*   **Origin:** Completed cognitive work
*   **Trigger:** Periodic learning / Reflection process
*   **Evaluator:** Reflection (evaluates cognitive performance)
*   **Evaluation Artifact:** `ReflectionReport`
*   **Producer:** Learning (synthesizes the improvement)
*   **Resulting Artifact:** `CandidateSnapshot`
*   **Lifecycle Manager:** Rollout Executor
*   **Rollout Lifecycle:** `Draft → Validated → Shadow → Canary → Active`
*   **Write Authority:** Learning (via Rollout Executor, strictly after `Active` promotion)

*(Note: The internal classification of `CandidateSnapshot` payload types—e.g., Models, Policies, Grammars—remains an implementation/architecture question only if later decisions require it. Q0.7-1 does not need to resolve individual Learning payload categories).*

#### D. Acquisition vs Learning Comparison

| Property | Acquisition Path | Learning Path |
|:---|:---|:---|
| **Origin** | Knowledge/Skill Acquisition | Learning |
| **Trigger** | `KnowledgeGap` / `SkillGap` | Reflection / learning process |
| **Primary purpose** | Acquire missing external capability/information | Improve cognitive behavior/strategy |
| **Validation** | P8 / P9 | `CandidateSnapshot` lifecycle |
| **CandidateSnapshot** | No | Yes |
| **Direct store write** | Yes, where authorized | No |
| **Rollout Executor** | No | Yes |
| **GapResolved** | Yes | Not the acquisition mechanism |
| **Continuation Episode**| Yes when resolving a Goal gap | Not inherently the same lifecycle |
| **Active-state mutation**| Validated acquisition path | Must pass rollout lifecycle |
| **Authority** | Acquisition subsystem | Learning + Rollout architecture |

#### E. Traceability Diagram

```text
                         STATE ENTERS IDUN
                                │
                  ┌─────────────┴─────────────┐
                  │                           │
             EXTERNAL GAP                INTERNAL LEARNING
                  │                           │
          ┌───────┴───────┐                   │
          │               │                   ▼
     KnowledgeGap      SkillGap           Reflection
          │               │                   │
          ▼               ▼                   ▼
 Knowledge Acquisition  Skill Acquisition  ReflectionReport
          │               │                   │
          │ P8            │ P9               ▼
          │               │              Learning
          ▼               ▼                   │
 Knowledge Store     Skill Registry           ▼
          │               │            CandidateSnapshot
          │               │                   │
          └───────┬───────┘                   ▼
                  │                     Rollout Executor
             GapResolved                      │
                  │                       Draft → Validated
                  ▼                       → Shadow → Canary
            Goal Manager                         │
                  │                              ▼
                  ▼                           Active
          Continuation E2
                  │
                  ▼
             Resume Goal
```

#### F. Ownership and Authority Mapping

```text
Acquisition
├── Knowledge Acquisition
│      └── Knowledge Store write authority
│
└── Skill Acquisition
       └── Skill Registry write authority

Learning
└── CandidateSnapshot production
       │
       └── NO direct active-state mutation

Rollout Executor
└── CandidateSnapshot activation lifecycle

Goal Manager
├── Goal lifecycle
├── Gap lifecycle
└── GapResolved / continuation coordination
```

#### G. The Shadow Sandbox (Resolved Execution Boundary)

**Finding:** Candidate execution is cognitive work and therefore must reuse the Episode lifecycle from Decision 0.6. However, executing a `CandidateSnapshot` within the Active `Goal Scope` would violate Decision 0.2 by corrupting authoritative state. Relying solely on `Episodic Scope` fails because Shadow child episodes (spawned via `GapSignal`) would lose access to the parent's context.

**Architectural Correction:** We establish the **Shadow Sandbox Rule**. Shadow is an Episode execution mode, not a second pipeline. It requires a dedicated `Shadow Scope` that spans the entire Shadow Episode tree, acting as a read-through cache over Active Working Memory while trapping all writes and side effects.

##### 1. Origin of the Candidate

*   **Reflection → Learning → CandidateSnapshot**
*   **Why Learning creates it:** Learning synthesizes improvements based on `ReflectionReport` evidence.
*   **What it represents:** A packaged structural or behavioral update to IDUN's cognitive strategies.
*   **Why it cannot directly modify active state:** Principle P6 mandates that Learning does not write directly. This prevents hallucinated or regressive logic from immediately corrupting IDUN.
*   **Persistence & Lifecycle:** The `CandidateSnapshot` is registered in the `SnapshotRegistry` and enters the controlled rollout lifecycle (`Draft → Validated → Shadow → Canary → Active`). 
*   **Acquisition Path Separation:** This remains strictly separate from the `KnowledgeGap`/`SkillGap` → Acquisition → Direct Write lifecycle (as established in Section A/B). 

##### 2. CandidateSnapshot Lifecycle

| State | Meaning | Owner | Allowed | Prohibited | Next | Authoritative? |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Draft** | Created, unvalidated | Learning | Internal synthesis | Execution | Validated | No |
| **Validated** | Passed structural checks | Validation Pipeline | Ready for rollout | Execution | Shadow | No |
| **Shadow** | Ephemeral parallel evaluation | Rollout Executor | Shadow execution | Real side effects | Canary | No |
| **Canary** | Bounded real-world trial | Rollout Executor | Real execution (limited) | Global activation | Active | Yes (bounded) |
| **Active** | System-wide baseline | Rollout Executor | Normal operations | N/A | Retired | **YES** |

##### 3. Shadow is an Episode Mode

Shadow execution is NOT a second cognitive pipeline (`ShadowRunner`, `ShadowBrain`, etc.). 
It is exactly: **Episode + (ExecutionMode = SHADOW)**
This strictly preserves Decision 0.6 (All cognitive work belongs in an Episode).

##### 4. Shadow Episode Identity

*   **EpisodeID:** Identifies the unique bounded cognitive execution. (Owned by Executive)
*   **SnapshotID:** Identifies the exact Candidate producing the behavior. (Owned by Learning/Rollout)
*   **EnvelopeID:** Identifies the originating stimulus. (Owned by Event Router)
*   **GoalID:** Identifies the active objective being solved. (Owned by Goal Manager)
*   **Lineage Binding:** The `EpisodeRecord` (specifically `ReplayMetadata` or an extended field) MUST carry the `SnapshotID` when executing in Shadow mode, ensuring all resulting `ReflectionReports` are cryptographically bound to the Candidate being evaluated.

##### 5. The Shadow Scope

The **Shadow Scope** is an ephemeral evaluation sandbox spanning the entire Shadow Episode tree. It is NOT a fourth authoritative Working Memory scope, nor a parallel cognitive state.

```text
              AUTHORITATIVE ACTIVE WM
             ┌────────────────────────┐
             │ Global Scope           │
             │ Goal Scope             │
             └───────────┬────────────┘
                         │
                    read-through
                         │
                         ▼
               ┌──────────────────────┐
               │     SHADOW SCOPE     │
               │   ephemeral sandbox  │
               └──────────┬───────────┘
                          │
             ┌────────────┼────────────┐
             ▼            ▼            ▼
        Shadow E1    Shadow E2    Shadow E3
             │            │            │
             └────────────┴────────────┘
                          │
                          ▼
                       DISCARD
```

**Why Shadow Scope is required:** If a Shadow Parent Episode emits a `GapSignal` and spawns a Shadow Child, the Child must be able to read the Parent's temporary context. An isolated `Episodic Scope` would prevent this. The `Shadow Scope` acts as a unified transparent layer over the Active state that persists only for the duration of the Candidate evaluation tree.

##### 6. Read and Write Semantics

**Read Precedence:**
Shadow reads strictly fall back: `Shadow-local state → Active Goal Scope → Active Global Scope`. Candidate and Active must begin from equivalent relevant cognitive context for evaluation to have meaning.

**Write Semantics (The Shadow Firewall Rule):**
*A Shadow Episode and every descendant Shadow Episode are permanently prohibited from committing state changes or external side effects to authoritative IDUN state or the real world. All mutable state generated by Shadow execution remains within the bounded Shadow Scope and is discarded when the Shadow execution terminates.*

```text
                 SHADOW EXECUTION
                       │
          ┌────────────┼─────────────┐
          │            │             │
          ▼            ▼             ▼
       WM Write     Store Write   External Side Effect
          │            │             │
          └────────────┼─────────────┘
                       ▼
                 SHADOW FIREWALL
                       │
                       ▼
                 LOCAL / SIMULATED
                       │
                       ▼
                  EVALUATION
```
*(Explicitly: Shadow → Active promotion of temporary state is permanently forbidden).*

##### 7. Evaluation Fidelity

**Fidelity Rule:** Shadow operations must be isolated without unnecessarily changing the Candidate's cognitive semantics. Shadow side effects are simulated only when doing so preserves the semantics required to evaluate the Candidate. When simulation cannot preserve those semantics, the Shadow environment produces a controlled outcome representing the relevant constraint or failure.

| Operation | Active | Shadow | Authoritative mutation? | Evidence? | Reason |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Global read | allowed | read-through | No | optional | Same context |
| Goal read | allowed | read-through | No | optional | Same context |
| Shadow-local write| allowed | local only | No | yes | Preserve semantics |
| Goal mutation | allowed | trapped/local | No | yes | Protect Active Goal |
| Knowledge write | allowed | trapped/local | No | yes | Prevent contamination |
| Skill write | allowed | trapped/local | No | yes | Prevent contamination |
| GapSignal | allowed | Shadow-only | No | yes | Preserve cognitive flow |
| GapResolved | allowed | Shadow-only | No | yes | Must not close real Goal |
| Host comms | real | no real send | No | yes | Prevent external side effect |
| External mutation | real | controlled outcome | No | yes | Safety + fidelity |
| Invalid API request| real error | controlled error | No | yes | Preserve failure semantics |

##### 8. GapSignal and Shadow Child Episodes

When a Shadow Episode emits a `GapSignal`:
1. The Shadow Firewall traps it.
2. The real Goal Manager ignores it.
3. A Shadow Child Episode is spawned within the same `Shadow Scope`.
4. When the Child completes, the `GapResolved` returns to the Shadow evaluation tree.
*(Runtime policy dictates recursion limits to prevent runaway shadow execution).*

##### 9. Knowledge and Skill Acquisition During Shadow

*   **Knowledge Acquisition:** Validated via P8, but the result is strictly stored in the Shadow Scope as evaluation evidence. The authoritative Knowledge Store remains untouched.
*   **Skill Acquisition:** Validated via P9, but the result is stored locally. The authoritative Skill Registry remains untouched.

##### 10. Host / External Effects

*   **Communication Planner:** Shadow candidates cannot send real messages to the Host. The intent is trapped as evidence.
*   **External APIs:** Shadow candidates cannot execute real-world mutations (payments, emails, destructive device actions). The request is structurally validated; if valid, a simulated controlled outcome is returned. If invalid, a simulated error forces the Candidate to handle the failure.

##### 11. Reflection Integration & Shadow Termination

**Evidence Flow:**
`Shadow Episode → Cognitive Work → Reflection → ReflectionReport → Candidate Evaluation Evidence`.

**Termination:**
When the Shadow tree completes, the `ReflectionReport` is generated. The entire `Shadow Scope` (including hallucinated knowledge, fake goal states, and temporary memory) is instantly destroyed. **Shadow state is never automatically promoted to Active Working Memory.**

##### 12. Ownership Matrix

| Responsibility | Owner | Input | Output | Authoritative? |
| :--- | :--- | :--- | :--- | :--- |
| Candidate creation | Learning | ReflectionReport | CandidateSnapshot | No |
| Candidate lifecycle | Rollout Executor | CandidateSnapshot | Lifecycle state | Yes for lifecycle |
| Shadow execution | Executive/runtime | Candidate + Episode | Shadow execution | No |
| Shadow isolation | Runtime boundary | EpisodeMode | isolated execution | No |
| Shadow memory | Shadow Scope | Active context + writes | temporary state | No |
| Cognitive evaluation | Reflection | Shadow cognitive work | ReflectionReport | No |
| Promotion governance | Promotion Authority | evaluation evidence | RolloutRecommendation | Yes for promotion |
| Lifecycle mutation | Rollout Executor | recommendation | new lifecycle state | Yes for lifecycle |

##### 13. Identity / Lineage Diagram

```text
Original stimulus
       │
       ▼
EnvelopeID
       │
       ├────────────── Active execution (GoalID)
       │
       └────────────── Shadow execution
                          │
                          ▼
                    Shadow Episode
                          │
                    SnapshotID
                          │
                          ▼
                    ReflectionReport
```

##### 14. Failure / Recovery Scenarios

*   **Scenario A (Shadow Crash):** Shadow Episode crashes → Shadow Scope discarded → Active state unaffected.
*   **Scenario B (Forbidden Write):** Candidate attempts real-world write → Shadow Firewall intercepts → Simulated controlled outcome returned → Recorded as evidence.
*   **Scenario C (Shadow Child Crash):** Shadow Child crashes → Parent remains in Sandbox → Real Goal unaffected.

##### 15. Cross-Decision Traceability

*   **Decision 0.1:** Goal/Gap lifecycle remains authoritative only for Active.
*   **Decision 0.2:** Active Working Memory remains the single authoritative source of truth. Shadow is an ephemeral sandbox, not a fourth authoritative scope.
*   **Decision 0.3:** Shadow communication cannot reach the real Host.
*   **Decision 0.4 & 0.6:** Shadow cognitive execution remains Episode-bounded.
*   **Decision 0.5:** Executive/runtime remains responsible for orchestration; there is no second cognitive pipeline.
*   **North Star:** Ensures offline-first safety, simplicity (reuses Episodes), and loose coupling (Reflection grades the candidate independently).

##### 16. Architecture vs Implementation vs Runtime Policy

*   **Architecture:** Shadow is an Episode execution mode. Shadow Scope exists as an ephemeral evaluation sandbox spanning the Shadow Episode tree. Authoritative state is read-through. Shadow writes cannot become authoritative. Real-world side effects are prohibited. Shadow must preserve evaluation semantics. Shadow evidence flows into Reflection.
*   **Implementation (Deferred):** Exact `EpisodeMode` struct fields, `ShadowScope` Go interfaces, context propagation mechanics, interceptor/decorator code, candidate routing mechanisms, serialization.
*   **Runtime Policy (Deferred):** Shadow execution timeouts, child Episode depth limits, evaluation sample sizes, promotion thresholds.

##### 17. Implementation Traceability Map

| Boundary | Producer | Consumer | Contract | Identity | Persistence | Failure behavior | Source Decision | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Candidate Creation | Learning | SnapshotRegistry | `CandidateSnapshot` | SnapshotID | Immutable store | Discard | Dec 0.7 | Implemented |
| Shadow Memory | Runtime | Shadow Episode | `ShadowScope` | EpisodeID | Ephemeral | Dropped | Dec 0.2/0.7 | Gap |
| Cognitive Eval | Shadow Episode | Reflection | `ReflectionReport` | SnapshotID | Immutable | Dropped | Dec 0.4/0.7 | Gap |

##### 18. Forbidden Interpretations (Guardrails)

Future implementations **MUST NOT** interpret this architecture to mean:
*   Shadow is a second cognitive pipeline.
*   Shadow is a fourth authoritative WM scope.
*   Shadow can modify Active Goal or Global state.
*   Shadow can write to the authoritative `KnowledgeStore` or `SkillRegistry`.
*   Shadow can send real Host communication or execute real external mutations.
*   Shadow state automatically becomes Active state.
*   Shadow evidence automatically equals promotion authorization.
*   Learning evaluates its own Candidate.
*   Every side effect should return simulated Success (malformed inputs must fail).
*   `Episodic Scope` alone is sufficient for an entire Shadow tree (it is not).

#### H. Reflection and Learning Feedback Architecture

**Architectural Principle:** Reflection is an observation/evaluation capability operating across multiple execution contexts. Learning consumes evidence from Reflection and synthesizes scoped improvements. Reflection does not directly modify active cognitive state, Learning does not directly promote its own changes, Governance remains responsible for Candidate promotion decisions, and Rollout Executor remains responsible for lifecycle mutation.

##### 1. Three Reflection Contexts

Reflection operates at three distinct architectural observation levels.

**A. Episode Reflection**
```text
Episode
   │
   ▼
Reflection
   │
   ▼
ReflectionReport
```
Evaluates a specific piece of cognitive work to observe how the Episode performed. Produces structured evidence (e.g., `WentWell`, `WentPoorly`, `CrossCognitiveFindings`). It does not mutate the Episode, Goal, Working Memory, or active cognitive state merely by reflecting. This is the existing P7-style reflection responsibility.

**B. Periodic / System Reflection**
```text
Recent completed Episodes
        │
        ▼
Periodic Reflection
        │
        ▼
Aggregated evidence
        │
        ▼
Recurring findings / weaknesses
        │
        ▼
Learning input
```
Answers: *"How is IDUN performing over time?"*
Aggregates historical cognitive evidence to identify repeated failures, recurring weaknesses, recurring clarification needs, repeated reasoning problems, persistent capability weaknesses, and observed regressions where existing evidence supports them.

**C. Candidate Reflection**
```text
CandidateSnapshot
       │
       ▼
Shadow / Canary Episode
       │
       ▼
Reflection
       │
       ▼
Candidate evaluation evidence
```
Candidate cognitive work is still an Episode. Reflection evaluates Candidate behavior exactly as it evaluates Active cognitive work, producing Candidate evaluation evidence. **Crucially: Candidate Reflection does not itself authorize promotion, nor does it directly mutate CandidateSnapshot lifecycle state.**

*(Note: The existence of Candidate Reflection does not by itself define the complete Candidate-vs-Active governance mechanism. Promotion Authority handles baseline identity, quantitative comparison, and promotion authorization).*

##### 2. Reflection Context Comparison

| Reflection Context | Input | Primary Question | Output | Downstream Consumer | State Mutation Authority |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Episode** | Completed cognitive Episode | How did this Episode perform? | `ReflectionReport` | Existing consumers / Learning | None |
| **Periodic/System** | Aggregated historical cognitive evidence | How is IDUN performing over time? | Structured findings/evidence | Learning | None |
| **Candidate** | Shadow/Canary cognitive Episode | How did this candidate behave? | Candidate evaluation evidence | Future Governance | None |

##### 3. The Reflection → Learning Relationship

```text
                    REFLECTION
                        │
          ┌─────────────┼─────────────┐
          │             │             │
          ▼             ▼             ▼
       Episode      Periodic      Candidate
       findings      findings       evidence
          │             │             │
          └───────┬─────┘             │
                  ▼                   │
               Learning               │
                  │                   │
                  ▼                   │
          CandidateSnapshot           │
                  │                   │
                  ▼                   │
             Shadow/Canary            │
                  └──────────┬────────┘
                             ▼
                         Governance [OPEN Q0.7-2]
                             │
                             ▼
                      Rollout Executor [OPEN Q0.7-3]
```
*(Note: Candidate evidence → Governance → Rollout remains tied to the unresolved Q0.7-2 architecture. This diagram demonstrates the conceptual closed loop).*

##### 4. Learning's Scope

**Architectural Rule:** Learning responds to evidence-backed weaknesses identified through Reflection and produces scoped CandidateSnapshots rather than indiscriminately modifying all cognitive capabilities.

The scope of a Learning-produced `CandidateSnapshot` should be the smallest scope that adequately addresses the evidence-backed weakness.

```text
Reflection
    │
    ▼
Observed weakness
    │
    ├───────────────┐
    │               │
    ▼               ▼
Local weakness   Systemic weakness
    │               │
    ▼               ▼
Targeted        Coordinated
Candidate       Candidate
    │               │
    └───────┬───────┘
            ▼
      CandidateSnapshot
```
*   **Local Example:** Planning repeatedly produces poor business plans → Learning targets Planning.
*   **Systemic Example:** Reasoning + Planning + Communication all fail because of a shared context-grounding weakness → A coordinated CandidateSnapshot may be appropriate.

##### 5. Responsibility Matrix

| Responsibility | Owner | Must NOT Own |
| :--- | :--- | :--- |
| Observe/evaluate cognitive work | Reflection | Candidate promotion |
| Identify recurring weaknesses | Reflection / System Reflection | Direct system mutation |
| Synthesize improvement | Learning | Active-state promotion |
| Represent proposed improvement | `CandidateSnapshot` | Automatic activation |
| Evaluate Candidate against baseline | Promotion Authority | Learning |
| Authorize promotion | Promotion Authority | Reflection |
| Mutate rollout lifecycle | Rollout Executor | Cognitive evaluation |
| Persist Active artifact | Existing store/write authority | Learning directly |

##### 6. The Anti-Self-Evaluation Rule

**Architectural Guardrail:** Learning must not be the final authority on whether its own `CandidateSnapshot` is successful. A subsystem that generates the change should not be the sole authority determining that the change succeeded.

```text
Learning
   │
   └── produces CandidateSnapshot
            │
            ▼
       Candidate execution
            │
            ▼
        Reflection
            │
            ▼
      independent evidence
            │
            ▼
        Promotion Authority
            │
            ▼
       promotion decision
```

##### 7. The Closed Learning Loop

Reflection is not exclusively a Candidate evaluator. It observes both normal IDUN operation (producing System Evidence for Learning) and Shadow/Canary Candidate operation (producing Candidate Evidence for Governance).

```text
                  IDUN RUNS
                      │
                      ▼
                  Reflection
                      │
          ┌───────────┴───────────┐
          │                       │
          ▼                       ▼
   System Findings        Candidate Evidence
          │                       │
          ▼                       │
       Learning                   │
          │                       │
          ▼                       │
   CandidateSnapshot              │
          │                       │
          ▼                       │
     Shadow/Canary                │
          │                       │
          └──────────┬────────────┘
                     ▼
             Promotion Authority
                     │
                     ▼
              Rollout Executor
                     │
                     ▼
               Active IDUN
                     │
                     └──────────→ IDUN runs again
```
*Architectural Rationale:* Reflection evidence is diagnostic. It identifies where performance is weak. Learning uses that evidence to synthesize a scoped improvement. A `CandidateSnapshot` is not a command to rewrite all active cognitive state.

##### 8. Cross-Decision Traceability

*   **Decision 0.2:** Working Memory and memory retrieval boundaries.
*   **Decision 0.3:** Communication behavior remains owned by Communication pipeline.
*   **Decision 0.4:** Candidate and Shadow cognitive execution still use Episode semantics.
*   **Decision 0.5:** Executive remains orchestrator and cognitive stage boundaries remain intact.
*   **Decision 0.6:** Shadow work is Episode-based cognitive work, not a second pipeline.
*   **Decision 0.7:** CandidateSnapshot and rollout architecture.
*   **P6:** Learning produces CandidateSnapshots and does not directly mutate active cognitive state.
*   **P7:** Reflection is read-only evaluation.
*   **P8/P9:** Acquisition validation remains separate from Learning-generated CandidateSnapshots.

##### 9. Architecture vs Implementation vs Runtime Policy

| Category | Description |
| :--- | :--- |
| **Architecture** | Reflection evaluates multiple execution contexts. Reflection can identify recurring system weaknesses. Learning consumes evidence and produces scoped `CandidateSnapshots`. Learning does not directly activate its own changes. Candidate evaluation remains separate from promotion authority. Promotion Authority is the mechanical promotion authority. Rollout Executor remains the lifecycle mutator. |
| **Implementation (Deferred)** | Exact Reflection aggregation code, exact candidate/evaluation data structures, exact linkage fields, exact Learning interfaces, exact event wiring, exact storage schema. |
| **Runtime Policy (Deferred)** | Periodic Reflection cadence, observation windows, sample sizes, thresholds, how many failures trigger a learning cycle, retention periods. |

##### 10. Forbidden Interpretations (Guardrails)

Future implementations **MUST NOT** interpret this architecture to mean:
*   Reflection only evaluates Candidates.
*   Reflection automatically promotes Candidates.
*   Reflection directly mutates `CandidateSnapshot` state.
*   Learning evaluates its own Candidate and declares success.
*   Learning retrains all of IDUN after any weakness.
*   One weakness necessarily means one module.
*   `CandidateSnapshot` necessarily affects the entire system.
*   Candidate evidence automatically equals promotion authorization.
*   Existing `ReflectionReport` is automatically sufficient for Governance.
*   Governance is already implemented.
*   Rollout Executor performs cognitive evaluation.
*   Periodic Reflection has a fixed runtime interval defined by architecture.

#### I. Candidate Routing (Resolved Execution Boundary)

**Architectural Purpose:** What problem does Candidate Routing solve?
Candidate Routing provides the controlled execution path that allows a `CandidateSnapshot` to be evaluated against the currently Active behavior using the same semantic stimulus, without making the Executive responsible for Candidate lifecycle decisions and without creating a second cognitive pipeline.

Without Candidate Routing, Learning produces a CandidateSnapshot but it never safely executes. With Candidate Routing, the rollout layer configures an upstream routing fork that safely triggers a Shadow execution, producing Reflection evidence for Governance. **Candidate Routing is an execution-routing concern, not a cognitive reasoning subsystem and not the Governance authority.**

##### 1. Complete Source-to-Destination Lifecycle

```text
                         ORIGINAL STIMULUS
                                │
                                ▼
                         SystemEvent / Envelope A
                                │
                                │ EnvelopeID = A
                                ▼
                          Event Router
                                │
                 ┌──────────────┴──────────────┐
                 │                             │
                 │ Active execution            │ Candidate eligible
                 ▼                             ▼
          ACTIVE ENVELOPE A              SHADOW FORK
                                             │
                                             ▼
                                      SHADOW ENVELOPE B
                                      │
                                      ├── ExecutionMode = SHADOW
                                      ├── SnapshotID = Candidate X
                                      └── OriginalEnvelopeID = A
                                             │
                                             ▼
                                          Executive
                                             │
                                             ▼
                                      Shadow Episode
                                             │
                                             ├── EpisodeID
                                             ├── SnapshotID
                                             ├── ExecutionMode
                                             └── ReplayMetadata
                                             │
                                             ▼
                                       Shadow Scope
                                             │
                                             ▼
                                      Candidate execution
                                             │
                                             ▼
                                         Reflection
                                             │
                                             ▼
                                   ReflectionReport
                                             │
                                             ├── SnapshotID
                                             ├── OriginalEnvelopeID
                                             └── evaluation evidence
                                             │
                                             ▼
                                    Promotion Authority
```

##### 2. Origin of the Original Stimulus

```text
Host / Timer / Sensor / Internal Event
              │
              ▼
          SystemEvent
              │
              ▼
        Event Router
              │
              ▼
           Envelope A
```
The original `SystemEvent` is created by the Host or sensors. The **Event Router** normalizes it into an `Envelope` with a unique `EnvelopeID`. The original Envelope remains authoritative for the Active execution. Candidate Routing must derive the Shadow execution from this exact same stimulus to ensure A/B evaluation validity.

##### 3. Event Router Owns the Fork

The architectural boundaries are strictly defined to preserve single responsibility:

*   **Rollout Executor** owns Candidate lifecycle and eligibility for Shadow evaluation.
*   **Event Router** owns creating the parallel Shadow stimulus representation *before* Executive orchestration.
*   **Executive** owns executing the Envelope as an Episode according to the execution information carried by the Envelope.

```text
Rollout Executor
     │
     │ candidate eligibility / routing configuration
     ▼
Event Router
     │
     │ stimulus fork
     ▼
Executive
     │
     │ Episode orchestration
     ▼
Shadow Episode
```
*Architectural Guardrail:* The Executive **must not** poll the Rollout Executor or inspect Candidate lifecycle state merely to decide whether to fork a stimulus. The Executive must remain blind to candidate rollout logic.

##### 4. Rollout Executor's Role

The Rollout Executor owns Candidate lifecycle and knows which CandidateSnapshots are eligible for Shadow. It configures Candidate routing rules on the Event Router. It **does not** perform cognitive execution, **does not** act as the Candidate Router or Executive, and **does not** perform Governance evaluation.

##### 5. The Envelope Fork

```text
Envelope A
   │
   │ semantic source
   ▼
Shadow fork
   │
   ▼
Envelope B
```

Envelope B preserves the same semantic stimulus while receiving a distinct runtime identity.

*   **`EnvelopeID`:** Identifies the actual Envelope instance being processed and avoids Event/WM deduplication collision.
*   **`ExecutionMode`:** Defines that this execution is SHADOW rather than normal Active execution.
*   **`SnapshotID`:** Identifies which CandidateSnapshot is being evaluated.
*   **`OriginalEnvelopeID`:** Expresses parallel evaluation lineage: *"This Shadow Envelope is an evaluation fork of Envelope A."*

*Why ParentRef is not used:* `ParentRef` implies a chronological/causal relationship (a reply). `OriginalEnvelopeID` implies parallel evaluation origin. These semantics must remain distinct.

##### 6. Identity Propagation Map

| Identity | Meaning | Authoritative Source | Propagated To | Why |
| :--- | :--- | :--- | :--- | :--- |
| **`EnvelopeID`** | Runtime identity of Envelope | Envelope | Episode / artifacts | Deduplication + tracing |
| **`OriginalEnvelopeID`**| Parallel fork origin | Shadow Envelope | Replay metadata / Reflection | Active-vs-Shadow correlation |
| **`SnapshotID`** | Candidate being evaluated | Shadow Envelope / candidate lifecycle| Episode / Reflection | Candidate provenance |
| **`EpisodeID`** | Cognitive execution identity | Executive | Reflection | Episode lifecycle |
| **`GoalID`** | Persistent Goal identity | Goal Manager | Goal-related execution | Goal lifecycle |
| **`ArtifactID`** | Internal payload identity | Existing contract | Downstream contracts | Provenance |

*(Note: We explicitly reject introducing `TurnID`, `ShadowTreeID`, or `EvaluationStimulusID` as they are redundant).*

##### 7. First-Class Candidate Execution Identity

`SnapshotID` is **not** stored inside `ModuleReferences`. `ModuleReferences` represents static capability/module bindings. `SnapshotID` represents execution provenance / candidate identity. Candidate identity must be elevated to a first-class field in the appropriate contract boundary:

```text
Envelope
   ↓
ReplayMetadata / Episode provenance
   ↓
ExecutiveEpisodeDefinition
   ↓
ReflectionReport
```

##### 8. Replay Determinism and Same-Stimulus Semantics

```text
Envelope A
   │
   ├── payload
   ├── ReplaySeed
   └── ReplayTimestamp
   │
   ▼
Shadow Envelope B
   │
   ├── same semantic payload
   ├── copied ReplayMetadata
   ├── new EnvelopeID
   ├── SnapshotID
   └── OriginalEnvelopeID = A
```
Envelope B guarantees equivalent evaluation context. `EnvelopeID` differs to protect deduplication, but the payload semantics remain identical. `ReplayMetadata` preserves deterministic conditions (seed, timestamp). Asynchronous processing does not alter the evaluation context because `ShadowScope` provides an equivalent read context bounded to the replay timestamp.

##### 9. Active vs Shadow Execution Context

```text
                   SAME STIMULUS
                        │
                ┌───────┴────────┐
                │                │
                ▼                ▼
           ACTIVE A          SHADOW B
           Envelope A        Envelope B
                │                │
                ▼                ▼
        Active Episode       Shadow Episode
                │                │
                ▼                ▼
          Active WM         Shadow Scope
                │                │
                ▼                ▼
        Active behavior    Candidate behavior
                │                │
                ▼                ▼
           Evidence A       Evidence B
                │                │
                └───────┬────────┘
                        ▼
                Promotion Authority
```

##### 10. Shadow Episode Creation

```text
Shadow Envelope B
      │
      ▼
Executive
      │
      ▼
ExecutiveEpisodeDefinition
      │
      ├── EpisodeID
      ├── ExecutionMode = SHADOW
      ├── SnapshotID
      └── replay / origin metadata
      │
      ▼
Shadow Episode
      │
      ▼
Shadow Scope
```
The Executive creates the Episode based on the Envelope. Candidate identity is already present in the Envelope's provenance. The Executive does not choose the Candidate or perform Governance; it simply orchestrates the requested execution mode.

##### 11. Candidate → Reflection Lineage

```text
SnapshotID
    │
    ▼
Shadow Envelope
    │
    ▼
ExecutiveEpisodeDefinition
    │
    ▼
Shadow Episode
    │
    ▼
Reasoning / cognitive work
    │
    ▼
Reflection
    │
    ▼
ReflectionReport
    │
    ├── SnapshotID
    ├── OriginalEnvelopeID
    └── EpisodeID
```
The `ReflectionReport` must be independently interpretable as: *"This evidence belongs to Candidate X executing the Shadow fork of original stimulus A during Episode E."* This empowers future Governance to compare evidence without reconstructing the full runtime graph.

##### 12. Shadow Child Lineage

```text
Original Envelope A
        │
        ▼
Shadow Parent Episode E1
        │
        ├── SnapshotID = X
        │
        ├── GapSignal
        │
        ▼
Shadow Child Episode E2
        │
        ├── ParentEpisodeID = E1
        ├── RootEpisodeID = E1
        ├── SnapshotID = X
        └── same Shadow Scope
```
No `ShadowTreeID` is necessary. The existing `ParentEpisodeID` and `RootEpisodeID` fields inherently preserve the execution tree within the unified `ShadowScope`.

##### 13. Failure / Recovery Scenarios

*   **Scenario A (Shadow Envelope fork fails):**
    `Envelope A` → `Shadow fork` → `FAIL`. The Active execution remains unaffected.
*   **Scenario B (Shadow Episode crashes):**
    `Shadow Episode` → `Crash`. Shadow state is discarded/recovered. Active state unaffected.
*   **Scenario C (Shadow duplicate / retry):**
    Distinct runtime IDs preserve deduplication while `OriginalEnvelopeID` preserves semantic correlation.
*   **Scenario D (Candidate routing rule changes):**
    If a Candidate becomes ineligible mid-flight, normal lifecycle semantics apply (e.g., in-flight episodes complete or are cooperatively cancelled based on policy, no complex rollbacks required).

##### 14. Ownership Matrix

| Responsibility | Owner | Must NOT Own |
| :--- | :--- | :--- |
| Candidate creation | Learning | Promotion |
| Candidate lifecycle | Rollout Executor | Cognitive execution |
| Candidate Shadow eligibility | Rollout Executor | Executive lifecycle |
| Routing configuration | Rollout Executor → Event Router | Cognitive reasoning |
| Envelope fork | Event Router | Candidate evaluation |
| Episode orchestration | Executive | Candidate selection |
| Shadow scope | Runtime/Shadow execution boundary | Authoritative WM |
| Cognitive evaluation | Reflection | Promotion authorization |
| Promotion decision | Promotion Authority | Episode execution |
| Lifecycle mutation | Rollout Executor | Cognitive evaluation |

##### 15. Architectural Rationale

*   **Why Event Router instead of Executive?** Because Executive should not become coupled to Candidate lifecycle/Governance.
*   **Why new Envelope identity?** Because reusing the original EnvelopeID collides with deduplication/idempotency.
*   **Why OriginalEnvelopeID?** Because normal ParentRef means causal/chronological relationship, not parallel evaluation.
*   **Why SnapshotID?** Because Candidate identity must survive independently from generic module references.
*   **Why ReplayMetadata?** Because the Shadow execution must preserve equivalent deterministic evaluation conditions.
*   **Why Shadow Scope?** Because the Candidate must be able to reason using equivalent context without contaminating Active state.
*   **Why ReflectionReport carries candidate provenance?** Because future Governance must be able to compare evidence without reconstructing the full runtime graph.

##### 16. Architecture vs Implementation vs Runtime Policy

| Category | Description |
| :--- | :--- |
| **Architecture (Resolved)** | Event Router owns stimulus fork. Rollout Executor owns Candidate lifecycle/eligibility. Executive owns Episode orchestration. Shadow execution is an Episode. Candidate identity is explicit. Fork lineage is explicit. Active and Shadow use semantically equivalent stimulus. Candidate identity propagates into Reflection evidence. |
| **Implementation (Deferred)** | Exact Event Router API, exact fork implementation, exact Envelope field types, exact database schema, exact serialization, exact ReplayMetadata implementation, exact routing-rule storage, exact Candidate selection code. |
| **Runtime Policy (Deferred)** | Shadow sampling percentage, candidate exposure percentage, shadow execution timeout, candidate eligibility duration, concurrency limits, routing refresh frequency. |

##### 17. The Promotion Authority Boundary

*Candidate Routing produces trustworthy Candidate execution evidence. It does not determine whether the Candidate is better than the Active baseline.*

```text
Candidate execution
      ↓
Reflection
      ↓
Candidate Evidence
      ↓
Promotion Authority
```

##### 18. Cross-Decision Traceability

*   **Decision 0.1:** No Shadow execution may alter the authoritative Goal/Gap lifecycle.
*   **Decision 0.2:** Active WM remains authoritative; ShadowScope remains temporary and isolated.
*   **Decision 0.3:** Shadow communication cannot reach the real Host.
*   **Decision 0.4:** Shadow execution and descendants remain bounded Episodes.
*   **Decision 0.5:** Executive remains the orchestrator and does not become candidate/rollout manager.
*   **Decision 0.6:** Candidate execution is Episode-based; no separate cognitive pipeline exists.
*   **Decision 0.7:** CandidateSnapshot → Routing → Shadow → Reflection is now fully documented.
*   **North Star:** Preserves single responsibility, loose coupling, simplicity, extensibility, and restart safety.

##### 19. Forbidden Interpretations (Guardrails)

Future implementations **MUST NOT** interpret this architecture to mean:
*   Executive polls Rollout Executor to choose Candidates.
*   Executive owns Candidate lifecycle.
*   Event Router performs cognitive evaluation.
*   Event Router becomes Governance.
*   `CandidateSnapshot` is stored in `ModuleReferences`.
*   `ParentRef` is used to mean Shadow fork lineage.
*   Shadow Envelope reuses original `EnvelopeID`.
*   Candidate execution bypasses Executive.
*   Candidate execution creates a second cognitive pipeline.
*   `SnapshotID` and `EpisodeID` are interchangeable.
*   `OriginalEnvelopeID` and `ParentRef` are interchangeable.
*   Reflection automatically approves promotion.
*   Candidate Routing automatically promotes a Candidate.
*   Candidate evidence automatically means Candidate is better.
*   Governance is already solved.
*   Shadow routing defines promotion thresholds.

#### J. Promotion Authority (Mechanical Governance)

**Architectural Purpose:** Promotion Authority exists as an independent, purely mechanical authorization boundary. It closes the gap between Candidate execution and Rollout promotion by comparing the cognitive evaluation evidence (`ReflectionReport`) of a Candidate against the Active baseline and authorizing the `RolloutExecutor` to advance the lifecycle.

This is a **new architectural boundary** that reuses and unites existing concepts (`ReflectionReport`, `RolloutRecommendation`, `RolloutExecutor`). It explicitly replaces the conflated use of `GovernanceBridge` for cognitive promotion.

##### 1. The Anti-Self-Evaluation Rule

Learning must not grade itself. Governance must not be cognitive. 

*   If Learning evaluated itself, it could hallucinate success. 
*   If Governance were a cognitive reasoning engine, it would violate Decision 0.6 by creating a hidden cognitive pipeline outside of an Episode. 

Instead, **Reflection** performs the cognitive evaluation (inside the Shadow Episode) and outputs structured metrics. **Promotion Authority** performs a purely *mechanical* evaluation by applying deterministic policies to those metrics.

##### 2. Complete Evidence Flow

```text
       ACTIVE A                     SHADOW B
     Active Episode             Shadow Episode
           │                            │
           ▼                            ▼
      Reflection                   Reflection
           │                            │
           ▼                            ▼
    Active Report                Candidate Report
           │                            │
           └─────────────┬──────────────┘
                         ▼
                Promotion Authority
               (Mechanical Governance)
                         │
                         ▼
               RolloutRecommendation
                  (Authorization)
                         │
                         ▼
                  Rollout Executor
                         │
                         ▼
             Candidate Lifecycle Mutation
```

##### 3. Candidate vs Active Comparison (The Evidence Join)

The Promotion Authority achieves precise A/B comparison without reconstructing complex runtime graphs by querying the Active and Shadow `ReflectionReports` and joining them.

**The Evidence Identity Contract:**
Evidence identity is not solely `OriginalEnvelopeID`. True evidence provenance requires the composite identity:
`OriginalEnvelopeID` + `SnapshotID` + `ExecutionMode` + `EpisodeID`

*   **`OriginalEnvelopeID`:** The join key. Guarantees both executions processed the exact same semantic workload (thanks to the Event Router fork).
*   **`SnapshotID`:** Distinguishes the Candidate from the Active baseline.
*   **`ExecutionMode`:** Prevents Shadow evidence from contaminating Active performance metrics.
*   **`EpisodeID`:** The primary key pointing to the detailed execution trace if deep-dive auditing is required.

*(Note: `ReflectionReport` must be structurally extended to natively carry `SnapshotID`, `ExecutionMode`, and `OriginalEnvelopeID`. They are required contract extensions, not currently implemented fields).*

##### 4. Recommendation vs Authorization Semantics

The `RolloutRecommendation` is an **Authorization**. 
The `RolloutExecutor` is mechanical infrastructure. It validates state-machine transitions (e.g., rejecting stale commands or invalid lifecycle jumps like Draft → Canary), but it **does not** second-guess the *reason* for promotion. If the Promotion Authority authorizes a valid transition, the Executor must execute it.

##### 5. Lifecycle Transition Matrix

| Transition | Authorized By | Trigger |
| :--- | :--- | :--- |
| **Draft → Validated** | `ValidationPipeline` | Pre-execution structural checks pass. |
| **Validated → Shadow** | `ExperimentManager` | Evaluation window opens. |
| **Shadow → Canary** | **Promotion Authority** | Mechanical A/B evidence shows improvement. |
| **Canary → Active** | **Promotion Authority** | Sustained positive evidence across population. |
| **Reject / Rollback** | **Promotion Authority** | Mechanical detection of regression. |

##### 6. Ownership Matrix

| Responsibility | Owner | Must NOT Own |
| :--- | :--- | :--- |
| Candidate Creation | Learning | Candidate Evaluation |
| Cognitive Evaluation | Reflection | Promotion Authorization |
| Process Diagnostics | `GovernanceBridge` | Cognitive Governance |
| Promotion Authorization | Promotion Authority | Cognitive Execution |
| Lifecycle Mutation | `RolloutExecutor` | Promotion Logic |

##### 7. Anti-Gaming and Safety Rules

The architecture inherently prevents "gaming" the promotion system:
*   **Holdout Evaluation:** Learning cannot write its own `ReflectionReport`. Reflection is a read-only observer.
*   **Same-Stimulus Comparison:** The Candidate cannot cherry-pick workloads; it must process the exact same `OriginalEnvelopeID` fork.
*   **Protected Metrics:** Promotion relies on deterministic metrics (e.g., Goal resolution, execution time, error rates) rather than self-reported confidence.

##### 8. Failure and Recovery Model

*   **Incomplete Evidence:** Promotion Authority takes no action. Candidate remains safely in Shadow.
*   **Promotion Authority Crashes:** Evaluation pauses. Candidate remains in current state (fail-safe).
*   **Rollout Executor Crashes:** `RolloutRecommendation` is idempotent. It can be safely re-processed upon restart.
*   **Evidence Disagrees (High Variance):** Mechanical policy defers to `ActionHumanReview`.

##### 9. Architecture vs Implementation vs Runtime Policy

| Boundary | Examples |
| :--- | :--- |
| **Architecture (Resolved)** | Promotion Authority exists as an independent mechanical authorization boundary. Learning cannot grade itself. Governance consumes joined `ReflectionReports`. Rollout Executor performs mutations. |
| **Implementation (Deferred)** | Exact Go interfaces, event bus handling, persistence schemas, precise `ReflectionReport` extensions, and the specific programmatic comparison functions. |
| **Runtime Policy (Deferred)** | Exact numeric promotion thresholds, required sample sizes, confidence intervals, Canary duration, and regression tolerance levels. |

---

### Decision 0.7 — Status

*   **Candidate Routing:** RESOLVED
*   **Shadow Sandbox:** RESOLVED
*   **Governance / Promotion Authority (Q0.7-2):** RESOLVED
*   **Rollout Executor (Q0.7-3):** RESOLVED

All major architectural boundaries for continuous cognitive evolution are now locked.

---

### Decision 0.8 — Reflection is Already Two-Mode

**Finding:** Reflection is already designed for `MODE_EPISODE` and `MODE_PERIODIC`. Both are implemented but neither is wired in production.

**Decision:** The Blueprint does not need to redesign Reflection. It needs to wire it. Phase 1 work.

---

## Part 1 — Architectural Principles

These 12 principles govern all architectural decisions in this Blueprint. Any proposed change that violates a principle requires explicit architectural justification.

| # | Principle | Implication |
|:--|:----------|:------------|
| **P1** | **Working Memory is omnipresent, not sequential.** | It is not a pipeline stage. It exists before, during, and after every cognitive stage. |
| **P2** | **Gap recognition is first-class.** | `KnowledgeGap`, `SkillGap`, `ClarificationGap`, `AuthorizationGap` are typed architectural signals, not error conditions. |
| **P3** | **Acquisition is not recursion.** | Knowledge and Skill acquisition trigger continuation episodes, not recursive subsystem re-invocation. |
| **P4** | **Conversation is a cognitive decision.** | What to communicate is a cognitive function. How to express it is a presentation function. These are distinct boundaries. |
| **P5** | **Goals are managed; episodes are bounded.** | A goal lives across multiple episodes. An episode is a bounded unit of work. A turn is a single stimulus-response cycle. |
| **P6** | **Learning does not write directly.** | Learning produces `CandidateSnapshot` objects. A Rollout Executor activates them. Learning never directly mutates active cognitive state. |
| **P7** | **Reflection evaluates; it does not modify.** | Reflection is a pure read-only evaluator. Its sole output is a structured `ReflectionReport`. |
| **P8** | **Acquired information is not automatically trusted.** | All acquired knowledge carries confidence, provenance, and verification status. |
| **P9** | **Acquired skills are not automatically executable.** | Skills must pass validation, security check, and sandbox before being marked `Available`. |
| **P10** | **Autonomy never bypasses authorization.** | Autonomy level governs what IDUN may do without asking. The Constitution governs what IDUN may never do regardless of autonomy level. |
| **P11** | **The architecture must survive technology changes.** | Models, hardware, APIs, and tools are components behind stable interfaces. The architecture is the system; components are replaceable. |
| **P12** | **Simplest design that meets the requirement.** | A new subsystem, service, or contract must justify its existence by solving a problem that cannot be solved by extending an existing mechanism. |

---

## Part 2 — The Canonical Cognitive Lifecycle

This is the authoritative description of how IDUN processes a stimulus. Every implementation must be consistent with this lifecycle.

### 2.1 Process Classifications

- **PIPELINE:** Sequential stages where the output of one feeds the next.
- **STORE:** Persistent structures read from and written to throughout the lifecycle.
- **ASYNC:** Processes that run concurrently or after the main pipeline without blocking it.
- **EVENT:** Signal-driven triggers that enter the pipeline from outside.
- **ROUTING:** Deterministic signal dispatch — no intelligence required.

### 2.2 The Lifecycle

```
════════════════════════════════════════
  STIMULUS ENTRY
════════════════════════════════════════

STIMULUS                                                    [EVENT]
(Host input / Timer / System event / Goal monitor / Tool result)
        │
        ▼
EVENT NORMALIZATION                                         [PIPELINE]
Wrap stimulus in normalized Event Envelope.
Attach metadata: source, timestamp, type, priority hint.
        │
        ▼
RELEVANCE CHECK                                             [PIPELINE]
Is this event relevant to any active goal?
Is it worth processing at all?
        │                         │
   Relevant                  Not relevant
        │                         │
        │                      Discard or log only
        ▼
AUTONOMY POLICY EVALUATION                                  [POLICY]
For non-Host events: what is IDUN permitted to do?
  → Route to full cognitive pipeline (if autonomy permits action)
  → Queue for Host confirmation (if action exceeds autonomy level)
  → Discard (if irrelevant and no notification warranted)
Note: all proactive notification paths flow through Attention → Conversation Planner.
Notification Service is never routed to directly from Event Router.
        │
        ▼
ATTENTION TRIAGE                                            [PIPELINE]
Receives: SystemEvent from Event Router.
Reads: Working Memory (ActiveGoals, CurrentEntities, ActiveTopic) — contextual salience scoring.
Evaluates: salience, urgency, contextual relevance.
Groups semantically related events where appropriate (semantic grouping is cognitive).
Assigns priority band (0–4). Reserves budget for this priority level.
Higher-band events preempt lower-band work in progress.
Emits: SalientEvent (original SystemEvent + salience score + urgency band + semantic grouping).
        │
        ▼
════════════════════════════════════════
  COGNITIVE PIPELINE
════════════════════════════════════════

UNDERSTANDING                                               [PIPELINE]
Normalize input text.
Multi-intent splitting with Connector Registry.
Specialist cascade: Grammar → Neural → Deliberative LLM (fallback).
Semantic slot extraction.
Produce: SemanticFrame / UnderstandingBatch (immutable).
        │
        ▼
CONTEXT RESOLUTION                                          [PIPELINE]
Resolve pronouns, ellipses, temporal references.
Enrich with dialogue state.
Produce: ResolvedContext (wraps UnderstandingBatch — does NOT mutate it).
        │
        ▼
WORKING MEMORY — WRITE (context update)                     [STORE WRITE]
Write resolved context into Working Memory.
Update: ActiveTopic, CurrentEntities, RecentCorrections.
Read: existing context, active goals, current plan.
(Working Memory exists before this and continues after — this is a write, not a stage.)
        │
        ▼
REASONING                                                   [PIPELINE]
Read: Working Memory context + Knowledge Store records.
11-stage cascade: Symbolic → Graph → CSP → Bayesian →
  Analogy → Beam → Calibration → Deliberative LLM → Constitution.
Emit: ReasoningResult + GoalProposal + GapSignals.
        │
        ├──── KnowledgeGap detected? ──────────────────────────────────┐
        │                                                              │
        │     KnowledgeGap emitted by Reasoning                       │
        │     → transported via Event Bus (content-blind)             │
        │     → received by Goal Manager                              │
        │                                                              │
        │     GOAL MANAGER — GAP LIFECYCLE                            │
        │     Pause goal.                                             │
        │     Create GapRecord (with Deadline = Level 2 timeout).     │
        │     Dispatch to Knowledge Acquisition via routeGap().       │
        │     Monitor resolution deadline.                            │
        │     On SUCCESS: schedule Continuation Episode.              │
        │     On FAILURE/TIMEOUT: retry / fail / Host intervention.   │
        │                    │                                         │
        └────────────────────┘ (continue after acquisition)           │
        │
        │ NOTE — ClarificationGap (RESOLVED — see §0.1.J):
        │ ClarificationGap originates at Context Resolver / Understanding
        │ BEFORE this stage. It always routes through Goal Manager.
        │ Goal Manager applies the two-tier wait model:
        │   Active goal blocked → persisted GapRecord → PAUSED
        │   No active goal blocked → PendingClarification token
        │                            in Working Memory
        │ Conversation Planner generates the clarification question.
        │ There is NO bypass path from Context Resolver to
        │ Conversation Planner.
        │
        │ NOTE — AuthorizationGap (see Constitutional Gate below):
        │ AuthorizationGap is produced by the Constitutional Gate.
        │ It routes through Goal Manager, which signals Conversation
        │ Planner to generate the authorization request.
        │
        ▼
GOAL PROPOSAL → GOAL MANAGER                                [EVENT]
Reasoning emits GoalProposal (not directly to TopicActiveGoals).
Goal Manager validates GoalProposal.
Registers goal if new. Updates goal state if continuation.
Provides active goals to Planning.
        │
        ▼
PLANNING                                                    [PIPELINE]
Read: Active goals (from Goal Manager) + Working Memory +
      Knowledge Store + Skill Registry + constraints.
HTN decomposition / GOAP / A* / Beam search.
Emit: CandidatePlans.
If gap encountered: emit SkillGap signal
  → transported via Event Bus (content-blind) → Goal Manager
  → Goal Manager manages SkillGap lifecycle (pause / dispatch / monitor).
        │
        ▼
DECISION                                                    [PIPELINE]
Read: CandidatePlans + Working Memory + Host preferences +
      Autonomy level + Calibration weights.
Evaluate: goal alignment, constraints, risk, cost, safety.
Emit: DecisionRecord.
  COMMIT             → Executive
  DEFER              → Goal Manager pauses; Working Memory updated
  ABSTAIN            → Conversation Planner explains
  ESCALATE           → Re-evaluate at deliberative depth
  REQUEST_CONFIRMATION → Conversation Planner asks Host
  REQUEST_INFORMATION  → GapSignal emitted → Goal Manager → Acquisition (via routeGap)
  SUGGEST_ALTERNATIVES → Conversation Planner presents options
        │
        ▼ (on COMMIT)
CONSTITUTIONAL GATE                                         [PIPELINE]
Pre-broadcast interception of world-modifying actions.
HMAC-signed approval required.
Veto → Conversation Planner explains; execution aborted.
Approval → Executive dispatches.
        │
        ├──── AuthorizationGap detected? ────────────────────────────────┐
        │                                                                │
        │     AuthorizationGap emitted by Constitutional Gate           │
        │     → transported via Event Bus (content-blind)               │
        │     → received by Goal Manager                                │
        │     → Goal Manager signals Conversation Planner to            │
        │       generate authorization request to Host                  │
        │     → Goal → AWAITING_HOST (Host-Response branch,             │
        │       no acquisition dispatch or Level 2 timeout)             │
        │     → Host approves or denies                                 │
        │                    │                                           │
        └────────────────────┘ (continue to Executive on approval)      │
        │
        ▼
EXECUTIVE                                                   [PIPELINE]
Content-blind orchestration.
Budget management and concurrency control.
Dispatch to Capability/Skill execution layer.
Emit: ActionExecution.
        │
        ▼
CAPABILITY / SKILL EXECUTION                                [PIPELINE]
Execute via registered SkillCard or native capability.
Produce: ExecutionOutcome + Observations.
        │
        ▼
WORKING MEMORY — WRITE (outcome update)                     [STORE WRITE]
Write execution outcome: RecentObservations, CurrentPlan status.
        │
        ▼
════════════════════════════════════════
  COMMUNICATION
════════════════════════════════════════

CONVERSATION PLANNER                                        [PIPELINE]
Receives: SalientEvent from Attention.
Reads: Reasoning result + Decision record + Execution outcome +
      Working Memory (ConversationTurns, ActiveTopic, HostPreferences) + Gap signals.
Determines communicative intent:
  ANSWER / CLARIFY / WARN / ACKNOWLEDGE / EXPLAIN /
  RECOMMEND / REFUSE / NOTIFY / ASK_CONFIRMATION /
  PRESENT_OPTIONS / REMAIN_SILENT
Writes: outbound intent to Working Memory (ConversationTurns) via WorkingMemoryManager
        so that an immediate Host response can be correctly understood by Understanding.
Emits: ConversativeIntent (structured — not yet language).
        │
        ▼
CONSTITUTIONAL GATE (communication path)                    [PIPELINE]
All ConversativeIntent — including proactive NOTIFY — must pass this gate.
Gate unavailable → Intent dropped; never fail open.
        │
        ▼
LANGUAGE REALIZATION                                        [PIPELINE]
Receives: ConversativeIntent (after Constitutional Gate approval).
Deterministic rendering for factual outputs (time, status, numbers).
LLM rendering for conversational outputs.
Emits: Rendered Message → Notification Service.
        │
        ▼
NOTIFICATION SERVICE                                        [DELIVERY]
Receives: Rendered Message from Language Realization.
Mechanical delivery operations: queue, retry, rate limit,
  mechanical deduplication, delivery ordering, modality/channel selection.
Delivers to Host / World.
        │
        ▼
HOST / WORLD                                                [OUTPUT]

════════════════════════════════════════
  EPISODE CLOSURE (if episode-level work)
════════════════════════════════════════

REFLECTION — Episode mode (async)                           [ASYNC]
Triggered by Executive after meaningful episode closes.
Read-only consumption of episode traces from Workspace.
11-specialist evaluation.
Emit: ReflectionReport to Workspace.
(Non-blocking — cognitive pipeline continues uninterrupted.)
        │
        ▼ (async)
LEARNING                                                    [ASYNC]
Consume: ReflectionReport (COGNITIVE_PERFORMANCE category only).
Aggregate: windowed experience corpus.
Generate: CandidateSnapshot per domain.
Validate: Draft → Validated.
Publish: to SnapshotRegistry.
(Rollout Executor activates via Shadow → Canary → Active pipeline.)
        │
        ▼ (async — Rollout Executor activates)
UPDATES PROPAGATE TO:
  Reasoning S1 rules (compiled from snapshot)
  Planning strategy (atomic snapshot swap)
  Decision policy profile (atomic snapshot swap)
  Attention weights (atomic snapshot swap)
  Knowledge Store (new verified knowledge)
  Skill Registry (skill confidence updates)
  Conversation preferences (interaction preference store)
```

### 2.3 Prohibited Recursive Loops

| Pattern | Why Prohibited | Correct Alternative |
|:--------|:--------------|:-------------------|
| Reasoning invokes itself | Unbounded depth, unclear episode boundary | Continuation episode with enriched Working Memory |
| Planning invokes Reasoning | Breaks responsibility boundary | Planning reads Reasoning result; does not invoke Reasoning |
| Learning writes directly to active state | Bypasses rollout safety | CandidateSnapshot → Rollout Executor → activation |
| Reflection modifies any subsystem | Feedback loops; pure evaluator boundary violated | Reflection → ReflectionReport → Learning → CandidateSnapshot |
| Working Memory directly triggers cognition | Working Memory is passive | Events trigger cognition; Working Memory is read during it |

---

## Part 3 — Subsystem Catalog

Each entry formally defines a subsystem or service.

**Status labels:**
- 🟢 **CURRENT** — Implemented and production-ready
- 🟡 **CURRENT / INCOMPLETE** — Implemented but not fully wired in production
- 🔵 **APPROVED ARCHITECTURE** — Specified; implementation pending
- 🟠 **PROPOSED REFINEMENT** — Proposed in Cognitive Operating Workflow; refined here
- ⚪ **NORTH STAR FUTURE** — Required for North Star maturity; not needed immediately

---

### SS-01 — World / Input Adapters

**Status:** 🟢 CURRENT | **Classification:** Service | **Package:** `world/`

**Purpose:** Provide IDUN's sensory connection to the external world. Normalize each input modality into an Event Envelope.

**Owns:** I/O lifecycle, input normalization, output delivery per modality.
**Must NOT Own:** Cognitive interpretation, routing, memory persistence.

**Inputs:** Raw stimuli (keyboard, voice, API calls, file content)
**Outputs:** Normalized `Event Envelope`
**Failure:** Log I/O failure; do not fabricate input.

---

### SS-02 — Event Router

**Status:** 🟠 PROPOSED REFINEMENT | **Classification:** Service | **Package:** `runtime/eventrouter`

**Purpose:** Normalize events from all sources into a unified `SystemEvent` envelope, apply Autonomy Policy, and route to Attention for cognitive salience evaluation.

**Owns:** `SystemEvent` normalization, source metadata, Autonomy Policy evaluation, relevance pre-screening.
**Must NOT Own:** Cognitive evaluation of event content, memory access, reasoning, or delivery decisions.

**Inputs:** Raw event signals from all sources (Host input, Scheduler, Goal Manager lifecycle, Executive progress)
**Outputs:** `SystemEvent` → Attention
**Persistent State:** None (stateless router)
**Failure:** Drop with log; never silently deliver malformed events. Never route directly to Notification Service — all proactive communication flows through the cognitive pipeline.

---

### SS-03 — Attention

**Status:** 🟡 CURRENT / INCOMPLETE — implemented; not wired in production
**Classification:** Service | **Package:** `intelligence/attention`

**Purpose:** Apply salience scoring, urgency assignment, priority band assignment, and semantic grouping of related events before events enter the cognitive pipeline. Produces the `SalientEvent` contract that Conversation Planner consumes.

**Priority Bands:**

| Band | Name | Examples | Preempts |
|:-----|:-----|:---------|:---------|
| 0 | Critical Safety | Constitutional violations, critical failures | All |
| 1 | Real-Time Interactive | Active Host conversation | 2–4 |
| 2 | Active Goal Execution | Executing an approved plan step | 3–4 |
| 3 | Background Goal Work | Research, monitoring | 4 |
| 4 | Maintenance | Reflection, periodic learning | Nothing |

**Owns:** Salience scoring, urgency evaluation, semantic grouping of related events, priority band assignment, budget reservation.
**Must NOT Own:** Content interpretation beyond salience scoring, communicative strategy, delivery decisions.

**Inputs:**
- `SystemEvent` from Event Router
- Working Memory (read): `ActiveGoals`, `CurrentEntities`, `ActiveTopic` — for contextual salience scoring

**Outputs:** `SalientEvent` → Conversation Planner
- `SalientEvent` carries: the original `SystemEvent`, salience score, urgency band, relevance reason, and semantic group identifier if events were grouped.

**Semantic Grouping:** When multiple related `SystemEvent`s arrive within a contextual window, Attention may group them into a single `SalientEvent` so Conversation Planner can generate one coherent `ConversativeIntent`. Grouping is cognitive (understands event relationships); mechanical deduplication remains with the Notification Service.

**Failure:** If Working Memory is unavailable, fall back to static priority band scoring without contextual enrichment. Log degradation. Do not block the pipeline.

**Pending Wiring:** Must be wired to workspace subscription in `runtime/host.go` (Phase 1).

---

### SS-04 — Understanding

**Status:** 🟡 CURRENT / INCOMPLETE — U8 frozen; Deliberative LLM not wired
**Classification:** Service | **Package:** `intelligence/understanding` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Transform raw natural language input into a structured, deterministic `SemanticFrame` / `UnderstandingBatch`.

**Owns:** Text normalization, multi-intent splitting, specialist cascade (Grammar → Neural → Deliberative LLM), semantic slot extraction, ambiguity beam construction.
**Must NOT Own:** Reference resolution, reasoning, memory access, world knowledge.

**Pending Wiring:** `InferenceService` must be injected for Deliberative LLM stage (Phase 1).

---

### SS-05 — Context Resolver

**Status:** 🟢 CURRENT — U7.5 frozen
**Classification:** Service | **Package:** `intelligence/context` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Resolve implicit linguistic references in `UnderstandingBatch` against conversational history.

**Boundary Clarification:** Context Resolver must produce a `ResolvedContext` wrapper. It must NOT mutate the `UnderstandingBatch` in place — that is the immutable Understanding contract.

**Owns:** Pronoun, ellipsis, temporal, confirmation strategy resolution.
**Must NOT Own:** Inventing new intents, planning, direct Working Memory writes.

---

### SS-06 — Working Memory

**Status:** 🔵 APPROVED ARCHITECTURE (with refinements) — new, highest priority
**Classification:** Store + Manager | **Package:** `intelligence/workingmemory`

**Purpose:** A bounded, continuously-updated, session-scoped cognitive context store. The shared active cognitive state readable by all subsystems and writable only by designated subsystems through defined contracts. Synchronous in-process access on the hot cognitive path.

**Problem Solved:** Without Working Memory, each episode re-derives context from scratch. Reasoning cannot see what goals are active. Planning cannot see what the Host said two turns ago. Conversation Planner cannot calibrate verbosity to Host preferences.

**Architecture:** Store + WorkingMemoryManager (see Decision 0.2 §0.2.A–§0.2.H for the authoritative specification). This section is the subsystem record.

**Data Model (Slots by Lifetime Scope):**

*GLOBAL scope — session-scoped, shared across all goals*

| Slot | Description | Write Authority | Checkpoint |
|:-----|:------------|:----------------|:-----------|
| `ConversationTurns` | Sliding window of recent turns | Understanding, Conversation Planner | NOT persisted |
| `ActiveTopic` | What is currently being discussed | Context Resolver | NOT persisted |
| `CurrentEntities` | Named entities from recent turns | Context Resolver | NOT persisted |
| `RecentCorrections` | Host corrections in this session | Understanding | NOT persisted |

*GOAL scope — per-GoalID, isolated between goals*

| Slot | Description | Write Authority | Checkpoint |
|:-----|:------------|:----------------|:-----------|
| `ActiveGoalID` | GoalID this context belongs to (reference) | Goal Manager | REQUIRED |
| `ActiveSubgoalIDs` | Active decomposed sub-goal IDs (references) | Goal Manager | REQUIRED |
| `ActiveGapIDs` | IDs of in-flight gaps for this goal (references only — GapRecord authority stays with Goal Manager) | Goal Manager | RECOMMENDED |

*EPISODE scope — per-GoalID, episode-bounded*

| Slot | Description | Write Authority | Checkpoint |
|:-----|:------------|:----------------|:-----------|
| `CurrentPlan` | Plan currently being executed | Planning | RECOMMENDED |
| `PlanCheckpoint` | Serialized paused plan state (for pause/resume) | Planning | **REQUIRED** |
| `ActiveBeliefs` | Working assumptions derived by Reasoning | Reasoning | RECOMMENDED |
| `UnresolvedQuestions` | Questions raised, not yet answered | Conversation Planner | RECOMMENDED |
| `RecentObservations` | Execution outcomes this episode | Executive | RECOMMENDED |
| `TemporaryResearch` | Acquired info, not yet verified | Knowledge Acquisition | **NOT persisted** |
| `RetrievedMemories` | LTM records fetched for this episode | MemoryRetriever (interface) | **NOT persisted — re-fetched on resume** |

*TURN scope — single stimulus-response cycle*

| Slot | Description | Write Authority | Checkpoint |
|:-----|:------------|:----------------|:-----------|
| `PendingClarification` | Transient clarification token (non-blocking path only — see §0.2.G) | Goal Manager | NOT persisted (transient path); GapRecord reference checkpointed (persistent path) |

**`PendingGaps` slot removed.** Replaced by `ActiveGapIDs` (reference-only). GapRecord authority remains exclusively with Goal Manager in `core/storage`. See Decision 0.2 §0.2.B.

**Read Authority:** All subsystems. Reading is non-exclusive and synchronous (~10 ns in-process).

**Write Authority Enforcement:** Write authority is checked inside the write lock. An unauthorized write returns an error and never silently succeeds.

**Capacity Rules:**
- `ConversationTurns`: Sliding window, max N turns (configurable; calibrate in Phase 2)
- `CurrentEntities`: Max N; lowest-priority evicted when full (configurable)
- `RecentObservations`: Sliding window, max N (configurable)
- `RecentCorrections`: Max N; oldest evicted (configurable)
- `RetrievedMemories`: Max N records per episode (configurable)
- `TemporaryResearch`: Max N records; evicted on episode close (configurable)
- Dormant goal context RAM limit: Configurable runtime policy (SS-32)

**Concurrency:** Slot-level RWMutex (or equivalent). Global context: shared RWMutex. Per-goal context: per-goal RWMutex. `ForegroundGoalID`: atomic pointer. Write authority enforced inside the lock. See Decision 0.2 §0.2.E.

**Conflict Handling:** When a new belief contradicts an active belief: mark old as `CONTRADICTED`, add `ContradictionFlag`, notify Reasoning on next invocation. Do NOT silently overwrite.

**Lifetime Scopes:** GLOBAL (session), GOAL (goal-lifetime), EPISODE (bounded), TURN (single cycle). See Decision 0.2 §0.2.A for the complete lifetime hierarchy.

**Persistent State:**
- GOAL + required EPISODE slots (especially `PlanCheckpoint`): checkpointed to `core/storage` atomically on episode pause
- SESSION + transient TURN slots: in-RAM only; not persisted
- `TemporaryResearch` and `RetrievedMemories`: not persisted; re-derived or re-fetched on resume

**Failure and Restart:**
- SESSION context: start fresh (do NOT restore stale session context into new session)
- GOAL + EPISODE checkpoint: restore from `core/storage` if checkpoint is present and consistent
- If no checkpoint or inconsistent: initialize empty goal context; Goal Manager re-derives minimal state from Goal record
- Checkpoint write must be atomic; failed checkpoint prevents episode pause acknowledgement (see §0.2.G)

**Consolidation at Episode Close:**
- Executive triggers `WorkingMemoryManager.Consolidate(episodeID)`
- WorkingMemoryManager applies retention policy (see §0.2.H)
- Reflection recommends; WorkingMemoryManager executes promotion to Long-Term Memory
- Reflection does NOT directly write to Long-Term Memory (AI-09 preserved)

**WorkingMemoryManager Interface:**
```
WorkingMemoryManager (interface — injected into all cognitive subsystems)
  Read(ctx, slotID) → any
  ReadGoalScope(ctx, goalID, slotID) → any
  Write(ctx, slotID, value) error
  WriteGoalScope(ctx, goalID, slotID, value) error
  SetForegroundGoal(ctx, goalID) error
  Checkpoint(ctx, episodeID) error
  Restore(ctx, sourceEpisodeID, targetEpisodeID) error  // Decision 0.4: cross-episode checkpoint transfer
  Consolidate(ctx, episodeID) error
  Reset(ctx, scope LifetimeScope) error
```

**Deferred Policies:** Exact capacities and weights are runtime configuration (see Decision 0.2 §0.2.J). Core architecture is fully resolved.

---

### SS-07 — Reasoning

**Status:** 🟡 CURRENT / INCOMPLETE — V3 frozen; Deliberative LLM not wired; GoalProposal and KnowledgeGap emission needed
**Classification:** Service | **Package:** `intelligence/reasoning` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Transform `ResolvedContext` and Working Memory context into a `ReasoningResult`, identifying goals, conclusions, and knowledge gaps.

**Boundary Refinement:** Reasoning currently publishes directly to `TopicActiveGoals`. The correct model: Reasoning emits a `GoalProposal` → Goal Manager validates and registers it.

**Gap Classification Ownership (confirmed by Decision 0.1 review):** Reasoning detects and classifies **`KnowledgeGap` only**. It does not classify `SkillGap` (Planning's responsibility), `ClarificationGap` (Context Resolver's responsibility), or `AuthorizationGap` (Constitutional Gate's responsibility).

**KnowledgeGap Extension (additive — does not modify frozen spec):**
```
ReasoningResult {
    ...existing fields...
    GoalProposal   *GoalProposal     // proposed goal for Goal Manager
    KnowledgeGaps  []KnowledgeGap    // gaps in IDUN's knowledge
}

KnowledgeGap {
    Type        GapType    // always KNOWLEDGE from Reasoning
    Domain      string     // knowledge domain that is missing
    Description string     // what specifically is unknown
    Urgency     float64    // 0.0 – 1.0 (how blocking is this gap)
}
```

All emitted `KnowledgeGap` signals are transported via the Event Bus (content-blind) and received by Goal Manager, which manages the gap lifecycle.

**Owns:** 11-stage cascade (S0–S10), symbolic inference, graph reasoning, CSP, Bayesian fusion, analogy, beam, calibration, deliberative LLM fallback, constitution integration, GoalProposal emission, `KnowledgeGap` detection and emission.
**Must NOT Own:** `SkillGap` detection (Planning), `ClarificationGap` detection (Context Resolver), `AuthorizationGap` detection (Constitutional Gate), generating candidate plans, selecting actions, storing facts, deciding communication, executing capabilities.

**Pending Wiring:** `InferenceService` injection for S8 Deliberative Specialist (Phase 1).

---

### SS-08 — Gap Signal Routing

**Status:** 🔵 APPROVED ARCHITECTURE — refined by Decision 0.1 review
**Classification:** Pure routing function embedded in Goal Manager (NOT a subsystem, NOT a runtime middleware component)

**Purpose:** Map a detected gap type to the correct dispatch topic so Goal Manager can send it to the appropriate acquisition subsystem.

**Architecture (confirmed):**

```
Gap-detecting subsystem
         │ emits GapSignal
         ▼
  Workspace Event Bus
  (content-blind transport)
         │
         ▼
    Goal Manager
         │ calls routeGap(signal.Type)
         ▼
  Pure routing function
  (no state, no goroutines, no lifecycle)
         │
    ┌────┴────┐
    ▼         ▼
 TopicKnow-  TopicSkill-
 ledgeGap    GapRequested
 Requested   (→ Skill Acq.)
 (→ Know. Acq.)
```

**Why NOT an infrastructure routing component:** The Event Bus must remain content-blind. Placing a routing rule table in the infrastructure layer requires the infrastructure to inspect cognitive gap types — a boundary violation. Goal Manager is the correct owner because it already owns the gap lifecycle.

**Why NOT a separate subsystem:** A pure deterministic function (`GapType → Topic`) does not require goroutines, state, or a lifecycle. Principle P12 applies.

**Gap Type → Dispatch Topic Mapping:**

| Gap Type | Detecting Subsystem | Dispatch Topic (via Goal Manager) | Branch |
|:---------|:--------------------|:----------------------------------|:-------|
| `KnowledgeGap` | Reasoning | `TopicKnowledgeGapRequested` → Knowledge Acquisition | Acquisition |
| `SkillGap` | Planning | `TopicSkillGapRequested` → Skill Acquisition | Acquisition |
| `ClarificationGap` | Context Resolver / Understanding | Goal Manager two-tier wait model (§0.1.J): persisted `GapRecord` or transient `PendingClarification` token → Conversation Planner → Host | Host-Response |
| `AuthorizationGap` | Constitutional Gate | Goal Manager → Conversation Planner → authorization request → Host | Host-Response |

---

### SS-09 — Goal Manager

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Service | **Package:** `intelligence/goalmanager`

**Purpose:** Manage the lifecycle of all active, suspended, queued, and completed goals. Own the complete gap resolution lifecycle for any gap that requires pausing a goal.

**Problem Solved:** Without Goal Manager, IDUN can pursue only one goal at a time, cannot remember suspended goals, and has no mechanism to resume a paused plan after acquisition completes. Without a single timeout owner, gap resolution can stall indefinitely.

**Gap Resolution Protocol (confirmed by Decision 0.1 review):**
1. Goal Manager receives `GapSignal` from Event Bus
2. Creates `GapRecord` with `{GapID, GoalID, Type, Domain, StartedAt, Deadline, RetryCount: 0}`
3. Transitions Goal: `ACTIVE → PAUSED`
4. Calls `routeGap(signal.Type)` to determine dispatch topic
5. Publishes gap to appropriate acquisition subsystem
6. Monitors Level 2 deadline
7. On `GapResolved{Status: SUCCESS}`: cancel timer, schedule Continuation Episode, transition Goal `PAUSED → ACTIVE`
8. On `GapResolved{Status: FAILURE}` or Level 2 timeout: retry (if RetryCount < Max) or escalate
9. On escalation: transition Goal `PAUSED → AWAITING_HOST`; notify via Conversation Planner

**Owns:** Goal registry, goal lifecycle state machine, gap lifecycle (GapRecord creation, dispatch via `routeGap()`, Level 2 timeout monitoring, retry, escalation), pause/resume protocol, multi-goal priority management, Continuation Episode scheduling, goal deadline monitoring, sub-goal management.

**Must NOT Own:** Planning strategy, decision logic, memory retrieval, cognitive interpretation, Level 1 operation timeouts (those are owned by Acquisition subsystems internally).

**Must NOT Become:** A cognitive subsystem that reasons about the content or meaning of gaps. It manages gap lifecycle mechanically; it does not understand what the gaps mean.

**Goal Lifecycle States:**
`Proposed → Validated → Active → Paused(gap_pending) → Active (resumed) → Monitoring → Active (condition met) → Completed | Failed | Abandoned → Archived`

Additionally: `Active | Paused → AWAITING_HOST` (when all retries exhausted)

**Multi-Goal Priority Rules:**
- Only one Band 1 (interactive) goal at a time
- Multiple Band 3–4 (background) goals permitted concurrently
- Band 0 (safety) always preempts everything
- Resource conflict resolved by priority score, not arrival order

**Persistent State:** Goal registry and GapRecords persisted to `core/storage` under `goals/` namespace. Must survive IDUN restarts. See Decision 0.1.I for restart recovery procedure.

---

### SS-10 — Knowledge Acquisition

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Service | **Package:** `intelligence/knowledge/acquisition`

**Purpose:** Resolve `KnowledgeGap` signals by acquiring missing information, evaluating source quality, and storing verified knowledge.

**Complete Lifecycle:**
```
KnowledgeGap signal received
        ↓
Classify gap type (factual-stable / factual-temporal / contextual / procedural / causal)
        ↓
Select acquisition strategy by type and online/offline state:
  Local:  Working Memory → Knowledge Store → Episodic Memory (analogy)
  Online: Tool use → Internet research
  Host:   Clarification request (automated acquisition insufficient)
        ↓
Acquire raw information from selected source
        ↓
Extract and normalize relevant claims
        ↓
Evaluate source quality:
  Source type · Freshness · Cross-reference · Internal consistency · Provenance
  → Assign confidence score (0.0–1.0)
        ↓
Cross-check against existing Knowledge Store
  Consistent  → Accept
  Contradicted → Store with ConflictFlag; do not overwrite existing record
  Uncertain    → Store as UNVERIFIED with low confidence
        ↓
Classify result durability:
  VERIFIED (confidence ≥ threshold) AND durable
        → WRITE Knowledge Store (insert-if-absent, F3)
        → Emit GapResolved{GapID, Status=SUCCESS, StoreKey} → Goal Manager
  UNVERIFIED (confidence < threshold) but durable domain knowledge
        → WRITE Knowledge Store with UNVERIFIED status + ConflictFlag if contradicted
        → Emit GapResolved{GapID, Status=SUCCESS, StoreKey} → Goal Manager
  TRANSIENT/CONTEXTUAL (inherently temporal — e.g., current weather)
        → DO NOT write to Knowledge Store (preserves LTM integrity)
        → Emit GapResolved{GapID, Status=SUCCESS, Payload=[opaque transient result]}
        → Goal Manager carries Payload through ContinuationRequest to Executive
        → Executive injects into E2 TemporaryResearch (EPISODE scope)
  CONFLICTED (contradicts existing Knowledge Store record)
        → WRITE Knowledge Store with ConflictFlag + Host notification
        → Emit GapResolved{GapID, Status=SUCCESS, StoreKey} → Goal Manager
  FAILURE (acquisition failed, all retries exhausted)
        → Emit GapResolved{GapID, Status=FAILURE} → Goal Manager
```

**Offline Behavior:** Local sources → defer with retry schedule → inform Host if blocking. Do NOT fabricate information.

**Owns:** Strategy selection, source evaluation, normalization, confidence scoring, cross-reference, Knowledge Store writes, contradiction detection.
**Must NOT Own:** Reasoning over acquired knowledge, deciding whether knowledge is useful for the goal, executing capabilities.

---

### SS-11 — Knowledge Verification

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Component (pipeline stage within Knowledge Acquisition) | **Location:** `intelligence/knowledge/acquisition`

**Why Not a Separate Package:** Verification is a stage within acquisition, not a lifecycle service. It has no goroutines, no separate state, and no independent triggers. Principle P12 applies.

**Verification Dimensions:**

| Dimension | Assessment |
|:----------|:-----------|
| Source type | Official / encyclopedia / reference / forum / unknown |
| Freshness | Source age × domain volatility coefficient |
| Cross-reference | N independent corroborating sources |
| Internal consistency | Does not contradict high-confidence existing records |
| Provenance | Source identifier stored for future re-verification |

**Unverifiable Information:** Store in Working Memory as `UNVERIFIED`. Reasoning uses with explicit caveat. Do NOT store as permanent Knowledge Store fact.

---

### SS-12 — Knowledge Store

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Store | **Package:** `intelligence/knowledge/store`

**Purpose:** A queryable store of verified, confidence-tagged, provenance-tracked knowledge records about the world.

**Why Separate from Long-Term Memory:**
- Long-Term Memory: Host-centric (who Host is, preferences, relationships, long-term goals)
- Knowledge Store: World-centric (facts about the world, acquired domain knowledge, verified claims)
- Different confidence models, retention policies, and query semantics

**Knowledge Record:**
```
KnowledgeRecord {
    ID                 string
    Claim              string
    Domain             string              // factual / procedural / contextual / causal
    Source             []SourceRef
    AcquiredAt         time.Time
    VerifiedAt         *time.Time
    VerificationStatus VerificationStatus  // verified / unverified / disputed / revoked
    Confidence         float64             // 0.0 – 1.0
    ExpiresAt          *time.Time
    ConflictsWith      []string
    UsedIn             []string            // Episode IDs for provenance
    StoredBy           string
}
```

**Write Authority:** Knowledge Acquisition, Learning (via Rollout Executor).
**Read Authority:** Reasoning, Planning, Knowledge Acquisition (cross-reference).

---

### SS-13 — Skill Acquisition

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Service | **Package:** `intelligence/skill/acquisition`

**Purpose:** Resolve `SkillGap` signals by acquiring knowledge for a new skill, designing its procedure, validating it, and registering a SkillCard.

**Critical Distinction:**
- Knowledge Acquisition: "I don't know what X is." (factual gap)
- Skill Acquisition: "I know what X is, but I don't know how to do it." (procedural gap)

**Complete Lifecycle:**
```
SkillGap signal received
        ↓
Search Skill Registry: exact match?
  YES → skill already available; this was a routing error → signal error
        ↓
Search Skill Registry: similar/adaptable skill?
  YES → Adaptation path:
        Extract relevant components from existing skill
        Adapt parameters or procedure
        Validate adapted skill → new SkillCard version
        ↓
  NO  → Acquisition path:
        Knowledge acquisition for required procedure
        Procedure design from acquired knowledge
        ↓
Validate procedure:
  Functional test, edge cases, resource bounds
        ↓
Security evaluation:
  Permissions required, risk level, requires confirmation?
  Constitutional pre-approval for high-risk?
        ↓
Sandbox execution:
  Run in isolated environment
  Verify outputs match postconditions
  Verify side effects are declared and bounded
        ↓
Create SkillCard
        ↓
Register in Skill Registry (status: VALIDATED)
        ↓
Constitutional Gate review (if high-risk)
        ↓
Mark AVAILABLE
        ↓
Emit: SkillAcquisitionComplete → Goal Manager
        (Goal Manager schedules Continuation Episode per Decision 0.4 §0.4.L)
```

**Principle:** Acquiring a procedure does not automatically make it trusted or executable. Validation → security evaluation → sandbox → AVAILABLE.

---

### SS-14 — SkillCard Contract

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Contract (typed data structure) | **Location:** `intelligence/types`

**Purpose:** The formal specification for every registered capability. Enables Planning to reason about whether a skill can achieve a goal without executing it.

```
SkillCard {
    // Identity
    SkillID              string
    Name                 string
    Version              string           // semantic version
    Description          string
    Purpose              string

    // Interface Contract
    Inputs               []ParameterSpec
    Outputs              []ParameterSpec
    SideEffects          []SideEffect     // ALL side effects explicitly declared

    // Conditions
    Preconditions        []Condition
    Postconditions       []Condition

    // Procedure
    ProcedureRef         string
    RequiredKnowledge    []string
    RequiredCapabilities []string

    // Dependencies
    DependsOn            []string         // other SkillCard IDs
    IncompatibleWith     []string

    // Authorization
    Permissions          []Permission
    RiskLevel            RiskLevel        // low / medium / high / critical
    RequiresConfirmation bool

    // Quality
    ValidationStatus     ValidationStatus // registered/validated/available/deprecated/revoked
    Confidence           float64
    SuccessRate          float64          // from Learning

    // Examples and Failures
    Examples             []SkillExample
    FailureModes         []FailureMode

    // Provenance
    Provenance           SkillProvenance
    CreatedAt            time.Time
    LastUsed             time.Time

    // History (bounded)
    UsageHistory         []SkillUsage
    EvaluationHistory    []SkillEval
}
```

**Ownership Matrix:**

| Action | Owner |
|:-------|:------|
| Create | Skill Acquisition |
| Validate | Skill Acquisition (validation pipeline) |
| Security review | Constitutional Gate (for high-risk) |
| Register | Skill Registry |
| Select for planning | Planning |
| Execute | Executive / Capability Binder |
| Evaluate | Reflection |
| Update confidence | Learning (via Rollout Executor) |
| Version | Skill Registry |
| Deprecate | Learning (automated) or Host (manual) |
| Revoke | Constitutional Gate or Host |

---

### SS-15 — Skill Registry

**Status:** 🟠 PROPOSED REFINEMENT — capabilities exist; formal SkillCard contract needed
**Classification:** Store | **Package:** `intelligence/skill/registry`

**Purpose:** Authoritative registry of all validated, available skills. Planning queries it to determine what IDUN can do.

**Why Separate from capabilities/ Package:** The `capabilities/` package contains implementations. The Skill Registry contains SkillCard *contracts* — what capabilities can do in terms Planning can reason about. The SkillCard points to the implementation; it is not the implementation.

**Write Authority:** Skill Acquisition, Learning (via Rollout Executor).
**Read Authority:** Planning, Skill Acquisition (similar-skill search), Conversation Planner.

**Query Interface:**
- `FindByGoalType(goal) []SkillCard` — skills that serve a goal type
- `FindSimilar(description) []SkillCard` — structurally similar skills
- `FindAvailable() []SkillCard` — all currently available skills
- `GetByID(id) *SkillCard`

---

### SS-16 — Planning

**Status:** 🟡 CURRENT / INCOMPLETE — V3 frozen; Planning → Reasoning import violation; gap signal emission and Goal Manager integration needed
**Classification:** Service | **Package:** `intelligence/planning` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Decompose active goals into structured, executable step sequences. Produce CandidatePlans for Decision.

**Boundary Issue:** `planning/types.go` imports `idun/intelligence/reasoning` directly. Resolution: both import `intelligence/types` for shared types. This must be resolved before Phase 2.

**Planning Can Pause:** When Planning discovers a gap:
1. Emit structured gap signal
2. Emit `PlanStatusInsufficientInfo` with gap reference
3. Goal Manager checkpoints planning state
4. Acquisition runs independently
5. On acquisition completion: Goal Manager triggers Continuation Episode
6. Planning resumes with updated context

**Owns:** HTN decomposition, GOAP, A*/Beam search, feasibility assessment, resource estimation, gap signal emission.
**Must NOT Own:** Selecting the best plan, authorizing execution, reasoning about user intent, acquiring knowledge.

---

### SS-17 — Decision

**Status:** 🟢 CURRENT — V3 frozen
**Classification:** Service | **Package:** `intelligence/decision` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Evaluate candidate plans and commit to the optimal choice, or defer, abstain, or escalate.

**The Decision Boundary — Permanent:**
```
Generate options   →   Evaluate options   →   Select/Commit   →   Execute
   (Planning)             (Decision)            (Decision)        (Executive)
```
These four stages are permanently distinct. Decision never generates its own options. Planning never authorizes its own plans. Executive never decides what to do.

---

### SS-18 — Executive

**Status:** 🟢 CURRENT — V2 frozen
**Classification:** Service | **Package:** `intelligence/executive` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Content-blind orchestration and dispatch of authorized plans to the execution layer.

**Executive Invariant:** Executive inspects only control-plane envelope metadata. It never dereferences payload content, never parses language, never performs reasoning or planning.

**Pending Wiring (Phase 1):**
- Trigger `reflection.ReflectEpisode` after each meaningful episode closes
- Schedule `Learning.RunCycle` during idle periods
- Schedule `reflection.ReflectPeriodic` on configured idle schedule

---

### SS-19 — Capabilities / Skills Execution Layer

**Status:** 🟢 CURRENT — V3 Binder frozen
**Classification:** Service | **Package:** `capabilities/`

**Purpose:** Execute authorized SkillCards and native capabilities.

**Boundary:** The Capability Binder executes what it is told. It does not evaluate whether to execute, perform authorization, or interpret intent behind requests. Authorization is complete before execution reaches this layer.

---

### SS-20 — Conversation Planner

**Status:** 🔵 APPROVED ARCHITECTURE — new, high priority
**Classification:** Service | **Package:** `intelligence/conversation`

**Purpose:** Determine *what* IDUN should communicate, in what communicative mode, and at what verbosity level. The cognitive bridge between intelligence output and the presentation layer.

**Problem Solved:** Language Realization currently contains conversational policy inside LLM prompts — a cognitive function leaking into a presentation component. The Conversation Planner is the correct home for communicative intent decisions.

**Communicative Intent Types:**

| Type | When Used |
|:-----|:----------|
| `ANSWER` | IDUN knows and can provide the answer |
| `ACKNOWLEDGE` | Confirm understanding or completion |
| `CLARIFY` | Ask Host to resolve ambiguity (ClarificationGap) |
| `WARN` | Alert Host to something important |
| `EXPLAIN` | Provide reasoning or justification |
| `RECOMMEND` | Suggest a course of action |
| `DISAGREE` | Respectfully challenge a premise |
| `APOLOGIZE` | Acknowledge an error |
| `REFUSE` | Decline with reason |
| `NOTIFY` | Proactive information Host should know |
| `ASK_CONFIRMATION` | Request explicit authorization (AuthorizationGap) |
| `PRESENT_OPTIONS` | Offer choices when no single best option exists |
| `REMAIN_SILENT` | Nothing useful to say; do not add noise |

**ConversativeIntent:**
```
ConversativeIntent {
    Type        IntentType
    Content     CognitiveResult     // structured — not yet language
    Tone        Tone                // from Host preferences
    Verbosity   Verbosity           // Concise / Normal / Detailed
    Urgency     float64
    GapDetails  *GapSignal         // if CLARIFY
    Options     []OptionSummary    // if PRESENT_OPTIONS
    Reasoning   *ReasoningSummary  // if EXPLAIN
}
```

**Inputs:** `SalientEvent` (from Attention), Reasoning result, Decision outcome, Execution outcome, Working Memory, Host preferences, Gap signals, Constitution rules.
**Outputs:** `ConversativeIntent` (to Constitutional Gate → Language Realization). It also records its outbound intent into Working Memory (`ConversationTurns`).
**Must NOT Own:** Natural language generation, domain reasoning, action decisions, capability execution, mechanical delivery/queueing.

---

### SS-21 — Language Realization

**Status:** 🟡 CURRENT / INCOMPLETE — exists; needs reorientation
**Classification:** Service | **Package:** `world/` (presentation layer)

**Purpose:** Transform `ConversativeIntent` (after Constitutional Gate approval) into natural language output.

**Current Problem:** Language Realization contains conversational policy logic inside LLM prompts. The cognitive system has no formal input into communicative intent. This is resolved by the Conversation Planner boundary.

**Required Change:** Language Realization must consume `ConversativeIntent` instead of `ExecutionResponse`. Its LLM prompt contains only expression guidance (tone, style, conciseness) — not communicative intent logic.

**Rendering Modes:**

| Output Type | Rendering |
|:------------|:----------|
| Time / date, status, numbers, lists | Deterministic template |
| Conversational response | LLM at low temperature |
| Notification | Concise contextual statement |
| Clarification question | Natural phrasing of gap |

**Must NOT Own:** Communicative intent decisions.

---

### SS-22 — Episodic Memory

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Store | **Package:** `intelligence/memory/episodic`

**Purpose:** Structured records of significant experiences — what happened, why, and what was learned. Used by Reflection for metacognitive audit and by Reasoning for analogical case retrieval.

**Why Separate from Long-Term Memory:**
- Long-Term Memory: persistent facts, preferences, Host relationships
- Episodic Memory: experiences — what happened in specific situations, outcomes, lessons
- Different query pattern: "What happened when I tried to do X?" vs. "What do I know about X?"

**EpisodeRecord:**
```
EpisodeRecord {
    EpisodeID          string
    GoalID             string
    GoalStatement      string
    EpisodeType        EpisodeType    // request / goal_step / background / proactive
    StartedAt          time.Time
    CompletedAt        time.Time
    Outcome            EpisodeOutcome // success / partial / failure / abandoned / vetoed
    Context            EpisodeContext // entities, beliefs, host state at start
    Actions            []ActionRecord
    Observations       []ObservationRecord
    KnowledgeGapsHit   []KnowledgeGapRecord
    SkillGapsHit       []SkillGapRecord
    ReflectionReport   *ReflectionReport
    LessonsLearned     []LessonRecord
}
```

**Episode Begins When:** New goal requires planning, paused goal resumes, autonomous trigger requires cognitive work.
**Episode Does NOT Begin For:** Trivial reflex responses, simple one-step queries.

**Write Authority:** Reflection (episode records), Executive (episode metadata at close).
**Read Authority:** Reasoning (analogy), Reflection (periodic review).

---

### SS-23 — Long-Term Memory

**Status:** 🟡 CURRENT / INCOMPLETE — core/memory exists; taxonomy and lifecycle not designed
**Classification:** Store | **Package:** `intelligence/memory/longterm`

**Purpose:** Persistent cross-session knowledge about the Host, relationships, and learned behaviors. Survives indefinitely.

**Memory Taxonomy:**

| Category | Key Prefix | Retention | Expiration |
|:---------|:-----------|:----------|:-----------|
| Host identity | `host.identity.*` | Very High | Never |
| Host preferences | `host.preferences.*` | High | On explicit change |
| Host relationships | `host.relationships.*` | High | On correction |
| Long-term goals | `host.goals.*` | High | On completion |
| Host corrections | `host.corrections.*` | Very High | Never |
| Project context | `host.projects.*` | High | On completion |
| Observed patterns | `host.patterns.*` | Medium | On contradiction |
| Interaction preferences | `interaction.*` | Medium | On correction |

**Retention Policy:**
```
Explicitly requested by Host?   → Store immediately, HIGH confidence
Is a correction?                → Store immediately, VERY HIGH; supersede conflicts
Used in N episodes?             → Promote (configurable threshold)
Reflection recommended?         → Store with Reflection confidence
Otherwise                       → Discard from Working Memory
```

**Conflict Handling:** Conflicts never silently overwritten. New records carry `conflicts_with` link. Host confirmation resolves. Nothing permanently deleted — records archived, not erased.

---

### SS-24 — Event Memory

**Status:** 🟡 CURRENT / INCOMPLETE — core/scheduler exists; formal Event Memory type not designed
**Classification:** Store | **Package:** `intelligence/memory/events`

**Purpose:** Time-anchored records that trigger future action or notification.

| Type | Trigger Condition |
|:-----|:------------------|
| `reminder` | Scheduled time |
| `deadline` | T-minus threshold |
| `appointment` | Scheduled time + pre-alert |
| `goal_checkpoint` | Date or condition |
| `scheduled_action` | Scheduled time |
| `recurring` | Cron expression |
| `condition_trigger` | Condition evaluation |

All event triggers flow through Event Router → Relevance Check → Autonomy Policy before reaching the cognitive pipeline. Events are cognitively evaluated, not forwarded raw.

---

### SS-25 — Reflection

**Status:** 🟡 CURRENT / INCOMPLETE — fully implemented; not wired in production
**Classification:** Service | **Package:** `intelligence/reflection` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Metacognitive evaluation of episodes and longitudinal trends. Produce structured findings for Learning.

**The Reflection Invariant:** Reflection is a pure read-only evaluator. It reads episode traces from Workspace (immutable). Its sole output is `ReflectionReport`. It modifies nothing.

**Operating Modes:**

| Mode | Trigger | Scope |
|:-----|:--------|:------|
| `MODE_EPISODE` | Executive after each meaningful episode closes | Single episode trace |
| `MODE_PERIODIC` | Scheduler during idle windows | Recent N episodes via `HistoricalSummary` |

**11 Specialist Evaluators:** Understanding, Reasoning, Decision, Planning, Learning, Attention, Conversation, Executive, Overall Cognitive Assessment, Trend Reflection, Reflection-on-Reflection.

**Pending Wiring:**
- Executive must trigger `MODE_EPISODE` after meaningful episode close (Phase 1)
- Scheduler must trigger `MODE_PERIODIC` on configured schedule (Phase 1)

---

### SS-26 — Learning

**Status:** 🟡 CURRENT / INCOMPLETE — fully implemented; not active in production
**Classification:** Service | **Package:** `intelligence/learning` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Synthesize improvements to cognitive strategies from accumulated Reflection experience. Produce `CandidateSnapshot` objects for Rollout Executor activation.

**The Learning Invariant:** Learning never writes directly to any active cognitive subsystem. All outputs are `CandidateSnapshot` objects activated through `Draft → Validated → Shadow → Canary → Active`.

**Constitutional Hard Limit:** No learning signal can produce a snapshot that causes IDUN to bypass constitutional constraints or act outside authorized permission boundaries.

**Typed Learning Outputs (not a generic "learning update"):**

| Domain | Snapshot Schema | Recipient |
|:-------|:----------------|:---------|
| Reasoning S1 rules | `idun.reasoning.strategy.v1` | Reasoning symbolic engine |
| Planning strategy | `idun.planning.strategy.v1` | Planning depth selector |
| Decision policy | `idun.decision.policy.v1` | Decision utility weights |
| Attention weights | `idun.attention.weights.v1` | Attention salience scorer |
| Conversation preferences | `idun.conversation.prefs.v1` | Conversation Planner |
| Research strategies | `idun.knowledge.strategy.v1` | Knowledge Acquisition |

**Pending Wiring:**
- Learning must be triggered by `TopicReflectionReports` (Phase 1)
- Rollout Executor must be connected to SnapshotRegistry (Phase 1)
- Learning cycles scheduled during idle periods via Scheduler (Phase 1)

---

### SS-27 — Autonomy Policy

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Policy (evaluated by Event Router) | **Location:** Embedded in `runtime/eventrouter`

**Purpose:** Determine what IDUN is permitted to do in response to a non-Host event, based on configured autonomy level and event nature.

**Why NOT a Separate Package:** Autonomy Policy is a lookup evaluation — event type × autonomy level → permitted action. No goroutines, no lifecycle, no state beyond current level configuration. Principle P12 applies.

**Autonomy Levels:**

| Level | Name | Permitted Without Asking |
|:------|:-----|:------------------------|
| 0 | Passive | Nothing — waits for Host input only |
| 1 | Informational | Proactive notifications (reminders, alerts, important changes) |
| 2 | Preparatory | Background research; no consequential action |
| 3 | Low-risk autonomous | Pre-approved reversible actions (create notes, set reminders) |
| 4 | Goal-directed | Execute approved multi-step goal plans |
| 5 | High-impact | NEVER automatic — always requires explicit per-action Host approval |

**Key Distinction:**
```
AUTONOMY LEVEL    governs what IDUN may do without asking
CONSTITUTION      governs what IDUN may NEVER do regardless of autonomy level
```
These are orthogonal controls. Higher autonomy does not weaken the Constitution.

---

### SS-28 — Notification Service

**Status:** 🔵 APPROVED ARCHITECTURE
**Classification:** Service | **Package:** `runtime/notification`

**Purpose:** Deliver `ConversativeIntent` of type `NOTIFY` to the Host through the appropriate modality.

**Notification Service Owns:** Mechanical queueing, retry, rate limiting, mechanical deduplication, modality selection, and delivery.
**Notification Service Does NOT Own:** What to communicate (Conversation Planner), whether to communicate (Conversation Planner), or semantic batching (Attention).

**Notification Decision Flow:**
```
Notification Service receives Rendered Message (via Language Realization) →
Mechanical Deduplication (has exact message been sent recently?) →
Queue/Rate Limit (wait for appropriate delivery timing based on urgency) →
Select Modality → Deliver to Host
```

**Modalities:**

| Modality | Status |
|:---------|:-------|
| Terminal output | 🟢 Current |
| Desktop notification | 🔵 Planned |
| Voice | ⚪ North Star Future |
| Mobile push | ⚪ North Star Future |

---

### SS-29 — Scheduler

**Status:** 🟢 CURRENT — `core/scheduler` exists
**Classification:** Service | **Package:** `core/scheduler`

**Purpose:** Fire scheduled events and manage recurring schedules. The time-based event source.

**Must NOT Own:** Deciding whether an event matters, cognitive processing.

**Extended Responsibilities (Phase 1 wiring):**
- Trigger periodic Reflection cycles
- Trigger Learning cycles during idle periods

---

### SS-30 — Constitution

**Status:** 🟢 CURRENT — fully implemented and frozen
**Classification:** Service | **Package:** `intelligence/constitution` | **Frozen:** `2.0.0-FROZEN`

**Purpose:** Non-negotiable hard boundary. All external world-modifying actions — including proactive communication via `ConversativeIntent` — must pass through the Constitutional Gate before execution. HMAC-SHA256 approval tokens required.

**Constitutional Rule — Absolute:** The constitution is not advisory. No learned behavior, optimization pressure, or autonomy level can cause IDUN to bypass it. If the Constitutional Gate is unavailable, the default is to **not execute** — never to proceed anyway.

---

### SS-31 — Calibration

**Status:** 🟢 CURRENT — fully implemented
**Classification:** Service | **Package:** `intelligence/calibration`

**Purpose:** Apply cross-episode historical trust adjustment to Reasoning confidence. Sole writer of `CalibratedConfidence`.

---

### SS-32 — Permissions and Runtime Policy

**Status:** 🟡 CURRENT / INCOMPLETE — basic permissions exist; no runtime policy engine
**Classification:** Policy store | **Package:** `intelligence/infrastructure/permissions`

**Purpose:** Maintain what IDUN is authorized to do on behalf of the Host, and what operational policies govern its behavior.

---

## Part 4 — Memory Architecture

### 4.1 Why Five Separate Stores

| Store | Purpose | Lifetime | Key Distinction |
|:------|:--------|:---------|:----------------|
| **Working Memory** | Active cognitive context | Session / Episode | Bounded, ephemeral, high-frequency writes |
| **Long-Term Memory** | Host and relationship knowledge | Indefinite | Host-centric; selective retention |
| **Episodic Memory** | Experience records | Indefinite (summarized) | Experience-centric; retrospective query |
| **Event Memory** | Time-anchored triggers | Until fired or cancelled | Time-centric; drives autonomous activation |
| **Knowledge Store** | Verified world knowledge | Until expired or revoked | World-centric; confidence and provenance |

Collapsing these into a single generic `Memory` table would make it impossible to apply different retention policies, different expiration strategies, and different query semantics per type.

### 4.2 Memory Lifecycle

```
New information observed during episode
        │
        ▼ [WRITE]
Working Memory (active, bounded)
        │
        ▼ [at episode or session end]
Consolidation Decision:
  Host explicitly requested retention? → Long-Term Memory (HIGH confidence)
  Was a correction?                    → Long-Term Memory (VERY HIGH confidence)
  Used N times in N episodes?          → Long-Term Memory
  Reflection recommended?              → Long-Term Memory (with confidence)
  Otherwise                            → Discard
        │
        ▼
Long-Term Memory (persistent)
        │
  Relevant to new episode → Retrieved → Working Memory (prefetch)
  Not accessed for N episodes → Flagged as Stale
  Contradicted by newer record → Disputed (not deleted)
  Host invalidates → Revoked
  Periodic Reflection confirms no longer relevant → Archived
```

---

## Part 5 — Gap Architecture

### 5.1 Four Gap Types — Formally Distinguished

| Gap Type | Meaning | Resolution Path |
|:---------|:--------|:----------------|
| `KnowledgeGap` | I don't know what this is | Knowledge Acquisition |
| `SkillGap` | I know what needs doing but not how | Skill Acquisition |
| `ClarificationGap` | I don't know what the Host means | Conversation Planner → Host |
| `AuthorizationGap` | I know how but I am not authorized | Conversation Planner → Host |

**Examples:**
```
"What is the GDP of Singapore?"
    → KnowledgeGap: factual gap → internet research

"Deploy this app using Kubernetes."
    → SkillGap: procedural gap → Skill Acquisition

"Do that thing we talked about before."
    → ClarificationGap: ambiguous reference → ask Host

"Send this email to John."
    → AuthorizationGap: email sending requires per-send authorization → ask Host
```

---

## Part 6 — Goal and Episode Architecture

### 6.1 Three-Level Temporal Model

```
GOAL                "Prepare for Singapore move"
Long-lived intention. Managed by Goal Manager.
Persists across days/weeks. Spawns multiple episodes.
        │
        ├── EPISODE 1    "Research Singapore visa requirements"
        │   Bounded cognitive work session.
        │   Has a start, end, outcome, and reflection record.
        │   Spawns multiple turns.
        │           │
        │           ├── TURN   "What is the processing time?"
        │           │   Single request-response cycle.
        │           │   May or may not create an episode.
        │           │   Always produces a response.
        │
        ├── EPISODE 2    "Compare accommodation options"
        ...
```

### 6.2 When to Create an Episode

| Scenario | Create Episode? |
|:---------|:---------------|
| Simple factual query ("What time is it?") | No — trivial turn |
| Conversational acknowledgment | No — trivial turn |
| Simple calculator query | No — trivial turn |
| Request requiring multi-step planning | Yes |
| Request requiring knowledge acquisition | Yes |
| Request requiring skill acquisition | Yes |
| Autonomous trigger requiring cognitive work | Yes |
| Background research for active goal | Yes |

---

## Part 7 — Communication Architecture

### 7.1 The Boundary

All communication — both reactive (responding to Host) and proactive (IDUN-initiated) — flows through the same canonical pipeline. The boundary between cognitive decisions and mechanical delivery is enforced by `ConversativeIntent`.

```text
═══════════════════════════════════════════════════════════════
  CANONICAL PROACTIVE COMMUNICATION FLOW
  (Decision 0.3 — Architecture Resolved / Safe to Freeze)
═══════════════════════════════════════════════════════════════

  LEGEND
  ──────────►  data / event flow (synchronous)
  - - - - - ►  asynchronous / event flow
  READ ►       read dependency on a store or context
  WRITE ►      write authority over a store or context
  [STORE]      persistent storage
  [WM]         Working Memory slot
  [EVENT]      event transport (content-blind)
  [CONTRACT]   typed contract in intelligence/types
  ⚠            failure / degraded path

═══════════════════════════════════════════════════════════════

  EVENT PRODUCERS
  ┌──────────────────────────────────────┐
  │  Executive / WorkflowCoordinator     │  (progress, lifecycle events)
  │  Goal Manager                        │  (goal lifecycle events)
  │  Scheduler                           │  (temporal triggers)
  └───────────────────┬──────────────────┘
                      │
                      │ produces: SystemEvent [CONTRACT]
                      │   fields: EventID, SourceSubsystem,
                      │           EventType, GoalID?, EpisodeID?,
                      │           Payload, Timestamp
                      │ lifecycle: depends on source (TURN/EPISODE/GOAL)
                      │ persistence: TRANSIENT — event transport only
                      ▼
  ┌─────────────────────────────────────────────────┐
  │  WORKSPACE EVENT BUS                            │
  │  (content-blind transport — Decision 0.1 §0.1.D)│
  │  Owner: Runtime infrastructure                  │
  └───────────────────┬─────────────────────────────┘
                      │
                      ▼
  ┌─────────────────────────────────────────────────┐
  │  EVENT ROUTER (SS-02 / SS-27)                   │
  │  Owner: Runtime                                 │
  │  Evaluates: Autonomy Policy (SS-27)             │
  │    Is IDUN permitted to act on this event?      │
  │  Failure: ⚠ Drop with log                       │
  └───────────────────┬─────────────────────────────┘
                      │
                      │ passes: SystemEvent (if permitted)
                      ▼
  ┌─────────────────────────────────────────────────┐
  │  ATTENTION (SS-03)                              │
  │  Owner: Cognitive (intelligence/attention)      │
  │                                                 │
  │  READ ► Working Memory (SS-06)                  │
  │    - ActiveGoals (GOAL scope)                   │
  │    - CurrentEntities (EPISODE scope)            │
  │    - ActiveTopic (TURN scope)                   │
  │                                                 │
  │  Evaluates: salience, urgency, relevance         │
  │  Performs: semantic grouping of related events  │
  │  Assigns: priority band (0–4)                   │
  │  Failure: ⚠ fall back to static band scoring;  │
  │           log WM unavailability                 │
  └───────────────────┬─────────────────────────────┘
                      │
                      │ produces: SalientEvent [CONTRACT]
                      │   fields: original SystemEvent,
                      │           SalienceScore, UrgencyBand,
                      │           RelevanceReason, SemanticGroupID?
                      │ persistence: TRANSIENT
                      ▼
  ┌─────────────────────────────────────────────────┐
  │  CONVERSATION PLANNER (SS-20)                   │
  │  Owner: Cognitive (intelligence/conversation)   │
  │                                                 │
  │  READ ► Working Memory (SS-06)                  │
  │    - ConversationTurns (TURN scope)             │
  │    - ActiveTopic, HostPreferences               │
  │    - ReasoningResult, DecisionRecord            │
  │    - GapSignals                                 │
  │                                                 │
  │  Decides: IntentType, tone, verbosity, urgency  │
  │                                                 │
  │  WRITE ► Working Memory (SS-06) via WMM         │
  │    - ConversationTurns [WM] (TURN scope)        │
  │      so Understanding can resolve a Host reply  │
  │                                                 │
  │  Must NOT own: delivery timing, dedup, queueing │
  └───────────────────┬─────────────────────────────┘
                      │
                      │ produces: ConversativeIntent [CONTRACT]
                      │   fields: Type (IntentType), Content,
                      │           Tone, Verbosity, Urgency,
                      │           GapDetails?, Options?, Reasoning?
                      │ persistence: TRANSIENT
                      ▼
  ┌─────────────────────────────────────────────────┐
  │  CONSTITUTIONAL GATE (SS-30)                    │
  │  Owner: Policy (intelligence/constitution)      │
  │  Applies to: ALL ConversativeIntent             │
  │    including proactive NOTIFY                   │
  │  Failure: ⚠ Hard stop — intent dropped;        │
  │           never fail open                       │
  └───────────────────┬─────────────────────────────┘
                      │
                      │ passes: ConversativeIntent (approved)
                      ▼
  ┌─────────────────────────────────────────────────┐
  │  LANGUAGE REALIZATION (SS-21)                   │
  │  Owner: Presentation (world/)                   │
  │  Input: ConversativeIntent (structured)         │
  │  Renders: natural language per rendering mode   │
  │  Must NOT own: communicative intent decisions   │
  │  Failure: ⚠ OPEN — FAILURE CONTRACT REQUIRED   │
  └───────────────────┬─────────────────────────────┘
                      │
                      │ produces: Rendered Message
                      │ persistence: TRANSIENT
                      ▼
  ┌─────────────────────────────────────────────────┐
  │  NOTIFICATION SERVICE (SS-28)                   │
  │  Owner: Presentation (runtime/notification)     │
  │  Mechanical delivery only:                      │
  │    queue, retry, rate limit,                    │
  │    mechanical deduplication,                    │
  │    delivery ordering, modality/channel          │
  │  Must NOT own: semantic reasoning,              │
  │    communicative strategy, intent decisions     │
  │  Failure: ⚠ queue with bounded retry;          │
  │           drop if queue full; log               │
  └───────────────────┬─────────────────────────────┘
                      │
                      ▼
                     HOST
```

This boundary must never be violated. Communicative policy belongs to Conversation Planner. Linguistic expression belongs to Language Realization. Mechanical delivery belongs to Notification Service.

### 7.2 Mid-Episode Communication

IDUN may communicate important information while a long-running Episode is still executing. This uses the same proactive communication pipeline above — no separate lifecycle is needed.

```text
  ACTIVE GOAL (lifecycle: GOAL)
       │
  RUNNING EPISODE (lifecycle: EPISODE)
       │
       │ Executive/WorkflowCoordinator emits
       ├─►  SystemEvent(type=Progress, GoalID=X, EpisodeID=Y)
       │         │
       │         ▼  [normal proactive pipeline]
       │    Event Bus → Event Router → Attention
       │         │  READ WM (ActiveGoals, CurrentEntities)
       │         ▼
       │    SalientEvent
       │         ▼
       │    Conversation Planner
       │         │  WRITE WM (ConversationTurns) via WMM
       │         ▼
       │    ConversativeIntent
       │         ▼
       │    Constitutional Gate
       │         ▼
       │    Language Realization
       │         ▼
       │    Notification Service → HOST
       │
       │  ← The original Episode CONTINUES here.
       │    The proactive communication is a side-effect;
       │    it does NOT create a new Episode.
       │
       │  ← If the Host responds:
       │    (new Turn enters normal cognitive pipeline)
       │    Understanding READS ConversationTurns from WM
       │    Context Resolver resolves references
       │    Reasoning / Planning / Goal Manager handle
       │    any goal modification or continuation
       ▼
  EPISODE CONTINUES or PAUSES/CANCELS (Goal Manager decides)
```

**Critical Rule:** A proactive communication mid-episode does NOT create another Episode. The Episode lifecycle is managed by Goal Manager. A Host response to a proactive notification enters the standard Turn → Understanding → Goal/Episode update lifecycle.

### 7.3 Responsibility Boundary

```text
┌─────────────────────────────────────────────────────────────┐
│                    COGNITIVE DOMAIN                         │
│                                                             │
│  Event Router (SS-02 / SS-27)                               │
│    → autonomy / routing policy                              │
│    → must not inspect event semantics                       │
│                                                             │
│  Attention (SS-03)                                          │
│    → salience / urgency / contextual relevance              │
│    → semantic grouping of related events                    │
│    → reads Working Memory for context                       │
│                                                             │
│  Conversation Planner (SS-20)                               │
│    → communication strategy (what to say)                   │
│    → writes ConversationTurns to Working Memory             │
│    → must not own delivery timing or dedup                  │
│                                                             │
│  Constitutional Gate (SS-30)                                │
│    → authorization / safety (for ALL ConversativeIntent)    │
└──────────────────────────┬──────────────────────────────────┘
                           │ ConversativeIntent
═══════════════════════════╪═══════════════════════════════════
                           │
┌─────────────────────────────────────────────────────────────┐
│                REALIZATION / DELIVERY DOMAIN                │
│                                                             │
│  Language Realization (SS-21)                               │
│    → natural language rendering                             │
│    → must not contain communicative intent logic            │
│                                                             │
│  Notification Service (SS-28)                               │
│    → queue / retry / rate limit / mechanical dedup          │
│    → delivery channel / modality selection                  │
│    → must not perform semantic reasoning                    │
└─────────────────────────────────────────────────────────────┘
```

### 7.4 Semantic vs Mechanical Batching

**Semantic grouping** is cognitive and belongs to Attention. **Mechanical deduplication and batching** is delivery infrastructure and belongs to Notification Service.

```text
  SEMANTIC GROUPING (Attention — cognitive):

  Multiple related SystemEvents
       │
       ▼
  Attention (SS-03)
       │ recognizes relationship; groups into single SalientEvent
       ▼
  Conversation Planner
       │ generates one combined ConversativeIntent
       ▼
  Constitutional Gate → Language Realization → Notification Service → Host

  ─────────────────────────────────────────────────────────────

  MECHANICAL DELIVERY (Notification Service — delivery infrastructure):

  Rendered Messages
       │
       ▼
  Notification Service (SS-28)
       ├── queue                (order messages)
       ├── retry                (resend on transient failure)
       ├── rate limit           (prevent flooding, runtime policy)
       ├── mechanical dedup     (suppress byte-identical messages)
       └── modality/channel     (terminal, desktop, voice, push)
       │
       ▼
  Host

  Notification Service has NO knowledge of semantic relationships.
  Semantic grouping has already occurred upstream in Attention.
```

### 7.5 Conversational Learning — Constitutional Boundary

| MAY be learned | MAY NEVER be learned |
|:--------------|:--------------------|
| Response verbosity preference | Safety constraints |
| Technical vs. simplified framing | Truthfulness requirements |
| Formality level | Authorization requirements |
| When to volunteer information | Constitutional boundaries |
| How to present options | Deception of any form |
| Tone calibration | Suppression or hiding of errors |

---

## Part 8 — Architectural Invariants

These invariants must never be violated by any implementation. Violations require explicit architectural review.

| ID | Invariant |
|:---|:----------|
| **AI-01** | Reasoning does not execute actions. |
| **AI-02** | Planning does not make final authorization decisions. |
| **AI-03** | Decision does not generate candidate plans. |
| **AI-04** | Executive does not interpret content, reason about intent, or make semantic decisions. |
| **AI-05** | Language Realization does not contain communicative intent logic. It renders; it does not decide. |
| **AI-06** | Conversation Planner does not perform domain reasoning. It decides what to say; Reasoning determines what is true. |
| **AI-07** | Memory stores do not silently rewrite or override active cognitive state. All modifications follow defined write authority. |
| **AI-08** | Learning cannot produce CandidateSnapshots that, when activated, would cause IDUN to bypass Constitutional constraints. |
| **AI-09** | Reflection does not modify any subsystem. It produces `ReflectionReport`. That is its complete authority. |
| **AI-10** | Acquired knowledge is not automatically trusted. Every acquired record carries verification status and confidence. |
| **AI-11** | Acquired skills are not automatically executable. Validation → security review → sandbox → AVAILABLE. |
| **AI-12** | Autonomy never bypasses authorization. Higher autonomy expands operational scope; it does not remove constitutional constraints. |
| **AI-13** | Working Memory is bounded. When capacity is exceeded, lowest-priority items are evicted. Never unbounded. |
| **AI-14** | Long-running goals are not single infinite episodes. Every goal is served by bounded episodes with checkpoints. |
| **AI-15** | Learning never invokes itself recursively. It reads ReflectionReport, produces CandidateSnapshot, and terminates. |
| **AI-16** | Planning cannot import Reasoning types directly. Shared types live in `intelligence/types`. |
| **AI-17** | Context Resolver does not mutate `UnderstandingBatch`. It produces `ResolvedContext` as a distinct wrapper type. |
| **AI-18** | Knowledge Acquisition does not reason over acquired knowledge. It acquires, normalizes, evaluates, and stores. Reasoning interprets. |
| **AI-19** | Notification Service does not perform semantic reasoning. It performs mechanical delivery operations only. Semantic grouping of related events belongs to Attention. |
| **AI-20** | Conversation Planner records every proactive outbound `ConversativeIntent` to Working Memory (`ConversationTurns`) before Language Realization, so that an immediate Host response can be correctly resolved by Understanding. |

---

## Part 9 — Failure and Degraded Operation

### 9.1 Per-Subsystem Failure Response

| Subsystem | Failure Response | Degrades? |
|:----------|:----------------|:----------|
| Understanding (deliberative LLM unavailable) | Fall back to grammar+neural; lower confidence | Yes — reduced intent coverage |
| Reasoning (LLM unavailable) | Fall back to symbolic+graph; emit low confidence | Yes — reduced reasoning depth |
| Knowledge Acquisition (internet unavailable) | Local sources only; defer or ask Host | Yes — reduced knowledge freshness |
| Skill Acquisition validation fails | Do not register skill; inform Host | Yes — capability unavailable |
| Working Memory at capacity | Evict lowest-priority items; log eviction | Graceful |
| Attention unavailable | Fall back to static priority band scoring; log degradation | Yes — reduced salience quality |
| Goal Manager unavailable | Complete current episode; no new goals started | Yes — goal tracking suspended |
| Constitutional Gate unavailable | Hard stop — no world-modifying action executed; no ConversativeIntent delivered | Critical — never fail open |
| Notification Service unavailable | Queue Rendered Messages with bounded retry; drop if queue full; log | Yes — delayed/lost proactive notifications |
| Reflection unavailable | Episode closes without reflection; no learning this cycle | Yes — learning paused |
| Learning unavailable | No strategy update; operates at current version | Yes — no improvement |
| Long-Term Memory unavailable | Operate from Working Memory only; stateless session | Yes — reduced context |

### 9.2 The Safe Failure Principle

> **The Constitutional Gate must never fail open.**

If the authorization mechanism for a world-modifying action is unavailable or uncertain, the default is to **not execute** — never to proceed. Safety is the last defense, not an optional optimization.

---

## Part 10 — Implementation Dependency Graph

```
TIER 0 — Foundation (must exist; all present)
─────────────────────────────────────────────
  core/storage ✅
  core/memory  ✅
  core/scheduler ✅

  intelligence/types  🆕 NEW FIRST
    Defines: GapSignal types, GoalProposal, ConversativeIntent, SystemEvent, SalientEvent,
             WorkingMemorySlice contract, EpisodeRecord, SkillCard
    Why first: violates AI-16 if skipped; everything imports from here

          ↓
TIER 1 — Wire Existing Architecture (no new packages required)
──────────────────────────────────────────────────────────────
  Understanding: inject InferenceService → Deliberative LLM active
  Reasoning:     inject InferenceService → Deliberative Specialist active
  Attention:     subscribe to workspace topics in runtime/host.go
  Reflection:    Executive triggers MODE_EPISODE after episode close
  Learning:      triggered by TopicReflectionReports; Rollout Executor connected
  Scheduler:     triggers Reflection periodic + Learning idle cycles

          ↓
TIER 2 — Working Memory (everything cognitive depends on this)
──────────────────────────────────────────────────────────────
  intelligence/workingmemory
    Write contracts per slot
    Capacity, priority, eviction rules
    Session boundary and consolidation logic
    Conflict detection and ContradictionFlag
    Integration: Understanding writes, Reasoning reads, Planning reads

          ↓
TIER 3 — Conversation Planner
──────────────────────────────
  intelligence/conversation
    Requires: Working Memory, intelligence/types
    ConversativeIntent type established
    Language Realization reoriented to consume ConversativeIntent

          ↓
TIER 4 — Goal Manager
──────────────────────
  intelligence/goalmanager
    Requires: Working Memory, intelligence/types, core/storage
    Goal lifecycle state machine
    Continuation protocol (pause/resume via continuation episode)
    Multi-goal priority management

          ↓
TIER 5 — Knowledge Gap Resolution
───────────────────────────────────
  intelligence/knowledge/store
    Requires: core/memory, intelligence/types

  intelligence/knowledge/acquisition
    Requires: Knowledge Store, Working Memory, Goal Manager
    Local sources → internet research (online mode)
    Source evaluation framework

  intelligence/skill/registry
    Requires: SkillCard from intelligence/types
    Formalizes existing capabilities/ as SkillCards

  intelligence/skill/acquisition
    Requires: Skill Registry, Knowledge Acquisition, Goal Manager
    Validation pipeline, security evaluation, sandbox

          ↓
TIER 6 — Episodic Memory + Memory Lifecycle
─────────────────────────────────────────────
  intelligence/memory/episodic
    Requires: Reflection (writes episode records), core/memory

  intelligence/memory/longterm  (wraps core/memory with taxonomy)
    Requires: Working Memory (consolidation), Reflection

  Working Memory → Long-Term Memory consolidation
  Memory expiration and staleness detection
  Periodic Reflection for staleness flagging

          ↓
TIER 7 — Autonomous Operation
───────────────────────────────
  runtime/eventrouter
    Requires: Autonomy Policy (embedded), Event Memory, Attention

  runtime/notification
    Requires: Conversation Planner, World output adapters

  Autonomy Levels 1–2 (informational + preparatory)
  Goal Monitor (condition-based triggers → Event Router)
  Scheduler: periodic triggers fully wired
```

---

## Part 11 — Implementation Phases with Exit Criteria

### Phase 1 — Wire Existing Architecture

**Objective:** Activate fully-implemented subsystems not yet connected to production.

**Dependencies:** None — existing codebase only.

**Scope:**
1. Inject `InferenceService` into `understanding.NewService` in `runtime/host.go`
2. Inject `InferenceService` into `reasoning.NewService` in `runtime/host.go`
3. Instantiate and wire `attention.NewService` to workspace topics in `runtime/host.go`
4. Wire Executive to trigger `reflection.ReflectEpisode` after each meaningful episode closes
5. Wire `TopicReflectionReports` → Learning input
6. Wire Rollout Executor to Learning SnapshotRegistry
7. Wire Scheduler to trigger `reflection.ReflectPeriodic` on idle schedule
8. Wire Scheduler to trigger `Learning.RunCycle` during idle periods

**Non-Goals:** No new packages. No new type contracts. No architectural changes.

**Exit Criteria:**
- `TestDeliberativeWorker` passes with live InferenceService (not mocked)
- `TestDeliberativeSpecialist` passes with live InferenceService (not mocked)
- Attention publishes TelemetrySnapshot showing salience scoring on real workspace traffic
- Reflection produces `ReflectionReport` after each cognitive episode closes (validated by integration test)
- Learning produces `CandidateSnapshot` from Reflection reports
- Rollout Executor activates at least one CandidateSnapshot through Shadow → Active
- No `-race` violations introduced
- No regressions in any existing acceptance tests

---

### Phase 2 — Shared Contracts and Working Memory

**Objective:** Establish the shared type contract package and Working Memory store.

**Dependencies:** Phase 1 complete.

**Scope:**
1. Create `intelligence/types`:
   - `GapSignal` types
   - `GoalProposal`
   - `ConversativeIntent`
   - `WorkingMemorySlice` (read contract for subsystems)
   - `EpisodeRecord`
   - `SkillCard`
2. Resolve Planning → Reasoning import violation: both import `intelligence/types`
3. Implement `intelligence/workingmemory`:
   - `WorkingMemoryStore` (typed in-process slots)
   - `WorkingMemoryManager` (capacity, eviction, expiration, conflict detection)
   - Compile-time-enforced write contracts per slot
4. Wire Understanding → Working Memory writes (context and entities)
5. Wire Context Resolver → Working Memory (topic write)
6. Wire Reasoning → Working Memory reads (context input)
7. Wire Planning → Working Memory reads (context input)
8. Implement Working Memory checkpoint (episode pause/resume)
9. Extend Reasoning to emit `GoalProposal` and `GapSignals`

**Non-Goals:** No Knowledge Acquisition, Skill Acquisition, Goal Manager.

**Exit Criteria:**
- `planning` package has zero direct imports from `reasoning` package
- Working Memory contains conversation context after Understanding processes input
- Reasoning reads working memory context (active goal, recent observations) as part of S0 ContextAssembler
- Planning reads entity context and constraints from Working Memory
- Working Memory capacity eviction fires correctly at boundary (tested with synthetic data)
- Working Memory checkpoint written on episode pause; fully restored on resume
- All Working Memory writes logged with subsystem attribution (audit trail)
- No `-race` violations
- Reasoning emits GoalProposal and GapSignals (verified in integration test with all gap types)

---

### Phase 3 — Conversation Planner

**Objective:** Separate communicative intent from language rendering.

**Dependencies:** Phase 2 (Working Memory, `intelligence/types`).

**Scope:**
1. Implement `intelligence/conversation.ConversationPlanner`:
   - All 13 IntentType strategies
   - Reads Reasoning result, Decision outcome, Execution outcome, Working Memory
   - Produces `ConversativeIntent`
2. Reorient Language Realization to consume `ConversativeIntent`
3. Remove embedded conversational policy from Language Realization LLM prompts
4. Wire ClarificationGap signals → Conversation Planner → Language Realization → Host
5. Wire AuthorizationGap signals → Conversation Planner → Language Realization → Host

**Non-Goals:** No learned conversation preferences yet. Default behavior only.

**Exit Criteria:**
- Language Realization prompt contains only expression guidance, no intent logic
- Every response trace shows a `ConversativeIntent` with correct `IntentType`
- `CLARIFY` intent generates a well-formed, natural clarification question (not a generic error)
- `AUTHORIZATION_GAP` generates a well-formed authorization request with clear reason
- `REMAIN_SILENT` produces no output and no error
- Existing acceptance tests produce equivalent or improved outputs
- No response produced without a ConversativeIntent trace record

---

### Phase 4 — Goal Manager

**Objective:** Enable multi-goal lifecycle management and pause/resume for gap resolution.

**Dependencies:** Phase 2 (Working Memory), Phase 3 (Conversation Planner for gap notifications to Host).

**Scope:**
1. Implement `intelligence/goalmanager`:
   - Goal registry with `core/storage` persistence
   - Goal state machine
   - Continuation protocol (pause goal → acquisition → continuation episode)
   - Multi-goal priority management
   - Sub-goal support
   - Goal deadline monitoring
2. Extend Reasoning: `GoalProposal` → Goal Manager (not directly to `TopicActiveGoals`)
3. Wire Goal Manager → Planning (Goal Manager provides active goals)
4. Implement Goal Monitor (deadline proximity triggers notification)

**Non-Goals:** No Knowledge/Skill Acquisition yet. Goal Manager receives gap signals but cannot resolve them.

**Exit Criteria:**
- Multiple goals can be active simultaneously; correct priority ordering enforced
- A goal paused due to gap signal resumes correctly with restored context after simulated acquisition completion
- Completed goals create EpisodeRecord skeleton (Reflection fills in details)
- Goal deadline monitoring fires notification at T-minus threshold
- Goal state persists across IDUN process restart
- No race conditions in multi-goal concurrent access under `-race`

---

### Phase 5 — Knowledge Gap Resolution

**Objective:** Enable IDUN to recognize when it doesn't know something and acquire the needed knowledge.

**Dependencies:** Phase 4 (Goal Manager for continuation episodes), Phase 2 (Working Memory).

**Scope:**
1. Create `intelligence/knowledge/store`
2. Create `intelligence/knowledge/acquisition`:
   - Local strategy (Knowledge Store, Working Memory, Episodic Memory analogy)
   - Source evaluation framework
   - Cross-reference check
   - Confidence scoring
   - Online strategy (internet research, gated by online mode flag)
3. Wire KnowledgeGap → Acquisition → Goal Manager (continuation)
4. Wire Acquisition → Working Memory (temporary research)
5. Wire Acquisition → Knowledge Store (verified records)

**Non-Goals:** No Skill Acquisition. No Skill Gap resolution.

**Exit Criteria:**
- Scenario: IDUN asked question beyond current knowledge → acquires answer → answers correctly (with source citation)
- Scenario: Internet unavailable → acknowledges limitation → defers or asks Host → does NOT fabricate
- Scenario: Acquired information conflicts with existing Knowledge Store record → conflict stored; Host notified; no silent overwrite
- All Knowledge Store records carry: source, confidence, verification status, timestamp
- Provenance trail: every knowledge record traceable to the acquisition operation and the episode it served

---

### Phase 6 — Skill Gap Resolution

**Objective:** Enable IDUN to recognize when it cannot do something and acquire the needed skill.

**Dependencies:** Phase 5 (Knowledge Acquisition — skills require knowledge), Phase 4 (Goal Manager).

**Scope:**
1. Formalize `intelligence/skill/registry` from existing `capabilities/`
2. Create `intelligence/skill/acquisition`:
   - Registry search (exact, similar, adaptable)
   - Knowledge acquisition for procedures
   - Procedure design
   - Validation pipeline
   - Security evaluation
   - Basic sandbox
3. Wire SkillGap → Skill Acquisition → Skill Registry → Planning
4. Connect Skill Registry to Planning (query by goal type)

**Exit Criteria:**
- Scenario: IDUN asked to do something with no matching skill → skill acquired, validated, registered, used in next turn
- Scenario: Skill validation fails → skill NOT registered; Host informed capability unavailable
- Scenario: High-risk skill → Constitutional Gate consulted; HMAC token required before first execution
- Existing capabilities fully described by SkillCards in registry
- Skill Registry queryable by goal type; returns ranked results

---

### Phase 7 — Event Router and Autonomous Operation

**Objective:** Enable IDUN to activate from non-Host events and operate proactively at Autonomy Level 1–2.

**Dependencies:** Phase 4 (Goal Manager), Phase 3 (Conversation Planner for notifications), Phase 1 (Attention wired).

**Scope:**
1. Create `runtime/eventrouter`: normalize events; evaluate Autonomy Policy
2. Implement Autonomy Policy (levels 0–5 as routing policy)
3. Create `runtime/notification`: terminal delivery; modality abstraction
4. Wire Scheduler → Event Router → Attention
5. Wire Goal Monitor → Event Router
6. Implement Autonomy Level 1 (informational notifications)
7. Implement Autonomy Level 2 (background research for active goals)

**Exit Criteria:**
- Scenario: Reminder fires at scheduled time → IDUN delivers concise notification without Host input
- Scenario: Knowledge change affects active goal → IDUN evaluates relevance → notifies if impactful
- Scenario: Non-relevant event → discarded silently; no notification
- Autonomy Policy correctly blocks Level 3+ actions when configured at Level 1
- No consequential autonomous actions taken without explicit Host authorization at Level 1

---

### Phase 8 — Episodic Memory and Memory Lifecycle

**Objective:** Complete the memory architecture with episodic memory, long-term memory taxonomy, and retention lifecycle.

**Dependencies:** Phase 1 (Reflection wired), Phase 2 (Working Memory consolidation).

**Scope:**
1. Create `intelligence/memory/episodic`
2. Create `intelligence/memory/longterm` (taxonomy + retention policy)
3. Implement Working Memory → Long-Term Memory consolidation with Reflection recommendation
4. Implement memory expiration (TTL-based) and staleness detection
5. Wire Reflection → Episodic Memory writes
6. Wire periodic Reflection for staleness review
7. Wire Reasoning → Episodic Memory for analogical case retrieval

**Exit Criteria:**
- Completed episode produces EpisodeRecord in Episodic Memory with Reflection findings
- Working Memory consolidation promotes Host corrections to Long-Term Memory automatically
- A Host preference observed 3 times is promoted automatically (configurable threshold)
- A stale Long-Term Memory record is flagged by periodic Reflection
- Reasoning retrieves an analogous past episode for a new similar goal
- Memory taxonomy correctly separates host knowledge, world knowledge, and experiences under distinct key namespaces

---

## Part 12 — Blueprint Audit Summary

### What Was Preserved

| Item | Notes |
|:-----|:------|
| All Layer 1 frozen contracts (`2.0.0-FROZEN`) | Unchanged. Wiring only where incomplete. |
| Understanding U8 frozen spec | Pending: Deliberative LLM injection |
| Reasoning V3 11-stage cascade | Pending: LLM injection + GoalProposal/GapSignal extensions |
| Planning V3 HTN/GOAP/A* | Pending: boundary fix + Knowledge Store reads |
| Decision V3 4-tier evaluation | Extended with new SUGGEST_ALTERNATIVES outcome type |
| Executive V2 content-blind orchestration | Pending: episode lifecycle wiring |
| Reflection 2-mode design | Pending: production wiring |
| Learning Snapshot lifecycle | Pending: production wiring |
| Constitutional Gate HMAC model | Authoritative — unchanged |
| Calibration single-owner model | Authoritative — unchanged |
| `core/memory`, `core/scheduler` | Preserved; extended by higher-level stores |

### What Was Refined

| Item | Refinement |
|:-----|:-----------|
| Gap Classifier → not a subsystem | Rejected dedicated Gap Classifier; classification belongs to the detecting stage (Decision 0.1) |
| Gap classification ownership | Corrected: 4 gap types have 4 different owners — Reasoning (Knowledge), Planning (Skill), Context Resolver (Clarification), Constitutional Gate (Authorization) |
| Event Bus gap routing removed | Event Bus transports gap signals content-blindly; Goal Manager dispatches via embedded `routeGap()` function |
| Goal Manager as gap lifecycle owner | Goal Manager owns the complete gap resolution lifecycle including Level 2 timeout |
| Two-level timeout model | Operation timeout (Level 1) owned by Acquisition; Resolution timeout (Level 2) owned exclusively by Goal Manager |
| Working Memory positioning | Corrected from "pipeline stage" to "omnipresent store" |
| Context Resolver output | Produces `ResolvedContext` wrapper; does NOT mutate `UnderstandingBatch` in place |
| Reasoning → TopicActiveGoals | Corrected to: Reasoning emits GoalProposal → Goal Manager validates and registers |
| Reasoning gap scope | Corrected: Reasoning emits `KnowledgeGap` only — not all four gap types |
| Language Realization input | Receives `ConversativeIntent` instead of raw `ExecutionResponse` |
| Gap resolution mechanism | Formalized as "continuation episode" model with Goal Manager orchestration, not recursive invocation |
| Episode vs. Turn distinction | Formally defined; not every Turn creates an Episode |
| Memory taxonomy | Formalized as 5 distinct stores |
| Planning → Reasoning import | Identified as boundary violation (AI-16); resolved via `intelligence/types` |

### What Was Challenged and Rejected

| Proposal | Reason Rejected |
|:---------|:----------------|
| Gap Classifier as separate subsystem | Deterministic routing by enum type needs no subsystem; Principle P12 |
| "Reason again" as architectural step | Recursive self-invocation violates bounded episode model; continuation episode is correct |
| Working Memory as a pipeline stage | Working Memory is omnipresent; sequential positioning was architecturally incorrect |
| Single generic Memory store | 5-store taxonomy serves genuinely different purposes with different lifecycle rules |

### What Remains Open

#### Decision 0.1 — Resolved Questions (formerly OQ-0.1-1 through OQ-0.1-4)

All four architectural open questions from the initial Decision 0.1 review have been resolved. See §0.1.N (21 confirmed decisions) and §0.1.O.

| Former OQ | Resolution | Section |
|:----------|:-----------|:--------|
| OQ-0.1-1 — `GapSignal.DependsOn` | **RESOLVED** — Do NOT add. Sequential deps handled by continuation-episode model; parallel gaps use `PendingCount` if ever needed. | §0.1.K |
| OQ-0.1-2 — Restart knowledge verification | **RESOLVED** — Goal Manager treats uncertain state as UNCERTAIN, re-dispatches same GapID; Acquisition subsystem owns idempotency check against output store. | §0.1.I |
| OQ-0.1-3 — `AWAITING_HOST` lifetime | **RESOLVED** — No universal mandatory timeout. Three separate concepts: `Goal.ExpiresAt` ≠ `GapRecord.Deadline` ≠ Host Staleness Policy (Levels 4–5). | §0.1.G |
| OQ-0.1-4 — `ClarificationGap` lifecycle | **RESOLVED** — Always routes through Goal Manager. Two-tier wait model: blocking goal → persisted `GapRecord`; non-blocking → transient `PendingClarification` in Working Memory. | §0.1.J |

#### Decision 0.1 — Remaining Open (Policy Decisions Only)

| Open Question | ID | Notes |
|:-------------|:---|:------|
| Exact `ReminderInterval` per priority band for Host staleness | OQ-0.1-3a | Configuration policy; owned by SS-32 (Runtime Policy). Do not hardcode. |
| `ArchiveThreshold` for stale goals | OQ-0.1-3b | Configuration policy; owned by SS-32. OFF by default. |
| Parallel gap `GapBatch.Strategy` + `PendingCount` mechanism | OQ-0.1-1a | Not needed until demonstrated requirement exists. |
| `PendingClarification` token lifetime in Working Memory | OQ-0.1-4a | Resolve in Working Memory design specification (Phase 2). |


#### Other Open Questions

| Open Question | Resolution Path |
|:-------------|:----------------|
| Conversation preference persistence boundary | Resolve in Phase 3 design specification |
| Multi-modal understanding scope | Define extension points in Phase 2; implementation is North Star future |
| Skill sandboxing mechanism (process isolation vs. in-process) | Resolve in Phase 6 design specification |
| Episodic Memory retention limits and summarization strategy | Resolve in Phase 8 design specification |
| Working Memory capacity constants (specific N values) | Configurable defaults; calibrate during Phase 2 integration testing |
| Internet research authorization (pre-approved or per-request) | Policy decision required from Host before Phase 5 |
| Cross-goal knowledge sharing | Automatic via Knowledge Store; Working Memory does not share across goals |

---

## Part 13 — North Star Alignment

| North Star Capability | Blueprint Coverage | Phase |
|:----------------------|:------------------|:------|
| Natural language understanding | Understanding (existing + Phase 1 deliberative LLM) | 1 |
| Conversational context | Working Memory | 2 |
| Problem solving | Reasoning (existing + Phase 1 LLM) | 1 |
| Mathematical reasoning | Reasoning + tool capabilities | 5 |
| Complex multi-step planning | Planning + Goal Manager | 4 |
| Recognizing missing info | Gap Signal types (Phase 2 contracts) | 2 |
| Acquiring missing knowledge | Knowledge Acquisition | 5 |
| Verifying acquired knowledge | Knowledge Verification (part of Acquisition) | 5 |
| Recognizing skill gaps | SkillGap signal type (Phase 2 contracts) | 2 |
| Acquiring skills | Skill Acquisition | 6 |
| Using tools | Capabilities (existing) + Skill Registry | 6 |
| Remembering | Long-Term Memory lifecycle | 8 |
| Forgetting / expiring | Memory expiration, staleness detection | 8 |
| Reflecting | Reflection (wired Phase 1) | 1 |
| Learning from experience | Learning (wired Phase 1) | 1 |
| Improving strategies | Learning Rollout Executor | 1 |
| Adapting interaction behavior | Conversation Learning | 3+ |
| Communicating naturally | Conversation Planner + Realization | 3 |
| Autonomous operation | Event Router + Autonomy Policy | 7 |
| Proactive notification | Notification Service | 7 |
| Safe, constitutional, offline-first | Constitution (existing, all phases) | All |

---

*Implementation must not begin on any phase until this Blueprint is reviewed and approved. After approval, implementation begins at Phase 1 — the highest-value, lowest-risk entry point that requires no new packages and no architectural changes.*