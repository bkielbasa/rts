package camera

import (
	emath "github.com/bklimczak/tanks/engine/math"
)

const (
	edgeScrollSpeed   = 8.0
	edgeScrollMargin  = 20.0
	keyScrollSpeed    = 10.0
)

// Camera handles viewport into the game world
type Camera struct {
	// Position is the top-left corner of the viewport in world coordinates
	Position    emath.Vec2
	// ViewportSize is the size of the visible area (screen size)
	ViewportSize emath.Vec2
	// WorldSize is the total size of the game world
	WorldSize   emath.Vec2
}

// New creates a new camera
func New(worldWidth, worldHeight, viewportWidth, viewportHeight float64) *Camera {
	return &Camera{
		Position:     emath.Vec2{X: 0, Y: 0},
		ViewportSize: emath.Vec2{X: viewportWidth, Y: viewportHeight},
		WorldSize:    emath.Vec2{X: worldWidth, Y: worldHeight},
	}
}

// SetViewportSize updates the viewport size (when window resizes)
func (c *Camera) SetViewportSize(width, height float64) {
	c.ViewportSize = emath.Vec2{X: width, Y: height}
	c.clampPosition()
}

// SetWorldSize updates the world size
func (c *Camera) SetWorldSize(width, height float64) {
	c.WorldSize = emath.Vec2{X: width, Y: height}
	c.clampPosition()
}

// Move moves the camera by a delta
func (c *Camera) Move(dx, dy float64) {
	c.Position.X += dx
	c.Position.Y += dy
	c.clampPosition()
}

// MoveTo moves the camera to center on a world position
func (c *Camera) MoveTo(worldPos emath.Vec2) {
	c.Position.X = worldPos.X - c.ViewportSize.X/2
	c.Position.Y = worldPos.Y - c.ViewportSize.Y/2
	c.clampPosition()
}

// clampPosition ensures camera stays within world bounds
func (c *Camera) clampPosition() {
	// Clamp X
	if c.Position.X < 0 {
		c.Position.X = 0
	}
	maxX := c.WorldSize.X - c.ViewportSize.X
	if maxX < 0 {
		maxX = 0
	}
	if c.Position.X > maxX {
		c.Position.X = maxX
	}

	// Clamp Y
	if c.Position.Y < 0 {
		c.Position.Y = 0
	}
	maxY := c.WorldSize.Y - c.ViewportSize.Y
	if maxY < 0 {
		maxY = 0
	}
	if c.Position.Y > maxY {
		c.Position.Y = maxY
	}
}

// WorldToScreen converts world coordinates to screen coordinates
func (c *Camera) WorldToScreen(worldPos emath.Vec2) emath.Vec2 {
	return emath.Vec2{
		X: worldPos.X - c.Position.X,
		Y: worldPos.Y - c.Position.Y,
	}
}

// ScreenToWorld converts screen coordinates to world coordinates
func (c *Camera) ScreenToWorld(screenPos emath.Vec2) emath.Vec2 {
	return emath.Vec2{
		X: screenPos.X + c.Position.X,
		Y: screenPos.Y + c.Position.Y,
	}
}

// GetViewportBounds returns the visible area in world coordinates
func (c *Camera) GetViewportBounds() emath.Rect {
	return emath.Rect{
		Pos:  c.Position,
		Size: c.ViewportSize,
	}
}

// IsVisible checks if a world rectangle is visible in the viewport
func (c *Camera) IsVisible(bounds emath.Rect) bool {
	viewport := c.GetViewportBounds()
	return viewport.Intersects(bounds)
}

// HandleEdgeScroll scrolls camera when mouse is near screen edges
// topOffset is the height of UI elements at the top (resource bar)
// leftOffset is the width of UI elements on the left (command panel)
func (c *Camera) HandleEdgeScroll(mouseX, mouseY float64, topOffset, leftOffset float64) {
	// Don't scroll if mouse is in UI areas
	if mouseY < topOffset || mouseX < leftOffset {
		return
	}

	// Right edge
	if mouseX > c.ViewportSize.X-edgeScrollMargin {
		c.Move(edgeScrollSpeed, 0)
	}
	// Left edge (but not in command panel area)
	if mouseX < leftOffset+edgeScrollMargin && mouseX >= leftOffset {
		c.Move(-edgeScrollSpeed, 0)
	}
	// Bottom edge
	if mouseY > c.ViewportSize.Y-edgeScrollMargin {
		c.Move(0, edgeScrollSpeed)
	}
	// Top edge (but not in resource bar area)
	if mouseY < topOffset+edgeScrollMargin && mouseY >= topOffset {
		c.Move(0, -edgeScrollSpeed)
	}
}

// HandleKeyScroll scrolls camera based on arrow keys or WASD
func (c *Camera) HandleKeyScroll(up, down, left, right bool) {
	if up {
		c.Move(0, -keyScrollSpeed)
	}
	if down {
		c.Move(0, keyScrollSpeed)
	}
	if left {
		c.Move(-keyScrollSpeed, 0)
	}
	if right {
		c.Move(keyScrollSpeed, 0)
	}
}
