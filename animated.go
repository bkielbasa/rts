package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type animatedSprite struct {
    spriteSheet  *ebiten.Image
    frameCounter int
    currentFrame int
    numFrames    int
    distanceBetweenFrames    int
		// offset locaiton of the sprite in the image
		xOffset int
		yOffset int

		size size
}


type animatedSpriteOption func (ss *animatedSprite)

func animatedSpriteOptSize(s size) func (ss *animatedSprite) {
	return func(ss *animatedSprite) {
		ss.size = s
	}
}

func animatedSpriteOptXOffset(o int) func (ss *animatedSprite) {
	return func(ss *animatedSprite) {
		ss.xOffset = o
	}
}

func animatedSpriteOptYOffset(o int) func (ss *animatedSprite) {
	return func(ss *animatedSprite) {
		ss.yOffset = o
	}
}


func newAnimatedSprite(img *ebiten.Image, opts ...animatedSpriteOption) *animatedSprite {
    a := &animatedSprite{
        spriteSheet:  img,
        frameCounter: 0,
        currentFrame: 0,
				size: size{
					width: 28,
					height: 25,
				},
        numFrames:    3,
				distanceBetweenFrames: 4,
    }

		for _, opt := range opts {
			opt(a)
		}

    return a
}

type staticSprite struct {
	sprite *animatedSprite
	position position
	size size

	yOffset int
	xOffset int
	blocking bool
}

type size struct {
	width, height int
}

type staticSpriteOption func (ss *staticSprite)

func staticSpriteOptSize(s size) func (ss *staticSprite) {
	return func(ss *staticSprite) {
		ss.size = s
	}
}

func staticSpriteOptXOffset(o int) func (ss *staticSprite) {
	return func(ss *staticSprite) {
		ss.xOffset = o
	}
}

func staticSpriteOptYOffset(o int) func (ss *staticSprite) {
	return func(ss *staticSprite) {
		ss.yOffset = o
	}
}

func staticSpriteOptBlocking(blocking bool) func (ss *staticSprite) {
	return func(ss *staticSprite) {
		ss.blocking = blocking
	}
}

func newStaticSprite(img *ebiten.Image, x, y int, opts ...staticSpriteOption) *staticSprite {
	ss := &staticSprite{
		position: position{
				x: x,
				y: y,
		},
	}

	for _, opt := range opts {
		opt(ss)
	}

	ss.sprite = newAnimatedSprite(img, 
		animatedSpriteOptSize(ss.size),
		animatedSpriteOptXOffset(ss.xOffset),
		animatedSpriteOptYOffset(ss.yOffset),
	)

	return ss
}

func (ss *staticSprite) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(float64(ss.position.x), float64(ss.position.y))
  screen.DrawImage(ss.currentFrame(), op)
}

func (ss *staticSprite) currentFrame() *ebiten.Image {
	return ss.sprite.sprite()
}

// GetBounds returns the collision bounding box for the static sprite
func (ss *staticSprite) GetBounds() (x, y, width, height int) {
	return ss.position.x, ss.position.y, ss.size.width, ss.size.height
}

// IsBlocking returns whether this sprite blocks movement
func (ss *staticSprite) IsBlocking() bool {
	return ss.blocking
}

func (a *animatedSprite) update() {
    a.frameCounter++
    
    if a.frameCounter >= 10 {
        a.currentFrame = (a.currentFrame + 1) % a.numFrames
        a.frameCounter = 0
    }
}

func (a *animatedSprite) sprite() *ebiten.Image {
		offset := a.currentFrame*a.distanceBetweenFrames
    x := (a.currentFrame * a.size.width)
		x += offset + a.xOffset

    y := a.yOffset
    
    return a.spriteSheet.SubImage(
        image.Rect(x, y, x+a.size.width, y+a.size.height),
    ).(*ebiten.Image)
}

type animatedUnit struct {
	sprite *animatedSprite

	downOffset int
	upOffset int
	rightOffset int
	leftOffset int

	direction direction
	speedPerFrame int
	position position

	health int
	maxHealth int
}

type position struct {
	x, y int
}

type direction int

const (
	directionDown direction = iota + 1
	directionUp
	directionLeft
	directionRight
)

func newAnimatedUnit(img *ebiten.Image) *animatedUnit {
	au := animatedUnit{
		sprite: newAnimatedSprite(img),
		downOffset: 0,
		leftOffset: 34,
		rightOffset: 66,
		upOffset: 92,
		speedPerFrame: 2,
		direction: directionDown,
		health: 100,
		maxHealth: 100,
	}

	return &au
}

type enemyAI struct {
	fireCooldown int
}

// newEnemyUnit creates a new enemy unit at a specific position
func newEnemyUnit(img *ebiten.Image, x, y int) *animatedUnit {
	au := newAnimatedUnit(img)
	au.position.x = x
	au.position.y = y
	au.speedPerFrame = 1 // Slower than player
	au.health = 20 // Enemies have less health
	au.maxHealth = 20
	return au
}

// CanFire returns true if enough time has passed since last fire
func (ai *enemyAI) CanFire() bool {
	return ai.fireCooldown <= 0
}

// ResetFireCooldown sets the fire cooldown
func (ai *enemyAI) ResetFireCooldown() {
	ai.fireCooldown = 90 // 1.5 seconds at 60 FPS
}

// UpdateCooldown decreases fire cooldown
func (ai *enemyAI) UpdateCooldown() {
	if ai.fireCooldown > 0 {
		ai.fireCooldown--
	}
}


func (au *animatedUnit) setDirection(d direction) {
	au.direction = d

	switch d {
		case directionLeft:
			au.sprite.yOffset = au.leftOffset
		case directionRight:
			au.sprite.yOffset = au.rightOffset
		case directionUp:
			au.sprite.yOffset = au.upOffset
		case directionDown:
			au.sprite.yOffset = au.downOffset
	}
}

func (au *animatedUnit) update() {
	au.sprite.update()
}

// tryMove attempts to move the unit in the given direction with collision checking
func (au *animatedUnit) tryMove(d direction, collisionSystem *CollisionSystem, obstacles []Collider) {
	newX, newY := au.position.x, au.position.y

	switch d {
		case directionLeft:
			newX -= au.speedPerFrame
		case directionRight:
			newX += au.speedPerFrame
		case directionUp:
			newY -= au.speedPerFrame
		case directionDown:
			newY += au.speedPerFrame
	}

	// Check if the new position is valid
	if collisionSystem.CanMoveTo(au, newX, newY, obstacles) {
		au.position.x = newX
		au.position.y = newY
	}
}

func (a *animatedUnit) currentFrame() *ebiten.Image {
	return a.sprite.sprite()
}

func (au *animatedUnit) Draw(screen *ebiten.Image, camera *Camera) {
	op := &ebiten.DrawImageOptions{}
	// Convert world position to screen position
	screenX, screenY := camera.WorldToScreen(au.position.x, au.position.y)
	op.GeoM.Translate(float64(screenX), float64(screenY))
	screen.DrawImage(au.currentFrame(), op)
}

// DrawHealthBar draws a health bar above the unit
func (au *animatedUnit) DrawHealthBar(screen *ebiten.Image, camera *Camera) {
	screenX, screenY := camera.WorldToScreen(au.position.x, au.position.y)

	// Health bar dimensions
	barWidth := 28 // Same as tank width
	barHeight := 4
	barX := float32(screenX)
	barY := float32(screenY - 6) // Above the tank

	// Background (red)
	vector.DrawFilledRect(screen, barX, barY, float32(barWidth), float32(barHeight), color.RGBA{100, 0, 0, 255}, false)

	// Health (green)
	healthWidth := float32(barWidth) * float32(au.health) / float32(au.maxHealth)
	if healthWidth > 0 {
		vector.DrawFilledRect(screen, barX, barY, healthWidth, float32(barHeight), color.RGBA{0, 200, 0, 255}, false)
	}

	// Border
	vector.StrokeRect(screen, barX, barY, float32(barWidth), float32(barHeight), 1, color.RGBA{255, 255, 255, 255}, false)
}

// GetBounds returns the collision bounding box for the animated unit
func (au *animatedUnit) GetBounds() (x, y, width, height int) {
	return au.position.x, au.position.y, au.sprite.size.width, au.sprite.size.height
}

// IsBlocking returns true since animated units block movement
func (au *animatedUnit) IsBlocking() bool {
	return true
}

// moveTowards moves the unit toward a target position with collision checking
func (au *animatedUnit) moveTowards(targetX, targetY int, collisionSystem *CollisionSystem, obstacles []Collider) {
	// Calculate direction to target
	dx := targetX - au.position.x
	dy := targetY - au.position.y

	// Determine primary movement direction
	var moveDir direction
	if abs(dx) > abs(dy) {
		// Move horizontally
		if dx > 0 {
			moveDir = directionRight
		} else {
			moveDir = directionLeft
		}
	} else {
		// Move vertically
		if dy > 0 {
			moveDir = directionDown
		} else {
			moveDir = directionUp
		}
	}

	// Set direction and try to move
	au.setDirection(moveDir)
	au.tryMove(moveDir, collisionSystem, obstacles)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TakeDamage reduces the unit's health
func (au *animatedUnit) TakeDamage(damage int) {
	au.health -= damage
	if au.health < 0 {
		au.health = 0
	}
}

// IsAlive returns true if the unit still has health
func (au *animatedUnit) IsAlive() bool {
	return au.health > 0
}

// Heal restores health to the unit
func (au *animatedUnit) Heal(amount int) {
	au.health += amount
	if au.health > au.maxHealth {
		au.health = au.maxHealth
	}
}
