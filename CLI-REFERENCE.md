# 4C Disk Benchmark - CLI Reference

## Overview

`4c` is a cross-platform disk I/O benchmark tool that measures the "4 corners" of storage performance:

- **Read Throughput** — Maximum read bandwidth (large block sequential)
- **Write Throughput** — Maximum write bandwidth (large block sequential)
- **Read IOPS** — Random read operations per second (small block random)
- **Write IOPS** — Random write operations per second (small block random)

Each test reports throughput (MB/s), IOPS, and latency (avg, p50, p99).

## Usage

```
4c --device <DEVICE> [OPTIONS]
```

## Required

| Option | Description |
|--------|-------------|
| `-d`, `--device <PATH>` | Device or file path to benchmark |

### Windows device paths
```
\\.\PhysicalDrive1       Physical drive
\\.\D:                   Volume
C:\test\benchmark.dat    File
```

### Linux device paths
```
/dev/sdb                 SATA/SAS drive
/dev/nvme0n1             NVMe drive
/mnt/storage/bench.dat   File
```

## Test Selection

| Option | Default | Description |
|--------|---------|-------------|
| `--tests <LIST>` | `all` | Comma-separated list of tests to run |

Values: `all`, `read-tp`, `write-tp`, `read-iops`, `write-iops`

```powershell
# Run all 4 tests
4c --device \\.\D:

# IOPS tests only
4c --device \\.\D: --tests read-iops,write-iops

# Read tests only
4c --device \\.\D: --tests read-tp,read-iops
```

## Duration

| Option | Default | Description |
|--------|---------|-------------|
| `--duration <SECS>` | `60` | Duration of each test in seconds |

## Thread Configuration

Each test type uses its own thread count. More threads generate more concurrent I/O.

| Option | Default | Description |
|--------|---------|-------------|
| `--read-tp-threads` | `30` | Threads for read throughput test |
| `--write-tp-threads` | `16` | Threads for write throughput test |
| `--read-iops-threads` | `120` | Threads for read IOPS test |
| `--write-iops-threads` | `120` | Threads for write IOPS test |

### Guidelines by device type

| Device | Throughput Threads | IOPS Threads |
|--------|--------------------|--------------|
| HDD | 8–16 | 32–64 |
| SATA SSD | 16–32 | 64–128 |
| NVMe SSD | 32–64 | 128–256 |
| High-end NVMe | 64–128 | 256–512 |

## Queue Depth

Queue depth controls how many I/Os each thread keeps in flight simultaneously. Higher queue depth drives more parallelism per thread. This is a key parameter for IOPS performance.

| Option | Default | Description |
|--------|---------|-------------|
| `--read-tp-qd` | `1` | Queue depth per thread for read throughput |
| `--write-tp-qd` | `1` | Queue depth per thread for write throughput |
| `--read-iops-qd` | `32` | Queue depth per thread for read IOPS |
| `--write-iops-qd` | `32` | Queue depth per thread for write IOPS |

Total concurrent I/Os = threads x queue depth. For example, the default IOPS config runs 120 threads x 32 QD = 3,840 concurrent I/Os.

## Block Size

Block size is the amount of data transferred per I/O operation, specified in KB.

| Option | Default | Description |
|--------|---------|-------------|
| `--read-tp-bs` | `128` | Block size (KB) for read throughput |
| `--write-tp-bs` | `64` | Block size (KB) for write throughput |
| `--read-iops-bs` | `4` | Block size (KB) for read IOPS |
| `--write-iops-bs` | `4` | Block size (KB) for write IOPS |

## File & Device Preparation

| Option | Default | Description |
|--------|---------|-------------|
| `--create-file` | off | Create a file device before testing |
| `--file-size <GB>` | `10` | Size of the file to create (in GB) |
| `--prep` | off | Write random data to device before testing |

Use `--create-file` to benchmark against a file instead of a raw device. Use `--prep` to pre-condition a device with random data for accurate first-write performance.

## Examples

### Quick 30-second test on a volume
```powershell
4c --device \\.\D: --duration 30
```

### Create a 20 GB test file and benchmark it
```powershell
4c --device C:\test\bench.dat --create-file --file-size 20
```

### IOPS-focused test with higher queue depth
```powershell
4c --device \\.\PhysicalDrive1 --tests read-iops,write-iops --read-iops-qd 64 --write-iops-qd 64
```

### High-thread NVMe test
```powershell
4c --device \\.\PhysicalDrive1 --read-iops-threads 256 --write-iops-threads 256
```

### Read-only test (safe for production volumes)
```powershell
4c --device \\.\D: --tests read-tp,read-iops
```

### Prep device then run full benchmark
```powershell
4c --device \\.\PhysicalDrive1 --prep
```

### Linux: NVMe full test
```bash
sudo ./4c --device /dev/nvme0n1 --duration 120
```

## Output

### Console
Real-time progress is printed every 5 seconds during each test:
```
Running Read IOPS Test...
  Read test: 4KB blocks, 120 threads, QD=32, 60 seconds
  Device size: 476.94 GB
    5s:  1234.56 MB/s |     316045 IOPS |    121.3 us avg lat
   10s:  1245.67 MB/s |     318891 IOPS |    119.8 us avg lat
  ...
  RESULT: 1240.12 MB/s | 317471 IOPS | avg 120.5 us | p50 98.2 us | p99 412.7 us
```

### Report Files
Two files are saved to the current directory after each run:

- `4c-report-YYYYMMDD-HHMMSS.txt` — Human-readable text report
- `4c-report-YYYYMMDD-HHMMSS.json` — Machine-readable JSON report

## Permissions

- **Windows**: Administrator required for raw devices (`\\.\PhysicalDrive#`, `\\.\D:`). Files work as regular user.
- **Linux**: Root/sudo required for block devices (`/dev/sd*`, `/dev/nvme*`). Files work as regular user.

## Safety

Write tests are destructive. They overwrite data on the target device. Do not run write tests against drives containing data you need. Use `--tests read-tp,read-iops` for read-only testing on production systems.

## Building

```
cargo build --release
```

Binary output: `target/release/4c.exe` (Windows) or `target/release/4c` (Linux).
