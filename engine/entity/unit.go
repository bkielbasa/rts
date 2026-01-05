package entity

import (
	"math"

	emath "github.com/bklimczak/tanks/engine/math"
)

// Unit represents a selectable, movable game unit
type Unit struct {
	Entity
	Type         UnitType
	Target       emath.Vec2
	HasTarget    bool
	Selected     bool
	Speed        float64
	StuckCounter int // Counts frames stuck in same position
	LastPosition emath.Vec2

	// Rotation (in radians, 0 = facing right, increases counter-clockwise)
	Angle         float64
	RotationSpeed float64 // Radians per frame

	// Combat
	Health       float64
	MaxHealth    float64
	Damage       float64 // Damage per shot
	Range        float64 // Attack range
	FireRate     float64 // Shots per second
	FireCooldown float64 // Time until next shot
	AttackTarget *Unit   // Current attack target

	// Build task for constructors
	BuildTarget  *Building    // Building being constructed
	BuildDef     *BuildingDef // What to build (before building is created)
	BuildPos     emath.Vec2   // Where to build
	HasBuildTask bool         // Whether unit has a build task
	IsBuilding   bool         // Whether unit is actively constructing
}

// Rotation speed constants (radians per frame)
const (
	RotationSpeedTank        = 0.05  // Base rotation speed
	RotationSpeedConstructor = 0.1   // 2x tank
	RotationSpeedScout       = 0.2   // 4x tank
	RotationSpeedBasic       = 0.08  // Default
)

// NewConstructor creates a constructor unit for a faction
func NewConstructor(id uint64, x, y float64, faction Faction) *Unit {
	def := UnitDefs[UnitTypeConstructor]
	return &Unit{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(def.Size, def.Size),
			Color:    GetFactionTintedColor(def.Color, faction),
			Active:   true,
			Faction:  faction,
		},
		Type:          UnitTypeConstructor,
		Speed:         def.Speed,
		RotationSpeed: RotationSpeedConstructor,
		Health:        def.Health,
		MaxHealth:     def.Health,
		Damage:        def.Damage,
		Range:         def.Range,
		FireRate:      def.FireRate,
	}
}

// NewTank creates a tank unit for a faction
func NewTank(id uint64, x, y float64, faction Faction) *Unit {
	def := UnitDefs[UnitTypeTank]
	return &Unit{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(def.Size, def.Size),
			Color:    GetFactionTintedColor(def.Color, faction),
			Active:   true,
			Faction:  faction,
		},
		Type:          UnitTypeTank,
		Speed:         def.Speed,
		RotationSpeed: RotationSpeedTank,
		Health:        def.Health,
		MaxHealth:     def.Health,
		Damage:        def.Damage,
		Range:         def.Range,
		FireRate:      def.FireRate,
	}
}

// NewScout creates a scout unit for a faction
func NewScout(id uint64, x, y float64, faction Faction) *Unit {
	def := UnitDefs[UnitTypeScout]
	return &Unit{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(def.Size, def.Size),
			Color:    GetFactionTintedColor(def.Color, faction),
			Active:   true,
			Faction:  faction,
		},
		Type:          UnitTypeScout,
		Speed:         def.Speed,
		RotationSpeed: RotationSpeedScout,
		Health:        def.Health,
		MaxHealth:     def.Health,
		Damage:        def.Damage,
		Range:         def.Range,
		FireRate:      def.FireRate,
	}
}

// NewUnitFromDef creates a unit from a unit definition for a faction
func NewUnitFromDef(id uint64, x, y float64, def *UnitDef, faction Faction) *Unit {
	rotSpeed := RotationSpeedBasic
	switch def.Type {
	case UnitTypeTank:
		rotSpeed = RotationSpeedTank
	case UnitTypeScout:
		rotSpeed = RotationSpeedScout
	case UnitTypeConstructor:
		rotSpeed = RotationSpeedConstructor
	}

	return &Unit{
		Entity: Entity{
			ID:       id,
			Position: emath.NewVec2(x, y),
			Size:     emath.NewVec2(def.Size, def.Size),
			Color:    GetFactionTintedColor(def.Color, faction),
			Active:   true,
			Faction:  faction,
		},
		Type:          def.Type,
		Speed:         def.Speed,
		RotationSpeed: rotSpeed,
		Health:        def.Health,
		MaxHealth:     def.Health,
		Damage:        def.Damage,
		Range:         def.Range,
		FireRate:      def.FireRate,
	}
}

// CanBuild returns true if this unit can build structures
func (u *Unit) CanBuild() bool {
	return u.Type == UnitTypeConstructor
}

// GetBuildOptions returns available building options for this unit
func (u *Unit) GetBuildOptions() []*BuildingDef {
	return GetBuildableDefs(u.Type)
}

// SetTarget sets the movement target for the unit
func (u *Unit) SetTarget(target emath.Vec2) {
	u.Target = target
	u.HasTarget = true
	u.StuckCounter = 0
}

// ClearTarget stops the unit's movement
func (u *Unit) ClearTarget() {
	u.HasTarget = false
	u.Velocity = emath.Vec2{}
	u.StuckCounter = 0
}

// Update moves the unit towards its target
// Returns the desired new position (collision system may modify it)
func (u *Unit) Update() emath.Vec2 {
	if !u.HasTarget {
		return u.Position
	}

	// Calculate direction to target (target is where unit center should go)
	currentCenter := u.Center()
	direction := u.Target.Sub(currentCenter)
	distSquared := direction.LengthSquared()

	// Arrived at target
	if distSquared < u.Speed*u.Speed {
		u.Position = u.Target.Sub(u.Size.Mul(0.5))
		u.ClearTarget()
		return u.Position
	}

	// Check if stuck (hasn't moved significantly in a while)
	if u.Position.DistanceSquared(u.LastPosition) < 0.1 {
		u.StuckCounter++
		if u.StuckCounter > 30 { // Stuck for ~0.5 seconds at 60fps
			u.ClearTarget()
			return u.Position
		}
	} else {
		u.StuckCounter = 0
	}
	u.LastPosition = u.Position

	// Calculate target angle
	targetAngle := math.Atan2(direction.Y, direction.X)

	// Rotate towards target angle
	u.rotateTowards(targetAngle)

	// Move in the direction we're facing (not directly towards target)
	u.Velocity = emath.Vec2{
		X: math.Cos(u.Angle) * u.Speed,
		Y: math.Sin(u.Angle) * u.Speed,
	}
	return u.Position.Add(u.Velocity)
}

// rotateTowards smoothly rotates the unit towards the target angle
func (u *Unit) rotateTowards(targetAngle float64) {
	// Normalize angles to [-Pi, Pi]
	diff := normalizeAngle(targetAngle - u.Angle)

	// Rotate by rotation speed, clamped to the difference
	if math.Abs(diff) < u.RotationSpeed {
		u.Angle = targetAngle
	} else if diff > 0 {
		u.Angle += u.RotationSpeed
	} else {
		u.Angle -= u.RotationSpeed
	}

	// Keep angle normalized
	u.Angle = normalizeAngle(u.Angle)
}

// normalizeAngle normalizes an angle to the range [-Pi, Pi]
func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// ApplyPosition sets the unit's position after collision resolution
func (u *Unit) ApplyPosition(pos emath.Vec2) {
	u.Position = pos
}

// SetBuildTask assigns a build task to the constructor
func (u *Unit) SetBuildTask(def *BuildingDef, pos emath.Vec2) {
	if !u.CanBuild() {
		return
	}
	u.BuildDef = def
	u.BuildPos = pos
	u.HasBuildTask = true
	u.IsBuilding = false
	u.BuildTarget = nil

	// Set movement target to build site (offset so unit is next to the building)
	buildSiteTarget := emath.Vec2{
		X: pos.X - u.Size.X - 5, // Stand to the left of the building
		Y: pos.Y + def.Size/2,   // Centered vertically
	}
	u.SetTarget(buildSiteTarget)
}

// ClearBuildTask cancels the current build task
func (u *Unit) ClearBuildTask() {
	u.BuildDef = nil
	u.BuildTarget = nil
	u.HasBuildTask = false
	u.IsBuilding = false
}

// IsNearBuildSite returns true if the unit is close enough to start building
func (u *Unit) IsNearBuildSite() bool {
	if !u.HasBuildTask || u.BuildDef == nil {
		return false
	}
	// Check if unit center is within range of build site
	unitCenter := u.Center()
	buildCenter := emath.Vec2{
		X: u.BuildPos.X + u.BuildDef.Size/2,
		Y: u.BuildPos.Y + u.BuildDef.Size/2,
	}
	maxDist := u.BuildDef.Size/2 + u.Size.X + 20 // Building radius + unit size + margin
	return unitCenter.DistanceSquared(buildCenter) < maxDist*maxDist
}

// CanAttack returns true if this unit can attack
func (u *Unit) CanAttack() bool {
	return u.Damage > 0 && u.Range > 0
}

// IsInRange returns true if the target is within attack range
func (u *Unit) IsInRange(target *Unit) bool {
	if target == nil {
		return false
	}
	dist := u.Center().Distance(target.Center())
	return dist <= u.Range
}

// TakeDamage applies damage to the unit, returns true if the unit dies
func (u *Unit) TakeDamage(damage float64) bool {
	u.Health -= damage
	if u.Health <= 0 {
		u.Health = 0
		u.Active = false
		return true
	}
	return false
}

// UpdateCombat handles attack cooldown, returns true if ready to fire
func (u *Unit) UpdateCombat(dt float64) bool {
	if u.FireCooldown > 0 {
		u.FireCooldown -= dt
	}
	if u.FireCooldown <= 0 && u.AttackTarget != nil && u.CanAttack() {
		if u.IsInRange(u.AttackTarget) && u.AttackTarget.Active {
			u.FireCooldown = 1.0 / u.FireRate
			return true
		}
	}
	return false
}

// SetAttackTarget sets the unit's attack target
func (u *Unit) SetAttackTarget(target *Unit) {
	u.AttackTarget = target
}

// ClearAttackTarget clears the attack target
func (u *Unit) ClearAttackTarget() {
	u.AttackTarget = nil
}

// HealthRatio returns current health as a ratio (0-1)
func (u *Unit) HealthRatio() float64 {
	if u.MaxHealth <= 0 {
		return 1
	}
	return u.Health / u.MaxHealth
}
