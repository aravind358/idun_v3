# ADR — World Owns All External Communication

## Purpose
Define a single boundary responsible for all communication between IDUN and the outside world, establishing a robust, plugin-based egress pipeline.

## Decision
- **World is the only ingress boundary.**
- **World is the only egress boundary.**
- Cognitive subsystems **never** communicate directly with external interfaces.
- Output handling is orchestrated by an **Output Manager** inside the World subsystem.
- **Aggregation comes first:** Multiple node results must be aggregated into a `CompositeResponse` before realization to solve multi-intent limitations.
- **Structured Output:** The realization engine returns a structured `OutputDocument`, not plain text.
- **Shared Formatting Abstraction:** Formatting is handled via a generic `Formatter` interface rather than bespoke plugin formatters.
- **Plugin Registry:** World utilizes a Plugin Registry to dynamically route output to multiple destinations (e.g., Console, Speaker, GUI) without hardcoded conditionals.
- **Output Plugins:** Modality extensions live in `world/plugins/`. Plugins are capability-based, implementing a standard `OutputPlugin` contract that defines their supported modalities and physical I/O writing.
- **Non-blocking Execution:** The Workspace event loop must never wait for realization. The Output Manager must dispatch realization asynchronously.
- Cognitive layers remain completely modality-agnostic. They never know whether the interaction is Text, Voice, Screen, Robot, API, or any future modality.

## Architectural Consequences
The standalone `presentation` package becomes an MVP artifact. Its reusable responsibilities will migrate into `world/output/` and `world/plugins/` using a safe, incremental O-Series roadmap.

This removes the current cyclic communication (`Presentation -> World`) and establishes a clean, plugin-based egress flow:
External Interface
        ↓
World (Ingress)
        ↓
Workspace
        ↓
(Cognitive Pipeline)
        ↓
World (Egress)
        ↓
Output Manager
        ↓
Aggregator
        ↓
Output Engine
        ↓
OutputDocument
        ↓
Plugin Registry
        ↓
┌──────────┬─────────┬─────────┬─────────┐
│          │         │         │         │
Text      Voice    Screen    Robot     API
Plugin    Plugin   Plugin    Plugin    Plugin
│          │         │         │         │
Formatter Formatter Formatter Formatter Formatter
│          │         │         │         │
Adapter   Adapter  Adapter   Adapter   Adapter
│          │         │         │         │
Console   Speaker  GUI       Hardware  HTTP

## Design Principles
This decision reinforces:
- Single Responsibility
- Separation of Concerns
- Loose Coupling
- Modality Independence
- Offline-first Architecture
- Scalable Multimodal Design
