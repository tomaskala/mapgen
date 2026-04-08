package city

import "tomaskala.com/mapgen/graph"

type City struct {
	MainRoads  graph.Graph
	MajorRoads graph.Graph
	MinorRoads graph.Graph
}
