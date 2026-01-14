package terrain

import (
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	TilesetTileSize = 16 // Size of tiles in the tileset image
)

type Sprites struct {
	grassTile *ebiten.Image
	loaded    bool
}

var tileSprites *Sprites

func LoadSprites() {
	tileSprites = &Sprites{}

	tileset, _, err := ebitenutil.NewImageFromFile("assets/terrain/grass.png")
	if err != nil {
		log.Printf("Warning: could not load grass tileset: %v", err)
		return
	}

	// Extract the first 16x16 tile (top-left corner)
	tileSprites.grassTile = tileset.SubImage(image.Rect(0, 0, TilesetTileSize, TilesetTileSize)).(*ebiten.Image)
	tileSprites.loaded = true
}

func GetGrassTile() *ebiten.Image {
	if tileSprites == nil || !tileSprites.loaded {
		return nil
	}
	return tileSprites.grassTile
}

func SpritesLoaded() bool {
	return tileSprites != nil && tileSprites.loaded
}
