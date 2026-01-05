package render

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

type Renderer struct {
	backgroundColor color.Color
	gridColor       color.Color
	gridSize        int
	showGrid        bool
}

func NewRenderer() *Renderer {
	return &Renderer{
		backgroundColor: color.RGBA{20, 20, 30, 255},
		gridColor:       color.RGBA{40, 40, 50, 255},
		gridSize:        50,
		showGrid:        true,
	}
}
func (r *Renderer) SetBackgroundColor(c color.Color) {
	r.backgroundColor = c
}
func (r *Renderer) SetGridVisible(visible bool) {
	r.showGrid = visible
}
func (r *Renderer) Clear(screen *ebiten.Image) {
	screen.Fill(r.backgroundColor)
}
func (r *Renderer) DrawGrid(screen *ebiten.Image, width, height int) {
	if !r.showGrid {
		return
	}
	for x := 0; x < width; x += r.gridSize {
		vector.StrokeLine(screen, float32(x), 0, float32(x), float32(height), 1, r.gridColor, false)
	}
	for y := 0; y < height; y += r.gridSize {
		vector.StrokeLine(screen, 0, float32(y), float32(width), float32(y), 1, r.gridColor, false)
	}
}
func (r *Renderer) DrawRect(screen *ebiten.Image, rect emath.Rect, c color.Color) {
	vector.FillRect(
		screen,
		float32(rect.Pos.X),
		float32(rect.Pos.Y),
		float32(rect.Size.X),
		float32(rect.Size.Y),
		c,
		false,
	)
}
func (r *Renderer) DrawRectOutline(screen *ebiten.Image, rect emath.Rect, strokeWidth float32, c color.Color) {
	vector.StrokeRect(
		screen,
		float32(rect.Pos.X),
		float32(rect.Pos.Y),
		float32(rect.Size.X),
		float32(rect.Size.Y),
		strokeWidth,
		c,
		false,
	)
}
func (r *Renderer) DrawCircle(screen *ebiten.Image, center emath.Vec2, radius float32, c color.Color) {
	vector.FillCircle(
		screen,
		float32(center.X),
		float32(center.Y),
		radius,
		c,
		false,
	)
}
func (r *Renderer) DrawLine(screen *ebiten.Image, from, to emath.Vec2, strokeWidth float32, c color.Color) {
	vector.StrokeLine(
		screen,
		float32(from.X),
		float32(from.Y),
		float32(to.X),
		float32(to.Y),
		strokeWidth,
		c,
		false,
	)
}
func (r *Renderer) DrawText(screen *ebiten.Image, text string) {
	ebitenutil.DebugPrint(screen, text)
}
func (r *Renderer) DrawTextAt(screen *ebiten.Image, text string, x, y int) {
	ebitenutil.DebugPrintAt(screen, text, x, y)
}
