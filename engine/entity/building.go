package entity

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
)

type Building struct {
	Entity
	Type                  BuildingType
	Def                   *BuildingDef
	Completed             bool
	BuildProgress         float64
	MetalSpent            float64
	EnergySpent           float64
	Producing             bool
	ProductionProgress    float64
	ProductionTime        float64
	CurrentProduction     *UnitDef
	ProductionQueue       []*UnitDef
	ProductionMetalSpent  float64
	ProductionEnergySpent float64
	RallyPoint            emath.Vec2
	HasRallyPoint         bool
	Selected              bool
	Health                float64
	MaxHealth             float64

	// Combat state for defensive buildings
	AttackTarget *Unit
	FireCooldown float64

	// Animation state
	AnimationTime  float64
	AnimationFrame int
}

func NewBuilding(id uint64, x, y float64, def *BuildingDef) *Building {
	w, h := def.GetWidth(), def.GetHeight()
	b := &Building{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(w, h),
			Color:    def.Color,
			Active:   true,
		},
		Type:          def.Type,
		Def:           def,
		Completed:     true,
		BuildProgress: 1.0,
		Health:        def.Health,
		MaxHealth:     def.Health,
	}
	b.RallyPoint = emath.Vec2{X: x + w + 20, Y: y + h/2}
	b.HasRallyPoint = true
	return b
}
func NewBuildingUnderConstruction(id uint64, x, y float64, def *BuildingDef) *Building {
	w, h := def.GetWidth(), def.GetHeight()
	b := &Building{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(w, h),
			Color:    def.Color,
			Active:   true,
		},
		Type:          def.Type,
		Def:           def,
		Completed:     false,
		BuildProgress: 0.0,
		Health:        def.Health * 0.1,
		MaxHealth:     def.Health,
	}
	b.RallyPoint = emath.Vec2{X: x + w + 20, Y: y + h/2}
	b.HasRallyPoint = true
	return b
}
func (b *Building) UpdateConstruction(dt float64, resources *resource.Manager) bool {
	if b.Completed {
		return false
	}
	progressDelta := dt / b.Def.BuildTime
	targetProgress := b.BuildProgress + progressDelta
	if targetProgress > 1.0 {
		targetProgress = 1.0
	}
	metalCost := b.Def.Cost[resource.Credits]
	energyCost := b.Def.Cost[resource.Energy]
	metalNeeded := targetProgress*metalCost - b.MetalSpent
	energyNeeded := targetProgress*energyCost - b.EnergySpent
	metalAvailable := resources.Get(resource.Credits).Current
	energyAvailable := resources.Get(resource.Energy).Current
	metalToSpend := metalNeeded
	if metalToSpend > metalAvailable {
		metalToSpend = metalAvailable
	}
	energyToSpend := energyNeeded
	if energyToSpend > energyAvailable {
		energyToSpend = energyAvailable
	}
	var actualProgress float64
	if metalCost > 0 && energyCost > 0 {
		metalProgress := (b.MetalSpent + metalToSpend) / metalCost
		energyProgress := (b.EnergySpent + energyToSpend) / energyCost
		if metalProgress < energyProgress {
			actualProgress = metalProgress
		} else {
			actualProgress = energyProgress
		}
	} else if metalCost > 0 {
		actualProgress = (b.MetalSpent + metalToSpend) / metalCost
	} else if energyCost > 0 {
		actualProgress = (b.EnergySpent + energyToSpend) / energyCost
	} else {
		actualProgress = targetProgress
	}
	if actualProgress > targetProgress {
		actualProgress = targetProgress
	}
	if metalCost > 0 {
		metalToSpend = actualProgress*metalCost - b.MetalSpent
		if metalToSpend > 0 {
			resources.Get(resource.Credits).SpendWithTracking(metalToSpend)
			b.MetalSpent += metalToSpend
		}
	}
	if energyCost > 0 {
		energyToSpend = actualProgress*energyCost - b.EnergySpent
		if energyToSpend > 0 {
			resources.Get(resource.Energy).SpendWithTracking(energyToSpend)
			b.EnergySpent += energyToSpend
		}
	}
	b.BuildProgress = actualProgress
	if b.BuildProgress >= 1.0 {
		b.BuildProgress = 1.0
		b.Completed = true
		return true
	}
	return false
}
func (b *Building) CanProduce() bool {
	return b.Def != nil && b.Def.IsFactory && b.Completed
}
func (b *Building) QueueProduction(unitDef *UnitDef) {
	if !b.CanProduce() || unitDef == nil {
		return
	}
	b.ProductionQueue = append(b.ProductionQueue, unitDef)
	if !b.Producing {
		b.startNextProduction()
	}
}
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
func (b *Building) GetTotalQueueCount() int {
	count := len(b.ProductionQueue)
	if b.Producing {
		count++
	}
	return count
}
func (b *Building) RemoveFromQueue(unitType UnitType, resources *resource.Manager) (metalRefund, energyRefund float64) {
	for i := len(b.ProductionQueue) - 1; i >= 0; i-- {
		if b.ProductionQueue[i].Type == unitType {
			b.ProductionQueue = append(b.ProductionQueue[:i], b.ProductionQueue[i+1:]...)
			return 0, 0
		}
	}
	if b.Producing && b.CurrentProduction != nil && b.CurrentProduction.Type == unitType {
		metalRefund = b.ProductionMetalSpent
		energyRefund = b.ProductionEnergySpent
		if resources != nil {
			if metalRefund > 0 {
				resources.Get(resource.Credits).Add(metalRefund)
			}
			if energyRefund > 0 {
				resources.Get(resource.Energy).Add(energyRefund)
			}
		}
		b.Producing = false
		b.ProductionProgress = 0
		b.CurrentProduction = nil
		b.ProductionMetalSpent = 0
		b.ProductionEnergySpent = 0
		b.startNextProduction()
		return metalRefund, energyRefund
	}
	return 0, 0
}
func (b *Building) UpdateProduction(dt float64, resources *resource.Manager) *UnitDef {
	if !b.Producing {
		return nil
	}
	progressDelta := dt / b.ProductionTime
	targetProgress := b.ProductionProgress + progressDelta
	if targetProgress > 1.0 {
		targetProgress = 1.0
	}
	metalCost := b.CurrentProduction.Cost[resource.Credits]
	energyCost := b.CurrentProduction.Cost[resource.Energy]
	metalNeeded := targetProgress*metalCost - b.ProductionMetalSpent
	energyNeeded := targetProgress*energyCost - b.ProductionEnergySpent
	metalAvailable := resources.Get(resource.Credits).Current
	energyAvailable := resources.Get(resource.Energy).Current
	metalToSpend := metalNeeded
	if metalToSpend > metalAvailable {
		metalToSpend = metalAvailable
	}
	energyToSpend := energyNeeded
	if energyToSpend > energyAvailable {
		energyToSpend = energyAvailable
	}
	var actualProgress float64
	if metalCost > 0 && energyCost > 0 {
		metalProgress := (b.ProductionMetalSpent + metalToSpend) / metalCost
		energyProgress := (b.ProductionEnergySpent + energyToSpend) / energyCost
		if metalProgress < energyProgress {
			actualProgress = metalProgress
		} else {
			actualProgress = energyProgress
		}
	} else if metalCost > 0 {
		actualProgress = (b.ProductionMetalSpent + metalToSpend) / metalCost
	} else if energyCost > 0 {
		actualProgress = (b.ProductionEnergySpent + energyToSpend) / energyCost
	} else {
		actualProgress = targetProgress
	}
	if actualProgress > targetProgress {
		actualProgress = targetProgress
	}
	if metalCost > 0 {
		metalToSpend = actualProgress*metalCost - b.ProductionMetalSpent
		if metalToSpend > 0 {
			resources.Get(resource.Credits).SpendWithTracking(metalToSpend)
			b.ProductionMetalSpent += metalToSpend
		}
	}
	if energyCost > 0 {
		energyToSpend = actualProgress*energyCost - b.ProductionEnergySpent
		if energyToSpend > 0 {
			resources.Get(resource.Energy).SpendWithTracking(energyToSpend)
			b.ProductionEnergySpent += energyToSpend
		}
	}
	b.ProductionProgress = actualProgress
	if b.ProductionProgress >= 1.0 {
		completedUnit := b.CurrentProduction
		b.Producing = false
		b.ProductionProgress = 0
		b.CurrentProduction = nil
		b.ProductionMetalSpent = 0
		b.ProductionEnergySpent = 0
		b.startNextProduction()
		return completedUnit
	}
	return nil
}
func (b *Building) GetSpawnPoint() emath.Vec2 {
	return emath.Vec2{
		X: b.Position.X + b.Size.X/2 - 11,
		Y: b.Position.Y + b.Size.Y + 5,
	}
}

func (b *Building) TakeDamage(damage float64) bool {
	b.Health -= damage
	if b.Health <= 0 {
		b.Health = 0
		b.Active = false
		return true
	}
	return false
}

func (b *Building) HealthRatio() float64 {
	if b.MaxHealth <= 0 {
		return 1
	}
	return b.Health / b.MaxHealth
}

// CanAttack returns true if this building can attack enemies
func (b *Building) CanAttack() bool {
	return b.Def != nil && b.Def.CanAttack && b.Completed
}

// IsInAttackRange checks if a target unit is within attack range
func (b *Building) IsInAttackRange(target *Unit) bool {
	if target == nil || b.Def == nil {
		return false
	}
	dist := b.Center().Distance(target.Center())
	return dist <= b.Def.AttackRange
}

// SetAttackTarget sets the building's current attack target
func (b *Building) SetAttackTarget(target *Unit) {
	b.AttackTarget = target
}

// ClearAttackTarget clears the building's attack target
func (b *Building) ClearAttackTarget() {
	b.AttackTarget = nil
}

// UpdateAnimation updates the building's animation frame
func (b *Building) UpdateAnimation(dt float64) {
	if b.Def == nil || b.Def.AnimationSpeed <= 0 || b.Def.SpriteHeight <= 0 {
		return
	}
	b.AnimationTime += dt
	frameDuration := 1.0 / b.Def.AnimationSpeed
	if b.AnimationTime >= frameDuration {
		b.AnimationTime -= frameDuration
		b.AnimationFrame++
	}
}
