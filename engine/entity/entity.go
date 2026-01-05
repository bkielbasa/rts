package entity

import (
	"image/color"

	emath "github.com/bklimczak/tanks/engine/math"
)

// Faction represents a team/side in the game
type Faction int

const (
	FactionPlayer Faction = iota
	FactionEnemy
	FactionNeutral
)

// FactionColors defines the primary color for each faction
var FactionColors = map[Faction]color.RGBA{
	FactionPlayer:  {50, 150, 50, 255},   // Green
	FactionEnemy:   {200, 50, 50, 255},   // Red
	FactionNeutral: {150, 150, 150, 255}, // Gray
}

// GetFactionTintedColor applies a faction tint to a base color
func GetFactionTintedColor(base color.Color, faction Faction) color.RGBA {
	factionColor := FactionColors[faction]
	r, g, b, a := base.RGBA()

	// Convert from 16-bit to 8-bit
	baseR := uint8(r >> 8)
	baseG := uint8(g >> 8)
	baseB := uint8(b >> 8)
	baseA := uint8(a >> 8)

	// Blend base color with faction color (70% base, 30% faction)
	blendR := uint8((float64(baseR)*0.7 + float64(factionColor.R)*0.3))
	blendG := uint8((float64(baseG)*0.7 + float64(factionColor.G)*0.3))
	blendB := uint8((float64(baseB)*0.7 + float64(factionColor.B)*0.3))

	return color.RGBA{blendR, blendG, blendB, baseA}
}

// Entity represents any game object with position and size
type Entity struct {
	ID       uint64
	Position emath.Vec2
	Size     emath.Vec2
	Velocity emath.Vec2
	Color    color.Color
	Active   bool
	Faction  Faction
}

// Bounds returns the bounding rectangle of the entity
func (e *Entity) Bounds() emath.Rect {
	return emath.Rect{Pos: e.Position, Size: e.Size}
}

// Center returns the center point of the entity
func (e *Entity) Center() emath.Vec2 {
	return e.Bounds().Center()
}

// Contains checks if a point is inside the entity
func (e *Entity) Contains(p emath.Vec2) bool {
	return e.Bounds().Contains(p)
}
