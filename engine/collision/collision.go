package collision

import (
	emath "github.com/bklimczak/tanks/engine/math"
)

// TerrainChecker interface for checking terrain passability
type TerrainChecker interface {
	IsPassable(bounds emath.Rect) bool
}

// Collidable interface for objects that can collide
type Collidable interface {
	Bounds() emath.Rect
}

// System handles collision detection and resolution
type System struct {
	worldBounds emath.Rect
	terrain     TerrainChecker
}

// NewSystem creates a new collision system
func NewSystem(worldWidth, worldHeight float64) *System {
	return &System{
		worldBounds: emath.NewRect(0, 0, worldWidth, worldHeight),
	}
}

// SetWorldBounds updates the world boundaries
func (s *System) SetWorldBounds(width, height float64) {
	s.worldBounds = emath.NewRect(0, 0, width, height)
}

// SetTerrain sets the terrain checker for passability checks
func (s *System) SetTerrain(t TerrainChecker) {
	s.terrain = t
}

// CheckCollision checks if two rectangles collide
func (s *System) CheckCollision(a, b emath.Rect) bool {
	return a.Intersects(b)
}

// isPositionValid checks if a position is valid (within bounds and passable terrain)
func (s *System) isPositionValid(bounds emath.Rect) bool {
	// Check world bounds
	if bounds.Pos.X < s.worldBounds.Pos.X || bounds.Pos.Y < s.worldBounds.Pos.Y {
		return false
	}
	if bounds.Pos.X+bounds.Size.X > s.worldBounds.Pos.X+s.worldBounds.Size.X {
		return false
	}
	if bounds.Pos.Y+bounds.Size.Y > s.worldBounds.Pos.Y+s.worldBounds.Size.Y {
		return false
	}

	// Check terrain passability
	if s.terrain != nil && !s.terrain.IsPassable(bounds) {
		return false
	}

	return true
}

// ResolveMovement resolves collision for a moving object against obstacles
// Returns the valid position after collision resolution
func (s *System) ResolveMovement(mover emath.Rect, desiredPos emath.Vec2, obstacles []emath.Rect) emath.Vec2 {
	newBounds := emath.Rect{Pos: desiredPos, Size: mover.Size}

	// Check world bounds first
	desiredPos = s.clampToWorld(desiredPos, mover.Size)
	newBounds.Pos = desiredPos

	// Check terrain passability
	if s.terrain != nil && !s.terrain.IsPassable(newBounds) {
		// Try moving only on X axis
		xOnlyPos := emath.Vec2{X: desiredPos.X, Y: mover.Pos.Y}
		xOnlyBounds := emath.Rect{Pos: xOnlyPos, Size: mover.Size}

		if s.terrain.IsPassable(xOnlyBounds) {
			desiredPos = xOnlyPos
			newBounds.Pos = desiredPos
		} else {
			// Try moving only on Y axis
			yOnlyPos := emath.Vec2{X: mover.Pos.X, Y: desiredPos.Y}
			yOnlyBounds := emath.Rect{Pos: yOnlyPos, Size: mover.Size}

			if s.terrain.IsPassable(yOnlyBounds) {
				desiredPos = yOnlyPos
				newBounds.Pos = desiredPos
			} else {
				// Can't move, return original
				return mover.Pos
			}
		}
	}

	// Check against each obstacle
	for _, obs := range obstacles {
		// Check if currently overlapping with this obstacle
		currentlyOverlapping := mover.Intersects(obs)

		if newBounds.Intersects(obs) {
			// If we're already overlapping and the new position reduces overlap, allow it
			if currentlyOverlapping {
				currentOverlap := s.overlapAmount(mover, obs)
				newOverlap := s.overlapAmount(newBounds, obs)
				if newOverlap < currentOverlap {
					// Moving away from obstacle, allow it
					continue
				}
			}

			// Try moving only on X axis
			xOnlyPos := emath.Vec2{X: desiredPos.X, Y: mover.Pos.Y}
			xOnlyBounds := emath.Rect{Pos: xOnlyPos, Size: mover.Size}

			if !xOnlyBounds.Intersects(obs) && (s.terrain == nil || s.terrain.IsPassable(xOnlyBounds)) {
				desiredPos = xOnlyPos
				newBounds.Pos = desiredPos
				continue
			}

			// If currently overlapping, check if X movement reduces overlap
			if currentlyOverlapping && xOnlyBounds.Intersects(obs) {
				currentOverlap := s.overlapAmount(mover, obs)
				xOverlap := s.overlapAmount(xOnlyBounds, obs)
				if xOverlap < currentOverlap && (s.terrain == nil || s.terrain.IsPassable(xOnlyBounds)) {
					desiredPos = xOnlyPos
					newBounds.Pos = desiredPos
					continue
				}
			}

			// Try moving only on Y axis
			yOnlyPos := emath.Vec2{X: mover.Pos.X, Y: desiredPos.Y}
			yOnlyBounds := emath.Rect{Pos: yOnlyPos, Size: mover.Size}

			if !yOnlyBounds.Intersects(obs) && (s.terrain == nil || s.terrain.IsPassable(yOnlyBounds)) {
				desiredPos = yOnlyPos
				newBounds.Pos = desiredPos
				continue
			}

			// If currently overlapping, check if Y movement reduces overlap
			if currentlyOverlapping && yOnlyBounds.Intersects(obs) {
				currentOverlap := s.overlapAmount(mover, obs)
				yOverlap := s.overlapAmount(yOnlyBounds, obs)
				if yOverlap < currentOverlap && (s.terrain == nil || s.terrain.IsPassable(yOnlyBounds)) {
					desiredPos = yOnlyPos
					newBounds.Pos = desiredPos
					continue
				}
			}

			// Can't move at all, return original position
			return mover.Pos
		}
	}

	return desiredPos
}

// overlapAmount calculates the overlap area between two rectangles
func (s *System) overlapAmount(a, b emath.Rect) float64 {
	overlapX := min(a.Pos.X+a.Size.X, b.Pos.X+b.Size.X) - max(a.Pos.X, b.Pos.X)
	overlapY := min(a.Pos.Y+a.Size.Y, b.Pos.Y+b.Size.Y) - max(a.Pos.Y, b.Pos.Y)

	if overlapX <= 0 || overlapY <= 0 {
		return 0
	}
	return overlapX * overlapY
}

// clampToWorld ensures a position stays within world bounds
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
