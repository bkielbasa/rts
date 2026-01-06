package fog

import (
	emath "github.com/bklimczak/tanks/engine/math"
)

type TileState int

const (
	Unexplored TileState = iota
	Explored
	Visible
)

type FogOfWar struct {
	Width    int
	Height   int
	Tiles    [][]TileState
	TileSize float64
	Version  int
}

func New(worldWidth, worldHeight, tileSize float64) *FogOfWar {
	width := int(worldWidth / tileSize)
	height := int(worldHeight / tileSize)
	tiles := make([][]TileState, height)
	for y := range tiles {
		tiles[y] = make([]TileState, width)
	}
	return &FogOfWar{
		Width:    width,
		Height:   height,
		Tiles:    tiles,
		TileSize: tileSize,
	}
}

func (f *FogOfWar) ClearVisibility() {
	changed := false
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			if f.Tiles[y][x] == Visible {
				f.Tiles[y][x] = Explored
				changed = true
			}
		}
	}
	if changed {
		f.Version++
	}
}

func (f *FogOfWar) RevealCircle(worldX, worldY, radius float64) {
	tileX := int(worldX / f.TileSize)
	tileY := int(worldY / f.TileSize)
	tileRadius := int(radius/f.TileSize) + 1

	for dy := -tileRadius; dy <= tileRadius; dy++ {
		for dx := -tileRadius; dx <= tileRadius; dx++ {
			checkX := tileX + dx
			checkY := tileY + dy

			if checkX < 0 || checkX >= f.Width || checkY < 0 || checkY >= f.Height {
				continue
			}

			tileCenterX := (float64(checkX) + 0.5) * f.TileSize
			tileCenterY := (float64(checkY) + 0.5) * f.TileSize
			distSq := (tileCenterX-worldX)*(tileCenterX-worldX) + (tileCenterY-worldY)*(tileCenterY-worldY)

			if distSq <= radius*radius {
				if f.Tiles[checkY][checkX] != Visible {
					f.Tiles[checkY][checkX] = Visible
					f.Version++
				}
			}
		}
	}
}

func (f *FogOfWar) GetTileState(worldX, worldY float64) TileState {
	tileX := int(worldX / f.TileSize)
	tileY := int(worldY / f.TileSize)

	if tileX < 0 || tileX >= f.Width || tileY < 0 || tileY >= f.Height {
		return Unexplored
	}

	return f.Tiles[tileY][tileX]
}

func (f *FogOfWar) GetTileStateAt(tileX, tileY int) TileState {
	if tileX < 0 || tileX >= f.Width || tileY < 0 || tileY >= f.Height {
		return Unexplored
	}
	return f.Tiles[tileY][tileX]
}

func (f *FogOfWar) IsVisible(bounds emath.Rect) bool {
	centerX := bounds.Pos.X + bounds.Size.X/2
	centerY := bounds.Pos.Y + bounds.Size.Y/2
	return f.GetTileState(centerX, centerY) == Visible
}

func (f *FogOfWar) IsExplored(bounds emath.Rect) bool {
	centerX := bounds.Pos.X + bounds.Size.X/2
	centerY := bounds.Pos.Y + bounds.Size.Y/2
	state := f.GetTileState(centerX, centerY)
	return state == Explored || state == Visible
}
