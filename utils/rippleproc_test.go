package utils

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestGraph_CutOneDegree(t *testing.T) {
	g, _ := ParseTestJson("../data/ten_nodes.json")
	g.CutOneDegree(2)
	if len(g.Nodes) != 9 || g.Nodes[1].Neighbours[0] != 2 {
		t.Fatal("cut failed")
	}
	g.CutOneDegree(2)
	spew.Dump(g.Nodes)
}

func TestGraph_ConvertToSeriesID(t *testing.T) {
	g, _ := ParseTestJson("../data/ten_nodes.json")
	g.CutOneDegree(2)
	g.CutOneDegree(2)
	spew.Dump(g.ConvertToSeriesID())
	spew.Dump(g.Nodes)
	spew.Dump(g.Channels)
}

func TestOriginDataSize(t *testing.T) {

	g := GetGraph("../data")
	for g.CutOneDegree(2) != 0 {
		spew.Dump(len(g.Nodes))
	}
}
