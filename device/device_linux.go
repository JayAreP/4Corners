//go:build linux
// +build linux

package device

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	BLKGETSIZE64 = 0x80081272 // Linux ioctl to get block device size
)

// Stub for cross-compilation
func listDevicesWindows() ([]Device, error) {
	return nil, nil
}

func getDeviceSizeWindows(devicePath string) (int64, error) {
	return 0, nil
}

func listDevicesLinux() ([]Device, error) {
	var devices []Device
	
	// List block devices from /sys/block
	entries, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("failed to read /sys/block: %v", err)
	}
	
	for _, entry := range entries {
		name := entry.Name()
		
		// Skip loop devices, ram devices, etc.
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}
		
		devicePath := filepath.Join("/dev", name)
		
		// Try to get size
		sizePath := filepath.Join("/sys/block", name, "size")
		sizeData, err := ioutil.ReadFile(sizePath)
		if err != nil {
			continue
		}
		
		// Size is in 512-byte sectors
		sectors, err := strconv.ParseInt(strings.TrimSpace(string(sizeData)), 10, 64)
		if err != nil {
			continue
		}
		size := sectors * 512
		
		// Check if we can access the device
		if _, err := os.Stat(devicePath); err != nil {
			continue
		}
		
		devices = append(devices, Device{
			Name: fmt.Sprintf("%s (%s)", name, FormatSize(size)),
			Path: devicePath,
			Size: size,
		})
	}
	
	return devices, nil
}

func getDeviceSizeLinux(devicePath string) (int64, error) {
	file, err := os.Open(devicePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	var size int64
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), BLKGETSIZE64, uintptr(unsafe.Pointer(&size)))
	if errno != 0 {
		// Fallback to seeking
		size, err = file.Seek(0, 2)
		if err != nil {
			return 0, err
		}
	}
	
	return size, nil
}
