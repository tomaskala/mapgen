package streamline

import (
	"math"

	"tomaskala.com/mapgen/field"
)

type Grid struct {
	width  int
	height int
	dsep   float64
	grid   [][]field.Vector
}

func NewGrid(width, height int, dsep float64) *Grid {
	w := int(math.Ceil(float64(width) / dsep))
	h := int(math.Ceil(float64(height) / dsep))
	grid := make([][]field.Vector, w*h)
	return &Grid{w, h, dsep, grid}
}

func (g *Grid) Add(v field.Vector) bool {
	cx, cy, ok := g.cell(v)
	if !ok {
		return false
	}

	off := g.offset(cx, cy)
	g.grid[off] = append(g.grid[off], v)
	return true
}

func (g *Grid) Neighbors(v field.Vector) []field.Vector {
	cx, cy, ok := g.cell(v)
	if !ok {
		return nil
	}
	var neighbors []field.Vector

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := cx+dx, cy+dy
			if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
				continue
			}
			neighbors = append(neighbors, g.grid[g.offset(nx, ny)]...)
		}
	}

	return neighbors
}

func (g *Grid) IsTooClose(v field.Vector, minDist float64) bool {
	cx, cy, ok := g.cell(v)
	if !ok {
		return false
	}
	minDist2 := minDist * minDist

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := cx+dx, cy+dy
			if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
				continue
			}

			neighbors := g.grid[g.offset(nx, ny)]
			for _, n := range neighbors {
				if v.Sub(n).NormSquared() < minDist2 {
					return true
				}
			}
		}
	}

	return false
}

func (g *Grid) cell(v field.Vector) (int, int, bool) {
	cx := int(v.X / g.dsep)
	cy := int(v.Y / g.dsep)

	if cx < 0 || cx >= g.width || cy < 0 || cy >= g.height {
		return 0, 0, false
	}
	return cx, cy, true
}

func (g *Grid) offset(x, y int) int {
	return y*g.width + x
}
