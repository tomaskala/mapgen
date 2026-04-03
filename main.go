package main

import (
	"math/rand/v2"

	"github.com/fogleman/gg"
	"tomaskala.com/mapgen/field"
	"tomaskala.com/mapgen/graph"
	"tomaskala.com/mapgen/streamline"
)

func main() {
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
	numSeeds := 25
	output := "image.png"
	width := 800
	height := 800

	dSep := 20.0
	dTest := 0.5 * dSep // Must be (0.0 * dSep, 1.0 * dSep)
	dLookahead := 2.0 * dSep
	rkStep := 0.5 * dSep
	maxLength := 2000.0

	majorGrid := streamline.NewGrid(width, height, dSep)
	minorGrid := streamline.NewGrid(width, height, dSep)
	seeds := make([]field.Vector, numSeeds)
	for i := range numSeeds {
		seeds[i] = field.Vector{
			X: rng.Float64() * float64(width),
			Y: rng.Float64() * float64(height),
		}
	}

	tracer := streamline.NewTracer(tf, dSep, dTest, dLookahead, rkStep, maxLength)
	majorLines, minorLines := tracer.Trace(majorGrid, minorGrid, seeds)

	graph := graph.BuildGraph(width, height, dSep, majorLines, minorLines)

	dc := gg.NewContext(width, height)

	dc.SetHexColor("#FF0000")
	for _, major := range majorLines {
		points := major.Points()
		if len(points) == 0 {
			continue
		}

		dc.MoveTo(points[0].X, points[0].Y)
		for _, p := range points[1:] {
			dc.LineTo(p.X, p.Y)
		}
	}
	dc.SetLineWidth(8)
	dc.Stroke()

	dc.SetHexColor("#00FF00")
	for _, minor := range minorLines {
		points := minor.Points()
		if len(points) == 0 {
			continue
		}

		dc.MoveTo(points[0].X, points[0].Y)
		for _, p := range points[1:] {
			dc.LineTo(p.X, p.Y)
		}
	}
	dc.SetLineWidth(4)
	dc.Stroke()

	dc.SetHexColor("#0000FF")
	for _, vertex := range graph.Vertices {
		dc.DrawPoint(vertex.Pos.X, vertex.Pos.Y, 10.0)
	}
	dc.Fill()

	dc.SetHexColor("#FFA500")
	for _, edge := range graph.Edges {
		if len(edge.Path) == 0 {
			continue
		}

		dc.MoveTo(edge.Path[0].X, edge.Path[0].Y)
		for _, p := range edge.Path[1:] {
			dc.LineTo(p.X, p.Y)
		}
	}
	dc.SetLineWidth(2)
	dc.Stroke()

	dc.SavePNG(output)
}
