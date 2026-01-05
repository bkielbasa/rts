package math

import "math"

type Vec2 struct {
	X, Y float64
}

func NewVec2(x, y float64) Vec2 {
	return Vec2{X: x, Y: y}
}
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}
func (v Vec2) Mul(s float64) Vec2 {
	return Vec2{X: v.X * s, Y: v.Y * s}
}
func (v Vec2) Div(s float64) Vec2 {
	if s == 0 {
		return Vec2{}
	}
	return Vec2{X: v.X / s, Y: v.Y / s}
}
func (v Vec2) Dot(other Vec2) float64 {
	return v.X*other.X + v.Y*other.Y
}
func (v Vec2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}
func (v Vec2) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y
}
func (v Vec2) Normalize() Vec2 {
	length := v.Length()
	if length == 0 {
		return Vec2{}
	}
	return v.Div(length)
}
func (v Vec2) Distance(other Vec2) float64 {
	return v.Sub(other).Length()
}
func (v Vec2) DistanceSquared(other Vec2) float64 {
	return v.Sub(other).LengthSquared()
}

type Rect struct {
	Pos  Vec2 // Top-left corner
	Size Vec2 // Width and height
}

func NewRect(x, y, w, h float64) Rect {
	return Rect{Pos: Vec2{x, y}, Size: Vec2{w, h}}
}
func (r Rect) Center() Vec2 {
	return Vec2{
		X: r.Pos.X + r.Size.X/2,
		Y: r.Pos.Y + r.Size.Y/2,
	}
}
func (r Rect) Contains(p Vec2) bool {
	return p.X >= r.Pos.X && p.X <= r.Pos.X+r.Size.X &&
		p.Y >= r.Pos.Y && p.Y <= r.Pos.Y+r.Size.Y
}
func (r Rect) Intersects(other Rect) bool {
	return r.Pos.X < other.Pos.X+other.Size.X &&
		r.Pos.X+r.Size.X > other.Pos.X &&
		r.Pos.Y < other.Pos.Y+other.Size.Y &&
		r.Pos.Y+r.Size.Y > other.Pos.Y
}
