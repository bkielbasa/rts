package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

type BulletOwner int

const (
	BulletOwnerPlayer BulletOwner = iota
	BulletOwnerEnemy
)

type Bullet struct {
	x, y      int
	direction direction
	speed     int
	active    bool
	owner     BulletOwner
}

func NewBullet(x, y int, dir direction, owner BulletOwner) *Bullet {
	return &Bullet{
		x:         x,
		y:         y,
		direction: dir,
		speed:     4,
		active:    true,
		owner:     owner,
	}
}

func (b *Bullet) Update(collisionSystem *CollisionSystem, terrainColliders []Collider) {
	if !b.active {
		return
	}

	// Move bullet
	switch b.direction {
	case directionUp:
		b.y -= b.speed
	case directionDown:
		b.y += b.speed
	case directionLeft:
		b.x -= b.speed
	case directionRight:
		b.x += b.speed
	}

	// Check if bullet is out of bounds
	if b.x < 0 || b.y < 0 || b.x >= collisionSystem.worldWidth || b.y >= collisionSystem.worldHeight {
		b.active = false
		return
	}

	// Check collision with terrain
	bulletRect := NewRectangle(b.x, b.y, 4, 4)
	for _, obstacle := range terrainColliders {
		if !obstacle.IsBlocking() {
			continue
		}
		obstX, obstY, obstW, obstH := obstacle.GetBounds()
		obstRect := NewRectangle(obstX, obstY, obstW, obstH)
		if bulletRect.Intersects(obstRect) {
			b.active = false
			return
		}
	}
}

func (b *Bullet) Draw(screen *ebiten.Image, camera *Camera) {
	if !b.active {
		return
	}

	screenX, screenY := camera.WorldToScreen(b.x, b.y)

	// Draw bullet with different colors based on owner
	bulletColor := color.RGBA{255, 255, 0, 255} // Yellow for player
	if b.owner == BulletOwnerEnemy {
		bulletColor = color.RGBA{255, 100, 0, 255} // Orange for enemy
	}

	vector.DrawFilledCircle(screen, float32(screenX+2), float32(screenY+2), 2, bulletColor, true)
}

func (b *Bullet) GetBounds() (x, y, width, height int) {
	return b.x, b.y, 4, 4
}

func (b *Bullet) IsActive() bool {
	return b.active
}

func (b *Bullet) Deactivate() {
	b.active = false
}

// CheckEnemyCollision checks if the bullet hits an enemy
func (b *Bullet) CheckEnemyCollision(enemy *animatedUnit) bool {
	if !b.active {
		return false
	}

	bulletRect := NewRectangle(b.x, b.y, 4, 4)
	enemyX, enemyY, enemyW, enemyH := enemy.GetBounds()
	enemyRect := NewRectangle(enemyX, enemyY, enemyW, enemyH)

	return bulletRect.Intersects(enemyRect)
}

// CheckBulletCollision checks if this bullet hits another bullet
func (b *Bullet) CheckBulletCollision(other *Bullet) bool {
	if !b.active || !other.active {
		return false
	}

	// Bullets from the same owner don't collide
	if b.owner == other.owner {
		return false
	}

	bulletRect := NewRectangle(b.x, b.y, 4, 4)
	otherRect := NewRectangle(other.x, other.y, 4, 4)

	return bulletRect.Intersects(otherRect)
}

// GetOwner returns the bullet owner
func (b *Bullet) GetOwner() BulletOwner {
	return b.owner
}
