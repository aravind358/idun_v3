import sys
import codecs

file_path = r"C:\Projects\idun_v3\.agents\AGENTS.md"

try:
    with open(file_path, "rb") as file:
        content = file.read()
    
    # Remove null bytes introduced by UTF-16LE
    content = content.replace(b'\x00', b'')
    
    text = content.decode("utf-8", errors="ignore")
    
    with open(file_path, "w", encoding="utf-8") as file:
        file.write(text)
except Exception as e:
    print(f"Error processing {file_path}: {e}")
