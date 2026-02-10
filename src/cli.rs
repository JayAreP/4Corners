use clap::Parser;

#[derive(Parser, Debug, Clone)]
#[command(name = "4c", about = "4Corners Disk Benchmark - CLI")]
pub struct Args {
    /// Device or file path(s) - can specify multiple times or comma-separated
    /// On Windows: use \\.\PhysicalDrive4 or just 4
    #[arg(short, long)]
    pub device: Vec<String>,

    /// Test duration in seconds
    #[arg(long, default_value_t = 30)]
    pub duration: u32,

    /// Read throughput threads
    #[arg(long, default_value_t = 30)]
    pub read_tp_threads: u32,

    /// Write throughput threads
    #[arg(long, default_value_t = 16)]
    pub write_tp_threads: u32,

    /// Read IOPS threads
    #[arg(long, default_value_t = 120)]
    pub read_iops_threads: u32,

    /// Write IOPS threads
    #[arg(long, default_value_t = 120)]
    pub write_iops_threads: u32,

    /// Read throughput queue depth per thread
    #[arg(long, default_value_t = 1)]
    pub read_tp_qd: u32,

    /// Write throughput queue depth per thread
    #[arg(long, default_value_t = 1)]
    pub write_tp_qd: u32,

    /// Read IOPS queue depth per thread
    #[arg(long, default_value_t = 1)]
    pub read_iops_qd: u32,

    /// Write IOPS queue depth per thread
    #[arg(long, default_value_t = 1)]
    pub write_iops_qd: u32,

    /// Read throughput block size (KB)
    #[arg(long, default_value_t = 128)]
    pub read_tp_bs: u32,

    /// Write throughput block size (KB)
    #[arg(long, default_value_t = 64)]
    pub write_tp_bs: u32,

    /// Read IOPS block size (KB)
    #[arg(long, default_value_t = 4)]
    pub read_iops_bs: u32,

    /// Write IOPS block size (KB)
    #[arg(long, default_value_t = 4)]
    pub write_iops_bs: u32,

    /// Prep device before testing (writes random data)
    #[arg(long)]
    pub prep: bool,

    /// Create a file device before testing
    #[arg(long)]
    pub create_file: bool,

    /// File device size in GB (if creating)
    #[arg(long, default_value_t = 10)]
    pub file_size: u64,

    /// Tests to run: all, read-tp, write-tp, read-iops, write-iops (comma-separated)
    #[arg(long, default_value = "all")]
    pub tests: String,
}
