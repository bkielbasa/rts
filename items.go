package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"math/rand"
)

type HeartItem struct {
	x, y          int
	active        bool
	pulseCounter  int
}

func NewHeartItem(worldWidth, worldHeight int) *HeartItem {
	// Random position on the map
	x := rand.Intn(worldWidth - 32) + 16
	y := rand.Intn(worldHeight - 32) + 16

	return &HeartItem{
		x:      x,
		y:      y,
		active: true,
	}
}

func (h *HeartItem) Update() {
	h.pulseCounter++
}

func (h *HeartItem) Draw(screen *ebiten.Image, camera *Camera) {
	if !h.active {
		return
	}

	screenX, screenY := camera.WorldToScreen(h.x, h.y)

	// Pulsing effect
	scale := 1.0 + 0.2*float32(h.pulseCounter%30)/30.0

	// Draw heart shape
	centerX := float32(screenX + 8)
	centerY := float32(screenY + 8)

	// Simple heart using circles and a triangle
	heartColor := color.RGBA{255, 0, 100, 255}

	// Left circle
	vector.DrawFilledCircle(screen, centerX-3*scale, centerY-2*scale, 4*scale, heartColor, true)
	// Right circle
	vector.DrawFilledCircle(screen, centerX+3*scale, centerY-2*scale, 4*scale, heartColor, true)

	// Bottom triangle (using a filled rect rotated)
	// Draw a simple filled shape for the bottom part
	vector.DrawFilledRect(screen, centerX-6*scale, centerY-1*scale, 12*scale, 8*scale, heartColor, false)

	// Outline for visibility
	vector.StrokeCircle(screen, centerX-3*scale, centerY-2*scale, 4*scale, 1, color.RGBA{200, 0, 80, 255}, true)
	vector.StrokeCircle(screen, centerX+3*scale, centerY-2*scale, 4*scale, 1, color.RGBA{200, 0, 80, 255}, true)
}

func (h *HeartItem) GetBounds() (x, y, width, height int) {
	return h.x, h.y, 16, 16
}

func (h *HeartItem) CheckCollision(unit *animatedUnit) bool {
	if !h.active {
		return false
	}

	heartRect := NewRectangle(h.x, h.y, 16, 16)
	unitX, unitY, unitW, unitH := unit.GetBounds()
	unitRect := NewRectangle(unitX, unitY, unitW, unitH)

	return heartRect.Intersects(unitRect)
}

func (h *HeartItem) Collect() {
	h.active = false
}

func (h *HeartItem) IsActive() bool {
	return h.active
}
