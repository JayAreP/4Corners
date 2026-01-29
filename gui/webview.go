package gui

import (
	"4corners/benchmark"
	"4corners/device"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	webview "github.com/webview/webview_go"
)

type GUI struct {
	w              webview.WebView
	selectedDevice string
	savePath       string
	benchEngine    *benchmark.Engine
	isRunning      bool
	currentTest    string
	logFile        *os.File
	outputBuffer   strings.Builder
}

func RunMainWindow() {
	debug := false
	w := webview.New(debug)
	defer w.Destroy()

	w.SetTitle("4Corners Disk Benchmark")
	w.SetSize(900, 700, webview.HintNone)

	// Create log file in current directory
	logFileName := fmt.Sprintf("4corners-log-%s.txt", time.Now().Format("20060102-150405"))
	logFile, err := os.Create(logFileName)
	if err != nil {
		fmt.Printf("Warning: Could not create log file: %v\n", err)
	}

	gui := &GUI{
		w:           w,
		benchEngine: benchmark.NewEngine(),
		logFile:     logFile,
	}

	if logFile != nil {
		defer logFile.Close()
		logFile.WriteString(fmt.Sprintf("4Corners Benchmark Log - %s\n", time.Now().Format("2006-01-02 15:04:05")))
		logFile.WriteString("===========================================\n\n")
	}

	// Bind Go functions to JavaScript
	w.Bind("listDevices", gui.listDevices)
	w.Bind("setDevice", gui.selectDevice)
	w.Bind("openFileDialog", gui.openFileDialog)
	w.Bind("createFileDevice", gui.createFileDevice)
	w.Bind("prepDevice", gui.prepDevice)
	w.Bind("runBenchmark", gui.runBenchmark)
	w.Bind("stopBenchmark", gui.stopBenchmark)
	w.Bind("setSaveLocation", gui.setSaveLocation)

	// Load HTML
	w.SetHtml(getHTML())
	w.Run()
}

func (g *GUI) logOutput(msg string) {
	g.outputBuffer.WriteString(msg)
	if g.logFile != nil {
		g.logFile.WriteString(msg)
		g.logFile.Sync()
	}
}

func (g *GUI) listDevices() string {
	devices, err := device.ListDevices()
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}

	data, _ := json.Marshal(devices)
	return string(data)
}

func (g *GUI) selectDevice(path string, name string) {
	g.selectedDevice = path
	g.eval(fmt.Sprintf(`updateStatus("Selected: %s");`, name))
}

func (g *GUI) openFileDialog() string {
	return openSaveFileDialog()
}

func (g *GUI) createFileDevice(path string, sizeGB int) {
	if path == "" {
		return
	}

	sizeBytes := int64(sizeGB) * 1024 * 1024 * 1024

	go func() {
		msg := "Creating file device...\n"
		g.logOutput(msg)
		g.eval(`appendOutput("Creating file device...\n");`)
		err := device.CreateFileDevice(path, sizeBytes, func(msg string) {
			g.logOutput(msg + "\n")
			escaped := strings.ReplaceAll(msg, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			g.eval(fmt.Sprintf(`appendOutput("%s\n");`, escaped))
		})

		if err != nil {
			errMsg := fmt.Sprintf("Error: %s\n", err.Error())
			g.logOutput(errMsg)
			escaped := strings.ReplaceAll(err.Error(), `"`, `\"`)
			g.eval(fmt.Sprintf(`appendOutput("Error: %s\n");`, escaped))
		} else {
			g.selectedDevice = path
			successMsg := "File device created successfully\n"
			g.logOutput(successMsg)
			g.eval(fmt.Sprintf(`updateStatus("Selected: %s"); appendOutput("File device created successfully\n");`, path))
		}
	}()
}

func (g *GUI) prepDevice() {
	if g.selectedDevice == "" {
		g.eval(`alert("Please select a device first");`)
		return
	}

	go func() {
		g.eval(`setRunning(true);`)
		msg := "Preparing device...\n"
		g.logOutput(msg)
		g.eval(`appendOutput("Preparing device...\n");`)

		err := g.benchEngine.PrepDevice(g.selectedDevice, func(msg string) {
			g.logOutput(msg + "\n")
			escaped := strings.ReplaceAll(msg, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			g.eval(fmt.Sprintf(`appendOutput("%s\n");`, escaped))
		})

		g.eval(`setRunning(false);`)
		if err != nil {
			errMsg := fmt.Sprintf("Prep failed: %s\n", err.Error())
			g.logOutput(errMsg)
			escaped := strings.ReplaceAll(err.Error(), `"`, `\"`)
			g.eval(fmt.Sprintf(`appendOutput("Prep failed: %s\n");`, escaped))
		} else {
			successMsg := "Device prep completed\n"
			g.logOutput(successMsg)
			g.eval(`appendOutput("Device prep completed\n");`)
		}
	}()
}

func (g *GUI) runBenchmark(config map[string]interface{}) {
	if g.selectedDevice == "" {
		g.eval(`alert("Please select a device first");`)
		return
	}

	benchConfig := benchmark.Config{
		Device:              g.selectedDevice,
		ReadTPIOSize:        fmt.Sprintf("%.0fk", config["readTPIOSize"]),
		WriteTPIOSize:       fmt.Sprintf("%.0fk", config["writeTPIOSize"]),
		ReadIOPSIOSize:      fmt.Sprintf("%.0fk", config["readIOPSIOSize"]),
		WriteIOPSIOSize:     fmt.Sprintf("%.0fk", config["writeIOPSIOSize"]),
		ReadTPThreads:       int(config["readTPThreads"].(float64)),
		WriteTPThreads:      int(config["writeTPThreads"].(float64)),
		ReadIOPSThreads:     int(config["readIOPSThreads"].(float64)),
		WriteIOPSThreads:    int(config["writeIOPSThreads"].(float64)),
		ReadTPDuration:      int(config["readTPDuration"].(float64)),
		WriteTPDuration:     int(config["writeTPDuration"].(float64)),
		ReadIOPSDuration:    int(config["readIOPSDuration"].(float64)),
		WriteIOPSDuration:   int(config["writeIOPSDuration"].(float64)),
		ReadTPQueueDepth:    int(config["readTPQueueDepth"].(float64)),
		WriteTPQueueDepth:   int(config["writeTPQueueDepth"].(float64)),
		ReadIOPSQueueDepth:  int(config["readIOPSQueueDepth"].(float64)),
		WriteIOPSQueueDepth: int(config["writeIOPSQueueDepth"].(float64)),
	}

	g.isRunning = true
	go func() {
		g.eval(`setRunning(true);`)
		startMsg := "\n=== Starting Benchmark ===\n"
		g.logOutput(startMsg)
		g.eval(`appendOutput("\n=== Starting Benchmark ===\n");`)
		g.eval(`clearGraphs();`)

		results, err := g.benchEngine.RunBenchmark(benchConfig, func(output string) {
			g.logOutput(output + "\n")
			escaped := strings.ReplaceAll(output, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			g.eval(fmt.Sprintf(`appendOutput("%s\n");`, escaped))

			// Parse output for graph updates
			g.parseOutputForGraphs(output)
		})

		g.isRunning = false
		g.eval(`setRunning(false);`)

		if err != nil {
			escaped := strings.ReplaceAll(err.Error(), `"`, `\"`)
			g.eval(fmt.Sprintf(`appendOutput("Benchmark failed: %s\n");`, escaped))
		} else {
			report := results.GenerateTextReport()
			escaped := strings.ReplaceAll(report, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			escaped = strings.ReplaceAll(escaped, "\n", `\n`)
			g.eval(fmt.Sprintf(`appendOutput("\n%s\n");`, escaped))

			if g.savePath != "" {
				results.SaveReport(g.savePath)
				g.eval(fmt.Sprintf(`appendOutput("Results saved to: %s\n");`, g.savePath))
			}
		}
	}()
}

func (g *GUI) stopBenchmark() {
	if g.isRunning {
		g.benchEngine.Stop()
		g.eval(`appendOutput("Stopping benchmark...\n");`)
	}
}

func (g *GUI) parseOutputForGraphs(output string) {
	// Track which test is currently running
	if strings.Contains(output, "Running Read Throughput") {
		g.currentTest = "read"
	} else if strings.Contains(output, "Running Read IOPS") {
		g.currentTest = "read"
	} else if strings.Contains(output, "Running Write Throughput") {
		g.currentTest = "write"
	} else if strings.Contains(output, "Running Write IOPS") {
		g.currentTest = "write"
	}

	// Parse real-time output for graph data
	// Format: "  25s: 1234.56 MB/s | 12345 IOPS | 1.23 ms"
	if strings.Contains(output, "MB/s") && strings.Contains(output, "IOPS") {
		parts := strings.Split(output, "|")
		if len(parts) >= 3 {
			testType := g.currentTest
			if testType == "" {
				testType = "read" // default to read if not set
			}

			// Extract throughput
			mbsPart := strings.TrimSpace(parts[0])
			if idx := strings.Index(mbsPart, "MB/s"); idx > 0 {
				if val := strings.Fields(mbsPart); len(val) >= 2 {
					g.eval(fmt.Sprintf(`updateGraph('%s', %s, 0, 0);`, testType, val[len(val)-2]))
				}
			}

			// Extract IOPS
			iopsPart := strings.TrimSpace(parts[1])
			if idx := strings.Index(iopsPart, "IOPS"); idx > 0 {
				if val := strings.Fields(iopsPart); len(val) >= 2 {
					g.eval(fmt.Sprintf(`updateGraph('%s', 0, %s, 0);`, testType, val[len(val)-2]))
				}
			}

			// Extract latency
			latPart := strings.TrimSpace(parts[2])
			if idx := strings.Index(latPart, "ms"); idx > 0 {
				if val := strings.Fields(latPart); len(val) >= 2 {
					g.eval(fmt.Sprintf(`updateGraph('%s', 0, 0, %s);`, testType, val[len(val)-2]))
				}
			}
		}
	}
}

func (g *GUI) setSaveLocation(path string) {
	g.savePath = path
	g.eval(fmt.Sprintf(`appendOutput("Save location set: %s\n");`, path))
}

func (g *GUI) eval(js string) {
	g.w.Dispatch(func() {
		g.w.Eval(js)
	})
}

func getHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
body {
	font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
	margin: 0;
	padding: 20px;
	background: #1e2836;
	color: #ecf0f1;
}
.container {
	max-width: 1400px;
	margin: 0 auto;
}
h1 {
	color: #ff5252;
	text-align: center;
	margin-bottom: 20px;
}
.status {
	background: #2d3e50;
	padding: 10px;
	border-radius: 5px;
	margin-bottom: 20px;
	font-weight: bold;
}
.section {
	background: #2d3e50;
	padding: 15px;
	border-radius: 5px;
	margin-bottom: 15px;
}
.section h2 {
	margin-top: 0;
	color: #52b8e8;
	font-size: 1.2em;
}
button {
	background: #ff1493;
	color: white;
	border: none;
	padding: 10px 20px;
	border-radius: 5px;
	cursor: pointer;
	margin-right: 10px;
	margin-bottom: 10px;
	font-size: 14px;
}
button:hover {
	background: #c2185b;
}
button:disabled {
	background: #7f8c8d;
	cursor: not-allowed;
}
.config-row {
	display: grid;
	grid-template-columns: 150px 80px 80px 80px 80px;
	gap: 10px;
	align-items: center;
	margin-bottom: 8px;
}
.config-row-label {
	color: #bdc3c7;
	font-size: 14px;
}
.config-header {
	display: grid;
	grid-template-columns: 150px 80px 80px 80px 80px;
	gap: 10px;
	margin-bottom: 8px;
	padding-bottom: 8px;
	border-bottom: 1px solid #7f8c8d;
}
.config-header-label {
	color: #52b8e8;
	font-size: 12px;
	font-weight: bold;
	text-align: center;
}
.button-row {
	display: flex;
	gap: 10px;
	margin-top: 15px;
}
label {
	margin-bottom: 5px;
	font-size: 12px;
	color: #bdc3c7;
}
input {
	padding: 8px;
	border: 1px solid #7f8c8d;
	border-radius: 3px;
	background: #1e2836;
	color: #ecf0f1;
	font-size: 14px;
	text-align: left;
}
#output {
	background: #1a252f;
	padding: 15px;
	border-radius: 5px;
	min-height: 200px;
	max-height: 250px;
	overflow-y: auto;
	font-family: 'Consolas', 'Monaco', monospace;
	font-size: 13px;
	white-space: pre-wrap;
	line-height: 1.5;
}
.warning {
	color: #e74c3c;
	font-weight: bold;
	margin: 10px 0;
}
select {
	padding: 8px;
	border: 1px solid #7f8c8d;
	border-radius: 3px;
	background: #1e2836;
	color: #ecf0f1;
	font-size: 14px;
	width: 100%;
	margin-bottom: 10px;
}
.graph-container {
	background: #1a252f;
	padding: 10px;
	border-radius: 5px;
	margin-bottom: 10px;
	position: relative;
}
.graph-title {
	color: #52b8e8;
	font-size: 14px;
	font-weight: bold;
	margin-bottom: 5px;
}
canvas {
	width: 100%;
	height: 120px;
	background: #0d1117;
	border-radius: 3px;
	cursor: crosshair;
}
.tooltip {
	position: fixed;
	background: rgba(0, 0, 0, 0.95);
	padding: 6px 10px;
	border-radius: 4px;
	font-size: 11px;
	font-family: 'Consolas', 'Monaco', monospace;
	pointer-events: none;
	display: none;
	z-index: 1000;
	border: 1px solid #7f8c8d;
	white-space: nowrap;
}
.tooltip div {
	color: #bdc3c7;
}
</style>
</head>
<body>
<div class="container">
	<h1>4Corners Disk Benchmark</h1>
	
	<div class="status" id="status">No device selected</div>
	
	<div class="section">
		<h2>Device Selection</h2>
		<button onclick="selectDevice()">Select Device</button>
		<button onclick="createFile()">Create File Device</button>
		<button onclick="prepDeviceWrapper()" id="prepBtn">Prep Device</button>
	</div>
	
	<div class="section">
		<h2>Test Configuration</h2>
		<div class="config-header">
			<div></div>
			<div class="config-header-label">IO Size</div>
			<div class="config-header-label">Threads</div>
			<div class="config-header-label">Queue Depth</div>
			<div class="config-header-label">Duration</div>
		</div>
		<div class="config-row">
			<div class="config-row-label">Read Throughput</div>
			<input type="text" id="readTPIOSize" value="128k">
			<input type="number" id="readTPThreads" value="30">
			<input type="number" id="readTPQueueDepth" value="4">
			<input type="number" id="readTPDuration" value="60">
		</div>
		<div class="config-row">
			<div class="config-row-label">Read IOPS</div>
			<input type="text" id="readIOPSIOSize" value="4k">
			<input type="number" id="readIOPSThreads" value="120">
			<input type="number" id="readIOPSQueueDepth" value="4">
			<input type="number" id="readIOPSDuration" value="60">
		</div>
		<div class="config-row">
			<div class="config-row-label">Write Throughput</div>
			<input type="text" id="writeTPIOSize" value="64k">
			<input type="number" id="writeTPThreads" value="16">
			<input type="number" id="writeTPQueueDepth" value="4">
			<input type="number" id="writeTPDuration" value="60">
		</div>
		<div class="config-row">
			<div class="config-row-label">Write IOPS</div>
			<input type="text" id="writeIOPSIOSize" value="4k">
			<input type="number" id="writeIOPSThreads" value="120">
			<input type="number" id="writeIOPSQueueDepth" value="4">
			<input type="number" id="writeIOPSDuration" value="60">
		</div>
		<div class="button-row">
			<button onclick="runBench()" id="runBtn">Run Benchmark</button>
			<button onclick="stopBench()" id="stopBtn" disabled>Stop</button>
			<span class="warning">⚠️  WARNING: Write tests will DESTROY all data on the selected device!</span>
		</div>
	</div>
	
	<div class="section">
		<h2>Real-Time Performance</h2>
		<div class="graph-container">
			<div class="graph-title">Throughput (MB/s)</div>
			<canvas id="throughputGraph"></canvas>
		</div>
		<div class="graph-container">
			<div class="graph-title">IOPS</div>
			<canvas id="iopsGraph"></canvas>
		</div>
		<div class="graph-container">
			<div class="graph-title">Latency (ms)</div>
			<canvas id="latencyGraph"></canvas>
		</div>
	</div>
	
	<div class="tooltip" id="tooltip"></div>
	
	<div class="section">
		<h2>Output</h2>
		<div id="output">4Corners Disk Benchmark - Ready
Select a device to begin.

⚠️  WARNING: Write tests will DESTROY all data on the selected device!
</div>
	</div>
</div>

<script>
const maxDataPoints = 60;
const throughputReadData = [];
const throughputWriteData = [];
const iopsReadData = [];
const iopsWriteData = [];
const latencyReadData = [];
const latencyWriteData = [];

let throughputCtx, iopsCtx, latencyCtx;
let mouseX = -1;
let showCrosshair = false;

window.onload = function() {
	throughputCtx = document.getElementById('throughputGraph').getContext('2d');
	iopsCtx = document.getElementById('iopsGraph').getContext('2d');
	latencyCtx = document.getElementById('latencyGraph').getContext('2d');
	
	// Set canvas sizes
	resizeCanvases();
	window.addEventListener('resize', resizeCanvases);
	
	// Add synchronized mouse tracking
	setupSynchronizedMouseTracking();
};

function setupSynchronizedMouseTracking() {
	const canvases = [document.getElementById('throughputGraph'), document.getElementById('iopsGraph'), document.getElementById('latencyGraph')];
	const tooltip = document.getElementById('tooltip');
	
	canvases.forEach(canvas => {
		canvas.addEventListener('mousemove', function(e) {
			const rect = canvas.getBoundingClientRect();
			const x = e.clientX - rect.left;
			mouseX = x;
			showCrosshair = true;
			
			const padding = 40;
			const graphWidth = canvas.width - padding * 2;
			const pixelsPerPoint = graphWidth / (maxDataPoints - 1);
			
			// Find closest data point index
			const maxLen = Math.max(throughputReadData.length, throughputWriteData.length, iopsReadData.length, iopsWriteData.length, latencyReadData.length, latencyWriteData.length);
			let closestIdx = -1;
			let minDist = Infinity;
			
			for (let i = 0; i < maxLen; i++) {
				const pointsFromRight = maxLen - 1 - i;
				const pointX = canvas.width - padding - (pointsFromRight * pixelsPerPoint);
				const dist = Math.abs(x - pointX);
				
				if (dist < minDist && dist < 20) {
					minDist = dist;
					closestIdx = i;
				}
			}
			
			if (closestIdx >= 0) {
				const timeAgo = maxLen - 1 - closestIdx;
				let html = '<div style="text-align:left;font-weight:bold;margin-bottom:3px;color:#fff;">Time: ' + timeAgo + 's</div>';
				
				// Throughput
				const tpRead = closestIdx < throughputReadData.length ? throughputReadData[closestIdx] : 0;
				const tpWrite = closestIdx < throughputWriteData.length ? throughputWriteData[closestIdx] : 0;
				html += '<div style="color:#bdc3c7;">Throughput: <span style="color:#52b8e8;">R: ' + tpRead.toFixed(1) + '</span> | <span style="color:#ff1493;">W: ' + tpWrite.toFixed(1) + '</span> MB/s</div>';
				
				// IOPS
				const iopsRead = closestIdx < iopsReadData.length ? iopsReadData[closestIdx] : 0;
				const iopsWrite = closestIdx < iopsWriteData.length ? iopsWriteData[closestIdx] : 0;
				html += '<div style="color:#bdc3c7;">IOPS: <span style="color:#52b8e8;">R: ' + iopsRead.toFixed(0) + '</span> | <span style="color:#ff1493;">W: ' + iopsWrite.toFixed(0) + '</span></div>';
				
				// Latency
				const latRead = closestIdx < latencyReadData.length ? latencyReadData[closestIdx] : 0;
				const latWrite = closestIdx < latencyWriteData.length ? latencyWriteData[closestIdx] : 0;
				html += '<div style="color:#bdc3c7;">Latency: <span style="color:#52b8e8;">R: ' + latRead.toFixed(2) + '</span> | <span style="color:#ff1493;">W: ' + latWrite.toFixed(2) + '</span> ms</div>';
				
				tooltip.innerHTML = html;
				tooltip.style.display = 'block';
				tooltip.style.left = (e.clientX + 15) + 'px';
				tooltip.style.top = (e.clientY - 30) + 'px';
			} else {
				showCrosshair = false;
				tooltip.style.display = 'none';
			}
			
			drawGraphs();
		});
		
		canvas.addEventListener('mouseleave', function() {
			showCrosshair = false;
			tooltip.style.display = 'none';
			drawGraphs();
		});
	});
}

function resizeCanvases() {
	const canvases = document.querySelectorAll('canvas');
	canvases.forEach(canvas => {
		const rect = canvas.getBoundingClientRect();
		canvas.width = rect.width;
		canvas.height = 120;
	});
	drawGraphs();
}

function clearGraphs() {
	throughputReadData.length = 0;
	throughputWriteData.length = 0;
	iopsReadData.length = 0;
	iopsWriteData.length = 0;
	latencyReadData.length = 0;
	latencyWriteData.length = 0;
	drawGraphs();
}

function updateGraph(testType, throughput, iops, latency) {
	const isRead = testType === 'read';
	
	// ALWAYS update BOTH read and write arrays so both lines scroll continuously
	if (throughput > 0) {
		if (isRead) {
			throughputReadData.push(throughput);
			throughputWriteData.push(0);  // Write line shows 0 during read test
		} else {
			throughputReadData.push(0);    // Read line shows 0 during write test
			throughputWriteData.push(throughput);
		}
		// Keep arrays in sync
		if (throughputReadData.length > maxDataPoints) throughputReadData.shift();
		if (throughputWriteData.length > maxDataPoints) throughputWriteData.shift();
	}
	
	if (iops > 0) {
		if (isRead) {
			iopsReadData.push(iops);
			iopsWriteData.push(0);
		} else {
			iopsReadData.push(0);
			iopsWriteData.push(iops);
		}
		if (iopsReadData.length > maxDataPoints) iopsReadData.shift();
		if (iopsWriteData.length > maxDataPoints) iopsWriteData.shift();
	}
	
	if (latency > 0) {
		if (isRead) {
			latencyReadData.push(latency);
			latencyWriteData.push(0);
		} else {
			latencyReadData.push(0);
			latencyWriteData.push(latency);
		}
		if (latencyReadData.length > maxDataPoints) latencyReadData.shift();
		if (latencyWriteData.length > maxDataPoints) latencyWriteData.shift();
	}
	
	drawGraphs();
}

function drawGraphs() {
	drawDualGraph(throughputCtx, throughputReadData, throughputWriteData, 'MB/s');
	drawDualGraph(iopsCtx, iopsReadData, iopsWriteData, 'IOPS');
	drawDualGraph(latencyCtx, latencyReadData, latencyWriteData, 'ms');
}

function drawDualGraph(ctx, readData, writeData, unit) {
	const canvas = ctx.canvas;
	const width = canvas.width;
	const height = canvas.height;
	const padding = 40;
	const rightPadding = 80;
	const graphWidth = width - padding - rightPadding;
	const graphHeight = height - 30;
	
	// Clear canvas
	ctx.clearRect(0, 0, width, height);
	
	// Calculate scale - combine both datasets
	const allValues = [...readData, ...writeData].filter(v => v > 0);
	const maxVal = allValues.length > 0 ? Math.max(...allValues) : 10;
	const minVal = 0;
	const range = maxVal - minVal;
	
	// Draw background grid
	ctx.strokeStyle = 'rgba(127, 140, 141, 0.2)';
	ctx.lineWidth = 1;
	for (let i = 0; i <= 4; i++) {
		const y = 10 + (graphHeight / 4) * i;
		ctx.beginPath();
		ctx.moveTo(padding, y);
		ctx.lineTo(padding + graphWidth, y);
		ctx.stroke();
	}
	
	// Draw Y-axis scale labels on right
	ctx.fillStyle = '#bdc3c7';
	ctx.font = '10px Arial';
	ctx.textAlign = 'left';
	for (let i = 0; i <= 4; i++) {
		const value = maxVal - (maxVal / 4) * i;
		const y = 10 + (graphHeight / 4) * i;
		ctx.fillText(value.toFixed(0), width - rightPadding + 5, y + 4);
	}
	
	// Draw unit label on right
	ctx.fillStyle = '#7f8c8d';
	ctx.font = 'bold 10px Arial';
	ctx.fillText(unit, width - rightPadding + 5, height / 2);
	
	// Draw legend in top-right
	ctx.font = 'bold 11px Arial';
	ctx.fillStyle = '#52b8e8';
	ctx.textAlign = 'right';
	ctx.fillText('Read', width - 10, 15);
	ctx.fillStyle = '#ff1493';
	ctx.fillText('Write', width - 10, 30);
	
	// Draw crosshair if hovering
	if (showCrosshair && mouseX >= padding && mouseX <= padding + graphWidth) {
		ctx.strokeStyle = 'rgba(255, 255, 255, 0.5)';
		ctx.lineWidth = 1;
		ctx.setLineDash([3, 3]);
		ctx.beginPath();
		ctx.moveTo(mouseX, 10);
		ctx.lineTo(mouseX, 10 + graphHeight);
		ctx.stroke();
		ctx.setLineDash([]);
	}
	
	// BLUE LINE FOR READS ONLY
	ctx.strokeStyle = '#52b8e8';
	ctx.lineWidth = 2;
	ctx.beginPath();
	
	if (readData.length === 0) {
		// No read data - flat blue line at bottom
		const y = 10 + graphHeight;
		ctx.moveTo(padding, y);
		ctx.lineTo(padding + graphWidth, y);
	} else {
		// Draw read data in blue
		const pointSpacing = graphWidth / (maxDataPoints - 1);
		for (let i = 0; i < readData.length; i++) {
			const dataIndex = readData.length - 1 - i;
			const x = padding + graphWidth - (dataIndex * pointSpacing);
			const normalizedValue = (readData[i] - minVal) / (range || 1);
			const y = 10 + graphHeight - (normalizedValue * graphHeight);
			
			if (i === 0) {
				ctx.moveTo(x, y);
			} else {
				ctx.lineTo(x, y);
			}
		}
	}
	ctx.stroke();
	
	// PINK LINE FOR WRITES ONLY
	ctx.strokeStyle = '#ff1493';
	ctx.lineWidth = 2;
	ctx.beginPath();
	
	if (writeData.length === 0) {
		// No write data - flat pink line at bottom
		const y = 10 + graphHeight;
		ctx.moveTo(padding, y);
		ctx.lineTo(padding + graphWidth, y);
	} else {
		// Draw write data in pink
		const pointSpacing = graphWidth / (maxDataPoints - 1);
		for (let i = 0; i < writeData.length; i++) {
			const dataIndex = writeData.length - 1 - i;
			const x = padding + graphWidth - (dataIndex * pointSpacing);
			const normalizedValue = (writeData[i] - minVal) / (range || 1);
			const y = 10 + graphHeight - (normalizedValue * graphHeight);
			
			if (i === 0) {
				ctx.moveTo(x, y);
			} else {
				ctx.lineTo(x, y);
			}
		}
	}
	ctx.stroke();
	
	// Draw current values in top-left
	ctx.font = 'bold 11px Arial';
	ctx.textAlign = 'left';
	
	const readVal = readData.length > 0 ? readData[readData.length - 1] : 0;
	const writeVal = writeData.length > 0 ? writeData[writeData.length - 1] : 0;
	
	ctx.fillStyle = '#52b8e8';
	ctx.fillText('R: ' + readVal.toFixed(1), padding + 5, 15);
	
	ctx.fillStyle = '#ff1493';
	ctx.fillText('W: ' + writeVal.toFixed(1), padding + 5, 30);
}

function updateStatus(msg) {
	document.getElementById('status').textContent = msg;
}

function appendOutput(text) {
	const output = document.getElementById('output');
	output.textContent += text;
	output.scrollTop = output.scrollHeight;
}

function setRunning(running) {
	document.getElementById('runBtn').disabled = running;
	document.getElementById('stopBtn').disabled = !running;
	document.getElementById('prepBtn').disabled = running;
}

async function selectDevice() {
	const devicesJson = await listDevices();
	const data = JSON.parse(devicesJson);
	
	if (data.error) {
		alert('Error: ' + data.error);
		return;
	}
	
	if (data.length === 0) {
		alert('No devices found');
		return;
	}
	
	let html = '<select id="deviceSelect" size="10" style="min-height:200px;">';
	data.forEach(d => {
		html += '<option value="' + d.Path + '">' + d.Name + '</option>';
	});
	html += '</select>';
	
	const div = document.createElement('div');
	div.innerHTML = '<div style="background:#34495e;padding:20px;border-radius:5px;position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:1000;min-width:400px;"><h3 style="color:#e91e63;margin-top:0;">Select Device</h3><p style="color:#e74c3c;">WARNING: Write tests will destroy all data!</p>' + html + '<br><br><button onclick="confirmDevice()">OK</button><button onclick="closeDialog()">Cancel</button></div><div onclick="closeDialog()" style="position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.7);z-index:999;"></div>';
	div.id = 'deviceDialog';
	document.body.appendChild(div);
}

function confirmDevice() {
	const sel = document.getElementById('deviceSelect');
	if (sel && sel.value) {
		setDevice(sel.value, sel.options[sel.selectedIndex].text);
		closeDialog();
	}
}

function closeDialog() {
	const dlg = document.getElementById('deviceDialog');
	if (dlg) dlg.remove();
	const fileDlg = document.getElementById('fileDialog');
	if (fileDlg) fileDlg.remove();
}

function createFile() {
	const div = document.createElement('div');
	div.innerHTML = '<div style="background:#34495e;padding:20px;border-radius:5px;position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:1000;min-width:500px;"><h3 style="color:#e91e63;margin-top:0;">Create File Device</h3>' +
		'<label>File Path:</label><br>' +
		'<div style="display:flex;gap:10px;margin:10px 0;"><input type="text" id="filePath" value="" placeholder="Click Browse to select..." style="flex:1;font-family:monospace;" readonly>' +
		'<button onclick="browsePath()" style="margin:0;">Browse...</button></div>' +
		'<label>Size (GB):</label><br>' +
		'<input type="number" id="fileSize" value="10" min="1" max="10000" style="width:95%;margin:10px 0;"><br><br>' +
		'<button onclick="confirmCreate()">Create</button>' +
		'<button onclick="closeDialog()">Cancel</button></div>' +
		'<div onclick="closeDialog()" style="position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.7);z-index:999;"></div>';
	div.id = 'fileDialog';
	document.body.appendChild(div);
}

async function browsePath() {
	const path = await openFileDialog();
	if (path) {
		document.getElementById('filePath').value = path;
	}
}

function confirmCreate() {
	const path = document.getElementById('filePath').value.trim();
	const size = parseInt(document.getElementById('fileSize').value);
	if (!path) {
		alert('Please select a file location using the Browse button');
		return;
	}
	if (!size || size < 1) {
		alert('Please enter a valid size (minimum 1 GB)');
		return;
	}
	createFileDevice(path, size);
	closeDialog();
}

function prepDeviceWrapper() {
	if (!confirm('This will fill the entire device with random data.\nThis may take a long time. Continue?')) {
		return;
	}
	prepDevice();
}

function runBench() {
	if (!confirm('⚠️  WARNING: Write tests will DESTROY all data on the device!\n\nContinue?')) {
		return;
	}
	
	function parseIOSize(val) {
		const str = val.toLowerCase().replace(/\s/g, '');
		const num = parseFloat(str);
		return num;
	}
	
	const config = {
		readTPIOSize: parseIOSize(document.getElementById('readTPIOSize').value),
		readTPThreads: parseFloat(document.getElementById('readTPThreads').value),
		readTPQueueDepth: parseFloat(document.getElementById('readTPQueueDepth').value),
		writeTPIOSize: parseIOSize(document.getElementById('writeTPIOSize').value),
		writeTPThreads: parseFloat(document.getElementById('writeTPThreads').value),
		writeTPQueueDepth: parseFloat(document.getElementById('writeTPQueueDepth').value),
		readIOPSIOSize: parseIOSize(document.getElementById('readIOPSIOSize').value),
		readIOPSThreads: parseFloat(document.getElementById('readIOPSThreads').value),
		readIOPSQueueDepth: parseFloat(document.getElementById('readIOPSQueueDepth').value),
		writeIOPSIOSize: parseIOSize(document.getElementById('writeIOPSIOSize').value),
		writeIOPSThreads: parseFloat(document.getElementById('writeIOPSThreads').value),
		writeIOPSQueueDepth: parseFloat(document.getElementById('writeIOPSQueueDepth').value),
		readTPDuration: parseFloat(document.getElementById('readTPDuration').value),
		writeTPDuration: parseFloat(document.getElementById('writeTPDuration').value),
		readIOPSDuration: parseFloat(document.getElementById('readIOPSDuration').value),
		writeIOPSDuration: parseFloat(document.getElementById('writeIOPSDuration').value)
	};
	
	runBenchmark(config);
}

function stopBench() {
	stopBenchmark();
}
</script>
</body>
</html>`
}
