package poissondisc

import (
	"math"

	"tomaskala.com/mapgen/field"
)

var sentinel = field.Vector{X: math.Inf(-1), Y: math.Inf(-1)}

type grid struct {
	width    int
	height   int
	cellSize float64
	radius2  float64
	cells    []field.Vector
}

func newGrid(width, height int, r float64) *grid {
	size := r / math.Sqrt(2)

	w := int(math.Floor(float64(width)/size)) + 1
	h := int(math.Floor(float64(height)/size)) + 1

	cells := make([]field.Vector, w*h)
	for i := range cells {
		cells[i] = sentinel
	}

	return &grid{width: w, height: h, cellSize: size, radius2: r * r, cells: cells}
}

func (g *grid) add(v field.Vector) bool {
	cx, cy := g.cell(v)
	w := g.cells[g.offset(cx, cy)]

	if w != sentinel && v.Dist2(w) < g.radius2 {
		return false
	}

	x0 := max(0, cx-2)
	y0 := max(0, cy-2)
	x1 := min(g.width, cx+3)
	y1 := min(g.height, cy+3)

	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			w := g.cells[g.offset(x, y)]

			if w != sentinel && v.Dist2(w) < g.radius2 {
				return false
			}
		}
	}

	g.cells[cy*g.width+cx] = v
	return true
}

func (g *grid) cell(v field.Vector) (int, int) {
	cx := int(v.X / g.cellSize)
	cy := int(v.Y / g.cellSize)
	return cx, cy
}

func (g *grid) offset(x, y int) int {
	return y*g.width + x
}
