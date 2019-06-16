package utils

import (
	"fmt"
	fibHeap "github.com/starwander/GoFibonacciHeap"
)

const INF = 0x3f3f3f3f

type disElement struct {
	distance float64
	id       RouterID
}

func (d *disElement) Tag() interface{} {
	return d.id
}

func (d *disElement) Key() float64 {
	return d.distance
}

func Dijkstra(nodes []*Node, start RouterID) (*DAG, map[RouterID]float64) {

	mNodes := CopyNodes(nodes)

	// 初始化距离和已求出最短距离的集合flag,distance表示start节点到其他节点的距离，
	// flag表示已经求出这个节点的最短距离
	// 最后空间换时间，distance数组用来直接索引距离，heap用来直接取最小值
	distance := make(map[RouterID]float64)
	heap := fibHeap.NewFibHeap()
	flag := make(map[RouterID]bool)
	for id := range mNodes {
		err := heap.InsertValue(&disElement{INF, RouterID(id)})
		if err != nil {
			fmt.Printf("insert value faced err :%v", err)
		}
		distance[RouterID(id)] = INF
		flag[RouterID(id)] = false
	}

	// 针对start节点初始化
	distance[start] = 0
	flag[start] = true
	for _, n := range nodes[start].Neighbours {
		distance[n] = 1
		err := heap.DecreaseKey(n, 1)
		if err != nil {
			fmt.Printf("decrease value faced err :%v", err)
		}
		mNodes[n].Parents = append(mNodes[n].Parents, start)
	}

	for i := 1; i < len(mNodes); i++ {
		tmpK, min := heap.ExtractMin()
		k := tmpK.(RouterID)
		flag[k] = true
		for _, node := range nodes[k].Neighbours {
			tmp := min + 1
			if !flag[node] {
				if tmp < distance[node] {
					distance[node] = tmp
					err := heap.DecreaseKey(node, tmp)
					if err != nil {
						fmt.Printf("decrease value faced err :%v", err)
					}
					mNodes[node].Parents = nil
					mNodes[node].Parents = append(mNodes[node].Parents, k)
				} else if tmp != INF && tmp == distance[node] {
					mNodes[node].Parents = append(mNodes[node].Parents, k)
				}
			}
		}
	}

	// 不光要让能找到父节点，还要能找到孩子节点
	for id, n := range mNodes {
		for _, p := range n.Parents {
			mNodes[p].Children = append(mNodes[p].Children, RouterID(id))
		}
	}

	return &DAG{
		Root:    mNodes[start],
		Vertexs: mNodes,
	}, distance
}
