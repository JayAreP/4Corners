use chrono::{DateTime, Local};
use serde::Serialize;
use std::fs;
use std::io;
use std::path::Path;

#[derive(Debug, Clone, Serialize)]
pub struct TestResult {
    pub throughput_mbps: f64,
    pub iops: f64,
    pub latency_avg_us: f64,
    pub latency_p50_us: f64,
    pub latency_p99_us: f64,
    pub threads: u32,
    pub queue_depth: u32,
    pub block_size_kb: u32,
    pub duration_secs: u32,
}

#[derive(Debug, Clone, Serialize)]
pub struct BenchmarkReport {
    pub test_date: DateTime<Local>,
    pub device: String,
    pub read_throughput: Option<TestResult>,
    pub write_throughput: Option<TestResult>,
    pub read_iops: Option<TestResult>,
    pub write_iops: Option<TestResult>,
}

impl BenchmarkReport {
    pub fn new(device: &str) -> Self {
        Self {
            test_date: Local::now(),
            device: device.to_string(),
            read_throughput: None,
            write_throughput: None,
            read_iops: None,
            write_iops: None,
        }
    }

    pub fn generate_text_report(&self) -> String {
        let mut s = String::new();
        s.push_str("========================================\n");
        s.push_str("4Corners Disk Benchmark Report\n");
        s.push_str("========================================\n\n");
        s.push_str(&format!(
            "Test Date: {}\n",
            self.test_date.format("%Y-%m-%d %H:%M:%S")
        ));
        s.push_str(&format!("Device: {}\n\n", self.device));

        if let Some(r) = &self.read_throughput {
            s.push_str("Read Throughput Test:\n");
            format_result(&mut s, r);
        }
        if let Some(r) = &self.write_throughput {
            s.push_str("Write Throughput Test:\n");
            format_result(&mut s, r);
        }
        if let Some(r) = &self.read_iops {
            s.push_str("Read IOPS Test:\n");
            format_result(&mut s, r);
        }
        if let Some(r) = &self.write_iops {
            s.push_str("Write IOPS Test:\n");
            format_result(&mut s, r);
        }

        s.push_str("========================================\n");
        s
    }

    pub fn save(&self, dir: &Path) -> io::Result<()> {
        let timestamp = self.test_date.format("%Y%m%d-%H%M%S");

        let text_path = dir.join(format!("4c-report-{}.txt", timestamp));
        fs::write(&text_path, self.generate_text_report())?;
        println!("Text report saved: {}", text_path.display());

        let json_path = dir.join(format!("4c-report-{}.json", timestamp));
        let json = serde_json::to_string_pretty(self).unwrap();
        fs::write(&json_path, json)?;
        println!("JSON report saved: {}", json_path.display());

        Ok(())
    }
}

fn format_result(s: &mut String, r: &TestResult) {
    s.push_str(&format!("  Threads:         {}\n", r.threads));
    s.push_str(&format!("  Queue Depth:     {}\n", r.queue_depth));
    s.push_str(&format!("  Block Size:      {} KB\n", r.block_size_kb));
    s.push_str(&format!("  Duration:        {} seconds\n", r.duration_secs));
    s.push_str(&format!("  Throughput:    {:>10.2} MB/s\n", r.throughput_mbps));
    s.push_str(&format!("  IOPS:          {:>10.0}\n", r.iops));
    s.push_str(&format!(
        "  Avg Latency:   {:>10.2} us\n",
        r.latency_avg_us
    ));
    s.push_str(&format!(
        "  P50 Latency:   {:>10.2} us\n",
        r.latency_p50_us
    ));
    s.push_str(&format!(
        "  P99 Latency:   {:>10.2} us\n",
        r.latency_p99_us
    ));
    s.push('\n');
}
