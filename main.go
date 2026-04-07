package main

import (
	"flag"
	"log"
	"math"
	oldrand "math/rand"
	"math/rand/v2"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/fogleman/gg"
	"github.com/fogleman/poissondisc"
	"tomaskala.com/mapgen/field"
	"tomaskala.com/mapgen/graph"
	"tomaskala.com/mapgen/renderer"
	"tomaskala.com/mapgen/streamline"
)

const (
	exitSuccess = 0
	exitIOError = 1
)

var (
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

func sampleTensorField(width, height int, r float64, rng *oldrand.Rand) field.TensorField {
	mainAngle := rng.Float64() * math.Pi / 2.0
	numGrid := 2 + rng.Intn(3)
	numRadial := 1 + rng.Intn(2)

	tf := make(field.TensorField, numGrid+numRadial)
	points := poissondisc.Sample(0.0, 0.0, float64(width), float64(height), r, 30, rng)
	rng.Shuffle(len(points), func(i, j int) {
		points[i], points[j] = points[j], points[i]
	})

	for i, p := range points[:numGrid] {
		center := field.Vector{X: p.X, Y: p.Y}
		theta := mainAngle + rng.NormFloat64()*math.Pi/24.0
		dir := field.Vector{X: math.Cos(theta), Y: math.Sin(theta)}
		radius := (0.5 + rng.Float64()*0.5) * float64(width)
		tf[i] = field.Grid(center, dir, radius)
	}

	for i := range numRadial {
		var center field.Vector
		if rng.Float64() < 0.4 {
			center = field.Vector{
				X: float64(width) * rng.Float64() * 0.3,
				Y: float64(height) * rng.Float64(),
			}
		} else {
			center = field.Vector{
				X: float64(width) * (0.1 + rng.Float64()*0.8),
				Y: float64(height) * (0.1 + rng.Float64()*0.8),
			}
		}
		radius := (0.25 + rng.Float64()*0.2) * float64(width)
		tf[numGrid+i] = field.Radial(center, radius)
	}

	return tf
}

func trace(
	width, height int,
	tf field.TensorField,
	cfg config,
	previous streamline.Trace,
	rng *rand.Rand,
) streamline.Trace {
	seeds := make([]field.Vector, cfg.numSeeds)
	for i := range seeds {
		seeds[i] = field.Vector{
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

	tracer := streamline.NewTracer(tf, cfg.dSep, cfg.dTest, cfg.dLookahead, cfg.rkStep, cfg.maxLength)
	return tracer.Run(majorGrid, minorGrid, seeds)
}

func debugGraph(output string, width, height int, tf field.TensorField, cfg config, rng *rand.Rand) int {
	majorGrid := streamline.NewGrid(width, height, cfg.dSep)
	minorGrid := streamline.NewGrid(width, height, cfg.dSep)
	seeds := make([]field.Vector, cfg.numSeeds)
	for i := range cfg.numSeeds {
		seeds[i] = field.Vector{
			X: rng.Float64() * float64(width),
			Y: rng.Float64() * float64(height),
		}
	}

	tracer := streamline.NewTracer(tf, cfg.dSep, cfg.dTest, cfg.dLookahead, cfg.rkStep, cfg.maxLength)
	trace := tracer.Run(majorGrid, minorGrid, seeds)

	g := graph.BuildGraph(width, height, cfg.dSep, trace)

	dc := gg.NewContext(width, height)
	renderer.DebugGraph(dc, trace, g)

	if err := dc.SavePNG(output); err != nil {
		return exitIOError
	}

	return exitSuccess
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
	oldRng := oldrand.New(oldrand.NewSource(5678))
	output := "image.png"
	width := 800
	height := 800

	tf := sampleTensorField(width, height, 50.0, oldRng)

	// debugGraph(output, width, height, tf, mainRoadCfg, rng)

	var fullTrace streamline.Trace
	mainStreamlines := trace(width, height, tf, mainRoadCfg, fullTrace, rng)

	fullTrace.Major = append(fullTrace.Major, mainStreamlines.Major...)
	fullTrace.Minor = append(fullTrace.Minor, mainStreamlines.Minor...)
	majorStreamlines := trace(width, height, tf, majorRoadCfg, fullTrace, rng)

	fullTrace.Major = append(fullTrace.Major, majorStreamlines.Major...)
	fullTrace.Minor = append(fullTrace.Minor, majorStreamlines.Minor...)
	minorStreamlines := trace(width, height, tf, minorRoadCfg, fullTrace, rng)

	mainGraph := graph.BuildGraph(width, height, mainRoadCfg.dSep, mainStreamlines)
	majorGraph := graph.BuildGraph(width, height, majorRoadCfg.dSep, majorStreamlines)
	minorGraph := graph.BuildGraph(width, height, minorRoadCfg.dSep, minorStreamlines)

	dc := gg.NewContext(width, height)

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
