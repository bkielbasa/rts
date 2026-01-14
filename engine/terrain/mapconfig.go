package terrain

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MapConfig represents a complete map configuration with factions and entities
type MapConfig struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description,omitempty"`
	Author      string          `yaml:"author,omitempty"`
	Version     string          `yaml:"version,omitempty"`
	Size        MapSize         `yaml:"size"`
	Terrain     *TerrainConfig  `yaml:"terrain,omitempty"`
	Factions    []FactionConfig `yaml:"factions"`
}

// MapSize defines the map dimensions
type MapSize struct {
	Width    int     `yaml:"width"`               // Width in tiles
	Height   int     `yaml:"height"`              // Height in tiles
	TileSize float64 `yaml:"tile_size,omitempty"` // Pixels per tile (default: 25)
}

// TerrainConfig defines terrain features (water, metal deposits, etc.)
// By default, everything is grass
type TerrainConfig struct {
	Water []WaterFeature `yaml:"water,omitempty"`
	Metal []MetalDeposit `yaml:"metal,omitempty"`
}

// WaterFeature defines a water body
type WaterFeature struct {
	Type   string `yaml:"type"` // "lake", "river"
	X      int    `yaml:"x"`
	Y      int    `yaml:"y"`
	Width  int    `yaml:"width,omitempty"`  // For lakes
	Height int    `yaml:"height,omitempty"` // For lakes
	EndX   int    `yaml:"end_x,omitempty"`  // For rivers
	EndY   int    `yaml:"end_y,omitempty"`  // For rivers
	Size   int    `yaml:"size,omitempty"`   // Thickness for rivers, radius for lakes
}

// MetalDeposit defines a metal resource location
type MetalDeposit struct {
	X      int     `yaml:"x"`
	Y      int     `yaml:"y"`
	Amount float64 `yaml:"amount,omitempty"` // Default: 2000
}

// FactionConfig groups all entities belonging to a faction/team
type FactionConfig struct {
	ID        string           `yaml:"id"`             // e.g., "player", "enemy_1", "neutral"
	Team      string           `yaml:"team,omitempty"` // e.g., "blue", "red" - for future alliance system
	Type      string           `yaml:"type"`           // "player", "ai", "neutral"
	Resources *ResourcesConfig `yaml:"resources,omitempty"`
	Buildings []BuildingConfig `yaml:"buildings,omitempty"`
	Units     []UnitConfig     `yaml:"units,omitempty"`
}

// ResourcesConfig defines starting resources for a faction
type ResourcesConfig struct {
	Metal  float64 `yaml:"metal"`
	Energy float64 `yaml:"energy"`
}

// BuildingConfig defines a building placement
type BuildingConfig struct {
	Type      string  `yaml:"type"` // e.g., "CommandNexus", "TanksFactory"
	X         float64 `yaml:"x"`    // Pixel coordinates
	Y         float64 `yaml:"y"`
	Completed bool    `yaml:"completed,omitempty"` // Default: true for pre-placed buildings
}

// UnitConfig defines a unit placement
type UnitConfig struct {
	Type  string  `yaml:"type"` // e.g., "Tank", "Scout"
	X     float64 `yaml:"x"`    // Pixel coordinates
	Y     float64 `yaml:"y"`
	Count int     `yaml:"count,omitempty"` // Default: 1
	// Tank customization
	Color string `yaml:"color,omitempty"` // color_a, color_b, color_c, color_d
	Hull  int    `yaml:"hull,omitempty"`  // 1-8
	Gun   int    `yaml:"gun,omitempty"`   // 1, 2, 4, 5, 7
}

// LoadMapConfig loads a map configuration from a YAML file
func LoadMapConfig(filename string) (*MapConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read map config: %w", err)
	}

	var config MapConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse map config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid map config: %w", err)
	}

	return &config, nil
}

// Validate checks if the map configuration is valid
func (mc *MapConfig) Validate() error {
	if mc.Name == "" {
		return fmt.Errorf("map name is required")
	}
	if mc.Size.Width <= 0 || mc.Size.Height <= 0 {
		return fmt.Errorf("invalid map size: %dx%d", mc.Size.Width, mc.Size.Height)
	}
	if len(mc.Factions) == 0 {
		return fmt.Errorf("at least one faction is required")
	}

	hasPlayer := false
	for _, f := range mc.Factions {
		if f.ID == "" {
			return fmt.Errorf("faction ID is required")
		}
		if f.Type == "" {
			return fmt.Errorf("faction type is required for faction %s", f.ID)
		}
		if f.Type == "player" {
			hasPlayer = true
		}
	}
	if !hasPlayer {
		return fmt.Errorf("at least one player faction is required")
	}

	return nil
}

// GetTileSize returns the tile size, using default if not specified
func (mc *MapConfig) GetTileSize() float64 {
	if mc.Size.TileSize <= 0 {
		return TileSize // Default from terrain.go
	}
	return mc.Size.TileSize
}

// GetPixelWidth returns the map width in pixels
func (mc *MapConfig) GetPixelWidth() float64 {
	return float64(mc.Size.Width) * mc.GetTileSize()
}

// GetPixelHeight returns the map height in pixels
func (mc *MapConfig) GetPixelHeight() float64 {
	return float64(mc.Size.Height) * mc.GetTileSize()
}

// ToMap creates a terrain Map from the configuration
func (mc *MapConfig) ToMap() *Map {
	tileSize := mc.GetTileSize()

	// Create map with all grass by default
	m := &Map{
		Width:       mc.Size.Width,
		Height:      mc.Size.Height,
		Tiles:       make([][]Tile, mc.Size.Height),
		PixelWidth:  float64(mc.Size.Width) * tileSize,
		PixelHeight: float64(mc.Size.Height) * tileSize,
	}

	// Initialize all tiles as grass
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, mc.Size.Width)
		for x := range m.Tiles[y] {
			m.Tiles[y][x] = Tile{
				Type:      TileGrass,
				Passable:  true,
				Buildable: true,
			}
		}
	}

	// Apply terrain features if specified
	if mc.Terrain != nil {
		mc.applyWaterFeatures(m)
		mc.applyMetalDeposits(m)
	}

	return m
}

// applyWaterFeatures adds water bodies to the map
func (mc *MapConfig) applyWaterFeatures(m *Map) {
	for _, water := range mc.Terrain.Water {
		switch water.Type {
		case "lake":
			mc.applyLake(m, water)
		case "river":
			mc.applyRiver(m, water)
		case "rect", "rectangle":
			mc.applyWaterRect(m, water)
		}
	}
}

// applyLake creates an elliptical lake
func (mc *MapConfig) applyLake(m *Map, water WaterFeature) {
	width := water.Width
	height := water.Height
	if width <= 0 {
		width = water.Size * 2
	}
	if height <= 0 {
		height = water.Size * 2
	}
	if width <= 0 {
		width = 5
	}
	if height <= 0 {
		height = 5
	}

	centerX := water.X + width/2
	centerY := water.Y + height/2
	radiusX := float64(width) / 2.0
	radiusY := float64(height) / 2.0

	for y := water.Y; y < water.Y+height && y < m.Height; y++ {
		for x := water.X; x < water.X+width && x < m.Width; x++ {
			if x < 0 || y < 0 {
				continue
			}
			// Check if point is inside ellipse
			dx := float64(x-centerX) / radiusX
			dy := float64(y-centerY) / radiusY
			if dx*dx+dy*dy <= 1.0 {
				m.Tiles[y][x] = Tile{
					Type:      TileWater,
					Passable:  false,
					Buildable: false,
				}
			}
		}
	}
}

// applyRiver creates a river from start to end point
func (mc *MapConfig) applyRiver(m *Map, water WaterFeature) {
	thickness := water.Size
	if thickness <= 0 {
		thickness = 2
	}

	// Simple line drawing with thickness
	x0, y0 := water.X, water.Y
	x1, y1 := water.EndX, water.EndY

	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		// Apply water in a square around current point
		for ty := y0 - thickness/2; ty <= y0+thickness/2; ty++ {
			for tx := x0 - thickness/2; tx <= x0+thickness/2; tx++ {
				if tx >= 0 && tx < m.Width && ty >= 0 && ty < m.Height {
					m.Tiles[ty][tx] = Tile{
						Type:      TileWater,
						Passable:  false,
						Buildable: false,
					}
				}
			}
		}

		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// applyWaterRect creates a rectangular water area
func (mc *MapConfig) applyWaterRect(m *Map, water WaterFeature) {
	for y := water.Y; y < water.Y+water.Height && y < m.Height; y++ {
		for x := water.X; x < water.X+water.Width && x < m.Width; x++ {
			if x >= 0 && y >= 0 {
				m.Tiles[y][x] = Tile{
					Type:      TileWater,
					Passable:  false,
					Buildable: false,
				}
			}
		}
	}
}

// applyMetalDeposits adds metal deposits to the map
func (mc *MapConfig) applyMetalDeposits(m *Map) {
	for _, metal := range mc.Terrain.Metal {
		if metal.X >= 0 && metal.X < m.Width && metal.Y >= 0 && metal.Y < m.Height {
			amount := metal.Amount
			if amount <= 0 {
				amount = 2000
			}
			m.Tiles[metal.Y][metal.X] = Tile{
				Type:        TileMetal,
				Passable:    true,
				Buildable:   true,
				HasMetal:    true,
				MetalAmount: amount,
			}
		}
	}
}

// GetPlayerFaction returns the first player-controlled faction
func (mc *MapConfig) GetPlayerFaction() *FactionConfig {
	for i := range mc.Factions {
		if mc.Factions[i].Type == "player" {
			return &mc.Factions[i]
		}
	}
	return nil
}

// GetAIFactions returns all AI-controlled factions
func (mc *MapConfig) GetAIFactions() []*FactionConfig {
	var factions []*FactionConfig
	for i := range mc.Factions {
		if mc.Factions[i].Type == "ai" {
			factions = append(factions, &mc.Factions[i])
		}
	}
	return factions
}

// GetFactionByID returns a faction by its ID
func (mc *MapConfig) GetFactionByID(id string) *FactionConfig {
	for i := range mc.Factions {
		if mc.Factions[i].ID == id {
			return &mc.Factions[i]
		}
	}
	return nil
}

// GetFactionsByTeam returns all factions belonging to a team
func (mc *MapConfig) GetFactionsByTeam(team string) []*FactionConfig {
	var factions []*FactionConfig
	for i := range mc.Factions {
		if mc.Factions[i].Team == team {
			factions = append(factions, &mc.Factions[i])
		}
	}
	return factions
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
