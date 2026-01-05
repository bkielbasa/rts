package ui

import (
	"image/color"

	"github.com/bklimczak/tanks/engine/camera"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/terrain"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// MinimapEntity represents something to draw on the minimap
type MinimapEntity struct {
	Position emath.Vec2
	Size     emath.Vec2
	Color    color.Color
}

// Minimap shows a bird's eye view of the entire map
type Minimap struct {
	bounds      emath.Rect
	worldSize   emath.Vec2
	borderColor color.Color
	bgColor     color.Color
}

// NewMinimap creates a new minimap
func NewMinimap(x, y, width, height float64) *Minimap {
	return &Minimap{
		bounds:      emath.NewRect(x, y, width, height),
		borderColor: color.RGBA{80, 80, 100, 255},
		bgColor:     color.RGBA{15, 15, 20, 255},
	}
}

// SetPosition updates the minimap position
func (m *Minimap) SetPosition(x, y float64) {
	m.bounds.Pos.X = x
	m.bounds.Pos.Y = y
}

// SetWorldSize sets the world dimensions for scale calculations
func (m *Minimap) SetWorldSize(width, height float64) {
	m.worldSize = emath.Vec2{X: width, Y: height}
}

// Bounds returns the minimap bounds
func (m *Minimap) Bounds() emath.Rect {
	return m.bounds
}

// Contains checks if a screen point is inside the minimap
func (m *Minimap) Contains(p emath.Vec2) bool {
	return m.bounds.Contains(p)
}

// ScreenToWorld converts a minimap screen position to world coordinates
func (m *Minimap) ScreenToWorld(screenPos emath.Vec2) emath.Vec2 {
	// Get relative position within minimap
	relX := (screenPos.X - m.bounds.Pos.X) / m.bounds.Size.X
	relY := (screenPos.Y - m.bounds.Pos.Y) / m.bounds.Size.Y

	// Convert to world coordinates
	return emath.Vec2{
		X: relX * m.worldSize.X,
		Y: relY * m.worldSize.Y,
	}
}

// worldToMinimap converts world coordinates to minimap screen position
func (m *Minimap) worldToMinimap(worldPos emath.Vec2) emath.Vec2 {
	scaleX := m.bounds.Size.X / m.worldSize.X
	scaleY := m.bounds.Size.Y / m.worldSize.Y

	return emath.Vec2{
		X: m.bounds.Pos.X + worldPos.X*scaleX,
		Y: m.bounds.Pos.Y + worldPos.Y*scaleY,
	}
}

// worldSizeToMinimap converts world size to minimap size
func (m *Minimap) worldSizeToMinimap(worldSize emath.Vec2) emath.Vec2 {
	scaleX := m.bounds.Size.X / m.worldSize.X
	scaleY := m.bounds.Size.Y / m.worldSize.Y

	return emath.Vec2{
		X: worldSize.X * scaleX,
		Y: worldSize.Y * scaleY,
	}
}

// Draw renders the minimap
func (m *Minimap) Draw(screen *ebiten.Image, cam *camera.Camera, terrainMap *terrain.Map, entities []MinimapEntity) {
	// Draw terrain as background
	if terrainMap != nil {
		m.drawTerrain(screen, terrainMap)
	} else {
		// Fallback to solid background
		vector.FillRect(
			screen,
			float32(m.bounds.Pos.X),
			float32(m.bounds.Pos.Y),
			float32(m.bounds.Size.X),
			float32(m.bounds.Size.Y),
			m.bgColor,
			false,
		)
	}

	// Draw entities
	for _, ent := range entities {
		pos := m.worldToMinimap(ent.Position)
		size := m.worldSizeToMinimap(ent.Size)

		// Ensure minimum visibility
		if size.X < 2 {
			size.X = 2
		}
		if size.Y < 2 {
			size.Y = 2
		}

		vector.FillRect(
			screen,
			float32(pos.X),
			float32(pos.Y),
			float32(size.X),
			float32(size.Y),
			ent.Color,
			false,
		)
	}

	// Draw camera viewport rectangle
	viewportPos := m.worldToMinimap(cam.Position)
	viewportSize := m.worldSizeToMinimap(cam.ViewportSize)

	vector.StrokeRect(
		screen,
		float32(viewportPos.X),
		float32(viewportPos.Y),
		float32(viewportSize.X),
		float32(viewportSize.Y),
		1,
		color.RGBA{255, 255, 255, 200},
		false,
	)

	// Draw border
	vector.StrokeRect(
		screen,
		float32(m.bounds.Pos.X),
		float32(m.bounds.Pos.Y),
		float32(m.bounds.Size.X),
		float32(m.bounds.Size.Y),
		2,
		m.borderColor,
		false,
	)
}

// drawTerrain renders the terrain on the minimap
func (m *Minimap) drawTerrain(screen *ebiten.Image, terrainMap *terrain.Map) {
	scaleX := m.bounds.Size.X / m.worldSize.X
	scaleY := m.bounds.Size.Y / m.worldSize.Y
	tileScreenWidth := terrain.TileSize * scaleX
	tileScreenHeight := terrain.TileSize * scaleY

	// Ensure minimum tile size for visibility
	if tileScreenWidth < 1 {
		tileScreenWidth = 1
	}
	if tileScreenHeight < 1 {
		tileScreenHeight = 1
	}

	for y := 0; y < terrainMap.Height; y++ {
		for x := 0; x < terrainMap.Width; x++ {
			tile := terrainMap.Tiles[y][x]
			tileColor := terrain.TileColors(tile.Type)

			screenX := m.bounds.Pos.X + float64(x)*tileScreenWidth
			screenY := m.bounds.Pos.Y + float64(y)*tileScreenHeight

			vector.FillRect(
				screen,
				float32(screenX),
				float32(screenY),
				float32(tileScreenWidth)+1, // +1 to avoid gaps
				float32(tileScreenHeight)+1,
				tileColor,
				false,
			)
		}
	}
}
