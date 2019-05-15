package mara

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
)

/*
type Link struct {
	From utils.RouterID
	To   utils.RouterID
}
*/

/**
 * Node 结构为Graph和DAG中的节点，在无线图中，只会用到Neighbours数组
 * 在DAG中会用到Parents和Children数组以表示方向
 */
type Node struct {
	ID         utils.RouterID
	Parents    []utils.RouterID
	Children   []utils.RouterID
	Neighbours []utils.RouterID
}

type DAG struct {
	Root    *Node
	vertexs []*Node
	//edges   []*Link
}

type Graph struct {
	Nodes []*Node
	// 图的channels信息
	Channels map[string]*utils.Link
	// 路由所需要的DAG，路由时是一步步往parents的方向传递路由包和支付请求
	DAGs map[utils.RouterID]*DAG
	// 最短路树，用来构建MC-SPE
	SPTs map[utils.RouterID]*DAG
}

func NewDAG(root *Node, len int) *DAG {
	return &DAG{
		Root:    root,
		vertexs: make([]*Node, len),
		//edges:   make([]*Link, 0),
	}
}

func (n *Node) checkLink(id utils.RouterID) bool {
	//fmt.Printf("node id is %v", spew.Sdump(n))
	for _, n := range n.Neighbours {
		if n == id {
			return true
		}
	}
	return false
}

func (n *Node) checkParent(id utils.RouterID) bool {
	for _, p := range n.Parents {
		if p == id {
			return true
		}
	}
	return false
}

func (n *Node) checkChild(id utils.RouterID) bool {
	for _, c := range n.Children {
		if c == id {
			return true
		}
	}
	return false
}

func copyNodes(src []*Node) []*Node {
	res := make([]*Node, len(src))
	for id, node := range src {
		n := &Node{
			ID:         utils.RouterID(id),
			Neighbours: node.Neighbours,
			Children:   node.Children,
			Parents:    node.Parents,
		}
		res[id] = n
	}
	return res
}

func (g *Graph) updateWeights(routes [][]utils.RouterID,
	amts []utils.Amount) error {

	if len(routes) != len(amts) {
		return fmt.Errorf("routes number is not equal to amts' ")
	}

	for idx, route := range routes {
		for i := 0; i < len(route)-1; i++ {
			// i 到 i+1 的钱减少
			err := utils.UpdateLinkValue(route[i], route[i+1],
				g.Channels, amts[idx], false)
			if err != nil {
				return err
			}
			// i+1 到 i 的钱增加
			err = utils.UpdateLinkValue(route[i+1], route[i],
				g.Channels, amts[idx], true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
