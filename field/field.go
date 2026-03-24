package field

import "math"

type TensorField []BasisField

func (tf TensorField) Evaluate(p Vector) Tensor {
	t := Tensor{}

	for _, f := range tf {
		dp := p.Sub(f.center)
		coef := math.Exp(-f.decay * dp.NormSquared())
		t = t.Add(f.Evaluate(p).Mul(coef))
	}

	return t
}
