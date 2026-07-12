# IDUN Intelligence Pillar: Executive Functions Version 2.0 (`idun/intelligence/executive`)

**Architecture Version:** `2.0.0-FROZEN`  
**Classification:** Core Executive Control Plane — Phase 5 Upgrade  
**Status:** FULLY IMPLEMENTED & TESTED (`-race` CLEAN)

---

## 1. Purpose & Version 1 → Version 2 Upgrade Summary
The `executive` package coordinates the seven immutable cognitive abilities.
- **Version 1 Heritage Preserved:** All existing V1 structures (`Service`, `WorkflowGraph`, `AttentionGate`, `PriorityEngine`, `BudgetManager`, `AbilityRegistry`) remain intact and are embedded within `ServiceV2`.
- **Version 2.0 Capabilities Added:**
  - **Global Workspace Integration (`Workspace()`):** Symmetric publish/subscribe arbitration across leveled topic channels.
  - **Epistemic Calibration Integration (`Calibration()`):** Evaluates candidate bids using Calibrated Effective Priority ($P_{\text{eff}}$), discounting overconfident modules.
  - **Constitutional Gate Integration (`Constitution()`):** Routes actions targeting `TopicActionExecution` through the Pre-Broadcast Constitutional Action Gate before physical execution.
  - **Multi-Horizon Scheduling (`Horizon`):** Decouples Reflexive ($<15\text{ms}$), Deliberative ($100\text{--}500\text{ms}$), and Background scheduling.
  - **SOAR-Style Content-Blind Impasse Emission:** Automatically publishes `TopicImpasses` events when no candidate bid meets admission threshold $\tau$.

---

## 2. Strict Content Blindness & Separation of Responsibilities
Executive Functions Version 2.0 inspects **only** control-plane Envelope metadata (`Source`, `Topic`, `RawConfidence`, `Urgency`, `CostEstimateUnits`, `PayloadRef`).
Executive **never** dereferences `PayloadRef`, **never** parses language, and **never** performs domain reasoning or planning.
