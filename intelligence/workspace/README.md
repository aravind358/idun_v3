# IDUN Intelligence Pillar: Global Workspace & Leveled Blackboard (`idun/intelligence/workspace`)

**Architecture Version:** `2.0.0-FROZEN`  
**Classification:** Core Cognitive Communication Layer — Phase 2  
**Status:** FULLY IMPLEMENTED & TESTED

---

## 1. Purpose
The `workspace` package implements the **Global Workspace & Leveled Blackboard Engine** specified in Architecture Version 2.0. It provides orthogonal leveled topic channels where cognitive abilities publish and subscribe as symmetric peers.

---

## 2. Key Features
- **Leveled Topic Channels:** Subscriptions are scoped to canonical `TopicID` channels (`perception`, `candidate-plans`, etc.), preventing quadratic broadcast storms.
- **Global Broadcast Escalation (`WithGlobalBroadcast`):** Allows critical safety or constitutional envelopes to reach all subscribers regardless of topic.
- **Thread-Safe Non-Blocking Dispatch:** Lock-free handler invocation avoids deadlocks across concurrent publishers.
- **Bounded Topic Buffers:** Enforces per-topic buffer limits to ensure stable memory over 20+ years.
