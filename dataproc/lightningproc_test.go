package dataproc

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestParseLightningGraph(t *testing.T) {
	graph, err := ParseLightningGraph("../data/lightning/testnetgraph.json")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(len(graph.Nodes))
	spew.Dump(len(graph.Channels))

	graph.CutOneDegree(2)
	//graph.ConvertToSeriesID()

	spew.Dump(len(graph.Nodes))
	graph.CutOneDegree(2)
	spew.Dump(len(graph.Nodes))
	graph.CutOneDegree(2)
	spew.Dump(len(graph.Nodes))
	graph.CutOneDegree(2)
	spew.Dump(len(graph.Nodes))

	//spew.Dump(len(graph.Channels))
	//spew.Dump(graph.Channels)

}

func TestGetTrans(t *testing.T) {

	_, err := GetLightningTrans(100, 100,
		"../data/ripple/ripple_val.csv", "../data/lightning/BitcoinVal.txt")
	if err != nil {
		panic(err)
	}
	//spew.Dump(trans)
}
