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

var (
	MetalColor  = color.RGBA{180, 180, 200, 255}
	EnergyColor = color.RGBA{80, 180, 255, 255}
)

type ResourceDisplay struct {
	resourceType resource.Type
	label        string
	symbol       string
	barColor     color.Color
	bounds       emath.Rect
	barHeight    float64
}

func NewResourceDisplay(t resource.Type, x, y, width float64) *ResourceDisplay {
	rd := &ResourceDisplay{
		resourceType: t,
		bounds:       emath.NewRect(x, y, width, 40),
		barHeight:    8,
	}
	switch t {
	case resource.Metal:
		rd.label = "METAL"
		rd.symbol = "M"
		rd.barColor = MetalColor
	case resource.Energy:
		rd.label = "ENERGY"
		rd.symbol = "⚡"
		rd.barColor = EnergyColor
	}
	return rd
}

func (rd *ResourceDisplay) Draw(screen *ebiten.Image, res *resource.Resource) {
	rd.DrawWithDrain(screen, res, 0)
}

func (rd *ResourceDisplay) DrawWithDrain(screen *ebiten.Image, res *resource.Resource, drainPerSecond float64) {
	x := rd.bounds.Pos.X
	y := rd.bounds.Pos.Y
	w := rd.bounds.Size.X

	ebitenutil.DebugPrintAt(screen, rd.label, int(x), int(y))

	barY := y + 14
	barWidth := w - 10

	vector.FillRect(screen, float32(x), float32(barY), float32(barWidth), float32(rd.barHeight), BarBackgroundColor, false)
	fillWidth := barWidth * res.Ratio()
	vector.FillRect(screen, float32(x), float32(barY), float32(fillWidth), float32(rd.barHeight), rd.barColor, false)
	vector.StrokeRect(screen, float32(x), float32(barY), float32(barWidth), float32(rd.barHeight), 1, BorderColor, false)

	statsY := int(barY + rd.barHeight + 4)
	currentStr := fmt.Sprintf("%.0f / %.0f", res.Current, res.Capacity)
	ebitenutil.DebugPrintAt(screen, currentStr, int(x), statsY)

	netFlow := res.NetFlow() - drainPerSecond
	flowStr := ""
	if netFlow >= 0 {
		flowStr = fmt.Sprintf("+%.1f", netFlow)
	} else {
		flowStr = fmt.Sprintf("%.1f", netFlow)
	}

	var flowDetailStr string
	if drainPerSecond > 0 {
		flowDetailStr = fmt.Sprintf("(+%.1f/-%.1f/B:%.1f)", res.Production, res.Consumption, drainPerSecond)
	} else {
		flowDetailStr = fmt.Sprintf("(+%.1f/-%.1f)", res.Production, res.Consumption)
	}

	flowX := int(x + w - 90)
	ebitenutil.DebugPrintAt(screen, flowStr, flowX, statsY)
	ebitenutil.DebugPrintAt(screen, flowDetailStr, flowX-70, int(y))
}

func (rd *ResourceDisplay) Width() float64 {
	return rd.bounds.Size.X
}

type ResourceBar struct {
	panel    *Panel
	displays []*ResourceDisplay
	height   float64
}

func NewResourceBar(screenWidth float64) *ResourceBar {
	height := 45.0
	rb := &ResourceBar{
		panel:  NewPanel(0, 0, screenWidth, height),
		height: height,
	}

	displayWidth := 180.0
	startX := 10.0
	spacing := 15.0

	rb.displays = append(rb.displays, NewResourceDisplay(resource.Metal, startX, 5, displayWidth))
	rb.displays = append(rb.displays, NewResourceDisplay(resource.Energy, startX+displayWidth+spacing, 5, displayWidth))

	return rb
}

func (rb *ResourceBar) Height() float64 {
	return rb.height
}

func (rb *ResourceBar) UpdateWidth(screenWidth float64) {
	rb.panel.Bounds.Size.X = screenWidth
}

func (rb *ResourceBar) Draw(screen *ebiten.Image, resources *resource.Manager) {
	rb.panel.Draw(screen)
	const tickRate = 1.0 / 60.0
	for _, display := range rb.displays {
		res := resources.Get(display.resourceType)
		if res != nil {
			drainPerSecond := res.ConstructionDrain / tickRate
			display.DrawWithDrain(screen, res, drainPerSecond)
		}
	}
}
