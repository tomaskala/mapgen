package main

type config struct {
	numSeeds   int
	dSep       float64
	dTest      float64 // Must be (0.0 * dSep, 1.0 * dSep)
	dLookahead float64
	rkStep     float64
	maxLength  float64
}

var (
	mainRoadCfg = config{
		numSeeds:   30,
		dSep:       200.0,
		dTest:      100.0,
		dLookahead: 300.0,
		rkStep:     1.0,
		maxLength:  1200.0,
	}

	majorRoadCfg = config{
		numSeeds:   30,
		dSep:       80.0,
		dTest:      40.0,
		dLookahead: 100.0,
		rkStep:     1.0,
		maxLength:  1000.0,
	}

	minorRoadCfg = config{
		numSeeds:   30,
		dSep:       20.0,
		dTest:      15.0,
		dLookahead: 40.0,
		rkStep:     1.0,
		maxLength:  800.0,
	}
)
