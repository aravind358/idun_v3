# Semantic Ontology

## Purpose
The semantic ontology defines the canonical vocabulary for the Understanding layer in IDUN V3. It represents the typed semantic objects that IDUN can construct from extracted raw slots.

The ontology is fully extensible but strictly decoupled from the extraction implementation. It categorizes entities by domain but enforces that only capabilities supported by the current Understanding layer will instantiate these objects at runtime.

## Extensible Hierarchy

### People
- **`EntityPerson`**: Represents an individual person (e.g., "John", "Mom").
- **`EntityContact`**: Represents a specific contact entity (e.g., from an address book).
- **`EntityOrganization`**: Represents a company, institution, or group.
- **`EntityRole`**: Represents a position or job function.
- **`EntityProfession`**: Represents an occupational class.

### Geography
- **`EntityLocation`**: Represents a generic geographic location, region, or point of interest.
- **`EntityCountry`**: Represents a sovereign nation.
- **`EntityState`**: Represents a state or province.
- **`EntityCity`**: Represents a municipality.
- **`EntityAddress`**: Represents a street address.
- **`EntityBuilding`**: Represents a specific structure.

### Files
- **`EntityFile`**: Represents a file on a system.
- **`EntityDirectory`**: Represents a folder or directory path.
- **`EntityDocument`**: Represents written content, notes, or descriptive text.
  - **`EntityDocumentTitle`**: (Future) Represents the title or header of a document.
  - **`EntityDocumentBody`**: (Future) Represents the main body content of a document.
- **`EntityArchive`**: Represents a compressed file structure.
- **`EntityExecutable`**: Represents a program or runnable binary.

### Computer / System Resources
- **`EntityApplication`**: Represents software running or installed on the system.
- **`EntityProcess`**: Represents a low-level OS process.
- **`EntityService`**: Represents a background system service.
- **`EntityCommand`**: Represents an OS or shell command.
- **`EntityDevice`**: Represents hardware connected to the computer.
- **`EntitySystemResource`**: Represents core hardware or virtual resources (e.g., cpu, battery, memory, disk).

### Numbers
- **`EntityNumber`**: Represents a pure numeric value.
- **`EntityQuantity`**: Represents a magnitude.
- **`EntityUnit`**: Represents a measurement scale.
- **`EntityPercentage`**: Represents a ratio.
- **`EntityCurrency`**: Represents a monetary value.
- **`EntityExpression`**: (Future) Represents a mathematical or logical formula before evaluation.

### Physical
- **`EntityProduct`**: Represents a physical or commercial item.
- **`EntityFood`**: Represents edible items.
- **`EntityTool`**: Represents an implement.
- **`EntityVehicle`**: Represents transport hardware.
- **`EntityMedicine`**: Represents pharmaceuticals.

### Communication
- **`EntityMessage`**: Represents an email, SMS, or chat message.
- **`EntityConversation`**: Represents a dialogue thread.
- **`EntityPhoneNumber`**: Represents a telecom number.
- **`EntityEmail`**: Represents an email address.

### Internet
- **`EntityURL`**: Represents a web link.
- **`EntityWebsite`**: Represents a web property.
- **`EntityDomain`**: Represents a DNS domain.
- **`EntityIPAddress`**: Represents a network IP.

### AI
- **`EntityPrompt`**: Represents an LLM input sequence.
- **`EntityModel`**: Represents an AI neural network model.
- **`EntityToolCall`**: Represents an agentic tool invocation.
- **`EntityCapability`**: Represents a registered agent skill.
- **`EntityWorkflow`**: Represents a sequence of orchestrated actions.
- **`EntityAgent`**: Represents an autonomous agent.

### Generic
- **`EntityUnknown`**: A fallback for unmapped or unrecognizable entities.
- **`EntityIdentifier`**: Represents a generic ID or UUID.
- **`EntityLabel`**: Represents a short descriptive name.
- **`EntityTag`**: Represents a metadata keyword.
- **`EntityMetadata`**: Represents unstructured key-value data.

---

## Temporal Anchor Types
- **`TempAbsolute`**: Exact chronological points (e.g., dates, explicit times).
- **`TempRelative`**: Chronological offsets (e.g., "tomorrow", "next week").
- **`TempDuration`**: Spans of time (e.g., "for 3 hours").
- **`TempRecurrence`**: Repeating schedules.

## Reference Types
- **`RefPronoun`**: Structural pointers requiring resolution (e.g., "me", "it").
- **`RefDemonstrative`**: Deictic pointers (e.g., "this", "that").
- **`RefDefiniteDescription`**: Descriptive pointers (e.g., "the file").

## Rule of Stability
As defined in the **Semantic Ontology Rule**: The semantic ontology is cumulative. Existing definitions must remain stable unless an architectural defect is identified. New capabilities must extend the ontology by introducing new types rather than modifying existing ones.

## Future Semantic Ontology Roadmap

The ontology should continue evolving cumulatively without breaking existing semantic types.

Recommended long-term structure:

``nEntity

+-- Person
+-- Organization
+-- Location

+-- Physical
¦   +-- Product
¦   +-- Food
¦   +-- Vehicle
¦   +-- Tool
¦   +-- Device

+-- FileSystem
¦   +-- File
¦   +-- Directory
¦   +-- Document
¦       +-- DocumentTitle
¦       +-- DocumentBody

+-- Computer
¦   +-- Application
¦   +-- Process
¦   +-- Service
¦   +-- Command
¦   +-- SystemResource

+-- AI
¦   +-- Prompt
¦   +-- Model
¦   +-- ToolCall
¦   +-- Capability
¦   +-- Workflow
¦   +-- Agent

+-- Communication
¦   +-- Message
¦   +-- Conversation
¦   +-- Contact
¦   +-- PhoneNumber
¦   +-- Email

+-- Numbers
¦   +-- Number
¦   +-- Quantity
¦   +-- Unit
¦   +-- Expression

+-- Generic
    +-- Identifier
    +-- Label
    +-- Tag
    +-- Metadata

## Temporal Anchors

The Understanding layer classifies temporal objects into the following ontology types:
- `ABSOLUTE_DATE`: Fully specified absolute dates (e.g., 2026-08-04)
- `RELATIVE_DATE`: Dates relative to today (e.g., today, tomorrow)
- `RELATIVE_WEEKDAY`: Specific upcoming weekdays (e.g., next Friday)
- `CLOCK_TIME`: Time of day (e.g., 5 PM, noon)
- `RELATIVE_DURATION`: Duration from now (e.g., in 3 hours)
- `TIME_INTERVAL`: A span of time (e.g., next week, this morning)
- `DAYPART`: General parts of a day (e.g., morning, evening)
- `RECURRENCE`: Repeating schedules (e.g., every day)

## Cognitive Enrichment Principle

**Each Understanding phase only enriches the representation produced by the previous phase.**
- No phase modifies or reinterprets information produced by earlier phases.
- No phase performs responsibilities owned by later cognitive layers.
