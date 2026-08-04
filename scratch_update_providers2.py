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
            
        text = content.decode("utf-8", errors="ignore")
        
        # Remove the messed up mock function
        if "GetBatteryInfo" in text:
            idx = text.find("func (p *MockSystemProvider) GetBatteryInfo")
            if idx != -1:
                text = text[:idx]
            idx2 = text.find("func (p *WindowsSystemProvider) GetBatteryInfo")
            if idx2 != -1:
                text = text[:idx2]
            idx3 = text.find("func (p *LinuxSystemProvider) GetBatteryInfo")
            if idx3 != -1:
                text = text[:idx3]
            idx4 = text.find("func (p *DarwinSystemProvider) GetBatteryInfo")
            if idx4 != -1:
                text = text[:idx4]

        if "mock" in f:
            text += "\nfunc (p *MockProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {\n\treturn map[string]interface{}{\"status\": \"Charging\", \"percentage\": 100}, nil\n}\n"
        else:
            text += "\nfunc (p *NativeProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {\n\treturn map[string]interface{}{\"status\": \"Battery not available on this system.\", \"percentage\": 0}, nil\n}\n"
                
        with open(f, "w", encoding="utf-8") as file:
            file.write(text)
    except Exception as e:
        print(f"Error processing {f}: {e}")
