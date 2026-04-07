package field

import "github.com/ojrac/opensimplex-go"

type Population struct {
	noise   opensimplex.Noise
	scale   float64
	octaves int
}

func NewPopulation(scale float64, octaves int, seed int64) Population {
	return Population{opensimplex.NewNormalized(seed), scale, octaves}
}

func (p Population) Density(v Vector) float64 {
	val := 0.0
	amp := 1.0
	totalAmp := 0.0
	freq := p.scale

	for range p.octaves {
		val += p.noise.Eval2(v.X*freq, v.Y*freq) * amp
		totalAmp += amp
		amp *= 0.5
		freq *= 2.0
	}

	return val / totalAmp
}
