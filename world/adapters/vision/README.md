# Vision Input Adapter — Future Enhancement (Post Layer 1)

**Architecture Version:** `2.0.0-FROZEN`
**Package:** `idun/world/adapters/vision`
**Status:** PLACEHOLDER — NOT IMPLEMENTED IN PHASE 1

---

## Purpose

This directory is reserved for the Vision Input Adapter, which will convert image,
video frame, or camera stream data into Interaction artifacts for the World subsystem.

## Planned Responsibilities

- Accept image or video data from cameras or image files
- Pre-process frames for content-addressed storage (PayloadRef)
- Return an Interaction with `Modality = ModalityVision` and `Origin = OriginVision`
- Store raw image data via PayloadStorer; never embed large payloads in Interaction fields

## Implementation Notes (When Ready)

- Must implement the `world.InputAdapter` interface including `AdapterVersion()` and `AdapterFingerprint()`
- Understanding — not World — will interpret the visual content
- Adapter fingerprint must include the vision pre-processing pipeline version
- Large images must be stored via PayloadStorer and referenced only by PayloadRef
