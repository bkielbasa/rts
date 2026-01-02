package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type TerrainType int

const (
	TerrainGrass TerrainType = iota
	TerrainSand
)

const TileSize = 16

// TerrainMap represents the game world terrain
type TerrainMap struct {
	width, height int
	tiles         [][]TerrainType
	spriteSheet   *ebiten.Image
}

// NewTerrainMap creates a new terrain map
func NewTerrainMap(width, height int, spriteSheet *ebiten.Image) *TerrainMap {
	tiles := make([][]TerrainType, height)
	for i := range tiles {
		tiles[i] = make([]TerrainType, width)
		// Default to grass
		for j := range tiles[i] {
			tiles[i][j] = TerrainGrass
		}
	}

	return &TerrainMap{
		width:       width,
		height:      height,
		tiles:       tiles,
		spriteSheet: spriteSheet,
	}
}

// SetTile sets the terrain type at a specific grid position
func (tm *TerrainMap) SetTile(x, y int, terrainType TerrainType) {
	if x >= 0 && x < tm.width && y >= 0 && y < tm.height {
		tm.tiles[y][x] = terrainType
	}
}

// SetRect sets a rectangular area to a specific terrain type
func (tm *TerrainMap) SetRect(x, y, width, height int, terrainType TerrainType) {
	for j := y; j < y+height && j < tm.height; j++ {
		for i := x; i < x+width && i < tm.width; i++ {
			tm.SetTile(i, j, terrainType)
		}
	}
}

// GetTile returns the terrain type at a specific grid position
func (tm *TerrainMap) GetTile(x, y int) TerrainType {
	if x >= 0 && x < tm.width && y >= 0 && y < tm.height {
		return tm.tiles[y][x]
	}
	return TerrainGrass
}

// GetColliders returns all blocking tiles as colliders
func (tm *TerrainMap) GetColliders() []Collider {
	colliders := []Collider{}
	for y := 0; y < tm.height; y++ {
		for x := 0; x < tm.width; x++ {
			if tm.isBlocking(tm.tiles[y][x]) {
				colliders = append(colliders, &TerrainTile{
					x:           x * TileSize,
					y:           y * TileSize,
					terrainType: tm.tiles[y][x],
				})
			}
		}
	}
	return colliders
}

func (tm *TerrainMap) isBlocking(t TerrainType) bool {
	switch t {
	case TerrainSand:
		return true
	default:
		return false
	}
}

// Draw renders the visible portion of the terrain
func (tm *TerrainMap) Draw(screen *ebiten.Image, camera *Camera) {
	// Calculate which tiles are visible
	startX := camera.X / TileSize
	startY := camera.Y / TileSize
	endX := (camera.X + camera.ViewWidth) / TileSize
	endY := (camera.Y + camera.ViewHeight) / TileSize

	// Add padding to avoid edge artifacts
	startX--
	startY--
	endX++
	endY++

	// Clamp to map bounds
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX >= tm.width {
		endX = tm.width - 1
	}
	if endY >= tm.height {
		endY = tm.height - 1
	}

	// Draw visible tiles
	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			tm.drawTile(screen, x, y, camera)
		}
	}
}

func (tm *TerrainMap) drawTile(screen *ebiten.Image, gridX, gridY int, camera *Camera) {
	terrainType := tm.GetTile(gridX, gridY)

	op := &ebiten.DrawImageOptions{}
	// World position
	worldX := gridX * TileSize
	worldY := gridY * TileSize
	// Screen position (world position - camera position)
	screenX := worldX - camera.X
	screenY := worldY - camera.Y

	op.GeoM.Translate(float64(screenX), float64(screenY))

	sprite := tm.getTileSprite(terrainType)
	screen.DrawImage(sprite, op)
}

func (tm *TerrainMap) getTileSprite(terrainType TerrainType) *ebiten.Image {
	opts := []animatedSpriteOption{
		animatedSpriteOptSize(size{TileSize, TileSize}),
	}

	switch terrainType {
	case TerrainSand:
		opts = append(opts, animatedSpriteOptXOffset(16))
	case TerrainGrass:
		// Default, no offset needed
	}

	sprite := newAnimatedSprite(tm.spriteSheet, opts...)
	return sprite.sprite()
}

// TerrainTile represents a single tile for collision detection
type TerrainTile struct {
	x, y        int
	terrainType TerrainType
}

func (tt *TerrainTile) GetBounds() (x, y, width, height int) {
	return tt.x, tt.y, TileSize, TileSize
}

func (tt *TerrainTile) IsBlocking() bool {
	switch tt.terrainType {
	case TerrainSand:
		return true
	default:
		return false
	}
}
