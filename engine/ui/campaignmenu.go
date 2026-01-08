package ui

import (
	"fmt"
	"github.com/bklimczak/tanks/engine/campaign"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

type CampaignMenu struct {
	missions        []*campaign.Mission
	selectedIndex   int
	hoveredIndex    int
	screenWidth     float64
	screenHeight    float64
	campaignManager *campaign.Manager
	currentCampaign string
	visible         bool
}

func NewCampaignMenu() *CampaignMenu {
	return &CampaignMenu{
		selectedIndex: 0,
		hoveredIndex:  -1,
	}
}

func (cm *CampaignMenu) SetCampaignManager(mgr *campaign.Manager) {
	cm.campaignManager = mgr
}

func (cm *CampaignMenu) Show(campaignID string, missions []*campaign.Mission) {
	cm.currentCampaign = campaignID
	cm.missions = missions
	cm.selectedIndex = 0
	cm.hoveredIndex = -1
	cm.visible = true
}

func (cm *CampaignMenu) Hide() {
	cm.visible = false
}

func (cm *CampaignMenu) IsVisible() bool {
	return cm.visible
}

func (cm *CampaignMenu) UpdateSize(w, h float64) {
	cm.screenWidth = w
	cm.screenHeight = h
}

func (cm *CampaignMenu) Update(up, down, enter, escape bool) (selectedMission *campaign.Mission, cancelled bool) {
	if escape {
		return nil, true
	}

	if up {
		cm.selectedIndex--
		if cm.selectedIndex < 0 {
			cm.selectedIndex = len(cm.missions) - 1
		}
		cm.hoveredIndex = -1
	}
	if down {
		cm.selectedIndex++
		if cm.selectedIndex >= len(cm.missions) {
			cm.selectedIndex = 0
		}
		cm.hoveredIndex = -1
	}
	if enter && len(cm.missions) > 0 {
		mission := cm.missions[cm.selectedIndex]
		if cm.isMissionUnlocked(mission) {
			return mission, false
		}
	}
	return nil, false
}

func (cm *CampaignMenu) UpdateHover(mousePos emath.Vec2) {
	cm.hoveredIndex = -1
	missionHeight := 60.0
	missionWidth := 500.0
	startX := cm.screenWidth/2 - missionWidth/2
	startY := cm.screenHeight/2 - float64(len(cm.missions))*missionHeight/2

	for i := range cm.missions {
		y := startY + float64(i)*missionHeight
		rect := emath.NewRect(startX, y, missionWidth, missionHeight-5)
		if rect.Contains(mousePos) {
			cm.hoveredIndex = i
			cm.selectedIndex = i
			break
		}
	}
}

func (cm *CampaignMenu) HandleClick(mousePos emath.Vec2) *campaign.Mission {
	missionHeight := 60.0
	missionWidth := 500.0
	startX := cm.screenWidth/2 - missionWidth/2
	startY := cm.screenHeight/2 - float64(len(cm.missions))*missionHeight/2

	for i, mission := range cm.missions {
		y := startY + float64(i)*missionHeight
		rect := emath.NewRect(startX, y, missionWidth, missionHeight-5)
		if rect.Contains(mousePos) {
			if cm.isMissionUnlocked(mission) {
				return mission
			}
			return nil
		}
	}
	return nil
}

func (cm *CampaignMenu) isMissionUnlocked(mission *campaign.Mission) bool {
	if cm.campaignManager == nil {
		return true
	}
	return cm.campaignManager.IsMissionUnlocked(cm.currentCampaign, mission.ID)
}

func (cm *CampaignMenu) isMissionCompleted(mission *campaign.Mission) bool {
	if cm.campaignManager == nil {
		return false
	}
	return cm.campaignManager.IsMissionCompleted(mission.ID)
}

func (cm *CampaignMenu) Draw(screen *ebiten.Image) {
	if !cm.visible {
		return
	}

	overlayColor := color.RGBA{0, 0, 0, 200}
	vector.FillRect(screen, 0, 0, float32(cm.screenWidth), float32(cm.screenHeight), overlayColor, false)

	boxWidth := 550.0
	boxHeight := float64(len(cm.missions))*60 + 100
	boxX := cm.screenWidth/2 - boxWidth/2
	boxY := cm.screenHeight/2 - boxHeight/2

	boxColor := color.RGBA{30, 30, 40, 240}
	borderColor := color.RGBA{80, 80, 100, 255}
	vector.FillRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), boxColor, false)
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), 2, borderColor, false)

	title := "SELECT MISSION"
	titleX := int(boxX) + int(boxWidth)/2 - len(title)*3
	titleY := int(boxY) + 15
	ebitenutil.DebugPrintAt(screen, title, titleX, titleY)

	missionHeight := 60.0
	missionWidth := 500.0
	startX := cm.screenWidth/2 - missionWidth/2
	startY := boxY + 50

	for i, mission := range cm.missions {
		y := startY + float64(i)*missionHeight

		isUnlocked := cm.isMissionUnlocked(mission)
		isCompleted := cm.isMissionCompleted(mission)

		var bgColor color.RGBA
		if !isUnlocked {
			bgColor = color.RGBA{30, 30, 35, 255}
		} else if cm.hoveredIndex == i || cm.selectedIndex == i {
			bgColor = color.RGBA{55, 55, 70, 255}
		} else {
			bgColor = color.RGBA{40, 40, 55, 255}
		}

		vector.FillRect(screen, float32(startX), float32(y), float32(missionWidth), float32(missionHeight-5), bgColor, false)

		slotBorderColor := color.RGBA{60, 60, 75, 255}
		if cm.selectedIndex == i && isUnlocked {
			slotBorderColor = color.RGBA{90, 90, 120, 255}
		}
		vector.StrokeRect(screen, float32(startX), float32(y), float32(missionWidth), float32(missionHeight-5), 1, slotBorderColor, false)

		statusIcon := "  "
		if isCompleted {
			statusIcon = "[X]"
		} else if isUnlocked {
			statusIcon = "[ ]"
		} else {
			statusIcon = "[?]"
		}

		nameText := fmt.Sprintf("%s %d. %s", statusIcon, i+1, mission.Name)
		textX := int(startX) + 10
		textY := int(y) + 8
		ebitenutil.DebugPrintAt(screen, nameText, textX, textY)

		descText := mission.Description
		if len(descText) > 60 {
			descText = descText[:57] + "..."
		}
		if !isUnlocked {
			descText = "Complete previous mission to unlock"
		}
		descY := int(y) + 28
		ebitenutil.DebugPrintAt(screen, descText, textX+20, descY)

		if cm.selectedIndex == i && isUnlocked {
			markerY := float32(y) + float32(missionHeight-5)/2
			vector.DrawFilledCircle(screen, float32(startX)-10, markerY, 4, color.RGBA{100, 200, 100, 255}, false)
		}
	}

	helpText := "UP/DOWN: Select | ENTER: Start | ESC: Back"
	helpX := int(boxX) + int(boxWidth)/2 - len(helpText)*3
	helpY := int(boxY) + int(boxHeight) - 25
	ebitenutil.DebugPrintAt(screen, helpText, helpX, helpY)
}
