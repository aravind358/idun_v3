# Phase 4B.4: Temporal Processing Walkthrough

## Overview

Phase 4B.4 (Temporal Processing) has been successfully completed and integrated into the Understanding V3 pipeline. This phase introduces deterministic normalization of temporal semantic objects, cleanly bridging the gap between raw natural language surface forms and standardized machine formats.

Crucially, this phase enforces the **Progressive Semantic Enrichment** architectural principle:
*   Extractors strictly extract.
*   Builders build.
*   Normalizers normalize.

## Key Accomplishments

### 1. Extended Core Time Service
The `idun/core/time` package now contains the generic, robust calendar utilities required by the normalizer without exposing any natural language components.
*   Added methods for precise date arithmetic: `AddDays`, `AddDuration`, `NextWeekday`, `StartOfDay`, `EndOfDay`.
*   Added `ParseClock` capable of correctly mapping string tokens (e.g. `5 PM`, `noon`) into base `time.Time` values.
*   Maintained `TimeService` as the singular source of truth for time operations in the project.

### 2. Additive Temporal Ontology
The semantic ontology for `v3` was expanded with precise temporal types without carrying over normalization logic:
*   `TempRelativeDate`
*   `TempRelativeWeekday`
*   `TempClockTime`
*   `TempRelativeDuration`
*   `TempTimeInterval`
*   `TempDaypart`
*   `TempAbsoluteDate`
*   `TempUnknown`

### 3. Explicit Normalization Pipeline
We created a strictly isolated processing package `intelligence/understanding/v3/normalizers/` with single responsibility. 
*   `NormalizerRunner` executes directly after `ExtractorRunner`, consuming the exact `TemporalAnchors` created during extraction and augmenting them via `TemporalNormalizer`.
*   The `Orchestrator` runs this sequentially and explicitly injects the normalizer using the standard interface.
*   No placeholder files were created for future components; deferred files (`number.go`, `unit.go`) have been documented in `TODO.md` as instructed.

The explicit pipeline executes as follows:
```text
User Input
      │
      ▼
Grammar Recognition
      │
      ▼
Intent Detection
      │
      ▼
Raw Slot Extraction
      │
      ▼
Semantic Object Construction
      │
      ▼
Temporal Extraction
      │
      ▼
Temporal Normalization
      │
      ▼
Semantic Interpretation
      │
      ▼
Reasoning
```

### 4. Semantic Enrichment, Not Overwrite
The `TemporalAnchor` behavior rigorously enforces the principle that raw user input is inviolable:
*   `Surface`: Retains the exact unmodified raw input string (`"tomorrow"`, `"5 PM"`).
*   `Type`: The assigned temporal classification (`TempRelativeDate`).
*   `Normalized`: The new dynamically generated canonical string mapping (e.g. `"2026-08-04"` or `"17:00"`).

For ambiguous temporal expressions like "later":
*   `Surface`: `"later"`
*   `Type`: `TempUnknown`
*   `Normalized`: `""`

Because "later" is intentionally treated as an ambiguous expression that depends on context or reasoning not available during semantic extraction, it is appropriately typed as `TempUnknown` and remains unnormalized.

### 5. Architectural Rule Enforcement
Added the **Temporal Processing Rule** to `.agents/AGENTS.md` and updated the project roadmaps in `understanding_coverage_matrix.md`, `tier_2_cognitive_intelligence_audit.md`, and `system_audit_report.md` to formally freeze Phase 4B.4.

### 6. Verification
The implementation adheres to strict verification criteria:
*   Every deterministic temporal object receives exactly one normalized representation.
*   Raw temporal values are never modified.
*   Unsupported or ambiguous expressions remain unnormalized.
*   Normalization uses the existing Core Time service.
*   No memory lookup occurs.
*   No entity grounding occurs.
*   No reasoning occurs.
*   No planning occurs.
*   Normalization is deterministic for the same reference time and configured timezone.

### 7. Build & Test Evidence
All `core/time` and `normalizers` tests pass. 
The system correctly intercepts surface strings like `"tomorrow"`, dynamically retrieves `timeSvc.Now()`, steps forward exactly one calendar day, and records the localized output into the `Normalized` payload without overriding any previous phase operations.

### 8. Next Steps
Phase 4B.5 (Natural Language Expansion & Error Recovery) is now the next active item on the Understanding Roadmap.
