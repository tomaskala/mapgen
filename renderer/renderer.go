package renderer

import (
	"github.com/fogleman/gg"
	"tomaskala.com/mapgen/graph"
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
