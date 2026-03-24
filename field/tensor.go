package field

import "math"

type Tensor Vector

func GridTensor(r, theta float64) Tensor {
	return Tensor{r * math.Cos(2.0*theta), r * math.Sin(2.0*theta)}
}

func RadialTensor(v Vector) Tensor {
	return Tensor{v.Y*v.Y - v.X*v.X, -2.0 * v.X * v.Y}
}

func (t Tensor) Add(r Tensor) Tensor {
	return Tensor(Vector(t).Add(Vector(r)))
}

func (t Tensor) Mul(alpha float64) Tensor {
	return Tensor(Vector(t).Mul(alpha))
}

func (t Tensor) MajorEigenvector() Vector {
	norm := Vector(t).Norm()
	if norm < eps {
		return Vector{1.0, 0.0}
	}

	if t.X > 0 {
		return Vector{norm + t.X, t.Y}.Normalized()
	}

	return Vector{t.Y, norm - t.X}.Normalized()
}

func (t Tensor) MinorEigenvector() Vector {
	v := t.MajorEigenvector()
	return Vector{-v.Y, v.X}
}
