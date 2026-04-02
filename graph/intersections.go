package graph

import (
	"math"

	"tomaskala.com/mapgen/field"
	"tomaskala.com/mapgen/streamline"
)

type Intersection struct {
	P field.Vector // TODO: Make private when debugging is finished.
}

type Segment struct {
	a field.Vector
	b field.Vector
}

func FindIntersections(width, height int, cellSize float64, major, minor []streamline.Streamline) []Intersection {
	var intersections []Intersection
	grid := NewGrid(width, height, cellSize)

	for _, streamline := range minor {
		points := streamline.Points()

		for i := range len(points) - 1 {
			grid.Add(Segment{points[i], points[i+1]})
		}
	}

	for _, streamline := range major {
		points := streamline.Points()

		for i := range len(points) - 1 {
			segment := Segment{points[i], points[i+1]}
			candidates := grid.Neighbors(segment)

			for candidate := range candidates {
				if t, ok := intersects(segment, candidate); ok {
					point := segment.a.Add(segment.b.Sub(segment.a).Mul(t))
					intersections = append(intersections, Intersection{point})
				}
			}
		}
	}

	return intersections
}

func intersects(s1, s2 Segment) (float64, bool) {
	q := s1.b.Sub(s1.a)
	r := s2.b.Sub(s2.a)

	denom := q.Y*r.X - q.X*r.Y
	if math.Abs(denom) < field.Eps*q.Norm()*r.Norm() {
		return 0.0, false
	}

	ac := s1.a.Sub(s2.a)
	t := (r.Y*ac.X - r.X*ac.Y) / denom
	u := (q.Y*ac.X - q.X*ac.Y) / denom

	return t, 0.0 <= t && t <= 1.0 && 0.0 <= u && u <= 1.0
}
