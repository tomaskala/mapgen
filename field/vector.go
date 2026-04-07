package field

import "math"

const Eps = 1e-9

type Vector struct {
	X float64
	Y float64
}

func (v Vector) Add(w Vector) Vector {
	return Vector{v.X + w.X, v.Y + w.Y}
}

func (v Vector) Sub(w Vector) Vector {
	return Vector{v.X - w.X, v.Y - w.Y}
}

func (v Vector) Mul(alpha float64) Vector {
	return Vector{alpha * v.X, alpha * v.Y}
}

func (v Vector) Dot(w Vector) float64 {
	return v.X*w.X + v.Y*w.Y
}

func (v Vector) normalized() Vector {
	norm := v.Norm()
	if norm < Eps {
		return Vector{}
	}
	return v.Mul(1.0 / norm)
}

func (v Vector) Norm2() float64 {
	return v.Dot(v)
}

func (v Vector) Norm() float64 {
	return math.Sqrt(v.Norm2())
}
