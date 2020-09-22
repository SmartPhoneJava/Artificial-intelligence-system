package tree

import (
	"fmt"
	"shiki/internal/data/anime"
	"shiki/internal/graphml"
)

type Tree struct {
	values     map[string]string
	animes     anime.Animes
	nodes      map[string]string
	nodesDepth map[string]int

	categories map[string]string
}

func NewTree() Tree {
	return Tree{
		values:     make(map[string]string),
		nodes:      make(map[string]string),
		animes:     []anime.Anime{},
		nodesDepth: make(map[string]int),
		categories: make(map[string]string)}
}

func (tree *Tree) FromGraphml(gr graphml.Graphml) {
	for _, e := range gr.Graph.Edge {
		tree.nodes[e.Target] = e.Source
	}
	for _, v := range gr.Graph.Node {
		var id = v.ID
		for _, d := range v.Data {
			if d.ShapeNode.NodeLabel != "" {
				tree.values[id] = d.ShapeNode.NodeLabel
			}
		}
	}
	for id := range tree.nodes {
		tree.nodesDepth[id] = tree.Depth(id)
	}

	var lastM = tree.nodesDepth
	for _, v := range gr.Graph.Edge {
		lastM[v.Source] = -1
		fmt.Println("gr.Graph.Edge source", tree.values[v.Source])
	}

	for id, v := range lastM {
		fmt.Println("gr.Graph.Edge v", v, tree.values[id])
		if v > 0 {
			tree.categories[tree.values[id]] = tree.getParent(id, 2)
		}
	}
}

func (tree *Tree) Depth(key string) int {
	var (
		ok   = true
		size = 0
	)
	for ok {
		key, ok = tree.nodes[key]
		size++
	}
	return size
}

func (tree *Tree) getParent(id string, index int) string {
	if index+1 < tree.nodesDepth[id] {
		var repeat = tree.nodesDepth[id] - index
		for i := 0; i < repeat; i++ {
			id = tree.nodes[id]
		}
		return tree.values[id]
	}
	return ""
}
