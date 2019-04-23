package mara

import (
	"github.com/lightningnetwork/simulator/utils"
)

type Link struct {
	From utils.RouterID
	To   utils.RouterID
}

type Node struct {
	ID         utils.RouterID
	Parents    []*Node
	Children   []*Node
	Neighbours []*Node
}

type DAG struct {
	Root    *Node
	vertexs map[utils.RouterID]*Node
	edges   []*Link
}

func NewDAG(root *Node) *DAG {
	return &DAG{
		Root:    root,
		vertexs: make(map[utils.RouterID]*Node),
		edges:   make([]*Link, 0),
	}
}

func (n *Node)checkLink(id utils.RouterID) bool {
	for _, nei := range n.Neighbours {
		if nei.ID == id {
			return true
		}
	}
	return false
}

func copyNodes(src map[utils.RouterID]*Node) map[utils.RouterID]*Node {
	res := make(map[utils.RouterID]*Node)
	for id, node := range src {
		res[id] = &Node{
			ID:id,
		}
		copy(node.Children,res[id].Children)
		copy(node.Parents,res[id].Parents)
		copy(node.Neighbours,res[id].Neighbours)
	}
	return res
}

