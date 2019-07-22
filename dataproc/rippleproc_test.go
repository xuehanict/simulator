package dataproc

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/simulator/utils"
	"sort"
	"testing"
)

func TestGraph_CutOneDegree(t *testing.T) {
	g, _ := utils.ParseTestJson("../data/ten_nodes.json")
	CutOneDegree(2, g)
	if _, ok := g.Nodes[1].Neighbours[2]; !ok {
		t.Fatal("cut failed")
	}
	CutOneDegree(2,g)
	spew.Dump(g.Nodes)
}

func TestGraph_ConvertToSeriesID(t *testing.T) {
	g, _ := utils.ParseTestJson("../data/ten_nodes.json")
	CutOneDegree(2,g)
	CutOneDegree(2,g)
	spew.Dump(ConvertToSeriesID(ORIGION_CHANNEL, g))
	spew.Dump(g.Nodes)
	spew.Dump(g.Channels)
}

func TestOriginDataSize(t *testing.T) {

	g := utils.GetGraph("../data")
	for CutOneDegree(2, g) != 0 {
		spew.Dump(len(g.Nodes))
	}
}

func TestSnapshotData(t *testing.T) {
	g := utils.GetGraphSnapshot("../data")
	spew.Dump(len(g.Nodes))
	for CutOneDegree(2, g) != 0 {
		spew.Dump(len(g.Nodes))
	}
}

func TestGetTransMaxMin(t *testing.T)  {
	oriTrans, _ := utils.GenerateTransFromPath("../data/finalSets/static/")
	values := make([]float64,0)
	for _, tran := range oriTrans{
		values = append(values, tran.Val)
	}
	sort.Float64s(values)

	t.Logf("min is %v", values[len(values)/10*1])
	t.Logf("max is %v", values[len(values) - len(values)/10])

}
