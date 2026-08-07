# IDUN Cognitive Pipeline Architecture

This document serves as the canonical reference for the end-to-end cognitive pipeline in IDUN V3, illustrating how data flows from raw perception to validated action.

## Core Design Principles
1. **Stateless Stages**: Each cognitive subsystem operates completely statelessly. They consume an event from the Workspace, process it deterministically, and publish a new event.
2. **Orchestrator Purity**: The Executive is a pure orchestrator. It executes validated capabilities but performs NO cognitive interpretation or semantic binding.
3. **Canonical Payload Evolution**: Unstructured input progressively transforms into a highly structured `UnderstandingBatch`, which is iteratively refined by downstream modules.

## Pipeline Flowchart

```mermaid
graph TD
    A[Perception Subsystem] -->|TopicRawInput| B(Understanding V3 Orchestrator)
    
    subgraph Understanding Subsystem (U8)
        B --> PSD[Protected Span Detection]
        PSD --> SP[O(N) Splitter & Connector Registry]
        SP --> C[Specialist Cascade: Grammar, Neural, Deliberative]
        C --> D[Normalizers]
        D --> E[Semantic Extractors]
        E --> F[UnderstandingBatch (Intent Order Preserved)]
    end
    
    F -->|TopicUserIntent| G(Context Resolver)
    
    subgraph Context Subsystem (U7.5)
        G --> H[Pronoun Strategy]
        G --> I[Ellipsis Strategy]
        G --> J[Temporal Strategy]
        G --> K[Confirmation Strategy]
        H & I & J & K --> L[Modified UnderstandingBatch]
    end
    
    L -->|TopicResolvedIntent| M(Reasoning V3)
    
    subgraph Reasoning Subsystem
        M --> N[Policy Evaluation]
        M --> O[Resource Allocation]
        N & O --> P[TopicEvaluatedOption]
    end
    
    P -->|TopicEvaluatedOption| Q(Executive)
    
    subgraph Executive & Native
        Q --> R[Application Capabilities]
        R --> S[Platform Authorization]
        S --> T[Native Execution]
    end
```

## Subsystem Details

### 1. Understanding (U8)
- **Role**: Parses raw text into semantic intents. Handles multi-intent parsing via a deterministic $O(N)$ Splitter, a Connector Registry, and Protected Span Detection to ensure temporal safety.
- **Data Structure Owned**: `UnderstandingBatch` (preserves original intent order, contains multiple `SemanticInterpretation` objects).
- **Subscribes**: `TopicRawInput`
- **Publishes**: `TopicUserIntent`

### 2. Context Resolver (U7.5)
- **Role**: Resolves implicit references (pronouns, ellipses, relative time) within an `UnderstandingBatch` against the current dialogue state.
- **Data Structure Owned**: Modifies `UnderstandingBatch` (updates `ResolutionStatus` and `References`).
- **Subscribes**: `TopicUserIntent`
- **Publishes**: `TopicResolvedIntent`

### 3. Reasoning
- **Role**: Synthesizes resolved intents into actionable plans, checking policies and evaluating constraints.
- **Data Structure Owned**: `EvaluatedOption`
- **Subscribes**: `TopicResolvedIntent`
- **Publishes**: `TopicEvaluatedOption`

### 4. Executive
- **Role**: Pure execution orchestrator. Dispatches options to specific Application Capabilities and handles errors/budget management.
- **Data Structure Owned**: `ExecutionTrace`
- **Subscribes**: `TopicEvaluatedOption`
- **Publishes**: `TopicSystemEvent` (Execution results)
