package ui

import (
	"fmt"
	"image/color"

	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Resource display colors (Total Annihilation style)
var (
	MetalColor  = color.RGBA{100, 100, 120, 255} // Grayish blue
	EnergyColor = color.RGBA{255, 220, 50, 255}  // Yellow/gold
)

// ResourceDisplay shows a single resource with bar and stats
type ResourceDisplay struct {
	resourceType resource.Type
	label        string
	barColor     color.Color
	bounds       emath.Rect
	barHeight    float64
}

// NewResourceDisplay creates a display for a resource type
func NewResourceDisplay(t resource.Type, x, y, width float64) *ResourceDisplay {
	rd := &ResourceDisplay{
		resourceType: t,
		bounds:       emath.NewRect(x, y, width, 40),
		barHeight:    8,
	}

	switch t {
	case resource.Metal:
		rd.label = "METAL"
		rd.barColor = MetalColor
	case resource.Energy:
		rd.label = "ENERGY"
		rd.barColor = EnergyColor
	}

	return rd
}

// Draw renders the resource display
func (rd *ResourceDisplay) Draw(screen *ebiten.Image, res *resource.Resource) {
	x := rd.bounds.Pos.X
	y := rd.bounds.Pos.Y
	w := rd.bounds.Size.X

	// Label
	ebitenutil.DebugPrintAt(screen, rd.label, int(x), int(y))

	// Progress bar
	barY := y + 14
	barWidth := w - 10

	// Bar background
	vector.FillRect(screen, float32(x), float32(barY), float32(barWidth), float32(rd.barHeight), BarBackgroundColor, false)

	// Bar fill
	fillWidth := barWidth * res.Ratio()
	vector.FillRect(screen, float32(x), float32(barY), float32(fillWidth), float32(rd.barHeight), rd.barColor, false)

	// Bar border
	vector.StrokeRect(screen, float32(x), float32(barY), float32(barWidth), float32(rd.barHeight), 1, BorderColor, false)

	// Stats: Current / Capacity
	statsY := int(barY + rd.barHeight + 4)
	currentStr := fmt.Sprintf("%.0f / %.0f", res.Current, res.Capacity)
	ebitenutil.DebugPrintAt(screen, currentStr, int(x), statsY)

	// Flow rate (production - consumption)
	netFlow := res.NetFlow()
	flowStr := ""
	if netFlow >= 0 {
		flowStr = fmt.Sprintf("+%.1f", netFlow)
	} else {
		flowStr = fmt.Sprintf("%.1f", netFlow)
	}

	// Production/Consumption detail
	flowDetailStr := fmt.Sprintf("(+%.1f / -%.1f)", res.Production, res.Consumption)

	// Position flow info on the right side
	flowX := int(x + w - 100)
	ebitenutil.DebugPrintAt(screen, flowStr, flowX, statsY)
	ebitenutil.DebugPrintAt(screen, flowDetailStr, flowX-60, int(y))
}

// Width returns the width of this display
func (rd *ResourceDisplay) Width() float64 {
	return rd.bounds.Size.X
}

// ResourceBar is the top bar showing all resources (TA style)
type ResourceBar struct {
	panel    *Panel
	displays []*ResourceDisplay
	height   float64
}

// NewResourceBar creates a new resource bar
func NewResourceBar(screenWidth float64) *ResourceBar {
	height := 45.0
	rb := &ResourceBar{
		panel:  NewPanel(0, 0, screenWidth, height),
		height: height,
	}

	// Create displays for each resource type
	displayWidth := 200.0
	startX := 10.0
	spacing := 20.0

	rb.displays = append(rb.displays, NewResourceDisplay(resource.Metal, startX, 5, displayWidth))
	rb.displays = append(rb.displays, NewResourceDisplay(resource.Energy, startX+displayWidth+spacing, 5, displayWidth))

	return rb
}

// Height returns the height of the resource bar
func (rb *ResourceBar) Height() float64 {
	return rb.height
}

// UpdateWidth updates the bar width when screen resizes
func (rb *ResourceBar) UpdateWidth(screenWidth float64) {
	rb.panel.Bounds.Size.X = screenWidth
}

// Draw renders the resource bar
func (rb *ResourceBar) Draw(screen *ebiten.Image, resources *resource.Manager) {
	// Draw panel background
	rb.panel.Draw(screen)

	// Draw each resource display
	for _, display := range rb.displays {
		res := resources.Get(display.resourceType)
		if res != nil {
			display.Draw(screen, res)
		}
	}
}
