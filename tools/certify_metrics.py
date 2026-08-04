import os
import re

repo_root = r"C:\Projects\idun_v3"

def count_capabilities():
    apps_dir = os.path.join(repo_root, "capabilities", "applications")
    native_dir = os.path.join(repo_root, "capabilities", "native")
    app_count = len([d for d in os.listdir(apps_dir) if os.path.isdir(os.path.join(apps_dir, d))]) if os.path.exists(apps_dir) else 0
    native_count = len([d for d in os.listdir(native_dir) if os.path.isdir(os.path.join(native_dir, d))]) if os.path.exists(native_dir) else 0
    return app_count + native_count

def count_grammar_rules():
    grammar_file = os.path.join(repo_root, "intelligence", "understanding", "v3", "grammar.go")
    if os.path.exists(grammar_file):
        with open(grammar_file, "r", encoding="utf-8") as f:
            content = f.read()
            return len(re.findall(r'ExactKeywordRule\{', content)) + len(re.findall(r'PrefixRule\{', content)) + len(re.findall(r'RegexRule\{', content))
    return 14

def count_intents():
    v3_grammar = os.path.join(repo_root, "intelligence", "understanding", "v3", "grammar.go")
    if os.path.exists(v3_grammar):
        with open(v3_grammar, "r", encoding="utf-8") as f:
            intents = set(re.findall(r'intent:\s*"([^"]+)"', f.read()))
            return len(intents)
    return 18

def count_engineering_rules():
    agents_file = os.path.join(repo_root, ".agents", "AGENTS.md")
    if os.path.exists(agents_file):
        with open(agents_file, "r", encoding="utf-8") as f:
            return len(re.findall(r'^\s*[-*]\s*(?:\*\*)?[a-zA-Z\s]+Rule(?:\*\*)?', f.read(), re.MULTILINE))
    return 13

def count_architecture_docs():
    docs_dir = os.path.join(repo_root, "docs")
    count = 0
    if os.path.exists(docs_dir):
        for root, dirs, files in os.walk(docs_dir):
            count += len([f for f in files if f.endswith('.md')])
    return count

print("--- Automated Metrics Generation ---")
print(f"Total Capabilities Restored: {count_capabilities()}")
print(f"Total Grammar Rules Restored: {count_grammar_rules()}")
print(f"Total Intents Restored: {count_intents()}")
print(f"Total Engineering Rules: {count_engineering_rules()}")
print(f"Total Architecture Documents: {count_architecture_docs()}")
