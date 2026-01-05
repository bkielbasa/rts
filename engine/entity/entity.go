package entity

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"image/color"
)

type Faction int

const (
	FactionPlayer Faction = iota
	FactionEnemy
	FactionNeutral
)

var FactionColors = map[Faction]color.RGBA{
	FactionPlayer:  {50, 150, 50, 255},   // Green
	FactionEnemy:   {200, 50, 50, 255},   // Red
	FactionNeutral: {150, 150, 150, 255}, // Gray
}

func GetFactionTintedColor(base color.Color, faction Faction) color.RGBA {
	factionColor := FactionColors[faction]
	r, g, b, a := base.RGBA()
	baseR := uint8(r >> 8)
	baseG := uint8(g >> 8)
	baseB := uint8(b >> 8)
	baseA := uint8(a >> 8)
	blendR := uint8((float64(baseR)*0.7 + float64(factionColor.R)*0.3))
	blendG := uint8((float64(baseG)*0.7 + float64(factionColor.G)*0.3))
	blendB := uint8((float64(baseB)*0.7 + float64(factionColor.B)*0.3))
	return color.RGBA{blendR, blendG, blendB, baseA}
}

type Entity struct {
	ID       uint64
	Position emath.Vec2
	Size     emath.Vec2
	Velocity emath.Vec2
	Color    color.Color
	Active   bool
	Faction  Faction
}

func (e *Entity) Bounds() emath.Rect {
	return emath.Rect{Pos: e.Position, Size: e.Size}
}
func (e *Entity) Center() emath.Vec2 {
	return e.Bounds().Center()
}
func (e *Entity) Contains(p emath.Vec2) bool {
	return e.Bounds().Contains(p)
}
