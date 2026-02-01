package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Server is the main game server
type Server struct {
	lobbyManager *LobbyManager
	players      map[string]*Player // Connection ID -> Player
	httpServer   *http.Server

	mu sync.RWMutex
}

// New creates a new game server
func New() *Server {
	return &Server{
		lobbyManager: NewLobbyManager(),
		players:      make(map[string]*Player),
	}
}

// HandleWebSocket handles WebSocket connections
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create player
	playerID := uuid.New().String()[:8]
	player := NewPlayer(playerID, conn)

	s.mu.Lock()
	s.players[playerID] = player
	s.mu.Unlock()

	log.Printf("Player connected: %s", playerID)

	// Send welcome message
	player.SendPayload(MsgWelcome, WelcomePayload{PlayerID: playerID})

	// Handle messages
	go s.handlePlayer(player)
}

// handlePlayer handles messages from a player
func (s *Server) handlePlayer(player *Player) {
	defer func() {
		s.handleDisconnect(player)
	}()

	for {
		msg, err := player.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		s.handleMessage(player, msg)
	}
}

// handleMessage handles a single message from a player
func (s *Server) handleMessage(player *Player, msg Message) {
	switch msg.Type {
	case MsgSetName:
		var payload SetNamePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			player.SendError("Invalid payload")
			return
		}
		player.SetName(payload.Name)
		log.Printf("Player %s set name to: %s", player.ID, payload.Name)

	case MsgListLobbies:
		lobbies := s.lobbyManager.ListLobbies()
		player.SendPayload(MsgLobbyList, LobbyListPayload{Lobbies: lobbies})

	case MsgCreateLobby:
		var payload CreateLobbyPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			player.SendError("Invalid payload")
			return
		}

		lobby, err := s.lobbyManager.CreateLobby(player, payload.Name, payload.MaxPlayers)
		if err != nil {
			player.SendError(err.Error())
			return
		}

		log.Printf("Player %s created lobby: %s", player.ID, lobby.ID)
		player.SendPayload(MsgLobbyCreated, LobbyCreatedPayload{Lobby: lobby.ToLobbyInfo()})

	case MsgJoinLobby:
		var payload JoinLobbyPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			player.SendError("Invalid payload")
			return
		}

		lobby, err := s.lobbyManager.JoinLobby(player, payload.LobbyID)
		if err != nil {
			player.SendError(err.Error())
			return
		}

		log.Printf("Player %s joined lobby: %s", player.ID, lobby.ID)
		player.SendPayload(MsgLobbyJoined, LobbyJoinedPayload{Lobby: lobby.ToLobbyInfo()})

		// Notify other players
		lobby.BroadcastPayload(MsgLobbyUpdate, LobbyUpdatePayload{Lobby: lobby.ToLobbyInfo()})

	case MsgLeaveLobby:
		lobby, err := s.lobbyManager.LeaveLobby(player.ID)
		if err != nil {
			player.SendError(err.Error())
			return
		}

		log.Printf("Player %s left lobby", player.ID)
		player.SendPayload(MsgLobbyLeft, nil)

		// Notify remaining players
		if lobby != nil {
			lobby.BroadcastPayload(MsgLobbyUpdate, LobbyUpdatePayload{Lobby: lobby.ToLobbyInfo()})
		}

	case MsgSetReady:
		var payload SetReadyPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			player.SendError("Invalid payload")
			return
		}

		lobby, ok := s.lobbyManager.GetPlayerLobby(player.ID)
		if !ok {
			player.SendError("Not in a lobby")
			return
		}

		if err := lobby.SetPlayerReady(player.ID, payload.Ready); err != nil {
			player.SendError(err.Error())
			return
		}

		log.Printf("Player %s set ready: %v", player.ID, payload.Ready)

		// Notify all players
		lobby.BroadcastPayload(MsgLobbyUpdate, LobbyUpdatePayload{Lobby: lobby.ToLobbyInfo()})

	case MsgStartGame:
		lobby, ok := s.lobbyManager.GetPlayerLobby(player.ID)
		if !ok {
			player.SendError("Not in a lobby")
			return
		}

		if lobby.HostID != player.ID {
			player.SendError("Only the host can start the game")
			return
		}

		if !lobby.CanStart() {
			player.SendError("Cannot start game: not all players ready or not enough players")
			return
		}

		if err := lobby.Start(); err != nil {
			player.SendError(err.Error())
			return
		}

		log.Printf("Game started in lobby: %s", lobby.ID)

		// Notify all players with their slot
		lobbyInfo := lobby.ToLobbyInfo()
		for _, p := range lobby.Players {
			p.SendPayload(MsgGameStarting, GameStartingPayload{
				Lobby:    lobbyInfo,
				YourSlot: p.Slot,
			})
		}

	case MsgGameCommand:
		var payload GameCommandPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			player.SendError("Invalid payload")
			return
		}

		lobby, ok := s.lobbyManager.GetPlayerLobby(player.ID)
		if !ok || lobby.State != LobbyPlaying || lobby.Game == nil {
			return
		}

		// Enqueue command for processing
		lobby.Game.EnqueueCommand(player.ID, player.Slot, payload.Command)

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleDisconnect handles a player disconnecting
func (s *Server) handleDisconnect(player *Player) {
	s.mu.Lock()
	delete(s.players, player.ID)
	s.mu.Unlock()

	// Remove from lobby
	lobby, _ := s.lobbyManager.LeaveLobby(player.ID)
	if lobby != nil {
		lobby.BroadcastPayload(MsgLobbyUpdate, LobbyUpdatePayload{Lobby: lobby.ToLobbyInfo()})
	}

	player.Close()
	log.Printf("Player disconnected: %s", player.ID)
}

// HandleLobbies handles REST API for listing lobbies
func (s *Server) HandleLobbies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lobbies := s.lobbyManager.ListLobbies()
	json.NewEncoder(w).Encode(LobbyListPayload{Lobbies: lobbies})
}

// Start starts the server on the given address
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.HandleWebSocket)
	mux.HandleFunc("/api/lobbies", s.HandleLobbies)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Server starting on %s", addr)
	log.Printf("WebSocket endpoint: ws://%s/ws", addr)
	log.Printf("REST API endpoint: http://%s/api/lobbies", addr)

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Server shutting down...")

	// Close all player connections
	s.mu.Lock()
	for _, player := range s.players {
		player.Close()
	}
	s.players = make(map[string]*Player)
	s.mu.Unlock()

	// Stop all lobbies
	s.lobbyManager.StopAll()

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	log.Println("Server shutdown complete")
	return nil
}

// GracefulShutdown performs a graceful shutdown with a timeout
func (s *Server) GracefulShutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}
