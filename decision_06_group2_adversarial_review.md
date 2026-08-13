# Decision 0.6 — Group 2/3 Proposal: Adversarial Challenge

## 1. Q0.6-3 Verdict (Turn fast path bypassing Planning/Decision)

**Verdict:** APPROVE

**Exact Reasoning:** 
Forcing a conversational Turn through Planning and Decision strictly to produce a "no-op" plan violates the North Star principle of simplicity. Planning is designed to formulate multi-step world-modifying actions; Decision is designed to evaluate and approve them. A conversational Turn (e.g., "Hello" or "What time is it?") requires neither. The Conversation Planner (Decision 0.3) is already the designated owner of communicative intent. By allowing the Executive to route trivial Turns directly to the Conversation Planner, the architecture elegantly bypasses unnecessary cognitive overhead while preserving strict subsystem boundaries.

**Cross-decision evidence & Contradiction Check:**
- **Decision 0.3 (Communication):** Perfectly compatible. Decision 0.3 explicitly routes `Conversation Planner → Constitutional Gate → Language Realization`. Bypassing Planning/Decision does *not* bypass the Constitutional Gate, ensuring safety guarantees remain intact.
- **Decision 0.5 (Executive Mapping):** Compatible. Decision 0.5 designates the Executive as the orchestrator. If the `ReasoningContext` reveals no Episode-level requirements, the Executive maps the flow to the Communication path instead of `PlanningInput`.
- **Decision 0.6 Definition:** "Episode = bounded cognitive work with a meaningful goal, plan, and outcome." This explicitly justifies why Planning belongs exclusively to Episodes.

## 2. Q0.6-5 Verdict (Turn Lineage via EnvelopeID)

**Verdict:** APPROVE

**Exact Reasoning:** 
Creating a distinct `TurnID` provides zero architectural value because a Turn is, by definition, the direct processing of a single stimulus. The `SystemEvent.EnvelopeID` already uniquely identifies that stimulus. 

**Cross-decision evidence & Contradiction Check:**
- **Decision 0.5 (Lineage):** `EnvelopeID` is already established as the root lineage identifier for all contracts (e.g., `ReasoningContext`).
- **Replay Safety:** Replayed events retain the same `EnvelopeID`, allowing consistent tracing across restarts without generating new intermediate identifiers.

## 3. Before/After Architecture Diagrams

**Proposed & Corrected Model:** (The proposed model is architecturally sound and accepted without structural changes).

```text
                         HOST
                           │
                           ▼
                  SystemEvent (EnvelopeID)
                           │
                           ▼
                    UNDERSTANDING
                           │
                           ▼
                 CONTEXT RESOLUTION
                           │
                           ▼
                      REASONING
                           │
                           ▼
                 ReasoningContext
                 (Traces to EnvelopeID)
                           │
                           ▼
                      EXECUTIVE
                           │
                     CLASSIFICATION
                           │
             ┌─────────────┴─────────────┐
             │                           │
      NO EPISODE REQUIREMENT      EPISODE REQUIREMENT
             │                           │
             ▼                           ▼
           TURN                       PROMOTION
             │                           │
             │                     Allocate EpisodeID
             │                           │
             │                     Goal Manager (if Goal/Gap needed)
             │                           │
             │                           ▼
             │                        EPISODE
             │                           │
             │                     Planning (Episode-level)
             │                           │
             │                      Decision (Episode-level)
             │                           │
             └──────────────┬────────────┘
                            ▼
                   COMMUNICATION PATH
                  (Conversation Planner)
                            │
                            ▼
                     Constitutional Gate
                            │
                            ▼
                           HOST
```

## 4. Ownership Analysis

- **Understanding:** Owns parsing raw stimulus into semantic structures.
- **Context Resolution:** Owns enriching semantics with Working Memory context (pronouns, active topics).
- **Reasoning:** Owns domain logic, belief evaluation, and emitting `ReasoningContext` (identifying GoalProposals/Gaps).
- **Executive:** Owns execution orchestration, Turn/Episode classification, and Episode lifecycle (`EpisodeID`/`EpisodeRecord`).
- **Goal Manager:** Owns Goal and Gap lifecycles (`GoalID`/`GapRecord`).
- **Planning:** Owns multi-step plan generation (Episode-level only).
- **Decision:** Owns plan approval and outcome execution (Episode-level only).
- **Conversation Planner:** Owns conversational policy and `ConversativeIntent` generation.

## 5. Identity/Lineage Table

| Identifier | Semantic Purpose | Necessary? |
|:---|:---|:---|
| **EnvelopeID** | Identity of the originating external stimulus/event. Provides Turn lineage. | **YES** |
| **ArtifactID** | Identity of an internal data payload (e.g., `ReasoningContext`), mapping back to `EnvelopeID`. | **YES** (Decision 0.5) |
| **EpisodeID** | Identity of bounded cognitive execution. Groups Planning/Decision/LTM. | **YES** |
| **GoalID** | Identity of a long-term persistent intention. | **YES** |
| **TurnID** | Redundant wrapper around `EnvelopeID`. | **NO** (Reject) |

## 6. Failure/Restart Analysis

**Scenario 1: Crash before Episode persistence**
`Turn → Reasoning → Promotion begins → EpisodeID allocated → CRASH`
- **Recovery:** The `SystemEvent` is replayed with the *same* `EnvelopeID`. The Executive dynamically allocates a *new* `EpisodeID`.
- **Safety:** Perfectly safe. The unpersisted `EpisodeID` is lost in RAM. The replay provides deterministic lineage because `EnvelopeID` is constant.

**Scenario 2: Crash after GapRecord persistence, before ACK**
`Turn → Reasoning → GapSignal → Promotion → Goal Manager → GapRecord saved → CRASH`
- **Recovery:** 
  1. Event Router replays the `SystemEvent` (Turn).
  2. Goal Manager re-dispatches the Gap via Decision 0.1 Saga recovery.
- **Safety:** The Executive/Goal Manager uses `EnvelopeID` as the idempotency correlation key during the replay to avoid creating duplicate GoalRecords. Acquisition uses F3 `insert-if-absent` checks. Safe and deterministic. 
- *(Note: `EnvelopeID` as the idempotency key for Goal creation is an Implementation Specification, not a new architectural construct).*

## 7. Cross-Decision Contradiction Scan

- **Decision 0.1:** Compatible. Goal/Gap lifecycle remains authoritative.
- **Decision 0.2:** Compatible. `TURN` WM scope uses sliding windows; no `TurnID` required.
- **Decision 0.3:** Compatible. Fast path routes directly to Conversation Planner and hits the Constitutional Gate.
- **Decision 0.4:** Compatible. E2 initialization explicitly supports the Episode-level mechanisms.
- **Decision 0.5:** Compatible. `EnvelopeID` and `ArtifactID` contracts remain unchanged. Executive Mapping handles the routing branch.
- **Decision 0.6 Group 1:** Compatible. Fits perfectly with the structural promotion rules.
- **North Star:** Compatible. Maximizes simplicity by discarding `TurnID` and avoiding empty Planning runs.

## 8. Final Recommendation

Both **Q0.6-3** and **Q0.6-5** are architecturally sound, thoroughly tested against existing constraints, and **READY TO FREEZE**. No new architectural questions were created, and no new subsystems or IDs are required.
