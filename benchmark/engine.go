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
	ReadTPQueueDepth    int
	WriteTPQueueDepth   int
	ReadIOPSQueueDepth  int
	WriteIOPSQueueDepth int
}

type Results struct {
	ReadThroughputMBps      float64
	ReadThroughputIOPS      float64
	ReadTPLatencyMs         float64
	ReadTPThreads           int
	ReadTPDuration          int
	ReadTPQueueDepth        int
	WriteThroughputMBps     float64
	WriteThroughputIOPS     float64
	WriteTPLatencyMs        float64
	WriteTPThreads          int
	WriteTPDuration         int
	WriteTPQueueDepth       int
	ReadIOPSThroughputMBps  float64
	ReadIOPS                float64
	ReadIOPSLatencyMs       float64
	ReadIOPSThreads         int
	ReadIOPSDuration        int
	ReadIOPSQueueDepth      int
	WriteIOPSThroughputMBps float64
	WriteIOPS               float64
	WriteIOPSLatencyMs      float64
	WriteIOPSThreads        int
	WriteIOPSDuration       int
	WriteIOPSQueueDepth     int
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

	// Get device size using platform-specific method
	size, err := getDeviceSize(file)
	if err != nil {
		return fmt.Errorf("failed to determine device size: %v", err)
	}

	progressCallback(fmt.Sprintf("Device size: %.2f GB", float64(size)/(1024*1024*1024)))

	// Use multiple threads for faster prep
	numThreads := 64
	chunkSize := int64(8 * 1024 * 1024) // 8MB chunks for better throughput
	var written atomic.Int64
	lastProgress := atomic.Int32{}

	// Pre-generate random buffers to avoid bottleneck on rand.Read
	randomBuffers := make([][]byte, numThreads)
	for i := 0; i < numThreads; i++ {
		randomBuffers[i] = make([]byte, chunkSize)
		rand.Read(randomBuffers[i])
	}

	// Channel for work distribution
	type writeTask struct {
		offset int64
		size   int64
	}
	tasks := make(chan writeTask, numThreads*4)
	errors := make(chan error, numThreads)
	var wg sync.WaitGroup

	// Worker goroutines
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			buffer := randomBuffers[workerID]

			for task := range tasks {
				// Write to device at offset using pre-generated random data
				n, err := file.WriteAt(buffer[:task.size], task.offset)
				if err != nil {
					select {
					case errors <- fmt.Errorf("write error at offset %d: %v", task.offset, err):
					default:
					}
					return
				}

				// Update progress
				currentWritten := written.Add(int64(n))
				progress := int32(float64(currentWritten) / float64(size) * 100)
				oldProgress := lastProgress.Load()
				if progress > oldProgress && progress%5 == 0 && lastProgress.CompareAndSwap(oldProgress, progress) {
					progressCallback(fmt.Sprintf("Progress: %d%% (%.2f GB / %.2f GB)",
						progress,
						float64(currentWritten)/(1024*1024*1024),
						float64(size)/(1024*1024*1024)))
				}
			}
		}(i)
	}

	// Distribute work
	go func() {
		offset := int64(0)
		for offset < size {
			writeSize := chunkSize
			if offset+writeSize > size {
				writeSize = size - offset
			}
			tasks <- writeTask{offset: offset, size: writeSize}
			offset += writeSize
		}
		close(tasks)
	}()

	// Wait for completion
	wg.Wait()
	close(errors)

	// Check for errors
	if err := <-errors; err != nil {
		return err
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
		queueDepth   int
		throughputFn func(float64)
		iopsFn       func(float64)
		latencyFn    func(float64)
		threadsFn    func(int)
		durationFn   func(int)
		queueDepthFn func(int)
	}{
		{"Read Throughput", config.ReadTPIOSize, false, config.ReadTPThreads, config.ReadTPDuration, config.ReadTPQueueDepth,
			func(v float64) { results.ReadThroughputMBps = v },
			func(v float64) { results.ReadThroughputIOPS = v },
			func(v float64) { results.ReadTPLatencyMs = v },
			func(v int) { results.ReadTPThreads = v },
			func(v int) { results.ReadTPDuration = v },
			func(v int) { results.ReadTPQueueDepth = v }},
		{"Read IOPS", config.ReadIOPSIOSize, false, config.ReadIOPSThreads, config.ReadIOPSDuration, config.ReadIOPSQueueDepth,
			func(v float64) { results.ReadIOPSThroughputMBps = v },
			func(v float64) { results.ReadIOPS = v },
			func(v float64) { results.ReadIOPSLatencyMs = v },
			func(v int) { results.ReadIOPSThreads = v },
			func(v int) { results.ReadIOPSDuration = v },
			func(v int) { results.ReadIOPSQueueDepth = v }},
		{"Write Throughput", config.WriteTPIOSize, true, config.WriteTPThreads, config.WriteTPDuration, config.WriteTPQueueDepth,
			func(v float64) { results.WriteThroughputMBps = v },
			func(v float64) { results.WriteThroughputIOPS = v },
			func(v float64) { results.WriteTPLatencyMs = v },
			func(v int) { results.WriteTPThreads = v },
			func(v int) { results.WriteTPDuration = v },
			func(v int) { results.WriteTPQueueDepth = v }},
		{"Write IOPS", config.WriteIOPSIOSize, true, config.WriteIOPSThreads, config.WriteIOPSDuration, config.WriteIOPSQueueDepth,
			func(v float64) { results.WriteIOPSThroughputMBps = v },
			func(v float64) { results.WriteIOPS = v },
			func(v float64) { results.WriteIOPSLatencyMs = v },
			func(v int) { results.WriteIOPSThreads = v },
			func(v int) { results.WriteIOPSDuration = v },
			func(v int) { results.WriteIOPSQueueDepth = v }},
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

		// Store threads, duration, and queue depth for this test
		test.threadsFn(test.threads)
		test.durationFn(test.duration)
		test.queueDepthFn(test.queueDepth)

		throughput, iops, latency, err := e.runSingleTest(config.Device, ioSizeBytes, test.threads, test.duration, test.queueDepth, test.isWrite, progressCallback)
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
func (e *Engine) RunSingleTest(devicePath string, ioSize int64, threads int, duration int, queueDepth int, isWrite bool, progressCallback func(string)) (float64, float64, float64, error) {
	return e.runSingleTest(devicePath, ioSize, threads, duration, queueDepth, isWrite, progressCallback)
}

func (e *Engine) runSingleTest(devicePath string, ioSize int64, threads int, duration int, queueDepth int, isWrite bool, progressCallback func(string)) (float64, float64, float64, error) {
	var totalOps atomic.Int64
	var totalBytes atomic.Int64
	var totalLatencyNs atomic.Int64
	var wg sync.WaitGroup
	stopChan := make(chan bool)
	workerErrors := make(chan error, threads)

	// Start worker threads
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			err := e.worker(devicePath, ioSize, queueDepth, isWrite, &totalOps, &totalBytes, &totalLatencyNs, stopChan)
			if err != nil {
				select {
				case workerErrors <- err:
				default:
				}
			}
		}(i)
	}

	// Check for immediate worker errors (within first second)
	time.Sleep(100 * time.Millisecond)
	select {
	case err := <-workerErrors:
		close(stopChan)
		wg.Wait()
		return 0, 0, 0, fmt.Errorf("worker error: %v", err)
	default:
		// No immediate errors, continue
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

				progressCallback(fmt.Sprintf("\r  %.0fs: %.2f MB/s | %.0f IOPS | %.2f ms                    ", elapsed, mbps, iops, latencyMs))
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

	// Print newline to move to next line after progress updates
	progressCallback("\n")

	elapsed := time.Since(startTime).Seconds()
	ops := totalOps.Load()
	bytes := totalBytes.Load()
	latencyNs := totalLatencyNs.Load()

	// Check if any operations completed
	if ops == 0 {
		return 0, 0, 0, fmt.Errorf("no I/O operations completed - this usually means the device could not be accessed. On Windows, physical drive access requires Administrator privileges")
	}

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

func (e *Engine) worker(devicePath string, ioSize int64, queueDepth int, isWrite bool, totalOps *atomic.Int64, totalBytes *atomic.Int64, totalLatencyNs *atomic.Int64, stopChan chan bool) error {
	var file *os.File
	var err error

	if isWrite {
		file, err = openDeviceForWrite(devicePath)
	} else {
		file, err = openDeviceForRead(devicePath)
	}

	if err != nil {
		return err
	}
	defer file.Close()

	// Get device size for random seeks
	deviceSize, err := getDeviceSize(file)
	if err != nil {
		return fmt.Errorf("failed to get device size: %v", err)
	}

	// Pre-allocate multiple buffers for queue depth
	if queueDepth <= 0 {
		queueDepth = 4 // Default
	}
	buffers := make([][]byte, queueDepth)
	for i := 0; i < queueDepth; i++ {
		buffers[i] = make([]byte, ioSize)
		if isWrite {
			rand.Read(buffers[i])
		}
	}

	// Pre-calculate random offsets for better performance
	maxOffset := (deviceSize / ioSize) - 1
	if maxOffset <= 0 {
		maxOffset = 1
	}
	offsets := make([]int64, 1000)
	randBuf := make([]byte, 8)
	for i := range offsets {
		rand.Read(randBuf)
		randVal := int64(randBuf[0]) | int64(randBuf[1])<<8 | int64(randBuf[2])<<16 | int64(randBuf[3])<<24 |
			int64(randBuf[4])<<32 | int64(randBuf[5])<<40 | int64(randBuf[6])<<48 | int64(randBuf[7])<<56
		if randVal < 0 {
			randVal = -randVal
		}
		offsets[i] = (randVal % maxOffset) * ioSize
	}
	offsetIdx := 0

	// Batch counters to reduce atomic contention - update every N operations
	const batchSize = 256
	localOps := int64(0)
	localBytes := int64(0)
	localLatencyNs := int64(0)
	
	// Sample latency every N operations to reduce overhead (1% sampling)
	latencySampleRate := 100
	
	// Fast path - tight loop with minimal overhead
	bufferIdx := 0
	loopCount := 0
	for {
		// Check stop channel less frequently
		if loopCount%queueDepth == 0 {
			select {
			case <-stopChan:
				// Flush remaining counts
				if localOps > 0 {
					totalOps.Add(localOps)
					totalBytes.Add(localBytes)
					totalLatencyNs.Add(localLatencyNs)
				}
				return nil
			default:
			}
		}
		loopCount++

		// Issue multiple IOs in a batch
		for burst := 0; burst < queueDepth; burst++ {
			offset := offsets[offsetIdx]
			offsetIdx = (offsetIdx + 1) % len(offsets)

			// Sample latency periodically to avoid overhead on every IO
			measureLatency := (localOps % int64(latencySampleRate)) == 0
			var opStart time.Time
			if measureLatency {
				opStart = time.Now()
			}

			var n int
			if isWrite {
				n, err = file.WriteAt(buffers[bufferIdx], offset)
			} else {
				n, err = file.ReadAt(buffers[bufferIdx], offset)
			}
			
			if err == nil && n > 0 {
				localOps++
				localBytes += int64(n)
				
				if measureLatency {
					latency := time.Since(opStart).Nanoseconds()
					localLatencyNs += latency
				}
				
				// Batch update to reduce atomic contention
				if localOps >= batchSize {
					totalOps.Add(localOps)
					totalBytes.Add(localBytes)
					totalLatencyNs.Add(localLatencyNs)
					localOps = 0
					localBytes = 0
					localLatencyNs = 0
				}
			}

			bufferIdx = (bufferIdx + 1) % queueDepth
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

// Platform-specific device size detection
func getDeviceSize(file *os.File) (int64, error) {
	if runtime.GOOS == "windows" {
		// On Windows with FILE_FLAG_NO_BUFFERING, Seek doesn't work on physical drives
		// Use DeviceIoControl instead
		return getDeviceSizeWindows(file)
	}
	// On Linux/Unix, standard Seek works fine
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	// Reset position to start (though we don't actually use sequential I/O)
	_, err = file.Seek(0, io.SeekStart)
	return size, err
}
