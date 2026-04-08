package graph

import (
	"iter"
	"math"

	"tomaskala.com/mapgen/vector"
)

type queryCount int

type segmentWithID struct {
	id  int
	seg segment
}

type grid struct {
	width      int
	height     int
	cellSize   float64
	cells      [][]segmentWithID
	seen       []queryCount
	segmentID  int
	queryCount queryCount
}

func newGrid(width, height int, cellSize float64) *grid {
	w := int(math.Ceil(float64(width) / cellSize))
	h := int(math.Ceil(float64(height) / cellSize))
	cells := make([][]segmentWithID, w*h)
	return &grid{width: w, height: h, cellSize: cellSize, cells: cells}
}

func (g *grid) add(s segment) {
	if !g.isInBounds(s.a) || !g.isInBounds(s.b) {
		return
	}

	if len(g.seen) == g.segmentID {
		g.seen = append(g.seen, 0)
	}
	g.segmentID++

	x0, y0 := g.cell(s.a)
	x1, y1 := g.cell(s.b)

	dx, dy := abs(x1-x0), abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	err := dx - dy
	x, y := x0, y0

	for {
		if 0 <= x && x < g.width && 0 <= y && y < g.height {
			off := g.offset(x, y)
			g.cells[off] = append(g.cells[off], segmentWithID{id: g.segmentID - 1, seg: s})
		}

		if x == x1 && y == y1 {
			break
		}

		e2 := 2 * err

		if e2 > -dy {
			err -= dy
			x += sx
		}

		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func (g *grid) neighbors(s segment) iter.Seq[segment] {
	return func(yield func(segment) bool) {
		if !g.isInBounds(s.a) || !g.isInBounds(s.b) {
			return
		}

		g.queryCount++

		x0, y0 := g.cell(s.a)
		x1, y1 := g.cell(s.b)

		dx, dy := abs(x1-x0), abs(y1-y0)
		sx, sy := sign(x1-x0), sign(y1-y0)
		err := dx - dy
		x, y := x0, y0

		for {
			if !g.iterateNeighborhood(x, y, yield) {
				return
			}

			if x == x1 && y == y1 {
				break
			}

			e2 := 2 * err

			if e2 > -dy {
				err -= dy
				x += sx
			}

			if e2 < dx {
				err += dx
				y += sy
			}
		}
	}
}

func (g *grid) iterateNeighborhood(x, y int, yield func(segment) bool) bool {
	for ny := y - 1; ny <= y+1; ny++ {
		for nx := x - 1; nx <= x+1; nx++ {
			if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
				continue
			}

			for _, neighbor := range g.cells[g.offset(nx, ny)] {
				if g.seen[neighbor.id] == g.queryCount {
					continue
				}

				g.seen[neighbor.id] = g.queryCount
				if !yield(neighbor.seg) {
					return false
				}
			}
		}
	}

	return true
}

func (g *grid) isInBounds(v vector.Vec2) bool {
	if v.X < 0 || v.Y < 0 {
		return false
	}

	cx := int(v.X / g.cellSize)
	cy := int(v.Y / g.cellSize)

	return cx >= 0 && cx < g.width && cy >= 0 && cy < g.height
}

func (g *grid) cell(v vector.Vec2) (int, int) {
	cx := int(v.X / g.cellSize)
	cy := int(v.Y / g.cellSize)
	return cx, cy
}

func (g *grid) offset(x, y int) int {
	return y*g.width + x
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	switch {
	case x < 0:
		return -1
	case x > 0:
		return 1
	default:
		return 0
	}
}
