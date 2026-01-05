package ui

import (
	"fmt"
	"image/color"

	"github.com/bklimczak/tanks/engine/entity"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	commandPanelWidth  = 180
	buttonHeight       = 50
	buttonMargin       = 5
	panelPadding       = 10
)

// ButtonState represents the state of a button
type ButtonState int

const (
	ButtonNormal ButtonState = iota
	ButtonHovered
	ButtonPressed
	ButtonDisabled
)

// CommandButton represents a clickable command button for buildings
type CommandButton struct {
	Bounds      emath.Rect
	Label       string
	Description string
	BuildingDef *entity.BuildingDef
	State       ButtonState
	OnClick     func()
}

// NewCommandButton creates a new command button
func NewCommandButton(x, y, w, h float64, def *entity.BuildingDef) *CommandButton {
	return &CommandButton{
		Bounds:      emath.NewRect(x, y, w, h),
		Label:       def.Name,
		Description: def.Description,
		BuildingDef: def,
		State:       ButtonNormal,
	}
}

// UnitButton represents a clickable button for unit production
type UnitButton struct {
	Bounds     emath.Rect
	Label      string
	UnitDef    *entity.UnitDef
	State      ButtonState
	QueueCount int
}

// NewUnitButton creates a new unit production button
func NewUnitButton(x, y, w, h float64, def *entity.UnitDef) *UnitButton {
	return &UnitButton{
		Bounds:  emath.NewRect(x, y, w, h),
		Label:   def.Name,
		UnitDef: def,
		State:   ButtonNormal,
	}
}

// Contains checks if a point is inside the unit button
func (b *UnitButton) Contains(p emath.Vec2) bool {
	return b.Bounds.Contains(p)
}

// Draw renders the unit button
func (b *UnitButton) Draw(screen *ebiten.Image, resources *resource.Manager) {
	x := float32(b.Bounds.Pos.X)
	y := float32(b.Bounds.Pos.Y)
	w := float32(b.Bounds.Size.X)
	h := float32(b.Bounds.Size.Y)

	// Determine colors based on state
	var bgColor, borderColor color.Color
	canAfford := resources.CanAfford(b.UnitDef.Cost)

	switch b.State {
	case ButtonHovered:
		if canAfford {
			bgColor = color.RGBA{60, 60, 80, 255}
		} else {
			bgColor = color.RGBA{80, 40, 40, 255}
		}
		borderColor = color.RGBA{100, 100, 120, 255}
	case ButtonPressed:
		bgColor = color.RGBA{40, 40, 60, 255}
		borderColor = color.RGBA{120, 120, 140, 255}
	case ButtonDisabled:
		bgColor = color.RGBA{30, 30, 35, 255}
		borderColor = color.RGBA{50, 50, 60, 255}
	default:
		if canAfford {
			bgColor = color.RGBA{45, 45, 60, 255}
		} else {
			bgColor = color.RGBA{60, 35, 35, 255}
		}
		borderColor = color.RGBA{70, 70, 90, 255}
	}

	// Draw background
	vector.FillRect(screen, x, y, w, h, bgColor, false)

	// Draw border
	vector.StrokeRect(screen, x, y, w, h, 1, borderColor, false)

	// Draw unit color preview
	previewSize := float32(16)
	previewX := x + 8
	previewY := y + 8
	vector.FillRect(screen, previewX, previewY, previewSize, previewSize, b.UnitDef.Color, false)
	vector.StrokeRect(screen, previewX, previewY, previewSize, previewSize, 1, color.RGBA{100, 100, 100, 255}, false)

	// Draw label
	labelX := int(previewX + previewSize + 8)
	labelY := int(y + 6)
	ebitenutil.DebugPrintAt(screen, b.Label, labelX, labelY)

	// Draw cost
	costY := int(y + 22)
	costStr := ""
	if metalCost, ok := b.UnitDef.Cost[resource.Metal]; ok && metalCost > 0 {
		costStr += fmt.Sprintf("M:%.0f ", metalCost)
	}
	if energyCost, ok := b.UnitDef.Cost[resource.Energy]; ok && energyCost > 0 {
		costStr += fmt.Sprintf("E:%.0f", energyCost)
	}
	ebitenutil.DebugPrintAt(screen, costStr, labelX, costY)

	// Draw build time
	timeStr := fmt.Sprintf("%.0fs", b.UnitDef.BuildTime)
	ebitenutil.DebugPrintAt(screen, timeStr, int(x+w-35), costY)

	// Draw queue count in bottom-right corner if > 0
	if b.QueueCount > 0 {
		countStr := fmt.Sprintf("%d", b.QueueCount)
		countX := int(x + w - 20)
		countY := int(y + h - 16)
		// Draw count background
		vector.FillRect(screen, float32(countX-4), float32(countY-2), 20, 14, color.RGBA{0, 100, 0, 200}, false)
		ebitenutil.DebugPrintAt(screen, countStr, countX, countY)
	}
}

// Contains checks if a point is inside the button
func (b *CommandButton) Contains(p emath.Vec2) bool {
	return b.Bounds.Contains(p)
}

// Draw renders the button
func (b *CommandButton) Draw(screen *ebiten.Image, resources *resource.Manager) {
	x := float32(b.Bounds.Pos.X)
	y := float32(b.Bounds.Pos.Y)
	w := float32(b.Bounds.Size.X)
	h := float32(b.Bounds.Size.Y)

	// Determine colors based on state
	var bgColor, borderColor color.Color
	canAfford := resources.CanAfford(b.BuildingDef.Cost)

	switch b.State {
	case ButtonHovered:
		if canAfford {
			bgColor = color.RGBA{60, 60, 80, 255}
		} else {
			bgColor = color.RGBA{80, 40, 40, 255}
		}
		borderColor = color.RGBA{100, 100, 120, 255}
	case ButtonPressed:
		bgColor = color.RGBA{40, 40, 60, 255}
		borderColor = color.RGBA{120, 120, 140, 255}
	case ButtonDisabled:
		bgColor = color.RGBA{30, 30, 35, 255}
		borderColor = color.RGBA{50, 50, 60, 255}
	default:
		if canAfford {
			bgColor = color.RGBA{45, 45, 60, 255}
		} else {
			bgColor = color.RGBA{60, 35, 35, 255}
		}
		borderColor = color.RGBA{70, 70, 90, 255}
	}

	// Draw background
	vector.FillRect(screen, x, y, w, h, bgColor, false)

	// Draw border
	vector.StrokeRect(screen, x, y, w, h, 1, borderColor, false)

	// Draw building color preview
	previewSize := float32(16)
	previewX := x + 8
	previewY := y + 8
	vector.FillRect(screen, previewX, previewY, previewSize, previewSize, b.BuildingDef.Color, false)
	vector.StrokeRect(screen, previewX, previewY, previewSize, previewSize, 1, color.RGBA{100, 100, 100, 255}, false)

	// Draw label
	labelX := int(previewX + previewSize + 8)
	labelY := int(y + 6)
	ebitenutil.DebugPrintAt(screen, b.Label, labelX, labelY)

	// Draw cost
	costY := int(y + 22)
	costStr := ""
	if metalCost, ok := b.BuildingDef.Cost[resource.Metal]; ok && metalCost > 0 {
		costStr += fmt.Sprintf("M:%.0f ", metalCost)
	}
	if energyCost, ok := b.BuildingDef.Cost[resource.Energy]; ok && energyCost > 0 {
		costStr += fmt.Sprintf("E:%.0f", energyCost)
	}

	// Color cost text based on affordability
	if canAfford {
		ebitenutil.DebugPrintAt(screen, costStr, labelX, costY)
	} else {
		// Draw in red-ish (we can't easily change color with DebugPrint, so we just draw it)
		ebitenutil.DebugPrintAt(screen, costStr, labelX, costY)
	}

	// Draw build time
	timeStr := fmt.Sprintf("%.0fs", b.BuildingDef.BuildTime)
	ebitenutil.DebugPrintAt(screen, timeStr, int(x+w-35), costY)
}

// CommandPanel shows commands/build options for selected units
type CommandPanel struct {
	panel           *Panel
	buttons         []*CommandButton
	unitButtons     []*UnitButton
	visible         bool
	topOffset       float64 // Offset from top (for resource bar)
	title           string
	selectedFactory *entity.Building // Reference to selected factory for queue updates
}

// NewCommandPanel creates a new command panel
func NewCommandPanel(topOffset float64, screenHeight float64) *CommandPanel {
	return &CommandPanel{
		panel: NewPanel(0, topOffset, commandPanelWidth, screenHeight-topOffset),
		topOffset: topOffset,
		visible:   false,
	}
}

// UpdateHeight updates the panel height when screen resizes
func (cp *CommandPanel) UpdateHeight(screenHeight float64) {
	cp.panel.Bounds.Size.Y = screenHeight - cp.topOffset
}

// Width returns the panel width
func (cp *CommandPanel) Width() float64 {
	return commandPanelWidth
}

// IsVisible returns whether the panel is visible
func (cp *CommandPanel) IsVisible() bool {
	return cp.visible
}

// SetVisible sets panel visibility
func (cp *CommandPanel) SetVisible(visible bool) {
	cp.visible = visible
	if !visible {
		cp.buttons = nil
		cp.unitButtons = nil
		cp.title = ""
		cp.selectedFactory = nil
	}
}

// SetBuildOptions populates the panel with building options for constructor
func (cp *CommandPanel) SetBuildOptions(units []*entity.Unit) {
	cp.buttons = nil
	cp.unitButtons = nil
	cp.title = ""
	cp.visible = false
	cp.selectedFactory = nil

	// Check if we have exactly one constructor selected
	var constructor *entity.Unit
	selectedCount := 0
	for _, u := range units {
		if u.Selected {
			selectedCount++
			if u.CanBuild() {
				constructor = u
			}
		}
	}

	// Only show build panel for single constructor selection
	if constructor == nil || selectedCount != 1 {
		return
	}

	cp.visible = true
	cp.title = "BUILD"

	// Create buttons for each building option
	options := constructor.GetBuildOptions()
	y := cp.topOffset + panelPadding + 20 // +20 for title

	for _, def := range options {
		btn := NewCommandButton(
			panelPadding,
			y,
			commandPanelWidth-panelPadding*2,
			buttonHeight,
			def,
		)
		cp.buttons = append(cp.buttons, btn)
		y += buttonHeight + buttonMargin
	}
}

// SetFactoryOptions populates the panel with unit production options for factory
func (cp *CommandPanel) SetFactoryOptions(factory *entity.Building) {
	cp.buttons = nil
	cp.unitButtons = nil
	cp.title = ""
	cp.visible = false
	cp.selectedFactory = nil

	if factory == nil || !factory.CanProduce() {
		return
	}

	cp.visible = true
	cp.selectedFactory = factory
	cp.title = "PRODUCE"

	// Create buttons for each unit option
	options := entity.GetProducibleUnits(factory.Type)
	y := cp.topOffset + panelPadding + 20 // +20 for title

	for _, def := range options {
		btn := NewUnitButton(
			panelPadding,
			y,
			commandPanelWidth-panelPadding*2,
			buttonHeight,
			def,
		)
		// Set initial queue count
		btn.QueueCount = factory.GetQueueCount(def.Type)
		cp.unitButtons = append(cp.unitButtons, btn)
		y += buttonHeight + buttonMargin
	}
}

// UpdateQueueCounts updates the queue count display for factory buttons
func (cp *CommandPanel) UpdateQueueCounts() {
	if cp.selectedFactory == nil {
		return
	}
	for _, btn := range cp.unitButtons {
		btn.QueueCount = cp.selectedFactory.GetQueueCount(btn.UnitDef.Type)
	}
}

// Update handles input for the command panel, returns building def if clicked
func (cp *CommandPanel) Update(mousePos emath.Vec2, leftClicked bool) *entity.BuildingDef {
	if !cp.visible {
		return nil
	}

	var clickedDef *entity.BuildingDef

	for _, btn := range cp.buttons {
		if btn.Contains(mousePos) {
			if leftClicked {
				btn.State = ButtonPressed
				clickedDef = btn.BuildingDef
			} else {
				btn.State = ButtonHovered
			}
		} else {
			btn.State = ButtonNormal
		}
	}

	// Also update unit button states (hover only, clicking handled by UpdateUnit)
	for _, btn := range cp.unitButtons {
		if btn.Contains(mousePos) {
			if !leftClicked {
				btn.State = ButtonHovered
			}
		} else {
			btn.State = ButtonNormal
		}
	}

	return clickedDef
}

// UpdateUnit handles input for unit production buttons, returns unit def if clicked
func (cp *CommandPanel) UpdateUnit(mousePos emath.Vec2, leftClicked bool) *entity.UnitDef {
	if !cp.visible {
		return nil
	}

	var clickedDef *entity.UnitDef

	for _, btn := range cp.unitButtons {
		if btn.Contains(mousePos) {
			if leftClicked {
				btn.State = ButtonPressed
				clickedDef = btn.UnitDef
			} else {
				btn.State = ButtonHovered
			}
		} else {
			btn.State = ButtonNormal
		}
	}

	return clickedDef
}

// Contains checks if a point is inside the panel
func (cp *CommandPanel) Contains(p emath.Vec2) bool {
	if !cp.visible {
		return false
	}
	return cp.panel.Bounds.Contains(p)
}

// Draw renders the command panel
func (cp *CommandPanel) Draw(screen *ebiten.Image, resources *resource.Manager) {
	if !cp.visible {
		return
	}

	// Draw panel background
	cp.panel.Draw(screen)

	// Draw title
	ebitenutil.DebugPrintAt(screen, cp.title, int(panelPadding), int(cp.topOffset+panelPadding))

	// Draw building buttons
	for _, btn := range cp.buttons {
		btn.Draw(screen, resources)
	}

	// Draw unit buttons
	for _, btn := range cp.unitButtons {
		btn.Draw(screen, resources)
	}
}
