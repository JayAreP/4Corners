# 4Corners Disk Benchmark

A cross-platform disk I/O benchmark tool with a GUI, designed to test the "4 corners" of storage performance.

## Features

- **Read Throughput**: Large block random reads (default 128k)
- **Write Throughput**: Large block random writes (default 64k)
- **Read IOPS**: Small block random reads (default 4k)
- **Write IOPS**: Small block random writes (default 4k)

## Features

- Cross-platform support (Windows & Linux)
- Single executable
- Direct I/O for accurate measurements
- Device preparation (fill with random data)
- Real-time progress reporting
- JSON and text report generation
- Customizable thread count and test duration

## Building

### Prerequisites

- Go 1.21 or later
- GCC (for CGo) or MinGW on Windows

### Build Commands

```bash
# Download dependencies
go mod download

# Build for current platform
go build -o 4corners

# Build for Windows (from any platform)
GOOS=windows GOARCH=amd64 go build -o 4corners.exe

# Build for Linux (from any platform)
GOOS=linux GOARCH=amd64 go build -o 4corners
```

## Usage

### GUI Application

**⚠️ OpenGL 2.0+ Required** - The GUI requires OpenGL support. If you encounter errors, see [GUI-REQUIREMENTS.md](GUI-REQUIREMENTS.md) for troubleshooting, or use the CLI version.

1. **Run the application**: Execute the compiled binary
2. **Select Device**: Click "Select Device" to choose a block device
3. **Prep Device** (Optional): Fill the device with random data for consistent testing
4. **Configure Tests**: Adjust IO sizes, thread count, and duration as needed
5. **Save Report** (Optional): Choose a folder to save results
6. **Run**: Click "Run" to start the benchmark suite

### CLI Alternative (Recommended for Remote/Automated Use)

For systems without GUI support or when running in automated scripts, use the CLI version:
```bash
# Windows (run as Administrator!)
4cornerscli.exe -device \\.\PhysicalDrive1

# Linux (run as root or with sudo)
sudo ./4cornerscli -device /dev/sdb
```

See [CLI-README.md](CLI-README.md) for full CLI documentation.

### Important Notes

- **Administrator/Root privileges required** to access raw block devices on Windows/Linux
- **Windows**: Must run Command Prompt/PowerShell as Administrator for physical drive access
- **GUI OpenGL Requirement**: If GUI fails with OpenGL errors, use the CLI version
- **WARNING**: Write tests can destroy data on the selected device
- Test duration is per-test (4 tests total)
- Recommended to prep device before first run

## Device Paths

### Windows
- Format: `\\.\PhysicalDrive0`, `\\.\PhysicalDrive1`, etc.
- Use Disk Management to identify drive numbers

### Linux
- Format: `/dev/sda`, `/dev/nvme0n1`, etc.
- Use `lsblk` to list available devices

## Configuration Options

- **Read Throughput IO Size**: Default 128k
- **Write Throughput IO Size**: Default 64k
- **Read IOPS IO Size**: Default 4k
- **Write IOPS IO Size**: Default 4k
- **Threads**: Number of concurrent I/O workers (default 64)
- **Duration**: Test duration in seconds per test (default 60)

## Output

Results are displayed in the GUI and optionally saved to:
- JSON file: `4corners-report-YYYYMMDD-HHMMSS.json`
- Text file: `4corners-report-YYYYMMDD-HHMMSS.txt`

## License

MIT License
