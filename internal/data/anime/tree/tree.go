package tree

import (
	"shiki/internal/graphml"
)

type TreeSettings struct {
	LeavesKnown bool
}

type Tree struct {
	NodesNames map[string]string

	NodesUp   map[string][]string
	NodesDown map[string][]string

	// for loading anime only
	Categories map[string]string

	nodesDepth map[string]int
}

func NewTree() Tree {
	return Tree{
		NodesNames: make(map[string]string),

		NodesUp:   make(map[string][]string),
		NodesDown: make(map[string][]string),

		Categories: make(map[string]string),

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
				str := tree.getType(id)
				if str != "" {
					tree.Categories[tree.NodesNames[id]] = str
				}
			}
		}
	}
}

func (tree *Tree) Branch(key string) []string {
	var arr = make([]string, 0)
	for {
		v := tree.NodesUp[key]
		if len(v) == 0 {
			break
		}
		key = v[0]
		arr = append([]string{tree.NodesNames[key]}, arr...)
	}

	return arr
}

func (tree *Tree) Depth(key string) int {
	var size = 1
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

func (tree *Tree) getType(id string) string {
	var strType = ""
	for {
		arr := tree.NodesUp[id]
		if len(arr) == 0 || tree.NodesNames[arr[0]] == "Аниме" {
			break
		}
		id = arr[0]
		strType = tree.NodesNames[id]
	}

	return strType
}
