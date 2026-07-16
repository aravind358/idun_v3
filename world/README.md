# IDUN V3 World Subsystem Architecture

**Architecture Version:** `2.0.0-FROZEN`
**Package:** `idun/world`
**Classification:** Layer 1 Boundary Subsystem
**Status:** PERMANENT ARCHITECTURE SPECIFICATION (FROZEN FOR 20–30 YEAR LIFECYCLE)

---

## 1. Purpose & Architectural Philosophy

World is the sole boundary between IDUN and the external world. It accepts input from external adapters, normalizes and fingerprints it, constructs `Interaction` artifacts, publishes them to the Global Workspace, and eventually presents `Response` artifacts via output adapters.

```
External World
      │
      ▼
┌────────────────────────────┐
│      InputAdapter          │  (Text, Voice†, Vision†, API†)
│  Receive() → Interaction   │
└────────────┬───────────────┘
             │ HandleInteraction()
             ▼
    ┌─────────────────────────────────────────────────┐
    │                 World.Service                   │
    │  ┌───────────────────────────────────────────┐  │
    │  │  WorldPolicyProfile (read-only from Exec) │  │
    │  │  WorldCapabilities  (immutable at start)  │  │
    │  │  WorldSummary       (bounded telemetry)   │  │
    │  └───────────────────────────────────────────┘  │
    └───────────────────┬─────────────────────────────┘
                        │ Publish → TopicPerception
                        ▼
    ┌─────────────────────────────────────────────────┐
    │              Global Workspace                   │
    │           (TopicPerception topic)               │
    └───────────────────┬─────────────────────────────┘
                        │ Understanding subscribes
                        ▼ interprets intent
                     Executive coordinates
                        │
                        │ TopicActionExecution
                        ▼
    ┌─────────────────────────────────────────────────┐
    │              World.Service                      │
    │    handleResponseEnvelope() → Response          │
    └───────────────────┬─────────────────────────────┘
                        │ OutputAdapter.Send()
                        ▼
┌────────────────────────────┐
│      OutputAdapter         │  (Text, Voice†, Screen†, Robot†)
│  Send(Response)            │
└────────────────────────────┘
                        │
                        ▼
              External World
```

† = Future adapter (Post Layer 1)

---

## 2. Constitutional Responsibilities

### World Owns
- Accepting raw input from external adapters
- Normalizing input (whitespace, UTF-8) according to `WorldPolicyProfile`
- Computing deterministic `InteractionFingerprint` (SHA-256)
- Constructing immutable `Interaction` artifacts with `OriginalInput` and `NormalizedInput`
- Storing large payloads via `PayloadStorer` (content-addressed, via Core.Storage)
- Publishing `Interaction` envelopes to the Global Workspace (`TopicPerception`)
- Subscribing to `TopicActionExecution` for asynchronous response delivery
- Presenting `Response` artifacts via output adapters
- Recording `WorldTrace` telemetry (write-only; Reflection analyzes it)
- Maintaining bounded `WorldSummary` statistics

### World Never Owns
- Reasoning, planning, decision making, learning, or reflection
- Interpreting whether input is a user intent, tool request, or system event (Understanding)
- Evaluating attention or salience (Attention)
- Memory ownership or workspace state management
- Modifying `WorldPolicyProfile` or `WorldCapabilities`
- Making cognitive judgments about safety or constitutional alignment (Value)

---

## 3. Event-Driven Architecture (Refinement 11)

World is **fully event-driven**. It never blocks waiting for Executive or any cognitive subsystem.

```
Receive → HandleInteraction → Publish → return immediately

                                         ↓ (asynchronous)
                       Workspace delivers TopicActionExecution envelope
                                         ↓
                       handleResponseEnvelope() builds Response
                                         ↓
                       OutputAdapter.Send(Response)
```

This preserves compatibility with:
- Background planning episodes (long-duration cognitive tasks)
- Streaming responses (future)
- Long-running Executive episodes
- Distributed interaction gateways (future)

---

## 4. Domain Model

### 4.1 Interaction

`Interaction` is the canonical immutable artifact representing one user turn.

| Field | Purpose |
|---|---|
| `InteractionID` | Unique identifier (secure random UUID) |
| `SessionID` | Groups interactions into a conversation session |
| `Origin` | Source of the interaction (`USER`, `VOICE`, `API`, etc.) |
| `Modality` | Channel type (`TEXT`, `VOICE`, `VISION`, `API`) |
| `OriginalInput` | **Exactly** what the external adapter received (verbatim) |
| `NormalizedInput` | Canonical form after policy normalization |
| `PayloadRef` | Content-addressed storage URI (via Core.Storage) |
| `CreatedAt` | UTC timestamp of construction |
| `ReplayMetadata` | Deterministic replay provenance |

### 4.2 Response

`Response` is the canonical immutable artifact presented to the external world.

| Field | Purpose |
|---|---|
| `ResponseID` | Unique identifier |
| `InteractionID` | Links response to originating Interaction |
| `Content` | Human-readable response (or empty if PayloadRef holds large content) |
| `Status` | Individual response outcome (`SUCCESS`, `ERROR`, `TIMEOUT`) |
| `ResultStatus` | Full lifecycle outcome (`SUCCESS`, `FAILED`, `TIMEOUT`, `DROPPED`, `INTERRUPTED`) |
| `TerminationReason` | Factual cause of non-success (`INVALID_INPUT`, `WORKSPACE_FAILURE`, etc.) |
| `ExecutionDuration` | Wall-clock time from Interaction creation to Response delivery |
| `ReplayMetadata` | Deterministic replay provenance |

### 4.3 InteractionOrigin vs Modality

These are orthogonal:
- **`Modality`** = the communication medium (TEXT, VOICE, VISION, API)
- **`InteractionOrigin`** = who sent the interaction (USER, ROBOT, SCHEDULER, SIMULATION, etc.)

World records both. Understanding interprets content.

---

## 5. WorldPolicyProfile

`WorldPolicyProfile` is owned externally (by Runtime or Executive) and consumed **read-only** by World. World never modifies it.

| Field | Purpose |
|---|---|
| `MaximumInputLength` | Maximum bytes accepted per interaction |
| `MaximumResponseLength` | Maximum bytes in response content |
| `NormalizeWhitespace` | Whether to collapse whitespace runs |
| `DropInvalidUTF8` | Whether to drop invalid UTF-8 bytes |
| `AllowEmptyInput` | Whether to accept empty inputs after normalization |
| `ResponseTimeout` | Maximum wait time before timeout |
| `PolicyFingerprint` | Deterministic SHA-256 digest of all policy fields |
| `PolicyVersion` | Policy version string |

---

## 6. WorldCapabilities

`WorldCapabilities` describes what this deployment can receive and present. Immutable after startup.

| Field | Meaning |
|---|---|
| `SupportsText` | Text I/O is available |
| `SupportsVoice` | Voice I/O is available (future) |
| `SupportsVision` | Vision input is available (future) |
| `SupportsAPI` | Machine API input is available (future) |
| `SupportsStreaming` | Streaming response delivery is available (future) |
| `SupportsAttachments` | Attachment payloads are supported (future) |
| `CapabilityFingerprint` | Deterministic SHA-256 digest of capability flags |

---

## 7. WorldTrace (Write-Only Telemetry)

Every interaction produces a `WorldTrace`. World writes it; **Reflection analyzes it**. World never evaluates its own traces.

| Field | Purpose |
|---|---|
| `InteractionID` | Links trace to its Interaction |
| `InteractionFingerprint` | Deterministic SHA-256 over interaction content + policy |
| `AdapterName` | Which InputAdapter received the interaction |
| `AdapterVersion` | Implementation version of the adapter |
| `AdapterFingerprint` | Deterministic identity of the adapter implementation |
| `Origin` / `InputModality` / `OutputModality` | Channel metadata |
| `ExecutionTime` | Wall-clock processing duration |
| `WorldVersion` / `PolicyFingerprint` / `CapabilityFingerprint` | Replay provenance |

---

## 8. Adapter Interface & Identity (Refinement 10)

All adapters carry immutable identity to enable exact replay when adapter implementations evolve:

```go
type InputAdapter interface {
    Receive(ctx context.Context) (*Interaction, error)
    Name() string
    AdapterVersion() string    // e.g., "2.0.0-FROZEN"
    AdapterFingerprint() string // SHA-256(Name + Version)
    Close() error
}

type OutputAdapter interface {
    Send(ctx context.Context, response *Response) error
    Name() string
    AdapterVersion() string
    AdapterFingerprint() string
    Close() error
}
```

---

## 9. Topic Ownership (Refinement 13)

World publishes to **`TopicPerception`** — the semantically closest existing topic for raw external world input. World is content-blind and does not classify whether input is a user intent, tool request, or system event. That interpretation belongs to **Understanding**.

> **Post Layer 1 Evolution (v3.0+):** A dedicated `TopicInteraction` will be introduced to allow Understanding to explicitly subscribe to all World inputs without sharing the generic `TopicPerception` channel. This is deferred to avoid breaking the frozen `communication` package.

---

## 10. Content-Addressed Storage (Refinement 12)

Large interaction payloads are stored via `PayloadStorer` (a narrow bridge to Core.Storage) and referenced only by `PayloadRef` (SHA-256 hex). Raw content is never embedded directly in Workspace envelopes.

---

## 11. Runtime Integration

`World.Service` is registered at **`kernel.PhaseBackground` (Phase 6)** in `RuntimeHost`, ensuring all cognitive subsystems are already running before the world boundary accepts external input.

```
Phase 1: Core (Memory, Storage, Scheduler)
Phase 2: Infrastructure (Registry, Bus, Boundary, Permission, Constitution, Calibration)
Phase 3: Workspace
Phase 4: Executive, Attention
Phase 5: Understanding, Reasoning, Planning, Decision
Phase 6: Reflection, Learning, World.Service ← here
```

---

## 12. Implemented Adapters (Phase 1)

| Adapter | Location | Status |
|---|---|---|
| `TextInputAdapter` | `adapters/text/` | ✅ Implemented |
| `TextOutputAdapter` | `adapters/text/` | ✅ Implemented |

---

## 13. Future Enhancements (Post Layer 1)

### Future Input Adapters
- **Voice** (`adapters/voice/`) — speech-to-text via pluggable TTS engine
- **Vision** (`adapters/vision/`) — image/video frame capture
- **API** (`adapters/api/`) — machine-to-machine structured clients
- **Robot** — sensor and actuator input bridges
- **Sensors** — IoT and environmental data streams

### Future Output Adapters
- **Speech** (`output/voice/`) — text-to-speech delivery
- **Screen UI** (`output/screen/`) — graphical frontend rendering
- **Robot Control** (`output/robot/`) — actuator command translation
- **Holographic UI** — volumetric display adapters
- **AR/VR** — augmented and virtual reality interface adapters

### Future World Features
- **Streaming Responses** — incremental token delivery for long responses
- **Attachment Handling** — binary payload forwarding with PayloadRef routing
- **Multimodal Sessions** — simultaneous text + voice + vision per session
- **Remote Client Connections** — distributed World endpoints across network
- **Distributed Interaction Gateways** — federated World boundary for multi-agent deployments
- **TopicInteraction** — dedicated workspace topic (v3.0+, without breaking Layer 1)

---

## 14. Responsibility Audit

| Concern | Owner | NOT World |
|---|---|---|
| Input normalization | World ✅ | — |
| Content interpretation | Understanding | ✅ |
| Salience evaluation | Attention | ✅ |
| Response generation | Executive + Cognitive | ✅ |
| Constitutional safety | Value (Constitution) | ✅ |
| Memory storage | Memory | ✅ |
| Workspace state | Workspace | ✅ |
| Policy creation | Runtime/Executive | ✅ |
| Replay telemetry | Reflection | ✅ |
| Cognitive episode management | Executive | ✅ |
