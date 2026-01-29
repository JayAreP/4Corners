# 4Corners Troubleshooting Guide

## Physical Device Access Issues on Windows

### Problem: CLI shows 0 MB/s and 0 IOPS when testing physical drives

**Symptoms:**
```
./4cornerscli.exe -device \\.\PhysicalDrive2
Starting benchmark tests...
Running Read Throughput Test...
  25s: 0.00 MB/s | 0 IOPS | 0.00 ms
```

**Root Cause:**
Windows requires **Administrator privileges** to access physical drives (\\.\PhysicalDrive*) with direct I/O (FILE_FLAG_NO_BUFFERING).

**Solutions:**

1. **Run as Administrator (Recommended):**
   - Right-click Command Prompt or PowerShell
   - Select "Run as Administrator"
   - Navigate to the 4corners directory
   - Run the command again:
     ```cmd
     4cornerscli.exe -device \\.\PhysicalDrive2
     ```

2. **Use File-Based Testing (Alternative):**
   If you can't get admin rights, test against a file instead:
   ```cmd
   4cornerscli.exe -device testfile.bin -create-file -file-size 10
   ```

3. **Verify Device Access:**
   Test if you can access the device:
   ```powershell
   # Run as Administrator
   $handle = [System.IO.File]::Open("\\.\PhysicalDrive2", "Open", "Read", "ReadWrite")
   $handle.Close()
   ```

---

## Fyne GUI OpenGL Error

### Problem: GUI fails to start with OpenGL error

**Symptoms:**
```
Fyne error: window creation error
cause: APIUnavailable: WGL: The driver does not appear to support OpenGL
```

**Root Cause:**
The Fyne GUI framework requires OpenGL 2.0+ support, which may not be available in:
- Virtual machines without 3D acceleration
- Remote desktop sessions
- Systems with outdated/missing graphics drivers

**Quick Solutions:**

1. **Use CLI Instead (Recommended):**
   The CLI version has the same functionality without GUI requirements:
   ```cmd
   4cornerscli.exe -device \\.\PhysicalDrive2 -duration 60
   ```

2. **Enable 3D Acceleration (VMs):**
   - VMware: Enable "Accelerate 3D graphics" in VM settings
   - VirtualBox: Enable "3D Acceleration" in Display settings
   - Hyper-V: Enhanced Session Mode may help

3. **Update Graphics Drivers:**
   - Download latest drivers from GPU manufacturer (NVIDIA, AMD, Intel)
   - Restart after installation

4. **Use Software Rendering (Last Resort):**
   Set environment variable before running:
   ```cmd
   set LIBGL_ALWAYS_SOFTWARE=1
   4corners.exe
   ```
   Note: This will be slow and may not work on all systems.

5. **Run Locally Instead of RDP:**
   If accessing via Remote Desktop, try running locally on the machine or use the CLI via SSH/remote PowerShell.

**For detailed GUI troubleshooting and requirements, see [GUI-REQUIREMENTS.md](GUI-REQUIREMENTS.md)**

---

## Permission Errors

### Error: "Access is denied" or "failed to open device"

**Solutions:**
1. Run as Administrator (see above)
2. Close any applications that might have the device locked (disk management, backup tools, etc.)
3. Ensure the device exists: `Get-PhysicalDisk` in PowerShell

---

## Performance Issues

### Test shows unexpectedly low performance

**Checklist:**
1. Ensure the device is not in use by other applications
2. Disable antivirus real-time scanning temporarily
3. Use appropriate thread counts (see CLI-README.md)
4. Verify storage controller is configured for performance (not power saving)
5. Check if write-cache is enabled on the device

---

## Build Issues

### CGo compilation errors

The GUI version requires CGo and a C compiler. If you only need the CLI:

```bash
# Build CLI only (no CGo required)
cd cmd/4cornerscli
go build -o 4cornerscli.exe
```

Or use the pre-built binaries in the releases.
