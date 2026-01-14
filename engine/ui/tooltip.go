package ui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/bklimczak/tanks/engine/entity"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	tooltipPadding  = 8
	tooltipMaxWidth = 220
	lineHeight      = 14
)

type Tooltip struct {
	visible      bool
	buildingDef  *entity.BuildingDef
	unitDef      *entity.UnitDef
	position     emath.Vec2
	screenWidth  float64
	screenHeight float64
}

func NewTooltip() *Tooltip {
	return &Tooltip{}
}

func (t *Tooltip) SetScreenSize(w, h float64) {
	t.screenWidth = w
	t.screenHeight = h
}

func (t *Tooltip) ShowBuilding(def *entity.BuildingDef, x, y float64) {
	t.buildingDef = def
	t.unitDef = nil
	t.position = emath.Vec2{X: x, Y: y}
	t.visible = true
}

func (t *Tooltip) ShowUnit(def *entity.UnitDef, x, y float64) {
	t.unitDef = def
	t.buildingDef = nil
	t.position = emath.Vec2{X: x, Y: y}
	t.visible = true
}

func (t *Tooltip) Hide() {
	t.visible = false
	t.buildingDef = nil
	t.unitDef = nil
}

func (t *Tooltip) IsVisible() bool {
	return t.visible
}

func (t *Tooltip) Draw(screen *ebiten.Image) {
	if !t.visible {
		return
	}

	var lines []string
	if t.buildingDef != nil {
		lines = t.getBuildingLines()
	} else if t.unitDef != nil {
		lines = t.getUnitLines()
	} else {
		return
	}

	height := float64(len(lines)*lineHeight + tooltipPadding*2)
	width := float64(tooltipMaxWidth)

	x := t.position.X
	y := t.position.Y

	if x+width > t.screenWidth {
		x = t.screenWidth - width - 5
	}
	if y+height > t.screenHeight {
		y = t.screenHeight - height - 5
	}
	if x < 0 {
		x = 5
	}
	if y < 0 {
		y = 5
	}

	bgColor := color.RGBA{30, 30, 40, 240}
	borderColor := color.RGBA{80, 80, 100, 255}
	vector.FillRect(screen, float32(x), float32(y), float32(width), float32(height), bgColor, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 1, borderColor, false)

	textY := int(y) + tooltipPadding
	for _, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, int(x)+tooltipPadding, textY)
		textY += lineHeight
	}
}

func (t *Tooltip) getBuildingLines() []string {
	def := t.buildingDef
	lines := []string{def.Name}

	descLines := wrapText(def.Description, 30)
	lines = append(lines, descLines...)
	lines = append(lines, "")

	if def.EnergyProduction > 0 {
		lines = append(lines, fmt.Sprintf("Produces: +%.0f Energy/s", def.EnergyProduction))
	}
	if def.MetalProduction > 0 {
		lines = append(lines, fmt.Sprintf("Produces: +%.0f Metal/s", def.MetalProduction))
	}

	if def.EnergyConsumption > 0 {
		lines = append(lines, fmt.Sprintf("Consumes: -%.0f Energy/s", def.EnergyConsumption))
	}
	if def.MetalConsumption > 0 {
		lines = append(lines, fmt.Sprintf("Consumes: -%.0f Metal/s", def.MetalConsumption))
	}

	if def.EnergyStorage > 0 || def.MetalStorage > 0 {
		storageStr := "Storage:"
		if def.MetalStorage > 0 {
			storageStr += fmt.Sprintf(" +%.0f M", def.MetalStorage)
		}
		if def.EnergyStorage > 0 {
			storageStr += fmt.Sprintf(" +%.0f E", def.EnergyStorage)
		}
		lines = append(lines, storageStr)
	}

	if def.IsFactory {
		lines = append(lines, "Can produce units")
	}

	if def.CanAttack {
		lines = append(lines, fmt.Sprintf("Damage: %.0f  Range: %.0f", def.Damage, def.AttackRange))
	}

	if def.RequiresDeposit {
		lines = append(lines, "Must be placed on deposit")
	}

	lines = append(lines, "")
	costStr := "Cost:"
	if m, ok := def.Cost[resource.Metal]; ok && m > 0 {
		costStr += fmt.Sprintf(" %.0fM", m)
	}
	if e, ok := def.Cost[resource.Energy]; ok && e > 0 {
		costStr += fmt.Sprintf(" %.0fE", e)
	}
	lines = append(lines, costStr)
	lines = append(lines, fmt.Sprintf("Build Time: %.0fs", def.BuildTime))

	return lines
}

func (t *Tooltip) getUnitLines() []string {
	def := t.unitDef
	lines := []string{def.Name}

	descLines := wrapText(def.Description, 30)
	lines = append(lines, descLines...)
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("Health: %.0f  Speed: %.0f", def.Health, def.Speed))

	if def.CanAttack() {
		lines = append(lines, fmt.Sprintf("Damage: %.0f  Range: %.0f", def.GetDamage(), def.GetRange()))
	}

	if def.CanConstruct() {
		lines = append(lines, "Can construct buildings")
	}

	if def.CanRepairUnits() {
		lines = append(lines, "Can repair units")
	}

	lines = append(lines, "")
	costStr := "Cost:"
	if m, ok := def.Cost[resource.Metal]; ok && m > 0 {
		costStr += fmt.Sprintf(" %.0fM", m)
	}
	if e, ok := def.Cost[resource.Energy]; ok && e > 0 {
		costStr += fmt.Sprintf(" %.0fE", e)
	}
	lines = append(lines, costStr)
	lines = append(lines, fmt.Sprintf("Build Time: %.0fs", def.BuildTime))

	return lines
}

func wrapText(text string, maxChars int) []string {
	words := strings.Fields(text)
	var lines []string
	currentLine := ""

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= maxChars {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
