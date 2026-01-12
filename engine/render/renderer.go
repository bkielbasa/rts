package render

import (
	"image"
	"image/color"
	"math"

	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var emptyImage = ebiten.NewImage(3, 3)
var emptySubImage = emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

func init() {
	emptyImage.Fill(color.White)
}

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

func (r *Renderer) DrawRotatedRect(screen *ebiten.Image, center emath.Vec2, width, height, angle float64, c color.Color) {
	halfW := width / 2
	halfH := height / 2
	corners := r.getRotatedCorners(center, halfW, halfH, angle)

	var path vector.Path
	path.MoveTo(float32(corners[0].X), float32(corners[0].Y))
	path.LineTo(float32(corners[1].X), float32(corners[1].Y))
	path.LineTo(float32(corners[2].X), float32(corners[2].Y))
	path.LineTo(float32(corners[3].X), float32(corners[3].Y))
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	cr, cg, cb, ca := c.RGBA()
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(cr) / 0xffff
		vs[i].ColorG = float32(cg) / 0xffff
		vs[i].ColorB = float32(cb) / 0xffff
		vs[i].ColorA = float32(ca) / 0xffff
	}
	screen.DrawTriangles(vs, is, emptySubImage, nil)
}

func (r *Renderer) getRotatedCorners(center emath.Vec2, halfW, halfH, angle float64) [4]emath.Vec2 {
	sin, cos := math.Sin(angle), math.Cos(angle)
	corners := [4]emath.Vec2{
		{X: -halfW, Y: -halfH},
		{X: halfW, Y: -halfH},
		{X: halfW, Y: halfH},
		{X: -halfW, Y: halfH},
	}
	for i, c := range corners {
		rotX := c.X*cos - c.Y*sin
		rotY := c.X*sin + c.Y*cos
		corners[i] = emath.Vec2{X: center.X + rotX, Y: center.Y + rotY}
	}
	return corners
}
func (r *Renderer) DrawText(screen *ebiten.Image, text string) {
	ebitenutil.DebugPrint(screen, text)
}
func (r *Renderer) DrawTextAt(screen *ebiten.Image, text string, x, y int) {
	ebitenutil.DebugPrintAt(screen, text, x, y)
}
