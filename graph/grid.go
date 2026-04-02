package graph

import (
	"iter"
	"math"

	"tomaskala.com/mapgen/field"
)

type Grid struct {
	width    int
	height   int
	cellSize float64
	cells    [][]Segment
}

func NewGrid(width, height int, cellSize float64) *Grid {
	w := int(math.Ceil(float64(width) / cellSize))
	h := int(math.Ceil(float64(height) / cellSize))
	cells := make([][]Segment, w*h)
	return &Grid{w, h, cellSize, cells}
}

func (g *Grid) Add(s Segment) {
	x0, y0 := g.cell(s.a)
	x1, y1 := g.cell(s.b)

	dx, dy := abs(x1-x0), abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	err := dx - dy
	x, y := x0, y0

	for {
		if 0 <= x && x < g.width && 0 <= y && y < g.height {
			off := g.offset(x, y)
			g.cells[off] = append(g.cells[off], s)
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

func (g *Grid) Neighbors(s Segment) iter.Seq[Segment] {
	return func(yield func(Segment) bool) {
		seen := make(map[Segment]struct{})

		x0, y0 := g.cell(s.a)
		x1, y1 := g.cell(s.b)

		dx, dy := abs(x1-x0), abs(y1-y0)
		sx, sy := sign(x1-x0), sign(y1-y0)
		err := dx - dy
		x, y := x0, y0

		for {
			if !g.iterateNeighborhood(seen, x, y, yield) {
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

func (g *Grid) iterateNeighborhood(seen map[Segment]struct{}, x, y int, yield func(Segment) bool) bool {
	for ny := y - 1; ny <= y+1; ny++ {
		for nx := x - 1; nx <= x+1; nx++ {
			if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
				continue
			}

			for _, neighbor := range g.cells[g.offset(nx, ny)] {
				if _, ok := seen[neighbor]; !ok {
					seen[neighbor] = struct{}{}
					if !yield(neighbor) {
						return false
					}
				}
			}
		}
	}

	return true
}

func (g *Grid) cell(v field.Vector) (int, int) {
	cx := int(v.X / g.cellSize)
	cy := int(v.Y / g.cellSize)
	return cx, cy
}

func (g *Grid) offset(x, y int) int {
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
