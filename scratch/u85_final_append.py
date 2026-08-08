import os

# Append to SYSTEM_ARCHITECTURE.md
arch_content = """
## 7. Ingress & Raw Input Preservation (U8.5)

- **Input fidelity is guaranteed relative to the ingress adapter.**
- **Every ingress adapter defines its own atomic unit of input.**
- **`TextInputAdapter` preserves one command exactly.** (As an interactive CLI REPL, the terminal consumes the newline, and a single command line is the atomic unit).
- **Future document-oriented adapters will preserve entire artifacts byte-for-byte.** (e.g., File, Document, PDF, Voice).
- **U8.5 is officially certified and frozen.** The Claim Check Pattern is fully implemented using the Core Storage subsystem, propagating only the `PayloadRef` through the Workspace.
"""

arch_path = r"c:\Projects\idun_v3\reports\SYSTEM_ARCHITECTURE.md"
with open(arch_path, "a", encoding="utf-8") as f:
    f.write("\n" + arch_content + "\n")

# Append to TODO.md
todo_content = """
### U8.5 Follow-ups: Ingress Adapters

- Add DocumentInputAdapter for full document ingestion.
- Add FileInputAdapter for byte-for-byte file preservation.
- Define fidelity guarantees for every ingress adapter.
- Document adapter-specific atomic input semantics.
- Evaluate multiline conversational input separately from document ingestion.
"""

todo_path = r"c:\Projects\idun_v3\TODO.md"
with open(todo_path, "a", encoding="utf-8") as f:
    f.write("\n" + todo_content + "\n")
