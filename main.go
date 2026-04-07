package main

import (
	"flag"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/fogleman/gg"
	"tomaskala.com/mapgen/field"
	"tomaskala.com/mapgen/graph"
	"tomaskala.com/mapgen/renderer"
	"tomaskala.com/mapgen/streamline"
)

const (
	exitSuccess = 0
	exitIOError = 1

	initialPopFraction = 0.6
	initialPopStep     = 4
)

var (
	width  = flag.Int("width", 800, "image width in pixels")
	height = flag.Int("height", 800, "image height in pixels")

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

type config struct {
	numSeeds   int
	dSep       float64
	dTest      float64 // Must be (0.0 * dSep, 1.0 * dSep)
	dLookahead float64
	rkStep     float64
	maxLength  float64
}

var mainRoadCfg = config{
	numSeeds:   30,
	dSep:       200.0,
	dTest:      100.0,
	dLookahead: 300.0,
	rkStep:     1.0,
	maxLength:  1200.0,
}

var majorRoadCfg = config{
	numSeeds:   30,
	dSep:       80.0,
	dTest:      40.0,
	dLookahead: 100.0,
	rkStep:     1.0,
	maxLength:  1000.0,
}

var minorRoadCfg = config{
	numSeeds:   30,
	dSep:       20.0,
	dTest:      15.0,
	dLookahead: 40.0,
	rkStep:     1.0,
	maxLength:  800.0,
}

func sampleTensorField(population field.Population, rng *rand.Rand) field.TensorField {
	minPop := math.MaxFloat64
	maxPop := -math.MaxFloat64

	for y := 0; y < *height; y += initialPopStep {
		for x := 0; x < *width; x += initialPopStep {
			v := field.Vector{X: float64(x), Y: float64(y)}
			density := population.Density(v)
			minPop = min(minPop, density)
			maxPop = max(maxPop, density)
		}
	}

	var candidates []field.Vector
	for y := 0; y < *height; y += initialPopStep {
		for x := 0; x < *width; x += initialPopStep {
			candidate := field.Vector{X: float64(x), Y: float64(y)}
			if population.Density(candidate) > (maxPop-minPop)*initialPopFraction {
				candidates = append(candidates, candidate)
			}
		}
	}

	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	makeGrid := func(p field.Vector) field.BasisField {
		theta := rng.Float64() * math.Pi
		dir := field.Vector{X: math.Cos(theta), Y: math.Sin(theta)}
		radius := (0.2 + 0.3*rng.Float64()) * float64(*width)
		return field.Grid(p, dir, radius)
	}

	makeRadial := func(p field.Vector) field.BasisField {
		radius := (0.1 + 0.15*rng.Float64()) * float64(*width)
		return field.Radial(p, radius)
	}

	numGrid := 2 + rng.IntN(3)
	numRadial := 1 + rng.IntN(2)
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

func trace(
	tf field.TensorField,
	population field.Population,
	cfg config,
	previous streamline.Trace,
	rng *rand.Rand,
) streamline.Trace {
	seeds := make([]field.Vector, cfg.numSeeds)
	for i := range seeds {
		seeds[i] = field.Vector{
			X: rng.Float64() * float64(*width),
			Y: rng.Float64() * float64(*height),
		}
	}

	majorGrid := streamline.NewGrid(*width, *height, cfg.dSep)
	for _, major := range previous.Major {
		majorGrid.AddAll(major.Points())
	}

	minorGrid := streamline.NewGrid(*width, *height, cfg.dSep)
	for _, minor := range previous.Minor {
		minorGrid.AddAll(minor.Points())
	}

	tracer := streamline.NewTracer(tf, population, cfg.dSep, cfg.dTest, cfg.dLookahead, cfg.rkStep, cfg.maxLength)
	return tracer.Run(majorGrid, minorGrid, seeds)
}

func run() int {
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Print("could not create CPU profile: ", err)
			return exitIOError
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Print("could not start CPU profile: ", err)
			return exitIOError
		}
		defer pprof.StopCPUProfile()
	}

	rng := rand.New(rand.NewPCG(1234, 1337))
	output := "image.png"

	population := field.NewPopulation(0.002, 3, int64(9432))
	tf := sampleTensorField(population, rng)

	var fullTrace streamline.Trace
	mainStreamlines := trace(tf, population, mainRoadCfg, fullTrace, rng)

	fullTrace.Major = append(fullTrace.Major, mainStreamlines.Major...)
	fullTrace.Minor = append(fullTrace.Minor, mainStreamlines.Minor...)
	majorStreamlines := trace(tf, population, majorRoadCfg, fullTrace, rng)

	fullTrace.Major = append(fullTrace.Major, majorStreamlines.Major...)
	fullTrace.Minor = append(fullTrace.Minor, majorStreamlines.Minor...)
	minorStreamlines := trace(tf, population, minorRoadCfg, fullTrace, rng)

	mainGraph := graph.BuildGraph(*width, *height, mainRoadCfg.dSep, mainStreamlines)
	majorGraph := graph.BuildGraph(*width, *height, majorRoadCfg.dSep, majorStreamlines)
	minorGraph := graph.BuildGraph(*width, *height, minorRoadCfg.dSep, minorStreamlines)

	dc := gg.NewContext(*width, *height)

	// Draw thicker in black first for the borders.
	renderer.RenderGraph(dc, minorGraph)
	dc.SetHexColor("#000000")
	dc.SetLineWidth(4)
	dc.Stroke()

	renderer.RenderGraph(dc, majorGraph)
	dc.SetHexColor("#000000")
	dc.SetLineWidth(6)
	dc.Stroke()

	renderer.RenderGraph(dc, mainGraph)
	dc.SetHexColor("#000000")
	dc.SetLineWidth(8)
	dc.Stroke()

	// Draw thinner in colors for the fill.
	renderer.RenderGraph(dc, minorGraph)
	dc.SetHexColor("#dcdcdc")
	dc.SetLineWidth(2)
	dc.Stroke()

	renderer.RenderGraph(dc, majorGraph)
	dc.SetHexColor("#708090")
	dc.SetLineWidth(4)
	dc.Stroke()

	renderer.RenderGraph(dc, mainGraph)
	dc.SetHexColor("#fff600")
	dc.SetLineWidth(6)
	dc.Stroke()

	if err := dc.SavePNG(output); err != nil {
		return exitIOError
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Print("could not create memory profile: ", err)
			return exitIOError
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			log.Print("could not write memory profile: ", err)
			return exitIOError
		}
	}

	return exitSuccess
}

func main() {
	flag.Parse()
	os.Exit(run())
}
