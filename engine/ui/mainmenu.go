package ui

import (
	"fmt"
	"image/color"
	"strings"

	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// MenuOption represents a selectable menu item
type MenuOption int

const (
	MenuOptionStartGame MenuOption = iota
	MenuOptionExit
	MenuOptionCount
)

// MainMenu represents the game's main menu
type MainMenu struct {
	screenWidth  float64
	screenHeight float64
	selected     MenuOption
	options      []string
}

// Menu colors
var (
	MenuBackgroundColor = color.RGBA{20, 25, 30, 255} // Dark blue-gray
	MenuSelectedColor   = color.RGBA{100, 200, 100, 255} // Green
)

// NewMainMenu creates a new main menu
func NewMainMenu() *MainMenu {
	return &MainMenu{
		screenWidth:  1280,
		screenHeight: 720,
		selected:     MenuOptionStartGame,
		options:      []string{"Start Game", "Exit"},
	}
}

// UpdateSize updates the menu dimensions
func (m *MainMenu) UpdateSize(width, height float64) {
	m.screenWidth = width
	m.screenHeight = height
}

// Update handles menu input, returns the selected option if confirmed, -1 otherwise
func (m *MainMenu) Update(upPressed, downPressed, confirmPressed bool) MenuOption {
	// Navigate up
	if upPressed {
		m.selected--
		if m.selected < 0 {
			m.selected = MenuOptionCount - 1
		}
	}

	// Navigate down
	if downPressed {
		m.selected++
		if m.selected >= MenuOptionCount {
			m.selected = 0
		}
	}

	// Confirm selection
	if confirmPressed {
		return m.selected
	}

	return -1
}

// HandleClick checks if a menu option was clicked, returns the option or -1
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

// UpdateHover updates which option is hovered based on mouse position
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

// Draw renders the main menu
func (m *MainMenu) Draw(screen *ebiten.Image) {
	// Draw background
	vector.FillRect(
		screen,
		0, 0,
		float32(m.screenWidth),
		float32(m.screenHeight),
		MenuBackgroundColor,
		false,
	)

	// Draw title using ASCII art style
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

	// Draw subtitle
	subtitle := "R T S"
	subtitleX := int(m.screenWidth/2) - len(subtitle)*3
	ebitenutil.DebugPrintAt(screen, subtitle, subtitleX, titleStartY+110)

	// Draw menu options
	optionHeight := 40
	startY := int(m.screenHeight/2) - 20

	for i, option := range m.options {
		optionY := startY + i*optionHeight

		// Build the option text with indicator
		var text string
		if MenuOption(i) == m.selected {
			text = fmt.Sprintf("  >>  %s  <<", strings.ToUpper(option))
			// Draw selection highlight box
			boxX := float32(m.screenWidth/2) - 100
			boxY := float32(optionY) - 5
			vector.FillRect(screen, boxX, boxY, 200, 30, color.RGBA{50, 80, 50, 200}, false)
			vector.StrokeRect(screen, boxX, boxY, 200, 30, 2, MenuSelectedColor, false)
		} else {
			text = fmt.Sprintf("      %s", option)
		}

		// Center the text
		textX := int(m.screenWidth/2) - len(text)*3
		ebitenutil.DebugPrintAt(screen, text, textX, optionY)
	}

	// Draw instructions at bottom
	instructions := "UP/DOWN: Navigate | ENTER/Click: Select | ESC: Quit"
	instructionsX := int(m.screenWidth/2) - len(instructions)*3
	ebitenutil.DebugPrintAt(screen, instructions, instructionsX, int(m.screenHeight)-50)

	// Draw version/credits
	version := "v0.1 - A Simple RTS Game"
	versionX := int(m.screenWidth/2) - len(version)*3
	ebitenutil.DebugPrintAt(screen, version, versionX, int(m.screenHeight)-30)
}
