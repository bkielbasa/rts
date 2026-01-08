package ui

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

type PauseMenuOption int

const (
	PauseOptionResume PauseMenuOption = iota
	PauseOptionSave
	PauseOptionLoad
	PauseOptionMainMenu
	PauseOptionQuit
	pauseOptionCount
)

type PauseMenu struct {
	selectedOption PauseMenuOption
	screenWidth    float64
	screenHeight   float64
	options        []string
	hoveredOption  int
}

func NewPauseMenu() *PauseMenu {
	return &PauseMenu{
		selectedOption: PauseOptionResume,
		options: []string{
			"Resume Game",
			"Save Game",
			"Load Game",
			"Main Menu",
			"Quit",
		},
		hoveredOption: -1,
	}
}

func (pm *PauseMenu) UpdateSize(w, h float64) {
	pm.screenWidth = w
	pm.screenHeight = h
}

func (pm *PauseMenu) Update(up, down, enter bool) PauseMenuOption {
	if up {
		pm.selectedOption--
		if pm.selectedOption < 0 {
			pm.selectedOption = pauseOptionCount - 1
		}
		pm.hoveredOption = -1
	}
	if down {
		pm.selectedOption++
		if pm.selectedOption >= pauseOptionCount {
			pm.selectedOption = 0
		}
		pm.hoveredOption = -1
	}
	if enter {
		return pm.selectedOption
	}
	return -1
}

func (pm *PauseMenu) UpdateHover(mousePos emath.Vec2) {
	pm.hoveredOption = -1
	_, buttonY, buttonW, buttonH := pm.getButtonMetrics()

	for i := range pm.options {
		y := buttonY + float64(i)*(buttonH+10)
		rect := emath.NewRect(pm.screenWidth/2-buttonW/2, y, buttonW, buttonH)
		if rect.Contains(mousePos) {
			pm.hoveredOption = i
			pm.selectedOption = PauseMenuOption(i)
			break
		}
	}
}

func (pm *PauseMenu) HandleClick(mousePos emath.Vec2) PauseMenuOption {
	_, buttonY, buttonW, buttonH := pm.getButtonMetrics()

	for i := range pm.options {
		y := buttonY + float64(i)*(buttonH+10)
		rect := emath.NewRect(pm.screenWidth/2-buttonW/2, y, buttonW, buttonH)
		if rect.Contains(mousePos) {
			return PauseMenuOption(i)
		}
	}
	return -1
}

func (pm *PauseMenu) getButtonMetrics() (boxY, buttonY, buttonW, buttonH float64) {
	boxWidth := 300.0
	boxHeight := 280.0
	_ = pm.screenWidth/2 - boxWidth/2
	boxY = pm.screenHeight/2 - boxHeight/2

	buttonW = 200.0
	buttonH = 35.0
	buttonY = boxY + 50

	return boxY, buttonY, buttonW, buttonH
}

func (pm *PauseMenu) Draw(screen *ebiten.Image) {
	overlayColor := color.RGBA{0, 0, 0, 180}
	vector.FillRect(screen, 0, 0, float32(pm.screenWidth), float32(pm.screenHeight), overlayColor, false)

	boxWidth := 300.0
	boxHeight := 280.0
	boxX := pm.screenWidth/2 - boxWidth/2
	boxY := pm.screenHeight/2 - boxHeight/2

	boxColor := color.RGBA{30, 30, 40, 240}
	borderColor := color.RGBA{80, 80, 100, 255}
	vector.FillRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), boxColor, false)
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), 2, borderColor, false)

	titleX := int(boxX) + int(boxWidth)/2 - 30
	titleY := int(boxY) + 15
	ebitenutil.DebugPrintAt(screen, "PAUSED", titleX, titleY)

	buttonWidth := 200.0
	buttonHeight := 35.0
	buttonX := pm.screenWidth/2 - buttonWidth/2
	startY := boxY + 50

	for i, option := range pm.options {
		y := startY + float64(i)*(buttonHeight+10)

		var bgColor color.RGBA
		if pm.hoveredOption == i || pm.selectedOption == PauseMenuOption(i) {
			bgColor = color.RGBA{60, 60, 80, 255}
		} else {
			bgColor = color.RGBA{45, 45, 60, 255}
		}

		vector.FillRect(screen, float32(buttonX), float32(y), float32(buttonWidth), float32(buttonHeight), bgColor, false)

		btnBorderColor := color.RGBA{70, 70, 90, 255}
		if pm.selectedOption == PauseMenuOption(i) {
			btnBorderColor = color.RGBA{100, 100, 140, 255}
		}
		vector.StrokeRect(screen, float32(buttonX), float32(y), float32(buttonWidth), float32(buttonHeight), 1, btnBorderColor, false)

		textX := int(buttonX) + int(buttonWidth)/2 - len(option)*3
		textY := int(y) + 12
		ebitenutil.DebugPrintAt(screen, option, textX, textY)

		if pm.selectedOption == PauseMenuOption(i) {
			markerY := float32(y) + float32(buttonHeight)/2
			vector.DrawFilledCircle(screen, float32(buttonX)-10, markerY, 4, color.RGBA{100, 200, 100, 255}, false)
		}
	}
}
