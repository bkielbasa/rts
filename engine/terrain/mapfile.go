package terrain

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// MapFile represents the YAML map file format
type MapFile struct {
	// Metadata
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Author      string `yaml:"author,omitempty"`
	Version     string `yaml:"version,omitempty"`

	// Map dimensions in tiles
	Width  int `yaml:"width"`
	Height int `yaml:"height"`

	// Tile size in pixels
	TileSize float64 `yaml:"tile_size"`

	// Terrain data - each character represents a tile type
	// G = Grass, W = Water, M = Metal deposit
	Terrain []string `yaml:"terrain"`

	// Metal amounts for metal tiles (optional, defaults to 2000)
	// Format: "x,y:amount" e.g. "10,5:3000"
	MetalAmounts []string `yaml:"metal_amounts,omitempty"`

	// Spawn points for different factions
	SpawnPoints []SpawnPoint `yaml:"spawn_points,omitempty"`
}

// SpawnPoint defines a starting location for units
type SpawnPoint struct {
	Faction string  `yaml:"faction"` // "player", "enemy", "neutral"
	X       float64 `yaml:"x"`       // Pixel coordinates
	Y       float64 `yaml:"y"`
}

// TileChar maps tile types to characters for the map file
var TileChar = map[TileType]rune{
	TileGrass: 'G',
	TileWater: 'W',
	TileMetal: 'M',
}

// CharToTile maps characters to tile types
var CharToTile = map[rune]TileType{
	'G': TileGrass,
	'.': TileGrass, // Alternative for grass
	'W': TileWater,
	'~': TileWater, // Alternative for water
	'M': TileMetal,
	'#': TileMetal, // Alternative for metal
}

// LoadMapFromFile loads a terrain map from a YAML file
func LoadMapFromFile(filename string) (*Map, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read map file: %w", err)
	}

	var mapFile MapFile
	if err := yaml.Unmarshal(data, &mapFile); err != nil {
		return nil, fmt.Errorf("failed to parse map file: %w", err)
	}

	return mapFile.ToMap()
}

// ToMap converts a MapFile to a terrain Map
func (mf *MapFile) ToMap() (*Map, error) {
	// Validate dimensions
	if mf.Width <= 0 || mf.Height <= 0 {
		return nil, fmt.Errorf("invalid map dimensions: %dx%d", mf.Width, mf.Height)
	}

	if len(mf.Terrain) != mf.Height {
		return nil, fmt.Errorf("terrain rows (%d) don't match height (%d)", len(mf.Terrain), mf.Height)
	}

	tileSize := mf.TileSize
	if tileSize <= 0 {
		tileSize = TileSize // Use default
	}

	// Create the map
	m := &Map{
		Width:       mf.Width,
		Height:      mf.Height,
		Tiles:       make([][]Tile, mf.Height),
		PixelWidth:  float64(mf.Width) * tileSize,
		PixelHeight: float64(mf.Height) * tileSize,
	}

	// Parse terrain data
	for y := 0; y < mf.Height; y++ {
		row := mf.Terrain[y]
		if len(row) != mf.Width {
			return nil, fmt.Errorf("row %d has %d tiles, expected %d", y, len(row), mf.Width)
		}

		m.Tiles[y] = make([]Tile, mf.Width)
		for x, char := range row {
			tileType, ok := CharToTile[char]
			if !ok {
				return nil, fmt.Errorf("unknown tile character '%c' at (%d, %d)", char, x, y)
			}

			tile := Tile{
				Type:      tileType,
				Passable:  tileType != TileWater,
				Buildable: tileType != TileWater,
			}

			if tileType == TileMetal {
				tile.HasMetal = true
				tile.MetalAmount = 2000 // Default
			}

			m.Tiles[y][x] = tile
		}
	}

	// Apply custom metal amounts
	for _, ma := range mf.MetalAmounts {
		var x, y int
		var amount float64
		if _, err := fmt.Sscanf(ma, "%d,%d:%f", &x, &y, &amount); err == nil {
			if x >= 0 && x < mf.Width && y >= 0 && y < mf.Height {
				if m.Tiles[y][x].HasMetal {
					m.Tiles[y][x].MetalAmount = amount
				}
			}
		}
	}

	return m, nil
}

// SaveMapToFile saves a terrain map to a YAML file
func SaveMapToFile(m *Map, filename, name, description, author string) error {
	mapFile := MapToFile(m, name, description, author)

	data, err := yaml.Marshal(mapFile)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	// Add a header comment
	header := `# Tanks RTS Map File
# Terrain characters:
#   G or . = Grass (passable, buildable)
#   W or ~ = Water (impassable)
#   M or # = Metal deposit (passable, buildable, extractable)
#
`
	fullData := header + string(data)

	if err := os.WriteFile(filename, []byte(fullData), 0644); err != nil {
		return fmt.Errorf("failed to write map file: %w", err)
	}

	return nil
}

// MapToFile converts a terrain Map to a MapFile for saving
func MapToFile(m *Map, name, description, author string) *MapFile {
	mf := &MapFile{
		Name:        name,
		Description: description,
		Author:      author,
		Version:     "1.0",
		Width:       m.Width,
		Height:      m.Height,
		TileSize:    TileSize,
		Terrain:     make([]string, m.Height),
	}

	// Build terrain strings and collect metal amounts
	var metalAmounts []string

	for y := 0; y < m.Height; y++ {
		var row strings.Builder
		for x := 0; x < m.Width; x++ {
			tile := m.Tiles[y][x]
			char, ok := TileChar[tile.Type]
			if !ok {
				char = 'G' // Default to grass
			}
			row.WriteRune(char)

			// Track non-default metal amounts
			if tile.HasMetal && tile.MetalAmount != 2000 {
				metalAmounts = append(metalAmounts, fmt.Sprintf("%d,%d:%.0f", x, y, tile.MetalAmount))
			}
		}
		mf.Terrain[y] = row.String()
	}

	if len(metalAmounts) > 0 {
		mf.MetalAmounts = metalAmounts
	}

	return mf
}

// GenerateAndSave generates a new map with the given seed and saves it
func GenerateAndSave(pixelWidth, pixelHeight float64, seed int64, filename, name, description, author string) error {
	m := NewMap(pixelWidth, pixelHeight)
	m.Generate(seed)
	return SaveMapToFile(m, filename, name, description, author)
}
