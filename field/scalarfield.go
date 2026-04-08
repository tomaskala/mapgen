package field

import (
	"github.com/ojrac/opensimplex-go"
	"tomaskala.com/mapgen/vector"
)

type NoiseField struct {
	noise   opensimplex.Noise
	scale   float64
	octaves int
}

func Noise(scale float64, octaves int, seed int64) NoiseField {
	return NoiseField{opensimplex.NewNormalized(seed), scale, octaves}
}

func (sf NoiseField) Evaluate(v vector.Vec2) float64 {
	val := 0.0
	amp := 1.0
	totalAmp := 0.0
	freq := sf.scale

	for range sf.octaves {
		val += sf.noise.Eval2(v.X*freq, v.Y*freq) * amp
		totalAmp += amp
		amp *= 0.5
		freq *= 2.0
	}

	return val / totalAmp
}
