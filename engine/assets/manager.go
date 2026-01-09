package assets

import (
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Manager struct {
	sprites  map[string]*ebiten.Image
	basePath string
}

func NewManager(basePath string) *Manager {
	return &Manager{
		sprites:  make(map[string]*ebiten.Image),
		basePath: basePath,
	}
}

func (m *Manager) GetSprite(relativePath string) *ebiten.Image {
	if relativePath == "" {
		return nil
	}

	if sprite, ok := m.sprites[relativePath]; ok {
		return sprite
	}

	fullPath := filepath.Join(m.basePath, relativePath)
	sprite, _, err := ebitenutil.NewImageFromFile(fullPath)
	if err != nil {
		log.Printf("Warning: could not load sprite %s: %v", fullPath, err)
		m.sprites[relativePath] = nil
		return nil
	}

	m.sprites[relativePath] = sprite
	return sprite
}
