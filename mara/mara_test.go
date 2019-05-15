package mara

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/simulator/utils"
	fibHeap "github.com/starwander/GoFibonacciHeap"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

const (
	tenNodesGraph        = "../data/ten_nodes.json"
	tenNodesGraphComplex = "../data/ten_nodes_complex.json"
	tenNodesGraphHalf    = "../data/ten_nodes_half.json"
)

/**
用来解析图的json文件的辅助结构和函数
*/
type testGraph struct {
	Info  []string   `json:"info"`
	Nodes []testNode `json:"nodes"`
	Edges []testEdge `json:"edges"`
}

type testNode struct {
	Id utils.RouterID `json:"id"`
}

type testEdge struct {
	Node1     utils.RouterID `json:"node_1"`
	Node2     utils.RouterID `json:"node_2"`
	Capacity1 float64        `json:"capacity1"`
	Capacity2 float64        `json:"capacity2"`
}

func TestNewDAG(t *testing.T) {
	graph, err := parseTestJson(tenNodesGraphComplex)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}
	startID := utils.RouterID(3)
	m.DAGs[startID] = m.MaraSpeOpt(startID)

	spew.Dump(m.DAGs[startID])
	t.Log("done")
}

func TestNewDAGMcOPT(t *testing.T) {
	graph, err := parseTestJson(tenNodesGraphComplex)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}
	startID := utils.RouterID(3)
	m.DAGs[startID] = m.MaraMcOPT(startID)

	//spew.Dump(m.DAGs[startID])
	t.Log("done")
}

func TestNewDAGSpeOPT(t *testing.T) {
	graph, err := parseTestJson(tenNodesGraphComplex)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}
	startID := utils.RouterID(3)
	m.DAGs[startID] = m.MaraSpeOpt(startID)

	//spew.Dump(m.DAGs[startID])
	t.Log("done")
}

func TestGetRoutes(t *testing.T) {
	graph, err := parseTestJson(tenNodesGraphComplex)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}
	startID := utils.RouterID(3)
	m.DAGs[startID] = m.MaraSpeOpt(startID)

	paths := m.getRoutes(9, 3, 10)
	spew.Dump(paths)

	t.Log("done")
}

func TestPayment(t *testing.T) {
	graph, err := parseTestJson(tenNodesGraphComplex)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}

	src := utils.RouterID(1)
	dest := utils.RouterID(9)

	paths := m.getRoutes(src, dest, 150)
	spew.Dump(paths)
	//	spew.Dump(m.SPTs[dest])
	_, _, result := m.SendPaymentWithBond(src, dest, 150, 6, 0.01)
	spew.Dump(result)
}

func TestGetRoutesSpec(t *testing.T) {
	graph, err := parseTestJson(tenNodesGraphHalf)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}

	src := utils.RouterID(2)
	dest := utils.RouterID(7)

	paths := m.getRoutes(src, dest, 150)
	spew.Dump(paths)
	//spew.Dump(m.SPTs[dest])
	result := m.SendPayment(src, dest, 150)
	spew.Dump(result)
}

func TestRipple(t *testing.T) {
	m, trans := GetRippleMaraAndTrans("../data")
	total := 0
	success := 0

	for _, tran := range trans {
		total++

		len1, len2, err := m.SendPaymentWithBond(utils.RouterID(tran.Src),
			utils.RouterID(tran.Dest), utils.Amount(tran.Val), 6, 0.01)
		if err == nil {
			success++
		}

		fmt.Printf("err :%v", err)
		fmt.Printf("; total:%v", total)
		fmt.Printf("; success:%v", success)
		fmt.Printf("; path number:%v", len1)
		fmt.Printf("; used path number:%v \n", len2)

		if total == 10000 {
			break
		}
	}
	fmt.Printf("total :%v\n", total)
	time.Sleep(3 * time.Second)
}

func parseTestJson(filePath string) (*Graph, error) {

	var g testGraph
	graphJson, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(graphJson, &g); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	nodes := make([]*Node, len(g.Nodes))
	edges := make(map[string]*utils.Link)

	for _, n := range g.Nodes {
		nodes[n.Id] = &Node{
			ID:         n.Id,
			Neighbours: make([]utils.RouterID, 0),
		}
	}
	for _, edge := range g.Edges {
		link := &utils.Link{
			Part1: utils.RouterID(edge.Node1),
			Part2: utils.RouterID(edge.Node2),
			Val1:  utils.Amount(edge.Capacity1),
			Val2:  utils.Amount(edge.Capacity2),
		}
		linkKey := utils.GetLinkKey(edge.Node1, edge.Node2)
		edges[linkKey] = link
		nodes[link.Part1].Neighbours = append(nodes[link.Part1].Neighbours, link.Part2)
		nodes[link.Part2].Neighbours = append(nodes[link.Part2].Neighbours, link.Part1)
	}

	graph := &Graph{
		Channels: edges,
		Nodes:    nodes,
		DAGs:     make(map[utils.RouterID]*DAG),
		SPTs:     make(map[utils.RouterID]*DAG),
	}
	return graph, nil
}

func TestFibHeap(T *testing.T) {

	heap := fibHeap.NewFibHeap()
	for i := 1; i < 100; i++ {
		if i%10 == 0 {
			err := heap.Insert(i, -1)
			if err != nil {
				fmt.Printf("faced error : %v", err)
			}
		} else {
			err := heap.Insert(i, float64(i))
			if err != nil {
				fmt.Printf("faced error : %v", err)
			}
		}
	}

	tag, _ := heap.ExtractMin()
	fmt.Printf("tag: %v", tag)

}
