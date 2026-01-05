package ui

import (
	"image/color"

	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Colors for UI elements
var (
	BackgroundColor    = color.RGBA{30, 30, 40, 240}
	BorderColor        = color.RGBA{60, 60, 80, 255}
	TextColor          = color.RGBA{200, 200, 200, 255}
	BarBackgroundColor = color.RGBA{20, 20, 25, 255}
)

// Panel represents a UI panel/container
type Panel struct {
	Bounds     emath.Rect
	Background color.Color
	Border     color.Color
	BorderWidth float32
}

// NewPanel creates a new panel
func NewPanel(x, y, w, h float64) *Panel {
	return &Panel{
		Bounds:      emath.NewRect(x, y, w, h),
		Background:  BackgroundColor,
		Border:      BorderColor,
		BorderWidth: 1,
	}
}

// Draw renders the panel
func (p *Panel) Draw(screen *ebiten.Image) {
	// Background
	vector.FillRect(
		screen,
		float32(p.Bounds.Pos.X),
		float32(p.Bounds.Pos.Y),
		float32(p.Bounds.Size.X),
		float32(p.Bounds.Size.Y),
		p.Background,
		false,
	)

	// Border
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

// ProgressBar represents a horizontal progress bar
type ProgressBar struct {
	Bounds     emath.Rect
	Progress   float64 // 0.0 to 1.0
	FillColor  color.Color
	Background color.Color
}

// NewProgressBar creates a new progress bar
func NewProgressBar(x, y, w, h float64, fillColor color.Color) *ProgressBar {
	return &ProgressBar{
		Bounds:     emath.NewRect(x, y, w, h),
		Progress:   0,
		FillColor:  fillColor,
		Background: BarBackgroundColor,
	}
}

// Draw renders the progress bar
func (pb *ProgressBar) Draw(screen *ebiten.Image) {
	// Background
	vector.FillRect(
		screen,
		float32(pb.Bounds.Pos.X),
		float32(pb.Bounds.Pos.Y),
		float32(pb.Bounds.Size.X),
		float32(pb.Bounds.Size.Y),
		pb.Background,
		false,
	)

	// Fill
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

// SetProgress sets the progress (clamped 0-1)
func (pb *ProgressBar) SetProgress(p float64) {
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	pb.Progress = p
}
