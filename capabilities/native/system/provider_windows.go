//go:build windows
// +build windows

package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type NativeProvider struct{}

func NewNativeProvider() *NativeProvider {
	return &NativeProvider{}
}

func (p *NativeProvider) GetSystemInfo(ctx context.Context) (map[string]interface{}, error) {
	hostname, _ := os.Hostname()
	return map[string]interface{}{
		"time":         time.Now().Format(time.RFC3339),
		"date":         time.Now().Format("2006-01-02"),
		"timezone":     time.Now().Location().String(),
		"hostname":     hostname,
		"os":           runtime.GOOS,
		"architecture": runtime.GOARCH,
	}, nil
}

func (p *NativeProvider) GetEnvInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"environment": os.Environ(),
	}, nil
}

func (p *NativeProvider) GetHostInfo(ctx context.Context) (map[string]interface{}, error) {
	hostname, _ := os.Hostname()
	return map[string]interface{}{
		"hostname": hostname,
		"uptime":   "mock_uptime_windows", // TODO: syscall GetTickCount64
	}, nil
}

func (p *NativeProvider) GetCPUInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"cpu_cores": runtime.NumCPU(),
	}, nil
}

func (p *NativeProvider) GetMemoryInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"memory": "mock_memory_info_windows", // TODO: GlobalMemoryStatusEx
	}, nil
}

func (p *NativeProvider) GetDiskInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"disk": "mock_disk_info_windows", // TODO: GetDiskFreeSpaceEx
	}, nil
}

func (p *NativeProvider) ExecutePower(ctx context.Context, action SystemOperation) error {
	var cmd *exec.Cmd
	switch action {
	case OperationShutdown:
		cmd = exec.CommandContext(ctx, "shutdown", "/s", "/t", "0")
	case OperationRestart:
		cmd = exec.CommandContext(ctx, "shutdown", "/r", "/t", "0")
	case OperationSleep:
		cmd = exec.CommandContext(ctx, "rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	case OperationLock:
		cmd = exec.CommandContext(ctx, "rundll32.exe", "user32.dll,LockWorkStation")
	default:
		return fmt.Errorf("unsupported power action '%s' on windows", action)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute %s: %w", action, err)
	}
	return nil
}
