package poissondisc

import (
	"math"
	"math/rand/v2"

	"tomaskala.com/mapgen/field"
)

func Sample(width, height int, r float64, k int, rng *rand.Rand) []field.Vector {
	var result []field.Vector
	var active []field.Vector
	grid := newGrid(width, height, r)

	x := float64(width) / 2.0
	y := float64(height) / 2.0
	p := field.Vector{X: x, Y: y}
	grid.add(p)
	active = append(active, p)
	result = append(result, p)

	for len(active) > 0 {
		index := rng.IntN(len(active))
		point := active[index]
		ok := false

		for range k {
			a := rng.Float64() * 2.0 * math.Pi
			d := r * math.Sqrt(rng.Float64()*3.0+1.0)

			x := point.X + math.Cos(a)*d
			y := point.Y + math.Sin(a)*d
			if x < 0.0 || int(x) > width || y < 0.0 || int(y) > height {
				continue
			}

			p := field.Vector{X: x, Y: y}
			if !grid.add(p) {
				continue
			}

			result = append(result, p)
			active = append(active, p)
			ok = true
			break
		}

		if !ok {
			active[index] = active[len(active)-1]
			active = active[:len(active)-1]
		}
	}

	return result
}
