package field

import (
	"math"

	"tomaskala.com/mapgen/vector"
)

type TensorField []BasisField

func (tf TensorField) Evaluate(p vector.Vec2) Tensor {
	t := Tensor{}

	for _, f := range tf {
		coef := math.Exp(-f.decay * p.Dist2(f.center))
		t = t.add(f.evaluate(p).mul(coef))
	}

	return t
}

type BasisType int

const (
	BasisGrid BasisType = iota
	BasisRadial
)

type BasisField struct {
	typ        BasisType
	center     vector.Vec2
	decay      float64
	baseTensor Tensor // Only used by the grid field.
}

func Grid(center, direction vector.Vec2, radius float64) BasisField {
	l := direction.Norm()
	theta := math.Atan2(direction.Y, direction.X)
	baseTensor := gridTensor(l, theta)
	return BasisField{
		typ:        BasisGrid,
		center:     center,
		decay:      radiusToDecay(radius),
		baseTensor: baseTensor,
	}
}

func Radial(center vector.Vec2, radius float64) BasisField {
	return BasisField{
		typ:    BasisRadial,
		center: center,
		decay:  radiusToDecay(radius),
	}
}

func (bf BasisField) evaluate(p vector.Vec2) Tensor {
	switch bf.typ {
	case BasisGrid:
		// A grid is spatially constant.
		return bf.baseTensor
	case BasisRadial:
		u := p.Sub(bf.center)
		return radialTensor(u)
	default:
		panic("unrecognized basis field type")
	}
}

func radiusToDecay(radius float64) float64 {
	return 1.0 / (radius * radius)
}
