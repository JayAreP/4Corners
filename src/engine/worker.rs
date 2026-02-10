use std::io;
use std::sync::atomic::AtomicBool;

use super::Metrics;

/// Main worker entry point - dispatches to platform-specific async I/O
pub fn run_worker(
    _thread_id: u32,
    device_path: &str,
    io_size: u64,
    queue_depth: u32,
    is_write: bool,
    test_range: u64,
    stop: &AtomicBool,
    metrics: &Metrics,
) -> io::Result<()> {
    #[cfg(windows)]
    {
        super::platform_windows::worker_iocp(
            device_path, io_size, queue_depth, is_write, test_range, stop, metrics,
        )
    }

    #[cfg(target_os = "linux")]
    {
        super::platform_linux::worker_io_uring(
            device_path, io_size, queue_depth, is_write, test_range, stop, metrics,
        )
    }

    #[cfg(not(any(windows, target_os = "linux")))]
    {
        Err(io::Error::new(
            io::ErrorKind::Unsupported,
            "Platform not supported",
        ))
    }
}
