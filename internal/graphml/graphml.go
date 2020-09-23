package graphml

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type Graphml struct {
	XMLName xml.Name `xml:"graphml"`
	Keys    []Key    `xml:"key"`
	Graph   Graph    `xml:"graph"`
}

type Key struct {
	AName string `xml:"attr.name,attr"`
	AType string `xml:"attr.type,attr"`

	For string `xml:"for,attr"`

	ID string `xml:"id,attr"`

	YType string `xml:"yfiles.type,attr"`

	Default string `xml:"default"`
}

type Graph struct {
	Edgedefault string `xml:"edgedefault,attr"`

	ID string `xml:"id,attr"`

	Node []Node `xml:"node"`
	Edge []Edge `xml:"edge"`
}

type GData struct {
	Key   string `xml:"key,attr"`
	Space string `xml:"http://www.yworks.com/xml/graphml space"`
}

type Node struct {
	Data []Data `xml:"data"`
	ID   string `xml:"id,attr"`
}

type Edge struct {
	ID     string `xml:"id,attr"`
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
}

type Data struct {
	Key       string    `xml:"key,attr"`
	ShapeNode ShapeNode `xml:"http://www.yworks.com/xml/graphml ShapeNode"`
	Value     string    `xml:",chardata"`
}

type ShapeNode struct {
	NodeLabel string `xml:"http://www.yworks.com/xml/graphml NodeLabel"`
}

func (gr *Graphml) Load(path string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(byteValue, &gr)

	return err
}
