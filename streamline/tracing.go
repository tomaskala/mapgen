package streamline

import (
	"container/heap"
	"math"

	"tomaskala.com/mapgen/field"
	"tomaskala.com/mapgen/vector"
)

const (
	populationScale = 3.0

	degenerateThreshold = 0.01
)

type Streamline struct {
	seed  vector.Vec2
	back  []vector.Vec2
	front []vector.Vec2
}

func (s Streamline) Points() []vector.Vec2 {
	points := make([]vector.Vec2, len(s.back)+1+len(s.front))

	for i, p := range s.back {
		points[len(s.back)-1-i] = p
	}

	points[len(s.back)] = s.seed
	copy(points[len(s.back)+1:], s.front)

	return points
}

type Tracer struct {
	tf         field.TensorField
	population field.NoiseField
	dSep       float64
	dTest      float64
	dLookahead float64
	rkStep     float64
	maxLength  float64
}

func NewTracer(
	tf field.TensorField,
	population field.NoiseField,
	dSep, dTest, dLookahead, rkStep, maxLength float64,
) Tracer {
	return Tracer{
		tf:         tf,
		population: population,
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

func (t Tracer) Run(majorGrid, minorGrid *Grid, seeds []vector.Vec2) Trace {
	priority := func(v vector.Vec2) float64 {
		pop := t.population.Evaluate(v)
		return math.Exp(populationScale * pop)
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

func findSeeds(line Streamline, dSep2 float64) []vector.Vec2 {
	var seeds []vector.Vec2
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
	forward := item.self.sel(t.tf.Evaluate(item.p)).Normalized()
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

func (t Tracer) traceHalfline(item Item, dir vector.Vec2) []vector.Vec2 {
	var halfline []vector.Vec2
	dist := 0.0
	dTest2 := t.dTest * t.dTest
	curr := item.p

	for {
		next := field.RungeKuttaStep(t.tf, dir, curr, t.rkStep, item.self.sel)
		dir = next.Sub(curr).Normalized()
		dist += next.Dist(curr)
		curr = next

		// Stopping criteria (1): out of domain boundary.
		if !item.self.grid.IsInBounds(curr) {
			break
		}

		// Stopping criteria (2): degenerate point.
		if t.tf.Evaluate(curr).Norm2() < degenerateThreshold {
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

func (t Tracer) lookahead(grid *Grid, dir, curr vector.Vec2, sel field.EigenSelector) (vector.Vec2, bool) {
	dist := 0.0
	dTest2 := t.dTest * t.dTest

	for dist < t.dLookahead {
		next := field.RungeKuttaStep(t.tf, dir, curr, t.rkStep, sel)
		dir = next.Sub(curr).Normalized()
		dist += next.Dist(curr)
		curr = next

		if !grid.IsInBounds(curr) {
			break
		}

		if grid.IsTooClose(curr, dTest2) {
			return curr, true
		}
	}

	return vector.Vec2{}, false
}
