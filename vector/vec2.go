package vector

import "math"

const Eps = 1e-9

type Vec2 struct {
	X float64
	Y float64
}

func (v Vec2) Add(w Vec2) Vec2 {
	return Vec2{v.X + w.X, v.Y + w.Y}
}

func (v Vec2) Sub(w Vec2) Vec2 {
	return Vec2{v.X - w.X, v.Y - w.Y}
}

func (v Vec2) Mul(alpha float64) Vec2 {
	return Vec2{alpha * v.X, alpha * v.Y}
}

func (v Vec2) Dot(w Vec2) float64 {
	return v.X*w.X + v.Y*w.Y
}

func (v Vec2) Normalized() Vec2 {
	return v.Mul(1.0 / v.Norm())
}

func (v Vec2) Dist2(w Vec2) float64 {
	return w.Sub(v).Norm2()
}

func (v Vec2) Dist(w Vec2) float64 {
	return math.Hypot(w.X-v.X, w.Y-v.Y)
}

func (v Vec2) Norm2() float64 {
	return v.Dot(v)
}

func (v Vec2) Norm() float64 {
	return math.Sqrt(v.Norm2())
}
