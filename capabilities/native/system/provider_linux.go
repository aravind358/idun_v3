//go:build linux
// +build linux

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
		"uptime":   "mock_uptime_linux", // TODO: sysinfo
	}, nil
}

func (p *NativeProvider) GetCPUInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"cpu_cores": runtime.NumCPU(),
	}, nil
}

func (p *NativeProvider) GetMemoryInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"memory": "mock_memory_info_linux", // TODO: read /proc/meminfo
	}, nil
}

func (p *NativeProvider) GetDiskInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"disk": "mock_disk_info_linux", // TODO: statvfs
	}, nil
}

func (p *NativeProvider) ExecutePower(ctx context.Context, action SystemOperation) error {
	var cmd *exec.Cmd
	switch action {
	case OperationShutdown:
		cmd = exec.CommandContext(ctx, "shutdown", "now")
	case OperationRestart:
		cmd = exec.CommandContext(ctx, "reboot")
	case OperationSleep:
		cmd = exec.CommandContext(ctx, "systemctl", "suspend")
	case OperationLock:
		cmd = exec.CommandContext(ctx, "loginctl", "lock-session")
	default:
		return fmt.Errorf("unsupported power action '%s' on linux", action)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute %s: %w", action, err)
	}
	return nil
}


func (p *NativeProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"status": "Battery not available on this system.", "percentage": 0}, nil
}
