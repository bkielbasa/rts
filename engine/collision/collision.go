package collision

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"math"
)

type TerrainChecker interface {
	IsPassable(bounds emath.Rect) bool
}
type Collidable interface {
	Bounds() emath.Rect
}
type System struct {
	worldBounds emath.Rect
	terrain     TerrainChecker
}

func NewSystem(worldWidth, worldHeight float64) *System {
	return &System{
		worldBounds: emath.NewRect(0, 0, worldWidth, worldHeight),
	}
}
func (s *System) SetWorldBounds(width, height float64) {
	s.worldBounds = emath.NewRect(0, 0, width, height)
}
func (s *System) SetTerrain(t TerrainChecker) {
	s.terrain = t
}
func (s *System) CheckCollision(a, b emath.Rect) bool {
	return a.Intersects(b)
}
func (s *System) isPositionValid(bounds emath.Rect) bool {
	if bounds.Pos.X < s.worldBounds.Pos.X || bounds.Pos.Y < s.worldBounds.Pos.Y {
		return false
	}
	if bounds.Pos.X+bounds.Size.X > s.worldBounds.Pos.X+s.worldBounds.Size.X {
		return false
	}
	if bounds.Pos.Y+bounds.Size.Y > s.worldBounds.Pos.Y+s.worldBounds.Size.Y {
		return false
	}
	if s.terrain != nil && !s.terrain.IsPassable(bounds) {
		return false
	}
	return true
}
func (s *System) ResolveMovement(mover emath.Rect, desiredPos emath.Vec2, obstacles []emath.Rect) emath.Vec2 {
	newBounds := emath.Rect{Pos: desiredPos, Size: mover.Size}
	desiredPos = s.clampToWorld(desiredPos, mover.Size)
	newBounds.Pos = desiredPos
	if s.terrain != nil && !s.terrain.IsPassable(newBounds) {
		xOnlyPos := emath.Vec2{X: desiredPos.X, Y: mover.Pos.Y}
		xOnlyBounds := emath.Rect{Pos: xOnlyPos, Size: mover.Size}
		if s.terrain.IsPassable(xOnlyBounds) {
			desiredPos = xOnlyPos
			newBounds.Pos = desiredPos
		} else {
			yOnlyPos := emath.Vec2{X: mover.Pos.X, Y: desiredPos.Y}
			yOnlyBounds := emath.Rect{Pos: yOnlyPos, Size: mover.Size}
			if s.terrain.IsPassable(yOnlyBounds) {
				desiredPos = yOnlyPos
				newBounds.Pos = desiredPos
			} else {
				return mover.Pos
			}
		}
	}
	for _, obs := range obstacles {
		currentlyOverlapping := mover.Intersects(obs)
		if newBounds.Intersects(obs) {
			if currentlyOverlapping {
				currentOverlap := s.overlapAmount(mover, obs)
				newOverlap := s.overlapAmount(newBounds, obs)
				if newOverlap < currentOverlap {
					continue
				}
			}
			xOnlyPos := emath.Vec2{X: desiredPos.X, Y: mover.Pos.Y}
			xOnlyBounds := emath.Rect{Pos: xOnlyPos, Size: mover.Size}
			if !xOnlyBounds.Intersects(obs) && (s.terrain == nil || s.terrain.IsPassable(xOnlyBounds)) {
				desiredPos = xOnlyPos
				newBounds.Pos = desiredPos
				continue
			}
			if currentlyOverlapping && xOnlyBounds.Intersects(obs) {
				currentOverlap := s.overlapAmount(mover, obs)
				xOverlap := s.overlapAmount(xOnlyBounds, obs)
				if xOverlap < currentOverlap && (s.terrain == nil || s.terrain.IsPassable(xOnlyBounds)) {
					desiredPos = xOnlyPos
					newBounds.Pos = desiredPos
					continue
				}
			}
			yOnlyPos := emath.Vec2{X: mover.Pos.X, Y: desiredPos.Y}
			yOnlyBounds := emath.Rect{Pos: yOnlyPos, Size: mover.Size}
			if !yOnlyBounds.Intersects(obs) && (s.terrain == nil || s.terrain.IsPassable(yOnlyBounds)) {
				desiredPos = yOnlyPos
				newBounds.Pos = desiredPos
				continue
			}
			if currentlyOverlapping && yOnlyBounds.Intersects(obs) {
				currentOverlap := s.overlapAmount(mover, obs)
				yOverlap := s.overlapAmount(yOnlyBounds, obs)
				if yOverlap < currentOverlap && (s.terrain == nil || s.terrain.IsPassable(yOnlyBounds)) {
					desiredPos = yOnlyPos
					newBounds.Pos = desiredPos
					continue
				}
			}
			return mover.Pos
		}
	}
	return desiredPos
}
func (s *System) overlapAmount(a, b emath.Rect) float64 {
	overlapX := min(a.Pos.X+a.Size.X, b.Pos.X+b.Size.X) - max(a.Pos.X, b.Pos.X)
	overlapY := min(a.Pos.Y+a.Size.Y, b.Pos.Y+b.Size.Y) - max(a.Pos.Y, b.Pos.Y)
	if overlapX <= 0 || overlapY <= 0 {
		return 0
	}
	return overlapX * overlapY
}
func (s *System) clampToWorld(pos emath.Vec2, size emath.Vec2) emath.Vec2 {
	if pos.X < s.worldBounds.Pos.X {
		pos.X = s.worldBounds.Pos.X
	}
	if pos.Y < s.worldBounds.Pos.Y {
		pos.Y = s.worldBounds.Pos.Y
	}
	if pos.X+size.X > s.worldBounds.Pos.X+s.worldBounds.Size.X {
		pos.X = s.worldBounds.Pos.X + s.worldBounds.Size.X - size.X
	}
	if pos.Y+size.Y > s.worldBounds.Pos.Y+s.worldBounds.Size.Y {
		pos.Y = s.worldBounds.Pos.Y + s.worldBounds.Size.Y - size.Y
	}
	return pos
}

// CalculateAvoidanceDirection finds an alternative movement direction when blocked.
// It tries multiple angles offset from the desired direction and returns the best passable one.
func (s *System) CalculateAvoidanceDirection(mover emath.Rect, target emath.Vec2, speed float64, obstacles []emath.Rect) emath.Vec2 {
	moverCenter := emath.Vec2{X: mover.Pos.X + mover.Size.X/2, Y: mover.Pos.Y + mover.Size.Y/2}
	toTarget := target.Sub(moverCenter)
	distToTarget := toTarget.Length()
	if distToTarget < 1 {
		return mover.Pos
	}

	baseAngle := math.Atan2(toTarget.Y, toTarget.X)

	// Try angles in alternating pattern: 0, +30, -30, +60, -60, +90, -90, etc.
	angles := []float64{0}
	for offset := math.Pi / 6; offset <= math.Pi; offset += math.Pi / 6 {
		angles = append(angles, offset, -offset)
	}

	for _, angleOffset := range angles {
		testAngle := baseAngle + angleOffset
		velocity := emath.Vec2{
			X: math.Cos(testAngle) * speed,
			Y: math.Sin(testAngle) * speed,
		}
		testPos := mover.Pos.Add(velocity)
		testPos = s.clampToWorld(testPos, mover.Size)
		testBounds := emath.Rect{Pos: testPos, Size: mover.Size}

		// Check terrain
		if s.terrain != nil && !s.terrain.IsPassable(testBounds) {
			continue
		}

		// Check obstacles
		blocked := false
		for _, obs := range obstacles {
			if testBounds.Intersects(obs) {
				blocked = true
				break
			}
		}

		if !blocked {
			return testPos
		}
	}

	// No valid direction found, stay in place
	return mover.Pos
}
