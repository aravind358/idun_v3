import sys
import codecs

file_path = r"C:\Projects\idun_v3\.agents\AGENTS.md"

rule = """
### Authorization Boundary Rule
Application capabilities own authorization decisions.

Platform capability checks are responsible only for verifying runtime capability availability.

Native capabilities execute only authorized mechanical operations.

Authorization responsibilities must never be duplicated across architectural layers.
"""

try:
    with open(file_path, "r", encoding="utf-8") as file:
        content = file.read()
    
    if "Authorization Boundary Rule" not in content:
        with open(file_path, "a", encoding="utf-8") as file:
            file.write(rule)
except Exception as e:
    print(f"Error processing {file_path}: {e}")
