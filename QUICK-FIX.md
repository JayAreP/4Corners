# Quick Fix Guide for 4Corners Issues

## Issue 1: CLI shows 0 MB/s and 0 IOPS on Windows Physical Drives

### Quick Fix
**Run as Administrator!**

```cmd
# Right-click Command Prompt or PowerShell
# Select "Run as Administrator"
# Then run:
cd C:\path\to\4corners
4cornerscli.exe -device \\.\PhysicalDrive2
```

### Why?
Windows requires Administrator privileges to access physical drives with direct I/O (FILE_FLAG_NO_BUFFERING).

### Alternative Without Admin Rights
Use file-based testing:
```cmd
4cornerscli.exe -device testfile.bin -create-file -file-size 10
```

---

## Issue 2: GUI Fails with OpenGL Error

### Error Message
```
Fyne error: window creation error
cause: APIUnavailable: WGL: The driver does not appear to support OpenGL
```

### Quick Fix
**Use the CLI version instead:**

```cmd
# Has all the same features, no GUI dependencies
4cornerscli.exe -device \\.\PhysicalDrive2 -duration 60
```

### Root Cause
- The GUI requires OpenGL 2.0+
- Often unavailable in VMs, Remote Desktop, or systems with old/missing graphics drivers

### If You Must Use GUI
1. **VM Users**: Enable 3D acceleration in VM settings
2. **Update Drivers**: Install latest GPU drivers from manufacturer
3. **Run Locally**: Don't use Remote Desktop, run directly on the machine

---

## Verification

### Test if you have proper access (Windows):
```powershell
# Run this as Administrator
Get-PhysicalDisk

# Should list all physical disks
# If you see disks listed, you have proper access
```

### Test the CLI (with proper permissions):
```cmd
# Quick 10-second test
4cornerscli.exe -device \\.\PhysicalDrive2 -duration 10 -tests read-tp
```

If you still see `0.00 MB/s`, check:
1. Are you running as Administrator?
2. Does the device exist? (check with `Get-PhysicalDisk`)
3. Is the device in use? (close Disk Management, backup software, etc.)

---

## Changes Made to Code

The following improvements were made to help diagnose issues:

1. **Better Error Messages**: Now explicitly tells you when Administrator privileges are needed
2. **Early Error Detection**: Checks for device access issues within first 100ms of test
3. **Zero Operation Detection**: Reports helpful error if no I/O operations complete
4. **Detailed Error Types**: Distinguishes between access denied, device not found, and device in use

After rebuilding, the CLI will provide much clearer error messages like:
```
Error: access denied - physical drive access requires Administrator privileges. Please run as Administrator
```

Instead of silently showing 0 MB/s.
