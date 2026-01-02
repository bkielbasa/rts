package main

// Collider represents any object that can participate in collision detection
type Collider interface {
	// GetBounds returns the collision bounding box (x, y, width, height)
	GetBounds() (x, y, width, height int)
	// IsBlocking returns true if this object blocks movement
	IsBlocking() bool
}

// Rectangle represents an axis-aligned bounding box
type Rectangle struct {
	X, Y, Width, Height int
}

// NewRectangle creates a new rectangle from bounds
func NewRectangle(x, y, width, height int) Rectangle {
	return Rectangle{X: x, Y: y, Width: width, Height: height}
}

// Intersects checks if two rectangles overlap
func (r Rectangle) Intersects(other Rectangle) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}

// CollisionSystem handles collision detection
type CollisionSystem struct {
	worldWidth  int
	worldHeight int
}

// NewCollisionSystem creates a new collision system
func NewCollisionSystem(worldWidth, worldHeight int) *CollisionSystem {
	return &CollisionSystem{
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
	}
}

// CheckCollision checks if a collider at a given position would collide with any other colliders
func (cs *CollisionSystem) CheckCollision(moving Collider, newX, newY int, obstacles []Collider) bool {
	// Get the bounds at the new position
	_, _, width, height := moving.GetBounds()
	movingRect := NewRectangle(newX, newY, width, height)

	// Check world boundaries
	if newX < 0 || newY < 0 || newX+width > cs.worldWidth || newY+height > cs.worldHeight {
		return true
	}

	// Check collision with other objects
	for _, obstacle := range obstacles {
		if !obstacle.IsBlocking() {
			continue
		}

		obstacleX, obstacleY, obstacleWidth, obstacleHeight := obstacle.GetBounds()
		obstacleRect := NewRectangle(obstacleX, obstacleY, obstacleWidth, obstacleHeight)

		if movingRect.Intersects(obstacleRect) {
			return true
		}
	}

	return false
}

// CanMoveTo checks if a collider can move to a new position
func (cs *CollisionSystem) CanMoveTo(moving Collider, newX, newY int, obstacles []Collider) bool {
	return !cs.CheckCollision(moving, newX, newY, obstacles)
}
