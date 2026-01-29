# 4cornerscli - Command Line Interface

Cross-platform disk I/O benchmark tool that measures the "4 corners" of storage performance:
- Read Throughput
- Write Throughput  
- Read IOPS
- Write IOPS

## Installation

Pre-built binaries are provided for:
- **Windows**: `4cornerscli.exe`
- **Linux**: `4cornerscli`

Simply download and run - no installation required.

## Quick Start

### Test an existing device:
```bash
# Linux
./4cornerscli -device /dev/sdb

# Windows
4cornerscli.exe -device \\.\PhysicalDrive1
```

### Create and test a file:
```bash
./4cornerscli -device testfile.bin -create-file -file-size 10
```

### Run specific tests only:
```bash
./4cornerscli -device /dev/sdb -tests read-tp,write-iops
```

## Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-device` | (required) | Device or file path to test |
| `-duration` | 60 | Test duration in seconds |
| `-read-tp-threads` | 30 | Read throughput thread count |
| `-write-tp-threads` | 16 | Write throughput thread count |
| `-read-iops-threads` | 120 | Read IOPS thread count |
| `-write-iops-threads` | 120 | Write IOPS thread count |
| `-read-tp-bs` | 128 | Read throughput block size (KB) |
| `-write-tp-bs` | 64 | Write throughput block size (KB) |
| `-read-iops-bs` | 4 | Read IOPS block size (KB) |
| `-write-iops-bs` | 4 | Write IOPS block size (KB) |
| `-prep` | false | Prep device before testing |
| `-create-file` | false | Create a file device |
| `-file-size` | 10 | File device size in GB |
| `-tests` | all | Tests to run (comma-separated) |

## Test Selection

Use the `-tests` flag to run specific tests:
- `all` - Run all 4 tests (default)
- `read-tp` - Read throughput only
- `write-tp` - Write throughput only
- `read-iops` - Read IOPS only
- `write-iops` - Write IOPS only

Combine multiple tests with commas:
```bash
./4cornerscli -device /dev/sdb -tests read-tp,read-iops
```

## Output

Results are displayed in real-time on stdout and saved to:
- `4corners-report-YYYYMMDD-HHMMSS.txt` - Human-readable report
- `4corners-report-YYYYMMDD-HHMMSS.json` - Machine-readable JSON

## Examples

### Quick 30-second test:
```bash
./4cornerscli -device /dev/nvme0n1 -duration 30
```

### High-thread IOPS test:
```bash
./4cornerscli -device /dev/sdb -read-iops-threads 256 -write-iops-threads 256
```

### Custom block sizes:
```bash
./4cornerscli -device /dev/sdc -read-tp-bs 256 -write-tp-bs 128
```

### Read-only testing:
```bash
./4cornerscli -device /dev/sdd -tests read-tp,read-iops
```

## Device Paths

### Linux
- Physical devices: `/dev/sda`, `/dev/nvme0n1`
- Logical volumes: `/dev/mapper/vg-lv`
- Files: `./testfile.bin`

### Windows  
- Physical drives: `\\.\PhysicalDrive0`, `\\.\PhysicalDrive1`
- Volumes: `\\.\C:`, `\\.\D:`
- Files: `C:\test\testfile.bin`

## Notes

- **Direct I/O**: Tests use direct I/O (O_DIRECT/FILE_FLAG_NO_BUFFERING) to bypass caching
- **Destructive Testing**: Write tests will overwrite data - use with caution!
- **Permissions**: Requires administrator/root access for device testing
- **Reports**: Automatically saved in current directory

## See Also

- GUI version: `4corners.exe` / `4corners`
