//go:build windows
// +build windows

package device

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Stub for cross-compilation
func listDevicesLinux() ([]Device, error) {
	return nil, nil
}

func getDeviceSizeLinux(devicePath string) (int64, error) {
	return 0, nil
}

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procDeviceIoControl = kernel32.NewProc("DeviceIoControl")
)

const (
	IOCTL_DISK_GET_DRIVE_GEOMETRY_EX = 0x700A0
)

type DISK_GEOMETRY_EX struct {
	Geometry         DISK_GEOMETRY
	DiskSize         int64
	Data             [1]byte
}

type DISK_GEOMETRY struct {
	Cylinders         int64
	MediaType         uint32
	TracksPerCylinder uint32
	SectorsPerTrack   uint32
	BytesPerSector    uint32
}

func listDevicesWindows() ([]Device, error) {
	var devices []Device
	
	// Check PhysicalDrive0 through PhysicalDrive9
	for i := 0; i < 10; i++ {
		devicePath := fmt.Sprintf("\\\\.\\PhysicalDrive%d", i)
		
		pathPtr, err := syscall.UTF16PtrFromString(devicePath)
		if err != nil {
			continue
		}
		
		handle, err := syscall.CreateFile(
			pathPtr,
			0, // No access required for listing
			syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
			nil,
			syscall.OPEN_EXISTING,
			0,
			0,
		)
		
		if err != nil {
			continue
		}
		
		size, _ := getDeviceSizeByHandle(handle)
		syscall.CloseHandle(handle)
		
		devices = append(devices, Device{
			Name: fmt.Sprintf("PhysicalDrive%d (%s)", i, FormatSize(size)),
			Path: devicePath,
			Size: size,
		})
	}
	
	return devices, nil
}

func getDeviceSizeWindows(devicePath string) (int64, error) {
	pathPtr, err := syscall.UTF16PtrFromString(devicePath)
	if err != nil {
		return 0, err
	}
	
	handle, err := syscall.CreateFile(
		pathPtr,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	
	if err != nil {
		return 0, err
	}
	defer syscall.CloseHandle(handle)
	
	return getDeviceSizeByHandle(handle)
}

func getDeviceSizeByHandle(handle syscall.Handle) (int64, error) {
	var geometry DISK_GEOMETRY_EX
	var bytesReturned uint32
	
	r1, _, err := procDeviceIoControl.Call(
		uintptr(handle),
		uintptr(IOCTL_DISK_GET_DRIVE_GEOMETRY_EX),
		0,
		0,
		uintptr(unsafe.Pointer(&geometry)),
		uintptr(unsafe.Sizeof(geometry)),
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	
	if r1 == 0 {
		return 0, err
	}
	
	return geometry.DiskSize, nil
}
