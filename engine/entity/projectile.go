package entity

import (
	"image/color"

	emath "github.com/bklimczak/tanks/engine/math"
)

// Projectile represents a fired bullet/shell traveling toward a target
type Projectile struct {
	Entity
	Damage float64
	Speed  float64
	Target *Unit // Target unit (for homing) or nil for straight shots
	// For straight shots, store direction
	Direction emath.Vec2
}

// ProjectileSpeed is the default speed for projectiles
const ProjectileSpeed = 400.0

// ProjectileSize is the default size for projectiles
const ProjectileSize = 4.0

// NewProjectile creates a projectile from shooter to target
func NewProjectile(id uint64, shooter *Unit, target *Unit) *Projectile {
	startPos := shooter.Center()
	targetPos := target.Center()

	// Calculate direction
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

// Update moves the projectile, returns true if it should be removed (hit or expired)
func (p *Projectile) Update(dt float64) bool {
	if !p.Active {
		return true
	}

	// Move projectile
	movement := p.Direction.Mul(p.Speed * dt)
	p.Position = p.Position.Add(movement)

	// Check if we've hit the target
	if p.Target != nil && p.Target.Active {
		// Check collision with target
		projectileCenter := p.Center()
		targetBounds := p.Target.Bounds()

		// Expand target bounds slightly for hit detection
		hitBounds := emath.Rect{
			Pos:  emath.Vec2{X: targetBounds.Pos.X - 2, Y: targetBounds.Pos.Y - 2},
			Size: emath.Vec2{X: targetBounds.Size.X + 4, Y: targetBounds.Size.Y + 4},
		}

		if hitBounds.Contains(projectileCenter) {
			// Hit! Apply damage
			p.Target.TakeDamage(p.Damage)
			p.Active = false
			return true
		}
	}

	// Check if projectile has traveled too far (cleanup)
	// Remove if target is dead or projectile is out of reasonable range
	if p.Target == nil || !p.Target.Active {
		// Target lost, remove projectile after a short distance
		p.Active = false
		return true
	}

	return false
}
