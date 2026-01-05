package ui

import (
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

var (
	BackgroundColor    = color.RGBA{30, 30, 40, 240}
	BorderColor        = color.RGBA{60, 60, 80, 255}
	TextColor          = color.RGBA{200, 200, 200, 255}
	BarBackgroundColor = color.RGBA{20, 20, 25, 255}
)

type Panel struct {
	Bounds      emath.Rect
	Background  color.Color
	Border      color.Color
	BorderWidth float32
}

func NewPanel(x, y, w, h float64) *Panel {
	return &Panel{
		Bounds:      emath.NewRect(x, y, w, h),
		Background:  BackgroundColor,
		Border:      BorderColor,
		BorderWidth: 1,
	}
}
func (p *Panel) Draw(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(p.Bounds.Pos.X),
		float32(p.Bounds.Pos.Y),
		float32(p.Bounds.Size.X),
		float32(p.Bounds.Size.Y),
		p.Background,
		false,
	)
	if p.BorderWidth > 0 {
		vector.StrokeRect(
			screen,
			float32(p.Bounds.Pos.X),
			float32(p.Bounds.Pos.Y),
			float32(p.Bounds.Size.X),
			float32(p.Bounds.Size.Y),
			p.BorderWidth,
			p.Border,
			false,
		)
	}
}

type ProgressBar struct {
	Bounds     emath.Rect
	Progress   float64 // 0.0 to 1.0
	FillColor  color.Color
	Background color.Color
}

func NewProgressBar(x, y, w, h float64, fillColor color.Color) *ProgressBar {
	return &ProgressBar{
		Bounds:     emath.NewRect(x, y, w, h),
		Progress:   0,
		FillColor:  fillColor,
		Background: BarBackgroundColor,
	}
}
func (pb *ProgressBar) Draw(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(pb.Bounds.Pos.X),
		float32(pb.Bounds.Pos.Y),
		float32(pb.Bounds.Size.X),
		float32(pb.Bounds.Size.Y),
		pb.Background,
		false,
	)
	if pb.Progress > 0 {
		fillWidth := pb.Bounds.Size.X * pb.Progress
		vector.FillRect(
			screen,
			float32(pb.Bounds.Pos.X),
			float32(pb.Bounds.Pos.Y),
			float32(fillWidth),
			float32(pb.Bounds.Size.Y),
			pb.FillColor,
			false,
		)
	}
}
func (pb *ProgressBar) SetProgress(p float64) {
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	pb.Progress = p
}
