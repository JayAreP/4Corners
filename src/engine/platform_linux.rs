use std::io;
use std::os::unix::io::{AsRawFd, RawFd};

/// Wrapper around a raw Linux file descriptor opened with O_DIRECT
pub struct DeviceHandle {
    fd: RawFd,
}

unsafe impl Send for DeviceHandle {}
unsafe impl Sync for DeviceHandle {}

impl Drop for DeviceHandle {
    fn drop(&mut self) {
        unsafe { libc::close(self.fd) };
    }
}

impl AsRawFd for DeviceHandle {
    fn as_raw_fd(&self) -> RawFd {
        self.fd
    }
}

/// Open device for reading with O_DIRECT
pub fn open_device_read(path: &str) -> io::Result<DeviceHandle> {
    open_device(path, false)
}

/// Open device for writing with O_DIRECT
pub fn open_device_write(path: &str) -> io::Result<DeviceHandle> {
    open_device(path, true)
}

fn open_device(path: &str, write: bool) -> io::Result<DeviceHandle> {
    let c_path = std::ffi::CString::new(path).unwrap();
    let flags = if write {
        libc::O_RDWR | libc::O_DIRECT
    } else {
        libc::O_RDONLY | libc::O_DIRECT
    };

    let fd = unsafe { libc::open(c_path.as_ptr(), flags) };
    if fd < 0 {
        return Err(io::Error::last_os_error());
    }

    Ok(DeviceHandle { fd })
}

/// Get device or file size
pub fn get_device_size(path: &str) -> io::Result<u64> {
    // Try as regular file first
    if let Ok(meta) = std::fs::metadata(path) {
        if meta.len() > 0 {
            return Ok(meta.len());
        }
    }

    // Try as block device using ioctl
    let c_path = std::ffi::CString::new(path).unwrap();
    let fd = unsafe { libc::open(c_path.as_ptr(), libc::O_RDONLY) };
    if fd < 0 {
        return Err(io::Error::last_os_error());
    }

    let mut size: u64 = 0;
    // BLKGETSIZE64 = 0x80081272 on x86_64
    const BLKGETSIZE64: libc::c_ulong = 0x80081272;
    let result = unsafe { libc::ioctl(fd, BLKGETSIZE64, &mut size) };
    unsafe { libc::close(fd) };

    if result < 0 {
        return Err(io::Error::last_os_error());
    }

    Ok(size)
}

/// Synchronous read at offset (for prep/simple operations)
pub fn read_at_raw(dev: &DeviceHandle, buf: &super::AlignedBuf, offset: u64) -> io::Result<u32> {
    let result = unsafe {
        libc::pread(dev.fd, buf.ptr as *mut libc::c_void, buf.len, offset as i64)
    };
    if result < 0 {
        return Err(io::Error::last_os_error());
    }
    Ok(result as u32)
}

/// Synchronous write at offset (for prep/simple operations)
pub fn write_at_raw(dev: &DeviceHandle, buf: &super::AlignedBuf, offset: u64) -> io::Result<u32> {
    let result = unsafe {
        libc::pwrite(dev.fd, buf.ptr as *const libc::c_void, buf.len, offset as i64)
    };
    if result < 0 {
        return Err(io::Error::last_os_error());
    }
    Ok(result as u32)
}

/// io_uring-based async I/O worker for maximum IOPS
pub fn worker_io_uring(
    device_path: &str,
    io_size: u64,
    queue_depth: u32,
    is_write: bool,
    test_range: u64,
    stop: &std::sync::atomic::AtomicBool,
    metrics: &super::Metrics,
) -> io::Result<()> {
    use io_uring::{opcode, types, IoUring};
    use std::sync::atomic::Ordering;

    let dev = if is_write {
        open_device_write(device_path)?
    } else {
        open_device_read(device_path)?
    };

    let qd = queue_depth as usize;
    let sector_size: usize = 4096;
    let max_offset = test_range / io_size;

    // Create io_uring instance
    let mut ring = IoUring::new(queue_depth)?;

    // Allocate aligned buffers per slot
    let mut buffers: Vec<super::AlignedBuf> = Vec::with_capacity(qd);
    for _ in 0..qd {
        let mut buf = super::alloc_aligned(io_size as usize, sector_size);
        if is_write {
            for chunk in buf.as_mut_slice().chunks_mut(8) {
                let val = rand::random::<u64>();
                let bytes = val.to_le_bytes();
                let len = chunk.len().min(8);
                chunk[..len].copy_from_slice(&bytes[..len]);
            }
        }
        buffers.push(buf);
    }

    // Pre-generate random offsets
    let mut offsets: Vec<u64> = Vec::with_capacity(16384);
    for _ in 0..16384 {
        let rand_val = rand::random::<u64>();
        let block_num = rand_val % max_offset;
        offsets.push(block_num * io_size);
    }
    let mut offset_idx: usize = 0;

    // Track start times
    let mut start_times: Vec<std::time::Instant> = vec![std::time::Instant::now(); qd];

    // Submit initial batch
    {
        let sq = ring.submission();
        for slot in 0..qd {
            let off = offsets[offset_idx];
            offset_idx = (offset_idx + 1) % offsets.len();
            start_times[slot] = std::time::Instant::now();

            let entry = if is_write {
                opcode::Write::new(
                    types::Fd(dev.fd),
                    buffers[slot].ptr,
                    io_size as u32,
                )
                .offset(off)
                .build()
                .user_data(slot as u64)
            } else {
                opcode::Read::new(
                    types::Fd(dev.fd),
                    buffers[slot].ptr,
                    io_size as u32,
                )
                .offset(off)
                .build()
                .user_data(slot as u64)
            };

            unsafe { sq.push(&entry).ok() };
        }
    }
    ring.submit()?;

    let mut local_ops: u64 = 0;
    let mut local_bytes: u64 = 0;
    let batch_size: u64 = 256;
    let mut op_count: u64 = 0;

    while !stop.load(Ordering::Relaxed) {
        // Wait for at least 1 completion
        ring.submit_and_wait(1)?;

        // Process all available completions
        let cq = ring.completion();
        for cqe in cq {
            let slot = cqe.user_data() as usize;
            let result = cqe.result();

            if result > 0 {
                op_count += 1;
                if op_count % 64 == 0 {
                    let lat_ns = start_times[slot].elapsed().as_nanos() as u64;
                    metrics.record_latency(lat_ns);
                }

                local_ops += 1;
                local_bytes += result as u64;
            }

            // Reissue I/O on this slot
            let off = offsets[offset_idx];
            offset_idx = (offset_idx + 1) % offsets.len();
            start_times[slot] = std::time::Instant::now();

            let entry = if is_write {
                opcode::Write::new(
                    types::Fd(dev.fd),
                    buffers[slot].ptr,
                    io_size as u32,
                )
                .offset(off)
                .build()
                .user_data(slot as u64)
            } else {
                opcode::Read::new(
                    types::Fd(dev.fd),
                    buffers[slot].ptr,
                    io_size as u32,
                )
                .offset(off)
                .build()
                .user_data(slot as u64)
            };

            unsafe { ring.submission().push(&entry).ok() };
        }

        // Batch update metrics
        if local_ops >= batch_size {
            metrics.total_ops.fetch_add(local_ops, Ordering::Relaxed);
            metrics.total_bytes.fetch_add(local_bytes, Ordering::Relaxed);
            local_ops = 0;
            local_bytes = 0;
        }
    }

    // Flush remaining
    if local_ops > 0 {
        metrics.total_ops.fetch_add(local_ops, Ordering::Relaxed);
        metrics.total_bytes.fetch_add(local_bytes, Ordering::Relaxed);
    }

    Ok(())
}
