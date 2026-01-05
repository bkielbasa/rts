package entity

import (
	emath "github.com/bklimczak/tanks/engine/math"
)

// Building represents a constructed building
type Building struct {
	Entity
	Type          BuildingType
	Def           *BuildingDef
	Completed     bool
	BuildProgress float64 // 0.0 to 1.0

	// Production
	Producing          bool
	ProductionProgress float64 // 0.0 to 1.0
	ProductionTime     float64 // Total time to produce
	CurrentProduction  *UnitDef // What unit type is being produced
	ProductionQueue    []*UnitDef // Queue of units to produce
	RallyPoint         emath.Vec2
	HasRallyPoint      bool
	Selected           bool
}

// NewBuilding creates a new completed building at the given position
func NewBuilding(id uint64, x, y float64, def *BuildingDef) *Building {
	b := &Building{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(def.Size, def.Size),
			Color:    def.Color,
			Active:   true,
		},
		Type:          def.Type,
		Def:           def,
		Completed:     true,
		BuildProgress: 1.0,
	}
	// Set default rally point to the right of the building
	b.RallyPoint = emath.Vec2{X: x + def.Size + 20, Y: y + def.Size/2}
	b.HasRallyPoint = true
	return b
}

// NewBuildingUnderConstruction creates a new building that needs to be constructed
func NewBuildingUnderConstruction(id uint64, x, y float64, def *BuildingDef) *Building {
	b := &Building{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(def.Size, def.Size),
			Color:    def.Color,
			Active:   true,
		},
		Type:          def.Type,
		Def:           def,
		Completed:     false,
		BuildProgress: 0.0,
	}
	// Set default rally point to the right of the building
	b.RallyPoint = emath.Vec2{X: x + def.Size + 20, Y: y + def.Size/2}
	b.HasRallyPoint = true
	return b
}

// UpdateConstruction advances construction progress, returns true when complete
func (b *Building) UpdateConstruction(dt float64) bool {
	if b.Completed {
		return false
	}

	b.BuildProgress += dt / b.Def.BuildTime
	if b.BuildProgress >= 1.0 {
		b.BuildProgress = 1.0
		b.Completed = true
		return true
	}
	return false
}

// CanProduce returns true if this building can produce units
func (b *Building) CanProduce() bool {
	return b.Type == BuildingTankFactory && b.Completed
}

// QueueProduction adds a unit to the production queue
func (b *Building) QueueProduction(unitDef *UnitDef) {
	if !b.CanProduce() || unitDef == nil {
		return
	}
	b.ProductionQueue = append(b.ProductionQueue, unitDef)

	// Start production if not already producing
	if !b.Producing {
		b.startNextProduction()
	}
}

// startNextProduction starts producing the next unit in queue
func (b *Building) startNextProduction() {
	if len(b.ProductionQueue) == 0 {
		return
	}

	b.CurrentProduction = b.ProductionQueue[0]
	b.ProductionQueue = b.ProductionQueue[1:]
	b.Producing = true
	b.ProductionProgress = 0
	b.ProductionTime = b.CurrentProduction.BuildTime
}

// GetQueueCount returns the count of a specific unit type in the queue (including current)
func (b *Building) GetQueueCount(unitType UnitType) int {
	count := 0
	if b.Producing && b.CurrentProduction != nil && b.CurrentProduction.Type == unitType {
		count++
	}
	for _, def := range b.ProductionQueue {
		if def.Type == unitType {
			count++
		}
	}
	return count
}

// GetTotalQueueCount returns total units in queue (including current)
func (b *Building) GetTotalQueueCount() int {
	count := len(b.ProductionQueue)
	if b.Producing {
		count++
	}
	return count
}

// UpdateProduction updates production progress, returns the completed unit def if ready
func (b *Building) UpdateProduction(dt float64) *UnitDef {
	if !b.Producing {
		return nil
	}

	b.ProductionProgress += dt / b.ProductionTime
	if b.ProductionProgress >= 1.0 {
		completedUnit := b.CurrentProduction
		b.Producing = false
		b.ProductionProgress = 0
		b.CurrentProduction = nil

		// Start next in queue
		b.startNextProduction()

		return completedUnit
	}
	return nil
}

// GetSpawnPoint returns where new units should spawn
func (b *Building) GetSpawnPoint() emath.Vec2 {
	// Spawn at the bottom-center of the building
	return emath.Vec2{
		X: b.Position.X + b.Size.X/2 - 11, // Center, offset by half tank size
		Y: b.Position.Y + b.Size.Y + 5,    // Just below the building
	}
}
