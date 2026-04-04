package renderer

import (
	"github.com/fogleman/gg"
	"tomaskala.com/mapgen/graph"
	"tomaskala.com/mapgen/streamline"
)

func RenderGraph(dc *gg.Context, g graph.Graph) {
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

func DebugGraph(dc *gg.Context, majorLines, minorLines []streamline.Streamline, g graph.Graph) {
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
	for _, vertex := range g.Vertices {
		dc.DrawPoint(vertex.Pos.X, vertex.Pos.Y, 10.0)
	}
	dc.Fill()
}
