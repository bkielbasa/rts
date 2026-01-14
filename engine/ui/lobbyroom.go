package ui

import (
	"fmt"
	"image/color"

	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type PlayerSlot struct {
	Name   string
	Ready  bool
	IsHost bool
	IsYou  bool
}

type LobbyRoomAction int

const (
	LobbyRoomActionNone LobbyRoomAction = iota
	LobbyRoomActionReady
	LobbyRoomActionStart
	LobbyRoomActionLeave
)

type LobbyRoom struct {
	screenWidth  float64
	screenHeight float64
	lobbyName    string
	lobbyID      string
	players      []PlayerSlot
	maxPlayers   int
	isHost       bool
	isReady      bool
	canStart     bool
	countdown    int
}

func NewLobbyRoom() *LobbyRoom {
	return &LobbyRoom{
		screenWidth:  1280,
		screenHeight: 720,
		maxPlayers:   4,
		players:      make([]PlayerSlot, 0),
	}
}

func (lr *LobbyRoom) UpdateSize(width, height float64) {
	lr.screenWidth = width
	lr.screenHeight = height
}

func (lr *LobbyRoom) SetLobby(id, name string, maxPlayers int, isHost bool) {
	lr.lobbyID = id
	lr.lobbyName = name
	lr.maxPlayers = maxPlayers
	lr.isHost = isHost
}

func (lr *LobbyRoom) SetPlayers(players []PlayerSlot) {
	lr.players = players
}

func (lr *LobbyRoom) SetReady(ready bool) {
	lr.isReady = ready
}

func (lr *LobbyRoom) SetCanStart(canStart bool) {
	lr.canStart = canStart
}

func (lr *LobbyRoom) SetCountdown(seconds int) {
	lr.countdown = seconds
}

func (lr *LobbyRoom) Update(confirmPressed bool) LobbyRoomAction {
	if confirmPressed {
		if lr.isHost && lr.canStart {
			return LobbyRoomActionStart
		}
		return LobbyRoomActionReady
	}
	return LobbyRoomActionNone
}

func (lr *LobbyRoom) HandleClick(pos emath.Vec2) LobbyRoomAction {
	panelWidth := 500.0
	panelHeight := 400.0
	panelX := (lr.screenWidth - panelWidth) / 2
	panelY := (lr.screenHeight - panelHeight) / 2

	buttonY := panelY + panelHeight - 50
	buttonWidth := 100.0
	buttonHeight := 35.0

	// Leave button
	leaveBounds := emath.NewRect(panelX+20, buttonY, buttonWidth, buttonHeight)
	if leaveBounds.Contains(pos) {
		return LobbyRoomActionLeave
	}

	// Ready button
	readyX := panelX + panelWidth/2 - buttonWidth/2
	readyBounds := emath.NewRect(readyX, buttonY, buttonWidth, buttonHeight)
	if readyBounds.Contains(pos) {
		return LobbyRoomActionReady
	}

	// Start button (host only)
	if lr.isHost {
		startBounds := emath.NewRect(panelX+panelWidth-buttonWidth-20, buttonY, buttonWidth, buttonHeight)
		if startBounds.Contains(pos) && lr.canStart {
			return LobbyRoomActionStart
		}
	}

	return LobbyRoomActionNone
}

func (lr *LobbyRoom) Draw(screen *ebiten.Image) {
	// Background overlay
	vector.FillRect(screen, 0, 0, float32(lr.screenWidth), float32(lr.screenHeight), color.RGBA{20, 25, 30, 255}, false)

	panelWidth := 500.0
	panelHeight := 400.0
	panelX := (lr.screenWidth - panelWidth) / 2
	panelY := (lr.screenHeight - panelHeight) / 2

	// Panel background
	panelColor := color.RGBA{40, 45, 55, 240}
	borderColor := color.RGBA{80, 100, 120, 255}
	vector.FillRect(screen, float32(panelX), float32(panelY), float32(panelWidth), float32(panelHeight), panelColor, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelWidth), float32(panelHeight), 2, borderColor, false)

	// Title
	title := lr.lobbyName
	if title == "" {
		title = "LOBBY"
	}
	titleX := int(panelX) + int(panelWidth)/2 - len(title)*3
	ebitenutil.DebugPrintAt(screen, title, titleX, int(panelY)+15)

	// Player slots
	slotY := panelY + 60
	slotHeight := 50.0
	slotWidth := panelWidth - 40

	playerColors := []color.RGBA{
		{100, 150, 255, 255}, // Blue - Player 1
		{255, 100, 100, 255}, // Red - Player 2
		{255, 255, 100, 255}, // Yellow - Player 3
		{200, 100, 255, 255}, // Purple - Player 4
	}

	for i := 0; i < lr.maxPlayers; i++ {
		y := slotY + float64(i)*slotHeight

		// Slot background
		var slotColor color.RGBA
		if i < len(lr.players) {
			slotColor = color.RGBA{50, 60, 70, 200}
		} else {
			slotColor = color.RGBA{35, 40, 50, 200}
		}
		vector.FillRect(screen, float32(panelX)+20, float32(y), float32(slotWidth), float32(slotHeight)-5, slotColor, false)

		// Player color indicator
		vector.FillRect(screen, float32(panelX)+25, float32(y)+5, 10, float32(slotHeight)-15, playerColors[i], false)

		if i < len(lr.players) {
			player := lr.players[i]

			// Player name
			name := player.Name
			if player.IsYou {
				name += " (You)"
			}
			if player.IsHost {
				name += " [Host]"
			}
			ebitenutil.DebugPrintAt(screen, name, int(panelX)+45, int(y)+15)

			// Ready status
			var statusText string
			var statusColor color.RGBA
			if player.Ready {
				statusText = "READY"
				statusColor = color.RGBA{100, 200, 100, 255}
			} else {
				statusText = "Not Ready"
				statusColor = color.RGBA{200, 200, 100, 255}
			}
			statusX := int(panelX) + int(slotWidth) - 50
			vector.FillRect(screen, float32(statusX)-10, float32(y)+10, 70, 25, statusColor, false)
			ebitenutil.DebugPrintAt(screen, statusText, statusX-5, int(y)+15)
		} else {
			// Empty slot
			ebitenutil.DebugPrintAt(screen, "Waiting for player...", int(panelX)+45, int(y)+15)
		}
	}

	// Countdown display
	if lr.countdown > 0 {
		countdownText := fmt.Sprintf("Starting in %d...", lr.countdown)
		countdownX := int(panelX) + int(panelWidth)/2 - len(countdownText)*3
		ebitenutil.DebugPrintAt(screen, countdownText, countdownX, int(panelY)+int(panelHeight)-90)
	}

	// Buttons
	buttonY := panelY + panelHeight - 50
	buttonWidth := 100.0
	buttonHeight := 35.0
	buttonColor := color.RGBA{60, 80, 100, 255}
	readyColor := color.RGBA{60, 120, 60, 255}
	startColor := color.RGBA{60, 60, 120, 255}

	// Leave button
	vector.FillRect(screen, float32(panelX)+20, float32(buttonY), float32(buttonWidth), float32(buttonHeight), buttonColor, false)
	vector.StrokeRect(screen, float32(panelX)+20, float32(buttonY), float32(buttonWidth), float32(buttonHeight), 1, borderColor, false)
	ebitenutil.DebugPrintAt(screen, "Leave", int(panelX)+20+int(buttonWidth)/2-15, int(buttonY)+10)

	// Ready button
	readyX := panelX + panelWidth/2 - buttonWidth/2
	btnColor := readyColor
	readyText := "Ready"
	if lr.isReady {
		btnColor = color.RGBA{120, 60, 60, 255}
		readyText = "Unready"
	}
	vector.FillRect(screen, float32(readyX), float32(buttonY), float32(buttonWidth), float32(buttonHeight), btnColor, false)
	vector.StrokeRect(screen, float32(readyX), float32(buttonY), float32(buttonWidth), float32(buttonHeight), 1, borderColor, false)
	ebitenutil.DebugPrintAt(screen, readyText, int(readyX)+int(buttonWidth)/2-len(readyText)*3, int(buttonY)+10)

	// Start button (host only)
	if lr.isHost {
		startX := panelX + panelWidth - buttonWidth - 20
		btnStartColor := startColor
		if !lr.canStart {
			btnStartColor = color.RGBA{50, 50, 60, 255}
		}
		vector.FillRect(screen, float32(startX), float32(buttonY), float32(buttonWidth), float32(buttonHeight), btnStartColor, false)
		vector.StrokeRect(screen, float32(startX), float32(buttonY), float32(buttonWidth), float32(buttonHeight), 1, borderColor, false)
		ebitenutil.DebugPrintAt(screen, "Start", int(startX)+int(buttonWidth)/2-15, int(buttonY)+10)
	}

	// Instructions
	var instructions string
	if lr.isHost {
		instructions = "ENTER: Toggle Ready | Click Start when all ready | ESC: Leave"
	} else {
		instructions = "ENTER: Toggle Ready | Waiting for host to start | ESC: Leave"
	}
	instrX := int(lr.screenWidth/2) - len(instructions)*3
	ebitenutil.DebugPrintAt(screen, instructions, instrX, int(lr.screenHeight)-30)
}
