package dataproc

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/simulator/utils"
	"math/rand"
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
	spew.Dump(ConvertToSeriesID(ORIGION_CHANNEL, g, 0))
	spew.Dump(g.Nodes)
	spew.Dump(g.Channels)
}

func TestOriginDataSize(t *testing.T) {

	g := utils.GetGraph("../data")
	for {
		delLen := CutOneDegree(2, g)
		spew.Dump(delLen)
		if delLen == 0 {
			break
		}
	}
}

func TestSnapshotData(t *testing.T) {
	g := utils.GetGraphSnapshot("../data", false)
	fmt.Printf("node number is %v\n", len(g.Nodes))
	CutOneDegree(2, g)
	RemoveZeroEdge(g)
	fmt.Print("remove zero value done\n")
	maxComponent := g.GetMaxComponent()
	spew.Dump(len(maxComponent))
	RemainNodes(maxComponent, g)
	spew.Dump(len(g.Channels))
	//t.Logf("结点个数为%d", len(g.Nodes))
}

func TestJSON(t *testing.T) {
	g, _ := utils.ParseTestJson("../data/ten_nodes.json")
	fmt.Printf("node number is %v\n", len(g.Nodes))
	CutOneDegree(2, g)
	RemoveZeroEdge(g)
	fmt.Print("remove zero value done\n")
	spew.Dump(g.Channels)
	spew.Dump(g.GetMaxComponent())

	//t.Logf("结点个数为%d", len(g.Nodes))
}

func TestExpotionDistribution(t *testing.T)  {
	r := rand.New(rand.NewSource(99))
	i := 0
	for {
		spew.Dump(r.ExpFloat64())
		i++
		if i == 100 {
			break
		}
	}

}

func TestSnapShotConnection(t *testing.T)  {
	g := utils.GetGraphSnapshot("../data", true)

	for CutOneDegree(2, g) != 0 {
		spew.Dump(len(g.Nodes))
	}
	//ConvertToSeriesID(ORIGION_CHANNEL, g)
	part1 := make(map[utils.RouterID]struct{})
	partOther := make(map[utils.RouterID]struct{})
	for id := range g.Nodes {
		//t.Logf("id is %v", id)
		path  := utils.BfsPath(g.Nodes, 4, id,false, g.Channels)
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

func TestAnalyze(t *testing.T) {
	trans, _ := utils.GenerateTransFromPath("../data/finalSets/static/")
	destNum := make(map[utils.RouterID]int)
	for _, tran := range trans {
		if _, ok := destNum[utils.RouterID(tran.Dest)]; ok {
			destNum[utils.RouterID(tran.Dest)] = destNum[utils.RouterID(tran.Dest)] + 1
		} else {
			destNum[utils.RouterID(tran.Dest)] = 1
		}
	}
	fmt.Printf("平均每个dest有：%v比交易\n", len(trans)/len(destNum))

	distri := make(map[int]int)
	for _, num := range destNum {
		if _, ok := distri[num]; ok {
			distri[num] = distri[num] + 1
		} else {
			distri[num] = 1
		}
	}
//	spew.Dump(distri)


	less10 := 0
	for bishu, destshu := range distri {
		if bishu < 10000{
			less10 = less10 + bishu * destshu
		}
	}
	spew.Dump(less10)
}