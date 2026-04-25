package graph

import (
	"cmp"
	"iter"
	"math"
	"slices"

	"tomaskala.com/mapgen/streamline"
	"tomaskala.com/mapgen/vector"
)

type intersection struct {
	pos             vector.Vec2
	id              int
	majorLineOffset float64
	minorLineOffset float64
	majorLineID     int
	minorLineID     int
	majorSegmentID  int
	minorSegmentID  int
}

type segment struct {
	a         vector.Vec2
	b         vector.Vec2
	lineID    int
	segmentID int
}

type VertexID int

type Vertex struct {
	ID  VertexID
	Pos vector.Vec2
}

type Edge struct {
	To       VertexID
	path     []vector.Vec2
	backward bool
}

func (e Edge) Path() iter.Seq[vector.Vec2] {
	if e.backward {
		return func(yield func(vector.Vec2) bool) {
			for _, v := range slices.Backward(e.path) {
				if !yield(v) {
					return
				}
			}
		}
	}

	return func(yield func(vector.Vec2) bool) {
		for _, v := range e.path {
			if !yield(v) {
				return
			}
		}
	}
}

type Graph struct {
	Vertices  []Vertex
	Adjacency [][]Edge
}

func (g *Graph) addVertex(pos vector.Vec2) {
	id := VertexID(len(g.Vertices))
	g.Vertices = append(g.Vertices, Vertex{id, pos})
	g.Adjacency = append(g.Adjacency, nil)
}

func (g *Graph) addEdge(from, to VertexID, path []vector.Vec2) {
	g.Adjacency[from] = append(g.Adjacency[from], Edge{To: to, path: path, backward: false})
	g.Adjacency[to] = append(g.Adjacency[to], Edge{To: from, path: path, backward: true})
}

func BuildGraph(width, height int, cellSize float64, trace streamline.Trace) *Graph {
	g := &Graph{}
	intersections := findIntersections(width, height, cellSize, trace)

	for _, intersection := range intersections {
		g.addVertex(intersection.pos)
	}

	extractEdges(
		g,
		intersections,
		trace.Major,
		func(i intersection) int { return i.majorLineID },
		func(i intersection) int { return i.majorSegmentID },
		func(i intersection) float64 { return i.majorLineOffset },
	)

	extractEdges(
		g,
		intersections,
		trace.Minor,
		func(i intersection) int { return i.minorLineID },
		func(i intersection) int { return i.minorSegmentID },
		func(i intersection) float64 { return i.minorLineOffset },
	)

	return g
}

func findIntersections(width, height int, cellSize float64, trace streamline.Trace) []intersection {
	var intersections []intersection
	grid := newGrid(width, height, cellSize)

	for i, streamline := range trace.Minor {
		points := streamline.Points()

		for p := range len(points) - 1 {
			grid.add(segment{a: points[p], b: points[p+1], lineID: i, segmentID: p})
		}
	}

	for i, streamline := range trace.Major {
		points := streamline.Points()

		for p := range len(points) - 1 {
			seg := segment{a: points[p], b: points[p+1], lineID: i, segmentID: p}
			candidates := grid.neighbors(seg)

			for candidate := range candidates {
				t, u, ok := intersects(seg, candidate)
				if !ok {
					continue
				}

				point := seg.a.Add(seg.b.Sub(seg.a).Mul(t))
				id := len(intersections)
				intersections = append(intersections, intersection{
					pos:             point,
					id:              id,
					majorLineOffset: t,
					minorLineOffset: u,
					majorLineID:     i,
					minorLineID:     candidate.lineID,
					majorSegmentID:  p,
					minorSegmentID:  candidate.segmentID,
				})
			}
		}
	}

	return intersections
}

type idSelector func(intersection) int

type offsetSelector func(intersection) float64

func extractEdges(
	g *Graph,
	intersections []intersection,
	streamlines []streamline.Streamline,
	lineIDSelector, segmentIDSelector idSelector,
	lineOffsetSelector offsetSelector,
) {
	groups := make([][]intersection, len(streamlines))

	for _, intersection := range intersections {
		id := lineIDSelector(intersection)
		groups[id] = append(groups[id], intersection)
	}

	for gID, group := range groups {
		slices.SortFunc(group, func(i1, i2 intersection) int {
			seg1 := segmentIDSelector(i1)
			seg2 := segmentIDSelector(i2)

			if seg1 != seg2 {
				return cmp.Compare(seg1, seg2)
			}

			return cmp.Compare(lineOffsetSelector(i1), lineOffsetSelector(i2))
		})

		points := streamlines[gID].Points()
		for i := range len(group) - 1 {
			var path []vector.Vec2

			startSeg := segmentIDSelector(group[i])
			endSeg := segmentIDSelector(group[i+1])

			path = append(path, group[i].pos)
			for segID := startSeg + 1; segID <= endSeg; segID++ {
				path = append(path, points[segID])
			}
			path = append(path, group[i+1].pos)

			g.addEdge(VertexID(group[i].id), VertexID(group[i+1].id), path)
		}
	}
}

func intersects(s1, s2 segment) (float64, float64, bool) {
	q := s1.b.Sub(s1.a)
	r := s2.b.Sub(s2.a)

	denom := q.Y*r.X - q.X*r.Y
	if math.Abs(denom) < vector.Eps*q.Norm()*r.Norm() {
		return 0.0, 0.0, false
	}

	ac := s1.a.Sub(s2.a)
	t := (r.Y*ac.X - r.X*ac.Y) / denom
	u := (q.Y*ac.X - q.X*ac.Y) / denom

	return t, u, 0.0 <= t && t <= 1.0 && 0.0 <= u && u <= 1.0
}
