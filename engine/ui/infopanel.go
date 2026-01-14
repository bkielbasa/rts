package ui

import (
	"fmt"
	"image/color"

	"github.com/bklimczak/tanks/engine/entity"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	infoPanelWidth  = 200
	infoPanelHeight = 180
	infoPadding     = 8
	infoLineHeight  = 14
)

type InfoPanel struct {
	panel        *Panel
	visible      bool
	building     *entity.Building
	screenWidth  float64
	screenHeight float64
}

func NewInfoPanel(screenWidth, screenHeight float64) *InfoPanel {
	x := screenWidth - infoPanelWidth - 10
	y := screenHeight - infoPanelHeight - 10
	return &InfoPanel{
		panel:        NewPanel(x, y, infoPanelWidth, infoPanelHeight),
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
}

func (ip *InfoPanel) UpdatePosition(screenWidth, screenHeight float64) {
	ip.screenWidth = screenWidth
	ip.screenHeight = screenHeight
	ip.panel.Bounds.Pos.X = screenWidth - infoPanelWidth - 10
	ip.panel.Bounds.Pos.Y = screenHeight - infoPanelHeight - 10
}

func (ip *InfoPanel) SetBuilding(b *entity.Building) {
	if b == nil {
		ip.visible = false
		ip.building = nil
		return
	}
	ip.building = b
	ip.visible = true
}

func (ip *InfoPanel) IsVisible() bool {
	return ip.visible
}

func (ip *InfoPanel) Hide() {
	ip.visible = false
	ip.building = nil
}

func (ip *InfoPanel) Contains(p emath.Vec2) bool {
	if !ip.visible {
		return false
	}
	return ip.panel.Bounds.Contains(p)
}

func (ip *InfoPanel) Draw(screen *ebiten.Image) {
	if !ip.visible || ip.building == nil {
		return
	}

	ip.panel.Draw(screen)

	b := ip.building
	def := b.Def
	x := int(ip.panel.Bounds.Pos.X) + infoPadding
	y := int(ip.panel.Bounds.Pos.Y) + infoPadding

	ebitenutil.DebugPrintAt(screen, def.Name, x, y)
	y += infoLineHeight + 4

	descLines := wrapText(def.Description, 26)
	for _, line := range descLines {
		ebitenutil.DebugPrintAt(screen, line, x, y)
		y += infoLineHeight
	}
	y += 4

	healthPct := b.Health / b.MaxHealth
	barWidth := float32(infoPanelWidth - infoPadding*2)
	barHeight := float32(8)
	barX := float32(ip.panel.Bounds.Pos.X) + float32(infoPadding)
	barY := float32(y)

	vector.FillRect(screen, barX, barY, barWidth, barHeight, color.RGBA{40, 40, 40, 255}, false)
	healthColor := color.RGBA{0, 200, 0, 255}
	if healthPct < 0.3 {
		healthColor = color.RGBA{200, 0, 0, 255}
	} else if healthPct < 0.6 {
		healthColor = color.RGBA{200, 200, 0, 255}
	}
	vector.FillRect(screen, barX, barY, barWidth*float32(healthPct), barHeight, healthColor, false)
	vector.StrokeRect(screen, barX, barY, barWidth, barHeight, 1, color.RGBA{80, 80, 80, 255}, false)
	y += int(barHeight) + 4

	healthStr := fmt.Sprintf("HP: %.0f / %.0f", b.Health, b.MaxHealth)
	ebitenutil.DebugPrintAt(screen, healthStr, x, y)
	y += infoLineHeight + 4

	if def.EnergyProduction > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("+%.0f Energy/s", def.EnergyProduction), x, y)
		y += infoLineHeight
	}
	if def.MetalProduction > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("+%.0f Metal/s", def.MetalProduction), x, y)
		y += infoLineHeight
	}

	if def.EnergyConsumption > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("-%.0f Energy/s", def.EnergyConsumption), x, y)
		y += infoLineHeight
	}
	if def.MetalConsumption > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("-%.0f Metal/s", def.MetalConsumption), x, y)
		y += infoLineHeight
	}

	if def.CanAttack {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Dmg: %.0f  Range: %.0f", def.Damage, def.AttackRange), x, y)
		y += infoLineHeight
	}

	if def.IsFactory && b.Producing {
		ebitenutil.DebugPrintAt(screen, "Producing...", x, y)
	}
}

func (ip *InfoPanel) GetSelectedBuildingDef() *entity.BuildingDef {
	if ip.building == nil {
		return nil
	}
	return ip.building.Def
}

func GetBuildingStats(def *entity.BuildingDef) string {
	stats := ""
	if def.EnergyProduction > 0 {
		stats += fmt.Sprintf("+%.0fE/s ", def.EnergyProduction)
	}
	if def.MetalProduction > 0 {
		stats += fmt.Sprintf("+%.0fM/s ", def.MetalProduction)
	}
	if def.EnergyConsumption > 0 {
		stats += fmt.Sprintf("-%.0fE/s ", def.EnergyConsumption)
	}
	if def.MetalConsumption > 0 {
		stats += fmt.Sprintf("-%.0fM/s ", def.MetalConsumption)
	}
	if def.EnergyStorage > 0 || def.MetalStorage > 0 {
		stats += "Storage "
	}
	return stats
}
