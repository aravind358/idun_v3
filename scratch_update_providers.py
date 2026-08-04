import sys

files = [
    r"C:\Projects\idun_v3\capabilities\native\system\provider_mock.go",
    r"C:\Projects\idun_v3\capabilities\native\system\provider_windows.go",
    r"C:\Projects\idun_v3\capabilities\native\system\provider_linux.go",
    r"C:\Projects\idun_v3\capabilities\native\system\provider_darwin.go",
]

for f in files:
    try:
        with open(f, "rb") as file:
            content = file.read()
        
        # If it was corrupted by utf-16le append, we might have null bytes. We can just clean it up if it's the mock file.
        if "mock" in f:
            content = content.replace(b'\x00', b'')
            
        text = content.decode("utf-8", errors="ignore")
        
        # Remove the messed up mock function
        if "mock" in f and "func (p *MockSystemProvider) GetBatteryInfo" in text:
            idx = text.find("func (p *MockSystemProvider) GetBatteryInfo")
            text = text[:idx]

        if "GetBatteryInfo" not in text:
            if "mock" in f:
                text += "\nfunc (p *MockSystemProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {\n\treturn map[string]interface{}{\"status\": \"Charging\", \"percentage\": 100}, nil\n}\n"
            elif "windows" in f:
                text += "\nfunc (p *WindowsSystemProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {\n\treturn map[string]interface{}{\"status\": \"Battery not available on this system.\", \"percentage\": 0}, nil\n}\n"
            elif "linux" in f:
                text += "\nfunc (p *LinuxSystemProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {\n\treturn map[string]interface{}{\"status\": \"Battery not available on this system.\", \"percentage\": 0}, nil\n}\n"
            elif "darwin" in f:
                text += "\nfunc (p *DarwinSystemProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {\n\treturn map[string]interface{}{\"status\": \"Battery not available on this system.\", \"percentage\": 0}, nil\n}\n"
                
        with open(f, "w", encoding="utf-8") as file:
            file.write(text)
    except Exception as e:
        print(f"Error processing {f}: {e}")
