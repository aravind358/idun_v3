# Architecture TODOs & Migration Tracking

This document tracks temporary architectural decisions, migration bridges, compatibility layers, and planned cleanup work. The goal is to ensure temporary solutions are not forgotten as IDUN evolves.

## Documentation
- Create `SYSTEM_ARCHITECTURE.md` (or `COGNITIVE_ARCHITECTURE.md`) after the major cognitive pillars stabilize.
- Document the complete top-to-bottom architecture using flowcharts and subsystem diagrams.
- Document every subsystem, its responsibilities, owned data structures, Workspace topics, and how information flows between modules.
- Maintain this document as the primary architectural reference to make onboarding, debugging, and future development easier.

## Context (Migration)
- Replace `dummyDialogueStateReader` (in `runtime/host.go`) once the real Dialogue State Manager is implemented at the global level.
- Remove any remaining obsolete V1 parsing/compatibility adapters across Reasoning and Context modules once the migration to V3 is completely stabilized.
- ~~Remove the temporary V1/V3 compatibility forwarding layer (in `intelligence/context/workspace_bridge.go`) after the Understanding pipeline fully migrates.~~ (Completed U7.5)
- ~~Update the Context Resolver to natively consume the V3 `UnderstandingBatch` containing `SemanticInterpretation` structures, instead of the legacy `SemanticFrame`.~~ (Completed U7.5)

## Context (Future Improvements)
- Replace `map[string]string` entity reference abstraction in Context Resolver with a more future-proof EntityReference abstraction or Entity Manager subsystem.
- Add `ResolutionMetadata` to Context payloads if useful.
- Add `PartiallyResolved` to `ResolutionStatus` if required by future reasoning capabilities.
- Improve context expiration with dynamic, goal-aware TTLs.
- Extend Context strategies to support richer references ("the first one", "the previous file", "the larger folder", etc.) when future phases require them.

## Understanding (U8 Multi-Intent & Composite Architecture)
- **Performance / N+1 Problem:** The current `deterministicSplitter` uses an $O(2^N)$ subset combination search where every combination check invokes `isValidGoal`, triggering a synchronous `cascadeAnalyze` (Grammar -> Neural -> Deliberative). This will cause exponential latency for utterances with many conjunctions. We must refactor this to either use a parser-based LLM segmentation step *before* individual goal evaluation, or parallelize the cascade evaluation.
- **Micro-splitting Conflict:** `extractors/temporal.go` handles micro-splitting ("tomorrow and friday"). We need to ensure the primary `Splitter` does not eagerly slice temporal boundaries into invalid intents before the temporal extractor can parse them.

## Understanding (U8 Future and Post-U8)
- **Dependency Metadata:** Future versions of `UnderstandingBatch` may include relationship metadata between split goals, such as sequential dependency, prerequisite, parallel, conditional, or parent/child intent. This belongs to future planning/reasoning work, not U8.
- **Pipeline Documentation:** Update `PIPELINE.md` after U8 is complete to show: Raw Input -> Normalizer -> Splitter -> Grammar -> Extractors -> UnderstandingBatch -> Context Resolver -> Reasoning.

## U8 TODO
- Replace heuristic connector list with a Connector Registry.
- Introduce a dedicated Temporal Span Detector module.
- Add dependency metadata to UnderstandingBatch (future U8.5/U9).
- Expand the certification corpus with linguistic edge cases.
- Document the Splitter algorithm in PIPELINE.md.

## Architecture Documentation
- Create a permanent `reports/SYSTEM_ARCHITECTURE.md` document that becomes the single source of truth for the entire IDUN architecture.
- Keep this document updated whenever a subsystem, pipeline, or architectural boundary changes.
- Ensure future architecture audits verify that this document stays synchronized with the implementation.

## Documentation Standards
- Every major subsystem should have clearly documented:
  - Purpose
  - Responsibilities
  - What it MUST do
  - What it MUST NOT do
  - Inputs
  - Outputs
  - Workspace Topics (if applicable)
  - Dependencies
  - Lifecycle
  - Roadmap milestone(s)
  - Future planned improvements
  - Related TODOs
