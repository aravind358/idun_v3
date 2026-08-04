import sys
import re

file_path = r"C:\Projects\idun_v3\TODO.md"

todo_addition = """
## Post-Restoration Governance Enhancements (Phase 5.x)

### Phase 5.x — Restoration Traceability Matrix
**Objective**: Introduce a formal Traceability Matrix linking every restored artifact across the architecture.

**Tasks**:
- Generate a Traceability Matrix linking:
  - Grammar Rules
  - Intents
  - Planning Mappings
  - Application Capabilities
  - Native Capabilities
  - Verification Status
- Detect orphaned or undocumented restoration artifacts.
- Use the matrix for future architectural audits and regression analysis.

### Phase 5.x — Architectural Exception Register
**Objective**: Maintain a centralized register of intentional architectural limitations.

**Tasks**:
For each exception, record:
- Component
- Reason
- Classification
- Future Phase
- Status

*Example entries*:
- Complex calculator expressions
- Reminder scheduler RFC3339 limitation
- Future temporal enhancements
- Future Unified Policy Engine migration

This register should clearly distinguish intentional MVP limitations from implementation defects.

### Phase 5.x — Repository Certification Baseline
**Objective**: Strengthen long-term auditability by introducing repository-level certification baselines.

**Tasks**:
- Create permanent Git tags for certified architectural baselines.
- Record repository commit hashes in certification artifacts.
- Allow future audits to compare against exact certified repository snapshots.

### Phase 5.x — Change Control Framework
**Objective**: Formalize architectural governance after restoration.

**Tasks**:
Introduce documented change-control procedures covering:
- Architectural review requirements
- Component impact analysis
- Baseline modification process
- Certification document revision policy

This expands the current Change Control Statement into a complete governance framework.

### Phase 5.x — Long-Term Certification Metrics
**Objective**: Extend architectural certification beyond the restoration effort.

Future certification metrics may include:
- Traceability completeness
- Architectural debt
- Documentation coverage
- Engineering rule coverage
- Governance compliance
- Long-term architectural drift analysis

### Deferred Rationale
These governance enhancements improve long-term maintainability, auditability, and architectural evolution. However, they are not prerequisites for certifying the deterministic restoration completed during Phase 4.

Sprint 7 already verifies:
- Restoration correctness
- Architecture compliance
- Runtime behavior
- Documentation consistency
- Engineering rule compliance
- Security boundaries
- Regression stability

That evidence is sufficient to certify and freeze the Phase 4 restoration baseline.

These governance enhancements are therefore intentionally scheduled for Phase 5.x, where the project transitions from restoration to long-term architectural evolution.
"""

try:
    with open(file_path, "r", encoding="utf-8") as file:
        content = file.read()
    
    if "Post-Restoration Governance Enhancements" not in content:
        with open(file_path, "a", encoding="utf-8") as file:
            file.write(todo_addition)
except Exception as e:
    print(f"Error processing {file_path}: {e}")
