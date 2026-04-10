package config

type Config struct {
	NumSeeds     int
	ConstDensity bool
	DSep         float64
	DTest        float64 // Must be (0.0 * dSep, 1.0 * dSep)
	DLookahead   float64
	RkStep       float64
	MaxLength    float64
}
