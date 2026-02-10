# 4C Disk Benchmark - Build Instructions

## Prerequisites

### Windows

1. **Rust Toolchain** — Install from https://rustup.rs/
   - Includes `rustc`, `cargo`, and standard library
   - Default installation includes MSVC toolchain (recommended)

2. **Visual Studio Build Tools** (required by MSVC)
   - Option A: Install "Desktop development with C++" workload from [Visual Studio Community](https://visualstudio.microsoft.com/downloads/)
   - Option B: Install [Build Tools for Visual Studio](https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2022) (lighter weight)

3. **Git** (optional, for cloning)
   - Download from https://git-scm.com/

### Linux

1. **Rust Toolchain** — Install from https://rustup.rs/
   ```bash
   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
   source $HOME/.cargo/env
   ```

2. **Build Essentials** — C compiler and development tools
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install build-essential pkg-config

   # Fedora/RHEL
   sudo dnf install gcc pkg-config

   # Arch
   sudo pacman -S base-devel
   ```

3. **Linux Headers** (for io_uring)
   ```bash
   # Ubuntu/Debian
   sudo apt-get install linux-headers-$(uname -r)

   # Fedora/RHEL
   sudo dnf install kernel-devel

   # Arch
   sudo pacman -S linux-headers
   ```

## Building

### Windows

1. Open PowerShell or Command Prompt

2. Navigate to the 4c directory:
   ```powershell
   cd C:\Users\Jar\Dropbox\VSCode2\4c
   ```

3. Build debug version (faster build, slower binary):
   ```powershell
   cargo build
   ```
   Binary: `target\debug\4c.exe`

4. Build release version (slower build, optimized binary):
   ```powershell
   cargo build --release
   ```
   Binary: `target\release\4c.exe`

### Linux

1. Open a terminal

2. Navigate to the 4c directory:
   ```bash
   cd ~/Dropbox/VSCode2/4c
   ```

3. Build debug version:
   ```bash
   cargo build
   ```
   Binary: `target/debug/4c`

4. Build release version (recommended):
   ```bash
   cargo build --release
   ```
   Binary: `target/release/4c`

## Running

### Windows (debug)
```powershell
.\target\debug\4c.exe --device \\.\D: --duration 30
```

### Windows (release)
```powershell
.\target\release\4c.exe --device \\.\D: --duration 30
```

### Linux (debug)
```bash
./target/debug/4c --device /dev/sdb --duration 30
```

### Linux (release)
```bash
./target/release/4c --device /dev/sdb --duration 30
```

## Installing (Optional)

### Windows

Copy the release binary to a directory in your PATH:

```powershell
Copy-Item .\target\release\4c.exe C:\Users\$env:USERNAME\AppData\Local\Programs\4c\
```

Then add `C:\Users\$env:USERNAME\AppData\Local\Programs\4c\` to your PATH environment variable.

Or run directly from the project directory:
```powershell
.\target\release\4c.exe --device ...
```

### Linux

Install to `/usr/local/bin`:

```bash
sudo install -v target/release/4c /usr/local/bin/4c
```

Then run from anywhere:
```bash
4c --device /dev/sdb
```

Or copy to a personal bin directory:
```bash
mkdir -p ~/.local/bin
cp target/release/4c ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"  # Add to ~/.bashrc or ~/.zshrc to persist
```

## Troubleshooting

### Windows: "Microsoft Visual C++ build tools not found"

Install Visual Studio Build Tools:
1. Go to https://visualstudio.microsoft.com/downloads/
2. Download "Build Tools for Visual Studio 2022"
3. Run installer, select "Desktop development with C++"
4. Retry `cargo build`

### Windows: "The system cannot find the file specified"

Make sure you're in the correct directory:
```powershell
cd C:\Users\Jar\Dropbox\VSCode2\4c
Get-Location  # Should show the 4c directory
cargo build
```

### Linux: "io_uring not found"

Install kernel development headers:
```bash
# Ubuntu/Debian
sudo apt-get install linux-headers-$(uname -r)

# Fedora
sudo dnf install kernel-devel

# Then retry
cargo build --release
```

### Linux: "Permission denied" when running

Make binary executable:
```bash
chmod +x target/release/4c
./target/release/4c --help
```

### Any platform: "Cargo not found"

Rust is not installed or not in PATH. Install from https://rustup.rs/ and restart your terminal.

## Build Variants

### Debug Build
```bash
cargo build
```
- Faster compilation
- Slower execution (~10-20% slower IOPS)
- Includes debug symbols
- Useful for development/testing

### Release Build
```bash
cargo build --release
```
- Slower compilation (~1-2 minutes)
- Optimized binary (3x+ faster in some cases)
- Stripped debug symbols
- Recommended for benchmarking

### Stripped Release Binary (Linux only)
```bash
cargo build --release
strip target/release/4c
```
Reduces binary size further.

## Clean Build

To remove all build artifacts and start fresh:

```bash
cargo clean
cargo build --release
```

## Development

### View code structure
```bash
# List all modules
cargo tree

# Check for issues
cargo clippy

# Format code
cargo fmt
```

### Run with verbose output
```bash
RUST_LOG=debug cargo run --release -- --device ... --duration 10
```

## Cross-Compilation (Building Linux Binary on Windows)

### Option 1: WSL2 (Recommended)

If you have Windows Subsystem for Linux installed:

```powershell
# From PowerShell, launch into WSL
wsl -d Ubuntu  # or your WSL distro name
```

Inside WSL:
```bash
# Navigate to the 4c directory
cd /mnt/c/Users/Jar/Dropbox/VSCode2/4c

# Build release binary
cargo build --release

# Binary will be at: target/release/4c
```

**Advantages:**
- Native Linux environment
- Full io_uring support
- Fastest cross-compilation option
- No Docker required

**Requirements:**
- WSL2 installed with a Linux distribution (Ubuntu, Debian, Fedora, etc.)
- Rust installed in WSL (separate from Windows Rust)

### Option 2: Docker + cross tool

If you have Docker Desktop installed:

```powershell
# Install cross (one-time)
cargo install cross

# Build Linux binary
cd C:\Users\Jar\Dropbox\VSCode2\4c
cross build --release --target x86_64-unknown-linux-gnu
```

Binary: `target/x86_64-unknown-linux-gnu/release/4c`

**Advantages:**
- Works without WSL2
- Minimal setup

**Requirements:**
- Docker Desktop running
- ~500MB+ disk space

### Option 3: Virtual Machine

Set up a Linux VM in VirtualBox, Hyper-V, or similar, and build natively on that platform. Most straightforward but slowest option.

## Performance Notes

- **Release builds are essential for benchmarking** — debug binaries run ~10-20% slower IOPS
- **Windows IOCP** requires FILE_FLAG_OVERLAPPED on device open (implemented)
- **Linux io_uring** requires kernel 5.1+ (check with `uname -r`)
- **Aligned I/O buffers** ensure direct I/O compatibility on all platforms
- **Linux binaries built on Windows** — Use WSL2 for the simplest, fastest cross-compilation experience
