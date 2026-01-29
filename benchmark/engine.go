package benchmark

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	Device              string
	ReadTPIOSize        string // e.g., "128k"
	WriteTPIOSize       string // e.g., "64k"
	ReadIOPSIOSize      string // e.g., "4k"
	WriteIOPSIOSize     string // e.g., "4k"
	ReadTPThreads       int
	WriteTPThreads      int
	ReadIOPSThreads     int
	WriteIOPSThreads    int
	ReadTPDuration      int
	WriteTPDuration     int
	ReadIOPSDuration    int
	WriteIOPSDuration   int
}

type Results struct {
	ReadThroughputMBps      float64
	ReadThroughputIOPS      float64
	ReadTPLatencyMs         float64
	ReadTPThreads           int
	ReadTPDuration          int
	WriteThroughputMBps     float64
	WriteThroughputIOPS     float64
	WriteTPLatencyMs        float64
	WriteTPThreads          int
	WriteTPDuration         int
	ReadIOPSThroughputMBps  float64
	ReadIOPS                float64
	ReadIOPSLatencyMs       float64
	ReadIOPSThreads         int
	ReadIOPSDuration        int
	WriteIOPSThroughputMBps float64
	WriteIOPS               float64
	WriteIOPSLatencyMs      float64
	WriteIOPSThreads        int
	WriteIOPSDuration       int
	TestDate                time.Time
	Config                  Config
}

type Engine struct {
	stopChan chan bool
	stopped  bool
}

func NewEngine() *Engine {
	return &Engine{
		stopChan: make(chan bool),
		stopped:  false,
	}
}

func (e *Engine) Stop() {
	if !e.stopped {
		e.stopped = true
		close(e.stopChan)
	}
}

func (e *Engine) Reset() {
	e.stopChan = make(chan bool)
	e.stopped = false
}

// ParseSize converts size strings like "4k", "128k", "1m" to bytes
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToLower(sizeStr))
	
	multiplier := int64(1)
	numStr := sizeStr
	
	if strings.HasSuffix(sizeStr, "k") {
		multiplier = 1024
		numStr = strings.TrimSuffix(sizeStr, "k")
	} else if strings.HasSuffix(sizeStr, "m") {
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "m")
	} else if strings.HasSuffix(sizeStr, "g") {
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "g")
	}
	
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	
	return num * multiplier, nil
}

func (e *Engine) PrepDevice(devicePath string, progressCallback func(string)) error {
	progressCallback("Opening device...")
	
	file, err := openDeviceForWrite(devicePath)
	if err != nil {
		return fmt.Errorf("failed to open device: %v", err)
	}
	defer file.Close()
	
	// Get device size
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to determine device size: %v", err)
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to start: %v", err)
	}
	
	progressCallback(fmt.Sprintf("Device size: %.2f GB", float64(size)/(1024*1024*1024)))
	
	// Write in 1MB chunks
	chunkSize := int64(1024 * 1024)
	buffer := make([]byte, chunkSize)
	written := int64(0)
	lastProgress := 0
	
	for written < size {
		// Fill buffer with random data
		_, err := rand.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to generate random data: %v", err)
		}
		
		// Write to device
		n, err := file.Write(buffer)
		if err != nil {
			return fmt.Errorf("write error: %v", err)
		}
		
		written += int64(n)
		progress := int(float64(written) / float64(size) * 100)
		
		if progress > lastProgress && progress%10 == 0 {
			progressCallback(fmt.Sprintf("Progress: %d%% (%.2f GB / %.2f GB)", 
				progress, 
				float64(written)/(1024*1024*1024),
				float64(size)/(1024*1024*1024)))
			lastProgress = progress
		}
	}
	
	// Sync to ensure data is written
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync error: %v", err)
	}
	
	progressCallback("Prep complete: 100%")
	return nil
}

func (e *Engine) RunBenchmark(config Config, progressCallback func(string)) (*Results, error) {
	// Reset stop state for new run
	e.Reset()
	
	results := &Results{
		TestDate: time.Now(),
		Config:   config,
	}
	
	// Run each test
	tests := []struct {
		name         string
		ioSize       string
		isWrite      bool
		threads      int
		duration     int
		throughputFn func(float64)
		iopsFn       func(float64)
		latencyFn    func(float64)
		threadsFn    func(int)
		durationFn   func(int)
	}{
		{"Read Throughput", config.ReadTPIOSize, false, config.ReadTPThreads, config.ReadTPDuration,
			func(v float64) { results.ReadThroughputMBps = v },
			func(v float64) { results.ReadThroughputIOPS = v },
			func(v float64) { results.ReadTPLatencyMs = v },
			func(v int) { results.ReadTPThreads = v },
			func(v int) { results.ReadTPDuration = v }},
		{"Write Throughput", config.WriteTPIOSize, true, config.WriteTPThreads, config.WriteTPDuration,
			func(v float64) { results.WriteThroughputMBps = v },
			func(v float64) { results.WriteThroughputIOPS = v },
			func(v float64) { results.WriteTPLatencyMs = v },
			func(v int) { results.WriteTPThreads = v },
			func(v int) { results.WriteTPDuration = v }},
		{"Read IOPS", config.ReadIOPSIOSize, false, config.ReadIOPSThreads, config.ReadIOPSDuration,
			func(v float64) { results.ReadIOPSThroughputMBps = v },
			func(v float64) { results.ReadIOPS = v },
			func(v float64) { results.ReadIOPSLatencyMs = v },
			func(v int) { results.ReadIOPSThreads = v },
			func(v int) { results.ReadIOPSDuration = v }},
		{"Write IOPS", config.WriteIOPSIOSize, true, config.WriteIOPSThreads, config.WriteIOPSDuration,
			func(v float64) { results.WriteIOPSThroughputMBps = v },
			func(v float64) { results.WriteIOPS = v },
			func(v float64) { results.WriteIOPSLatencyMs = v },
			func(v int) { results.WriteIOPSThreads = v },
			func(v int) { results.WriteIOPSDuration = v }},
	}
	
	for _, test := range tests {
		// Check if stopped
		if e.stopped {
			return nil, fmt.Errorf("benchmark stopped by user")
		}
		
		progressCallback(fmt.Sprintf("\nRunning %s test...", test.name))
		
		ioSizeBytes, err := ParseSize(test.ioSize)
		if err != nil {
			return nil, fmt.Errorf("invalid IO size for %s: %v", test.name, err)
		}
		
		// Store threads and duration for this test
		test.threadsFn(test.threads)
		test.durationFn(test.duration)
		
		throughput, iops, latency, err := e.runSingleTest(config.Device, ioSizeBytes, test.threads, test.duration, test.isWrite, progressCallback)
		if err != nil {
			if e.stopped {
				return nil, fmt.Errorf("benchmark stopped by user")
			}
			return nil, fmt.Errorf("%s test failed: %v", test.name, err)
		}
		
		test.throughputFn(throughput)
		test.iopsFn(iops)
		test.latencyFn(latency)
		
		progressCallback(fmt.Sprintf("%s: %.2f MB/s | %.0f IOPS | %.2f ms", test.name, throughput, iops, latency))
	}
	
	return results, nil
}

// RunSingleTest runs a single benchmark test and returns throughput, IOPS, and latency
func (e *Engine) RunSingleTest(devicePath string, ioSize int64, threads int, duration int, isWrite bool, progressCallback func(string)) (float64, float64, float64, error) {
	return e.runSingleTest(devicePath, ioSize, threads, duration, isWrite, progressCallback)
}

func (e *Engine) runSingleTest(devicePath string, ioSize int64, threads int, duration int, isWrite bool, progressCallback func(string)) (float64, float64, float64, error) {
	var totalOps atomic.Int64
	var totalBytes atomic.Int64
	var totalLatencyNs atomic.Int64
	var wg sync.WaitGroup
	stopChan := make(chan bool)
	
	// Start worker threads
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			e.worker(devicePath, ioSize, isWrite, &totalOps, &totalBytes, &totalLatencyNs, stopChan)
		}(i)
	}
	
	// Run for specified duration
	startTime := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	go func() {
		lastOps := int64(0)
		lastBytes := int64(0)
		lastLatency := int64(0)
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime).Seconds()
				currentOps := totalOps.Load()
				currentBytes := totalBytes.Load()
				currentLatency := totalLatencyNs.Load()
				
				opsPerSec := float64(currentOps-lastOps) / 1.0
				bytesPerSec := float64(currentBytes-lastBytes) / 1.0
				latencyDelta := currentLatency - lastLatency
				opsDelta := currentOps - lastOps
				
				lastOps = currentOps
				lastBytes = currentBytes
				lastLatency = currentLatency
				
				// Calculate all three metrics
				mbps := bytesPerSec / (1024 * 1024)
				iops := opsPerSec
				latencyMs := float64(0)
				if opsDelta > 0 {
					latencyMs = float64(latencyDelta) / float64(opsDelta) / 1000000.0
				}
				
				progressCallback(fmt.Sprintf("  %.0fs: %.2f MB/s | %.0f IOPS | %.2f ms", elapsed, mbps, iops, latencyMs))
			case <-stopChan:
				return
			case <-e.stopChan:
				return
			}
		}
	}()
	
	// Wait for duration or stop signal
	select {
	case <-time.After(time.Duration(duration) * time.Second):
		// Normal completion
	case <-e.stopChan:
		// User stopped the test
		progressCallback("  Test stopped by user")
	}
	
	close(stopChan)
	wg.Wait()
	
	elapsed := time.Since(startTime).Seconds()
	ops := totalOps.Load()
	bytes := totalBytes.Load()
	latencyNs := totalLatencyNs.Load()
	
	// Calculate average latency in milliseconds
	avgLatencyMs := float64(0)
	if ops > 0 {
		avgLatencyMs = float64(latencyNs) / float64(ops) / 1000000.0
	}
	
	// Return all three metrics: throughput, IOPS, latency
	throughputMBps := (float64(bytes) / elapsed) / (1024 * 1024)
	iops := float64(ops) / elapsed
	
	return throughputMBps, iops, avgLatencyMs, nil
}

func (e *Engine) worker(devicePath string, ioSize int64, isWrite bool, totalOps *atomic.Int64, totalBytes *atomic.Int64, totalLatencyNs *atomic.Int64, stopChan chan bool) {
	var file *os.File
	var err error
	
	if isWrite {
		file, err = openDeviceForWrite(devicePath)
	} else {
		file, err = openDeviceForRead(devicePath)
	}
	
	if err != nil {
		return
	}
	defer file.Close()
	
	// Get device size for random seeks
	deviceSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}
	
	buffer := make([]byte, ioSize)
	
	// Pre-fill buffer with random data for writes
	if isWrite {
		rand.Read(buffer)
	}
	
	for {
		select {
		case <-stopChan:
			return
		default:
			// Random offset aligned to IO size
			maxOffset := (deviceSize / ioSize) - 1
			if maxOffset <= 0 {
				maxOffset = 1
			}
			randomBlock := (time.Now().UnixNano() % maxOffset)
			offset := randomBlock * ioSize
			
			_, err := file.Seek(offset, io.SeekStart)
			if err != nil {
				continue
			}
			
			// Measure latency
			opStart := time.Now()
			
			if isWrite {
				n, err := file.Write(buffer)
				if err == nil && n > 0 {
					latency := time.Since(opStart).Nanoseconds()
					totalOps.Add(1)
					totalBytes.Add(int64(n))
					totalLatencyNs.Add(latency)
				}
			} else {
				n, err := file.Read(buffer)
				if err == nil && n > 0 {
					latency := time.Since(opStart).Nanoseconds()
					totalOps.Add(1)
					totalBytes.Add(int64(n))
					totalLatencyNs.Add(latency)
				}
			}
		}
	}
}

// Platform-specific device opening
func openDeviceForRead(devicePath string) (*os.File, error) {
	if runtime.GOOS == "windows" {
		return openDeviceWindows(devicePath, false)
	}
	return openDeviceLinux(devicePath, false)
}

func openDeviceForWrite(devicePath string) (*os.File, error) {
	if runtime.GOOS == "windows" {
		return openDeviceWindows(devicePath, true)
	}
	return openDeviceLinux(devicePath, true)
}
