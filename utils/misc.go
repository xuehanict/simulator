package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
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
	Id RouterID `json:"id"`
}

type testEdge struct {
	Node1     RouterID `json:"node_1"`
	Node2     RouterID `json:"node_2"`
	Capacity1 float64        `json:"capacity1"`
	Capacity2 float64        `json:"capacity2"`
}

func ParseTestJson(filePath string) (*Graph, error) {

	var g testGraph
	graphJson, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(graphJson, &g); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	nodes := make(map[RouterID]*Node)
	edges := make(map[string]*Link)

	for _, n := range g.Nodes {
		nodes[n.Id] = &Node{
			ID:         n.Id,
			Neighbours: make(map[RouterID]struct{}),
		}
	}
	for _, edge := range g.Edges {
		link := &Link{
			Part1: RouterID(edge.Node1),
			Part2: RouterID(edge.Node2),
			Val1:  Amount(edge.Capacity1),
			Val2:  Amount(edge.Capacity2),
		}
		linkKey := GetLinkKey(edge.Node1, edge.Node2)
		edges[linkKey] = link
		nodes[link.Part1].Neighbours[link.Part2] = struct{}{}
		nodes[link.Part2].Neighbours[link.Part1] = struct{}{}
	}

	graph := &Graph{
		Channels: edges,
		Nodes:    nodes,
		DAGs:     make(map[RouterID]*DAG),
		SPTs:     make(map[RouterID]*DAG),
		Distance: make(map[RouterID]map[RouterID]float64),
	}
	RanddomFeeRate(graph.Channels)
	return graph, nil
}

func GetMap(filePath string) map[RouterID]RouterID {
	resMap := make(map[RouterID]RouterID)
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return nil
		}
		splitStr := strings.Split(string(line), " ")
		src, _ := strconv.Atoi(splitStr[0])
		dest, _ := strconv.Atoi(splitStr[1])
		resMap[RouterID(src)]=RouterID(dest)
	}
	return resMap
}

