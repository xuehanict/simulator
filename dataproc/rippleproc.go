package dataproc

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"math/rand"
	"time"
)

func  CutOneDegree(i int, g *utils.Graph) int {
	nodesToDelete := make(map[utils.RouterID]struct{})
	for _, n := range g.Nodes {
		if len(n.Neighbours) < i {
			nodesToDelete[n.ID] = struct{}{}
		}
	}

	fmt.Printf("node to delete length is %v\n", len(nodesToDelete))
	for id := range g.Nodes {
		if _, ok := nodesToDelete[id]; ok {
			delete(g.Nodes, id)
		}
	}
	fmt.Printf("remove node done\n")

	for _, node := range g.Nodes {
		for nToD := range nodesToDelete {
			node.RemoveNei(nToD)
		}
	}
	return len(nodesToDelete)
}

func ConvertToSeriesID(balance bool,g * utils.Graph) map[utils.RouterID]utils.RouterID {
	i := utils.RouterID(0)
	IDMap := make(map[utils.RouterID]utils.RouterID)
	finalNodes := make(map[utils.RouterID]*utils.Node)
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
	channels := make(map[string]*utils.Link)
	for _, link := range g.Channels {
		if _, ok := g.Nodes[link.Part1]; !ok {
			continue
		}
		if _, ok := g.Nodes[link.Part2]; !ok {
			continue
		}

		mapped1 := IDMap[link.Part1]
		mapped2 := IDMap[link.Part2]
		linkKey := utils.GetLinkKey(mapped1, mapped2)
		linkValue := (link.Val2 + link.Val1) / 2
		var newLink *utils.Link
		if mapped1 < mapped2 {
			newLink = &utils.Link{
				Part1: mapped1,
				Part2: mapped2,
				Val1:  linkValue,
				Val2:  linkValue,
			}
			if !balance {
				newLink.Val1 = link.Val1
				newLink.Val2 = link.Val2
			}
		} else {
			newLink = &utils.Link{
				Part1: mapped2,
				Part2: mapped1,
				Val1:  linkValue,
				Val2:  linkValue,
			}
			if !balance {
				newLink.Val2 = link.Val1
				newLink.Val1 = link.Val2
			}
		}
		channels[linkKey] = newLink
	}
	g.Channels = channels
	g.Nodes = finalNodes
	return IDMap
}

func RandomTrans(trans []utils.Tran, IDMap map[utils.RouterID]utils.RouterID, transNum int) []utils.Tran {
	resTrans := make([]utils.Tran, 0)
	rand.Seed(time.Now().UnixNano())
	for i := 0; len(resTrans) < transNum; i++ {
		tran := trans[rand.Intn(len(trans))]

			if _, ok := IDMap[utils.RouterID(tran.Src)]; !ok {
				continue
			}
			if _, ok := IDMap[utils.RouterID(tran.Dest)]; !ok {
				continue
			}


			newTran := utils.Tran{
				Src:  int(IDMap[utils.RouterID(tran.Src)]),
				Dest: int(IDMap[utils.RouterID(tran.Dest)]),
				Val:  tran.Val,
			}
		/*
		newTran := Tran{
			Src:  tran.Src % len(IDMap),
			Dest: tran.Dest % len(IDMap),
			Val:  tran.Val,
		}*/
		resTrans = append(resTrans, newTran)
	}
	return resTrans
}
