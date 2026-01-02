package main

// Camera represents the viewport into the game world
type Camera struct {
	X, Y              int // Camera position in world coordinates
	ViewWidth         int // Width of the viewport (screen width)
	ViewHeight        int // Height of the viewport (screen height)
	WorldWidth        int // Total world width in pixels
	WorldHeight       int // Total world height in pixels
}

// NewCamera creates a new camera
func NewCamera(viewWidth, viewHeight, worldWidth, worldHeight int) *Camera {
	return &Camera{
		X:           0,
		Y:           0,
		ViewWidth:   viewWidth,
		ViewHeight:  viewHeight,
		WorldWidth:  worldWidth,
		WorldHeight: worldHeight,
	}
}

// CenterOn centers the camera on a specific world position
func (c *Camera) CenterOn(x, y int) {
	c.X = x - c.ViewWidth/2
	c.Y = y - c.ViewHeight/2

	// Clamp camera to world bounds
	c.clampToWorld()
}

// clampToWorld ensures the camera doesn't show area outside the world
func (c *Camera) clampToWorld() {
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.X+c.ViewWidth > c.WorldWidth {
		c.X = c.WorldWidth - c.ViewWidth
	}
	if c.Y+c.ViewHeight > c.WorldHeight {
		c.Y = c.WorldHeight - c.ViewHeight
	}

	// If the world is smaller than the view, center it
	if c.WorldWidth < c.ViewWidth {
		c.X = -(c.ViewWidth - c.WorldWidth) / 2
	}
	if c.WorldHeight < c.ViewHeight {
		c.Y = -(c.ViewHeight - c.WorldHeight) / 2
	}
}

// WorldToScreen converts world coordinates to screen coordinates
func (c *Camera) WorldToScreen(worldX, worldY int) (screenX, screenY int) {
	return worldX - c.X, worldY - c.Y
}

// ScreenToWorld converts screen coordinates to world coordinates
func (c *Camera) ScreenToWorld(screenX, screenY int) (worldX, worldY int) {
	return screenX + c.X, screenY + c.Y
}
