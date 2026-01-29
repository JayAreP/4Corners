//go:build linux
// +build linux

package benchmark

import (
	"fmt"
	"os"
	"syscall"
)

// Stub for cross-compilation
func openDeviceWindows(devicePath string, write bool) (*os.File, error) {
	return nil, nil
}

func openDeviceLinux(devicePath string, write bool) (*os.File, error) {
	flags := os.O_RDONLY | syscall.O_DIRECT
	if write {
		flags = os.O_RDWR | syscall.O_DIRECT
	}
	
	file, err := os.OpenFile(devicePath, flags, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %v", err)
	}
	
	return file, nil
}
