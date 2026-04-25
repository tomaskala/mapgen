package main

import (
	"image/color"
	"math/rand/v2"

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
	renderPolygons(dc, c.Polygons)

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

// TODO: This draws each edge in both directions.
func renderGraph(dc *gg.Context, g *graph.Graph) {
	for _, v := range g.Vertices {
		for _, edge := range g.Adjacency[v.ID] {
			isFirst := true
			for p := range edge.Path() {
				if isFirst {
					dc.MoveTo(p.X, p.Y)
					isFirst = false
				} else {
					dc.LineTo(p.X, p.Y)
				}
			}
		}
	}
}

func renderPolygons(dc *gg.Context, polygons []graph.Polygon) {
	for _, polygon := range polygons {
		isFirstPolygonPoint := true
		for _, edge := range polygon.Edges {
			for p := range edge.Path() {
				if isFirstPolygonPoint {
					dc.MoveTo(p.X, p.Y)
					isFirstPolygonPoint = false
				} else {
					dc.LineTo(p.X, p.Y)
				}
			}
		}

		dc.SetColor(
			color.RGBA{uint8(rand.IntN(256)), uint8(rand.IntN(256)), uint8(rand.IntN(256)), uint8(rand.IntN(256))},
		)

		dc.Fill()
	}
}
