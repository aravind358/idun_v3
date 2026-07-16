# API Input Adapter — Future Enhancement (Post Layer 1)

**Architecture Version:** `2.0.0-FROZEN`
**Package:** `idun/world/adapters/api`
**Status:** PLACEHOLDER — NOT IMPLEMENTED IN PHASE 1

---

## Purpose

This directory is reserved for the API Input Adapter, which will accept structured
machine-to-machine requests (HTTP, gRPC, WebSocket, etc.) and convert them into
Interaction artifacts for the World subsystem.

## Planned Responsibilities

- Accept structured payloads from remote machine clients
- Validate API request schema and extract the canonical input payload
- Return an Interaction with `Modality = ModalityAPI` and `Origin = OriginAPI`
- Support authentication metadata passthrough (not evaluated by World)

## Implementation Notes (When Ready)

- Must implement the `world.InputAdapter` interface including `AdapterVersion()` and `AdapterFingerprint()`
- API schema version must be included in the adapter fingerprint for replay determinism
- Large structured payloads must be stored via PayloadStorer
- Authentication and rate-limiting are infrastructure concerns, not World's responsibility
