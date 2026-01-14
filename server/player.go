package server

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Player represents a connected player session
type Player struct {
	ID        string
	Name      string
	Conn      *websocket.Conn
	Slot      int // Player slot 0-3 in game (assigned when game starts)
	Ready     bool
	Connected bool
	Alive     bool // In-game status

	sendChan  chan Message
	closeChan chan struct{}
	closeOnce sync.Once

	mu sync.RWMutex
}

// NewPlayer creates a new player session
func NewPlayer(id string, conn *websocket.Conn) *Player {
	p := &Player{
		ID:        id,
		Name:      "Player",
		Conn:      conn,
		Connected: true,
		Alive:     true,
		sendChan:  make(chan Message, 64),
		closeChan: make(chan struct{}),
	}

	go p.writePump()

	return p
}

// SetName sets the player's display name
func (p *Player) SetName(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if name != "" {
		p.Name = name
	}
}

// GetName returns the player's display name
func (p *Player) GetName() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Name
}

// SetReady sets the player's ready status
func (p *Player) SetReady(ready bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Ready = ready
}

// IsReady returns whether the player is ready
func (p *Player) IsReady() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Ready
}

// Send queues a message to be sent to the player
func (p *Player) Send(msg Message) error {
	select {
	case p.sendChan <- msg:
		return nil
	case <-p.closeChan:
		return websocket.ErrCloseSent
	default:
		// Channel full, drop message
		log.Printf("Warning: message dropped for player %s (channel full)", p.ID)
		return nil
	}
}

// SendPayload sends a message with the given type and payload
func (p *Player) SendPayload(msgType MessageType, payload interface{}) error {
	msg, err := NewMessage(msgType, payload)
	if err != nil {
		return err
	}
	return p.Send(msg)
}

// SendError sends an error message to the player
func (p *Player) SendError(message string) error {
	return p.SendPayload(MsgError, ErrorPayload{Message: message})
}

// Close closes the player's connection
func (p *Player) Close() {
	p.closeOnce.Do(func() {
		close(p.closeChan)
		p.mu.Lock()
		p.Connected = false
		p.mu.Unlock()
		if p.Conn != nil {
			p.Conn.Close()
		}
	})
}

// IsConnected returns whether the player is still connected
func (p *Player) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Connected
}

// writePump handles sending messages to the WebSocket
func (p *Player) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		p.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-p.sendChan:
			if !ok {
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			p.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			if err := p.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				return
			}

		case <-ticker.C:
			p.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-p.closeChan:
			return
		}
	}
}

// ReadMessage reads the next message from the player's WebSocket
func (p *Player) ReadMessage() (Message, error) {
	_, data, err := p.Conn.ReadMessage()
	if err != nil {
		return Message{}, err
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return Message{}, err
	}

	return msg, nil
}

// ToPlayerInfo converts the player to a PlayerInfo for sending to clients
func (p *Player) ToPlayerInfo() PlayerInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return PlayerInfo{
		ID:      p.ID,
		Name:    p.Name,
		Ready:   p.Ready,
		Faction: p.Slot,
		Alive:   p.Alive,
	}
}
