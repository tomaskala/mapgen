package main

import (
	"flag"
	"log"
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
	exitSuccess      = 0
	exitIOError      = 1
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
	numSeeds:   300,
	dSep:       200.0,
	dTest:      100.0,
	dLookahead: 300.0,
	rkStep:     10.0,
	maxLength:  2000.0,
}

var majorRoadCfg = config{
	numSeeds:   300,
	dSep:       100.0,
	dTest:      30.0,
	dLookahead: 200.0,
	rkStep:     10.0,
	maxLength:  2000.0,
}

var minorRoadCfg = config{
	numSeeds:   300,
	dSep:       20.0,
	dTest:      15.0,
	dLookahead: 40.0,
	rkStep:     1.0,
	maxLength:  2000.0,
}

func buildGraph(width, height int, tf field.TensorField, cfg config, rng *rand.Rand) graph.Graph {
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
	majorLines, minorLines := tracer.Trace(majorGrid, minorGrid, seeds)

	return graph.BuildGraph(width, height, cfg.dSep, majorLines, minorLines)
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

	tf := field.TensorField{
		field.Grid(field.Vector{X: 320.0, Y: 100.0}, field.Vector{X: 1.0, Y: 0.0}, 400.0),
		field.Grid(field.Vector{X: 120.0, Y: 700.0}, field.Vector{X: 0.0, Y: -1.0}, 300.0),
		field.Grid(field.Vector{X: 400.0, Y: 400.0}, field.Vector{X: 1.0, Y: 1.0}, 200.0),
		field.Radial(field.Vector{X: 200.0, Y: 200.0}, 100.0),
		field.Radial(field.Vector{X: 350.0, Y: 500.0}, 75.0),
	}

	/*
		// Grid only
		tf := field.TensorField{
			field.Grid(field.Vector{X: 400, Y: 400}, field.Vector{X: 1, Y: 0}, 600.0),
			field.Grid(field.Vector{X: 300, Y: 300}, field.Vector{X: 1, Y: 0.4}, 150.0),
		}
	*/

	/*
		// Central radius
		tf := field.TensorField{
			field.Radial(field.Vector{X: 400, Y: 400}, 300.0),
			field.Radial(field.Vector{X: 500, Y: 350}, 150.0),
			field.Grid(field.Vector{X: 400, Y: 400}, field.Vector{X: 1, Y: 0}, 500.0),
		}
	*/

	/*
		// Radius and a curve
		tf := field.TensorField{
			field.Grid(field.Vector{X: 400, Y: 600}, field.Vector{X: 1, Y: 0}, 500.0),
			field.Grid(field.Vector{X: 200, Y: 750}, field.Vector{X: 1, Y: 0.6}, 200.0),
			field.Grid(field.Vector{X: 600, Y: 750}, field.Vector{X: 1, Y: -0.6}, 200.0),
			field.Radial(field.Vector{X: 400, Y: 300}, 120.0),
		}
	*/

	rng := rand.New(rand.NewPCG(1234, 1337))
	output := "image.png"
	width := 800
	height := 800

	mainGraph := buildGraph(width, height, tf, mainRoadCfg, rng)
	majorGraph := buildGraph(width, height, tf, majorRoadCfg, rng)
	minorGraph := buildGraph(width, height, tf, minorRoadCfg, rng)

	dc := gg.NewContext(width, height)

	// renderer.DebugGraph(dc, majorLines, minorLines, graph)

	// Draw thicker in black first for the borders.
	renderer.RenderGraph(dc, mainGraph)
	dc.SetHexColor("#000000")
	dc.SetLineWidth(8)
	dc.Stroke()

	renderer.RenderGraph(dc, majorGraph)
	dc.SetHexColor("#000000")
	dc.SetLineWidth(6)
	dc.Stroke()

	renderer.RenderGraph(dc, minorGraph)
	dc.SetHexColor("#000000")
	dc.SetLineWidth(4)
	dc.Stroke()

	// Draw thinner in colors for the fill.
	renderer.RenderGraph(dc, mainGraph)
	dc.SetHexColor("#fff600")
	dc.SetLineWidth(6)
	dc.Stroke()

	renderer.RenderGraph(dc, majorGraph)
	dc.SetHexColor("#708090")
	dc.SetLineWidth(4)
	dc.Stroke()

	renderer.RenderGraph(dc, minorGraph)
	dc.SetHexColor("#dcdcdc")
	dc.SetLineWidth(2)
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
