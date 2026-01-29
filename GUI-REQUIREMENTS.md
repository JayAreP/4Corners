# 4Corners GUI Requirements

The 4Corners GUI application uses the Fyne framework, which requires **OpenGL 2.0 or higher**.

## System Requirements

### Windows
- **OpenGL 2.0+** capable graphics card and drivers
- Windows 7 or later
- Graphics drivers installed and up to date
- Not suitable for: Windows Server Core, Hyper-V VMs without RemoteFX, basic RDP sessions

### Linux
- **OpenGL 2.0+** capable graphics card and drivers
- X11 or Wayland display server
- Graphics drivers installed (mesa, nvidia, or amd)

### Important Notes
- **Virtual Machines**: Requires 3D acceleration enabled in VM settings
- **Remote Desktop**: May not work over standard RDP/VNC without GPU passthrough
- **Headless Servers**: Cannot run GUI (use CLI version instead)

## Common OpenGL Issues

### Issue: "WGL: The driver does not appear to support OpenGL"

**This means:** Your system doesn't have OpenGL 2.0+ available.

**Common Scenarios:**
1. **Virtual Machine without 3D acceleration**
   - VMware: Enable "Accelerate 3D graphics" in VM settings (requires VMware Tools)
   - VirtualBox: Enable "3D Acceleration" in Display settings (requires Guest Additions)
   - Hyper-V: Enable RemoteFX vGPU (Enhanced Session Mode)
   - QEMU/KVM: Use virtio-gpu with virgl enabled

2. **Remote Desktop Session**
   - Standard Windows RDP doesn't provide GPU access
   - Use alternate methods: VNC with GPU passthrough, Parsec, or local access
   - Better option: Use the CLI version via SSH/PowerShell remoting

3. **Missing/Outdated Graphics Drivers**
   - Update from manufacturer: NVIDIA, AMD, Intel
   - Windows: Use Device Manager or manufacturer's driver software
   - Linux: Install mesa-utils and check with `glxinfo | grep "OpenGL version"`

4. **Windows Server Core or Minimal Installations**
   - GUI components not installed
   - Use the CLI version

## Solutions

### Option 1: Use the CLI Version (Recommended)

The CLI has **identical functionality** to the GUI:

```bash
# Windows (as Administrator)
4cornerscli.exe -device \\.\PhysicalDrive1

# Linux (as root/sudo)
sudo ./4cornerscli -device /dev/sdb
```

**Advantages:**
- No OpenGL requirement
- Works over SSH/remote sessions
- Suitable for automation/scripting
- Smaller resource footprint

See [CLI-README.md](CLI-README.md) for full documentation.

### Option 2: Enable 3D Acceleration (VMs)

#### VMware
1. Power off the VM
2. Edit VM Settings → Display
3. Check "Accelerate 3D graphics"
4. Set video memory to at least 128MB
5. Install/update VMware Tools
6. Restart the VM

#### VirtualBox
1. Power off the VM
2. Settings → Display
3. Check "Enable 3D Acceleration"
4. Set video memory to at least 128MB
5. Install/update Guest Additions
6. Restart the VM

**Note:** May require host system to have GPU acceleration available.

### Option 3: Update Graphics Drivers

#### Windows
```powershell
# Check current driver
Get-WmiObject Win32_VideoController | Select-Object Name, DriverVersion

# Update options:
# 1. Windows Update
# 2. Device Manager → Display adapters → Update driver
# 3. Download from manufacturer website
```

#### Linux
```bash
# Check OpenGL support
glxinfo | grep "OpenGL version"

# Install/update Mesa (Intel/AMD)
sudo apt install mesa-utils libgl1-mesa-dri  # Debian/Ubuntu
sudo dnf install mesa-dri-drivers            # Fedora/RHEL

# NVIDIA proprietary drivers
ubuntu-drivers devices                        # Check available drivers
sudo ubuntu-drivers autoinstall              # Install recommended
```

### Option 4: Software Rendering (Last Resort)

**Warning:** This will be **very slow** and may not work on all systems.

#### Windows
```cmd
set LIBGL_ALWAYS_SOFTWARE=1
4corners.exe
```

#### Linux
```bash
LIBGL_ALWAYS_SOFTWARE=1 ./4corners
```

## Verification

### Windows
Check if OpenGL is available:
```powershell
# Download and run OpenGL Extensions Viewer or GPU-Z
# Or check in Device Manager → Display adapters
```

### Linux
```bash
# Check OpenGL version
glxinfo | grep "OpenGL version"

# Test OpenGL with glxgears
glxgears

# Check 3D acceleration is working
glxinfo | grep "direct rendering"
# Should output: direct rendering: Yes
```

## Build Requirements

If building from source, the GUI version requires:
- Go 1.21 or later
- GCC/MinGW (for CGo)
- Graphics libraries (handled by Fyne)

The CLI version has **no special requirements** and can be built with just Go.

## Getting Help

If you continue to have issues:

1. Check your error message in [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
2. Verify OpenGL support using the verification commands above
3. Consider using the CLI version for production use
4. Check the [Fyne documentation](https://docs.fyne.io/) for platform-specific issues

## Why Does the CLI Work When GUI Doesn't?

- **GUI**: Uses Fyne framework → Requires OpenGL for rendering
- **CLI**: Pure Go + syscalls → No graphics dependencies

Both versions have identical benchmarking capabilities. The GUI just provides a visual interface for configuration and real-time graphs.
