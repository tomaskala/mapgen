package field

import "math"

type BasisType int

const (
	BasisGrid BasisType = iota
	BasisRadial
)

type BasisField struct {
	typ        BasisType
	center     Vector
	decay      float64
	baseTensor Tensor // Only used by the grid field.
}

func Grid(center, direction Vector, radius float64) BasisField {
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

func Radial(center Vector, radius float64) BasisField {
	return BasisField{
		typ:    BasisRadial,
		center: center,
		decay:  radiusToDecay(radius),
	}
}

func (bf BasisField) evaluate(p Vector) Tensor {
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
