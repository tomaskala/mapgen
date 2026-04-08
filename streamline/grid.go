package streamline

import (
	"math"

	"tomaskala.com/mapgen/vector"
)

type Grid struct {
	width    int
	height   int
	cellSize float64
	cells    [][]vector.Vec2
}

func NewGrid(width, height int, cellSize float64) *Grid {
	w := int(math.Ceil(float64(width) / cellSize))
	h := int(math.Ceil(float64(height) / cellSize))
	cells := make([][]vector.Vec2, w*h)
	return &Grid{w, h, cellSize, cells}
}

func (g *Grid) Add(v vector.Vec2) {
	if !g.IsInBounds(v) {
		return
	}

	cx, cy := g.cell(v)
	off := g.offset(cx, cy)
	g.cells[off] = append(g.cells[off], v)
}

func (g *Grid) AddAll(vs []vector.Vec2) {
	for _, v := range vs {
		g.Add(v)
	}
}

func (g *Grid) IsInBounds(v vector.Vec2) bool {
	if v.X < 0 || v.Y < 0 {
		return false
	}

	cx := int(v.X / g.cellSize)
	cy := int(v.Y / g.cellSize)

	return cx >= 0 && cx < g.width && cy >= 0 && cy < g.height
}

func (g *Grid) IsTooClose(v vector.Vec2, minDist2 float64) bool {
	if !g.IsInBounds(v) {
		return false
	}

	cx, cy := g.cell(v)

	for ny := cy - 1; ny <= cy+1; ny++ {
		for nx := cx - 1; nx <= cx+1; nx++ {
			if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
				continue
			}

			for _, neighbor := range g.cells[g.offset(nx, ny)] {
				if v.Dist2(neighbor) < minDist2 {
					return true
				}
			}
		}
	}

	return false
}

func (g *Grid) cell(v vector.Vec2) (int, int) {
	cx := int(v.X / g.cellSize)
	cy := int(v.Y / g.cellSize)
	return cx, cy
}

func (g *Grid) offset(x, y int) int {
	return y*g.width + x
}
