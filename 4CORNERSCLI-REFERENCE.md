# 4cornerscli - Comprehensive Reference Guide

## Overview

`4cornerscli` is a cross-platform command-line disk I/O benchmark tool that measures the "4 corners" of storage performance:

- **Read Throughput** - Maximum sequential read bandwidth
- **Write Throughput** - Maximum sequential write bandwidth  
- **Read IOPS** - Random read operations per second
- **Write IOPS** - Random write operations per second

Each test measures three key metrics:
- **Throughput** (MB/s) - Data transfer rate
- **IOPS** - Operations per second
- **Latency** (ms) - Average operation time

## Installation

### Windows
Download `4cornerscli.exe` and run directly from Command Prompt or PowerShell. No installation required.

### Linux
Download `4cornerscli`, make it executable, and run:
```bash
chmod +x 4cornerscli
./4cornerscli [options]
```

## Basic Usage

### Quick Start

**Windows:**
```powershell
# Test a physical drive
4cornerscli.exe -device \\.\PhysicalDrive1

# Test a volume
4cornerscli.exe -device \\.\D:

# Create and test a file
4cornerscli.exe -device C:\test\benchmark.dat -create-file -file-size 10
```

**Linux:**
```bash
# Test a block device
./4cornerscli -device /dev/sdb

# Test a file
./4cornerscli -device /mnt/storage/benchmark.dat -create-file -file-size 10
```

## Command-Line Options

### Required Options

| Option | Type | Description |
|--------|------|-------------|
| `-device` | string | Device or file path to test (see Device Paths section) |

### Test Configuration

| Option | Default | Description |
|--------|---------|-------------|
| `-duration` | 60 | Test duration in seconds for each test |
| `-tests` | all | Tests to run (see Test Selection section) |

### Thread Configuration

Each test type can use a different thread count for optimal performance:

| Option | Default | Description |
|--------|---------|-------------|
| `-read-tp-threads` | 30 | Read throughput thread count |
| `-write-tp-threads` | 16 | Write throughput thread count |
| `-read-iops-threads` | 120 | Read IOPS thread count |
| `-write-iops-threads` | 120 | Write IOPS thread count |

### Block Size Configuration

Block sizes determine how much data is transferred in each I/O operation:

| Option | Default | Description |
|--------|---------|-------------|
| `-read-tp-bs` | 128 | Read throughput block size in KB |
| `-write-tp-bs` | 64 | Write throughput block size in KB |
| `-read-iops-bs` | 4 | Read IOPS block size in KB |
| `-write-iops-bs` | 4 | Write IOPS block size in KB |

### File Device Options

For testing on files rather than physical devices:

| Option | Default | Description |
|--------|---------|-------------|
| `-create-file` | false | Create a new file device before testing |
| `-file-size` | 10 | File device size in GB (only used with -create-file) |

### Advanced Options

| Option | Default | Description |
|--------|---------|-------------|
| `-prep` | false | Prep device by writing random data before testing |

## Test Selection

The `-tests` flag controls which benchmark tests run. You can specify:

### All Tests (Default)
```bash
./4cornerscli -device /dev/sdb -tests all
```
Runs all four benchmark tests in sequence.

### Individual Tests
- `read-tp` - Read Throughput only
- `write-tp` - Write Throughput only
- `read-iops` - Read IOPS only
- `write-iops` - Write IOPS only

### Multiple Tests
Combine tests with commas (no spaces):
```bash
# Run only throughput tests
./4cornerscli -device /dev/sdb -tests read-tp,write-tp

# Run only IOPS tests
./4cornerscli -device /dev/sdb -tests read-iops,write-iops

# Run read tests only
./4cornerscli -device /dev/sdb -tests read-tp,read-iops
```

## Device Paths

### Windows Device Paths

**Physical Drives:**
```powershell
# First physical drive (usually C:)
\\.\PhysicalDrive0

# Second physical drive
\\.\PhysicalDrive1

# List all physical drives
wmic diskdrive list brief
```

**Volumes:**
```powershell
# C: drive
\\.\C:

# D: drive
\\.\D:

# List volumes
wmic volume get DeviceID, Caption, DriveLetter
```

**Files:**
```powershell
# File on C: drive
C:\benchmark\testfile.dat

# File on network share
\\server\share\testfile.dat
```

### Linux Device Paths

**Block Devices:**
```bash
# SATA/SAS drive
/dev/sda
/dev/sdb

# NVMe drive
/dev/nvme0n1
/dev/nvme1n1

# Logical volume
/dev/mapper/vg-lv

# List block devices
lsblk
```

**Files:**
```bash
# File in current directory
./benchmark.dat

# File with absolute path
/mnt/storage/benchmark.dat

# File on mounted filesystem
/media/user/drive/benchmark.dat
```

## Usage Examples

### Example 1: Quick 30-Second Test

**Windows:**
```powershell
4cornerscli.exe -device \\.\PhysicalDrive1 -duration 30
```

**Linux:**
```bash
./4cornerscli -device /dev/sdb -duration 30
```

### Example 2: Custom Thread Counts

Test with higher thread counts for high-performance NVMe:

**Windows:**
```powershell
4cornerscli.exe -device \\.\PhysicalDrive1 -read-iops-threads 256 -write-iops-threads 256
```

**Linux:**
```bash
./4cornerscli -device /dev/nvme0n1 -read-iops-threads 256 -write-iops-threads 256
```

### Example 3: Custom Block Sizes

Test with larger block sizes:

**Windows:**
```powershell
4cornerscli.exe -device \\.\D: -read-tp-bs 256 -write-tp-bs 128
```

**Linux:**
```bash
./4cornerscli -device /dev/sdb -read-tp-bs 256 -write-tp-bs 128
```

### Example 4: Create and Test File Device

**Windows:**
```powershell
# Create 20 GB file and test
4cornerscli.exe -device C:\test\benchmark.dat -create-file -file-size 20
```

**Linux:**
```bash
# Create 50 GB file and test
./4cornerscli -device /mnt/storage/benchmark.dat -create-file -file-size 50
```

### Example 5: Read-Only Testing

Test without writing (requires existing data):

**Windows:**
```powershell
4cornerscli.exe -device \\.\D: -tests read-tp,read-iops
```

**Linux:**
```bash
./4cornerscli -device /dev/sdb -tests read-tp,read-iops
```

### Example 6: Write-Only Testing

Test write performance:

**Windows:**
```powershell
4cornerscli.exe -device \\.\D: -tests write-tp,write-iops
```

**Linux:**
```bash
./4cornerscli -device /dev/sdb -tests write-tp,write-iops
```

### Example 7: IOPS-Only Testing

Focus on random I/O performance:

**Windows:**
```powershell
4cornerscli.exe -device \\.\PhysicalDrive1 -tests read-iops,write-iops -duration 120
```

**Linux:**
```bash
./4cornerscli -device /dev/nvme0n1 -tests read-iops,write-iops -duration 120
```

### Example 8: Prep Device Before Testing

Write random data to device first (ensures accurate first-write performance):

**Windows:**
```powershell
4cornerscli.exe -device \\.\D: -prep
```

**Linux:**
```bash
./4cornerscli -device /dev/sdb -prep
```

### Example 9: Full Custom Configuration

**Windows:**
```powershell
4cornerscli.exe -device C:\test\benchmark.dat `
  -create-file -file-size 100 `
  -duration 180 `
  -read-tp-threads 64 -write-tp-threads 32 `
  -read-iops-threads 256 -write-iops-threads 256 `
  -read-tp-bs 256 -write-tp-bs 128 `
  -read-iops-bs 8 -write-iops-bs 8
```

**Linux:**
```bash
./4cornerscli -device /mnt/storage/benchmark.dat \
  -create-file -file-size 100 \
  -duration 180 \
  -read-tp-threads 64 -write-tp-threads 32 \
  -read-iops-threads 256 -write-iops-threads 256 \
  -read-tp-bs 256 -write-tp-bs 128 \
  -read-iops-bs 8 -write-iops-bs 8
```

## Output and Reports

### Console Output

Results are displayed in real-time during testing:
```
4Corners Disk Benchmark - CLI
==============================

Starting benchmark tests...

Running Read Throughput Test...
  5s: 1234.56 MB/s | 9876 IOPS | 1.23 ms
  10s: 1245.67 MB/s | 9965 IOPS | 1.21 ms
  ...

Benchmark completed!
```

### Report Files

Two report files are automatically saved in the current directory:

**Text Report:** `4corners-report-YYYYMMDD-HHMMSS.txt`
- Human-readable format
- Contains all test results
- Includes configuration details
- Shows threads and duration per test

**JSON Report:** `4corners-report-YYYYMMDD-HHMMSS.json`
- Machine-readable format
- Same data as text report
- Suitable for parsing/automation

### Report Format

```
========================================
4Corners Disk Benchmark Report
========================================

Test Date: 2026-01-28 14:30:45
Device: /dev/sdb

Configuration:
  Read Throughput IO Size: 128k
  Write Throughput IO Size: 64k
  Read IOPS IO Size: 4k
  Write IOPS IO Size: 4k

Results:
========================================
Read Throughput Test:
  Threads:         30
  Duration:        60 seconds
  Throughput:    1234.56 MB/s
  IOPS:            9876 IOPS
  Latency:          1.23 ms

Write Throughput Test:
  Threads:         16
  Duration:        60 seconds
  Throughput:     876.54 MB/s
  IOPS:            14024 IOPS
  Latency:          1.14 ms

Read IOPS Test:
  Threads:        120
  Duration:        60 seconds
  Throughput:     234.56 MB/s
  IOPS:           60123 IOPS
  Latency:          2.00 ms

Write IOPS Test:
  Threads:        120
  Duration:        60 seconds
  Throughput:     123.45 MB/s
  IOPS:           31603 IOPS
  Latency:          3.80 ms
========================================
```

## Performance Tuning

### Thread Count Guidelines

**HDDs (Spinning Disks):**
- Read Throughput: 8-16 threads
- Write Throughput: 8-16 threads
- Read IOPS: 32-64 threads
- Write IOPS: 32-64 threads

**SATA SSDs:**
- Read Throughput: 16-32 threads
- Write Throughput: 16-32 threads
- Read IOPS: 64-128 threads
- Write IOPS: 64-128 threads

**NVMe SSDs:**
- Read Throughput: 32-64 threads
- Write Throughput: 16-32 threads
- Read IOPS: 128-256 threads
- Write IOPS: 128-256 threads

**High-End NVMe:**
- Read Throughput: 64-128 threads
- Write Throughput: 32-64 threads
- Read IOPS: 256-512 threads
- Write IOPS: 256-512 threads

### Block Size Guidelines

**Throughput Tests:**
- Typical: 64-128 KB
- Maximum bandwidth: 256-1024 KB
- Smaller blocks: Lower throughput, higher CPU

**IOPS Tests:**
- Standard: 4 KB (industry standard)
- Database workloads: 8-16 KB
- Application specific: Match actual workload

### Duration Guidelines

- **Quick test**: 30 seconds (good for comparisons)
- **Standard test**: 60 seconds (balanced)
- **Thorough test**: 120-300 seconds (more stable results)
- **Stress test**: 600+ seconds (long-term stability)

## Permissions and Access

### Windows

**Administrator Required:**
- Testing physical drives (`\\.\PhysicalDrive#`)
- Testing volumes (`\\.\C:`, `\\.\D:`)

**Run as Administrator:**
```powershell
# Right-click PowerShell/CMD → "Run as Administrator"
# Or use elevation:
Start-Process powershell -Verb RunAs -ArgumentList "-NoExit", "-Command", "cd C:\path\to\4cornerscli; .\4cornerscli.exe -device \\.\PhysicalDrive1"
```

**Regular User:**
- Testing files (no special permissions needed)

### Linux

**Root Required:**
- Testing block devices (`/dev/sd*`, `/dev/nvme*`)
- Direct device access

**Run as Root:**
```bash
# Using sudo
sudo ./4cornerscli -device /dev/sdb

# As root user
su -
./4cornerscli -device /dev/sdb
```

**Regular User:**
- Testing files in owned directories
- May need write permissions for file creation

## Best Practices

### Before Testing

1. **Backup Data**: Write tests destroy data on the target device
2. **Close Applications**: Stop applications using the device
3. **Unmount Filesystems** (Linux): `umount /dev/sdb1`
4. **Check Device**: Verify correct device path before testing
5. **Free Space**: Ensure adequate space for file devices

### During Testing

1. **Stable System**: Avoid heavy workloads during testing
2. **Consistent Environment**: Same conditions for comparative tests
3. **Monitor Progress**: Watch for errors or anomalies
4. **Patience**: Let tests complete for accurate results

### After Testing

1. **Review Reports**: Check both console output and saved reports
2. **Compare Results**: Use consistent parameters for comparisons
3. **Archive Reports**: Save reports for future reference
4. **Remount Filesystems** (Linux): `mount /dev/sdb1 /mnt/disk`

## Troubleshooting

### Common Issues

**"Permission denied"**
- **Solution**: Run as Administrator (Windows) or root (Linux)

**"Device is busy"**
- **Solution**: Close applications, unmount filesystem, stop services

**"Invalid device path"**
- **Windows**: Check with `wmic diskdrive list brief`
- **Linux**: Check with `lsblk` or `fdisk -l`

**Low performance results**
- Increase thread counts
- Check for background processes
- Verify device isn't throttling (temperature, power)
- Ensure device supports direct I/O

**"File too large"**
- Check available disk space
- Use smaller `-file-size` value
- Verify filesystem limits (FAT32: 4GB max)

**Inconsistent results**
- Run longer duration tests (120+ seconds)
- Ensure system is idle
- Check for thermal throttling
- Verify device health

## Technical Details

### I/O Method

- **Direct I/O**: Bypasses OS cache for accurate results
  - Linux: `O_DIRECT` flag
  - Windows: `FILE_FLAG_NO_BUFFERING`
- **Unbuffered**: No write caching
- **Synchronous**: Waits for I/O completion

### Test Methodology

**Throughput Tests:**
1. Multiple threads perform sequential I/O
2. Large block sizes (64-128 KB default)
3. Measures maximum sustained bandwidth
4. Calculates average latency per operation

**IOPS Tests:**
1. Multiple threads perform random I/O
2. Small block sizes (4 KB default)
3. Measures maximum operations per second
4. Calculates average latency per operation

### Metrics Calculation

- **Throughput (MB/s)**: Total bytes transferred / elapsed time
- **IOPS**: Total operations / elapsed time
- **Latency (ms)**: Total operation time / operations count

## Safety Warnings

⚠️ **DESTRUCTIVE TESTING**: Write tests **DESTROY ALL DATA** on the target device.

**DO:**
- Test on empty devices
- Test on file devices
- Backup critical data first
- Verify device path before running
- Use read-only tests on production systems

**DO NOT:**
- Test production devices with write tests
- Test system drives (C:, /, /boot)
- Run without verifying device path
- Interrupt write tests (may leave device in bad state)

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (see console output for details) |

## Version Information

To view help and options:
```bash
# Windows
4cornerscli.exe -help

# Linux
./4cornerscli -help
```

## Related Tools

**GUI Version**: `4corners` / `4corners.exe`
- Graphical interface
- Real-time performance graphs
- Visual device selection
- Interactive configuration

## Support and Feedback

For issues, questions, or feedback:
- Check this documentation
- Review console error messages
- Verify command-line syntax
- Check device paths and permissions

## License

This tool is provided as-is for benchmarking and testing purposes.
