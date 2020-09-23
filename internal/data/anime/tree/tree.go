package tree

import (
	"log"
	"shiki/internal/data/anime"
	"shiki/internal/graphml"
)

type TreeSettings struct {
	LeavesKnown bool
}

type Tree struct {
	NodesNames  map[string]string
	NodesAnimes map[string]anime.Anime

	NodesUp   map[string][]string
	NodesDown map[string][]string

	// for loading anime only
	Categories map[string]string

	animes     anime.Animes
	nodesDepth map[string]int
}

func NewTree() Tree {
	return Tree{
		NodesNames:  make(map[string]string),
		NodesAnimes: make(map[string]anime.Anime),

		NodesUp:   make(map[string][]string),
		NodesDown: make(map[string][]string),

		Categories: make(map[string]string),

		animes:     []anime.Anime{},
		nodesDepth: make(map[string]int),
	}

}

func (tree *Tree) FromGraphml(gr graphml.Graphml, settings *TreeSettings) {
	if settings == nil {
		settings = &TreeSettings{true}
	}
	for _, e := range gr.Graph.Edge {
		tree.NodesUp[e.Target] = append(tree.NodesUp[e.Target], e.Source)
		tree.NodesDown[e.Source] = append(tree.NodesUp[e.Source], e.Target)
	}
	for _, v := range gr.Graph.Node {
		var id = v.ID
		for _, d := range v.Data {
			if d.Key == "d6" {
				if d.ShapeNode.NodeLabel != "" {
					tree.NodesNames[id] = d.ShapeNode.NodeLabel
				}
			} else if d.Key == "d5" {
				if settings.LeavesKnown {
					log.Println("d.Value", d.Value)
				}
			}
		}
	}
	if !settings.LeavesKnown {
		for id := range tree.NodesUp {
			tree.nodesDepth[id] = tree.Depth(id)
		}

		var lastM = tree.nodesDepth
		for _, v := range gr.Graph.Edge {
			lastM[v.Source] = -1
		}

		for id, v := range lastM {
			if v > 0 {
				arr := tree.getParent(id, 2)
				if len(arr) != 0 {
					tree.Categories[tree.NodesUp[id][0]] = arr[0]
				}
			}
		}
	}
}

func (tree *Tree) Depth(key string) int {
	var size = 0
	for {
		v := tree.NodesUp[key]
		if len(v) == 0 {
			break
		}
		key = v[0]
		size++
	}
	return size
}

func (tree *Tree) getParent(id string, index int) []string {
	if index+1 < tree.nodesDepth[id] {
		var repeat = tree.nodesDepth[id] - index
		var arr []string
		for i := 0; i < repeat; i++ {
			arr = tree.NodesUp[id]
			if len(arr) == 0 {
				break
			}
			id = arr[0]
		}
		return tree.NodesUp[id]
	}
	return []string{}
}
