
## Executable Organization Rule
The repository root must not contain standalone executable Go programs (package main) unless the repository itself is a single-binary application.
Production entry points belong under cmd/.
Internal developer utilities belong under tools/.
This rule ensures go build ./... succeeds consistently and keeps the project aligned with standard Go project structure.

## Engineering Documentation Rule

Every engineering topic must have exactly one authoritative document.
Future work updates these documents instead of creating duplicate reports.

Examples:
- one Cognitive API Specification
- one Understanding Coverage Matrix
- one Tier 2 Cognitive Intelligence Audit
- one Build Baseline Verification

### Audit Documentation Rule
All engineering audits, coverage reports, implementation verification reports, architecture reviews, maturity assessments, and baseline verification reports must live under reports/audits/.
Historical audits belong in reports/audits/legacy/.

Any future system audit should update `reports/audits/current/system_audit_report.md` rather than creating a new standalone report.
Build and runtime health are maintained in `build_baseline_verification.md`.
Overall cognitive maturity is maintained in `tier_2_cognitive_intelligence_audit.md`.

### Walkthrough Rule
Each implementation phase has exactly one walkthrough.
Examples: phase_4b1.md, phase_4b2.md, phase_4b3.md.

### Trace Rule
Separate traces into:
- 
raw/
- parsed/
- 
results/
Generated trace artifacts should never be mixed with engineering documentation.

### Architecture Rule
Architecture specifications, semantic contracts, ADRs, and architecture requests belong under 
reports/architecture/.

### Archive Rule
The 
reports/archive/ directory is reserved for deprecated, obsolete, or intentionally retired documentation.
It must not be used for active engineering documents or historical V2 audits.



## Semantic Ontology Rule

The semantic ontology is cumulative.

Existing semantic object definitions must remain stable unless an architectural defect is identified.

New capabilities must extend the ontology by introducing new semantic object types rather than modifying existing ones.

The semantic ontology defines the vocabulary of the Understanding layer and remains independent of any specific extractor implementation.

## Cognitive Enrichment Principle

**Each Understanding phase only enriches the representation produced by the previous phase.**
- No phase modifies or reinterprets information produced by earlier phases.
- No phase performs responsibilities owned by later cognitive layers.

## Temporal Processing Rule

The Understanding layer owns:
- Temporal classification
- Deterministic temporal normalization

The Core Time service owns:
- Calendar utilities
- Clock utilities
- Timezone management
- Generic date/time calculations

The Reasoning layer owns:
- Contextual interpretation
- Time grounding
- Context-dependent temporal reasoning

Responsibilities must never cross these boundaries.

## Architecture Documentation Rule

Permanent architectural principles and accepted engineering decisions must be documented as Architecture Decision Records (ADRs) under reports/architecture/architecture_decisions/.

Every ADR must also be listed in architecture/architecture_decisions/README.md.

The top-level reports/README.md should remain the master documentation index and link to the architecture indexes rather than attempting to duplicate their contents.

## Capability Behavior Rule

A single capability may support multiple semantic intents when those intents operate on the same underlying resource.

Behavior differences must be determined using the preserved semantic intent, not by creating duplicate capabilities.

Example:

query_time
        │
        ▼
sys-time-1
        │
        ▼
ResponseType = time
query_date
        │
        ▼
sys-time-1
        │
        ▼
ResponseType = date

This preserves:
- capability reuse
- semantic separation
- presentation flexibility

without multiplying capabilities unnecessarily.

## Presentation Formatting Rule

Capabilities must return structured domain data, not presentation-formatted strings.

Human-readable formatting (dates, times, numbers, units, etc.) belongs exclusively to the Presentation layer.

The Realization layer consumes already formatted content and must not become responsible for domain-specific formatting.

This rule generalizes beyond time and date. The same principle will apply later to:
- file sizes (KB, MB, GB)
- percentages
- temperatures
- battery levels
- memory usage
- currency
- durations

## Legacy Adapter Rule

Legacy adapters may translate parameter names between cognitive semantics and legacy capability interfaces.

They must never:
- modify intent
- modify slots
- infer values
- compute missing parameters

They may only perform deterministic field translation.

Examples:
operand1 -> a
operand2 -> b
operator "+" -> operation "add"

No other semantic transformation is permitted.

## Temporal Ownership Rule

The Understanding layer owns temporal extraction and deterministic temporal normalization.

Capabilities must consume normalized temporal values only.

Capabilities must never perform natural-language temporal parsing.

## Temporal Data Preservation Rule

Temporal Composition must never replace or destroy normalized TemporalAnchors.

It may create additional composed temporal artifacts.

Original normalized TemporalAnchors must remain available for:
- downstream reasoning
- auditing
- clarification
- future inference
- explainability
### Temporal Composition Completeness Rule
Temporal Composition may produce a composed timestamp only when sufficient normalized temporal information exists.
If the available temporal information is incomplete, the original normalized TemporalAnchors must be preserved.
Temporal Composition must never:
- invent missing dates,
- invent missing times,
- infer unspecified temporal values,
- or discard user-provided temporal information.
## Legacy Capability Accommodation Rule
The cognitive contract reflects the user's semantic intent.
Legacy capability requirements must be satisfied by deterministic adapter logic whenever possible.
The user should not be forced to speak in terms dictated by legacy capability interfaces.
Adapters may provide documented deterministic defaults for optional legacy parameters.
They must never invent semantic meaning.
## Legacy Note Naming Rule
If a legacy note capability requires a title and the user did not provide one:
- generate a deterministic default title,
- never derive the title from arbitrary content unless that policy is explicitly documented,
- never modify user-provided titles,
- never overwrite existing notes.

## Note Persistence Rule
A note restoration sprint is not complete until:
- semantic extraction succeeds,
- the capability executes successfully,
- persistent storage is verified,
- retrieval returns identical content,
- deletion removes the stored artifact,
- listing reflects the persistent state.

## Legacy Note Naming Rule
If a legacy note capability requires a title and the user did not provide one:
1. generate a deterministic default title,
2. never derive the title from arbitrary content unless that policy is explicitly documented,
3. never modify user-provided titles,
4. never overwrite existing notes.

## Note Persistence Rule
A note restoration sprint is not complete until:
- semantic extraction succeeds,
- the capability executes successfully,
- persistent storage is verified,
- retrieval returns identical content,
- deletion removes the stored artifact,
- listing reflects the persistent state.

## File Ownership Rule
The Understanding layer owns file semantics.
The Executive layer owns parameter translation.
The application capability owns authorization policy.
The native capability owns filesystem execution.
No layer may assume the responsibilities of another.

## Filesystem Security Rule
All authorization decisions for filesystem operations shall occur exclusively within the application capability (app-files-1) through the documented security policy.
Mechanical capabilities must never implement application-specific authorization logic.

## Workspace Resolution Rule
Workspace resolution determines where a request refers.
Authorization determines whether that location may be accessed.
These responsibilities must remain independent.


### System Policy Rule
Application capabilities own authorization and safety evaluation for operating system actions.
Native system capabilities own only mechanical execution.
Native system capabilities must never:
- evaluate safety
- interpret user intent
- require confirmations
- implement business policies

### System Operation Translation Rule
Application capabilities may:
- translate semantic operations into native system operations,
- apply documented security policy,
- perform deterministic parameter translation.

They must never:
- change user intent,
- invent operations,
- bypass documented policy,
- execute native operations directly.

Native capabilities remain purely mechanical.

### Native Capability Purity Rule
Native capabilities are responsible only for executing mechanical operations.

They must never:
- interpret natural language
- perform semantic reasoning
- enforce application security policy
- perform user confirmation
- modify cognitive intent

They may only:
- validate required parameters
- perform platform capability checks
- execute the requested operation
- return structured results

### Authorization Boundary Rule
Application capabilities own authorization decisions.

Platform capability checks are responsible only for verifying runtime capability availability.

Native capabilities execute only authorized mechanical operations.

Authorization responsibilities must never be duplicated across architectural layers.

### Realization Ownership Rule
Capabilities produce semantic facts only. They must never generate user-facing language or formatted responses.

Presentation prepares those semantic facts for realization. It does not decide which realization engine to use.

The RealizationPolicy is the sole component responsible for selecting the realization strategy. Realization-selection logic must never appear inside the Router.

Realization Engines convert structured semantic data into user-facing communication. They do not perform semantic reasoning or modify the meaning of capability output.

The Router orchestrates realization only — it retrieves capability results, builds a PresentationContext, delegates engine selection to the RealizationPolicy, and forwards the realized output. The Router must remain unchanged when the RealizationPolicy implementation is replaced.

The World layer only displays or speaks the realized output produced by a Realization Engine. It does not perform realization, formatting, or semantic interpretation.
