# Phase 4B.3: Semantic Object Construction Readiness Audit

## 1. Raw Slot Inventory

The following raw slots are currently produced by the deterministic `GrammarSpecialist` (Phase 4B.2 boundary):

| Capability Family | Intent | Raw Slots |
| :--- | :--- | :--- |
| **Calculator** | `calculate` | `operator`, `operand1`, `operand2`, `expression` |
| **Files** | `file_operation`, `create_directory`, `list_files` | `operation`, `path`, `destination`, `source`, `directory`, `filename`, `extension` |
| **Reminder** | `create_reminder` | `task`, `target`, `person`, `date`, `time`, `duration` |
| **Weather** | `query_weather` | `location`, `date`, `duration`, `daypart` |
| **Notes** | `take_note`, `delete_note`, `read_note` | `title`, `content` |
| **System** | `query_time`, `query_date`, `query_battery`, `query_cpu`, `query_memory`, `query_disk`, `system_shutdown`, `system_restart`, `system_lock` | `target`, `operation`, `date`, `time` |

## 2. Semantic Mapping Matrix

During Phase 4B.3, the raw slots above must be mapped into typed semantic objects without modification, normalization, or external resolution.

| Raw Slot | Semantic Object Type |
| :--- | :--- |
| `person` | `EntityPerson` |
| `location` | `EntityLocation` |
| `filename`, `extension` | `EntityFile` |
| `path` | `EntityFile` or `EntityDirectory` (context-dependent) |
| `directory` | `EntityDirectory` |
| `title`, `content`, `task` | `EntityDocument` |
| `operand1`, `operand2`, `expression` | `EntityNumber` / `EntityQuantity` |
| `target` (pronouns) | `Reference` |
| `target` (system targets like "cpu", "battery") | `EntityProduct` / `EntityConcept` |
| `date`, `time`, `daypart`, `duration` | `TemporalAnchor` |

## 3. Existing Semantic Object Audit

A review of `intelligence/understanding/v3/semantic_interpretation.go` reveals the current state of semantic objects:

- **Existing & Ready**: `EntityPerson`, `EntityOrganization`, `EntityLocation`, `EntityProduct`, `EntityQuantity`, `EntityConcept`, `EntityFile`.
- **Partially Implemented**: 
  - `TemporalAnchor` exists, but `daypart` is not currently extracted by `SlotBasedTemporalExtractor`.
  - `Reference` exists, but `SlotBasedReferenceExtractor` blindly scans all slot values for words like "it/this/that", which is dangerous (e.g., a note content of "do it" will misfire as a Reference).
- **Missing**:
  - `EntityDocument` (for notes and task contents)
  - `EntityDirectory` (for file operations)
  - `EntityNumber` (for mathematical operands)
  - `EntityUnit` (for future quantity extraction)
- **Duplication**: `extractors.go` loops over slots multiple times across different extractors, often with overlapping fallback logic. `EntityExtractor` assigns `EntityUnknown` to `content` and `task`, which is semantically weak.

## 4. Boundary Verification

Phase 4B.3 is strictly defined. The planned semantic extraction layer adheres to all architectural boundaries:
- **Construct semantic objects**: Yes, by mapping typed structs (e.g., `Entity`, `Reference`).
- **Preserve raw slot values**: Yes, values are copied to the `.surface` field.
- **Avoid temporal normalization**: Yes, strings like "tomorrow" will remain "tomorrow" inside the `TemporalAnchor`.
- **Avoid memory lookup / entity grounding**: Yes, `CanonicalName` and `GroundingID` will remain empty or equal to surface text.
- **Avoid reference resolution**: Yes, `anchorHint` and `resolved` will remain empty/false.
- **Avoid reasoning/planning**: Yes.

## 5. Extractor Architecture Review

**Current Architecture:**
`extractors.go` currently defines a monolithic file containing multiple independent heuristic extractors (`SlotBasedEntityExtractor`, `SlotBasedReferenceExtractor`, `SlotBasedTemporalExtractor`). They iterate over `Hypothesis.Slots()` independently.

**Recommendation for Evolution:**
As we move to deterministic rule-based semantic extraction and prepare for future neural hybrids (Phase 4C), the extractors should be modularized into a dedicated sub-package (`intelligence/understanding/v3/extractors/`).
- `extractors/entity.go`
- `extractors/temporal.go`
- `extractors/reference.go`

This isolates domain-specific extraction rules. Additionally, a `Context` or `SemanticFrame` pointer should be passed through the extractor pipeline so that extractors don't conflict (e.g., if a Reference extractor claims a word, the Entity extractor knows to skip it). 

*Refactoring will occur in the implementation phase.*

## 6. Phase 4B.3 Implementation Plan

**1. Update V3 Semantic Contracts (`semantic_interpretation.go`)**
- Add missing `EntityType` constants: `EntityDocument`, `EntityDirectory`, `EntityNumber`, `EntityUnit`.

**2. Modularize Extractors**
- Create `extractors/` subdirectory.
- Move and rename `extractors.go` logic into domain-specific files (`entity.go`, `temporal.go`, `reference.go`).

**3. Implement Deterministic Semantic Mapping**
- **Temporal**: Map `date`, `time`, `duration`, `daypart` slots strictly into `TemporalAnchor` objects without normalization.
- **Entities**: Map `person`, `location`, `filename`, `directory`, `title`, `content`, `operand1`, `operand2` strictly into `Entity` objects with appropriate `EntityType`.
- **References**: Map `target` into `Reference` strictly when it equals known pronouns (me, us, it, this, that), avoiding blind substring scans on content slots.

**4. Data Flow**
`GrammarSpecialist.Evaluate` -> `Hypothesis` -> `Orchestrator` -> `Extractors` (append to `SemanticFrame`).

**5. Verification Strategy**
- Write `extractors_test.go` verifying that semantic objects are generated correctly for each test case from the Phase 4B.2 Closure Sprint.
- Ensure that `TemporalAnchor.normalized` remains blank.
- Ensure that `Entity.groundingID` remains blank.

**6. Success Criteria**
- All 34 baseline test cases produce correct `SemanticFrame` objects containing the typed entities, temporal anchors, and references.
- No parsing regressions in the core grammar suite.
