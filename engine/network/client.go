package network

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	// Client -> Server
	MsgSetName     MessageType = "set_name"
	MsgCreateLobby MessageType = "create_lobby"
	MsgJoinLobby   MessageType = "join_lobby"
	MsgLeaveLobby  MessageType = "leave_lobby"
	MsgListLobbies MessageType = "list_lobbies"
	MsgSetReady    MessageType = "set_ready"
	MsgStartGame   MessageType = "start_game"
	MsgGameCommand MessageType = "game_command"

	// Server -> Client
	MsgWelcome      MessageType = "welcome"
	MsgLobbyList    MessageType = "lobby_list"
	MsgLobbyCreated MessageType = "lobby_created"
	MsgLobbyJoined  MessageType = "lobby_joined"
	MsgLobbyLeft    MessageType = "lobby_left"
	MsgLobbyUpdate  MessageType = "lobby_update"
	MsgGameStarting MessageType = "game_starting"
	MsgGameState    MessageType = "game_state"
	MsgGameEnd      MessageType = "game_end"
	MsgError        MessageType = "error"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type PlayerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Ready   bool   `json:"ready"`
	Faction int    `json:"faction"`
	Alive   bool   `json:"alive"`
}

type LobbyInfo struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	HostID     string       `json:"hostId"`
	HostName   string       `json:"hostName"`
	Players    []PlayerInfo `json:"players"`
	MaxPlayers int          `json:"maxPlayers"`
	State      string       `json:"state"`
}

type LobbyListPayload struct {
	Lobbies []LobbyInfo `json:"lobbies"`
}

type LobbyCreatedPayload struct {
	Lobby LobbyInfo `json:"lobby"`
}

type LobbyJoinedPayload struct {
	Lobby LobbyInfo `json:"lobby"`
}

type LobbyUpdatePayload struct {
	Lobby LobbyInfo `json:"lobby"`
}

type GameStartingPayload struct {
	Lobby    LobbyInfo `json:"lobby"`
	YourSlot int       `json:"yourSlot"`
}

type WelcomePayload struct {
	PlayerID string `json:"playerId"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type UnitState struct {
	ID        uint64  `json:"id"`
	Type      int     `json:"type"`
	OwnerSlot int     `json:"owner"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Health    float64 `json:"hp"`
	MaxHealth float64 `json:"maxHp"`
	Angle     float64 `json:"angle"`
}

type BuildingState struct {
	ID            uint64  `json:"id"`
	Type          int     `json:"type"`
	OwnerSlot     int     `json:"owner"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Health        float64 `json:"hp"`
	MaxHealth     float64 `json:"maxHp"`
	Completed     bool    `json:"done"`
	BuildProgress float64 `json:"progress"`
}

type ProjectileState struct {
	ID        uint64  `json:"id"`
	OwnerSlot int     `json:"owner"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}

type ResourceStateNet struct {
	Metal     float64 `json:"metal"`
	MetalCap  float64 `json:"metalCap"`
	Energy    float64 `json:"energy"`
	EnergyCap float64 `json:"energyCap"`
}

type PlayerGameState struct {
	Slot      int              `json:"slot"`
	Name      string           `json:"name"`
	Alive     bool             `json:"alive"`
	Resources ResourceStateNet `json:"resources"`
}

type GameStatePayload struct {
	Tick        uint64            `json:"tick"`
	Players     []PlayerGameState `json:"players"`
	Units       []UnitState       `json:"units"`
	Buildings   []BuildingState   `json:"buildings"`
	Projectiles []ProjectileState `json:"projectiles"`
}

type GameEndPayload struct {
	WinnerSlot int    `json:"winnerSlot"`
	WinnerName string `json:"winnerName"`
}

type Client struct {
	conn         *websocket.Conn
	serverAddr   string
	connected    bool
	playerName   string
	playerID     string
	currentLobby *LobbyInfo

	mu          sync.RWMutex
	lobbies     []LobbyInfo
	gameState   *GameStatePayload
	gameStarted bool
	gameEnded   bool
	gameEndInfo *GameEndPayload
	lastError   string
	yourSlot    int
	isHost      bool
}

func NewClient(playerName string) *Client {
	return &Client{
		playerName: playerName,
		yourSlot:   -1,
	}
}

func (c *Client) Connect(serverAddr string) error {
	url := fmt.Sprintf("ws://%s/ws", serverAddr)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.serverAddr = serverAddr
	c.connected = true

	go c.readLoop()

	// Send player name
	c.send(Message{Type: MsgSetName, Payload: mustMarshal(map[string]string{"name": c.playerName})})

	return nil
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func (c *Client) Disconnect() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
	c.currentLobby = nil
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) readLoop() {
	defer func() {
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
		}
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case MsgWelcome:
		var payload WelcomePayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.playerID = payload.PlayerID
			c.mu.Unlock()
			log.Printf("Connected as player: %s", payload.PlayerID)
		}

	case MsgLobbyList:
		var payload LobbyListPayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.lobbies = payload.Lobbies
			c.mu.Unlock()
		}

	case MsgLobbyCreated:
		var payload LobbyCreatedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.currentLobby = &payload.Lobby
			c.isHost = true
			c.yourSlot = 0
			c.mu.Unlock()
			log.Printf("Created lobby: %s", payload.Lobby.Name)
		}

	case MsgLobbyJoined:
		var payload LobbyJoinedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.currentLobby = &payload.Lobby
			c.isHost = false
			// Find our slot
			for i, p := range payload.Lobby.Players {
				if p.ID == c.playerID {
					c.yourSlot = i
					break
				}
			}
			c.mu.Unlock()
			log.Printf("Joined lobby: %s", payload.Lobby.Name)
		}

	case MsgLobbyUpdate:
		var payload LobbyUpdatePayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.currentLobby = &payload.Lobby
			// Update our slot
			for i, p := range payload.Lobby.Players {
				if p.ID == c.playerID {
					c.yourSlot = i
					break
				}
			}
			c.mu.Unlock()
		}

	case MsgLobbyLeft:
		c.mu.Lock()
		c.currentLobby = nil
		c.isHost = false
		c.yourSlot = -1
		c.mu.Unlock()

	case MsgGameStarting:
		var payload GameStartingPayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.gameStarted = true
			c.yourSlot = payload.YourSlot
			c.mu.Unlock()
			log.Printf("Game starting! Your slot: %d", payload.YourSlot)
		}

	case MsgGameState:
		var payload GameStatePayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.gameState = &payload
			c.mu.Unlock()
		}

	case MsgGameEnd:
		var payload GameEndPayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.gameEnded = true
			c.gameEndInfo = &payload
			c.mu.Unlock()
		}

	case MsgError:
		var payload ErrorPayload
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.mu.Lock()
			c.lastError = payload.Message
			c.mu.Unlock()
			log.Printf("Server error: %s", payload.Message)
		}
	}
}

func (c *Client) send(msg Message) error {
	if !c.connected || c.conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) RequestLobbyList() error {
	return c.send(Message{Type: MsgListLobbies})
}

func (c *Client) CreateLobby(name string, maxPlayers int) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"name":       name,
		"maxPlayers": maxPlayers,
	})
	return c.send(Message{Type: MsgCreateLobby, Payload: payload})
}

func (c *Client) JoinLobby(lobbyID string) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"lobbyId": lobbyID,
	})
	return c.send(Message{Type: MsgJoinLobby, Payload: payload})
}

func (c *Client) LeaveLobby() error {
	err := c.send(Message{Type: MsgLeaveLobby})
	return err
}

func (c *Client) SetReady(ready bool) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"ready": ready,
	})
	return c.send(Message{Type: MsgSetReady, Payload: payload})
}

func (c *Client) StartGame() error {
	return c.send(Message{Type: MsgStartGame})
}

func (c *Client) SendCommand(command string, data interface{}) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"command": map[string]interface{}{
			"type": command,
		},
	})
	return c.send(Message{Type: MsgGameCommand, Payload: payload})
}

func (c *Client) SendMoveCommand(unitIDs []uint64, x, y float64) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"command": map[string]interface{}{
			"type":    "move",
			"unitIds": unitIDs,
			"x":       x,
			"y":       y,
		},
	})
	return c.send(Message{Type: MsgGameCommand, Payload: payload})
}

func (c *Client) SendProduceUnitCommand(buildingID uint64, unitType int) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"command": map[string]interface{}{
			"type":       "produce_unit",
			"buildingId": buildingID,
			"unitType":   unitType,
		},
	})
	return c.send(Message{Type: MsgGameCommand, Payload: payload})
}

func (c *Client) SendCancelProductionCommand(buildingID uint64, unitType int) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"command": map[string]interface{}{
			"type":       "cancel_production",
			"buildingId": buildingID,
			"unitType":   unitType,
		},
	})
	return c.send(Message{Type: MsgGameCommand, Payload: payload})
}

func (c *Client) SendPlaceBuildingCommand(buildingType int, x, y float64) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"command": map[string]interface{}{
			"type":         "place_building",
			"buildingType": buildingType,
			"x":            x,
			"y":            y,
		},
	})
	return c.send(Message{Type: MsgGameCommand, Payload: payload})
}

func (c *Client) GetLobbies() []LobbyInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lobbies
}

func (c *Client) GetCurrentLobby() *LobbyInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentLobby
}

func (c *Client) GetGameState() *GameStatePayload {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gameState
}

func (c *Client) IsGameStarted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gameStarted
}

func (c *Client) IsGameEnded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gameEnded
}

func (c *Client) GetGameEndInfo() *GameEndPayload {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gameEndInfo
}

func (c *Client) GetLastError() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastError
}

func (c *Client) ClearError() {
	c.mu.Lock()
	c.lastError = ""
	c.mu.Unlock()
}

func (c *Client) GetYourSlot() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.yourSlot
}

func (c *Client) InLobby() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentLobby != nil
}

func (c *Client) IsHost() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isHost
}

func (c *Client) GetPlayerID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.playerID
}

func (c *Client) ResetGameState() {
	c.mu.Lock()
	c.gameStarted = false
	c.gameEnded = false
	c.gameState = nil
	c.gameEndInfo = nil
	c.currentLobby = nil
	c.mu.Unlock()
}
