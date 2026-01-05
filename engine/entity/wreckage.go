package entity

import (
	"image/color"

	"github.com/bklimczak/tanks/engine/resource"
)

// Wreckage represents a destroyed unit that can be reclaimed for metal
type Wreckage struct {
	Entity
	MetalValue float64 // Amount of metal that can be reclaimed
}

// WreckageColor is the color for all wreckage
var WreckageColor = color.RGBA{30, 30, 30, 255} // Dark gray/black

// NewWreckageFromUnit creates wreckage at the unit's position with proper metal value
func NewWreckageFromUnit(id uint64, unit *Unit) *Wreckage {
	// Calculate metal value from UnitDef cost (50% of original metal cost)
	metalValue := 25.0 // Default fallback
	if def, ok := UnitDefs[unit.Type]; ok {
		if cost, exists := def.Cost[resource.Metal]; exists {
			metalValue = cost * 0.5
		}
	}

	return &Wreckage{
		Entity: Entity{
			ID:       id,
			Position: unit.Position,
			Size:     unit.Size,
			Color:    WreckageColor,
			Active:   true,
			Faction:  FactionNeutral,
		},
		MetalValue: metalValue,
	}
}
