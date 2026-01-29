//go:build windows
// +build windows

package benchmark

import (
	"fmt"
	"os"
	"syscall"
)

// Stub for cross-compilation
func openDeviceLinux(devicePath string, write bool) (*os.File, error) {
	return nil, nil
}

const (
	FILE_FLAG_NO_BUFFERING = 0x20000000
	FILE_FLAG_WRITE_THROUGH = 0x80000000
)

func openDeviceWindows(devicePath string, write bool) (*os.File, error) {
	access := uint32(syscall.GENERIC_READ)
	if write {
		access = syscall.GENERIC_READ | syscall.GENERIC_WRITE
	}
	
	pathPtr, err := syscall.UTF16PtrFromString(devicePath)
	if err != nil {
		return nil, err
	}
	
	handle, err := syscall.CreateFile(
		pathPtr,
		access,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		FILE_FLAG_NO_BUFFERING|FILE_FLAG_WRITE_THROUGH,
		0,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %v", err)
	}
	
	return os.NewFile(uintptr(handle), devicePath), nil
}
