package main

import (
	"github.com/fogleman/gg"
	"tomaskala.com/mapgen/city"
	"tomaskala.com/mapgen/graph"
)

const (
	roadBorderColor = "#000000"
	mainRoadColor   = "#fff600"
	majorRoadColor  = "#708090"
	minorRoadColor  = "#dcdcdc"

	roadBorderWidth = 1.0
	mainRoadWidth   = 6.0
	majorRoadWidth  = 4.0
	minorRoadWidth  = 2.0
)

func renderCity(dc *gg.Context, c city.City) {
	// Draw thicker first for the borders.
	renderGraph(dc, c.MinorRoads)
	dc.SetHexColor(roadBorderColor)
	dc.SetLineWidth(2*roadBorderWidth + minorRoadWidth)
	dc.Stroke()

	renderGraph(dc, c.MajorRoads)
	dc.SetHexColor(roadBorderColor)
	dc.SetLineWidth(2*roadBorderWidth + majorRoadWidth)
	dc.Stroke()

	renderGraph(dc, c.MainRoads)
	dc.SetHexColor(roadBorderColor)
	dc.SetLineWidth(2*roadBorderWidth + mainRoadWidth)
	dc.Stroke()

	// Draw thinner next for the fill.
	renderGraph(dc, c.MinorRoads)
	dc.SetHexColor(minorRoadColor)
	dc.SetLineWidth(minorRoadWidth)
	dc.Stroke()

	renderGraph(dc, c.MajorRoads)
	dc.SetHexColor(majorRoadColor)
	dc.SetLineWidth(majorRoadWidth)
	dc.Stroke()

	renderGraph(dc, c.MainRoads)
	dc.SetHexColor(mainRoadColor)
	dc.SetLineWidth(mainRoadWidth)
	dc.Stroke()
}

func renderGraph(dc *gg.Context, g graph.Graph) {
	for _, edge := range g.Edges {
		if len(edge.Path) == 0 {
			continue
		}

		dc.MoveTo(edge.Path[0].X, edge.Path[0].Y)
		for _, p := range edge.Path[1:] {
			dc.LineTo(p.X, p.Y)
		}
	}
}
