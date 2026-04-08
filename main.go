package main

import (
	"flag"
	"log"
	"math/rand/v2"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/fogleman/gg"
)

const (
	exitSuccess = 0
	exitIOError = 1
)

var (
	width  = flag.Int("width", 800, "image width in pixels")
	height = flag.Int("height", 800, "image height in pixels")
	output = flag.String("output", "city.png", "name for the generated image")

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

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
	city := buildCityMap(*width, *height, rng)
	dc := gg.NewContext(*width, *height)
	renderCity(dc, city)

	if err := dc.SavePNG(*output); err != nil {
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
