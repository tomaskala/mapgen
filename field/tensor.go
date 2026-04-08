package field

import (
	"math"

	"tomaskala.com/mapgen/vector"
)

// Tensor represents a traceless symmetrix 2x2 matrix of the form [a, b; b, -a].
type Tensor struct {
	a float64
	b float64
}

func gridTensor(r, theta float64) Tensor {
	return Tensor{r * math.Cos(2.0*theta), r * math.Sin(2.0*theta)}
}

func radialTensor(v vector.Vec2) Tensor {
	return Tensor{v.Y*v.Y - v.X*v.X, -2.0 * v.X * v.Y}
}

func (t Tensor) add(r Tensor) Tensor {
	return Tensor{t.a + r.a, t.b + r.b}
}

func (t Tensor) mul(alpha float64) Tensor {
	return Tensor{alpha * t.a, alpha * t.b}
}

func (t Tensor) Norm2() float64 {
	return t.a*t.a + t.b*t.b
}

func (t Tensor) MajorEigenvector() vector.Vec2 {
	norm := math.Sqrt(t.Norm2())
	if norm < vector.Eps {
		return vector.Vec2{X: 1.0, Y: 0.0}
	}

	if t.a > 0 {
		return vector.Vec2{X: norm + t.a, Y: t.b}.Normalized()
	}

	return vector.Vec2{X: t.b, Y: norm - t.a}.Normalized()
}

func (t Tensor) MinorEigenvector() vector.Vec2 {
	v := t.MajorEigenvector()
	return vector.Vec2{X: -v.Y, Y: v.X}
}
