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
	commandPanelWidth = 180
	buttonHeight      = 50
	buttonMargin      = 5
	panelPadding      = 10
	scrollSpeed       = 30.0
)

type ButtonState int

const (
	ButtonNormal ButtonState = iota
	ButtonHovered
	ButtonPressed
	ButtonDisabled
)

type CommandButton struct {
	Bounds      emath.Rect
	Label       string
	Description string
	BuildingDef *entity.BuildingDef
	State       ButtonState
	OnClick     func()
}

func NewCommandButton(x, y, w, h float64, def *entity.BuildingDef) *CommandButton {
	return &CommandButton{
		Bounds:      emath.NewRect(x, y, w, h),
		Label:       def.Name,
		Description: def.Description,
		BuildingDef: def,
		State:       ButtonNormal,
	}
}

type UnitButton struct {
	Bounds     emath.Rect
	Label      string
	UnitDef    *entity.UnitDef
	State      ButtonState
	QueueCount int
}

func NewUnitButton(x, y, w, h float64, def *entity.UnitDef) *UnitButton {
	return &UnitButton{
		Bounds:  emath.NewRect(x, y, w, h),
		Label:   def.Name,
		UnitDef: def,
		State:   ButtonNormal,
	}
}
func (b *UnitButton) Contains(p emath.Vec2) bool {
	return b.Bounds.Contains(p)
}
func (b *UnitButton) Draw(screen *ebiten.Image, resources *resource.Manager) {
	x := float32(b.Bounds.Pos.X)
	y := float32(b.Bounds.Pos.Y)
	w := float32(b.Bounds.Size.X)
	h := float32(b.Bounds.Size.Y)
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
	vector.FillRect(screen, x, y, w, h, bgColor, false)
	vector.StrokeRect(screen, x, y, w, h, 1, borderColor, false)
	previewSize := float32(16)
	previewX := x + 8
	previewY := y + 8
	vector.FillRect(screen, previewX, previewY, previewSize, previewSize, b.UnitDef.Color, false)
	vector.StrokeRect(screen, previewX, previewY, previewSize, previewSize, 1, color.RGBA{100, 100, 100, 255}, false)
	labelX := int(previewX + previewSize + 8)
	labelY := int(y + 6)
	ebitenutil.DebugPrintAt(screen, b.Label, labelX, labelY)
	costY := int(y + 22)
	costStr := ""
	if metal, ok := b.UnitDef.Cost[resource.Metal]; ok && metal > 0 {
		costStr += fmt.Sprintf("M:%.0f ", metal)
	}
	if energy, ok := b.UnitDef.Cost[resource.Energy]; ok && energy > 0 {
		costStr += fmt.Sprintf("E:%.0f", energy)
	}
	ebitenutil.DebugPrintAt(screen, costStr, labelX, costY)
	timeStr := fmt.Sprintf("%.0fs", b.UnitDef.BuildTime)
	ebitenutil.DebugPrintAt(screen, timeStr, int(x+w-35), costY)
	if b.QueueCount > 0 {
		countStr := fmt.Sprintf("%d", b.QueueCount)
		countX := int(x + w - 20)
		countY := int(y + h - 16)
		vector.FillRect(screen, float32(countX-4), float32(countY-2), 20, 14, color.RGBA{0, 100, 0, 200}, false)
		ebitenutil.DebugPrintAt(screen, countStr, countX, countY)
	}
}
func (b *CommandButton) Contains(p emath.Vec2) bool {
	return b.Bounds.Contains(p)
}
func (b *CommandButton) Draw(screen *ebiten.Image, resources *resource.Manager) {
	x := float32(b.Bounds.Pos.X)
	y := float32(b.Bounds.Pos.Y)
	w := float32(b.Bounds.Size.X)
	h := float32(b.Bounds.Size.Y)
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
	vector.FillRect(screen, x, y, w, h, bgColor, false)
	vector.StrokeRect(screen, x, y, w, h, 1, borderColor, false)
	previewSize := float32(16)
	previewX := x + 8
	previewY := y + 8
	vector.FillRect(screen, previewX, previewY, previewSize, previewSize, b.BuildingDef.Color, false)
	vector.StrokeRect(screen, previewX, previewY, previewSize, previewSize, 1, color.RGBA{100, 100, 100, 255}, false)
	labelX := int(previewX + previewSize + 8)
	labelY := int(y + 6)
	ebitenutil.DebugPrintAt(screen, b.Label, labelX, labelY)
	costY := int(y + 22)
	costStr := ""
	if metal, ok := b.BuildingDef.Cost[resource.Metal]; ok && metal > 0 {
		costStr += fmt.Sprintf("M:%.0f ", metal)
	}
	if energy, ok := b.BuildingDef.Cost[resource.Energy]; ok && energy > 0 {
		costStr += fmt.Sprintf("E:%.0f", energy)
	}
	ebitenutil.DebugPrintAt(screen, costStr, labelX, costY)
	timeStr := fmt.Sprintf("%.0fs", b.BuildingDef.BuildTime)
	ebitenutil.DebugPrintAt(screen, timeStr, int(x+w-35), costY)
}

type CommandPanel struct {
	panel           *Panel
	buttons         []*CommandButton
	unitButtons     []*UnitButton
	visible         bool
	topOffset       float64
	title           string
	selectedFactory *entity.Building

	// Scroll state
	scrollOffset    float64
	maxScrollOffset float64
	contentHeight   float64
	visibleHeight   float64
}

func NewCommandPanel(topOffset float64, screenHeight float64) *CommandPanel {
	return &CommandPanel{
		panel:     NewPanel(0, topOffset, commandPanelWidth, screenHeight-topOffset),
		topOffset: topOffset,
		visible:   false,
	}
}
func (cp *CommandPanel) UpdateHeight(screenHeight float64) {
	cp.panel.Bounds.Size.Y = screenHeight - cp.topOffset
	cp.calculateScrollBounds()
}
func (cp *CommandPanel) Width() float64 {
	return commandPanelWidth
}

func (cp *CommandPanel) calculateScrollBounds() {
	buttonCount := len(cp.buttons) + len(cp.unitButtons)
	if buttonCount == 0 {
		cp.contentHeight = 0
		cp.maxScrollOffset = 0
		return
	}

	cp.contentHeight = float64(buttonCount) * (buttonHeight + buttonMargin)

	titleAreaHeight := 20.0 + panelPadding
	cp.visibleHeight = cp.panel.Bounds.Size.Y - titleAreaHeight - panelPadding*2

	if cp.contentHeight > cp.visibleHeight {
		cp.maxScrollOffset = cp.contentHeight - cp.visibleHeight
	} else {
		cp.maxScrollOffset = 0
	}

	cp.clampScroll()
}

func (cp *CommandPanel) clampScroll() {
	if cp.scrollOffset < 0 {
		cp.scrollOffset = 0
	}
	if cp.scrollOffset > cp.maxScrollOffset {
		cp.scrollOffset = cp.maxScrollOffset
	}
}

func (cp *CommandPanel) HandleScroll(wheelY float64) {
	if !cp.visible || cp.maxScrollOffset == 0 {
		return
	}

	cp.scrollOffset -= wheelY * scrollSpeed
	cp.clampScroll()
	cp.updateButtonPositions()
}

func (cp *CommandPanel) isButtonVisible(btnY, btnHeight float64) bool {
	buttonsStartY := cp.topOffset + panelPadding + 20
	buttonsEndY := cp.topOffset + cp.panel.Bounds.Size.Y - panelPadding

	return btnY+btnHeight >= buttonsStartY && btnY <= buttonsEndY
}

func (cp *CommandPanel) updateButtonPositions() {
	buttonsStartY := cp.topOffset + panelPadding + 20

	for i, btn := range cp.buttons {
		baseY := buttonsStartY + float64(i)*(buttonHeight+buttonMargin)
		btn.Bounds.Pos.Y = baseY - cp.scrollOffset
	}

	buildingCount := len(cp.buttons)
	for i, btn := range cp.unitButtons {
		baseY := buttonsStartY + float64(buildingCount+i)*(buttonHeight+buttonMargin)
		btn.Bounds.Pos.Y = baseY - cp.scrollOffset
	}
}
func (cp *CommandPanel) IsVisible() bool {
	return cp.visible
}
func (cp *CommandPanel) SetVisible(visible bool) {
	cp.visible = visible
	if !visible {
		cp.buttons = nil
		cp.unitButtons = nil
		cp.title = ""
		cp.selectedFactory = nil
		cp.scrollOffset = 0
	}
}
func (cp *CommandPanel) SetBuildOptions(units []*entity.Unit) {
	cp.buttons = nil
	cp.unitButtons = nil
	cp.title = ""
	cp.visible = false
	cp.selectedFactory = nil
	cp.scrollOffset = 0
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
	if constructor == nil || selectedCount != 1 {
		return
	}
	cp.visible = true
	cp.title = "BUILD"
	options := constructor.GetBuildOptions()
	for i, def := range options {
		btn := NewCommandButton(
			panelPadding,
			0,
			commandPanelWidth-panelPadding*2,
			buttonHeight,
			def,
		)
		cp.buttons = append(cp.buttons, btn)
		_ = i
	}
	cp.calculateScrollBounds()
	cp.updateButtonPositions()
}
func (cp *CommandPanel) SetFactoryOptions(factory *entity.Building) {
	cp.buttons = nil
	cp.unitButtons = nil
	cp.title = ""
	cp.visible = false
	cp.selectedFactory = nil
	cp.scrollOffset = 0
	if factory == nil || !factory.CanProduce() {
		return
	}
	cp.visible = true
	cp.selectedFactory = factory
	cp.title = "PRODUCE"
	options := entity.GetProducibleUnits(factory.Type)
	for i, def := range options {
		btn := NewUnitButton(
			panelPadding,
			0,
			commandPanelWidth-panelPadding*2,
			buttonHeight,
			def,
		)
		btn.QueueCount = factory.GetQueueCount(def.Type)
		cp.unitButtons = append(cp.unitButtons, btn)
		_ = i
	}
	cp.calculateScrollBounds()
	cp.updateButtonPositions()
}

func (cp *CommandPanel) SetBuildingBuildOptions(building *entity.Building) {
	cp.buttons = nil
	cp.unitButtons = nil
	cp.title = ""
	cp.visible = false
	cp.selectedFactory = nil
	cp.scrollOffset = 0
	if building == nil || building.Def == nil || len(building.Def.BuildableStructures) == 0 {
		return
	}
	cp.visible = true
	cp.title = "BUILD"
	options := entity.GetBuildableStructures(building.Type)
	for i, def := range options {
		btn := NewCommandButton(
			panelPadding,
			0,
			commandPanelWidth-panelPadding*2,
			buttonHeight,
			def,
		)
		cp.buttons = append(cp.buttons, btn)
		_ = i
	}
	cp.calculateScrollBounds()
	cp.updateButtonPositions()
}

func (cp *CommandPanel) UpdateQueueCounts() {
	if cp.selectedFactory == nil {
		return
	}
	for _, btn := range cp.unitButtons {
		btn.QueueCount = cp.selectedFactory.GetQueueCount(btn.UnitDef.Type)
	}
}
func (cp *CommandPanel) Update(mousePos emath.Vec2, leftClicked bool) *entity.BuildingDef {
	if !cp.visible {
		return nil
	}
	var clickedDef *entity.BuildingDef
	for _, btn := range cp.buttons {
		if !cp.isButtonVisible(btn.Bounds.Pos.Y, btn.Bounds.Size.Y) {
			btn.State = ButtonNormal
			continue
		}
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
	for _, btn := range cp.unitButtons {
		if !cp.isButtonVisible(btn.Bounds.Pos.Y, btn.Bounds.Size.Y) {
			btn.State = ButtonNormal
			continue
		}
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

func (cp *CommandPanel) GetHoveredBuilding(mousePos emath.Vec2) *entity.BuildingDef {
	if !cp.visible {
		return nil
	}
	for _, btn := range cp.buttons {
		if !cp.isButtonVisible(btn.Bounds.Pos.Y, btn.Bounds.Size.Y) {
			continue
		}
		if btn.Contains(mousePos) {
			return btn.BuildingDef
		}
	}
	return nil
}

func (cp *CommandPanel) GetHoveredUnit(mousePos emath.Vec2) *entity.UnitDef {
	if !cp.visible {
		return nil
	}
	for _, btn := range cp.unitButtons {
		if !cp.isButtonVisible(btn.Bounds.Pos.Y, btn.Bounds.Size.Y) {
			continue
		}
		if btn.Contains(mousePos) {
			return btn.UnitDef
		}
	}
	return nil
}

func (cp *CommandPanel) GetHoveredButtonBounds(mousePos emath.Vec2) *emath.Rect {
	if !cp.visible {
		return nil
	}
	for _, btn := range cp.buttons {
		if !cp.isButtonVisible(btn.Bounds.Pos.Y, btn.Bounds.Size.Y) {
			continue
		}
		if btn.Contains(mousePos) {
			return &btn.Bounds
		}
	}
	for _, btn := range cp.unitButtons {
		if !cp.isButtonVisible(btn.Bounds.Pos.Y, btn.Bounds.Size.Y) {
			continue
		}
		if btn.Contains(mousePos) {
			return &btn.Bounds
		}
	}
	return nil
}
func (cp *CommandPanel) UpdateUnit(mousePos emath.Vec2, leftClicked bool) *entity.UnitDef {
	if !cp.visible {
		return nil
	}
	var clickedDef *entity.UnitDef
	for _, btn := range cp.unitButtons {
		if !cp.isButtonVisible(btn.Bounds.Pos.Y, btn.Bounds.Size.Y) {
			btn.State = ButtonNormal
			continue
		}
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
func (cp *CommandPanel) UpdateUnitRightClick(mousePos emath.Vec2, rightClicked bool) *entity.UnitDef {
	if !cp.visible || !rightClicked {
		return nil
	}
	for _, btn := range cp.unitButtons {
		if !cp.isButtonVisible(btn.Bounds.Pos.Y, btn.Bounds.Size.Y) {
			continue
		}
		if btn.Contains(mousePos) && btn.QueueCount > 0 {
			return btn.UnitDef
		}
	}
	return nil
}
func (cp *CommandPanel) Contains(p emath.Vec2) bool {
	if !cp.visible {
		return false
	}
	return cp.panel.Bounds.Contains(p)
}
func (cp *CommandPanel) Draw(screen *ebiten.Image, resources *resource.Manager) {
	if !cp.visible {
		return
	}
	cp.panel.Draw(screen)
	ebitenutil.DebugPrintAt(screen, cp.title, int(panelPadding), int(cp.topOffset+panelPadding))

	for _, btn := range cp.buttons {
		if cp.isButtonVisible(btn.Bounds.Pos.Y, buttonHeight) {
			btn.Draw(screen, resources)
		}
	}

	for _, btn := range cp.unitButtons {
		if cp.isButtonVisible(btn.Bounds.Pos.Y, buttonHeight) {
			btn.Draw(screen, resources)
		}
	}

	if cp.maxScrollOffset > 0 {
		buttonsStartY := cp.topOffset + panelPadding + 20
		cp.drawScrollIndicator(screen, buttonsStartY)
	}
}

func (cp *CommandPanel) drawScrollIndicator(screen *ebiten.Image, startY float64) {
	indicatorWidth := 4.0
	indicatorX := commandPanelWidth - indicatorWidth - 2
	endY := cp.topOffset + cp.panel.Bounds.Size.Y - panelPadding
	trackHeight := endY - startY

	trackColor := color.RGBA{40, 40, 50, 255}
	vector.FillRect(screen, float32(indicatorX), float32(startY),
		float32(indicatorWidth), float32(trackHeight), trackColor, false)

	visibleRatio := cp.visibleHeight / cp.contentHeight
	thumbHeight := trackHeight * visibleRatio
	if thumbHeight < 20 {
		thumbHeight = 20
	}

	scrollRatio := 0.0
	if cp.maxScrollOffset > 0 {
		scrollRatio = cp.scrollOffset / cp.maxScrollOffset
	}
	thumbY := startY + (trackHeight-thumbHeight)*scrollRatio

	thumbColor := color.RGBA{100, 100, 120, 255}
	vector.FillRect(screen, float32(indicatorX), float32(thumbY),
		float32(indicatorWidth), float32(thumbHeight), thumbColor, false)
}
