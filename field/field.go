package field

import "math"

type TensorField []BasisField

func (tf TensorField) Evaluate(p Vector) Tensor {
	t := Tensor{}

	for _, f := range tf {
		dp := p.Sub(f.center)
		coef := math.Exp(-f.decay * dp.Norm2())
		t = t.add(f.evaluate(p).mul(coef))
	}

	return t
}
