# Voice Input Adapter — Future Enhancement (Post Layer 1)

**Architecture Version:** `2.0.0-FROZEN`
**Package:** `idun/world/adapters/voice`
**Status:** PLACEHOLDER — NOT IMPLEMENTED IN PHASE 1

---

## Purpose

This directory is reserved for the Voice Input Adapter, which will convert speech-to-text
audio streams into Interaction artifacts for the World subsystem.

## Planned Responsibilities

- Accept audio input from microphone hardware or streaming audio sources
- Convert speech to raw text via a pluggable speech-to-text engine
- Normalize the transcription according to WorldPolicyProfile
- Return an Interaction with `Modality = ModalityVoice` and `Origin = OriginVoice`

## Implementation Notes (When Ready)

- Must implement the `world.InputAdapter` interface including `AdapterVersion()` and `AdapterFingerprint()`
- Must NOT perform any intent interpretation (that belongs to Understanding)
- Must be content-blind: pass raw transcription to World, not semantic labels
- Adapter fingerprint must change if the speech-to-text engine version changes (for replay determinism)

## See Also

- [World subsystem architecture](../../README.md)
- [TextInputAdapter](../text/adapter.go) — the Phase 1 reference implementation
