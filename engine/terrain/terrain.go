package terrain

import (
	"image/color"
	"math/rand"

	emath "github.com/bklimczak/tanks/engine/math"
)

// TileType represents the type of terrain
type TileType int

const (
	TileGrass TileType = iota
	TileWater
	TileMetal // Metal deposits for extraction
)

// TileSize is the size of each terrain tile in pixels
const TileSize = 50.0

// Tile represents a single terrain tile
type Tile struct {
	Type       TileType
	Passable   bool    // Can units walk on this?
	Buildable  bool    // Can buildings be placed here?
	HasMetal   bool    // Does this tile have metal deposits?
	MetalAmount float64 // Amount of metal if HasMetal is true
}

// TileColors returns the color for each tile type
func TileColors(t TileType) color.Color {
	switch t {
	case TileGrass:
		return color.RGBA{45, 85, 45, 255} // Dark green
	case TileWater:
		return color.RGBA{30, 60, 120, 255} // Dark blue
	case TileMetal:
		return color.RGBA{80, 80, 95, 255} // Gray with slight blue
	default:
		return color.RGBA{50, 50, 50, 255}
	}
}

// TileColorVariation returns a slightly varied color for visual interest
func TileColorVariation(t TileType, x, y int) color.Color {
	base := TileColors(t)
	r, g, b, a := base.RGBA()

	// Create subtle variation based on position
	variation := ((x * 7) + (y * 13)) % 20
	offset := int32(variation) - 10

	// Convert from 16-bit to 8-bit and apply variation
	r8 := clampColor(int32(r>>8) + offset)
	g8 := clampColor(int32(g>>8) + offset)
	b8 := clampColor(int32(b>>8) + offset)

	return color.RGBA{r8, g8, b8, uint8(a >> 8)}
}

func clampColor(v int32) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// Map represents the terrain grid
type Map struct {
	Width    int // Width in tiles
	Height   int // Height in tiles
	Tiles    [][]Tile
	PixelWidth  float64
	PixelHeight float64
}

// NewMap creates a new terrain map
func NewMap(pixelWidth, pixelHeight float64) *Map {
	width := int(pixelWidth / TileSize)
	height := int(pixelHeight / TileSize)

	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
		for x := range tiles[y] {
			tiles[y][x] = Tile{
				Type:      TileGrass,
				Passable:  true,
				Buildable: true,
			}
		}
	}

	return &Map{
		Width:       width,
		Height:      height,
		Tiles:       tiles,
		PixelWidth:  pixelWidth,
		PixelHeight: pixelHeight,
	}
}

// Generate creates terrain using simple procedural generation
func (m *Map) Generate(seed int64) {
	rng := rand.New(rand.NewSource(seed))

	// Generate water bodies (lakes/rivers)
	m.generateWater(rng)

	// Generate metal deposits
	m.generateMetal(rng)
}

// generateWater creates water bodies
func (m *Map) generateWater(rng *rand.Rand) {
	// Create several lake "seeds" and grow them
	numLakes := 5 + rng.Intn(5)

	for i := 0; i < numLakes; i++ {
		// Random lake center
		centerX := rng.Intn(m.Width)
		centerY := rng.Intn(m.Height)

		// Random lake size
		size := 3 + rng.Intn(8)

		// Grow lake using random walk
		m.growWaterBody(rng, centerX, centerY, size)
	}

	// Create a couple of rivers
	numRivers := 1 + rng.Intn(2)
	for i := 0; i < numRivers; i++ {
		m.generateRiver(rng)
	}
}

// growWaterBody expands water from a center point
func (m *Map) growWaterBody(rng *rand.Rand, startX, startY, size int) {
	// Simple flood fill with randomness
	toProcess := []struct{ x, y int }{{startX, startY}}
	processed := make(map[int]bool)

	tilesPlaced := 0
	maxTiles := size * size

	for len(toProcess) > 0 && tilesPlaced < maxTiles {
		// Pop from queue
		current := toProcess[0]
		toProcess = toProcess[1:]

		key := current.y*m.Width + current.x
		if processed[key] {
			continue
		}
		processed[key] = true

		// Check bounds
		if current.x < 0 || current.x >= m.Width || current.y < 0 || current.y >= m.Height {
			continue
		}

		// Place water
		m.Tiles[current.y][current.x] = Tile{
			Type:      TileWater,
			Passable:  false,
			Buildable: false,
		}
		tilesPlaced++

		// Add neighbors with some randomness
		neighbors := []struct{ x, y int }{
			{current.x - 1, current.y},
			{current.x + 1, current.y},
			{current.x, current.y - 1},
			{current.x, current.y + 1},
		}

		for _, n := range neighbors {
			if rng.Float64() < 0.6 { // 60% chance to expand
				toProcess = append(toProcess, n)
			}
		}
	}
}

// generateRiver creates a winding river
func (m *Map) generateRiver(rng *rand.Rand) {
	// Start from a random edge
	var x, y int
	var dx, dy int

	side := rng.Intn(4)
	switch side {
	case 0: // Top
		x, y = rng.Intn(m.Width), 0
		dx, dy = 0, 1
	case 1: // Bottom
		x, y = rng.Intn(m.Width), m.Height-1
		dx, dy = 0, -1
	case 2: // Left
		x, y = 0, rng.Intn(m.Height)
		dx, dy = 1, 0
	case 3: // Right
		x, y = m.Width-1, rng.Intn(m.Height)
		dx, dy = -1, 0
	}

	riverWidth := 2 + rng.Intn(2)

	// Carve river
	for x >= 0 && x < m.Width && y >= 0 && y < m.Height {
		// Place water in a width around the current point
		for w := -riverWidth / 2; w <= riverWidth/2; w++ {
			tx, ty := x, y
			if dx == 0 {
				tx += w
			} else {
				ty += w
			}

			if tx >= 0 && tx < m.Width && ty >= 0 && ty < m.Height {
				m.Tiles[ty][tx] = Tile{
					Type:      TileWater,
					Passable:  false,
					Buildable: false,
				}
			}
		}

		// Move forward
		x += dx
		y += dy

		// Random wandering
		if rng.Float64() < 0.3 {
			if dx == 0 {
				x += rng.Intn(3) - 1
			} else {
				y += rng.Intn(3) - 1
			}
		}
	}
}

// generateMetal places metal deposits
func (m *Map) generateMetal(rng *rand.Rand) {
	// Place metal deposits in clusters
	numDeposits := 10 + rng.Intn(10)

	for i := 0; i < numDeposits; i++ {
		// Find a valid grass location
		attempts := 0
		for attempts < 100 {
			x := rng.Intn(m.Width)
			y := rng.Intn(m.Height)

			if m.Tiles[y][x].Type == TileGrass {
				// Place metal deposit (1-3 tiles)
				clusterSize := 1 + rng.Intn(3)
				m.placeMetalCluster(rng, x, y, clusterSize)
				break
			}
			attempts++
		}
	}
}

// placeMetalCluster places a cluster of metal tiles
func (m *Map) placeMetalCluster(rng *rand.Rand, startX, startY, size int) {
	positions := []struct{ x, y int }{{startX, startY}}

	for i := 0; i < size && len(positions) > 0; i++ {
		// Pick a random position from the list
		idx := rng.Intn(len(positions))
		pos := positions[idx]

		// Check bounds and current tile
		if pos.x >= 0 && pos.x < m.Width && pos.y >= 0 && pos.y < m.Height {
			if m.Tiles[pos.y][pos.x].Type == TileGrass {
				m.Tiles[pos.y][pos.x] = Tile{
					Type:        TileMetal,
					Passable:    true,
					Buildable:   true,
					HasMetal:    true,
					MetalAmount: 1000 + rng.Float64()*2000, // 1000-3000 metal
				}

				// Add neighbors as candidates
				neighbors := []struct{ x, y int }{
					{pos.x - 1, pos.y},
					{pos.x + 1, pos.y},
					{pos.x, pos.y - 1},
					{pos.x, pos.y + 1},
				}
				positions = append(positions, neighbors...)
			}
		}

		// Remove used position
		positions = append(positions[:idx], positions[idx+1:]...)
	}
}

// GetTileAt returns the tile at pixel coordinates
func (m *Map) GetTileAt(pixelX, pixelY float64) *Tile {
	tileX := int(pixelX / TileSize)
	tileY := int(pixelY / TileSize)

	if tileX < 0 || tileX >= m.Width || tileY < 0 || tileY >= m.Height {
		return nil
	}

	return &m.Tiles[tileY][tileX]
}

// GetTileCoords converts pixel coordinates to tile coordinates
func (m *Map) GetTileCoords(pixelX, pixelY float64) (int, int) {
	return int(pixelX / TileSize), int(pixelY / TileSize)
}

// GetPixelCoords converts tile coordinates to pixel coordinates (top-left of tile)
func (m *Map) GetPixelCoords(tileX, tileY int) (float64, float64) {
	return float64(tileX) * TileSize, float64(tileY) * TileSize
}

// IsPassable checks if a rectangle area is passable
func (m *Map) IsPassable(bounds emath.Rect) bool {
	// Check all tiles that the bounds overlap
	startX, startY := m.GetTileCoords(bounds.Pos.X, bounds.Pos.Y)
	endX, endY := m.GetTileCoords(bounds.Pos.X+bounds.Size.X, bounds.Pos.Y+bounds.Size.Y)

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
				return false
			}
			if !m.Tiles[y][x].Passable {
				return false
			}
		}
	}
	return true
}

// IsBuildable checks if a rectangle area is buildable
func (m *Map) IsBuildable(bounds emath.Rect) bool {
	startX, startY := m.GetTileCoords(bounds.Pos.X, bounds.Pos.Y)
	endX, endY := m.GetTileCoords(bounds.Pos.X+bounds.Size.X, bounds.Pos.Y+bounds.Size.Y)

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
				return false
			}
			if !m.Tiles[y][x].Buildable {
				return false
			}
		}
	}
	return true
}

// GetVisibleTiles returns the range of tiles visible in the viewport
func (m *Map) GetVisibleTiles(camPosX, camPosY, viewportWidth, viewportHeight float64) (startX, startY, endX, endY int) {
	startX = int(camPosX / TileSize)
	startY = int(camPosY / TileSize)
	endX = int((camPosX + viewportWidth) / TileSize) + 1
	endY = int((camPosY + viewportHeight) / TileSize) + 1

	// Clamp to map bounds
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX > m.Width {
		endX = m.Width
	}
	if endY > m.Height {
		endY = m.Height
	}

	return
}

// PlaceMetalDeposit places a metal deposit at the given pixel coordinates
// Returns true if successful, false if the location is not suitable (e.g., water)
func (m *Map) PlaceMetalDeposit(pixelX, pixelY float64) bool {
	tileX, tileY := m.GetTileCoords(pixelX, pixelY)

	if tileX < 0 || tileX >= m.Width || tileY < 0 || tileY >= m.Height {
		return false
	}

	tile := &m.Tiles[tileY][tileX]

	// Don't place on water
	if tile.Type == TileWater {
		return false
	}

	// Convert to metal tile
	tile.Type = TileMetal
	tile.Passable = true
	tile.Buildable = true
	tile.HasMetal = true
	tile.MetalAmount = 2000

	return true
}
