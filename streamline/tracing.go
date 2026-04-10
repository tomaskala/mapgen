package streamline

import (
	"container/heap"
	"math"

	"tomaskala.com/mapgen/config"
	"tomaskala.com/mapgen/field"
	"tomaskala.com/mapgen/vector"
)

const (
	populationScale = 3.0

	degenerateThreshold = 0.01
	localDensityDamping = 0.2
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
	cfg        config.Config
}

func NewTracer(tf field.TensorField, population field.NoiseField, cfg config.Config) Tracer {
	return Tracer{tf: tf, population: population, cfg: cfg}
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
		dSep := t.dSep(curr.p)
		if !curr.self.grid.IsInBounds(curr.p) || curr.self.grid.IsTooClose(curr.p, dSep*dSep) {
			continue
		}

		line := t.traceStreamline(curr)
		*curr.self.lines = append(*curr.self.lines, line)

		nextSeeds := t.findSeeds(line)
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

func (t Tracer) findSeeds(line Streamline) []vector.Vec2 {
	var seeds []vector.Vec2
	prev := line.seed

	for _, p := range line.front {
		dSep := t.dSep(p)
		if p.Dist2(prev) >= dSep*dSep {
			seeds = append(seeds, p)
			prev = p
		}
	}

	prev = line.seed
	for _, p := range line.back {
		dSep := t.dSep(p)
		if p.Dist2(prev) >= dSep*dSep {
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
	curr := item.p

	for {
		next := field.RungeKuttaStep(t.tf, dir, curr, t.cfg.RkStep, item.self.sel)
		dir = next.Sub(curr).Normalized()
		dist += next.Dist(curr)
		curr = next
		dTest := t.dTest(curr)

		// Stopping criteria (1): out of domain boundary.
		if !item.self.grid.IsInBounds(curr) {
			break
		}

		// Stopping criteria (2): degenerate point.
		if t.tf.Evaluate(curr).Norm2() < degenerateThreshold {
			break
		}

		// Stopping criteria (3): loop.
		if dist > dTest && curr.Dist2(item.p) < dTest*dTest {
			break
		}

		// Stopping criteria (4): length exceeded.
		if dist > t.cfg.MaxLength {
			if lookahead, found := t.lookahead(item.other.grid, dir, curr, item.self.sel); found {
				halfline = append(halfline, lookahead)
			}
			break
		}

		// Stopping criteria (5): too close to an existing streamline.
		if item.self.grid.IsTooClose(curr, dTest*dTest) {
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
	dLookahead := t.dLookahead(curr)

	for dist < dLookahead {
		next := field.RungeKuttaStep(t.tf, dir, curr, t.cfg.RkStep, sel)
		dir = next.Sub(curr).Normalized()
		dist += next.Dist(curr)
		curr = next
		dTest := t.dTest(curr)
		dLookahead = t.dLookahead(curr)

		if !grid.IsInBounds(curr) {
			break
		}

		if grid.IsTooClose(curr, dTest*dTest) {
			return curr, true
		}
	}

	return vector.Vec2{}, false
}

func (t Tracer) dSep(v vector.Vec2) float64 {
	if t.cfg.ConstDensity {
		return t.cfg.DSep
	}

	population := t.population.Evaluate(v)
	return t.cfg.DSep / (population + localDensityDamping)
}

func (t Tracer) dTest(v vector.Vec2) float64 {
	if t.cfg.ConstDensity {
		return t.cfg.DTest
	}

	population := t.population.Evaluate(v)
	return t.cfg.DTest / (population + localDensityDamping)
}

func (t Tracer) dLookahead(v vector.Vec2) float64 {
	if t.cfg.ConstDensity {
		return t.cfg.DLookahead
	}

	population := t.population.Evaluate(v)
	return t.cfg.DLookahead / (population + localDensityDamping)
}
