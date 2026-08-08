content = """
### Storage Subsystem Enhancements

- Implement TTL (Time-To-Live) and Garbage Collection policies for the Core Storage subsystem to prevent unbounded disk growth from preserved raw inputs.
- Extend the PayloadRef mechanism with metadata capabilities (e.g., MimeType, SourceID) to natively distinguish between Text, Voice, Vision, and binary (PDF) payloads.
- Evaluate a unified artifact indexing strategy for raw inputs once multimedia (images, audio) ingress is fully introduced.
"""

with open(r"c:\Projects\idun_v3\TODO.md", "a", encoding="utf-8") as f:
    f.write("\n" + content + "\n")
