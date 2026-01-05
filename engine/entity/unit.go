package entity

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"math"
)

type Unit struct {
	Entity
	Type          UnitType
	Target        emath.Vec2
	HasTarget     bool
	Selected      bool
	Speed         float64
	StuckCounter  int
	LastPosition  emath.Vec2
	Angle         float64
	RotationSpeed float64
	Health        float64
	MaxHealth     float64
	Damage        float64
	Range         float64
	FireRate      float64
	FireCooldown  float64
	AttackTarget  *Unit
	BuildTarget   *Building
	BuildDef      *BuildingDef
	BuildPos      emath.Vec2
	HasBuildTask  bool
	IsBuilding    bool
}

const (
	RotationSpeedTank        = 0.05
	RotationSpeedConstructor = 0.1
	RotationSpeedScout       = 0.2
	RotationSpeedBasic       = 0.08
)

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
func (u *Unit) CanBuild() bool {
	return u.Type == UnitTypeConstructor
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
	buildSiteTarget := emath.Vec2{
		X: pos.X - u.Size.X - 5,
		Y: pos.Y + def.Size/2,
	}
	u.SetTarget(buildSiteTarget)
}
func (u *Unit) ClearBuildTask() {
	u.BuildDef = nil
	u.BuildTarget = nil
	u.HasBuildTask = false
	u.IsBuilding = false
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
	if u.FireCooldown <= 0 && u.AttackTarget != nil && u.CanAttack() {
		if u.IsInRange(u.AttackTarget) && u.AttackTarget.Active {
			u.FireCooldown = 1.0 / u.FireRate
			return true
		}
	}
	return false
}
func (u *Unit) SetAttackTarget(target *Unit) {
	u.AttackTarget = target
}
func (u *Unit) ClearAttackTarget() {
	u.AttackTarget = nil
}
func (u *Unit) HealthRatio() float64 {
	if u.MaxHealth <= 0 {
		return 1
	}
	return u.Health / u.MaxHealth
}
