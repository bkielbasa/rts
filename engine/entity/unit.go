package entity

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"math"
)

// BuildTask represents a queued building construction task
type BuildTask struct {
	Def *BuildingDef
	Pos emath.Vec2
}

type Unit struct {
	Entity
	Def                  *UnitDef // Reference to unit definition for data-driven behavior
	Type                 UnitType
	Target               emath.Vec2
	HasTarget            bool
	Selected             bool
	Speed                float64
	StuckCounter         int
	LastPosition         emath.Vec2
	Angle                float64       // Body angle (direction unit is facing)
	TurretAngle          float64       // Turret angle (direction turret is aiming)
	RotationSpeed        float64       // Body rotation speed
	TurretRotationSpeed  float64       // Turret rotation speed (faster than body)
	Health               float64
	MaxHealth            float64
	Damage               float64
	Range                float64
	FireRate             float64
	FireCooldown         float64
	AttackTarget         *Unit
	BuildingAttackTarget *Building
	VisionRange          float64
	PursuitRange         float64   // Range to keep chasing an enemy (usually > fire range)
	BuildTarget          *Building
	BuildDef             *BuildingDef
	BuildPos             emath.Vec2
	HasBuildTask         bool
	IsBuilding           bool
	BuildQueue           []BuildTask // Queue of pending build tasks
	RepairRate           float64     // Health per second when repairing
	RepairRange          float64     // Range to repair units
	RepairTarget         *Unit       // Unit being repaired
}

const (
	RotationSpeedBasic = 0.08 // Default fallback rotation speed
)

// NewConstructor creates a constructor unit (convenience function)
func NewConstructor(id uint64, x, y float64, faction Faction) *Unit {
	return NewUnitFromDef(id, x, y, UnitDefs[UnitTypeConstructor], faction)
}

// NewTank creates a tank unit (convenience function)
func NewTank(id uint64, x, y float64, faction Faction) *Unit {
	return NewUnitFromDef(id, x, y, UnitDefs[UnitTypeTank], faction)
}

// NewScout creates a scout unit (convenience function)
func NewScout(id uint64, x, y float64, faction Faction) *Unit {
	return NewUnitFromDef(id, x, y, UnitDefs[UnitTypeScout], faction)
}

// NewUnitFromDef creates a unit from a definition (data-driven)
func NewUnitFromDef(id uint64, x, y float64, def *UnitDef, faction Faction) *Unit {
	rotSpeed := def.RotationSpeed
	if rotSpeed == 0 {
		rotSpeed = RotationSpeedBasic
	}
	turretRotSpeed := def.GetTurretRotationSpeed()
	if turretRotSpeed == 0 {
		turretRotSpeed = rotSpeed * 2 // Default: turret rotates faster than body
	}
	return &Unit{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(def.GetWidth(), def.GetHeight()),
			Color:    GetFactionTintedColor(def.Color, faction),
			Active:   true,
			Faction:  faction,
		},
		Def:                 def,
		Type:                def.Type,
		Speed:               def.Speed,
		RotationSpeed:       rotSpeed,
		TurretRotationSpeed: turretRotSpeed,
		Health:              def.Health,
		MaxHealth:           def.Health,
		Damage:              def.GetDamage(),
		Range:               def.GetRange(),
		FireRate:            def.GetFireRate(),
		VisionRange:         def.VisionRange,
		PursuitRange:        def.GetRange() * 1.5, // Chase enemies 1.5x further than fire range
		RepairRate:          def.GetRepairRate(),
		RepairRange:         def.GetRepairRange(),
	}
}
func (u *Unit) CanBuild() bool {
	return u.Def != nil && u.Def.CanConstruct()
}

// Center returns the center point of the unit
func (u *Unit) Center() emath.Vec2 {
	return emath.Vec2{
		X: u.Position.X + u.Size.X/2,
		Y: u.Position.Y + u.Size.Y/2,
	}
}

func (u *Unit) CanRepair() bool {
	return u.Def != nil && u.Def.CanRepairUnits() && u.RepairRate > 0 && u.RepairRange > 0
}
func (u *Unit) IsInRepairRange(target *Unit) bool {
	if target == nil {
		return false
	}
	dist := u.Center().Distance(target.Center())
	return dist <= u.RepairRange
}
func (u *Unit) SetRepairTarget(target *Unit) {
	u.RepairTarget = target
}
func (u *Unit) ClearRepairTarget() {
	u.RepairTarget = nil
}
func (u *Unit) GetBuildOptions() []*BuildingDef {
	return GetBuildableDefs(u.Type)
}
func (u *Unit) SetTarget(target emath.Vec2) {
	u.Target = target
	u.HasTarget = true
	u.StuckCounter = 0
}
func (u *Unit) ClearTarget() {
	u.HasTarget = false
	u.Velocity = emath.Vec2{}
	u.StuckCounter = 0
}
func (u *Unit) Update() emath.Vec2 {
	if !u.HasTarget {
		return u.Position
	}
	currentCenter := u.Center()
	direction := u.Target.Sub(currentCenter)
	distSquared := direction.LengthSquared()
	if distSquared < u.Speed*u.Speed {
		u.Position = u.Target.Sub(u.Size.Mul(0.5))
		u.ClearTarget()
		return u.Position
	}
	if u.Position.DistanceSquared(u.LastPosition) < 0.1 {
		u.StuckCounter++
		if u.StuckCounter > 30 {
			u.ClearTarget()
			return u.Position
		}
	} else {
		u.StuckCounter = 0
	}
	u.LastPosition = u.Position
	targetAngle := math.Atan2(direction.Y, direction.X)
	u.rotateTowards(targetAngle)
	u.Velocity = emath.Vec2{
		X: math.Cos(u.Angle) * u.Speed,
		Y: math.Sin(u.Angle) * u.Speed,
	}
	return u.Position.Add(u.Velocity)
}
func (u *Unit) rotateTowards(targetAngle float64) {
	diff := normalizeAngle(targetAngle - u.Angle)
	if math.Abs(diff) < u.RotationSpeed {
		u.Angle = targetAngle
	} else if diff > 0 {
		u.Angle += u.RotationSpeed
	} else {
		u.Angle -= u.RotationSpeed
	}
	u.Angle = normalizeAngle(u.Angle)
}
func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}
func (u *Unit) ApplyPosition(pos emath.Vec2) {
	u.Position = pos
}
func (u *Unit) SetBuildTask(def *BuildingDef, pos emath.Vec2) {
	if !u.CanBuild() {
		return
	}
	u.BuildDef = def
	u.BuildPos = pos
	u.HasBuildTask = true
	u.IsBuilding = false
	u.BuildTarget = nil
	// Target a position just below (south of) the building site, so constructor doesn't end up inside
	buildSiteTarget := emath.Vec2{
		X: pos.X + def.Size/2,
		Y: pos.Y + def.Size + u.Size.Y/2 + 5, // Below the building
	}
	u.SetTarget(buildSiteTarget)
}

// QueueBuildTask adds a build task to the queue, or sets it as current if no active task
func (u *Unit) QueueBuildTask(def *BuildingDef, pos emath.Vec2) {
	if !u.CanBuild() {
		return
	}
	if !u.HasBuildTask {
		// No current task, start immediately
		u.SetBuildTask(def, pos)
	} else {
		// Add to queue
		u.BuildQueue = append(u.BuildQueue, BuildTask{Def: def, Pos: pos})
	}
}

func (u *Unit) ClearBuildTask() {
	u.BuildDef = nil
	u.BuildTarget = nil
	u.HasBuildTask = false
	u.IsBuilding = false
	// Start next task from queue if available
	u.StartNextBuildTask()
}

// StartNextBuildTask starts the next build task from the queue if available
func (u *Unit) StartNextBuildTask() {
	if len(u.BuildQueue) == 0 {
		return
	}
	// Pop first task from queue
	nextTask := u.BuildQueue[0]
	u.BuildQueue = u.BuildQueue[1:]
	u.SetBuildTask(nextTask.Def, nextTask.Pos)
}

// GetBuildQueueLength returns the number of queued build tasks (including current)
func (u *Unit) GetBuildQueueLength() int {
	count := len(u.BuildQueue)
	if u.HasBuildTask {
		count++
	}
	return count
}
func (u *Unit) IsNearBuildSite() bool {
	if !u.HasBuildTask || u.BuildDef == nil {
		return false
	}
	unitCenter := u.Center()
	buildCenter := emath.Vec2{
		X: u.BuildPos.X + u.BuildDef.Size/2,
		Y: u.BuildPos.Y + u.BuildDef.Size/2,
	}
	maxDist := u.BuildDef.Size/2 + u.Size.X + 20
	return unitCenter.DistanceSquared(buildCenter) < maxDist*maxDist
}
func (u *Unit) CanAttack() bool {
	return u.Damage > 0 && u.Range > 0
}
func (u *Unit) IsInRange(target *Unit) bool {
	if target == nil {
		return false
	}
	dist := u.Center().Distance(target.Center())
	return dist <= u.Range
}

func (u *Unit) IsBuildingInRange(target *Building) bool {
	if target == nil {
		return false
	}
	dist := u.Center().Distance(target.Center())
	return dist <= u.Range
}

func (u *Unit) IsInPursuitRange(target *Unit) bool {
	if target == nil {
		return false
	}
	dist := u.Center().Distance(target.Center())
	return dist <= u.PursuitRange
}

func (u *Unit) IsBuildingInPursuitRange(target *Building) bool {
	if target == nil {
		return false
	}
	dist := u.Center().Distance(target.Center())
	return dist <= u.PursuitRange
}

func (u *Unit) TakeDamage(damage float64) bool {
	u.Health -= damage
	if u.Health <= 0 {
		u.Health = 0
		u.Active = false
		return true
	}
	return false
}
func (u *Unit) UpdateCombat(dt float64) bool {
	if u.FireCooldown > 0 {
		u.FireCooldown -= dt
	}

	// Update turret rotation towards target
	turretAimed := u.UpdateTurret()

	// Only fire if cooldown ready, can attack, and turret is aimed
	if u.FireCooldown <= 0 && u.CanAttack() && turretAimed {
		if u.AttackTarget != nil && u.IsInRange(u.AttackTarget) && u.AttackTarget.Active {
			u.FireCooldown = 1.0 / u.FireRate
			return true
		}
		if u.BuildingAttackTarget != nil && u.IsBuildingInRange(u.BuildingAttackTarget) && u.BuildingAttackTarget.Active {
			u.FireCooldown = 1.0 / u.FireRate
			return true
		}
	}
	return false
}
func (u *Unit) SetAttackTarget(target *Unit) {
	u.AttackTarget = target
	u.BuildingAttackTarget = nil
}

func (u *Unit) SetBuildingAttackTarget(target *Building) {
	u.BuildingAttackTarget = target
	u.AttackTarget = nil
}

func (u *Unit) ClearAttackTarget() {
	u.AttackTarget = nil
	u.BuildingAttackTarget = nil
}

func (u *Unit) HasAnyAttackTarget() bool {
	return u.AttackTarget != nil || u.BuildingAttackTarget != nil
}
func (u *Unit) HealthRatio() float64 {
	if u.MaxHealth <= 0 {
		return 1
	}
	return u.Health / u.MaxHealth
}

// UpdateTurret rotates the turret towards the current attack target
// Returns true if the turret is aimed at the target (within tolerance)
func (u *Unit) UpdateTurret() bool {
	var targetPos emath.Vec2
	hasTarget := false

	if u.AttackTarget != nil && u.AttackTarget.Active {
		targetPos = u.AttackTarget.Center()
		hasTarget = true
	} else if u.BuildingAttackTarget != nil && u.BuildingAttackTarget.Active {
		targetPos = u.BuildingAttackTarget.Center()
		hasTarget = true
	}

	if !hasTarget {
		return false
	}

	// Calculate angle to target
	myCenter := u.Center()
	toTarget := targetPos.Sub(myCenter)
	targetAngle := math.Atan2(toTarget.Y, toTarget.X)

	// Rotate turret towards target
	return u.rotateTurretTowards(targetAngle)
}

// rotateTurretTowards rotates the turret towards the target angle
// Returns true if the turret is aimed (within tolerance)
func (u *Unit) rotateTurretTowards(targetAngle float64) bool {
	diff := normalizeAngle(targetAngle - u.TurretAngle)
	aimTolerance := 0.1 // About 6 degrees

	if math.Abs(diff) < aimTolerance {
		u.TurretAngle = targetAngle
		return true
	}

	if math.Abs(diff) < u.TurretRotationSpeed {
		u.TurretAngle = targetAngle
	} else if diff > 0 {
		u.TurretAngle += u.TurretRotationSpeed
	} else {
		u.TurretAngle -= u.TurretRotationSpeed
	}
	u.TurretAngle = normalizeAngle(u.TurretAngle)
	return false
}

// IsTurretAimed returns true if the turret is aimed at the current target
func (u *Unit) IsTurretAimed() bool {
	var targetPos emath.Vec2
	hasTarget := false

	if u.AttackTarget != nil && u.AttackTarget.Active {
		targetPos = u.AttackTarget.Center()
		hasTarget = true
	} else if u.BuildingAttackTarget != nil && u.BuildingAttackTarget.Active {
		targetPos = u.BuildingAttackTarget.Center()
		hasTarget = true
	}

	if !hasTarget {
		return false
	}

	myCenter := u.Center()
	toTarget := targetPos.Sub(myCenter)
	targetAngle := math.Atan2(toTarget.Y, toTarget.X)
	diff := normalizeAngle(targetAngle - u.TurretAngle)

	return math.Abs(diff) < 0.1
}
