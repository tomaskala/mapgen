package main

import (
	"math"
	"math/rand/v2"

	"tomaskala.com/mapgen/city"
	"tomaskala.com/mapgen/field"
	"tomaskala.com/mapgen/graph"
	"tomaskala.com/mapgen/streamline"
	"tomaskala.com/mapgen/vector"
)

const (
	initialPopFraction = 0.6
	initialPopStep     = 4
	initialPopScale    = 0.002
	initialPopOctaves  = 3

	minGrid = 2
	maxGrid = 4

	minRadial = 1
	maxRadial = 2
)

func buildCityMap(width, height int, rng *rand.Rand) city.City {
	population := field.Noise(initialPopScale, initialPopOctaves, rng.Int64())
	tf := initializeTensorField(width, height, population, rng)

	var fullTrace streamline.Trace
	mainStreamlines := traceStreamlines(width, height, tf, population, fullTrace, mainRoadCfg, rng)

	fullTrace.Major = append(fullTrace.Major, mainStreamlines.Major...)
	fullTrace.Minor = append(fullTrace.Minor, mainStreamlines.Minor...)
	majorStreamlines := traceStreamlines(width, height, tf, population, fullTrace, majorRoadCfg, rng)

	fullTrace.Major = append(fullTrace.Major, majorStreamlines.Major...)
	fullTrace.Minor = append(fullTrace.Minor, majorStreamlines.Minor...)
	minorStreamlines := traceStreamlines(width, height, tf, population, fullTrace, minorRoadCfg, rng)

	mainGraph := graph.BuildGraph(width, height, mainRoadCfg.dSep, mainStreamlines)
	majorGraph := graph.BuildGraph(width, height, majorRoadCfg.dSep, majorStreamlines)
	minorGraph := graph.BuildGraph(width, height, minorRoadCfg.dSep, minorStreamlines)

	return city.City{MainRoads: mainGraph, MajorRoads: majorGraph, MinorRoads: minorGraph}
}

func traceStreamlines(
	width, height int,
	tf field.TensorField,
	population field.NoiseField,
	previous streamline.Trace,
	cfg config,
	rng *rand.Rand,
) streamline.Trace {
	seeds := make([]vector.Vec2, cfg.numSeeds)
	for i := range seeds {
		seeds[i] = vector.Vec2{
			X: rng.Float64() * float64(width),
			Y: rng.Float64() * float64(height),
		}
	}

	majorGrid := streamline.NewGrid(width, height, cfg.dSep)
	for _, major := range previous.Major {
		majorGrid.AddAll(major.Points())
	}

	minorGrid := streamline.NewGrid(width, height, cfg.dSep)
	for _, minor := range previous.Minor {
		minorGrid.AddAll(minor.Points())
	}

	tracer := streamline.NewTracer(tf, population, cfg.dSep, cfg.dTest, cfg.dLookahead, cfg.rkStep, cfg.maxLength)
	return tracer.Run(majorGrid, minorGrid, seeds)
}

func initializeTensorField(width, height int, population field.NoiseField, rng *rand.Rand) field.TensorField {
	minPop := math.MaxFloat64
	maxPop := -math.MaxFloat64

	for y := 0; y < height; y += initialPopStep {
		for x := 0; x < width; x += initialPopStep {
			v := vector.Vec2{X: float64(x), Y: float64(y)}
			density := population.Evaluate(v)
			minPop = min(minPop, density)
			maxPop = max(maxPop, density)
		}
	}

	var candidates []vector.Vec2
	for y := 0; y < height; y += initialPopStep {
		for x := 0; x < width; x += initialPopStep {
			candidate := vector.Vec2{X: float64(x), Y: float64(y)}
			if population.Evaluate(candidate) > (maxPop-minPop)*initialPopFraction {
				candidates = append(candidates, candidate)
			}
		}
	}

	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	makeGrid := func(p vector.Vec2) field.BasisField {
		theta := rng.Float64() * math.Pi
		dir := vector.Vec2{X: math.Cos(theta), Y: math.Sin(theta)}
		radius := (0.2 + 0.3*rng.Float64()) * float64(width)
		return field.Grid(p, dir, radius)
	}

	makeRadial := func(p vector.Vec2) field.BasisField {
		radius := (0.1 + 0.15*rng.Float64()) * float64(width)
		return field.Radial(p, radius)
	}

	numGrid := minGrid + rng.IntN(maxGrid-minGrid+1)
	numRadial := minRadial + rng.IntN(maxRadial-minRadial+1)

	if len(candidates) < numGrid+numRadial {
		panic("not enough candidate points for sampling tensor fields")
	}

	var tf field.TensorField

	for i, p := range candidates {
		if i == numGrid {
			break
		}
		tf = append(tf, makeGrid(p))
	}

	for i, p := range candidates[numGrid:] {
		if i == numRadial {
			break
		}
		tf = append(tf, makeRadial(p))
	}

	return tf
}
