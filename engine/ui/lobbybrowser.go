package ui

import (
	"fmt"
	"image/color"

	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type LobbyInfo struct {
	ID          string
	Name        string
	PlayerCount int
	MaxPlayers  int
	State       string
}

type LobbyBrowserAction int

const (
	LobbyActionNone LobbyBrowserAction = iota
	LobbyActionJoin
	LobbyActionCreate
	LobbyActionRefresh
	LobbyActionBack
)

type LobbyBrowser struct {
	screenWidth   float64
	screenHeight  float64
	lobbies       []LobbyInfo
	selectedIndex int
	scrollOffset  int
	maxVisible    int
	connecting    bool
	errorMessage  string
	serverAddress string
}

func NewLobbyBrowser() *LobbyBrowser {
	return &LobbyBrowser{
		screenWidth:   1280,
		screenHeight:  720,
		lobbies:       nil,
		selectedIndex: -1,
		scrollOffset:  0,
		maxVisible:    8,
		serverAddress: "localhost:8080",
	}
}

func (lb *LobbyBrowser) UpdateSize(width, height float64) {
	lb.screenWidth = width
	lb.screenHeight = height
}

func (lb *LobbyBrowser) SetLobbies(lobbies []LobbyInfo) {
	lb.lobbies = lobbies
	if lb.selectedIndex >= len(lobbies) {
		lb.selectedIndex = len(lobbies) - 1
	}
}

func (lb *LobbyBrowser) SetConnecting(connecting bool) {
	lb.connecting = connecting
}

func (lb *LobbyBrowser) SetError(err string) {
	lb.errorMessage = err
}

func (lb *LobbyBrowser) ClearError() {
	lb.errorMessage = ""
}

func (lb *LobbyBrowser) GetSelectedLobby() *LobbyInfo {
	if lb.selectedIndex >= 0 && lb.selectedIndex < len(lb.lobbies) {
		return &lb.lobbies[lb.selectedIndex]
	}
	return nil
}

func (lb *LobbyBrowser) Update(upPressed, downPressed, confirmPressed bool) LobbyBrowserAction {
	if upPressed && lb.selectedIndex > 0 {
		lb.selectedIndex--
		if lb.selectedIndex < lb.scrollOffset {
			lb.scrollOffset = lb.selectedIndex
		}
	}
	if downPressed && lb.selectedIndex < len(lb.lobbies)-1 {
		lb.selectedIndex++
		if lb.selectedIndex >= lb.scrollOffset+lb.maxVisible {
			lb.scrollOffset = lb.selectedIndex - lb.maxVisible + 1
		}
	}
	if confirmPressed && lb.selectedIndex >= 0 {
		return LobbyActionJoin
	}
	return LobbyActionNone
}

func (lb *LobbyBrowser) HandleClick(pos emath.Vec2) LobbyBrowserAction {
	panelWidth := 600.0
	panelHeight := 450.0
	panelX := (lb.screenWidth - panelWidth) / 2
	panelY := (lb.screenHeight - panelHeight) / 2

	// Check lobby list clicks
	listY := panelY + 60
	itemHeight := 40.0
	for i := 0; i < lb.maxVisible && lb.scrollOffset+i < len(lb.lobbies); i++ {
		itemY := listY + float64(i)*itemHeight
		bounds := emath.NewRect(panelX+20, itemY, panelWidth-40, itemHeight-5)
		if bounds.Contains(pos) {
			lb.selectedIndex = lb.scrollOffset + i
			return LobbyActionNone
		}
	}

	// Check button clicks
	buttonY := panelY + panelHeight - 50
	buttonWidth := 100.0
	buttonHeight := 35.0
	buttonSpacing := 20.0

	// Back button
	backBounds := emath.NewRect(panelX+20, buttonY, buttonWidth, buttonHeight)
	if backBounds.Contains(pos) {
		return LobbyActionBack
	}

	// Refresh button
	refreshBounds := emath.NewRect(panelX+20+buttonWidth+buttonSpacing, buttonY, buttonWidth, buttonHeight)
	if refreshBounds.Contains(pos) {
		return LobbyActionRefresh
	}

	// Create button
	createBounds := emath.NewRect(panelX+panelWidth-buttonWidth-20-buttonWidth-buttonSpacing, buttonY, buttonWidth, buttonHeight)
	if createBounds.Contains(pos) {
		return LobbyActionCreate
	}

	// Join button
	joinBounds := emath.NewRect(panelX+panelWidth-buttonWidth-20, buttonY, buttonWidth, buttonHeight)
	if joinBounds.Contains(pos) && lb.selectedIndex >= 0 {
		return LobbyActionJoin
	}

	return LobbyActionNone
}

func (lb *LobbyBrowser) UpdateHover(pos emath.Vec2) {
	panelWidth := 600.0
	panelHeight := 450.0
	panelX := (lb.screenWidth - panelWidth) / 2
	panelY := (lb.screenHeight - panelHeight) / 2

	listY := panelY + 60
	itemHeight := 40.0
	for i := 0; i < lb.maxVisible && lb.scrollOffset+i < len(lb.lobbies); i++ {
		itemY := listY + float64(i)*itemHeight
		bounds := emath.NewRect(panelX+20, itemY, panelWidth-40, itemHeight-5)
		if bounds.Contains(pos) {
			lb.selectedIndex = lb.scrollOffset + i
			return
		}
	}
}

func (lb *LobbyBrowser) Draw(screen *ebiten.Image) {
	// Background overlay
	vector.FillRect(screen, 0, 0, float32(lb.screenWidth), float32(lb.screenHeight), color.RGBA{20, 25, 30, 255}, false)

	panelWidth := 600.0
	panelHeight := 450.0
	panelX := (lb.screenWidth - panelWidth) / 2
	panelY := (lb.screenHeight - panelHeight) / 2

	// Panel background
	panelColor := color.RGBA{40, 45, 55, 240}
	borderColor := color.RGBA{80, 100, 120, 255}
	vector.FillRect(screen, float32(panelX), float32(panelY), float32(panelWidth), float32(panelHeight), panelColor, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelWidth), float32(panelHeight), 2, borderColor, false)

	// Title
	title := "MULTIPLAYER LOBBY BROWSER"
	titleX := int(panelX) + int(panelWidth)/2 - len(title)*3
	ebitenutil.DebugPrintAt(screen, title, titleX, int(panelY)+15)

	// Server address
	serverText := fmt.Sprintf("Server: %s", lb.serverAddress)
	ebitenutil.DebugPrintAt(screen, serverText, int(panelX)+20, int(panelY)+40)

	// Lobby list
	listY := panelY + 60
	itemHeight := 40.0

	if lb.connecting {
		connectingText := "Connecting to server..."
		ebitenutil.DebugPrintAt(screen, connectingText, int(panelX)+int(panelWidth)/2-len(connectingText)*3, int(listY)+100)
	} else if lb.errorMessage != "" {
		errorText := fmt.Sprintf("Error: %s", lb.errorMessage)
		ebitenutil.DebugPrintAt(screen, errorText, int(panelX)+20, int(listY)+100)
	} else if len(lb.lobbies) == 0 {
		noLobbiesText := "No lobbies available. Create one!"
		ebitenutil.DebugPrintAt(screen, noLobbiesText, int(panelX)+int(panelWidth)/2-len(noLobbiesText)*3, int(listY)+100)
	} else {
		// Draw lobby items
		for i := 0; i < lb.maxVisible && lb.scrollOffset+i < len(lb.lobbies); i++ {
			lobby := lb.lobbies[lb.scrollOffset+i]
			itemY := listY + float64(i)*itemHeight

			// Item background
			var itemColor color.RGBA
			if lb.scrollOffset+i == lb.selectedIndex {
				itemColor = color.RGBA{60, 100, 60, 200}
			} else {
				itemColor = color.RGBA{50, 55, 65, 200}
			}
			vector.FillRect(screen, float32(panelX)+20, float32(itemY), float32(panelWidth)-40, float32(itemHeight)-5, itemColor, false)

			if lb.scrollOffset+i == lb.selectedIndex {
				vector.StrokeRect(screen, float32(panelX)+20, float32(itemY), float32(panelWidth)-40, float32(itemHeight)-5, 2, color.RGBA{100, 200, 100, 255}, false)
			}

			// Lobby info
			lobbyText := fmt.Sprintf("%s  [%d/%d players]  %s", lobby.Name, lobby.PlayerCount, lobby.MaxPlayers, lobby.State)
			ebitenutil.DebugPrintAt(screen, lobbyText, int(panelX)+30, int(itemY)+12)
		}
	}

	// Buttons
	buttonY := panelY + panelHeight - 50
	buttonWidth := 100.0
	buttonHeight := 35.0
	buttonSpacing := 20.0
	buttonColor := color.RGBA{60, 80, 100, 255}
	buttonHoverColor := color.RGBA{80, 120, 140, 255}

	// Back button
	vector.FillRect(screen, float32(panelX)+20, float32(buttonY), float32(buttonWidth), float32(buttonHeight), buttonColor, false)
	vector.StrokeRect(screen, float32(panelX)+20, float32(buttonY), float32(buttonWidth), float32(buttonHeight), 1, borderColor, false)
	ebitenutil.DebugPrintAt(screen, "Back", int(panelX)+20+int(buttonWidth)/2-12, int(buttonY)+10)

	// Refresh button
	vector.FillRect(screen, float32(panelX)+20+float32(buttonWidth)+float32(buttonSpacing), float32(buttonY), float32(buttonWidth), float32(buttonHeight), buttonColor, false)
	vector.StrokeRect(screen, float32(panelX)+20+float32(buttonWidth)+float32(buttonSpacing), float32(buttonY), float32(buttonWidth), float32(buttonHeight), 1, borderColor, false)
	ebitenutil.DebugPrintAt(screen, "Refresh", int(panelX)+20+int(buttonWidth)+int(buttonSpacing)+int(buttonWidth)/2-21, int(buttonY)+10)

	// Create button
	createX := panelX + panelWidth - buttonWidth - 20 - buttonWidth - buttonSpacing
	vector.FillRect(screen, float32(createX), float32(buttonY), float32(buttonWidth), float32(buttonHeight), buttonHoverColor, false)
	vector.StrokeRect(screen, float32(createX), float32(buttonY), float32(buttonWidth), float32(buttonHeight), 1, borderColor, false)
	ebitenutil.DebugPrintAt(screen, "Create", int(createX)+int(buttonWidth)/2-18, int(buttonY)+10)

	// Join button
	joinX := panelX + panelWidth - buttonWidth - 20
	joinColor := buttonColor
	if lb.selectedIndex >= 0 {
		joinColor = buttonHoverColor
	}
	vector.FillRect(screen, float32(joinX), float32(buttonY), float32(buttonWidth), float32(buttonHeight), joinColor, false)
	vector.StrokeRect(screen, float32(joinX), float32(buttonY), float32(buttonWidth), float32(buttonHeight), 1, borderColor, false)
	ebitenutil.DebugPrintAt(screen, "Join", int(joinX)+int(buttonWidth)/2-12, int(buttonY)+10)

	// Instructions
	instructions := "UP/DOWN: Select | ENTER: Join | ESC: Back"
	instrX := int(lb.screenWidth/2) - len(instructions)*3
	ebitenutil.DebugPrintAt(screen, instructions, instrX, int(lb.screenHeight)-30)
}
