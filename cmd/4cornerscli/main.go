package main

import (
	"4corners/benchmark"
	"4corners/device"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type config struct {
	devicePath         string
	duration           int
	readTPThreads      int
	writeTPThreads     int
	readIOPSThreads    int
	writeIOPSThreads   int
	readTPBlockSize    int
	writeTPBlockSize   int
	readIOPSBlockSize  int
	writeIOPSBlockSize int
	readTPQueueDepth   int
	writeTPQueueDepth  int
	readIOPSQueueDepth int
	writeIOPSQueueDepth int
	prepDevice         bool
	createFile         bool
	fileSize           int64
	runTests           string
}

func main() {
	cfg := parseFlags()
	
	if cfg.devicePath == "" {
		fmt.Println("Error: device path is required")
		flag.Usage()
		os.Exit(1)
	}
	
	fmt.Println("4Corners Disk Benchmark - CLI")
	fmt.Println("==============================")
	fmt.Println()
	
	// Create file device if requested
	if cfg.createFile {
		fmt.Printf("Creating file device: %s (%d GB)\n", cfg.devicePath, cfg.fileSize)
		err := device.CreateFileDevice(cfg.devicePath, cfg.fileSize*1024*1024*1024, func(msg string) {
			fmt.Println(msg)
		})
		if err != nil {
			fmt.Printf("Error creating file device: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("File device created successfully")
	}
	
	// Create benchmark engine
	engine := benchmark.NewEngine()
	
	// Prep device if requested
	if cfg.prepDevice {
		fmt.Printf("Preparing device: %s\n", cfg.devicePath)
		err := engine.PrepDevice(cfg.devicePath, func(msg string) {
			fmt.Println(msg)
		})
		if err != nil {
			fmt.Printf("Error preparing device: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Device prepared successfully")
		fmt.Println()
	}
	
	// Determine which tests to run
	runAll := cfg.runTests == "all"
	runReadTP := runAll || strings.Contains(cfg.runTests, "read-tp")
	runWriteTP := runAll || strings.Contains(cfg.runTests, "write-tp")
	runReadIOPS := runAll || strings.Contains(cfg.runTests, "read-iops")
	runWriteIOPS := runAll || strings.Contains(cfg.runTests, "write-iops")
	
	fmt.Println("Starting benchmark tests...")
	fmt.Println()
	
	// Configure benchmark
	config := benchmark.Config{
		Device:              cfg.devicePath,
		ReadTPIOSize:        fmt.Sprintf("%dk", cfg.readTPBlockSize),
		WriteTPIOSize:       fmt.Sprintf("%dk", cfg.writeTPBlockSize),
		ReadIOPSIOSize:      fmt.Sprintf("%dk", cfg.readIOPSBlockSize),
		WriteIOPSIOSize:     fmt.Sprintf("%dk", cfg.writeIOPSBlockSize),
		ReadTPThreads:       cfg.readTPThreads,
		WriteTPThreads:      cfg.writeTPThreads,
		ReadIOPSThreads:     cfg.readIOPSThreads,
		WriteIOPSThreads:    cfg.writeIOPSThreads,
		ReadTPDuration:      cfg.duration,
		WriteTPDuration:     cfg.duration,
		ReadIOPSDuration:    cfg.duration,
		WriteIOPSDuration:   cfg.duration,
		ReadTPQueueDepth:    cfg.readTPQueueDepth,
		WriteTPQueueDepth:   cfg.writeTPQueueDepth,
		ReadIOPSQueueDepth:  cfg.readIOPSQueueDepth,
		WriteIOPSQueueDepth: cfg.writeIOPSQueueDepth,
	}
	
	results := &benchmark.Results{
		TestDate: time.Now(),
		Config:   config,
	}
	
	// Progress callback
	progressCallback := func(output string) {
		fmt.Print(output)
	}
	
	// Run Read Throughput test
	if runReadTP {
		fmt.Println("Running Read Throughput Test...")
		throughput, iops, latency, err := engine.RunSingleTest(cfg.devicePath, int64(cfg.readTPBlockSize*1024), cfg.readTPThreads, cfg.duration, cfg.readTPQueueDepth, false, progressCallback)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			results.ReadThroughputMBps = throughput
			results.ReadThroughputIOPS = iops
			results.ReadTPLatencyMs = latency
			results.ReadTPThreads = cfg.readTPThreads
			results.ReadTPDuration = cfg.duration
			results.ReadTPQueueDepth = cfg.readTPQueueDepth
		}
		fmt.Println()
	}
	
	// Run Write Throughput test
	if runWriteTP {
		fmt.Println("Running Write Throughput Test...")
		throughput, iops, latency, err := engine.RunSingleTest(cfg.devicePath, int64(cfg.writeTPBlockSize*1024), cfg.writeTPThreads, cfg.duration, cfg.writeTPQueueDepth, true, progressCallback)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			results.WriteThroughputMBps = throughput
			results.WriteThroughputIOPS = iops
			results.WriteTPLatencyMs = latency
			results.WriteTPThreads = cfg.writeTPThreads
			results.WriteTPDuration = cfg.duration
			results.WriteTPQueueDepth = cfg.writeTPQueueDepth
		}
		fmt.Println()
	}
	
	// Run Read IOPS test
	if runReadIOPS {
		fmt.Println("Running Read IOPS Test...")
		throughput, iops, latency, err := engine.RunSingleTest(cfg.devicePath, int64(cfg.readIOPSBlockSize*1024), cfg.readIOPSThreads, cfg.duration, cfg.readIOPSQueueDepth, false, progressCallback)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			results.ReadIOPSThroughputMBps = throughput
			results.ReadIOPS = iops
			results.ReadIOPSLatencyMs = latency
			results.ReadIOPSThreads = cfg.readIOPSThreads
			results.ReadIOPSDuration = cfg.duration
			results.ReadIOPSQueueDepth = cfg.readIOPSQueueDepth
		}
		fmt.Println()
	}
	
	// Run Write IOPS test
	if runWriteIOPS {
		fmt.Println("Running Write IOPS Test...")
		throughput, iops, latency, err := engine.RunSingleTest(cfg.devicePath, int64(cfg.writeIOPSBlockSize*1024), cfg.writeIOPSThreads, cfg.duration, cfg.writeIOPSQueueDepth, true, progressCallback)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			results.WriteIOPSThroughputMBps = throughput
			results.WriteIOPS = iops
			results.WriteIOPSLatencyMs = latency
			results.WriteIOPSThreads = cfg.writeIOPSThreads
			results.WriteIOPSDuration = cfg.duration
			results.WriteIOPSQueueDepth = cfg.writeIOPSQueueDepth
		}
		fmt.Println()
	}
	
	fmt.Println()
	fmt.Println("Benchmark completed!")
	fmt.Println()
	
	// Generate and display report
	report := results.GenerateTextReport()
	fmt.Println(report)
	
	// Save report files
	err := results.SaveReport(".")
	if err != nil {
		fmt.Printf("Warning: failed to save reports: %v\n", err)
	} else {
		fmt.Println("Reports saved successfully")
	}
}

func parseFlags() config {
	cfg := config{}
	
	flag.StringVar(&cfg.devicePath, "device", "", "Device or file path (required)")
	flag.IntVar(&cfg.duration, "duration", 60, "Test duration in seconds")
	
	flag.IntVar(&cfg.readTPThreads, "read-tp-threads", 30, "Read throughput threads")
	flag.IntVar(&cfg.writeTPThreads, "write-tp-threads", 16, "Write throughput threads")
	flag.IntVar(&cfg.readIOPSThreads, "read-iops-threads", 120, "Read IOPS threads")
	flag.IntVar(&cfg.writeIOPSThreads, "write-iops-threads", 120, "Write IOPS threads")
	
	flag.IntVar(&cfg.readTPQueueDepth, "read-tp-qd", 4, "Read throughput queue depth per thread")
	flag.IntVar(&cfg.writeTPQueueDepth, "write-tp-qd", 4, "Write throughput queue depth per thread")
	flag.IntVar(&cfg.readIOPSQueueDepth, "read-iops-qd", 4, "Read IOPS queue depth per thread")
	flag.IntVar(&cfg.writeIOPSQueueDepth, "write-iops-qd", 4, "Write IOPS queue depth per thread")
	
	flag.IntVar(&cfg.readTPBlockSize, "read-tp-bs", 128, "Read throughput block size (KB)")
	flag.IntVar(&cfg.writeTPBlockSize, "write-tp-bs", 64, "Write throughput block size (KB)")
	flag.IntVar(&cfg.readIOPSBlockSize, "read-iops-bs", 4, "Read IOPS block size (KB)")
	flag.IntVar(&cfg.writeIOPSBlockSize, "write-iops-bs", 4, "Write IOPS block size (KB)")
	
	flag.BoolVar(&cfg.prepDevice, "prep", false, "Prep device before testing")
	flag.BoolVar(&cfg.createFile, "create-file", false, "Create a file device")
	flag.Int64Var(&cfg.fileSize, "file-size", 10, "File device size in GB (if creating)")
	
	flag.StringVar(&cfg.runTests, "tests", "all", "Tests to run: all, read-tp, write-tp, read-iops, write-iops (comma-separated)")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "4Corners Disk Benchmark - CLI\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Test existing device\n")
		fmt.Fprintf(os.Stderr, "  %s -device /dev/sdb\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Create and test file device\n")
		fmt.Fprintf(os.Stderr, "  %s -device testfile.bin -create-file -file-size 10\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Run only specific tests\n")
		fmt.Fprintf(os.Stderr, "  %s -device /dev/sdb -tests read-tp,write-tp\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Custom thread counts and duration\n")
		fmt.Fprintf(os.Stderr, "  %s -device /dev/sdb -duration 120 -read-tp-threads 64 -write-tp-threads 32\n\n", os.Args[0])
	}
	
	flag.Parse()
	
	return cfg
}
