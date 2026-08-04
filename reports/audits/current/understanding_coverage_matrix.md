# Phase 4B: Understanding Coverage Matrix

This design report establishes the current baseline of the Understanding layer. It documents exactly what is supported by the `GrammarSpecialist` and the semantic extractors.

## 1. Grammar Specialist Coverage Matrix

The following table documents every intent currently registered in the `GrammarSpecialist` and maps its capabilities through the semantic extraction pipeline.

| Intent | Grammar Support | Slot Extraction | Entities | References | Temporal Anchors |
|---|:---:|:---:|:---:|:---:|:---:|
| `greet_user` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `farewell_user` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `query_identity` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `query_wellbeing` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `express_thanks` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `confirm` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `cancel_action` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `calculate` | ✅ (Pattern) | ✅ (`expression`, `operand1`, `operand2`, `operator`) | ✅ (`EntityQuantity`) | ❌ | ❌ |
| `file_operation` | ✅ (Pattern) | ✅ (`operation`, `filename`, `destination`, `source`, `extension`, `directory`, `path`) | ✅ (`EntityFile`) | ✅ (`RefPronoun`) | ❌ |
| `create_directory` | ✅ (Pattern) | ✅ (`directory`) | ✅ (`EntityFile`) | ❌ | ❌ |
| `list_files` | ✅ (Pattern) | ✅ (`directory`) | ✅ (`EntityFile`) | ❌ | ❌ |
| `create_reminder` | ✅ (Pattern) | ✅ (`target`, `task`, `time`, `person`, `date`, `duration`) | ✅ (`EntityUnknown`) | ✅ (`RefPronoun`) | ✅ (`TempAbsolute`) |
| `query_weather` | ✅ (Pattern) | ✅ (`location`, `date`, `duration`, `daypart`) | ✅ (`EntityLocation`) | ❌ | ✅ (`TempAbsolute`) |
| `take_note` | ✅ (Pattern) | ✅ (`content`, `title`) | ✅ (`EntityUnknown`) | ❌ | ❌ |
| `delete_note` | ✅ (Pattern) | ✅ (`title`) | ❌ | ✅ (`RefPronoun`) | ❌ |
| `read_note` | ✅ (Pattern) | ✅ (`title`) | ❌ | ✅ (`RefPronoun`) | ❌ |
| `query_time` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `query_date` | ✅ (Pattern) | ❌ | ❌ | ❌ | ❌ |
| `query_battery` | ✅ (Pattern) | ✅ (`target`) | ✅ (`EntitySystemResource`) | ❌ | ❌ |
| `query_cpu` | ✅ (Pattern) | ✅ (`target`) | ✅ (`EntitySystemResource`) | ❌ | ❌ |
| `query_memory` | ✅ (Pattern) | ✅ (`target`) | ✅ (`EntitySystemResource`) | ❌ | ❌ |
| `query_disk` | ✅ (Pattern) | ✅ (`target`) | ✅ (`EntitySystemResource`) | ❌ | ❌ |
| `system_shutdown` | ✅ (Pattern) | ✅ (`operation`, `target`, `date`, `time`) | ✅ (`EntitySystemResource`, `EntityUnknown`) | ❌ | ✅ (`TempAbsolute`) |
| `system_restart` | ✅ (Pattern) | ✅ (`operation`, `target`, `date`, `time`) | ✅ (`EntitySystemResource`, `EntityUnknown`) | ❌ | ✅ (`TempAbsolute`) |
| `system_lock` | ✅ (Pattern) | ✅ (`operation`, `target`) | ✅ (`EntitySystemResource`, `EntityUnknown`) | ❌ | ❌ |
| `query_status` | ✅ (Exact - Legacy) | ❌ | ❌ | ❌ | ❌ |
| `buy` | ✅ (Prefix - Legacy) | ✅ (`item`) | ✅ (`EntityProduct`) | ❌ | ❌ |

## 2. Intent Coverage Map

| Intent | Architectural Coverage | User-Facing Capability |
|---|:---:|:---:|
| `greet_user` | 🟢 Full | 🟢 Full |
| `query_status` | 🟢 Full | 🟡 Partial |
| `cancel_action` | 🟢 Full | 🟡 Partial |
| `query_identity` | 🟢 Full | 🟢 Full |
| `query_wellbeing`| 🟢 Full | 🟢 Full |
| `farewell_user` | 🟢 Full | 🟢 Full |
| `query_time` | 🟢 Full | 🟡 Partial |
| `set_alarm` | 🟡 Partial | 🟡 Partial |
| `create_reminder`| 🟡 Partial | 🟡 Partial |
| `query_weather` | 🟡 Partial | 🟡 Partial |
| `take_note` | 🟡 Partial | 🟡 Partial |
| `buy` | 🟡 Partial | 🟡 Partial |
| `delete` | 🟡 Partial | 🟡 Partial |
| `calculate` | 🟢 Full | 🟢 Full |
| `file_operation` | 🟢 Full | 🟢 Full |
| `create_directory`| 🟢 Full | 🟢 Full |
| `list_files` | 🟢 Full | 🟢 Full |
| `system_shutdown` | 🟢 Full | 🟢 Full |
| `system_restart` | 🟢 Full | 🟢 Full |
| `system_lock` | 🟢 Full | 🟢 Full |
| `query_battery` | 🟢 Full | 🟢 Full |
| `query_disk` | 🟢 Full | 🟢 Full |
| `query_cpu` | 🟢 Full | 🟢 Full |
| `query_memory` | 🟢 Full | 🟢 Full |
| `query_date` | 🟢 Full | 🟢 Full |
| `express_thanks` | 🟢 Full | 🟢 Full |
| `confirm` | 🟢 Full | 🟢 Full |
| `delete_note` | 🟢 Full | 🟢 Full |
| `read_note` | 🟢 Full | 🟢 Full |

## 3. Coverage Analysis

### Coverage Percentage
Based on the Phase 4B.2 Closure Sprint trace verification (34/34 targeted inputs):
- **Grammar Recognition:** 100% on defined deterministic corpus (all capability families)
- **FAILED_IMPASSE Rate:** Pre-closure 17.4% (69-input trace); post-closure shadowing eliminated for Weather/Notes
- **Supported Intents:** 27 deterministic intents registered across 7 highest-impact capability families (25 V3 Pattern intents, 2 legacy).
- **Slot Extraction:** 100% on closure sprint corpus. All required raw slots extracted across all 6 capability families.
  - **Files**: Added `path` slot; Windows (`C:\...`, `C:/...`) and Unix (`/...`) absolute paths now extracted.
  - **System**: Added `operation` and `time` slots. All system action verbs extracted as raw strings.
  - **Weather**: Added `daypart` slot (morning, afternoon, evening, night, noon, midnight).
  - **Shadowing Fixed**: Rule registration reordered — Weather and Notes now evaluate before Calculator and Files fallbacks.
- **Entity Extraction:** Extended to properly map `EntityFile`, `EntityQuantity`, and generalized `EntityUnknown` for content.
- **Reference Extraction:** Generalized pronoun detection resolving contextual targets dynamically.

### Missing Grammar Rules
- **Rich Temporal Queries:** (e.g., `what is on my schedule next Friday`)
- **Natural Conversational Idioms:** Variations in phrasing outside the defined deterministic bounds.

### Missing Slot Extraction
- **Quantities/Amounts:** Needs broader extraction beyond explicit math operands.

### Missing Semantic Extraction
- **Sub-slot Entity Parsing:** Identifying a `person` inside a `task` string without explicit pattern boundary separation.

## 4. Architectural Recommendations

1. **Keep Semantic Extraction and Normalization Separate**
   - Maintain a clean pipeline boundary: `Grammar` -> `Slots` -> `TemporalExtractor` -> `TemporalNormalizer` -> `TemporalAnchors`.
2. **Delay Compound Intent Detection**
   - Ensure the single-intent pipeline is fully robust before introducing the complexity of parsing multiple intents (e.g. `and`/`then`).
3. **Modularize Extractor Implementation**
   - Begin modularizing the `extractors.go` file before it grows into a monolithic structure.

## 5. Remaining Implementation (Phase 4B.5+)

### Phase 4B.5: Natural Language Expansion & Error Recovery
1. **Ambiguity Resolution**
   - Implement handlers for ambiguous expressions that do not fit strictly into deterministic ontology constraints.

*(Note: Entity Grounding, Memory Lookup, and Reference Resolution are responsibilities strictly owned by the Reasoning layer, not Understanding).*

## 6. Understanding Baseline Status

- **Architecture**: ✅ Frozen
- **Cognitive API Specification**: ✅ Frozen
- **Semantic Contracts**: ✅ Frozen
- **Schema**: ✅ Frozen
- **Phase 4B.1**: ✅ Complete
- **Phase 4B.2**: ✅ Complete
- **Phase 4B.3**: ✅ Complete
- **Phase 4B.4**: ✅ **FROZEN** — Temporal Processing. Deterministic normalization of temporal semantic objects completed. Extractors extract, builders build, normalizers normalize.
- **Current Focus**: Phase 4B.5 — Natural Language Expansion & Error Recovery

**Baseline Freeze Statement**
Phase 4B.4 is officially frozen as of 2026-08-03. The Understanding layer successfully converts all recognized raw slots into strongly-typed semantic objects according to the extensible semantic ontology and normalizes temporal objects into canonical machine representations. Strict separation of concerns is maintained: meaning is assigned deterministically and temporal objects are normalized, but no entity grounding, memory lookup, or reasoning occurs. The pipeline successfully executes: Grammar -> Raw Slots -> Semantic Objects -> Temporal Normalization -> Semantic Interpretation. All tests pass and the Semantic Ontology is formally documented.

## 7. Understanding Roadmap

✅ Phase 4B.1 — Deterministic Language Understanding
✅ Phase 4B.2 — Raw Slot Extraction
✅ Phase 4B.3 — Semantic Object Construction
✅ Phase 4B.4 — Temporal Processing
Phase 4B.5 — Natural Language Expansion & Error Recovery
Phase 4B.6 — Compound Intent Detection
Understanding Frozen
Reasoning Evolution (Grounding, Memory Lookup, Reference Resolution, Context Resolution, Confidence)

## 8. Cognitive Enrichment Principle

**Each Understanding phase only enriches the representation produced by the previous phase.**
- No phase modifies or reinterprets information produced by earlier phases.
- No phase performs responsibilities owned by later cognitive layers.
