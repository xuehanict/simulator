package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func (g *Graph)CutOneDegree() int {
	nodesToDelete := make(map[RouterID]struct{})
	for _, n := range g.Nodes {
		if len(n.Neighbours) < 4 {
			nodesToDelete[n.ID]= struct{}{}
		}
	}

	fmt.Printf("node to delete length is %v\n", len(nodesToDelete))
	for id := range g.Nodes  {
		if _, ok := nodesToDelete[id]; ok {
			delete(g.Nodes, id)
		}
	}
	fmt.Printf("remove node done\n")

	for _, node := range g.Nodes  {
		for nToD := range nodesToDelete {
			node.RemoveNei(nToD)
		}
	}
	return len(nodesToDelete)
}

func (g *Graph)ConvertToSeriesID() map[RouterID]RouterID {
	i := RouterID(0)
	IDMap := make(map[RouterID]RouterID)
	finalNodes := make(map[RouterID]*Node)
	for orgID, node := range g.Nodes {
		IDMap[orgID] = i
		finalNodes[i] = node
		i++
	}

	for id, node := range finalNodes {
		for index, neiID := range node.Neighbours {
			node.Neighbours[index] = IDMap[neiID]
		}
		node.ID = id
	}

	//Convert channels
	channels := make(map[string]*Link)
	for _, link := range g.Channels {
		if _, ok := g.Nodes[link.Part1]; !ok {
			continue
		}
		if _, ok := g.Nodes[link.Part2]; !ok {
			continue
		}

		mapped1 := IDMap[link.Part1]
		mapped2 := IDMap[link.Part2]
		linkKey := GetLinkKey(mapped1, mapped2)
		linkValue := (link.Val2 + link.Val1)/2
		if mapped1 < mapped2{
			newLink := &Link{
				Part1: mapped1,
				Part2: mapped2,
				Val1: linkValue,
				Val2: linkValue,
			}
			channels[linkKey] = newLink
		} else {
			newLink := &Link {
				Part1: mapped2,
				Part2: mapped1,
				Val1: linkValue,
				Val2: linkValue,
			}
			channels[linkKey] = newLink
		}
	}
	g.Channels = channels
	g.Nodes = finalNodes
	return IDMap
}

func RandomTrans(trans []Tran, IDMap map[RouterID]RouterID, transNum int) []Tran {
	resTrans := make([]Tran, 0)
	rand.Seed(time.Now().UnixNano())
	for i := 0; len(resTrans)<transNum; i++ {
		tran := trans[rand.Intn(len(trans))]
		if _, ok := IDMap[RouterID(tran.Src)]; !ok {
			continue
		}
		if _, ok := IDMap[RouterID(tran.Dest)]; !ok {
			continue
		}

		newTran := Tran{
			Src: int(IDMap[RouterID(tran.Src)]),
			Dest: int(IDMap[RouterID(tran.Dest)]),
			Val: tran.Val,
		}
		resTrans = append(resTrans, newTran)
	}
	return resTrans
}


