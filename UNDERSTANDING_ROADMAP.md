# Understanding Subsystem Roadmap

This document serves as the official roadmap for the IDUN Understanding pillar. It tracks the evolution of the cognitive pipeline from raw perception to resolved semantic interpretation.

## Overview
The Understanding subsystem is responsible for translating unstructured user input into deterministic, canonical structures that can be executed safely by the IDUN cognitive architecture. It operates purely as a stateless pipeline, applying grammar, neural, and deliberative specialists in a cascaded architecture, followed by normalizers and semantic extractors.

## Current Architecture Status
- **Language Engine**: V3 (U6) Architecture — **FROZEN**
- **Context Subsystem**: U7 Architecture — **FROZEN** (Operates via Workspace on TopicUserIntent)
- **Data Payload**: Transitioning from `understanding.SemanticFrame` (V1) to `UnderstandingBatch` (V3).

---

## Completed Milestones

### U1 - U5: Foundation & Legacy Parsing
- **Purpose**: Establish basic NLP capabilities, exact string matching, and the first iterations of `SemanticFrame`.
- **Goals**: Parse simple commands deterministically.
- **Status**: Deprecated/Superceded by U6.

### U6: Language Engine (V3 Architecture)
- **Purpose**: Introduce a robust, multi-layered parsing pipeline.
- **Goals**: Implement the Specialist Cascade (Grammar -> Neural -> Deliberative), deterministic normalizers, and transition the canonical payload to `UnderstandingBatch`.
- **Exit Criteria**: Strict separation of extraction and evaluation; all capabilities addressable via deterministic models before falling back to neural.
- **Status**: **COMPLETE & FROZEN**.

### U7: Context Subsystem
- **Purpose**: Introduce stateless, deterministic context resolution.
- **Goals**: Resolve temporal anchors, pronouns, and implicit context using explicit Dialogue State.
- **Exit Criteria**: Context Resolver operates independently on the Workspace without Executive meddling, relying entirely on Strategy interfaces.
- **Status**: **COMPLETE & FROZEN**. (Note: Currently utilizes a transparent forwarding layer for V3 compatibility).

---

## Updated Roadmap (Upcoming Milestones)

### U1-U7 ✅ Complete
- All foundational parsing, V3 architecture, and Context (U7.5) migrations are complete and frozen.

### U8 ← Current Sprint: Multi-Intent & Composite Understanding
- **Purpose**: Process complex, multi-part utterances natively within the cognitive pipeline.
- **Goals**: 
  - Enhance the existing `splitter` to support cross-intent semantic binding (e.g., "Find the file and delete it").
  - Resolve the $O(2^N)$ backtracking performance issue in the current `deterministicSplitter` by shifting to a pre-segmentation model or parallelized evaluation.
- **Certification Requirements**: A multi-intent command must map securely to multiple independent capability executions without cross-contamination.
- **Exit Criteria**: A single utterance can yield a batch of resolved intents that Reasoning can sequence properly without exponential latency.

#### U8.1: Splitter Refactoring
- Replace backtracking subset evaluation with a deterministic pre-segmentation pipeline.
#### U8.2: Micro-Splitting Conflict Resolution
- Ensure Temporal Extractors handle conjunctions natively without conflicting with utterance-level Splitters.

### U8.5: Raw Input Preservation
- **Purpose**: Ensure that the exact verbatim utterance is retained immutably throughout the pipeline for auditing, reflection, and debugging.
- **Goals**: Attach verifiable cryptographic proofs of the original user input to the final executed actions.
- **Certification Requirements**: Every `TopicResolvedIntent` must trace back to an unmodified `RawInput` via content-addressed storage.
- **Exit Criteria**: The Executive can audit any capability execution against the user's exact words.

### U9: Recovery & Clarification
- **Purpose**: Handle ambiguous, incomplete, or rejected intents gracefully.
- **Goals**: When `Understanding` yields a low-confidence hypothesis or `Context Resolver` encounters an unresolvable pronoun, the subsystem must generate a structured request for clarification.
- **Certification Requirements**: Must not guess destructive operations; must halt the pipeline and request explicit user confirmation.
- **Exit Criteria**: Ambiguous commands trigger a deterministic dialogue clarification loop instead of failure or hallucinated action.

### U10: Adaptive Learning
- **Purpose**: Improve the deterministic grammar and normalizers automatically over time.
- **Goals**: The subsystem should learn from clarification loops (U9) and user corrections to dynamically compile new `ExactKeywordRule` or `PatternRule` entries.
- **Certification Requirements**: Learning must be heavily constrained, logged, and reversible to prevent prompt-injection or gradual degradation of deterministic bounds.
- **Exit Criteria**: Repeatedly corrected neural intents are eventually promoted to deterministic grammar rules automatically.

---

## Roadmap Evolution Summary
- **U1–U6** : Completed (Foundational parsing and V3 Architecture)
- **U7** : Context
- **U7.5** : V3 Context Migration
- **U8** : Multi-Intent & Composite Understanding (Current)
- **U8.5** : Raw Input Preservation
- **U9** : Recovery & Clarification
- **U10** : Learning

*(Note: U-series milestones belong exclusively to the Understanding pillar and do not include Reasoning, Planning, Memory, or other system milestones.)*
