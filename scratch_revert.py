import re

file_path = r"C:\Users\ARAVIND\.gemini\antigravity-ide\brain\7c4fb2f4-19fa-471a-b495-230d5834d673\implementation_plan.md"

with open(file_path, "r", encoding="utf-8") as f:
    text = f.read()

# 1. Remove Certification Scope
text = re.sub(r"## Certification Scope.*?## 1\. Restoration Manifest", "## 1. Restoration Manifest", text, flags=re.DOTALL)

# 2. Remove tag info from Restoration Manifest
text = text.replace("- Repository Commit / Tag used for certification (e.g., `v3-phase4-baseline`)", "- Repository Commit / Tag used for certification")
text = re.sub(r"As part of the final freeze, the exact repository state will be recorded.*?certified baseline\.\n\n", "", text, flags=re.DOTALL)

# 3. Remove Traceability Matrix (section 3)
text = re.sub(r"## 3\. Traceability Matrix.*?## 4\. End-to-End Trace Audit", "## 3. End-to-End Trace Audit", text, flags=re.DOTALL)

# Renumber sections 4 through 11 -> 3 through 10
text = text.replace("## 4. End-to-End Trace Audit", "## 3. End-to-End Trace Audit")
text = text.replace("## 5. Regression Audit", "## 4. Regression Audit")
text = text.replace("## 6. Unresolved Intent Audit", "## 5. Unresolved Intent Audit")
text = text.replace("## 7. Architecture Drift Audit", "## 6. Architecture Drift Audit")
text = text.replace("## 8. Architecture Compliance Audit", "## 7. Architecture Compliance Audit")
text = text.replace("## 9. Legacy Adapter Audit", "## 8. Legacy Adapter Audit")
text = text.replace("## 10. Engineering Rules Audit", "## 9. Engineering Rules Audit")
text = text.replace("## 11. Documentation Consistency Audit", "## 10. Documentation Consistency Audit")

# 4. Remove Architectural Exception Register (section 12)
text = re.sub(r"## 12\. Architectural Exception Register.*?## Final Deliverable", "## Final Deliverable", text, flags=re.DOTALL)

# 5. Fix Final Deliverable list
text = re.sub(r"3\. Traceability Matrix\n", "", text)
text = re.sub(r"13\. Architectural Exception Register\n", "", text)
# renumber list manually for safety in python
list_str = """1. Restoration Manifest
2. Restoration Inventory
3. End-to-End Trace Audit
4. Architecture Compliance Audit
5. Legacy Adapter Audit
6. Engineering Rules Audit
7. Documentation Consistency Audit
8. Regression Audit
9. Unresolved Intent Audit
10. Architecture Drift Audit
11. Security Audit
12. Certification Decision Matrix
13. Final Restoration Certification
14. Change Control Statement"""
text = re.sub(r"1\. Restoration Manifest.*?16\. Change Control Statement", list_str, text, flags=re.DOTALL)

with open(file_path, "w", encoding="utf-8") as f:
    f.write(text)

print("Done")
