# Phase 4B.3: Semantic Extraction Design

## 1. Goal
Convert raw slots (strings extracted deterministically via grammar patterns) into typed semantic objects without performing temporal normalization, spatial grounding, global memory lookup, or deliberative reasoning. 

The scope of this phase is purely localized semantic identification. Raw slots become semantic meaning.

## 2. Principle and Pipeline Boundary
This design must preserve the established single-responsibility pipeline:
```
User Input
        │
Grammar Recognition      (Phase 4B.1)
        │
Intent Detection         (Phase 4B.1)
        │
Raw Slot Extraction      (Phase 4B.2)
        │
Semantic Extraction      (Phase 4B.3 - THIS PHASE)
        │
Semantic Objects         (Output of Phase 4B.3)
        │
Reasoning                (Future)
```

**Boundary Rules:**
- Semantic Extraction may add categorical meaning (e.g., this string is a `Person`).
- Semantic Extraction must **never** normalize dates (e.g., "tomorrow" remains "tomorrow", not an ISO8601 string).
- Semantic Extraction must **never** canonicalize entities using external state or memory (e.g., "John" remains `EntityPerson{Surface: "John"}`, not `EntityPerson{ID: "usr_123"}`).
- Semantic Extraction must **never** perform compound intent splitting or LLM inference.

## 3. Scope of Semantic Extraction

The extraction phase maps slots (and sub-strings within slots) to the following types:

### 3.1 Entities
- **Person** (`EntityPerson`): e.g., "John", "Sarah", "mom".
- **Location** (`EntityLocation`): e.g., "Tokyo", "New York", "home".
- **Product** (`EntityProduct`): e.g., "Milk", "shoes".
- **File** (`EntityFile`): e.g., "report.pdf", "C:/Projects/notes.txt".
- **Document** (`EntityDocument`): e.g., "Shopping List", "Ideas". (For notes/titles).
- **Directory** (`EntityDirectory`): e.g., "C:/archive", "/var/log".
- **Quantity** (`EntityQuantity`): e.g., "100 kg", "5 liters".
- **Number** (`EntityNumber`): e.g., "100", "5". (Often used as operands in calculation).
- **Unit** (`EntityUnit`): e.g., "kg", "liters".

### 3.2 Temporal Anchors (Semantic Only)
- **Date**: e.g., "tomorrow", "next Monday", "Friday".
- **Time**: e.g., "5 PM", "9 AM".
- **Duration**: e.g., "3 hours", "15 minutes".
*(Note: These will be typed as `TemporalAnchor` but the raw surface text remains unchanged. Normalization belongs to a future phase.)*

### 3.3 References
- **Pronouns/Pointers** (`Reference`): e.g., "this", "that", "it", "them", "me", "us".

## 4. Implementation Strategy

To prevent a monolithic `extractors.go` file from becoming unmaintainable, semantic extractors should be modularized. 

### 4.1 Interface Design
A new interface will govern semantic extraction passes:

```go
type SemanticExtractor interface {
    Extract(hyp *Hypothesis) error
}
```

### 4.2 Targeted Extractors
Instead of one massive switch block based on Intent, extractors should be specialized domain pipelines:

- `EntityExtractor`: Scans all slots for recognizable generic entities (e.g., mapping a `person` slot to `EntityPerson`, or finding a person name hidden inside a `task` string).
- `FileExtractor`: Converts `filename`, `path`, and `directory` slots into `EntityFile` and `EntityDirectory`.
- `MathExtractor`: Converts `operand1` and `operand2` slots into `EntityNumber`.
- `TemporalExtractor`: Converts `date`, `time`, and `duration` slots into `TemporalAnchor` structs.
- `ReferenceExtractor`: Identifies pointers like "me", "that", "it" from target/source slots and populates `References`.

### 4.3 Pipeline Integration
The Orchestrator will run these extractors sequentially over the primary hypothesis produced by the Grammar Specialist:

```go
func (o *Orchestrator) Interpret(ctx context.Context, env Envelope) (SemanticFrame, error) {
    // 1. Evaluate grammar (yields Intent + Raw Slots)
    hyp, matched := o.grammar.Evaluate(normText)
    
    // 2. Semantic Extraction (yields Entities, Refs, Anchors)
    if matched {
        for _, extractor := range o.extractors {
            extractor.Extract(&hyp)
        }
    }
    // ...
}
```

## 5. Non-Goals
To reiterate, the following are strictly prohibited in Phase 4B.3:
- Temporal normalization
- Memory lookup
- Entity resolution (canonicalization)
- Reasoning, Planning, Decision making
- Neural / LLM parsing
- Compound intents
