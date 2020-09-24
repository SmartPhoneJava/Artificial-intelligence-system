package eatree

import (
	"os"
	"shiki/internal/data/anime/tree"
	"shiki/internal/graphml"

	"github.com/beevik/etree"
)

type EDoc struct {
	tree   *etree.Document
	Leaves []AnimeLeaf
	Tree   tree.Tree
}

type AnimeLeaf struct {
	NodeID     string
	Name       *etree.Element
	Desription *etree.Element
}

func NewEdoc(path string) (EDoc, error) {
	var edoc EDoc
	edoc.tree = etree.NewDocument()
	if err := edoc.tree.ReadFromFile(path); err != nil {
		return edoc, err
	}

	var graphml = new(graphml.Graphml)
	err := graphml.Load(path)
	if err != nil {
		return edoc, err
	}

	edoc.Tree = tree.NewTree()
	edoc.Tree.FromGraphml(*graphml, &tree.TreeSettings{
		LeavesKnown: true,
	})

	edoc.Leaves = make([]AnimeLeaf, 0)
	for _, v := range edoc.tree.ChildElements() {
		for _, v1 := range v.ChildElements() {
			if v1.Tag == "graph" {
				for _, v2 := range v1.ChildElements() {
					if v2.Tag == "node" {
						var (
							id = v2.Attr[0].Value
							d5 *etree.Element
							d6 *etree.Element
						)
						for _, v3 := range v2.ChildElements() {
							if v3.Attr[0].Value == "d6" {
								for _, v4 := range v3.ChildElements() {
									for _, v5 := range v4.ChildElements() {
										if v5.Tag == "NodeLabel" {
											d6 = v5
										}
									}
								}
							} else if v3.Attr[0].Value == "d5" {
								d5 = v3

							}

							if d5 != nil && d6 != nil {
								edoc.Leaves = append(edoc.Leaves, AnimeLeaf{
									Name:       d6,
									Desription: d5,
									NodeID:     id,
								})
							}
						}

					}
				}
			}
		}
	}

	return edoc, nil
}

func (ed EDoc) Save(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	_, err = ed.tree.WriteTo(f)

	if err != nil {
		return err
	}
	return f.Close()
}
