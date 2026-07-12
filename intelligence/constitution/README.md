# IDUN Intelligence Pillar: Constitutional Action Gate (`idun/intelligence/constitution`)

**Architecture Version:** `2.0.0-FROZEN`  
**Classification:** Core Cognitive Communication Layer — Phase 4  
**Status:** FULLY IMPLEMENTED & TESTED

---

## 1. Purpose
The `constitution` package implements the **Pre-Broadcast Constitutional Action Gate (`ActionGate`)** specified in Architecture Version 2.0. It intercept all external world-modifying actions (`TopicActionExecution`) before physical/actuator broadcast.

---

## 2. Key Features
- **Independent Invariant Rules (`Rule`):** Constitutional logic resides outside Executive Functions. Executive communicates only via public interfaces (`EvaluateAction`, `InterceptAndPublish`).
- **Cryptographic Approval Signatures:** Approved actions receive a secure HMAC signature (`EvaluationResult.Signature`).
- **Automatic Veto & User Escalation Alerts:** Unconstitutional actions are blocked and diverted to `TopicValueFlags` (high-urgency veto alert) or `TopicUserIntent` (user escalation inquiry).
