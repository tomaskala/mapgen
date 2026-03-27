package streamline

import (
	"slices"

	"tomaskala.com/mapgen/field"
)

type Streamline []field.Vector

type Tracer struct {
	grid       *Grid
	tf         field.TensorField
	rkStep     float64
	dTest      float64
	dLookahead float64
	maxLength  float64
}

func NewTracer(grid *Grid, tf field.TensorField, rkStep, dTest, dLookahead, maxLength float64) Tracer {
	return Tracer{
		grid:       grid,
		tf:         tf,
		rkStep:     rkStep,
		dTest:      dTest,
		dLookahead: dLookahead,
		maxLength:  maxLength,
	}
}

func (t Tracer) Trace(seed field.Vector) {
	// TODO: Jobard and Lefer's algorithm.
}

func (t Tracer) traceStreamline(seed field.Vector, sel field.EigenSelector) Streamline {
	forward := sel(t.tf.Evaluate(seed))
	backward := forward.Mul(-1.0)
	t.grid.Add(seed)

	front := t.traceHalfline(seed, forward, sel)
	back := t.traceHalfline(seed, backward, sel)
	slices.Reverse(back)

	back = append(back, seed)
	return append(back, front...)
}

func (t Tracer) traceHalfline(seed, dir field.Vector, sel field.EigenSelector) Streamline {
	var halfline Streamline
	dist := 0.0
	dTestSq := t.dTest * t.dTest
	curr := seed

	for {
		dir, curr = t.step(dir, curr, sel)
		dist += dir.Norm()

		// Stopping criteria (1): out of domain boundary.
		if !t.grid.IsInBounds(curr) {
			break
		}

		// Stopping criteria (2): degenerate point.
		if t.tf.Evaluate(curr).NormSquared() < field.Eps {
			break
		}

		// Stopping criteria (3): loop.
		if dist > t.dTest && curr.Sub(seed).NormSquared() < dTestSq {
			break
		}

		// Stopping criteria (4): length exceeded.
		if dist > t.maxLength {
			if lookahead, found := t.lookahead(dir, curr, sel); found {
				halfline = append(halfline, lookahead)
			}
			break
		}

		// Stopping criteria (5): too close to an existing streamline.
		if t.grid.IsTooClose(curr, dTestSq) {
			if lookahead, found := t.lookahead(dir, curr, sel); found {
				halfline = append(halfline, lookahead)
			}
			break
		}

		t.grid.Add(curr) // Ignore the boolean, we have already tested out of bounds in (1).
		halfline = append(halfline, curr)
	}

	return halfline
}

func (t Tracer) step(dir, curr field.Vector, sel field.EigenSelector) (field.Vector, field.Vector) {
	next := field.RungeKuttaStep(t.tf, dir, curr, t.rkStep, sel)
	delta := next.Sub(curr)
	return delta, next
}

func (t Tracer) lookahead(dir, curr field.Vector, sel field.EigenSelector) (field.Vector, bool) {
	dist := 0.0
	dTestSq := t.dTest * t.dTest

	for dist < t.dLookahead {
		dir, curr = t.step(dir, curr, sel)
		dist += dir.Norm()

		if !t.grid.IsInBounds(curr) {
			break
		}

		if t.grid.IsTooClose(curr, dTestSq) {
			return curr, true
		}
	}

	return field.Vector{}, false
}
