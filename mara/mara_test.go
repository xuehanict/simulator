package mara

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/simulator/utils"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	tenNodesGraph = "../data/ten_nodes.json"
	tenNodesGraphComplex = "../data/ten_nodes_complex.json"
	tenNodesGraphHalf = "../data/ten_nodes_half.json"
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
	Capacity1 float64   `json:"capacity1"`
	Capacity2 float64   `json:"capacity2"`
}

func TestNewDAG(t *testing.T)  {
	graph, err := parseTestJson(tenNodesGraphComplex)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}
	startID := utils.RouterID(3)
	m.DAGs[startID] = m.MaraSPE(startID)

	spew.Dump(m.DAGs[startID])
	t.Log("done")
}

func TestNewDAGMcOPT(t *testing.T)  {
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

func TestNewDAGSpeOPT(t *testing.T)  {
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


func TestGetRoutes(t *testing.T)  {
	graph, err := parseTestJson(tenNodesGraphComplex)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}
	startID := utils.RouterID(3)
	m.DAGs[startID] = m.MaraSPE(startID)

	paths := m.getRoutes(9,3,  10)
	spew.Dump(paths)

	t.Log("done")
}

func TestPayment(t *testing.T)  {
	graph, err := parseTestJson(tenNodesGraphComplex)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}

	src := utils.RouterID(3)
	dest := utils.RouterID(8)

	paths := m.getRoutes(src,dest,  150)
	spew.Dump(paths)
	spew.Dump(m.SPTs[dest])
	result := m.sendPayment(src,dest,150)
	spew.Dump(result)
}

func TestGetRoutesSpec(t *testing.T)  {
	graph, err := parseTestJson(tenNodesGraphHalf)
	if err != nil {
		t.Fatalf("%v", err)
	}
	m := &Mara{
		Graph: graph,
	}

	src := utils.RouterID(3)
	dest := utils.RouterID(8)

	paths := m.getRoutes(src,dest,  150)
	spew.Dump(paths)
	//spew.Dump(m.SPTs[dest])
	result := m.sendPayment(src,dest,150)
	spew.Dump(result)
}

func TestRipple(t *testing.T)  {
	f, err := os.Open("../data/finalSets/static/ripple-lcc.graph_CREDIT_LINKS")
	if err != nil {
		fmt.Println("os Open error: ", err)
		return
	}
	defer f.Close()


	br := bufio.NewReader(f)
	lineNum := 1
	links := make(map[string]*utils.Link,0)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return
		}
		//
		if lineNum < 5 {
			lineNum ++
			continue
		}

		splitted := strings.Split(string(line), " ")
		id1, _ := strconv.Atoi(splitted[0])
		id2, _ := strconv.Atoi(splitted[1])
		v1, _ := strconv.ParseFloat(splitted[2],64)
		v2,_ := strconv.ParseFloat(splitted[3], 64)
		v3,_ := strconv.ParseFloat(splitted[4], 64)
		link := &utils.Link{
			Part1: utils.RouterID(id1),
			Part2: utils.RouterID(id2),
			Val1: utils.Amount(v3 - v2),
			Val2: utils.Amount(v2 - v1),
		}
		links[utils.GetLinkKey(link.Part1,link.Part2)] = link
	}

	nodes := make(map[utils.RouterID]*Node, 0)
	for i:=0; i<67149; i++ {
		router := &Node{
			ID: utils.RouterID(i),
			Parents: make([]utils.RouterID,0),
			Children:make([]utils.RouterID,0),
			Neighbours:make(map[utils.RouterID]struct{},0),
		}
		nodes[utils.RouterID(i)] = router
	}

	for _, edge := range links {
		nodes[edge.Part1].Neighbours[edge.Part2] = struct{}{}
		nodes[edge.Part2].Neighbours[edge.Part1] = struct{}{}
	}

	m := &Mara{
		Graph: &Graph{
			Nodes: nodes,
			Channels: links,
			DAGs: make(map[utils.RouterID]*DAG),
			SPTs: make(map[utils.RouterID]*DAG),
		},
	}


	fmt.Printf("节点link数据解析完成\n")

	trans := generateTrans("../data/finalSets/static/sampleTr-1.txt")
	total := 0
	success := 0

	fmt.Printf("交易数据解析完成\n")

	for _, tran := range trans{
		total ++
		err := m.sendPayment(utils.RouterID(tran.src),
			utils.RouterID(tran.dest), utils.Amount(tran.val))
		if err == nil {
			success++
		}

		fmt.Printf("err :%v\n", err)
		fmt.Printf("total:%v\n", total)
		fmt.Printf("success:%v\n", success)
		if total == 50000 {
			break
		}
	}
	fmt.Printf("total :%v\n", total)

	time.Sleep(3 * time.Second)
}


func parseTestJson(filePath string) (*Graph, error){

	nodes := make(map[utils.RouterID]*Node)
	edges := make(map[string]*utils.Link)

	var g testGraph
	graphJson, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(graphJson, &g); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	for _, n := range g.Nodes {
		nodes[n.Id] = &Node{
			ID: n.Id,
			Neighbours: make(map[utils.RouterID]struct{}),
		}
	}
	for _, edge := range g.Edges {
		link := &utils.Link{
			Part1: utils.RouterID(edge.Node1),
			Part2: utils.RouterID(edge.Node2),
			Val1: utils.Amount(edge.Capacity1),
			Val2: utils.Amount(edge.Capacity2),
		}
		linkKey := utils.GetLinkKey(edge.Node1,edge.Node2)
		edges[linkKey] = link
		nodes[link.Part1].Neighbours[link.Part2] = struct{}{}
		nodes[link.Part2].Neighbours[link.Part1] = struct{}{}
	}

	graph := &Graph{
		Channels:edges,
		Nodes:nodes,
		DAGs: make(map[utils.RouterID]*DAG),
		SPTs: make(map[utils.RouterID]*DAG),
	}
	return graph, nil
}

type tran struct {
	src int
	dest int
	val float64
}

func generateTrans (filePath string) []tran {

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	trans := make([]tran,0)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return nil
		}
		splitStr := strings.Split(string(line), " ")
		val, _ := strconv.ParseFloat(splitStr[0], 64)
		src, _ := strconv.Atoi(splitStr[1])
		dest, _:= strconv.Atoi(splitStr[2])

		trans = append(trans,tran{
			src: src,
			dest: dest,
			val: val,
		})
	}

	return trans
}

