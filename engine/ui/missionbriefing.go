package ui

import (
	"github.com/bklimczak/tanks/engine/campaign"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

type MissionBriefing struct {
	mission      *campaign.Mission
	screenWidth  float64
	screenHeight float64
	visible      bool
}

func NewMissionBriefing() *MissionBriefing {
	return &MissionBriefing{}
}

func (mb *MissionBriefing) Show(mission *campaign.Mission) {
	mb.mission = mission
	mb.visible = true
}

func (mb *MissionBriefing) Hide() {
	mb.visible = false
}

func (mb *MissionBriefing) IsVisible() bool {
	return mb.visible
}

func (mb *MissionBriefing) UpdateSize(w, h float64) {
	mb.screenWidth = w
	mb.screenHeight = h
}

func (mb *MissionBriefing) Update(enter, escape bool) (start, cancelled bool) {
	if escape {
		return false, true
	}
	if enter {
		return true, false
	}
	return false, false
}

func (mb *MissionBriefing) Draw(screen *ebiten.Image) {
	if !mb.visible || mb.mission == nil {
		return
	}

	overlayColor := color.RGBA{0, 0, 0, 220}
	vector.FillRect(screen, 0, 0, float32(mb.screenWidth), float32(mb.screenHeight), overlayColor, false)

	boxWidth := 600.0
	boxHeight := 400.0
	boxX := mb.screenWidth/2 - boxWidth/2
	boxY := mb.screenHeight/2 - boxHeight/2

	boxColor := color.RGBA{25, 30, 40, 250}
	borderColor := color.RGBA{70, 90, 120, 255}
	vector.FillRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), boxColor, false)
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), 2, borderColor, false)

	headerColor := color.RGBA{35, 45, 60, 255}
	vector.FillRect(screen, float32(boxX), float32(boxY), float32(boxWidth), 50, headerColor, false)

	title := mb.mission.Name
	if mb.mission.Briefing != nil && mb.mission.Briefing.Title != "" {
		title = mb.mission.Briefing.Title
	}
	titleX := int(boxX) + int(boxWidth)/2 - len(title)*3
	titleY := int(boxY) + 18
	ebitenutil.DebugPrintAt(screen, title, titleX, titleY)

	contentX := int(boxX) + 30
	currentY := int(boxY) + 70

	ebitenutil.DebugPrintAt(screen, "MISSION BRIEFING", contentX, currentY)
	currentY += 25

	bgText := mb.mission.Description
	if mb.mission.Briefing != nil && mb.mission.Briefing.Background != "" {
		bgText = mb.mission.Briefing.Background
	}
	lines := wrapTextForBriefing(bgText, 70)
	for _, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, contentX, currentY)
		currentY += 16
	}
	currentY += 15

	ebitenutil.DebugPrintAt(screen, "OBJECTIVES:", contentX, currentY)
	currentY += 20

	var objectives []string
	if mb.mission.Briefing != nil {
		objectives = mb.mission.Briefing.Objectives
	}
	if len(objectives) == 0 {
		for _, vc := range mb.mission.VictoryConditions {
			objectives = append(objectives, getVictoryDescription(vc))
		}
	}

	for i, obj := range objectives {
		bulletText := "  > " + obj
		if i < 9 {
			bulletText = "  " + string('1'+i) + ". " + obj
		}
		ebitenutil.DebugPrintAt(screen, bulletText, contentX, currentY)
		currentY += 18
	}
	currentY += 15

	if len(mb.mission.DefeatConditions) > 0 {
		ebitenutil.DebugPrintAt(screen, "DEFEAT CONDITIONS:", contentX, currentY)
		currentY += 20

		for _, dc := range mb.mission.DefeatConditions {
			defeatText := "  ! " + getDefeatDescription(dc)
			ebitenutil.DebugPrintAt(screen, defeatText, contentX, currentY)
			currentY += 18
		}
	}

	helpText := "ENTER: Start Mission | ESC: Back"
	helpX := int(boxX) + int(boxWidth)/2 - len(helpText)*3
	helpY := int(boxY) + int(boxHeight) - 30
	ebitenutil.DebugPrintAt(screen, helpText, helpX, helpY)
}

func wrapTextForBriefing(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var lines []string
	for len(text) > maxLen {
		splitAt := maxLen
		for i := maxLen; i > 0; i-- {
			if text[i] == ' ' {
				splitAt = i
				break
			}
		}
		lines = append(lines, text[:splitAt])
		text = text[splitAt+1:]
	}
	if len(text) > 0 {
		lines = append(lines, text)
	}
	return lines
}

func getVictoryDescription(vc campaign.VictoryConditionDef) string {
	switch vc.Type {
	case "destroy_all_enemies":
		return "Destroy all enemy units and buildings"
	case "survive":
		return "Survive for " + formatDuration(vc.Duration)
	case "reach_zone":
		return "Move units to the target zone"
	case "reach_zone_all":
		return "Move all units to the target zone"
	case "destroy_count":
		return "Destroy " + intToString(vc.Count) + " enemies"
	case "build_structure":
		return "Build a " + vc.Target
	case "train_units":
		return "Train " + intToString(vc.Count) + " units"
	default:
		return "Complete objective"
	}
}

func getDefeatDescription(dc campaign.DefeatConditionDef) string {
	switch dc.Type {
	case "lose_all_units":
		return "Lose all units"
	case "lose_all_buildings":
		return "Lose all buildings"
	case "lose_all":
		return "Lose all units and buildings"
	case "time_limit":
		return "Time limit: " + formatDuration(dc.Duration) + " seconds"
	default:
		return "Mission failed"
	}
}

func formatDuration(seconds float64) string {
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	if mins > 0 {
		return intToString(mins) + "m " + intToString(secs) + "s"
	}
	return intToString(int(seconds)) + "s"
}

func intToString(val int) string {
	if val == 0 {
		return "0"
	}
	if val < 0 {
		return "-" + intToString(-val)
	}
	result := ""
	for val > 0 {
		result = string('0'+byte(val%10)) + result
		val /= 10
	}
	return result
}

