# IDUN Intelligence Pillar: Cognitive Communication Substrate (`idun/intelligence/communication`)

**Architecture Version:** `2.0.0-FROZEN`  
**Classification:** Core Cognitive Communication Layer — Phase 1  
**Status:** FULLY IMPLEMENTED & TESTED

---

## 1. Purpose
The `communication` package establishes the immutable control-plane message wrapper (`Envelope`) and leveled semantic channels (`TopicID`) that cognitive abilities use to publish and subscribe across the Global Workspace.

---

## 2. Architectural Invariants
1. **Content-Blind Control Plane:** `Envelope` carries only control metadata (`Source`, `Topic`, `RawConfidence`, `Urgency`, `CostEstimateUnits`, `PayloadRef`).
2. **Payload Decoupling:** Executive Functions and the communication bus **never** dereference, inspect, or parse the contents of `PayloadRef`. Heavy payloads reside in `idun/core/storage`.
3. **Calibrated Effective Priority:** Bids are evaluated using Calibrated Effective Priority ($P_{\text{eff}}$), combining confidence, calibration trust weights, urgency, and budget cost.
