use std::io;
use std::os::windows::ffi::OsStrExt;
use std::ffi::OsStr;
use std::ptr;
use windows_sys::Win32::Foundation::*;
use windows_sys::Win32::Storage::FileSystem::*;
use windows_sys::Win32::System::IO::*;
use windows_sys::Win32::System::Threading::*;

/// Wrapper around a raw Windows HANDLE
pub struct DeviceHandle {
    pub handle: HANDLE,
}

unsafe impl Send for DeviceHandle {}
unsafe impl Sync for DeviceHandle {}

impl Drop for DeviceHandle {
    fn drop(&mut self) {
        if self.handle != INVALID_HANDLE_VALUE {
            unsafe { CloseHandle(self.handle) };
        }
    }
}

fn to_wide(s: &str) -> Vec<u16> {
    OsStr::new(s).encode_wide().chain(std::iter::once(0)).collect()
}

/// Open device for reading with direct I/O + overlapped
pub fn open_device_read(path: &str) -> io::Result<DeviceHandle> {
    open_device(path, false)
}

/// Open device for writing with direct I/O + overlapped
pub fn open_device_write(path: &str) -> io::Result<DeviceHandle> {
    open_device(path, true)
}

fn open_device(path: &str, write: bool) -> io::Result<DeviceHandle> {
    let wide_path = to_wide(path);
    let access = if write {
        GENERIC_READ | GENERIC_WRITE
    } else {
        GENERIC_READ
    };

    let flags = FILE_FLAG_NO_BUFFERING | FILE_FLAG_WRITE_THROUGH | FILE_FLAG_OVERLAPPED;

    let handle = unsafe {
        CreateFileW(
            wide_path.as_ptr(),
            access,
            FILE_SHARE_READ | FILE_SHARE_WRITE,
            ptr::null(),
            OPEN_EXISTING,
            flags,
            ptr::null_mut(),
        )
    };

    if handle == INVALID_HANDLE_VALUE {
        return Err(io::Error::last_os_error());
    }

    Ok(DeviceHandle { handle })
}

/// Get device or file size
pub fn get_device_size(path: &str) -> io::Result<u64> {
    // Try as regular file first
    if let Ok(meta) = std::fs::metadata(path) {
        if meta.len() > 0 {
            return Ok(meta.len());
        }
    }

    // Try as device - use IOCTL_DISK_GET_LENGTH_INFO
    let wide_path = to_wide(path);
    let handle = unsafe {
        CreateFileW(
            wide_path.as_ptr(),
            GENERIC_READ,
            FILE_SHARE_READ | FILE_SHARE_WRITE,
            ptr::null(),
            OPEN_EXISTING,
            0,
            ptr::null_mut(),
        )
    };

    if handle == INVALID_HANDLE_VALUE {
        return Err(io::Error::last_os_error());
    }

    // IOCTL_DISK_GET_LENGTH_INFO = 0x0007405C
    const IOCTL_DISK_GET_LENGTH_INFO: u32 = 0x0007405C;
    let mut length: i64 = 0;
    let mut bytes_returned: u32 = 0;

    let result = unsafe {
        DeviceIoControl(
            handle,
            IOCTL_DISK_GET_LENGTH_INFO,
            ptr::null(),
            0,
            &mut length as *mut i64 as *mut _,
            std::mem::size_of::<i64>() as u32,
            &mut bytes_returned,
            ptr::null_mut(),
        )
    };

    unsafe { CloseHandle(handle) };

    if result == 0 {
        return Err(io::Error::last_os_error());
    }

    Ok(length as u64)
}

/// Synchronous read at offset (for prep/simple operations)
pub fn read_at_raw(dev: &DeviceHandle, buf: &super::AlignedBuf, offset: u64) -> io::Result<u32> {
    let mut overlapped: OVERLAPPED = unsafe { std::mem::zeroed() };
    overlapped.Anonymous.Anonymous.Offset = offset as u32;
    overlapped.Anonymous.Anonymous.OffsetHigh = (offset >> 32) as u32;

    let event = unsafe { CreateEventW(ptr::null(), 1, 0, ptr::null()) };
    overlapped.hEvent = event;

    let mut bytes_read: u32 = 0;
    let result = unsafe {
        ReadFile(
            dev.handle,
            buf.ptr as *mut _,
            buf.len as u32,
            &mut bytes_read,
            &mut overlapped,
        )
    };

    if result == 0 {
        let err = unsafe { GetLastError() };
        if err == ERROR_IO_PENDING {
            unsafe {
                GetOverlappedResult(dev.handle, &overlapped, &mut bytes_read, 1);
            }
        } else {
            unsafe { CloseHandle(event) };
            return Err(io::Error::from_raw_os_error(err as i32));
        }
    }

    unsafe { CloseHandle(event) };
    Ok(bytes_read)
}

/// Synchronous write at offset (for prep/simple operations)
pub fn write_at_raw(dev: &DeviceHandle, buf: &super::AlignedBuf, offset: u64) -> io::Result<u32> {
    let mut overlapped: OVERLAPPED = unsafe { std::mem::zeroed() };
    overlapped.Anonymous.Anonymous.Offset = offset as u32;
    overlapped.Anonymous.Anonymous.OffsetHigh = (offset >> 32) as u32;

    let event = unsafe { CreateEventW(ptr::null(), 1, 0, ptr::null()) };
    overlapped.hEvent = event;

    let mut bytes_written: u32 = 0;
    let result = unsafe {
        WriteFile(
            dev.handle,
            buf.ptr as *const _,
            buf.len as u32,
            &mut bytes_written,
            &mut overlapped,
        )
    };

    if result == 0 {
        let err = unsafe { GetLastError() };
        if err == ERROR_IO_PENDING {
            unsafe {
                GetOverlappedResult(dev.handle, &overlapped, &mut bytes_written, 1);
            }
        } else {
            unsafe { CloseHandle(event) };
            return Err(io::Error::from_raw_os_error(err as i32));
        }
    }

    unsafe { CloseHandle(event) };
    Ok(bytes_written)
}

/// IOCP-based async I/O worker for maximum IOPS
/// Each call submits `queue_depth` overlapped I/Os and polls for completion
pub fn worker_iocp(
    device_path: &str,
    io_size: u64,
    queue_depth: u32,
    is_write: bool,
    test_range: u64,
    stop: &std::sync::atomic::AtomicBool,
    metrics: &super::Metrics,
) -> io::Result<()> {
    let dev = if is_write {
        open_device_write(device_path)?
    } else {
        open_device_read(device_path)?
    };

    // Create IOCP and associate the file handle
    let iocp = unsafe { CreateIoCompletionPort(dev.handle, ptr::null_mut(), 0, 0) };
    if iocp.is_null() {
        return Err(io::Error::last_os_error());
    }

    let qd = queue_depth as usize;
    let sector_size: u64 = 4096;
    let max_offset = test_range / io_size;

    // Allocate aligned buffers and overlapped structures per slot
    let mut buffers: Vec<super::AlignedBuf> = Vec::with_capacity(qd);
    let mut overlappeds: Vec<OVERLAPPED> = Vec::with_capacity(qd);

    for _ in 0..qd {
        let mut buf = super::alloc_aligned(io_size as usize, sector_size as usize);
        // Fill write buffers with random data
        if is_write {
            for chunk in buf.as_mut_slice().chunks_mut(8) {
                let val = rand::random::<u64>();
                let bytes = val.to_le_bytes();
                let len = chunk.len().min(8);
                chunk[..len].copy_from_slice(&bytes[..len]);
            }
        }
        buffers.push(buf);
        overlappeds.push(unsafe { std::mem::zeroed() });
    }

    // Pre-generate random offsets
    let mut offsets: Vec<i64> = Vec::with_capacity(16384);
    for _ in 0..16384 {
        let rand_val = rand::random::<u64>();
        let block_num = rand_val % max_offset;
        offsets.push((block_num * io_size) as i64);
    }
    let mut offset_idx: usize = 0;

    // Track start times for latency measurement
    let mut start_times: Vec<std::time::Instant> = vec![std::time::Instant::now(); qd];

    // Submit initial batch of I/Os
    for slot in 0..qd {
        let off = offsets[offset_idx] as u64;
        offset_idx = (offset_idx + 1) % offsets.len();

        overlappeds[slot].Anonymous.Anonymous.Offset = off as u32;
        overlappeds[slot].Anonymous.Anonymous.OffsetHigh = (off >> 32) as u32;
        start_times[slot] = std::time::Instant::now();

        if is_write {
            unsafe {
                WriteFile(
                    dev.handle,
                    buffers[slot].ptr as *const _,
                    io_size as u32,
                    ptr::null_mut(),
                    &mut overlappeds[slot],
                );
            }
        } else {
            unsafe {
                ReadFile(
                    dev.handle,
                    buffers[slot].ptr as *mut _,
                    io_size as u32,
                    ptr::null_mut(),
                    &mut overlappeds[slot],
                );
            }
        }
    }

    // Completion loop
    let mut local_ops: u64 = 0;
    let mut local_bytes: u64 = 0;
    let batch_size: u64 = 256;
    let mut op_count: u64 = 0;

    while !stop.load(std::sync::atomic::Ordering::Relaxed) {
        let mut bytes_transferred: u32 = 0;
        let mut completion_key: usize = 0;
        let mut overlapped_ptr: *mut OVERLAPPED = ptr::null_mut();

        // Wait for one completion (short timeout to check stop flag)
        let result = unsafe {
            GetQueuedCompletionStatus(
                iocp,
                &mut bytes_transferred,
                &mut completion_key,
                &mut overlapped_ptr,
                1, // 1ms timeout
            )
        };

        if result == 0 {
            let err = unsafe { GetLastError() };
            if err == WAIT_TIMEOUT {
                continue;
            }
            if overlapped_ptr.is_null() {
                continue;
            }
            // I/O error on this slot - skip and reissue
        }

        if overlapped_ptr.is_null() {
            continue;
        }

        // Find which slot completed
        let slot = {
            let base = overlappeds.as_ptr() as usize;
            let completed = overlapped_ptr as usize;
            (completed - base) / std::mem::size_of::<OVERLAPPED>()
        };

        if slot >= qd {
            continue;
        }

        // Record latency (sample every 64th operation)
        op_count += 1;
        if op_count % 64 == 0 {
            let lat_ns = start_times[slot].elapsed().as_nanos() as u64;
            metrics.record_latency(lat_ns);
        }

        local_ops += 1;
        local_bytes += bytes_transferred as u64;

        // Batch update metrics
        if local_ops >= batch_size {
            metrics
                .total_ops
                .fetch_add(local_ops, std::sync::atomic::Ordering::Relaxed);
            metrics
                .total_bytes
                .fetch_add(local_bytes, std::sync::atomic::Ordering::Relaxed);
            local_ops = 0;
            local_bytes = 0;
        }

        // Reissue I/O on the completed slot
        let off = offsets[offset_idx] as u64;
        offset_idx = (offset_idx + 1) % offsets.len();

        overlappeds[slot] = unsafe { std::mem::zeroed() };
        overlappeds[slot].Anonymous.Anonymous.Offset = off as u32;
        overlappeds[slot].Anonymous.Anonymous.OffsetHigh = (off >> 32) as u32;
        start_times[slot] = std::time::Instant::now();

        if is_write {
            unsafe {
                WriteFile(
                    dev.handle,
                    buffers[slot].ptr as *const _,
                    io_size as u32,
                    ptr::null_mut(),
                    &mut overlappeds[slot],
                );
            }
        } else {
            unsafe {
                ReadFile(
                    dev.handle,
                    buffers[slot].ptr as *mut _,
                    io_size as u32,
                    ptr::null_mut(),
                    &mut overlappeds[slot],
                );
            }
        }
    }

    // Flush remaining local counters
    if local_ops > 0 {
        metrics
            .total_ops
            .fetch_add(local_ops, std::sync::atomic::Ordering::Relaxed);
        metrics
            .total_bytes
            .fetch_add(local_bytes, std::sync::atomic::Ordering::Relaxed);
    }

    // Cancel any outstanding I/Os
    unsafe { CancelIo(dev.handle) };

    // Drain remaining completions
    loop {
        let mut bytes: u32 = 0;
        let mut key: usize = 0;
        let mut olp: *mut OVERLAPPED = ptr::null_mut();
        let r = unsafe { GetQueuedCompletionStatus(iocp, &mut bytes, &mut key, &mut olp, 0) };
        if r == 0 || olp.is_null() {
            break;
        }
    }

    unsafe { CloseHandle(iocp) };
    Ok(())
}
