package engine

import (
	"github.com/bklimczak/tanks/engine/camera"
	"github.com/bklimczak/tanks/engine/collision"
	"github.com/bklimczak/tanks/engine/input"
	"github.com/bklimczak/tanks/engine/render"
	"github.com/bklimczak/tanks/engine/resource"
)

type Engine struct {
	Input     *input.Manager
	Collision *collision.System
	Renderer  *render.Renderer
	Resources *resource.Manager
	Camera    *camera.Camera
}

func New(worldWidth, worldHeight, viewportWidth, viewportHeight float64) *Engine {
	return &Engine{
		Input:     input.NewManager(),
		Collision: collision.NewSystem(worldWidth, worldHeight),
		Renderer:  render.NewRenderer(),
		Resources: resource.NewManager(),
		Camera:    camera.New(worldWidth, worldHeight, viewportWidth, viewportHeight),
	}
}
func (e *Engine) UpdateWorldSize(width, height float64) {
	e.Collision.SetWorldBounds(width, height)
	e.Camera.SetWorldSize(width, height)
}
func (e *Engine) UpdateViewportSize(width, height float64) {
	e.Camera.SetViewportSize(width, height)
}
