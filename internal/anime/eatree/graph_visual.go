package eatree

import (
	"io"
	"log"

	"github.com/go-echarts/go-echarts/charts"
)

type ChartsGraph struct {
	graph *charts.Graph
}

func (c ChartsGraph) chooseColor(v int) string {
	switch v {
	case 1:
		return "brown"
	case 2:
		return "orange"
	case 3:
		return "blue"
	case 4:
		return "green"
	case 5:
		return "red"
	}
	return "black"
}

func (c ChartsGraph) chooseForm(v int) string {
	// switch v {
	// case 1:
	// 	return "pin"
	// case 2:
	// 	return "rect"
	// case 3:
	// 	return "roundRect"
	// case 4:
	// 	return "diamond"
	// case 5:
	// 	return "triangle"
	// }
	return "diamond"
}

func (c ChartsGraph) chooseSize(v int) []int {
	switch v {
	case 1:
		return []int{50, 50}
	case 2:
		return []int{40, 40}
	case 3:
		return []int{30, 30}
	case 4:
		return []int{20, 20}
	case 5:
		return []int{15, 15}
	}
	return []int{10, 10}
}

func NewChartsGraph(path string) (ChartsGraph, error) {
	var eadoc, err = NewEdoc(path)
	if err != nil {
		log.Fatal(err)
	}

	var (
		cg = ChartsGraph{}

		graphNodes = make([]charts.GraphNode, 0)
		graphLinks = make([]charts.GraphLink, 0)

		m = make(map[string]bool)
	)

	for _, v := range eadoc.Leaves {
		ok := true
		var key = v.NodeID

		for ok {
			if !m[key] {
				var depth = eadoc.Tree.Depth(key)
				graphNodes = append(
					graphNodes,
					charts.GraphNode{
						Name:       key,
						Symbol:     cg.chooseForm(depth),
						SymbolSize: cg.chooseSize(depth),
						ItemStyle: charts.ItemStyleOpts{
							Color: cg.chooseColor(depth),
						},
					})
			}
			m[key] = true
			nods := eadoc.Tree.NodesUp[key]
			if len(nods) == 0 {
				break
			}

			graphLinks = append(graphLinks, charts.GraphLink{Source: nods[0], Target: key})
			key = nods[0]
		}
	}

	cg.graph = charts.NewGraph().SetGlobalOptions(
		charts.ColorOpts{"green", "red", "blue"},
		charts.TitleOpts{
			Title: "Вернуться назад",
			Link:  "/",
		},
		charts.LegendOpts{Right: "20%"},
		charts.ToolboxOpts{Show: true},
		charts.InitOpts{
			PageTitle: "Визуализация классификации аниме",
			Width:     "720px", Height: "750px",
			BackgroundColor: "#f5f5dc"},
		// charts.DataZoomOpts{XAxisIndex: []int{0}, Start: 50, End: 100},
	).Add("graph", graphNodes, graphLinks,
		charts.GraphOpts{Roam: true, FocusNodeAdjacency: true, Force: charts.GraphForce{
			Repulsion: 100,
		}},
		charts.EmphasisOpts{Label: charts.LabelTextOpts{Show: true, Position: "left", Color: "black"},
			ItemStyle: charts.ItemStyleOpts{Color: "yellow"}},
		charts.LineStyleOpts{Curveness: 0.2})
	return cg, nil
}

func (c ChartsGraph) Render(w ...io.Writer) error {
	return c.graph.Render(w...)
}
