package gui

import (
	"4corners/benchmark"
	"4corners/device"
	_ "embed"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed slik_logo.ico
var iconData []byte

// Custom theme with pink accents and dark background
type customTheme struct{}

var (
	darkBackground = color.NRGBA{R: 54, G: 60, B: 74, A: 255}    // #363C4A
	pinkAccent     = color.NRGBA{R: 255, G: 20, B: 147, A: 255}  // Deep Pink
	lightGray      = color.NRGBA{R: 172, G: 206, B: 208, A: 255} // Light blue-gray
	greenButton    = color.NRGBA{R: 46, G: 204, B: 113, A: 255}  // Green
	redButton      = color.NRGBA{R: 231, G: 76, B: 60, A: 255}   // Red
)

func (m customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return darkBackground
	case theme.ColorNameButton:
		return pinkAccent
	case theme.ColorNamePrimary:
		return pinkAccent
	case theme.ColorNameForeground:
		return lightGray
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 100, G: 106, B: 120, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m customTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

type MainWindow struct {
	window           fyne.Window
	deviceLabel      *widget.Label
	readTPIOSize     *widget.Entry
	writeTPIOSize    *widget.Entry
	readIOPSIOSize   *widget.Entry
	writeIOPSIOSize  *widget.Entry
	threadsEntry     *widget.Entry
	durationEntry    *widget.Entry
	readTPThreads    *widget.Entry
	writeTPThreads   *widget.Entry
	readIOPSThreads  *widget.Entry
	writeIOPSThreads *widget.Entry
	readTPDuration   *widget.Entry
	writeTPDuration  *widget.Entry
	readIOPSDuration *widget.Entry
	writeIOPSDuration *widget.Entry
	outputText       *widget.Label
	outputScroll     *container.Scroll
	runButton        *widget.Button
	stopButton       *widget.Button
	prepButton       *widget.Button
	saveButton       *widget.Button
	showGraphCheck   *widget.Check
	graphPanel       *GraphPanel
	selectedDevice   string
	savePath         string
	benchEngine      *benchmark.Engine
	isRunning        bool
	currentTestWrite bool
}

func ShowMainWindow(a fyne.App) {
	a.Settings().SetTheme(&customTheme{})

	w := a.NewWindow("4Corners Disk Benchmark")
	w.Resize(fyne.NewSize(900, 700))
	w.SetFixedSize(false)
	
	// Set window icon
	if len(iconData) > 0 {
		icon := fyne.NewStaticResource("icon.ico", iconData)
		w.SetIcon(icon)
	}

	mw := &MainWindow{
		window:      w,
		benchEngine: benchmark.NewEngine(),
	}

	mw.buildUI()
	
	// Ensure proper cleanup on window close
	w.SetOnClosed(func() {
		a.Quit()
	})
	
	w.ShowAndRun()
}

func (mw *MainWindow) buildUI() {
	// Title
	title := canvas.NewText("4Corners Disk Benchmark", pinkAccent)
	title.TextSize = 20
	title.Alignment = fyne.TextAlignCenter

	// Create entries with default values
	mw.readTPIOSize = widget.NewEntry()
	mw.readTPIOSize.SetText("128k")
	
	mw.writeTPIOSize = widget.NewEntry()
	mw.writeTPIOSize.SetText("64k")
	
	mw.readIOPSIOSize = widget.NewEntry()
	mw.readIOPSIOSize.SetText("4k")
	
	mw.writeIOPSIOSize = widget.NewEntry()
	mw.writeIOPSIOSize.SetText("4k")

	mw.threadsEntry = widget.NewEntry()
	mw.threadsEntry.SetText("64")

	mw.durationEntry = widget.NewEntry()
	mw.durationEntry.SetText("60")

	// Create thread entries for each row (shared values)
	mw.readTPThreads = widget.NewEntry()
	mw.readTPThreads.SetText("30")
	mw.readTPThreads.OnChanged = func(s string) {
		mw.threadsEntry.SetText(s)
	}
	
	mw.writeTPThreads = widget.NewEntry()
	mw.writeTPThreads.SetText("16")
	mw.writeTPThreads.OnChanged = func(s string) {
		mw.threadsEntry.SetText(s)
	}
	
	mw.readIOPSThreads = widget.NewEntry()
	mw.readIOPSThreads.SetText("120")
	mw.readIOPSThreads.OnChanged = func(s string) {
		mw.threadsEntry.SetText(s)
	}
	
	mw.writeIOPSThreads = widget.NewEntry()
	mw.writeIOPSThreads.SetText("120")
	mw.writeIOPSThreads.OnChanged = func(s string) {
		mw.threadsEntry.SetText(s)
	}

	// Create duration entries for each row (shared values)
	mw.readTPDuration = widget.NewEntry()
	mw.readTPDuration.SetText("60")
	mw.readTPDuration.OnChanged = func(s string) {
		mw.durationEntry.SetText(s)
	}
	
	mw.writeTPDuration = widget.NewEntry()
	mw.writeTPDuration.SetText("60")
	mw.writeTPDuration.OnChanged = func(s string) {
		mw.durationEntry.SetText(s)
	}
	
	mw.readIOPSDuration = widget.NewEntry()
	mw.readIOPSDuration.SetText("60")
	mw.readIOPSDuration.OnChanged = func(s string) {
		mw.durationEntry.SetText(s)
	}
	
	mw.writeIOPSDuration = widget.NewEntry()
	mw.writeIOPSDuration.SetText("60")
	mw.writeIOPSDuration.OnChanged = func(s string) {
		mw.durationEntry.SetText(s)
	}

	// Header labels
	ioSizeLabel := widget.NewLabelWithStyle("IO Size", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	threadsLabel := widget.NewLabelWithStyle("Threads", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	durationLabel := widget.NewLabelWithStyle("Duration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Grid for test parameters - 4 columns (Label, IO Size, Threads, Duration)
	paramGrid := container.NewGridWithColumns(4,
		// Header row
		widget.NewLabel(""),
		ioSizeLabel,
		threadsLabel,
		durationLabel,
		
		// Read Throughput row
		widget.NewLabel("Read Throughput"),
		mw.readTPIOSize,
		mw.readTPThreads,
		mw.readTPDuration,
		
		// Write Throughput row
		widget.NewLabel("Write Throughput"),
		mw.writeTPIOSize,
		mw.writeTPThreads,
		mw.writeTPDuration,
		
		// Read IOPS row
		widget.NewLabel("Read IOPS"),
		mw.readIOPSIOSize,
		mw.readIOPSThreads,
		mw.readIOPSDuration,
		
		// Write IOPS row
		widget.NewLabel("Write IOPS"),
		mw.writeIOPSIOSize,
		mw.writeIOPSThreads,
		mw.writeIOPSDuration,
	)

	// Buttons - Left side 2x2 grid
	selectDeviceButton := widget.NewButton("Select Raw Device", mw.onSelectDevice)
	createFileButton := widget.NewButton("Create File Device", mw.onCreateFileDevice)
	mw.prepButton = widget.NewButton("Prep Device", mw.onPrepDevice)
	mw.saveButton = widget.NewButton("Save Report", mw.onSaveReport)
	
	// Right side Run/Stop buttons
	mw.runButton = widget.NewButton("Run", mw.onRun)
	mw.stopButton = widget.NewButton("Stop Test", mw.onStop)
	mw.stopButton.Disable()

	// Style left buttons with pink background
	selectDeviceButton.Importance = widget.HighImportance
	createFileButton.Importance = widget.HighImportance
	mw.prepButton.Importance = widget.HighImportance
	mw.saveButton.Importance = widget.HighImportance

	// Create 2x2 grid for left buttons
	leftButtonGrid := container.NewGridWithColumns(2,
		selectDeviceButton,
		createFileButton,
		mw.prepButton,
		mw.saveButton,
	)
	
	// Create custom colored buttons for Run (green) and Stop (red)
	// Make them same size
	runRect := canvas.NewRectangle(greenButton)
	runRect.SetMinSize(fyne.NewSize(100, 40))
	runButtonContainer := container.NewStack(
		runRect,
		mw.runButton,
	)
	
	stopRect := canvas.NewRectangle(redButton)
	stopRect.SetMinSize(fyne.NewSize(100, 40))
	stopButtonContainer := container.NewStack(
		stopRect,
		mw.stopButton,
	)
	
	// Right side button container
	rightButtonContainer := container.NewHBox(
		runButtonContainer,
		stopButtonContainer,
	)
	
	// Main button layout
	buttonContainer := container.NewBorder(
		nil, nil, leftButtonGrid, rightButtonContainer,
	)

	// Device display label
	mw.deviceLabel = widget.NewLabel("Selected Device: None")

	// Output area - use Label for proper vertical display
	outputLabel := widget.NewLabelWithStyle("Output", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	mw.outputText = widget.NewLabel("")
	mw.outputText.Wrapping = fyne.TextWrapWord
	mw.outputText.Alignment = fyne.TextAlignLeading
	
	mw.outputScroll = container.NewScroll(mw.outputText)
	// Allow horizontal expansion, set minimum vertical height
	mw.outputScroll.SetMinSize(fyne.NewSize(850, 300))

	outputContainer := container.NewMax(mw.outputScroll)

	// Graph panel
	mw.graphPanel = NewGraphPanel()
	
	// Show Graph checkbox - toggles between text output and graph
	mw.showGraphCheck = widget.NewCheck("Show Graph", func(checked bool) {
		if checked {
			mw.graphPanel.SetVisible(true)
			outputContainer.Hide()
		} else {
			mw.graphPanel.SetVisible(false)
			outputContainer.Show()
		}
	})

	// Main layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		paramGrid,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
		mw.deviceLabel,
		mw.showGraphCheck,
		widget.NewSeparator(),
		outputLabel,
		container.NewStack(
			outputContainer,
			mw.graphPanel.GetContainer(),
		),
	)

	mw.window.SetContent(container.NewPadded(content))
}

func (mw *MainWindow) onSelectDevice() {
	devices, err := device.ListDevices()
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to list devices: %v", err), mw.window)
		return
	}

	if len(devices) == 0 {
		dialog.ShowInformation("No Devices", "No suitable block devices found. You may need administrator/root privileges.", mw.window)
		return
	}

	deviceList := widget.NewList(
		func() int { return len(devices) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel(""))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*fyne.Container).Objects[0].(*widget.Label).SetText(devices[id].Name)
		},
	)

	var selectedIdx int
	deviceList.OnSelected = func(id widget.ListItemID) {
		selectedIdx = id
	}

	d := dialog.NewCustomConfirm("Select Block Device", "Select", "Cancel",
		container.NewBorder(
			widget.NewLabel("Available Devices:"),
			nil, nil, nil,
			deviceList,
		),
		func(confirmed bool) {
			if confirmed && selectedIdx >= 0 && selectedIdx < len(devices) {
				mw.selectedDevice = devices[selectedIdx].Path
				mw.deviceLabel.SetText(fmt.Sprintf("Selected Device: %s", devices[selectedIdx].Name))
				mw.appendOutput(fmt.Sprintf("Selected device: %s (%s)", devices[selectedIdx].Name, devices[selectedIdx].Path))
			}
		},
		mw.window,
	)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

func (mw *MainWindow) onCreateFileDevice() {
	// Create form for file device parameters
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("C:\\benchmark\\test.dat")
	
	sizeEntry := widget.NewEntry()
	sizeEntry.SetText("10")
	sizeEntry.SetPlaceHolder("10")
	
	sizeUnitSelect := widget.NewSelect([]string{"GB", "MB"}, nil)
	sizeUnitSelect.SetSelected("GB")
	
	browseButton := widget.NewButton("Browse...", func() {
		fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, mw.window)
				return
			}
			if writer != nil {
				writer.Close()
				pathEntry.SetText(writer.URI().Path())
			}
		}, mw.window)
		fd.SetFileName("benchmark.dat")
		fd.Show()
	})
	
	form := container.NewVBox(
		widget.NewLabel("Create a file to use as a benchmark device:"),
		widget.NewSeparator(),
		widget.NewLabel("File Path:"),
		container.NewBorder(nil, nil, nil, browseButton, pathEntry),
		widget.NewLabel("Size:"),
		container.NewBorder(nil, nil, nil, sizeUnitSelect, sizeEntry),
	)
	
	d := dialog.NewCustomConfirm("Create File Device", "Create", "Cancel", form,
		func(confirmed bool) {
			if !confirmed {
				return
			}
			
			filePath := pathEntry.Text
			if filePath == "" {
				dialog.ShowError(fmt.Errorf("Please specify a file path"), mw.window)
				return
			}
			
			sizeStr := sizeEntry.Text
			if sizeStr == "" {
				dialog.ShowError(fmt.Errorf("Please specify a size"), mw.window)
				return
			}
			
			size, err := strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size <= 0 {
				dialog.ShowError(fmt.Errorf("Invalid size value"), mw.window)
				return
			}
			
			// Convert to bytes based on unit
			multiplier := int64(1024 * 1024 * 1024) // GB
			if sizeUnitSelect.Selected == "MB" {
				multiplier = 1024 * 1024
			}
			sizeBytes := size * multiplier
			
			mw.appendOutput(fmt.Sprintf("\nCreating file device: %s (%d %s)", filePath, size, sizeUnitSelect.Selected))
			
			// Create file in background
			go func() {
				err := device.CreateFileDevice(filePath, sizeBytes, func(progress string) {
					mw.appendOutput(progress)
				})
				
				if err != nil {
					mw.appendOutput(fmt.Sprintf("Error: %v", err))
					return
				}
				
				mw.selectedDevice = filePath
				mw.deviceLabel.SetText(fmt.Sprintf("Selected Device: %s", filePath))
				mw.appendOutput(fmt.Sprintf("File device created and selected: %s", filePath))
			}()
		},
		mw.window,
	)
	d.Resize(fyne.NewSize(500, 250))
	d.Show()
}

func (mw *MainWindow) onPrepDevice() {
	if mw.selectedDevice == "" {
		dialog.ShowError(fmt.Errorf("No device selected"), mw.window)
		return
	}

	mw.appendOutput(fmt.Sprintf("\nPreparing device: %s", mw.selectedDevice))
	mw.appendOutput("Filling device with random data...")

	// Run prep in goroutine to avoid blocking UI
	go func() {
		err := mw.benchEngine.PrepDevice(mw.selectedDevice, func(progress string) {
			mw.appendOutput(progress)
		})
		if err != nil {
			mw.appendOutput(fmt.Sprintf("Error: %v", err))
		} else {
			mw.appendOutput("Device preparation complete!")
		}
	}()
}

func (mw *MainWindow) onSaveReport() {
	fd := dialog.NewFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, mw.window)
			return
		}
		if dir != nil {
			mw.savePath = dir.Path()
			mw.appendOutput(fmt.Sprintf("Report will be saved to: %s", mw.savePath))
		}
	}, mw.window)
	fd.Show()
}

func (mw *MainWindow) onRun() {
	if mw.selectedDevice == "" {
		dialog.ShowError(fmt.Errorf("Please select a device first"), mw.window)
		return
	}

	// Parse thread counts
	readTPThreads, err := strconv.Atoi(mw.readTPThreads.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Invalid read throughput threads value"), mw.window)
		return
	}
	
	writeTPThreads, err := strconv.Atoi(mw.writeTPThreads.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Invalid write throughput threads value"), mw.window)
		return
	}
	
	readIOPSThreads, err := strconv.Atoi(mw.readIOPSThreads.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Invalid read IOPS threads value"), mw.window)
		return
	}
	
	writeIOPSThreads, err := strconv.Atoi(mw.writeIOPSThreads.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Invalid write IOPS threads value"), mw.window)
		return
	}
	
	// Parse durations
	readTPDuration, err := strconv.Atoi(mw.readTPDuration.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Invalid read throughput duration value"), mw.window)
		return
	}
	
	writeTPDuration, err := strconv.Atoi(mw.writeTPDuration.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Invalid write throughput duration value"), mw.window)
		return
	}
	
	readIOPSDuration, err := strconv.Atoi(mw.readIOPSDuration.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Invalid read IOPS duration value"), mw.window)
		return
	}
	
	writeIOPSDuration, err := strconv.Atoi(mw.writeIOPSDuration.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Invalid write IOPS duration value"), mw.window)
		return
	}

	config := benchmark.Config{
		Device:            mw.selectedDevice,
		ReadTPIOSize:      mw.readTPIOSize.Text,
		WriteTPIOSize:     mw.writeTPIOSize.Text,
		ReadIOPSIOSize:    mw.readIOPSIOSize.Text,
		WriteIOPSIOSize:   mw.writeIOPSIOSize.Text,
		ReadTPThreads:     readTPThreads,
		WriteTPThreads:    writeTPThreads,
		ReadIOPSThreads:   readIOPSThreads,
		WriteIOPSThreads:  writeIOPSThreads,
		ReadTPDuration:    readTPDuration,
		WriteTPDuration:   writeTPDuration,
		ReadIOPSDuration:  readIOPSDuration,
		WriteIOPSDuration: writeIOPSDuration,
	}

	mw.appendOutput("\n========================================")
	mw.appendOutput("Starting 4Corners Benchmark Suite")
	mw.appendOutput("========================================")
	mw.runButton.Disable()
	mw.stopButton.Enable()
	mw.isRunning = true
	mw.currentTestWrite = false
	
	// Clear graph data
	if mw.showGraphCheck.Checked {
		mw.graphPanel.Clear()
	}

	go func() {
		defer func() {
			mw.runButton.Enable()
			mw.stopButton.Disable()
			mw.isRunning = false
		}()

		results, err := mw.benchEngine.RunBenchmark(config, func(progress string) {
			// Update text output and graph data (graph refreshes on timer)
			mw.appendOutput(progress)
			mw.updateGraphFromProgress(progress)
		})

		if err != nil {
			mw.appendOutput(fmt.Sprintf("\nError: %v", err))
			return
		}

		mw.appendOutput("\n========================================")
		mw.appendOutput("Benchmark Results:")
		mw.appendOutput("========================================")
		mw.appendOutput(fmt.Sprintf("Read Throughput:  %.2f MB/s (Avg Latency: %.2f ms)", results.ReadThroughputMBps, results.ReadTPLatencyMs))
		mw.appendOutput(fmt.Sprintf("Write Throughput: %.2f MB/s (Avg Latency: %.2f ms)", results.WriteThroughputMBps, results.WriteTPLatencyMs))
		mw.appendOutput(fmt.Sprintf("Read IOPS:        %.0f (Avg Latency: %.2f ms)", results.ReadIOPS, results.ReadIOPSLatencyMs))
		mw.appendOutput(fmt.Sprintf("Write IOPS:       %.0f (Avg Latency: %.2f ms)", results.WriteIOPS, results.WriteIOPSLatencyMs))
		mw.appendOutput("========================================")

		// Save report - use current directory if no path specified
		savePath := mw.savePath
		if savePath == "" {
			// Get current working directory
			cwd, err := os.Getwd()
			if err == nil {
				savePath = cwd
			} else {
				savePath = "."
			}
		}
		
		err = results.SaveReport(savePath)
		if err != nil {
			mw.appendOutput(fmt.Sprintf("Failed to save report: %v", err))
		} else {
			mw.appendOutput(fmt.Sprintf("Report saved to: %s", savePath))
		}
	}()
}

func (mw *MainWindow) onStop() {
	if mw.isRunning {
		mw.appendOutput("\n*** STOPPING TEST ***")
		mw.benchEngine.Stop()
		mw.stopButton.Disable()
	}
}

func (mw *MainWindow) updateGraphFromProgress(progress string) {
	if !mw.showGraphCheck.Checked {
		return
	}
	
	// Parse progress strings like "  5s: 123.45 MB/s | 987 IOPS | 1.23 ms"
	// or "Running Write Throughput test..."
	if strings.Contains(progress, "Running") && strings.Contains(progress, "test") {
		// Detect test type
		if strings.Contains(progress, "Write") {
			mw.currentTestWrite = true
		} else {
			mw.currentTestWrite = false
		}
		return
	}
	
	// Parse data line
	if strings.Contains(progress, "MB/s") && strings.Contains(progress, "IOPS") && strings.Contains(progress, "ms") {
		// Extract time, throughput, iops, latency
		var time, throughput, iops, latency float64
		_, err := fmt.Sscanf(progress, "  %fs: %f MB/s | %f IOPS | %f ms", &time, &throughput, &iops, &latency)
		if err == nil {
			mw.graphPanel.UpdateData(time, throughput, iops, latency, mw.currentTestWrite)
		}
	}
}

func (mw *MainWindow) appendOutput(text string) {
	currentText := mw.outputText.Text
	if currentText != "" {
		mw.outputText.SetText(currentText + "\n" + text)
	} else {
		mw.outputText.SetText(text)
	}
	mw.outputText.Refresh()
	// Auto-scroll to bottom
	mw.outputScroll.ScrollToBottom()
}
