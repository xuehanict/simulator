package mara

import "github.com/lightningnetwork/simulator/utils"

const INF = 0x3f3f3f3f

func dijkstra(nodes map[utils.RouterID]*Node, start utils.RouterID) *DAG {

	mNodes := copyNodes(nodes)

	// 初始化距离和已求出最短距离的集合flag,distance表示start节点到其他节点的距离，
	// flag表示已经求出这个节点的最短距离
	distance := make(map[utils.RouterID]int)
	flag := make(map[utils.RouterID]bool)
	for id := range mNodes {
		distance[id] = INF
		flag[id] = false
	}

	// 针对start节点初始化
	distance[start] = 0
	flag[start]  = true

	for i:=1; i < len(mNodes); i++ {
		min := INF
		k := utils.RouterID(0)
		for node := range mNodes{
			if !flag[node] && distance[node] < min {
				min = distance[node]
				k = node
			}
		}

		flag[k] = true
		for node := range mNodes {
			tmp := 0
			if mNodes[k].checkLink(node) {
				tmp = min + 1
			} else {
				tmp = INF
			}

			if !flag[node] && tmp < distance[node] {
				distance[node] = tmp
				mNodes[node].Parents = append(mNodes[node].Parents, mNodes[k])
				mNodes[k].Children = append(mNodes[k].Children, mNodes[node])
			}
		}
	}

	return &DAG{
		Root: mNodes[start],
		vertexs: mNodes,
	}
}










