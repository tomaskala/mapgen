package graph

import (
	"math"
)

const maxLength = 100

type Polygon struct {
	Vertices []Vertex
	Edges    []Edge
}

type directedEdge struct {
	from VertexID
	to   VertexID
}

func DetectPolygons(g *Graph) []Polygon {
	var polygons []Polygon
	visited := make(map[directedEdge]struct{})

	for i, v := range g.Vertices {
		if len(g.Adjacency[i]) < 2 {
			continue
		}

		for _, edge := range g.Adjacency[i] {
			directed := directedEdge{VertexID(i), edge.To}
			if _, ok := visited[directed]; ok {
				continue
			}
			visited[directed] = struct{}{}

			vertices := []Vertex{v, g.Vertices[edge.To]}
			edges := []Edge{edge}
			failed := false
			prevVertex := edge.To

			for {
				if len(vertices) >= maxLength {
					failed = true
					break
				}

				nextVertex, nextEdge, ok := turnRight(g, vertices[len(vertices)-2], vertices[len(vertices)-1])
				if !ok {
					failed = true
					break
				}

				visited[directedEdge{prevVertex, nextEdge.To}] = struct{}{}
				prevVertex = nextEdge.To

				edges = append(edges, nextEdge)
				if nextVertex.ID == vertices[0].ID {
					break
				}
				vertices = append(vertices, nextVertex)
			}

			if failed {
				continue
			}

			polygons = append(polygons, Polygon{vertices, edges})
		}
	}

	return polygons
}

func turnRight(g *Graph, v, w Vertex) (Vertex, Edge, bool) {
	diff := v.Pos.Sub(w.Pos)
	angle := math.Atan2(diff.Y, diff.X)

	var nextVertex Vertex
	var nextEdge Edge
	found := false
	minAngle := 2.0 * math.Pi

	for _, edge := range g.Adjacency[w.ID] {
		if edge.To == v.ID {
			continue
		}

		n := g.Vertices[edge.To]
		nextDiff := n.Pos.Sub(w.Pos)
		nextAngle := math.Atan2(nextDiff.Y, nextDiff.X) - angle

		if nextAngle < 0.0 {
			nextAngle += 2.0 * math.Pi
		}

		if nextAngle < minAngle {
			nextVertex = g.Vertices[edge.To]
			nextEdge = edge
			found = true
			minAngle = nextAngle
		}
	}

	return nextVertex, nextEdge, found
}
