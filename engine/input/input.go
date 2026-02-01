package input

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type State struct {
	MousePos         emath.Vec2
	LeftPressed      bool
	LeftJustPressed  bool
	LeftJustReleased bool
	RightPressed     bool
	RightJustPressed bool
	ShiftHeld        bool
	EscapePressed    bool
	ScrollUp         bool
	ScrollDown       bool
	ScrollLeft       bool
	ScrollRight      bool
	BuildTankPressed bool // T key to build tank
	MenuUp           bool // Up arrow only (not W, for menu)
	MenuDown         bool // Down arrow only (not S, for menu)
	EnterPressed     bool // Enter/Return key
	BackspacePressed bool // Backspace key for text input
	IsDragging       bool
	DragStart        emath.Vec2
	DragEnd          emath.Vec2
	MouseWheelY      float64 // Mouse wheel vertical scroll (positive = up/zoom in)
}
type Manager struct {
	state         State
	dragStarted   bool
	dragThreshold float64
}

func NewManager() *Manager {
	return &Manager{
		dragThreshold: 5.0, // pixels
	}
}
func (m *Manager) Update() {
	x, y := ebiten.CursorPosition()
	m.state.MousePos = emath.NewVec2(float64(x), float64(y))
	m.state.LeftPressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	m.state.LeftJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	m.state.LeftJustReleased = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	m.state.RightPressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	m.state.RightJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
	m.state.ShiftHeld = ebiten.IsKeyPressed(ebiten.KeyShift)
	m.state.EscapePressed = inpututil.IsKeyJustPressed(ebiten.KeyEscape)
	m.state.ScrollUp = ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
	m.state.ScrollDown = ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)
	m.state.ScrollLeft = ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	m.state.ScrollRight = ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)
	m.state.BuildTankPressed = inpututil.IsKeyJustPressed(ebiten.KeyT)
	m.state.MenuUp = inpututil.IsKeyJustPressed(ebiten.KeyUp)
	m.state.MenuDown = inpututil.IsKeyJustPressed(ebiten.KeyDown)
	m.state.EnterPressed = inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	m.state.BackspacePressed = inpututil.IsKeyJustPressed(ebiten.KeyBackspace)
	// Mouse wheel for zoom
	_, wheelY := ebiten.Wheel()
	m.state.MouseWheelY = wheelY
	if m.state.LeftJustPressed {
		m.state.DragStart = m.state.MousePos
		m.dragStarted = true
		m.state.IsDragging = false
	}
	if m.dragStarted && m.state.LeftPressed {
		dist := m.state.MousePos.Distance(m.state.DragStart)
		if dist > m.dragThreshold {
			m.state.IsDragging = true
		}
		m.state.DragEnd = m.state.MousePos
	}
	if m.state.LeftJustReleased {
		m.state.DragEnd = m.state.MousePos
		m.dragStarted = false
	}
}
func (m *Manager) State() State {
	return m.state
}
func (m *Manager) GetSelectionBox() emath.Rect {
	x1, y1 := m.state.DragStart.X, m.state.DragStart.Y
	x2, y2 := m.state.DragEnd.X, m.state.DragEnd.Y
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	return emath.NewRect(x1, y1, x2-x1, y2-y1)
}
func (m *Manager) ResetDrag() {
	m.state.IsDragging = false
	m.dragStarted = false
}
