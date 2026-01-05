package camera

import (
	emath "github.com/bklimczak/tanks/engine/math"
)

const (
	edgeScrollSpeed  = 8.0
	edgeScrollMargin = 20.0
	keyScrollSpeed   = 10.0
)

type Camera struct {
	Position     emath.Vec2
	ViewportSize emath.Vec2
	WorldSize    emath.Vec2
}

func New(worldWidth, worldHeight, viewportWidth, viewportHeight float64) *Camera {
	return &Camera{
		Position:     emath.Vec2{X: 0, Y: 0},
		ViewportSize: emath.Vec2{X: viewportWidth, Y: viewportHeight},
		WorldSize:    emath.Vec2{X: worldWidth, Y: worldHeight},
	}
}
func (c *Camera) SetViewportSize(width, height float64) {
	c.ViewportSize = emath.Vec2{X: width, Y: height}
	c.clampPosition()
}
func (c *Camera) SetWorldSize(width, height float64) {
	c.WorldSize = emath.Vec2{X: width, Y: height}
	c.clampPosition()
}
func (c *Camera) Move(dx, dy float64) {
	c.Position.X += dx
	c.Position.Y += dy
	c.clampPosition()
}
func (c *Camera) MoveTo(worldPos emath.Vec2) {
	c.Position.X = worldPos.X - c.ViewportSize.X/2
	c.Position.Y = worldPos.Y - c.ViewportSize.Y/2
	c.clampPosition()
}
func (c *Camera) clampPosition() {
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
func (c *Camera) WorldToScreen(worldPos emath.Vec2) emath.Vec2 {
	return emath.Vec2{
		X: worldPos.X - c.Position.X,
		Y: worldPos.Y - c.Position.Y,
	}
}
func (c *Camera) ScreenToWorld(screenPos emath.Vec2) emath.Vec2 {
	return emath.Vec2{
		X: screenPos.X + c.Position.X,
		Y: screenPos.Y + c.Position.Y,
	}
}
func (c *Camera) GetViewportBounds() emath.Rect {
	return emath.Rect{
		Pos:  c.Position,
		Size: c.ViewportSize,
	}
}
func (c *Camera) IsVisible(bounds emath.Rect) bool {
	viewport := c.GetViewportBounds()
	return viewport.Intersects(bounds)
}
func (c *Camera) HandleEdgeScroll(mouseX, mouseY float64, topOffset, leftOffset float64) {
	if mouseY < topOffset || mouseX < leftOffset {
		return
	}
	if mouseX > c.ViewportSize.X-edgeScrollMargin {
		c.Move(edgeScrollSpeed, 0)
	}
	if mouseX < leftOffset+edgeScrollMargin && mouseX >= leftOffset {
		c.Move(-edgeScrollSpeed, 0)
	}
	if mouseY > c.ViewportSize.Y-edgeScrollMargin {
		c.Move(0, edgeScrollSpeed)
	}
	if mouseY < topOffset+edgeScrollMargin && mouseY >= topOffset {
		c.Move(0, -edgeScrollSpeed)
	}
}
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
