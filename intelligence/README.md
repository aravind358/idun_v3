# IDUN Intelligence Pillar (`idun/intelligence`)

**System Pillar:** Intelligence (`idun/intelligence`)  
**Architecture Specification:** Version `1.0.0-FROZEN`  
**Classification:** First-Class Top-Level System Pillar

---

## Architectural Philosophy & Pillar Status

Within IDUN's architecture, **Intelligence is a first-class system pillar equivalent to the Kernel**, not a utility Core Service.

- **Core Services** (`idun/core/`) provide shared, non-cognitive infrastructure (`logger`, `storage`, `memory`, `scheduler`).
- **Intelligence Pillar** (`idun/intelligence/`) is responsible for all cognitive processing, reasoning, planning, decision-making, learning, reflection, and constitutional alignment. It will evolve independently over the lifetime of IDUN (20+ years).

---

## Pillar Directory Layout

```
idun/
├── kernel/                   # IDUN Operating System Kernel
├── core/                     # Infrastructure Core Services (Logger, Storage, Memory, Scheduler)
└── intelligence/             # First-Class Intelligence Pillar
    ├── executive/            # Executive Functions (Attentional Gating, Priority, Workflow Coordinator)
    ├── understanding/        # Perceptual Parsing, Semantic Decoding, Intent Resolution
    ├── reasoning/            # Logical Inference, Deductive/Inductive/Abductive Reasoning
    ├── decision/             # Action Selection, Trade-Off Optimization, Policy Choice
    ├── planning/             # Hierarchical Task Networks (HTN), Multi-Step Goal Scheduling
    ├── learning/             # Pattern Synthesis, Skill Generalization, Episodic Distillation
    ├── reflection/           # Metacognition, Contradiction Auditing, Epistemic Self-Critique
    ├── value/                # Constitutional Alignment & Axiom Safety Verification
    ├── interfaces/           # Shared Intelligence Pillar ABIs & Contracts
    ├── types/                # Shared Intelligence Pillar Domain Types
    └── README.md             # This document
```

---

## Separation of Concerns & Anti-God-Object Rules

1. **Executive Functions (`idun/intelligence/executive`)** coordinates cognitive workflows, attentional triage, and priority bands, but **never** performs domain thinking.
2. **Cognitive Abilities (`understanding`, `reasoning`, `decision`, `planning`, `learning`, `reflection`, `value`)** perform domain-specific cognitive tasks. They communicate strictly through versioned interfaces defined in `executive/interfaces.go` (`ICognitiveAbilityDriver`) and shared contracts.
