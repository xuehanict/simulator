package mara

import "github.com/lightningnetwork/simulator/utils"

func MaraMC(nodes map[utils.RouterID]*Node, start *Node) *DAG {

	S := make(map[utils.RouterID]struct{})
	T := make(map[utils.RouterID]struct{})

	for _, node := range nodes {
		T[node.ID] = struct{}{}
	}
	startID := start.ID
	delete(T, startID)

	S[startID] = struct{}{}

	ordering := make([]utils.RouterID, 1)
	ordering[0] = startID

	for {
		if len(ordering) == len(nodes) {
			break
		}
		v := utils.RouterID(-1)
		largestD := 0
		for vtx := range T {
			tmpConn := 0
			for _, n := range nodes[vtx].Neighbours {
				if _, ok := S[n.ID]; ok {
					tmpConn++
				}
			}
			if tmpConn > largestD {
				largestD = tmpConn
				v = vtx
			}
		}
		ordering = append(ordering, v)
		S[v] = struct{}{}
		delete(T, v)
	}
	return getDAG(ordering,nodes)
}

func MaraSPE(nodes map[utils.RouterID]*Node, start *Node, spf *DAG) *DAG {
	S := make(map[utils.RouterID]struct{})

	startID := start.ID
	S[startID] = struct{}{}

	ordering := make([]utils.RouterID, 1)
	ordering[0] = startID

	for {
		if len(ordering) == len(nodes) {
			break
		}
		v := utils.RouterID(-1)
		T := computeT(spf,S)
		largestD := 0
		for vtx := range T {
			tmpConn := 0
			for _, n := range nodes[vtx].Neighbours {
				if _, ok := S[n.ID]; ok {
					tmpConn++
				}
			}
			if tmpConn > largestD {
				largestD = tmpConn
				v = vtx
			}
		}
		ordering = append(ordering, v)
		S[v] = struct{}{}
	}
	return getDAG(ordering,nodes)
}

func getDAG(ordering []utils.RouterID, nodes map[utils.RouterID]*Node) *DAG {

	mapOrdering := make(map[utils.RouterID]int, len(ordering))
	for index, id := range ordering {
		mapOrdering[id] = index
	}

	tmpLinks := make([]*Link,0)
	dag := NewDAG(nodes[ordering[0]])
	dag.edges = tmpLinks
	tmpNodes := copyNodes(nodes)
	dag.vertexs = tmpNodes

	for i := 0; i < len(ordering); i++ {
		for _, n := range nodes[ordering[i]].Neighbours {
			if mapOrdering[n.ID] > i {
				link := &Link{
					From: n.ID,
					To:   ordering[i],
				}
				tmpNodes[ordering[i]].Parents = append(tmpNodes[ordering[i]].Parents,
					tmpNodes[n.ID])
				tmpNodes[n.ID].Children = append(tmpNodes[n.ID].Children,
					tmpNodes[ordering[i]])
				tmpLinks = append(tmpLinks, link)
			}
		}
	}
	return dag
}

func computeT (dag *DAG, S map[utils.RouterID]struct{}) map[utils.RouterID]struct{} {

	U := dag.vertexs
	T := make(map[utils.RouterID]struct{})

	for id, node := range U {
		if _, ok := S[id]; ok {
			continue
		}
		for _, parent := range node.Parents {
			if _, ok := S[parent.ID]; ok {
				T[id] = struct{}{}
			}
		}
	}
	return T
}



