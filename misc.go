package main

import sw "github.com/lightningnetwork/sm/silentWhisper"

const (
	tenNodesGraph = "data/ten_nodes.json"
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
	Id sw.RouteID `json:"id"`
}

type testEdge struct {
	Node1     sw.RouteID `json:"node_1"`
	Node2     sw.RouteID `json:"node_2"`
	Capacity1 float64   `json:"capacity1"`
	Capacity2 float64   `json:"capacity2"`
}
