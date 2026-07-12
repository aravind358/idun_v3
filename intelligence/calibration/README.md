# IDUN Intelligence Pillar: Epistemic Calibration System (`idun/intelligence/calibration`)

**Architecture Version:** `2.0.0-FROZEN`  
**Classification:** Core Cognitive Communication Layer — Phase 3  
**Status:** FULLY IMPLEMENTED & TESTED

---

## 1. Purpose
The `calibration` package implements the **Epistemic Calibration System** specified in Architecture Version 2.0. It tracks historical epistemic accuracy per module source and workspace topic, discounting uncalibrated over-confidence without requiring Executive Functions to inspect semantic payloads.

---

## 2. Key Features
- **Historical Trust Ledger:** Maintains empirical audit records (`AuditRecord`) comparing reported confidence against actual task accuracy.
- **Pluggable Weight Strategy (`WeightStrategy`):** Decouples calibration algorithms from Executive Functions.
- **Calibrated Effective Priority (`CalibrateEnvelope`):** Directly computes $P_{\text{eff}}$ for bidding envelopes.
