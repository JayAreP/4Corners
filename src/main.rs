mod cli;
mod engine;
mod report;

use clap::Parser;
use cli::Args;
use engine::TestConfig;
use report::BenchmarkReport;
use std::path::Path;

fn main() {
    let args = Args::parse();

    println!("4Corners Disk Benchmark (Rust)");
    println!("==============================");
    println!();

    // Create file device if requested
    if args.create_file {
        if let Err(e) = engine::create_file_device(&args.device, args.file_size) {
            eprintln!("Error creating file device: {}", e);
            std::process::exit(1);
        }
        println!("File device created successfully");
        println!();
    }

    // Prep device if requested
    if args.prep {
        if let Err(e) = engine::prep_device(&args.device) {
            eprintln!("Error preparing device: {}", e);
            std::process::exit(1);
        }
        println!("Device prepared successfully");
        println!();
    }

    // Determine which tests to run
    let run_all = args.tests == "all";
    let run_read_tp = run_all || args.tests.contains("read-tp");
    let run_write_tp = run_all || args.tests.contains("write-tp");
    let run_read_iops = run_all || args.tests.contains("read-iops");
    let run_write_iops = run_all || args.tests.contains("write-iops");

    let mut report = BenchmarkReport::new(&args.device);

    println!("Starting benchmark tests...");
    println!();

    // Read Throughput
    if run_read_tp {
        println!("Running Read Throughput Test...");
        let config = TestConfig {
            device_path: args.device.clone(),
            io_size: args.read_tp_bs as u64 * 1024,
            threads: args.read_tp_threads,
            queue_depth: args.read_tp_qd,
            duration_secs: args.duration,
            is_write: false,
        };
        match engine::run_test(&config) {
            Ok(result) => report.read_throughput = Some(result),
            Err(e) => eprintln!("Read throughput error: {}", e),
        }
        println!();
    }

    // Write Throughput
    if run_write_tp {
        println!("Running Write Throughput Test...");
        let config = TestConfig {
            device_path: args.device.clone(),
            io_size: args.write_tp_bs as u64 * 1024,
            threads: args.write_tp_threads,
            queue_depth: args.write_tp_qd,
            duration_secs: args.duration,
            is_write: true,
        };
        match engine::run_test(&config) {
            Ok(result) => report.write_throughput = Some(result),
            Err(e) => eprintln!("Write throughput error: {}", e),
        }
        println!();
    }

    // Read IOPS
    if run_read_iops {
        println!("Running Read IOPS Test...");
        let config = TestConfig {
            device_path: args.device.clone(),
            io_size: args.read_iops_bs as u64 * 1024,
            threads: args.read_iops_threads,
            queue_depth: args.read_iops_qd,
            duration_secs: args.duration,
            is_write: false,
        };
        match engine::run_test(&config) {
            Ok(result) => report.read_iops = Some(result),
            Err(e) => eprintln!("Read IOPS error: {}", e),
        }
        println!();
    }

    // Write IOPS
    if run_write_iops {
        println!("Running Write IOPS Test...");
        let config = TestConfig {
            device_path: args.device.clone(),
            io_size: args.write_iops_bs as u64 * 1024,
            threads: args.write_iops_threads,
            queue_depth: args.write_iops_qd,
            duration_secs: args.duration,
            is_write: true,
        };
        match engine::run_test(&config) {
            Ok(result) => report.write_iops = Some(result),
            Err(e) => eprintln!("Write IOPS error: {}", e),
        }
        println!();
    }

    println!("Benchmark completed!");
    println!();
    println!("{}", report.generate_text_report());

    if let Err(e) = report.save(Path::new(".")) {
        eprintln!("Warning: failed to save reports: {}", e);
    }
}
