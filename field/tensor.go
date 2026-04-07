package field

import "math"

// Tensor represents a traceless symmetrix 2x2 matrix of the form [a, b; b, -a].
type Tensor struct {
	a float64
	b float64
}

func gridTensor(r, theta float64) Tensor {
	return Tensor{r * math.Cos(2.0*theta), r * math.Sin(2.0*theta)}
}

func radialTensor(v Vector) Tensor {
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

func (t Tensor) MajorEigenvector() Vector {
	norm := math.Sqrt(t.Norm2())
	if norm < Eps {
		return Vector{1.0, 0.0}
	}

	if t.a > 0 {
		return Vector{norm + t.a, t.b}.normalized()
	}

	return Vector{t.b, norm - t.a}.normalized()
}

func (t Tensor) MinorEigenvector() Vector {
	v := t.MajorEigenvector()
	return Vector{-v.Y, v.X}
}
