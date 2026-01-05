package ui

import (
	"github.com/bklimczak/tanks/engine/camera"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/terrain"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

type MinimapEntity struct {
	Position emath.Vec2
	Size     emath.Vec2
	Color    color.Color
}
type Minimap struct {
	bounds      emath.Rect
	worldSize   emath.Vec2
	borderColor color.Color
	bgColor     color.Color
}

func NewMinimap(x, y, width, height float64) *Minimap {
	return &Minimap{
		bounds:      emath.NewRect(x, y, width, height),
		borderColor: color.RGBA{80, 80, 100, 255},
		bgColor:     color.RGBA{15, 15, 20, 255},
	}
}
func (m *Minimap) SetPosition(x, y float64) {
	m.bounds.Pos.X = x
	m.bounds.Pos.Y = y
}
func (m *Minimap) SetWorldSize(width, height float64) {
	m.worldSize = emath.Vec2{X: width, Y: height}
}
func (m *Minimap) Bounds() emath.Rect {
	return m.bounds
}
func (m *Minimap) Contains(p emath.Vec2) bool {
	return m.bounds.Contains(p)
}
func (m *Minimap) ScreenToWorld(screenPos emath.Vec2) emath.Vec2 {
	relX := (screenPos.X - m.bounds.Pos.X) / m.bounds.Size.X
	relY := (screenPos.Y - m.bounds.Pos.Y) / m.bounds.Size.Y
	return emath.Vec2{
		X: relX * m.worldSize.X,
		Y: relY * m.worldSize.Y,
	}
}
func (m *Minimap) worldToMinimap(worldPos emath.Vec2) emath.Vec2 {
	scaleX := m.bounds.Size.X / m.worldSize.X
	scaleY := m.bounds.Size.Y / m.worldSize.Y
	return emath.Vec2{
		X: m.bounds.Pos.X + worldPos.X*scaleX,
		Y: m.bounds.Pos.Y + worldPos.Y*scaleY,
	}
}
func (m *Minimap) worldSizeToMinimap(worldSize emath.Vec2) emath.Vec2 {
	scaleX := m.bounds.Size.X / m.worldSize.X
	scaleY := m.bounds.Size.Y / m.worldSize.Y
	return emath.Vec2{
		X: worldSize.X * scaleX,
		Y: worldSize.Y * scaleY,
	}
}
func (m *Minimap) Draw(screen *ebiten.Image, cam *camera.Camera, terrainMap *terrain.Map, entities []MinimapEntity) {
	if terrainMap != nil {
		m.drawTerrain(screen, terrainMap)
	} else {
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
	for _, ent := range entities {
		pos := m.worldToMinimap(ent.Position)
		size := m.worldSizeToMinimap(ent.Size)
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
func (m *Minimap) drawTerrain(screen *ebiten.Image, terrainMap *terrain.Map) {
	scaleX := m.bounds.Size.X / m.worldSize.X
	scaleY := m.bounds.Size.Y / m.worldSize.Y
	tileScreenWidth := terrain.TileSize * scaleX
	tileScreenHeight := terrain.TileSize * scaleY
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
