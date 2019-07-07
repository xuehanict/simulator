package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	tenNodesGraph     = "data/ten_nodes.json"
	tenNodesGraphHalf = "data/ten_nodes_half.json"
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

type tran struct {
	src  int
	dest int
	val  float64
}

func GenerateTrans(filePath string) []tran {

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	trans := make([]tran, 0)
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
		dest, _ := strconv.Atoi(splitStr[2])

		trans = append(trans, tran{
			src:  src,
			dest: dest,
			val:  val,
		})
	}

	return trans
}

func ParseTestJson(filePath string) (*utils.Graph, error) {

	var g testGraph
	graphJson, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(graphJson, &g); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	nodes := make([]*utils.Node, len(g.Nodes))
	edges := make(map[string]*utils.Link)

	for _, n := range g.Nodes {
		nodes[n.Id] = &utils.Node{
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

	graph := &utils.Graph{
		Channels: edges,
		Nodes:    nodes,
		DAGs:     make(map[utils.RouterID]*utils.DAG),
		SPTs:     make(map[utils.RouterID]*utils.DAG),
		Distance: make(map[utils.RouterID]map[utils.RouterID]float64),
	}
	return graph, nil
}

