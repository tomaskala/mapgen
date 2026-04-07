package streamline

import (
	"container/heap"

	"tomaskala.com/mapgen/field"
)

type Streamline struct {
	seed  field.Vector
	back  []field.Vector
	front []field.Vector
}

func (s Streamline) Points() []field.Vector {
	points := make([]field.Vector, len(s.back)+1+len(s.front))

	for i, p := range s.back {
		points[len(s.back)-1-i] = p
	}

	points[len(s.back)] = s.seed
	copy(points[len(s.back)+1:], s.front)

	return points
}

type Tracer struct {
	tf         field.TensorField
	dSep       float64
	dTest      float64
	dLookahead float64
	rkStep     float64
	maxLength  float64
}

func NewTracer(tf field.TensorField, dSep, dTest, dLookahead, rkStep, maxLength float64) Tracer {
	return Tracer{
		tf:         tf,
		dSep:       dSep,
		dTest:      dTest,
		dLookahead: dLookahead,
		rkStep:     rkStep,
		maxLength:  maxLength,
	}
}

type Trace struct {
	Major []Streamline
	Minor []Streamline
}

func (t Tracer) Run(majorGrid, minorGrid *Grid, seeds []field.Vector) Trace {
	priority := func(field.Vector) float64 {
		// TODO: Calculate priority based on the paper.
		return 0.0
	}

	var majorLines, minorLines []Streamline
	dSep2 := t.dSep * t.dSep

	pq := make(PriorityQueue, len(seeds))
	for i, seed := range seeds {
		pq[i] = Item{
			p:        seed,
			self:     Family{majorGrid, &majorLines, field.Tensor.MajorEigenvector},
			other:    Family{minorGrid, &minorLines, field.Tensor.MinorEigenvector},
			priority: priority(seed),
		}
	}
	heap.Init(&pq)

	for pq.Len() > 0 {
		curr := heap.Pop(&pq).(Item)
		if !curr.self.grid.IsInBounds(curr.p) || curr.self.grid.IsTooClose(curr.p, dSep2) {
			continue
		}

		line := t.traceStreamline(curr)
		*curr.self.lines = append(*curr.self.lines, line)

		nextSeeds := findSeeds(line, dSep2)
		for _, next := range nextSeeds {
			heap.Push(&pq, Item{
				p:        next,
				self:     curr.other,
				other:    curr.self,
				priority: priority(next),
			})
		}
	}

	return Trace{Major: majorLines, Minor: minorLines}
}

func findSeeds(line Streamline, dSep2 float64) []field.Vector {
	var seeds []field.Vector
	prev := line.seed

	for _, p := range line.front {
		if p.Dist2(prev) >= dSep2 {
			seeds = append(seeds, p)
			prev = p
		}
	}

	prev = line.seed
	for _, p := range line.back {
		if p.Dist2(prev) >= dSep2 {
			seeds = append(seeds, p)
			prev = p
		}
	}

	return seeds
}

func (t Tracer) traceStreamline(item Item) Streamline {
	forward := item.self.sel(t.tf.Evaluate(item.p))
	backward := forward.Mul(-1.0)

	back := t.traceHalfline(item, backward)
	front := t.traceHalfline(item, forward)

	item.self.grid.Add(item.p)
	item.self.grid.AddAll(back)
	item.self.grid.AddAll(front)

	return Streamline{
		seed:  item.p,
		back:  back,
		front: front,
	}
}

func (t Tracer) traceHalfline(item Item, dir field.Vector) []field.Vector {
	var halfline []field.Vector
	dist := 0.0
	dTest2 := t.dTest * t.dTest
	curr := item.p

	for {
		dir, curr = t.step(dir, curr, item.self.sel)
		dist += dir.Norm()

		// Stopping criteria (1): out of domain boundary.
		if !item.self.grid.IsInBounds(curr) {
			break
		}

		// Stopping criteria (2): degenerate point.
		if t.tf.Evaluate(curr).Norm2() < field.Eps {
			break
		}

		// Stopping criteria (3): loop.
		if dist > t.dTest && curr.Dist2(item.p) < dTest2 {
			break
		}

		// Stopping criteria (4): length exceeded.
		if dist > t.maxLength {
			if lookahead, found := t.lookahead(item.other.grid, dir, curr, item.self.sel); found {
				halfline = append(halfline, lookahead)
			}
			break
		}

		// Stopping criteria (5): too close to an existing streamline.
		if item.self.grid.IsTooClose(curr, dTest2) {
			if lookahead, found := t.lookahead(item.other.grid, dir, curr, item.self.sel); found {
				halfline = append(halfline, lookahead)
			}
			break
		}

		halfline = append(halfline, curr)
	}

	return halfline
}

func (t Tracer) step(dir, curr field.Vector, sel field.EigenSelector) (field.Vector, field.Vector) {
	next := field.RungeKuttaStep(t.tf, dir, curr, t.rkStep, sel)
	delta := next.Sub(curr)
	return delta, next
}

func (t Tracer) lookahead(grid *Grid, dir, curr field.Vector, sel field.EigenSelector) (field.Vector, bool) {
	dist := 0.0
	dTest2 := t.dTest * t.dTest

	for dist < t.dLookahead {
		dir, curr = t.step(dir, curr, sel)
		dist += dir.Norm()

		if !grid.IsInBounds(curr) {
			break
		}

		if grid.IsTooClose(curr, dTest2) {
			return curr, true
		}
	}

	return field.Vector{}, false
}
