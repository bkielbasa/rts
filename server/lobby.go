package server

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// LobbyState represents the current state of a lobby
type LobbyState string

const (
	LobbyWaiting  LobbyState = "waiting"
	LobbyPlaying  LobbyState = "playing"
	LobbyFinished LobbyState = "finished"
)

const (
	MinPlayers = 2
	MaxPlayers = 4
)

// Lobby represents a game lobby
type Lobby struct {
	ID          string
	Name        string
	State       LobbyState
	HostID      string
	Players     map[string]*Player // PlayerID -> Player
	PlayerOrder []string           // Ordered list of player IDs (for slot assignment)
	MaxPlayers  int

	Game       *Simulation
	gameCancel context.CancelFunc

	mu sync.RWMutex
}

// NewLobby creates a new lobby
func NewLobby(name string, host *Player, maxPlayers int) *Lobby {
	if maxPlayers < MinPlayers {
		maxPlayers = MinPlayers
	}
	if maxPlayers > MaxPlayers {
		maxPlayers = MaxPlayers
	}

	lobby := &Lobby{
		ID:          uuid.New().String()[:8],
		Name:        name,
		State:       LobbyWaiting,
		HostID:      host.ID,
		Players:     make(map[string]*Player),
		PlayerOrder: make([]string, 0, maxPlayers),
		MaxPlayers:  maxPlayers,
	}

	lobby.Players[host.ID] = host
	lobby.PlayerOrder = append(lobby.PlayerOrder, host.ID)

	return lobby
}

// AddPlayer adds a player to the lobby
func (l *Lobby) AddPlayer(p *Player) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != LobbyWaiting {
		return errors.New("lobby is not accepting players")
	}

	if len(l.Players) >= l.MaxPlayers {
		return errors.New("lobby is full")
	}

	if _, exists := l.Players[p.ID]; exists {
		return errors.New("player already in lobby")
	}

	l.Players[p.ID] = p
	l.PlayerOrder = append(l.PlayerOrder, p.ID)
	p.Ready = false

	return nil
}

// RemovePlayer removes a player from the lobby
func (l *Lobby) RemovePlayer(playerID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.Players[playerID]; !exists {
		return errors.New("player not in lobby")
	}

	delete(l.Players, playerID)

	// Remove from order
	for i, id := range l.PlayerOrder {
		if id == playerID {
			l.PlayerOrder = append(l.PlayerOrder[:i], l.PlayerOrder[i+1:]...)
			break
		}
	}

	// If host left, assign new host
	if l.HostID == playerID && len(l.PlayerOrder) > 0 {
		l.HostID = l.PlayerOrder[0]
	}

	return nil
}

// SetPlayerReady sets a player's ready status
func (l *Lobby) SetPlayerReady(playerID string, ready bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	player, exists := l.Players[playerID]
	if !exists {
		return errors.New("player not in lobby")
	}

	player.SetReady(ready)
	return nil
}

// CanStart returns true if the game can be started
func (l *Lobby) CanStart() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.State != LobbyWaiting {
		return false
	}

	if len(l.Players) < MinPlayers {
		return false
	}

	// All players must be ready
	for _, p := range l.Players {
		if !p.IsReady() {
			return false
		}
	}

	return true
}

// Start starts the game
func (l *Lobby) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != LobbyWaiting {
		return errors.New("lobby is not in waiting state")
	}

	if len(l.Players) < MinPlayers {
		return errors.New("not enough players")
	}

	// Assign slots to players
	for i, playerID := range l.PlayerOrder {
		if player, ok := l.Players[playerID]; ok {
			player.Slot = i
			player.Alive = true
		}
	}

	// Create game simulation
	playerSetups := make([]PlayerSetup, 0, len(l.Players))
	for _, playerID := range l.PlayerOrder {
		if player, ok := l.Players[playerID]; ok {
			playerSetups = append(playerSetups, PlayerSetup{
				PlayerID: player.ID,
				Name:     player.GetName(),
				Slot:     player.Slot,
			})
		}
	}

	l.Game = NewSimulation(playerSetups)
	l.State = LobbyPlaying

	// Start game loop in background
	ctx, cancel := context.WithCancel(context.Background())
	l.gameCancel = cancel
	go l.Game.Run(ctx, l)

	return nil
}

// Stop stops the game
func (l *Lobby) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.gameCancel != nil {
		l.gameCancel()
		l.gameCancel = nil
	}
	l.Game = nil
	l.State = LobbyFinished
}

// Broadcast sends a message to all players in the lobby
func (l *Lobby) Broadcast(msg Message) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, p := range l.Players {
		p.Send(msg)
	}
}

// BroadcastPayload sends a message with the given type and payload to all players
func (l *Lobby) BroadcastPayload(msgType MessageType, payload interface{}) {
	msg, err := NewMessage(msgType, payload)
	if err != nil {
		return
	}
	l.Broadcast(msg)
}

// GetPlayerSlot returns the slot number for a player
func (l *Lobby) GetPlayerSlot(playerID string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if player, ok := l.Players[playerID]; ok {
		return player.Slot
	}
	return -1
}

// ToLobbyInfo converts the lobby to a LobbyInfo for sending to clients
func (l *Lobby) ToLobbyInfo() LobbyInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	players := make([]PlayerInfo, 0, len(l.Players))
	for _, p := range l.Players {
		players = append(players, p.ToPlayerInfo())
	}

	hostName := ""
	if host, ok := l.Players[l.HostID]; ok {
		hostName = host.GetName()
	}

	return LobbyInfo{
		ID:         l.ID,
		Name:       l.Name,
		HostID:     l.HostID,
		HostName:   hostName,
		Players:    players,
		MaxPlayers: l.MaxPlayers,
		State:      string(l.State),
	}
}

// IsEmpty returns true if the lobby has no players
func (l *Lobby) IsEmpty() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.Players) == 0
}

// LobbyManager manages all active lobbies
type LobbyManager struct {
	lobbies   map[string]*Lobby // LobbyID -> Lobby
	playerMap map[string]string // PlayerID -> LobbyID

	mu sync.RWMutex
}

// NewLobbyManager creates a new lobby manager
func NewLobbyManager() *LobbyManager {
	return &LobbyManager{
		lobbies:   make(map[string]*Lobby),
		playerMap: make(map[string]string),
	}
}

// CreateLobby creates a new lobby with the given host
func (m *LobbyManager) CreateLobby(host *Player, name string, maxPlayers int) (*Lobby, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if player is already in a lobby
	if _, inLobby := m.playerMap[host.ID]; inLobby {
		return nil, errors.New("player is already in a lobby")
	}

	lobby := NewLobby(name, host, maxPlayers)
	m.lobbies[lobby.ID] = lobby
	m.playerMap[host.ID] = lobby.ID

	return lobby, nil
}

// JoinLobby adds a player to an existing lobby
func (m *LobbyManager) JoinLobby(player *Player, lobbyID string) (*Lobby, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if player is already in a lobby
	if _, inLobby := m.playerMap[player.ID]; inLobby {
		return nil, errors.New("player is already in a lobby")
	}

	lobby, exists := m.lobbies[lobbyID]
	if !exists {
		return nil, errors.New("lobby not found")
	}

	if err := lobby.AddPlayer(player); err != nil {
		return nil, err
	}

	m.playerMap[player.ID] = lobbyID
	return lobby, nil
}

// LeaveLobby removes a player from their current lobby
func (m *LobbyManager) LeaveLobby(playerID string) (*Lobby, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lobbyID, inLobby := m.playerMap[playerID]
	if !inLobby {
		return nil, errors.New("player is not in a lobby")
	}

	lobby, exists := m.lobbies[lobbyID]
	if !exists {
		delete(m.playerMap, playerID)
		return nil, errors.New("lobby not found")
	}

	if err := lobby.RemovePlayer(playerID); err != nil {
		return nil, err
	}

	delete(m.playerMap, playerID)

	// Clean up empty lobbies
	if lobby.IsEmpty() {
		lobby.Stop()
		delete(m.lobbies, lobbyID)
		return nil, nil
	}

	return lobby, nil
}

// GetLobby returns a lobby by ID
func (m *LobbyManager) GetLobby(lobbyID string) (*Lobby, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lobby, exists := m.lobbies[lobbyID]
	return lobby, exists
}

// GetPlayerLobby returns the lobby a player is in
func (m *LobbyManager) GetPlayerLobby(playerID string) (*Lobby, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lobbyID, inLobby := m.playerMap[playerID]
	if !inLobby {
		return nil, false
	}

	lobby, exists := m.lobbies[lobbyID]
	return lobby, exists
}

// ListLobbies returns all lobbies that are waiting for players
func (m *LobbyManager) ListLobbies() []LobbyInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lobbies := make([]LobbyInfo, 0)
	for _, lobby := range m.lobbies {
		if lobby.State == LobbyWaiting {
			lobbies = append(lobbies, lobby.ToLobbyInfo())
		}
	}

	return lobbies
}

// CleanupFinished removes finished lobbies
func (m *LobbyManager) CleanupFinished() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, lobby := range m.lobbies {
		if lobby.State == LobbyFinished && lobby.IsEmpty() {
			delete(m.lobbies, id)
		}
	}
}

// StopAll stops all active lobbies and games
func (m *LobbyManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, lobby := range m.lobbies {
		lobby.Stop()
	}
	m.lobbies = make(map[string]*Lobby)
	m.playerMap = make(map[string]string)
}
