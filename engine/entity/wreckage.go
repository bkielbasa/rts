package entity

import (
	"image/color"

	"github.com/bklimczak/tanks/engine/resource"
)

type Wreckage struct {
	Entity
	MetalValue float64 // Amount of metal that can be reclaimed
}

var WreckageColor = color.RGBA{30, 30, 30, 255} // Dark gray/black
func NewWreckageFromUnit(id uint64, unit *Unit) *Wreckage {
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

func NewWreckageFromBuilding(id uint64, building *Building) *Wreckage {
	metalValue := 50.0 // Default fallback
	if building.Def != nil {
		if cost, exists := building.Def.Cost[resource.Metal]; exists {
			metalValue = cost * 0.5
		}
	}
	return &Wreckage{
		Entity: Entity{
			ID:       id,
			Position: building.Position,
			Size:     building.Size,
			Color:    WreckageColor,
			Active:   true,
			Faction:  FactionNeutral,
		},
		MetalValue: metalValue,
	}
}
