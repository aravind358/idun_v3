# IDUN Intelligence Infrastructure: Model & Capability Registry (`idun/intelligence/infrastructure/registry`)

**Architecture Version:** `1.0.0-FROZEN-SPRINT1`  
**Classification:** Shared Computational Substrate — Package 1 of 3  
**Status:** FULLY IMPLEMENTED & TESTED

---

## 1. Purpose
The `registry` package provides the single source of truth mapping stable, semantic **Logical Model Identifiers (`ModelID`)** known to cognitive abilities (`Understanding`, `Reasoning`, etc.) to physical **Backend Descriptors (`BackendDescriptor`)** managed by infrastructure and hardware.

---

## 2. Interface Segregation & Access Governance
- **`Resolver` (Read-Only Capability Interface):** Injected into `InferenceService` and `EmbeddingService`. Cognitive abilities and execution engines interact with the registry strictly through `Resolver.Resolve(ctx, modelID)`.
- **`ModelRegistry` (Administrative Interface):** Used by system host/kernel wiring to register, deregister, update health, and rollback model backends.
- **`TelemetryProvider` (Observability Interface):** Exposes operational snapshots strictly to Host/Kernel monitors.

---

## 3. Key Features
- **Thread-Safe Architecture:** Protected by `sync.RWMutex` with atomic telemetry counters.
- **Open Execution Schemes:** Supports `"grpc"`, `"local-bin"`, `"neuromorphic-pci"`, and future hardware execution protocols.
- **Atomic Rollback:** Preserves full version history per `ModelID` with instant rollback via `Rollback(ctx, modelID, version)`.
- **Health Gating:** Unhealthy backends (`HealthUnhealthy`) instantly return `ErrBackendUnavailable` on resolution without blocking execution queues.
