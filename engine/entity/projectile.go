package entity

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"image/color"
)

type Projectile struct {
	Entity
	Damage    float64
	Speed     float64
	Target    *Unit // Target unit (for homing) or nil for straight shots
	Direction emath.Vec2
}

const ProjectileSpeed = 400.0
const ProjectileSize = 4.0

func NewProjectile(id uint64, shooter *Unit, target *Unit) *Projectile {
	startPos := shooter.Center()
	targetPos := target.Center()
	dir := targetPos.Sub(startPos).Normalize()
	return &Projectile{
		Entity: Entity{
			ID:       id,
			Position: emath.Vec2{X: startPos.X - ProjectileSize/2, Y: startPos.Y - ProjectileSize/2},
			Size:     emath.Vec2{X: ProjectileSize, Y: ProjectileSize},
			Color:    color.RGBA{255, 200, 50, 255}, // Yellow/orange
			Active:   true,
			Faction:  shooter.Faction,
		},
		Damage:    shooter.Damage,
		Speed:     ProjectileSpeed,
		Target:    target,
		Direction: dir,
	}
}
func (p *Projectile) Update(dt float64) bool {
	if !p.Active {
		return true
	}
	movement := p.Direction.Mul(p.Speed * dt)
	p.Position = p.Position.Add(movement)
	if p.Target != nil && p.Target.Active {
		projectileCenter := p.Center()
		targetBounds := p.Target.Bounds()
		hitBounds := emath.Rect{
			Pos:  emath.Vec2{X: targetBounds.Pos.X - 2, Y: targetBounds.Pos.Y - 2},
			Size: emath.Vec2{X: targetBounds.Size.X + 4, Y: targetBounds.Size.Y + 4},
		}
		if hitBounds.Contains(projectileCenter) {
			p.Target.TakeDamage(p.Damage)
			p.Active = false
			return true
		}
	}
	if p.Target == nil || !p.Target.Active {
		p.Active = false
		return true
	}
	return false
}
