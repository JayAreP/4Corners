package device

import (
	"crypto/rand"
	"fmt"
	"os"
	"runtime"
)

type Device struct {
	Name string
	Path string
	Size int64
}

func ListDevices() ([]Device, error) {
	if runtime.GOOS == "windows" {
		return listDevicesWindows()
	}
	return listDevicesLinux()
}

func GetDeviceSize(devicePath string) (int64, error) {
	if runtime.GOOS == "windows" {
		return getDeviceSizeWindows(devicePath)
	}
	return getDeviceSizeLinux(devicePath)
}

// Common function to format size
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CreateFileDevice creates a file that can be used as a benchmark device
func CreateFileDevice(filePath string, sizeBytes int64, progressCallback func(string)) error {
	progressCallback(fmt.Sprintf("Creating file: %s", filePath))
	progressCallback(fmt.Sprintf("Size: %s", FormatSize(sizeBytes)))
	
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	
	// Write in 10MB chunks to show progress
	chunkSize := int64(10 * 1024 * 1024) // 10 MB
	buffer := make([]byte, chunkSize)
	written := int64(0)
	lastProgress := 0
	
	for written < sizeBytes {
		// Determine how much to write this iteration
		toWrite := chunkSize
		if written+chunkSize > sizeBytes {
			toWrite = sizeBytes - written
			buffer = make([]byte, toWrite)
		}
		
		// Fill buffer with random data
		_, err := rand.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to generate random data: %v", err)
		}
		
		// Write to file
		n, err := file.Write(buffer)
		if err != nil {
			return fmt.Errorf("write error: %v", err)
		}
		
		written += int64(n)
		progress := int(float64(written) / float64(sizeBytes) * 100)
		
		if progress > lastProgress && (progress%10 == 0 || progress == 100) {
			progressCallback(fmt.Sprintf("Progress: %d%% (%s / %s)", 
				progress, 
				FormatSize(written),
				FormatSize(sizeBytes)))
			lastProgress = progress
		}
	}
	
	// Sync to ensure data is written
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync error: %v", err)
	}
	
	progressCallback("File creation complete!")
	return nil
}
