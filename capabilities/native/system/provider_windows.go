//go:build windows
// +build windows

package system

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemPower = kernel32.NewProc("GetSystemPowerStatus")
	procGlobalMemory   = kernel32.NewProc("GlobalMemoryStatusEx")
)

type SYSTEM_POWER_STATUS struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

type MEMORYSTATUSEX struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

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
	var msx MEMORYSTATUSEX
	msx.Length = uint32(unsafe.Sizeof(msx))
	ret, _, _ := procGlobalMemory.Call(uintptr(unsafe.Pointer(&msx)))
	if ret == 0 {
		return nil, fmt.Errorf("GlobalMemoryStatusEx failed")
	}

	used := msx.TotalPhys - msx.AvailPhys
	return map[string]interface{}{
		"Total": fmt.Sprintf("%.1f GB", float64(msx.TotalPhys)/(1024*1024*1024)),
		"Used":  fmt.Sprintf("%.1f GB", float64(used)/(1024*1024*1024)),
		"Avail": fmt.Sprintf("%.1f GB", float64(msx.AvailPhys)/(1024*1024*1024)),
		"Load":  msx.MemoryLoad,
	}, nil
}

func (p *NativeProvider) GetDiskInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"disk": "mock_disk_info_windows", // TODO: GetDiskFreeSpaceEx
	}, nil
}

func (p *NativeProvider) ExecutePower(ctx context.Context, action SystemOperation) error {
	// Dangerous Operations Mode is currently deferred.
	// We intentionally do not execute real power operations to prevent accidentally shutting down the host machine during tests.
	return fmt.Errorf("Operation blocked because Dangerous Operations Mode is disabled. Real %s is disabled.", action)
}

func (p *NativeProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {
	var sps SYSTEM_POWER_STATUS
	ret, _, _ := procGetSystemPower.Call(uintptr(unsafe.Pointer(&sps)))
	if ret == 0 {
		return nil, fmt.Errorf("GetSystemPowerStatus failed")
	}

	status := "Unknown"
	charging := false
	if sps.ACLineStatus == 1 {
		status = "Charging"
		charging = true
	} else if sps.ACLineStatus == 0 {
		status = "Discharging"
	}

	if sps.BatteryFlag == 128 || sps.BatteryFlag == 255 {
		status = "Battery not available on this system."
	}

	return map[string]interface{}{
		"status":     status,
		"percentage": sps.BatteryLifePercent,
		"Charging":   charging,
		"Percentage": sps.BatteryLifePercent,
	}, nil
}
