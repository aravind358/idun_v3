# IDUN Intelligence Infrastructure: Shared Inference Service (`idun/intelligence/infrastructure/inference`)

**Architecture Version:** `1.0.0-FROZEN-SPRINT1`  
**Classification:** Shared Computational Substrate — Package 2 of 3  
**Status:** FULLY IMPLEMENTED & TESTED

---

## 1. Purpose
The `inference` package provides cognitive abilities (`Understanding`, `Reasoning`, etc.) with a shared, thread-safe, budget-governed computational engine for analytical, generative, and causal inference.

---

## 2. Backend-Agnostic Execution Hints (`ExecutionHints`)
To insulate cognitive abilities from LLM-specific hyperparameters (`temperature`, `max_tokens`) that will become obsolete as AI evolves to Causal World Models, callers specify `ExecutionHints`:
- `ExploratoryVariance`: `0.0` (deterministic logical reasoning) to `1.0` (exploratory association).
- `ComputeBudgetUnits`: Normalized computation limit hint.
- `OutputDetailHint`: Response verbosity hint (`"compact"`, `"standard"`, `"comprehensive"`).

---

## 3. Key Architectural Features
- **Priority-Bucketed Work Queues:** Schedules tasks across `REFLEXIVE` (<15ms SLA), `STANDARD` (<250ms SLA), and `DELIBERATIVE` tiers.
- **Content-Addressed Exact Caching:** SHA-256 caching over `idun/core/storage` guarantees instant 0ms responses for repeated queries.
- **Interface Segregation:** Cognitive abilities receive `InferenceService`. Kernel monitors receive `TelemetryProvider`.
