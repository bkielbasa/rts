package ui

import (
	"fmt"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/save"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

type SaveMenuMode int

const (
	SaveModeSave SaveMenuMode = iota
	SaveModeLoad
)

type SaveMenu struct {
	mode          SaveMenuMode
	selectedSlot  int
	screenWidth   float64
	screenHeight  float64
	slots         []save.SaveSlotInfo
	hoveredSlot   int
	visible       bool
}

func NewSaveMenu() *SaveMenu {
	return &SaveMenu{
		selectedSlot: 0,
		hoveredSlot:  -1,
	}
}

func (sm *SaveMenu) Show(mode SaveMenuMode, slots []save.SaveSlotInfo) {
	sm.mode = mode
	sm.slots = slots
	sm.selectedSlot = 0
	sm.hoveredSlot = -1
	sm.visible = true
}

func (sm *SaveMenu) Hide() {
	sm.visible = false
}

func (sm *SaveMenu) IsVisible() bool {
	return sm.visible
}

func (sm *SaveMenu) Mode() SaveMenuMode {
	return sm.mode
}

func (sm *SaveMenu) UpdateSize(w, h float64) {
	sm.screenWidth = w
	sm.screenHeight = h
}

func (sm *SaveMenu) Update(up, down, enter, escape bool) (selectedSlot int, cancelled bool) {
	if escape {
		return -1, true
	}

	if up {
		sm.selectedSlot--
		if sm.selectedSlot < 0 {
			sm.selectedSlot = len(sm.slots) - 1
		}
		sm.hoveredSlot = -1
	}
	if down {
		sm.selectedSlot++
		if sm.selectedSlot >= len(sm.slots) {
			sm.selectedSlot = 0
		}
		sm.hoveredSlot = -1
	}
	if enter {
		if sm.mode == SaveModeLoad && sm.slots[sm.selectedSlot].Empty {
			return -1, false
		}
		return sm.selectedSlot, false
	}
	return -1, false
}

func (sm *SaveMenu) UpdateHover(mousePos emath.Vec2) {
	sm.hoveredSlot = -1
	slotHeight := 45.0
	slotWidth := 400.0
	startX := sm.screenWidth/2 - slotWidth/2
	startY := sm.screenHeight/2 - float64(len(sm.slots))*slotHeight/2 + 30

	for i := range sm.slots {
		y := startY + float64(i)*slotHeight
		rect := emath.NewRect(startX, y, slotWidth, slotHeight-5)
		if rect.Contains(mousePos) {
			sm.hoveredSlot = i
			sm.selectedSlot = i
			break
		}
	}
}

func (sm *SaveMenu) HandleClick(mousePos emath.Vec2) int {
	slotHeight := 45.0
	slotWidth := 400.0
	startX := sm.screenWidth/2 - slotWidth/2
	startY := sm.screenHeight/2 - float64(len(sm.slots))*slotHeight/2 + 30

	for i := range sm.slots {
		y := startY + float64(i)*slotHeight
		rect := emath.NewRect(startX, y, slotWidth, slotHeight-5)
		if rect.Contains(mousePos) {
			if sm.mode == SaveModeLoad && sm.slots[i].Empty {
				return -1
			}
			return i
		}
	}
	return -1
}

func (sm *SaveMenu) Draw(screen *ebiten.Image) {
	if !sm.visible {
		return
	}

	overlayColor := color.RGBA{0, 0, 0, 200}
	vector.FillRect(screen, 0, 0, float32(sm.screenWidth), float32(sm.screenHeight), overlayColor, false)

	boxWidth := 450.0
	boxHeight := float64(len(sm.slots))*45 + 80
	boxX := sm.screenWidth/2 - boxWidth/2
	boxY := sm.screenHeight/2 - boxHeight/2

	boxColor := color.RGBA{30, 30, 40, 240}
	borderColor := color.RGBA{80, 80, 100, 255}
	vector.FillRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), boxColor, false)
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), 2, borderColor, false)

	title := "SAVE GAME"
	if sm.mode == SaveModeLoad {
		title = "LOAD GAME"
	}
	titleX := int(boxX) + int(boxWidth)/2 - len(title)*3
	titleY := int(boxY) + 15
	ebitenutil.DebugPrintAt(screen, title, titleX, titleY)

	slotHeight := 45.0
	slotWidth := 400.0
	startX := sm.screenWidth/2 - slotWidth/2
	startY := boxY + 50

	for i, slot := range sm.slots {
		y := startY + float64(i)*slotHeight

		var bgColor color.RGBA
		if sm.hoveredSlot == i || sm.selectedSlot == i {
			bgColor = color.RGBA{55, 55, 70, 255}
		} else {
			bgColor = color.RGBA{40, 40, 55, 255}
		}

		if sm.mode == SaveModeLoad && slot.Empty {
			bgColor = color.RGBA{35, 35, 40, 255}
		}

		vector.FillRect(screen, float32(startX), float32(y), float32(slotWidth), float32(slotHeight-5), bgColor, false)

		slotBorderColor := color.RGBA{60, 60, 75, 255}
		if sm.selectedSlot == i {
			slotBorderColor = color.RGBA{90, 90, 120, 255}
		}
		vector.StrokeRect(screen, float32(startX), float32(y), float32(slotWidth), float32(slotHeight-5), 1, slotBorderColor, false)

		slotText := fmt.Sprintf("Slot %d - Empty", slot.Slot+1)
		if !slot.Empty && slot.Metadata != nil {
			timestamp := slot.Metadata.Timestamp.Format("Jan 2, 15:04")
			playTime := int(slot.Metadata.PlayTime)
			playMinutes := playTime / 60
			playSeconds := playTime % 60
			slotText = fmt.Sprintf("Slot %d: %s (%s) - %dm %ds",
				slot.Slot+1,
				slot.Metadata.Name,
				timestamp,
				playMinutes,
				playSeconds,
			)
		}

		textColor := color.RGBA{200, 200, 200, 255}
		if sm.mode == SaveModeLoad && slot.Empty {
			textColor = color.RGBA{100, 100, 100, 255}
		}
		_ = textColor

		textX := int(startX) + 10
		textY := int(y) + 12
		ebitenutil.DebugPrintAt(screen, slotText, textX, textY)

		if sm.selectedSlot == i {
			markerY := float32(y) + float32(slotHeight-5)/2
			vector.DrawFilledCircle(screen, float32(startX)-10, markerY, 4, color.RGBA{100, 200, 100, 255}, false)
		}
	}

	helpText := "UP/DOWN: Select | ENTER: Confirm | ESC: Cancel"
	helpX := int(boxX) + int(boxWidth)/2 - len(helpText)*3
	helpY := int(boxY) + int(boxHeight) - 25
	ebitenutil.DebugPrintAt(screen, helpText, helpX, helpY)
}
