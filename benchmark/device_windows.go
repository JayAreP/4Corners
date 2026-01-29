//go:build windows
// +build windows

package benchmark

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Stub for cross-compilation
func openDeviceLinux(devicePath string, write bool) (*os.File, error) {
	return nil, nil
}

const (
	FILE_FLAG_NO_BUFFERING     = 0x20000000
	FILE_FLAG_WRITE_THROUGH    = 0x80000000
	FILE_FLAG_OVERLAPPED       = 0x40000000
	ERROR_SHARING_VIOLATION    = 32
	ERROR_IO_PENDING           = 997
	IOCTL_DISK_GET_LENGTH_INFO = 0x0007405C
)

type GET_LENGTH_INFORMATION struct {
	Length int64
}

var (
	kernel32DLL         = syscall.NewLazyDLL("kernel32.dll")
	procDeviceIoControl = kernel32DLL.NewProc("DeviceIoControl")
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
		FILE_FLAG_NO_BUFFERING|FILE_FLAG_WRITE_THROUGH|FILE_FLAG_OVERLAPPED,
		0,
	)

	if err != nil {
		// Check for common error conditions
		if err == syscall.ERROR_ACCESS_DENIED {
			return nil, fmt.Errorf("access denied - physical drive access requires Administrator privileges. Please run as Administrator")
		}
		if err == syscall.ERROR_FILE_NOT_FOUND {
			return nil, fmt.Errorf("device not found: %s", devicePath)
		}
		errno, ok := err.(syscall.Errno)
		if ok && errno == ERROR_SHARING_VIOLATION {
			return nil, fmt.Errorf("device is in use by another process: %s", devicePath)
		}
		return nil, fmt.Errorf("failed to open device %s: %v (Run as Administrator if accessing physical drives)", devicePath, err)
	}

	// Validate that the handle is actually usable
	if handle == syscall.InvalidHandle {
		return nil, fmt.Errorf("invalid device handle for %s - may require Administrator privileges", devicePath)
	}

	return os.NewFile(uintptr(handle), devicePath), nil
}

// getDeviceSizeWindows gets the size of a Windows device
// Tries Seek first (works for regular files), falls back to DeviceIoControl for physical drives
func getDeviceSizeWindows(file *os.File) (int64, error) {
	// First try using Seek - this works for regular files
	size, err := file.Seek(0, 2)
	if err == nil {
		// Seek worked, reset position and return size
		file.Seek(0, 0)
		return size, nil
	}

	// Seek failed (expected for physical drives with FILE_FLAG_NO_BUFFERING)
	// Fall back to DeviceIoControl
	handle := syscall.Handle(file.Fd())

	var lengthInfo GET_LENGTH_INFORMATION
	var bytesReturned uint32

	r1, _, ioErr := procDeviceIoControl.Call(
		uintptr(handle),
		uintptr(IOCTL_DISK_GET_LENGTH_INFO),
		0,
		0,
		uintptr(unsafe.Pointer(&lengthInfo)),
		uintptr(unsafe.Sizeof(lengthInfo)),
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)

	if r1 == 0 {
		return 0, fmt.Errorf("DeviceIoControl failed (Seek also failed with: %v): %v", err, ioErr)
	}

	return lengthInfo.Length, nil
}
