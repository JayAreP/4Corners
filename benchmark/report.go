package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (r *Results) SaveReport(outputDir string) error {
	timestamp := time.Now().Format("20060102-150405")
	
	// Save JSON report
	jsonPath := filepath.Join(outputDir, fmt.Sprintf("4corners-report-%s.json", timestamp))
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	
	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON report: %v", err)
	}
	
	// Save text report
	textPath := filepath.Join(outputDir, fmt.Sprintf("4corners-report-%s.txt", timestamp))
	textReport := r.GenerateTextReport()
	err = os.WriteFile(textPath, []byte(textReport), 0644)
	if err != nil {
		return fmt.Errorf("failed to write text report: %v", err)
	}
	
	return nil
}

func (r *Results) GenerateTextReport() string {
	report := "========================================\n"
	report += "4Corners Disk Benchmark Report\n"
	report += "========================================\n\n"
	report += fmt.Sprintf("Test Date: %s\n", r.TestDate.Format("2006-01-02 15:04:05"))
	report += fmt.Sprintf("Device: %s\n\n", r.Config.Device)
	
	report += "Configuration:\n"
	report += fmt.Sprintf("  Read Throughput IO Size: %s\n", r.Config.ReadTPIOSize)
	report += fmt.Sprintf("  Write Throughput IO Size: %s\n", r.Config.WriteTPIOSize)
	report += fmt.Sprintf("  Read IOPS IO Size: %s\n", r.Config.ReadIOPSIOSize)
	report += fmt.Sprintf("  Write IOPS IO Size: %s\n\n", r.Config.WriteIOPSIOSize)
	
	report += "Results:\n"
	report += "========================================\n"
	report += "Read Throughput Test:\n"
	if r.ReadTPThreads > 0 {
		report += fmt.Sprintf("  Threads:    %10d\n", r.ReadTPThreads)
		report += fmt.Sprintf("  Duration:   %10d seconds\n", r.ReadTPDuration)
	}
	report += fmt.Sprintf("  Throughput: %10.2f MB/s\n", r.ReadThroughputMBps)
	report += fmt.Sprintf("  IOPS:       %10.0f IOPS\n", r.ReadThroughputIOPS)
	report += fmt.Sprintf("  Latency:    %10.2f ms\n\n", r.ReadTPLatencyMs)
	
	report += "Write Throughput Test:\n"
	if r.WriteTPThreads > 0 {
		report += fmt.Sprintf("  Threads:    %10d\n", r.WriteTPThreads)
		report += fmt.Sprintf("  Duration:   %10d seconds\n", r.WriteTPDuration)
	}
	report += fmt.Sprintf("  Throughput: %10.2f MB/s\n", r.WriteThroughputMBps)
	report += fmt.Sprintf("  IOPS:       %10.0f IOPS\n", r.WriteThroughputIOPS)
	report += fmt.Sprintf("  Latency:    %10.2f ms\n\n", r.WriteTPLatencyMs)
	
	report += "Read IOPS Test:\n"
	if r.ReadIOPSThreads > 0 {
		report += fmt.Sprintf("  Threads:    %10d\n", r.ReadIOPSThreads)
		report += fmt.Sprintf("  Duration:   %10d seconds\n", r.ReadIOPSDuration)
	}
	report += fmt.Sprintf("  Throughput: %10.2f MB/s\n", r.ReadIOPSThroughputMBps)
	report += fmt.Sprintf("  IOPS:       %10.0f IOPS\n", r.ReadIOPS)
	report += fmt.Sprintf("  Latency:    %10.2f ms\n\n", r.ReadIOPSLatencyMs)
	
	report += "Write IOPS Test:\n"
	if r.WriteIOPSThreads > 0 {
		report += fmt.Sprintf("  Threads:    %10d\n", r.WriteIOPSThreads)
		report += fmt.Sprintf("  Duration:   %10d seconds\n", r.WriteIOPSDuration)
	}
	report += fmt.Sprintf("  Throughput: %10.2f MB/s\n", r.WriteIOPSThroughputMBps)
	report += fmt.Sprintf("  IOPS:       %10.0f IOPS\n", r.WriteIOPS)
	report += fmt.Sprintf("  Latency:    %10.2f ms\n", r.WriteIOPSLatencyMs)
	report += "========================================\n"
	
	return report
}
