package gui

import (
	"fmt"
	"image/color"
	"sync"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// fixedSizeContainer wraps a container with a fixed minimum size
type fixedSizeContainer struct {
	widget.BaseWidget
	content fyne.CanvasObject
	size    fyne.Size
}

func newFixedSizeContainer(content fyne.CanvasObject, size fyne.Size) *fixedSizeContainer {
	f := &fixedSizeContainer{
		content: content,
		size:    size,
	}
	f.ExtendBaseWidget(f)
	return f
}

func (f *fixedSizeContainer) CreateRenderer() fyne.WidgetRenderer {
	return &fixedSizeRenderer{
		container: f,
		objects:   []fyne.CanvasObject{f.content},
	}
}

func (f *fixedSizeContainer) MinSize() fyne.Size {
	return f.size
}

type fixedSizeRenderer struct {
	container *fixedSizeContainer
	objects   []fyne.CanvasObject
}

func (r *fixedSizeRenderer) Layout(size fyne.Size) {
	r.container.content.Resize(size)
	r.container.content.Move(fyne.NewPos(0, 0))
}

func (r *fixedSizeRenderer) MinSize() fyne.Size {
	return r.container.size
}

func (r *fixedSizeRenderer) Refresh() {
	canvas.Refresh(r.container.content)
}

func (r *fixedSizeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *fixedSizeRenderer) Destroy() {}

// interactiveOverlay is a transparent overlay that captures mouse events
type interactiveOverlay struct {
	widget.BaseWidget
	panel *GraphPanel
	size  fyne.Size
}

func newInteractiveOverlay(panel *GraphPanel, size fyne.Size) *interactiveOverlay {
	o := &interactiveOverlay{
		panel: panel,
		size:  size,
	}
	o.ExtendBaseWidget(o)
	return o
}

// Implement desktop.Hoverable interface
func (o *interactiveOverlay) MouseIn(*desktop.MouseEvent) {
	o.panel.mu.Lock()
	o.panel.hoverActive = true
	o.panel.mu.Unlock()
}

func (o *interactiveOverlay) MouseMoved(ev *desktop.MouseEvent) {
	o.panel.mu.Lock()
	o.panel.hoverX = ev.Position.X
	o.panel.hoverActive = true
	o.panel.mu.Unlock()
	o.panel.redrawAll()
}

func (o *interactiveOverlay) MouseOut() {
	o.panel.mu.Lock()
	o.panel.hoverActive = false
	o.panel.mu.Unlock()
	o.panel.redrawAll()
}

func (o *interactiveOverlay) CreateRenderer() fyne.WidgetRenderer {
	return &interactiveOverlayRenderer{
		overlay: o,
	}
}

type interactiveOverlayRenderer struct {
	overlay *interactiveOverlay
}

func (r *interactiveOverlayRenderer) Layout(size fyne.Size) {}
func (r *interactiveOverlayRenderer) MinSize() fyne.Size {
	return r.overlay.size
}
func (r *interactiveOverlayRenderer) Refresh() {}
func (r *interactiveOverlayRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{}
}
func (r *interactiveOverlayRenderer) Destroy() {}

type GraphPanel struct {
	container          *fyne.Container
	
	throughputContainer fyne.CanvasObject
	iopsContainer       fyne.CanvasObject
	latencyContainer    fyne.CanvasObject
	
	throughputInnerContainer *fyne.Container
	iopsInnerContainer       *fyne.Container
	latencyInnerContainer    *fyne.Container
	
	throughputData      []DataPoint
	iopsData            []DataPoint
	latencyData         []DataPoint
	
	maxDataPoints       int
	maxThroughput       float64
	maxIOPS             float64
	maxLatency          float64
	
	hoverX              float32
	hoverActive         bool
	
	mu                  sync.Mutex
}

type DataPoint struct {
	Time       float64
	ReadValue  float64
	WriteValue float64
}

func NewGraphPanel() *GraphPanel {
	gp := &GraphPanel{
		throughputData: make([]DataPoint, 0),
		iopsData:       make([]DataPoint, 0),
		latencyData:    make([]DataPoint, 0),
		maxDataPoints:  60,
		maxThroughput:  2000,    // Start at 2000 MB/s
		maxIOPS:        100000,  // Start at 100,000 IOPS
		maxLatency:     5,       // Start at 5 ms
	}
	
	// Create three separate graph containers
	throughputInner, throughputOuter := gp.createEmptyGraph("Throughput (MB/s)", 2000)
	iopsInner, iopsOuter := gp.createEmptyGraph("IOPS", 100000)
	latencyInner, latencyOuter := gp.createEmptyGraph("Latency (ms)", 5)
	
	gp.throughputInnerContainer = throughputInner
	gp.iopsInnerContainer = iopsInner
	gp.latencyInnerContainer = latencyInner
	
	gp.throughputContainer = throughputOuter
	gp.iopsContainer = iopsOuter
	gp.latencyContainer = latencyOuter
	
	// Stack them vertically with spacing
	gp.container = container.NewVBox(
		gp.throughputContainer,
		widget.NewSeparator(),
		gp.iopsContainer,
		widget.NewSeparator(),
		gp.latencyContainer,
	)
	
	// Add interactive overlay for mouse events
	overlay := newInteractiveOverlay(gp, fyne.NewSize(1100, 500))
	gp.container = container.NewStack(
		container.NewVBox(
			gp.throughputContainer,
			widget.NewSeparator(),
			gp.iopsContainer,
			widget.NewSeparator(),
			gp.latencyContainer,
		),
		overlay,
	)
	
	gp.container.Hide()
	
	return gp
}

func (gp *GraphPanel) createEmptyGraph(title string, maxValue float64) (*fyne.Container, fyne.CanvasObject) {
	objects := make([]fyne.CanvasObject, 0)
	
	// Background rectangle
	bg := canvas.NewRectangle(color.RGBA{R: 40, G: 44, B: 52, A: 255})
	bg.Resize(fyne.NewSize(1100, 150))
	bg.Move(fyne.NewPos(0, 0))
	objects = append(objects, bg)
	
	// Title
	titleText := canvas.NewText(title, color.White)
	titleText.TextSize = 14
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.Move(fyne.NewPos(10, 10))
	objects = append(objects, titleText)
	
	// Y-axis max value label
	maxLabel := canvas.NewText(fmt.Sprintf("%.0f", maxValue), color.RGBA{R: 150, G: 150, B: 150, A: 255})
	maxLabel.TextSize = 10
	maxLabel.Move(fyne.NewPos(5, 35))
	objects = append(objects, maxLabel)
	
	// Y-axis min value label
	minLabel := canvas.NewText("0", color.RGBA{R: 150, G: 150, B: 150, A: 255})
	minLabel.TextSize = 10
	minLabel.Move(fyne.NewPos(5, 135))
	objects = append(objects, minLabel)
	
	// Grid lines (5 horizontal lines)
	for i := 0; i <= 4; i++ {
		y := float32(35 + i*25)
		line := canvas.NewLine(color.RGBA{R: 60, G: 64, B: 72, A: 255})
		line.Position1 = fyne.NewPos(40, y)
		line.Position2 = fyne.NewPos(1090, y)
		line.StrokeWidth = 1
		objects = append(objects, line)
	}
	
	// Legend labels
	readLegend := canvas.NewText("Read", color.RGBA{R: 100, G: 149, B: 237, A: 255})
	readLegend.TextSize = 10
	readLegend.Move(fyne.NewPos(950, 10))
	objects = append(objects, readLegend)
	
	writeLegend := canvas.NewText("Write", color.RGBA{R: 255, G: 20, B: 147, A: 255})
	writeLegend.TextSize = 10
	writeLegend.Move(fyne.NewPos(1020, 10))
	objects = append(objects, writeLegend)
	
	c := container.NewWithoutLayout(objects...)
	c.Resize(fyne.NewSize(1100, 150))
	
	// Wrap in fixed size container to enforce dimensions
	wrapper := newFixedSizeContainer(c, fyne.NewSize(1100, 150))
	
	return c, wrapper
}

func (gp *GraphPanel) UpdateData(time, throughput, iops, latency float64, isWrite bool) {
	gp.mu.Lock()
	defer gp.mu.Unlock()
	
	// Add data points (avoid duplicates for same time)
	if len(gp.throughputData) == 0 || gp.throughputData[len(gp.throughputData)-1].Time != time {
		readTP, writeTP := 0.0, 0.0
		readIOPS, writeIOPS := 0.0, 0.0
		readLat, writeLat := 0.0, 0.0
		
		if isWrite {
			writeTP = throughput
			writeIOPS = iops
			writeLat = latency
		} else {
			readTP = throughput
			readIOPS = iops
			readLat = latency
		}
		
		gp.throughputData = append(gp.throughputData, DataPoint{time, readTP, writeTP})
		gp.iopsData = append(gp.iopsData, DataPoint{time, readIOPS, writeIOPS})
		gp.latencyData = append(gp.latencyData, DataPoint{time, readLat, writeLat})
		
		// Keep only last N points
		if len(gp.throughputData) > gp.maxDataPoints {
			gp.throughputData = gp.throughputData[1:]
			gp.iopsData = gp.iopsData[1:]
			gp.latencyData = gp.latencyData[1:]
		}
		
		// Check for scaling
		gp.checkAndScale()
		
		// Redraw all three graphs
		gp.redrawAll()
	}
}

func (gp *GraphPanel) redrawAll() {
	gp.redrawGraph(gp.throughputInnerContainer, gp.throughputData, gp.maxThroughput, "Throughput (MB/s)")
	gp.redrawGraph(gp.iopsInnerContainer, gp.iopsData, gp.maxIOPS, "IOPS")
	gp.redrawGraph(gp.latencyInnerContainer, gp.latencyData, gp.maxLatency, "Latency (ms)")
}

func (gp *GraphPanel) checkAndScale() {
	// Throughput: scale up if any value exceeds 90%, scale down if max value is below 60%
	maxThroughput := 0.0
	for _, dp := range gp.throughputData {
		if dp.ReadValue > maxThroughput {
			maxThroughput = dp.ReadValue
		}
		if dp.WriteValue > maxThroughput {
			maxThroughput = dp.WriteValue
		}
	}
	
	// Scale up by 50% if exceeding 90%
	for maxThroughput > gp.maxThroughput*0.9 {
		gp.maxThroughput *= 1.5
	}
	
	// Scale down if current max data is below 60% of scale (with minimum of 2000)
	if len(gp.throughputData) > 0 && maxThroughput < gp.maxThroughput*0.6 && gp.maxThroughput > 2000 {
		newMax := maxThroughput * 1.2 // 20% headroom above current max
		if newMax < 2000 {
			newMax = 2000
		}
		gp.maxThroughput = newMax
	}
	
	// IOPS: scale up if any value exceeds 90%, scale down if max value is below 60%
	maxIOPS := 0.0
	for _, dp := range gp.iopsData {
		if dp.ReadValue > maxIOPS {
			maxIOPS = dp.ReadValue
		}
		if dp.WriteValue > maxIOPS {
			maxIOPS = dp.WriteValue
		}
	}
	
	// Scale up by 50% if exceeding 90%
	for maxIOPS > gp.maxIOPS*0.9 {
		gp.maxIOPS *= 1.5
	}
	
	// Scale down if current max data is below 60% of scale (with minimum of 100000)
	if len(gp.iopsData) > 0 && maxIOPS < gp.maxIOPS*0.6 && gp.maxIOPS > 100000 {
		newMax := maxIOPS * 1.2 // 20% headroom above current max
		if newMax < 100000 {
			newMax = 100000
		}
		gp.maxIOPS = newMax
	}
	
	// Latency: scale up if any value exceeds 90%, scale down if max value is below 60%
	maxLatency := 0.0
	for _, dp := range gp.latencyData {
		if dp.ReadValue > maxLatency {
			maxLatency = dp.ReadValue
		}
		if dp.WriteValue > maxLatency {
			maxLatency = dp.WriteValue
		}
	}
	
	// Scale up by 50% if exceeding 90%
	for maxLatency > gp.maxLatency*0.9 {
		gp.maxLatency *= 1.5
	}
	
	// Scale down if current max data is below 60% of scale (with minimum of 5)
	if len(gp.latencyData) > 0 && maxLatency < gp.maxLatency*0.6 && gp.maxLatency > 5 {
		newMax := maxLatency * 1.2 // 20% headroom above current max
		if newMax < 5 {
			newMax = 5
		}
		gp.maxLatency = newMax
	}
}

func (gp *GraphPanel) redrawGraph(graphContainer *fyne.Container, data []DataPoint, maxValue float64, title string) {
	objects := make([]fyne.CanvasObject, 0)
	
	// Background
	bg := canvas.NewRectangle(color.RGBA{R: 40, G: 44, B: 52, A: 255})
	bg.Resize(fyne.NewSize(1100, 150))
	bg.Move(fyne.NewPos(0, 0))
	objects = append(objects, bg)
	
	// Title
	titleText := canvas.NewText(title, color.White)
	titleText.TextSize = 14
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.Move(fyne.NewPos(10, 10))
	objects = append(objects, titleText)
	
	// Y-axis labels
	maxLabel := canvas.NewText(fmt.Sprintf("%.0f", maxValue), color.RGBA{R: 150, G: 150, B: 150, A: 255})
	maxLabel.TextSize = 10
	maxLabel.Move(fyne.NewPos(5, 35))
	objects = append(objects, maxLabel)
	
	minLabel := canvas.NewText("0", color.RGBA{R: 150, G: 150, B: 150, A: 255})
	minLabel.TextSize = 10
	minLabel.Move(fyne.NewPos(5, 135))
	objects = append(objects, minLabel)
	
	// Grid lines
	for i := 0; i <= 4; i++ {
		y := float32(35 + i*25)
		line := canvas.NewLine(color.RGBA{R: 60, G: 64, B: 72, A: 255})
		line.Position1 = fyne.NewPos(40, y)
		line.Position2 = fyne.NewPos(1090, y)
		line.StrokeWidth = 1
		objects = append(objects, line)
	}
	
	// Current value labels
	if len(data) > 0 {
		last := data[len(data)-1]
		
		readValueLabel := canvas.NewText(fmt.Sprintf("Read: %.1f", last.ReadValue), color.RGBA{R: 100, G: 149, B: 237, A: 255})
		readValueLabel.TextSize = 10
		readValueLabel.Move(fyne.NewPos(250, 10))
		objects = append(objects, readValueLabel)
		
		writeValueLabel := canvas.NewText(fmt.Sprintf("Write: %.1f", last.WriteValue), color.RGBA{R: 255, G: 20, B: 147, A: 255})
		writeValueLabel.TextSize = 10
		writeValueLabel.Move(fyne.NewPos(350, 10))
		objects = append(objects, writeValueLabel)
	}
	
	// Legend
	readLegend := canvas.NewText("Read", color.RGBA{R: 100, G: 149, B: 237, A: 255})
	readLegend.TextSize = 10
	readLegend.Move(fyne.NewPos(950, 10))
	objects = append(objects, readLegend)
	
	writeLegend := canvas.NewText("Write", color.RGBA{R: 255, G: 20, B: 147, A: 255})
	writeLegend.TextSize = 10
	writeLegend.Move(fyne.NewPos(1020, 10))
	objects = append(objects, writeLegend)
	
	// Draw data lines (if we have at least 2 points)
	if len(data) >= 2 {
		graphHeight := float32(100)  // Height of plot area
		graphTop := float32(35)      // Top of plot area
		graphLeft := float32(40)     // Left edge of plot area
		graphWidth := float32(1050)  // Width of plot area (1090 - 40)
		
		dataPoints := len(data)
		
		// Blue line for Read values - RIGHT TO LEFT (newest on right)
		readColor := color.RGBA{R: 100, G: 149, B: 237, A: 255}
		for i := 0; i < len(data)-1; i++ {
			// Calculate X position from RIGHT (newest = rightmost)
			x1 := graphLeft + graphWidth - float32(dataPoints-1-i)*graphWidth/float32(gp.maxDataPoints-1)
			x2 := graphLeft + graphWidth - float32(dataPoints-2-i)*graphWidth/float32(gp.maxDataPoints-1)
			
			// Calculate Y position (inverted - 0 at bottom)
			y1 := graphTop + graphHeight - float32(data[i].ReadValue/maxValue)*graphHeight
			y2 := graphTop + graphHeight - float32(data[i+1].ReadValue/maxValue)*graphHeight
			
			// Clamp to graph area
			if y1 < graphTop {
				y1 = graphTop
			}
			if y2 < graphTop {
				y2 = graphTop
			}
			if y1 > graphTop+graphHeight {
				y1 = graphTop + graphHeight
			}
			if y2 > graphTop+graphHeight {
				y2 = graphTop + graphHeight
			}
			
			line := canvas.NewLine(readColor)
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
		
		// Pink line for Write values - RIGHT TO LEFT (newest on right)
		writeColor := color.RGBA{R: 255, G: 20, B: 147, A: 255}
		for i := 0; i < len(data)-1; i++ {
			// Calculate X position from RIGHT (newest = rightmost)
			x1 := graphLeft + graphWidth - float32(dataPoints-1-i)*graphWidth/float32(gp.maxDataPoints-1)
			x2 := graphLeft + graphWidth - float32(dataPoints-2-i)*graphWidth/float32(gp.maxDataPoints-1)
			
			// Calculate Y position (inverted - 0 at bottom)
			y1 := graphTop + graphHeight - float32(data[i].WriteValue/maxValue)*graphHeight
			y2 := graphTop + graphHeight - float32(data[i+1].WriteValue/maxValue)*graphHeight
			
			// Clamp to graph area
			if y1 < graphTop {
				y1 = graphTop
			}
			if y2 < graphTop {
				y2 = graphTop
			}
			if y1 > graphTop+graphHeight {
				y1 = graphTop + graphHeight
			}
			if y2 > graphTop+graphHeight {
				y2 = graphTop + graphHeight
			}
			
			line := canvas.NewLine(writeColor)
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}
	
	// Define constants for hover line and tooltips
	graphHeight := float32(100)
	graphTop := float32(35)
	graphLeft := float32(40)
	graphWidth := float32(1050)
	
	// Draw hover line and tooltips if hover is active
	if gp.hoverActive && gp.hoverX >= graphLeft && gp.hoverX <= graphLeft+graphWidth {
		// Vertical hover line
		hoverLine := canvas.NewLine(color.RGBA{R: 255, G: 255, B: 255, A: 200})
		hoverLine.Position1 = fyne.NewPos(gp.hoverX, graphTop)
		hoverLine.Position2 = fyne.NewPos(gp.hoverX, graphTop+graphHeight)
		hoverLine.StrokeWidth = 1
		objects = append(objects, hoverLine)
		
		// Find the data point at this X position
		if len(data) > 0 {
			relativeX := gp.hoverX - graphLeft
			dataPoints := len(data)
			
			// The plotting formula is: x = graphLeft + graphWidth - (dataPoints-1-i)*graphWidth/(maxDataPoints-1)
			// Solving for i: i = dataPoints - 1 - (graphLeft + graphWidth - hoverX) * (maxDataPoints-1) / graphWidth
			// Simplifying with relativeX = hoverX - graphLeft:
			// i = dataPoints - 1 - (graphWidth - relativeX) * (maxDataPoints-1) / graphWidth
			
			dataPointsFromRight := (graphWidth - relativeX) * float32(gp.maxDataPoints-1) / graphWidth
			dataIndex := dataPoints - 1 - int(dataPointsFromRight)
			
			// Clamp to valid range
			if dataIndex < 0 {
				dataIndex = 0
			}
			if dataIndex >= len(data) {
				dataIndex = len(data) - 1
			}
			
			if dataIndex >= 0 && dataIndex < len(data) {
				dp := data[dataIndex]
				
				// Create tooltip background
				tooltipBg := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 230})
				tooltipBg.Resize(fyne.NewSize(140, 50))
				tooltipX := gp.hoverX + 10
				if tooltipX + 140 > graphLeft + graphWidth {
					tooltipX = gp.hoverX - 150 // Show on left if too close to right edge
				}
				tooltipBg.Move(fyne.NewPos(tooltipX, graphTop+10))
				objects = append(objects, tooltipBg)
				
				// Time label
				timeLabel := canvas.NewText(fmt.Sprintf("Time: %.0fs", dp.Time), color.White)
				timeLabel.TextSize = 10
				timeLabel.Move(fyne.NewPos(tooltipX+5, graphTop+15))
				objects = append(objects, timeLabel)
				
				// Read value in blue
				readLabel := canvas.NewText(fmt.Sprintf("Read: %.1f", dp.ReadValue), color.RGBA{R: 100, G: 149, B: 237, A: 255})
				readLabel.TextSize = 10
				readLabel.Move(fyne.NewPos(tooltipX+5, graphTop+28))
				objects = append(objects, readLabel)
				
				// Write value in pink
				writeLabel := canvas.NewText(fmt.Sprintf("Write: %.1f", dp.WriteValue), color.RGBA{R: 255, G: 20, B: 147, A: 255})
				writeLabel.TextSize = 10
				writeLabel.Move(fyne.NewPos(tooltipX+5, graphTop+41))
				objects = append(objects, writeLabel)
			}
		}
	}
	
	graphContainer.Objects = objects
	graphContainer.Refresh()
}

func (gp *GraphPanel) Clear() {
	gp.mu.Lock()
	defer gp.mu.Unlock()
	
	gp.throughputData = make([]DataPoint, 0)
	gp.iopsData = make([]DataPoint, 0)
	gp.latencyData = make([]DataPoint, 0)
	
	gp.maxThroughput = 2000
	gp.maxIOPS = 100000
	gp.maxLatency = 5
	
	gp.redrawAll()
}

func (gp *GraphPanel) GetContainer() *fyne.Container {
	return gp.container
}

func (gp *GraphPanel) SetVisible(visible bool) {
	if visible {
		gp.container.Show()
	} else {
		gp.container.Hide()
	}
}
