package math

import "math"

// Vec2 represents a 2D vector
type Vec2 struct {
	X, Y float64
}

// NewVec2 creates a new Vec2
func NewVec2(x, y float64) Vec2 {
	return Vec2{X: x, Y: y}
}

// Add returns the sum of two vectors
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub returns the difference of two vectors
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

// Mul returns the vector scaled by a scalar
func (v Vec2) Mul(s float64) Vec2 {
	return Vec2{X: v.X * s, Y: v.Y * s}
}

// Div returns the vector divided by a scalar
func (v Vec2) Div(s float64) Vec2 {
	if s == 0 {
		return Vec2{}
	}
	return Vec2{X: v.X / s, Y: v.Y / s}
}

// Dot returns the dot product of two vectors
func (v Vec2) Dot(other Vec2) float64 {
	return v.X*other.X + v.Y*other.Y
}

// Length returns the magnitude of the vector
func (v Vec2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// LengthSquared returns the squared magnitude (faster, no sqrt)
func (v Vec2) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y
}

// Normalize returns a unit vector in the same direction
func (v Vec2) Normalize() Vec2 {
	length := v.Length()
	if length == 0 {
		return Vec2{}
	}
	return v.Div(length)
}

// Distance returns the distance between two points
func (v Vec2) Distance(other Vec2) float64 {
	return v.Sub(other).Length()
}

// DistanceSquared returns the squared distance (faster, no sqrt)
func (v Vec2) DistanceSquared(other Vec2) float64 {
	return v.Sub(other).LengthSquared()
}

// Rect represents an axis-aligned bounding box
type Rect struct {
	Pos  Vec2 // Top-left corner
	Size Vec2 // Width and height
}

// NewRect creates a new Rect
func NewRect(x, y, w, h float64) Rect {
	return Rect{Pos: Vec2{x, y}, Size: Vec2{w, h}}
}

// Center returns the center point of the rectangle
func (r Rect) Center() Vec2 {
	return Vec2{
		X: r.Pos.X + r.Size.X/2,
		Y: r.Pos.Y + r.Size.Y/2,
	}
}

// Contains checks if a point is inside the rectangle
func (r Rect) Contains(p Vec2) bool {
	return p.X >= r.Pos.X && p.X <= r.Pos.X+r.Size.X &&
		p.Y >= r.Pos.Y && p.Y <= r.Pos.Y+r.Size.Y
}

// Intersects checks if two rectangles overlap
func (r Rect) Intersects(other Rect) bool {
	return r.Pos.X < other.Pos.X+other.Size.X &&
		r.Pos.X+r.Size.X > other.Pos.X &&
		r.Pos.Y < other.Pos.Y+other.Size.Y &&
		r.Pos.Y+r.Size.Y > other.Pos.Y
}
