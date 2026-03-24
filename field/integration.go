package field

type RungeKutta struct {
	field TensorField
}

type EigenSelector func(Tensor) Vector

func (rk RungeKutta) Step(prev, curr Vector, step float64, sel EigenSelector) Vector {
	k1 := rk.evaluate(prev, curr, sel)
	k2 := rk.evaluate(k1, curr.Add(k1.Mul(step/2.0)), sel)
	k3 := rk.evaluate(k2, curr.Add(k2.Mul(step/2.0)), sel)
	k4 := rk.evaluate(k3, curr.Add(k3.Mul(step)), sel)

	return curr.Add(k1.Add(k2.Mul(2.0)).Add(k3.Mul(2.0)).Add(k4).Mul(step / 6.0))
}

func (rk RungeKutta) evaluate(prev, curr Vector, sel EigenSelector) Vector {
	t := rk.field.Evaluate(curr)
	dir := sel(t)

	if dir.Dot(prev) < 0.0 {
		return dir.Mul(-1.0)
	}
	return dir
}
