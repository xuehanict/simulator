package dataproc

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/simulator/utils"
	"testing"
)

func TestParseLightningGraph(t *testing.T) {
	graph, err := ParseLightningGraph("../data/lightning/testnetgraph.json")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(len(graph.Nodes))
	spew.Dump(len(graph.Channels))

	CutOneDegree(2, graph)
	//graph.ConvertToSeriesID()

	spew.Dump(len(graph.Nodes))
	CutOneDegree(2, graph)
	spew.Dump(len(graph.Nodes))
	CutOneDegree(2, graph)
	spew.Dump(len(graph.Nodes))
	CutOneDegree(2,graph)
	spew.Dump(len(graph.Nodes))

	//spew.Dump(len(graph.Channels))
	//spew.Dump(graph.Channels)

}

func TestConnection(t *testing.T)  {
	g, err := ParseLightningGraph("../data/lightning/mainnetgraph.json")
	if err != nil {
		t.Fatal(err)
	}

	part1 := make(map[utils.RouterID]struct{})
	partOther := make(map[utils.RouterID]struct{})
	for id := range g.Nodes {
		path  := utils.BfsPath(g.Nodes, 1, id,false, g.Channels)
		if path == nil || len(path) == 0 {
			partOther[id] = struct{}{}
		} else {
			part1[id] = struct{}{}
		}
	}

	spew.Dump(len(g.Nodes))
	spew.Dump(len(part1))
	spew.Dump(len(partOther))
	RemoveNotConnectNodes(g, partOther)
	spew.Dump(len(g.Nodes))
	for ; CutOneDegree(2,g) != 0; {
		t.Logf("one round \n")
	}
	spew.Dump(len(g.Nodes))

}




func TestGetTrans(t *testing.T) {

	_, err := GetLightningTrans(100, 100,
		"../data/ripple/ripple_val.csv", "../data/lightning/BitcoinVal.txt")
	if err != nil {
		panic(err)
	}
	//spew.Dump(trans)
}
