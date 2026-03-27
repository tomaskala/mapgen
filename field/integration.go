package field

type EigenSelector func(Tensor) Vector

func RungeKuttaStep(tf TensorField, prev, curr Vector, step float64, sel EigenSelector) Vector {
	k1 := rungeKuttaEvaluate(tf, prev, curr, sel)
	k2 := rungeKuttaEvaluate(tf, k1, curr.Add(k1.Mul(step/2.0)), sel)
	k3 := rungeKuttaEvaluate(tf, k2, curr.Add(k2.Mul(step/2.0)), sel)
	k4 := rungeKuttaEvaluate(tf, k3, curr.Add(k3.Mul(step)), sel)

	return curr.Add(k1.Add(k2.Mul(2.0)).Add(k3.Mul(2.0)).Add(k4).Mul(step / 6.0))
}

func rungeKuttaEvaluate(tf TensorField, prev, curr Vector, sel EigenSelector) Vector {
	dir := sel(tf.Evaluate(curr))

	if dir.Dot(prev) < 0.0 {
		return dir.Mul(-1.0)
	}
	return dir
}
