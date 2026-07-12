# IDUN Intelligence Infrastructure: Shared Embedding Service (`idun/intelligence/infrastructure/embedding`)

**Architecture Version:** `1.0.0-FROZEN-SPRINT1`  
**Classification:** Shared Computational Substrate — Package 3 of 3  
**Status:** FULLY IMPLEMENTED & TESTED

---

## 1. Purpose
The `embedding` package projects multi-modal content (`text`, `audio`, `structured`) into a **Canonical Single Semantic Vector Space**.

---

## 2. Opaque Semantic Handles (`VectorRef`)
To ensure 20+ year architectural invariance, cognitive abilities **never** inspect float slices or vector dimensionality (`768`, `1536`). 
- Embeddings are returned as opaque handles (`VectorRef string`) referencing immutable payloads stored in `idun/core/storage`.
- Metric comparisons are performed via `EmbeddingService.Similarity(ctx, vectorRefA, vectorRefB) -> float64`.

---

## 3. Key Features
- **Batch Processing:** Concurrent multi-document projection via `EmbedBatch`.
- **SHA-256 Content Caching:** Exact cache hits avoid duplicate GPU/backend vector computations.
- **Interface Segregation:** Cognitive abilities receive only `EmbeddingService`. Host/Kernel monitors receive `TelemetryProvider`.
