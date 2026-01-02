package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type game struct {
	tank            *animatedUnit
	enemies         []*animatedUnit
	enemyAI         []*enemyAI
	bullets         []*Bullet
	heartItem       *HeartItem
	terrainMap      *TerrainMap
	camera          *Camera
	collisionSystem *CollisionSystem
	fireCooldown    int
}

func newGame() *game {
	const (
		screenWidth  = 320
		screenHeight = 240
		mapWidth     = 50 // tiles
		mapHeight    = 50 // tiles
	)

	worldWidth := mapWidth * TileSize
	worldHeight := mapHeight * TileSize

	tankSprite, _, err := ebitenutil.NewImageFromFile("tank2.png")
	if err != nil {
		panic("cannot read tank sprite")
	}

	enemySprite, _, err := ebitenutil.NewImageFromFile("tank1.png")
	if err != nil {
		panic("cannot read enemy sprite")
	}

	terrainSprite, _, err := ebitenutil.NewImageFromFile("terrain.png")
	if err != nil {
		panic("cannot read terrain sprite")
	}

	// Create terrain map
	terrainMap := NewTerrainMap(mapWidth, mapHeight, terrainSprite)

	// Add some sand obstacles
	terrainMap.SetRect(10, 10, 10, 5, TerrainSand)
	terrainMap.SetRect(25, 15, 8, 8, TerrainSand)
	terrainMap.SetRect(5, 30, 15, 3, TerrainSand)
	terrainMap.SetRect(35, 25, 5, 10, TerrainSand)

	// Create tank in the center of the world
	tank := newAnimatedUnit(tankSprite)
	tank.position.x = worldWidth / 2
	tank.position.y = worldHeight / 2

	// Create enemies at various positions
	enemies := []*animatedUnit{
		newEnemyUnit(enemySprite, 10*TileSize, 5*TileSize),
		newEnemyUnit(enemySprite, 30*TileSize, 10*TileSize),
		newEnemyUnit(enemySprite, 15*TileSize, 25*TileSize),
		newEnemyUnit(enemySprite, 40*TileSize, 20*TileSize),
		newEnemyUnit(enemySprite, 8*TileSize, 35*TileSize),
		newEnemyUnit(enemySprite, 35*TileSize, 40*TileSize),
	}

	// Create AI for each enemy
	var enemyAIList []*enemyAI
	for range enemies {
		enemyAIList = append(enemyAIList, &enemyAI{})
	}

	// Create heart item
	heartItem := NewHeartItem(worldWidth, worldHeight)

	g := &game{
		tank:            tank,
		enemies:         enemies,
		enemyAI:         enemyAIList,
		heartItem:       heartItem,
		terrainMap:      terrainMap,
		camera:          NewCamera(screenWidth, screenHeight, worldWidth, worldHeight),
		collisionSystem: NewCollisionSystem(worldWidth, worldHeight),
	}

	// Center camera on tank
	g.camera.CenterOn(g.tank.position.x+14, g.tank.position.y+12)

	return g
}

func (g *game) fireBullet() {
	// Check cooldown
	if g.fireCooldown > 0 {
		return
	}

	// Calculate bullet spawn position based on tank direction
	bulletX := g.tank.position.x + 14 // Tank center X
	bulletY := g.tank.position.y + 12 // Tank center Y

	// Offset bullet position based on tank direction
	switch g.tank.direction {
	case directionUp:
		bulletY -= 10
	case directionDown:
		bulletY += 10
	case directionLeft:
		bulletX -= 10
	case directionRight:
		bulletX += 10
	}

	bullet := NewBullet(bulletX, bulletY, g.tank.direction, BulletOwnerPlayer)
	g.bullets = append(g.bullets, bullet)

	// Set cooldown (15 frames = ~0.25 seconds at 60 FPS)
	g.fireCooldown = 15
}

func (g *game) Update() error {
	// Decrease fire cooldown
	if g.fireCooldown > 0 {
		g.fireCooldown--
	}

	// Get colliders from terrain
	obstacles := g.terrainMap.GetColliders()

	// Add enemies as obstacles
	for _, enemy := range g.enemies {
		obstacles = append(obstacles, enemy)
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.tank.setDirection(directionUp)
		g.tank.tryMove(directionUp, g.collisionSystem, obstacles)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.tank.setDirection(directionRight)
		g.tank.tryMove(directionRight, g.collisionSystem, obstacles)
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.tank.setDirection(directionDown)
		g.tank.tryMove(directionDown, g.collisionSystem, obstacles)
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.tank.setDirection(directionLeft)
		g.tank.tryMove(directionLeft, g.collisionSystem, obstacles)
	}

	// Fire bullet with space key
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.fireBullet()
	}

	g.tank.update()

	// Update enemies - make them follow the tank and fire
	for i, enemy := range g.enemies {
		// Build obstacles list excluding the current enemy
		enemyObstacles := g.terrainMap.GetColliders()
		enemyObstacles = append(enemyObstacles, g.tank)

		// Add other enemies as obstacles
		for j, otherEnemy := range g.enemies {
			if i != j {
				enemyObstacles = append(enemyObstacles, otherEnemy)
			}
		}

		// Move toward tank center
		tankCenterX := g.tank.position.x + 14
		tankCenterY := g.tank.position.y + 12
		enemy.moveTowards(tankCenterX, tankCenterY, g.collisionSystem, enemyObstacles)
		enemy.update()

		// Enemy AI - firing
		ai := g.enemyAI[i]
		ai.UpdateCooldown()
		if ai.CanFire() {
			// Fire bullet toward player
			bulletX := enemy.position.x + 14
			bulletY := enemy.position.y + 12

			// Offset based on enemy direction
			switch enemy.direction {
			case directionUp:
				bulletY -= 10
			case directionDown:
				bulletY += 10
			case directionLeft:
				bulletX -= 10
			case directionRight:
				bulletX += 10
			}

			enemyBullet := NewBullet(bulletX, bulletY, enemy.direction, BulletOwnerEnemy)
			g.bullets = append(g.bullets, enemyBullet)
			ai.ResetFireCooldown()
		}
	}

	// Update bullets
	terrainColliders := g.terrainMap.GetColliders()
	activeBullets := []*Bullet{}
	for _, bullet := range g.bullets {
		bullet.Update(g.collisionSystem, terrainColliders)
		if bullet.IsActive() {
			activeBullets = append(activeBullets, bullet)
		}
	}
	g.bullets = activeBullets

	// Check bullet-bullet collisions
	for i := 0; i < len(g.bullets); i++ {
		for j := i + 1; j < len(g.bullets); j++ {
			if g.bullets[i].CheckBulletCollision(g.bullets[j]) {
				g.bullets[i].Deactivate()
				g.bullets[j].Deactivate()
			}
		}
	}

	// Check bullet-enemy collisions (only player bullets hurt enemies)
	for _, bullet := range g.bullets {
		if bullet.GetOwner() == BulletOwnerPlayer {
			for i := len(g.enemies) - 1; i >= 0; i-- {
				enemy := g.enemies[i]
				if bullet.CheckEnemyCollision(enemy) {
					// Damage enemy
					enemy.TakeDamage(20)
					// Remove enemy if dead
					if !enemy.IsAlive() {
						g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
						g.enemyAI = append(g.enemyAI[:i], g.enemyAI[i+1:]...)
					}
					// Deactivate bullet
					bullet.Deactivate()
					break
				}
			}
		}
	}

	// Check enemy bullet-player collisions
	for _, bullet := range g.bullets {
		if bullet.GetOwner() == BulletOwnerEnemy {
			if bullet.CheckEnemyCollision(g.tank) {
				// Damage player
				g.tank.TakeDamage(20)
				// Deactivate bullet
				bullet.Deactivate()
			}
		}
	}

	// Update heart item
	if g.heartItem != nil && g.heartItem.IsActive() {
		g.heartItem.Update()

		// Check if player collects the heart
		if g.heartItem.CheckCollision(g.tank) {
			// Heal player by 50%
			healAmount := g.tank.maxHealth / 2
			g.tank.Heal(healAmount)
			g.heartItem.Collect()

			// Spawn a new heart at a random location
			g.heartItem = NewHeartItem(g.collisionSystem.worldWidth, g.collisionSystem.worldHeight)
		}
	}

	// Update camera to follow tank (center on tank sprite center)
	tankCenterX := g.tank.position.x + 14
	tankCenterY := g.tank.position.y + 12
	g.camera.CenterOn(tankCenterX, tankCenterY)

	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	// Draw terrain with camera offset
	g.terrainMap.Draw(screen, g.camera)

	// Draw heart item
	if g.heartItem != nil && g.heartItem.IsActive() {
		g.heartItem.Draw(screen, g.camera)
	}

	// Draw enemies with camera offset
	for _, enemy := range g.enemies {
		enemy.Draw(screen, g.camera)
	}

	// Draw bullets
	for _, bullet := range g.bullets {
		bullet.Draw(screen, g.camera)
	}

	// Draw tank with camera offset (on top of enemies)
	g.tank.Draw(screen, g.camera)

	// Draw player health bar above tank
	g.tank.DrawHealthBar(screen, g.camera)
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Tanks!")
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
