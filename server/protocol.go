package server

import (
	"encoding/json"
)

// MessageType identifies the type of WebSocket message
type MessageType string

const (
	// Client -> Server messages
	MsgSetName     MessageType = "set_name"
	MsgCreateLobby MessageType = "create_lobby"
	MsgJoinLobby   MessageType = "join_lobby"
	MsgLeaveLobby  MessageType = "leave_lobby"
	MsgListLobbies MessageType = "list_lobbies"
	MsgSetReady    MessageType = "set_ready"
	MsgStartGame   MessageType = "start_game"
	MsgGameCommand MessageType = "game_command"

	// Server -> Client messages
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

// Message is the envelope for all WebSocket communication
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewMessage creates a message with the given type and payload
func NewMessage(msgType MessageType, payload interface{}) (Message, error) {
	var raw json.RawMessage
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return Message{}, err
		}
		raw = data
	}
	return Message{Type: msgType, Payload: raw}, nil
}

// CommandType identifies game commands
type CommandType string

const (
	CmdMove             CommandType = "move"
	CmdAttack           CommandType = "attack"
	CmdAttackMove       CommandType = "attack_move"
	CmdStop             CommandType = "stop"
	CmdPlaceBuilding    CommandType = "place_building"
	CmdProduceUnit      CommandType = "produce_unit"
	CmdCancelProduction CommandType = "cancel_production"
	CmdSetRallyPoint    CommandType = "set_rally"
)

// GameCommand represents a player action in the game
type GameCommand struct {
	Type         CommandType `json:"type"`
	UnitIDs      []uint64    `json:"unitIds,omitempty"`
	BuildingID   uint64      `json:"buildingId,omitempty"`
	TargetX      float64     `json:"x,omitempty"`
	TargetY      float64     `json:"y,omitempty"`
	TargetID     uint64      `json:"targetId,omitempty"`
	BuildingType int         `json:"buildingType,omitempty"`
	UnitType     int         `json:"unitType,omitempty"`
}

// Client -> Server payloads

type SetNamePayload struct {
	Name string `json:"name"`
}

type CreateLobbyPayload struct {
	Name       string `json:"name"`
	MaxPlayers int    `json:"maxPlayers"`
}

type JoinLobbyPayload struct {
	LobbyID string `json:"lobbyId"`
}

type SetReadyPayload struct {
	Ready bool `json:"ready"`
}

type GameCommandPayload struct {
	Command GameCommand `json:"command"`
}

// Server -> Client payloads

type WelcomePayload struct {
	PlayerID string `json:"playerId"`
}

type LobbyInfo struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	HostID     string       `json:"hostId"`
	HostName   string       `json:"hostName"`
	Players    []PlayerInfo `json:"players"`
	MaxPlayers int          `json:"maxPlayers"`
	State      string       `json:"state"` // "waiting", "playing", "finished"
}

type PlayerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Ready   bool   `json:"ready"`
	Faction int    `json:"faction"`
	Alive   bool   `json:"alive"`
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
	YourSlot int       `json:"yourSlot"` // 0-3, determines spawn position
}

type ErrorPayload struct {
	Message string `json:"message"`
}

// Game state payloads

type UnitState struct {
	ID          uint64  `json:"id"`
	Type        int     `json:"type"`
	OwnerSlot   int     `json:"owner"` // Player slot 0-3
	PosX        float64 `json:"x"`
	PosY        float64 `json:"y"`
	Health      float64 `json:"hp"`
	MaxHealth   float64 `json:"maxHp"`
	Angle       float64 `json:"angle"`
	TurretAngle float64 `json:"turretAngle"`
	HasTarget   bool    `json:"hasTarget,omitempty"`
	TargetX     float64 `json:"tx,omitempty"`
	TargetY     float64 `json:"ty,omitempty"`
}

type BuildingState struct {
	ID            uint64  `json:"id"`
	Type          int     `json:"type"`
	OwnerSlot     int     `json:"owner"` // Player slot 0-3
	PosX          float64 `json:"x"`
	PosY          float64 `json:"y"`
	Health        float64 `json:"hp"`
	MaxHealth     float64 `json:"maxHp"`
	Completed     bool    `json:"done"`
	BuildProgress float64 `json:"progress,omitempty"`
	Producing     bool    `json:"producing,omitempty"`
	ProdProgress  float64 `json:"prodProgress,omitempty"`
	ProdType      int     `json:"prodType,omitempty"`
}

type ProjectileState struct {
	ID        uint64  `json:"id"`
	OwnerSlot int     `json:"owner"`
	PosX      float64 `json:"x"`
	PosY      float64 `json:"y"`
	TargetX   float64 `json:"tx"`
	TargetY   float64 `json:"ty"`
}

type ResourceStateNet struct {
	Metal      float64 `json:"metal"`
	MetalCap   float64 `json:"metalCap"`
	MetalProd  float64 `json:"metalProd"`
	Energy     float64 `json:"energy"`
	EnergyCap  float64 `json:"energyCap"`
	EnergyProd float64 `json:"energyProd"`
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
	Reason     string `json:"reason"` // "last_standing", "surrender"
}
