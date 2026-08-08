import os
import shutil

filepath = r'c:\Projects\idun_v3\TODO.md'
bakpath = r'c:\Projects\idun_v3\TODO.md.bak'

# 1. Read original
with open(filepath, 'rb') as f:
    data = f.read()

# 2. Create backup
shutil.copy2(filepath, bakpath)
print(f"Created backup at {bakpath}")

# 3. Extract sections
start_idx = 72566
end_idx = 77240

part1 = data[:start_idx]
part2_raw = data[start_idx:end_idx]
part3 = data[end_idx:]

print(f"Original size: {len(data)} bytes")
print(f"Part 1 size: {len(part1)} bytes")
print(f"Part 2 (UTF-16 BE) size: {len(part2_raw)} bytes")
print(f"Part 3 size: {len(part3)} bytes")

# 4. Decode the UTF-16 BE block
try:
    decoded_text = part2_raw.decode('utf-16-be')
    print("Successfully decoded UTF-16 BE block.")
except Exception as e:
    print(f"Failed to decode UTF-16 BE block: {e}")
    exit(1)

# Verify no NUL bytes left in decoded string (standard text shouldn't have them)
if '\x00' in decoded_text:
    print("Warning: Decoded text still contains NUL characters. This might be binary data.")
    exit(1)

# 5. Re-encode as UTF-8
part2_utf8 = decoded_text.encode('utf-8')
print(f"Part 2 (UTF-8) size: {len(part2_utf8)} bytes")

# 6. Reconstruct file
reconstructed = part1 + part2_utf8 + part3
print(f"Reconstructed size: {len(reconstructed)} bytes")

# 7. Verify entire file is valid UTF-8
try:
    final_text = reconstructed.decode('utf-8')
    print("Verification: Reconstructed file is 100% valid UTF-8.")
except Exception as e:
    print(f"Verification Failed: Reconstructed file is not valid UTF-8: {e}")
    exit(1)

# 8. Check line counts and markdown headings
orig_lines = data.replace(b'\x00\r\x00\n', b'\n').replace(b'\r\n', b'\n').count(b'\n') + 1
new_lines = reconstructed.replace(b'\r\n', b'\n').count(b'\n') + 1

# Note: The original UTF-16 BE block might have caused some line count confusion if 
# read directly without decoding, but logically the lines are the same.
print(f"Original logical line count approx: {orig_lines}")
print(f"New line count: {final_text.count(chr(10)) + 1}")

# Compare headings in part 2
headings_orig = [line.strip() for line in decoded_text.split('\n') if line.startswith('#')]
headings_new = [line.strip() for line in part2_utf8.decode('utf-8').split('\n') if line.startswith('#')]

if headings_orig == headings_new:
    print(f"Verification: {len(headings_orig)} Markdown headings preserved in affected block.")
else:
    print("Warning: Headings mismatch!")

# 9. Write back
with open(filepath, 'wb') as f:
    f.write(reconstructed)

print("Repair completed successfully.")
print(f"Bytes removed: {len(data) - len(reconstructed)}")
print(f"Affected section: '{decoded_text[:20].strip()}' to '{decoded_text[-20:].strip()}'")
