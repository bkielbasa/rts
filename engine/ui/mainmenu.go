package ui

import (
	"fmt"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"strings"
)

type MenuOption int

const (
	MenuOptionStartGame MenuOption = iota
	MenuOptionExit
	MenuOptionCount
)

type MainMenu struct {
	screenWidth  float64
	screenHeight float64
	selected     MenuOption
	options      []string
}

var (
	MenuBackgroundColor = color.RGBA{20, 25, 30, 255}    // Dark blue-gray
	MenuSelectedColor   = color.RGBA{100, 200, 100, 255} // Green
)

func NewMainMenu() *MainMenu {
	return &MainMenu{
		screenWidth:  1280,
		screenHeight: 720,
		selected:     MenuOptionStartGame,
		options:      []string{"Start Game", "Exit"},
	}
}
func (m *MainMenu) UpdateSize(width, height float64) {
	m.screenWidth = width
	m.screenHeight = height
}
func (m *MainMenu) Update(upPressed, downPressed, confirmPressed bool) MenuOption {
	if upPressed {
		m.selected--
		if m.selected < 0 {
			m.selected = MenuOptionCount - 1
		}
	}
	if downPressed {
		m.selected++
		if m.selected >= MenuOptionCount {
			m.selected = 0
		}
	}
	if confirmPressed {
		return m.selected
	}
	return -1
}
func (m *MainMenu) HandleClick(pos emath.Vec2) MenuOption {
	optionHeight := 40.0
	optionWidth := 200.0
	startY := m.screenHeight/2 - 20
	for i := range m.options {
		optionY := startY + float64(i)*optionHeight
		optionX := m.screenWidth/2 - optionWidth/2
		bounds := emath.NewRect(optionX, optionY, optionWidth, optionHeight)
		if bounds.Contains(pos) {
			return MenuOption(i)
		}
	}
	return -1
}
func (m *MainMenu) UpdateHover(pos emath.Vec2) {
	optionHeight := 40.0
	optionWidth := 200.0
	startY := m.screenHeight/2 - 20
	for i := range m.options {
		optionY := startY + float64(i)*optionHeight
		optionX := m.screenWidth/2 - optionWidth/2
		bounds := emath.NewRect(optionX, optionY, optionWidth, optionHeight)
		if bounds.Contains(pos) {
			m.selected = MenuOption(i)
			return
		}
	}
}
func (m *MainMenu) Draw(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0, 0,
		float32(m.screenWidth),
		float32(m.screenHeight),
		MenuBackgroundColor,
		false,
	)
	titleLines := []string{
		"████████╗ █████╗ ███╗   ██╗██╗  ██╗███████╗",
		"╚══██╔══╝██╔══██╗████╗  ██║██║ ██╔╝██╔════╝",
		"   ██║   ███████║██╔██╗ ██║█████╔╝ ███████╗",
		"   ██║   ██╔══██║██║╚██╗██║██╔═██╗ ╚════██║",
		"   ██║   ██║  ██║██║ ╚████║██║  ██╗███████║",
		"   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝",
	}
	titleStartY := int(m.screenHeight/2) - 180
	for i, line := range titleLines {
		x := int(m.screenWidth/2) - len(line)*3 // Approximate centering
		ebitenutil.DebugPrintAt(screen, line, x, titleStartY+i*16)
	}
	subtitle := "R T S"
	subtitleX := int(m.screenWidth/2) - len(subtitle)*3
	ebitenutil.DebugPrintAt(screen, subtitle, subtitleX, titleStartY+110)
	optionHeight := 40
	startY := int(m.screenHeight/2) - 20
	for i, option := range m.options {
		optionY := startY + i*optionHeight
		var text string
		if MenuOption(i) == m.selected {
			text = fmt.Sprintf("  >>  %s  <<", strings.ToUpper(option))
			boxX := float32(m.screenWidth/2) - 100
			boxY := float32(optionY) - 5
			vector.FillRect(screen, boxX, boxY, 200, 30, color.RGBA{50, 80, 50, 200}, false)
			vector.StrokeRect(screen, boxX, boxY, 200, 30, 2, MenuSelectedColor, false)
		} else {
			text = fmt.Sprintf("      %s", option)
		}
		textX := int(m.screenWidth/2) - len(text)*3
		ebitenutil.DebugPrintAt(screen, text, textX, optionY)
	}
	instructions := "UP/DOWN: Navigate | ENTER/Click: Select | ESC: Quit"
	instructionsX := int(m.screenWidth/2) - len(instructions)*3
	ebitenutil.DebugPrintAt(screen, instructions, instructionsX, int(m.screenHeight)-50)
	version := "v0.1 - A Simple RTS Game"
	versionX := int(m.screenWidth/2) - len(version)*3
	ebitenutil.DebugPrintAt(screen, version, versionX, int(m.screenHeight)-30)
}
