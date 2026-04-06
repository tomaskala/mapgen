package graph

import (
	"cmp"
	"math"
	"slices"

	"tomaskala.com/mapgen/field"
	"tomaskala.com/mapgen/streamline"
)

type intersection struct {
	pos             field.Vector
	id              int
	majorLineOffset float64
	minorLineOffset float64
	majorLineID     int
	minorLineID     int
	majorSegmentID  int
	minorSegmentID  int
}

type segment struct {
	a         field.Vector
	b         field.Vector
	lineID    int
	segmentID int
}

type VertexID int

type EdgeID int

type Vertex struct {
	Pos   field.Vector
	Edges []EdgeID
}

type Edge struct {
	A    VertexID
	B    VertexID
	Path []field.Vector
}

type Graph struct {
	Vertices []Vertex
	Edges    []Edge
}

func BuildGraph(width, height int, cellSize float64, major, minor []streamline.Streamline) Graph {
	intersections := findIntersections(width, height, cellSize, major, minor)

	vertices := make([]Vertex, len(intersections))
	for i, intersection := range intersections {
		vertices[i] = Vertex{Pos: intersection.pos}
	}

	var edges []Edge
	edges = append(
		edges,
		extractEdges(
			intersections,
			major,
			func(i intersection) int { return i.majorLineID },
			func(i intersection) int { return i.majorSegmentID },
			func(i intersection) float64 { return i.majorLineOffset },
		)...)
	edges = append(
		edges,
		extractEdges(
			intersections,
			minor,
			func(i intersection) int { return i.minorLineID },
			func(i intersection) int { return i.minorSegmentID },
			func(i intersection) float64 { return i.minorLineOffset },
		)...)

	for id, edge := range edges {
		vertices[edge.A].Edges = append(vertices[edge.A].Edges, EdgeID(id))
		vertices[edge.B].Edges = append(vertices[edge.B].Edges, EdgeID(id))
	}

	return Graph{
		Vertices: vertices,
		Edges:    edges,
	}
}

func findIntersections(width, height int, cellSize float64, major, minor []streamline.Streamline) []intersection {
	var intersections []intersection
	grid := newGrid(width, height, cellSize)

	for i, streamline := range minor {
		points := streamline.Points()

		for p := range len(points) - 1 {
			grid.add(segment{a: points[p], b: points[p+1], lineID: i, segmentID: p})
		}
	}

	for i, streamline := range major {
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
	intersections []intersection,
	streamlines []streamline.Streamline,
	lineIDSelector, segmentIDSelector idSelector,
	lineOffsetSelector offsetSelector,
) []Edge {
	var edges []Edge
	groups := make([][]intersection, len(streamlines))

	for _, intersection := range intersections {
		id := lineIDSelector(intersection)
		groups[id] = append(groups[id], intersection)
	}

	for g, group := range groups {
		slices.SortFunc(group, func(i1, i2 intersection) int {
			seg1 := segmentIDSelector(i1)
			seg2 := segmentIDSelector(i2)

			if seg1 != seg2 {
				return cmp.Compare(seg1, seg2)
			}

			return cmp.Compare(lineOffsetSelector(i1), lineOffsetSelector(i2))
		})

		points := streamlines[g].Points()
		for i := range len(group) - 1 {
			var path []field.Vector

			startSeg := segmentIDSelector(group[i])
			endSeg := segmentIDSelector(group[i+1])

			path = append(path, group[i].pos)
			for segID := startSeg + 1; segID <= endSeg; segID++ {
				path = append(path, points[segID])
			}
			path = append(path, group[i+1].pos)

			edges = append(edges, Edge{A: VertexID(group[i].id), B: VertexID(group[i+1].id), Path: path})
		}
	}

	return edges
}

func intersects(s1, s2 segment) (float64, float64, bool) {
	q := s1.b.Sub(s1.a)
	r := s2.b.Sub(s2.a)

	denom := q.Y*r.X - q.X*r.Y
	if math.Abs(denom) < field.Eps*q.Norm()*r.Norm() {
		return 0.0, 0.0, false
	}

	ac := s1.a.Sub(s2.a)
	t := (r.Y*ac.X - r.X*ac.Y) / denom
	u := (q.Y*ac.X - q.X*ac.Y) / denom

	return t, u, 0.0 <= t && t <= 1.0 && 0.0 <= u && u <= 1.0
}
