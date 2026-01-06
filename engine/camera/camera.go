package camera

import (
	emath "github.com/bklimczak/tanks/engine/math"
)

const (
	edgeScrollSpeed  = 8.0
	edgeScrollMargin = 20.0
	keyScrollSpeed   = 10.0
	minZoom          = 0.5
	maxZoom          = 2.0
	zoomSpeed        = 0.1
)

type Camera struct {
	Position     emath.Vec2
	ViewportSize emath.Vec2
	WorldSize    emath.Vec2
	Zoom         float64 // 1.0 = normal, < 1.0 = zoomed out, > 1.0 = zoomed in
}

func New(worldWidth, worldHeight, viewportWidth, viewportHeight float64) *Camera {
	return &Camera{
		Position:     emath.Vec2{X: 0, Y: 0},
		ViewportSize: emath.Vec2{X: viewportWidth, Y: viewportHeight},
		WorldSize:    emath.Vec2{X: worldWidth, Y: worldHeight},
		Zoom:         1.0,
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
	// Visible world size depends on zoom (zoomed out = see more world)
	visibleWidth := c.ViewportSize.X / c.Zoom
	visibleHeight := c.ViewportSize.Y / c.Zoom

	if c.Position.X < 0 {
		c.Position.X = 0
	}
	maxX := c.WorldSize.X - visibleWidth
	if maxX < 0 {
		maxX = 0
	}
	if c.Position.X > maxX {
		c.Position.X = maxX
	}
	if c.Position.Y < 0 {
		c.Position.Y = 0
	}
	maxY := c.WorldSize.Y - visibleHeight
	if maxY < 0 {
		maxY = 0
	}
	if c.Position.Y > maxY {
		c.Position.Y = maxY
	}
}
func (c *Camera) WorldToScreen(worldPos emath.Vec2) emath.Vec2 {
	return emath.Vec2{
		X: (worldPos.X - c.Position.X) * c.Zoom,
		Y: (worldPos.Y - c.Position.Y) * c.Zoom,
	}
}
func (c *Camera) ScreenToWorld(screenPos emath.Vec2) emath.Vec2 {
	return emath.Vec2{
		X: screenPos.X/c.Zoom + c.Position.X,
		Y: screenPos.Y/c.Zoom + c.Position.Y,
	}
}
func (c *Camera) GetViewportBounds() emath.Rect {
	return emath.Rect{
		Pos: c.Position,
		Size: emath.Vec2{
			X: c.ViewportSize.X / c.Zoom,
			Y: c.ViewportSize.Y / c.Zoom,
		},
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

// ZoomIn increases the zoom level (things appear larger)
func (c *Camera) ZoomIn(mouseScreenPos emath.Vec2) {
	c.zoomAtPoint(zoomSpeed, mouseScreenPos)
}

// ZoomOut decreases the zoom level (things appear smaller)
func (c *Camera) ZoomOut(mouseScreenPos emath.Vec2) {
	c.zoomAtPoint(-zoomSpeed, mouseScreenPos)
}

// zoomAtPoint zooms centered on a screen position (typically the mouse cursor)
func (c *Camera) zoomAtPoint(delta float64, screenPos emath.Vec2) {
	// Get world position under cursor before zoom
	worldBefore := c.ScreenToWorld(screenPos)

	// Apply zoom
	oldZoom := c.Zoom
	c.Zoom += delta
	if c.Zoom < minZoom {
		c.Zoom = minZoom
	}
	if c.Zoom > maxZoom {
		c.Zoom = maxZoom
	}

	// If zoom didn't change, nothing to do
	if c.Zoom == oldZoom {
		return
	}

	// Get world position under cursor after zoom
	worldAfter := c.ScreenToWorld(screenPos)

	// Adjust camera position to keep the same world point under cursor
	c.Position.X += worldBefore.X - worldAfter.X
	c.Position.Y += worldBefore.Y - worldAfter.Y

	c.clampPosition()
}

// GetZoom returns the current zoom level
func (c *Camera) GetZoom() float64 {
	return c.Zoom
}
