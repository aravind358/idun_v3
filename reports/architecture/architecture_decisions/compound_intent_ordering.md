# ADR: Compound Intent Ordering

## Status
Accepted

## Decision
The Understanding layer preserves only the user's spoken order of goals.
The Planning layer exclusively owns execution order and dependency construction.
Understanding must never encode execution dependencies, workflow logic, or scheduling decisions.

## Current Planning Policy
For Phase 4, Planning constructs a deterministic sequential execution graph by creating dependency edges between consecutive goals in the order they were spoken.
This is an implementation policy of the Planning layer and not part of the Understanding contract.

## Reason
A Semantic Dependency Analyzer has not yet been implemented.
Until Planning can formally determine whether goals are independent, mutually dependent, or safe to execute concurrently, sequential execution provides deterministic and predictable behavior while preserving the user's requested order.

## Future Evolution
Future versions of the Planning layer may:
- Reorder goals when it is safe to do so.
- Execute independent goals in parallel.
- Merge compatible goals into optimized workflows.
- Introduce synchronization points where required.
- Build more sophisticated DAGs based on semantic dependency analysis.

These improvements must occur entirely within the Planning layer.

## Architectural Boundary
Understanding is responsible for:
- Detecting compound intents.
- Splitting utterances into independent goals.
- Preserving the user's spoken order.
- Producing an UnderstandingBatch.

Planning is responsible for:
- Determining execution order.
- Constructing the execution DAG.
- Creating dependencies between goals.
- Optimizing execution strategies.

The Understanding contract must remain unchanged as Planning evolves.

## Compatibility Guarantee
Future enhancements to Planning—including semantic dependency analysis, parallel execution, workflow optimization, or synchronization—must not require changes to:
- SemanticInterpretation
- UnderstandingBatch
- The Understanding pipeline
- The frozen Cognitive API Specification

This ADR formalizes the separation of responsibilities between Understanding and Planning and ensures future execution strategies can evolve without impacting the Understanding layer.